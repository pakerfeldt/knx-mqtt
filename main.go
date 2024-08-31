package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pakerfeldt/knx-mqtt/models"
	"github.com/pakerfeldt/knx-mqtt/parser"
	"github.com/pakerfeldt/knx-mqtt/protocols"
	"github.com/pakerfeldt/knx-mqtt/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vapourismo/knx-go/knx"
)

func main() {

	var knxClient protocols.KnxClient
	var knxItems *models.KNX

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load the configuration
	var configPath, exists = os.LookupEnv("KNX_MQTT_CONFIG")
	if !exists {
		configPath = "config.yaml"
	}
	cfg, err := parser.LoadConfig(configPath)
	if err != nil {
		log.Fatal().Str("error", fmt.Sprintf("%+v", err)).Msg("Error loading config")
		os.Exit(1)
	}

	utils.SetupLogging(cfg.LogLevel, cfg.KNX.EnableLogs)

	if cfg.KNX.ETSExport != "" {
		knxItems, err = parser.ParseGroupAddressExport(cfg.KNX.ETSExport)
		if err != nil {
			log.Fatal().Str("error", fmt.Sprintf("%+v", err)).Msg("Error parsing KNX XML")
			os.Exit(1)
		}
	} else {
		if cfg.OutgoingMqttMessage.Type != "bytes" {
			log.Fatal().Msg("Outgoing MQTT message type can only be 'bytes' when no KNX address are imported. Change your config.")
			os.Exit(1)
		}
		log.Info().Msg("Outgoing MQTT messages will only be emitted using their address.")
		cfg.OutgoingMqttMessage.EmitUsingAddress = true
		cfg.OutgoingMqttMessage.EmitUsingName = false
		emptyKnx := models.EmptyKNX()
		knxItems = &emptyKnx
	}

	// Create a context that is cancelled on SIGINT (Ctrl+C) or SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var (
		gt *knx.GroupTunnel
		gr *knx.GroupRouter
	)

	if cfg.KNX.TunnelMode {
		tunnel, knxError := knx.NewGroupTunnel(cfg.KNX.Endpoint, knx.DefaultTunnelConfig)
		if knxError != nil {
			log.Fatal().Msgf("Error connecting to KNX endpoint %+v", knxError)
			os.Exit(1)
		}
		gt = &tunnel
	} else {
		router, knxError := knx.NewGroupRouter(cfg.KNX.Endpoint, knx.DefaultRouterConfig)
		if knxError != nil {
			log.Fatal().Msgf("Error connecting to KNX endpoint %+v", knxError)
			os.Exit(1)
		}
		gr = &router
	}
	knxClient = protocols.NewKnxClient(gt, gr)

	mqttOptions := mqtt.NewClientOptions()
	if cfg.MQTT.Username != nil {
		mqttOptions.SetUsername(*cfg.MQTT.Username)
	}
	if cfg.MQTT.Password != nil {
		mqttOptions.SetPassword(*cfg.MQTT.Password)
	}
	if cfg.MQTT.ClientID != nil {
		mqttOptions.SetClientID(*cfg.MQTT.ClientID)
	} else {
		mqttOptions.SetClientID("knx-mqtt")
	}

	mqttOpts := mqttOptions.AddBroker(cfg.MQTT.URL)
	mqttOpts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Error().Str("error", fmt.Sprintf("%+v", err)).Msg("Connection to MQTT broker lost")
	}
	mqttClient := mqtt.NewClient(mqttOpts)

	// Close upon exiting.
	defer knxClient.Close()
	defer mqttClient.Disconnect(1)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal().Str("error", fmt.Sprintf("%+v", token.Error())).Msg("Failed to establish connection to MQTT broker")
		os.Exit(1)
	}

	incomingMqttMessageHandler := protocols.IncomingMqttMessageHandler(cfg.MQTT.TopicPrefix, knxItems, knxClient)
	if token := mqttClient.Subscribe(cfg.MQTT.TopicPrefix+"+/+/+/+", 0, incomingMqttMessageHandler); token.Wait() && token.Error() != nil {
		log.Fatal().Str("error", fmt.Sprintf("%+v", token.Error())).Msg("Failed to subscribe to MQTT")
		os.Exit(1)
	}

	incomingKnxEventHandler := protocols.IncomingKnxEventHandler(mqttClient, knxItems, cfg.OutgoingMqttMessage, cfg.MQTT, cfg.KNX.IgnoreUnknownGroupAddresses)
	protocols.SubscribeKnx(knxClient, incomingKnxEventHandler)

	<-ctx.Done()

	stop()
	log.Info().Msg("Shutting down ...")
	mqttClient.Disconnect(250)
}
