package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pakerfeldt/knx-mqtt/internal/bridge"
	"github.com/pakerfeldt/knx-mqtt/internal/knx"
	"github.com/pakerfeldt/knx-mqtt/internal/models"
	"github.com/pakerfeldt/knx-mqtt/internal/mqtt"
	"github.com/pakerfeldt/knx-mqtt/internal/parser"
	"github.com/pakerfeldt/knx-mqtt/internal/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	var knxItems *models.KNX
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
		knxItems, err = parser.ReadGroupsFromFile(cfg.KNX.ETSExport, cfg.KNX.GaTranslation)
		if err != nil {
			log.Fatal().Str("error", fmt.Sprintf("%+v", err)).Msg("Error parsing KNX XML")
			os.Exit(1)
		}
	} else {
		if cfg.OutgoingMqttMessage.Type != "bytes" {
			log.Fatal().Msg("Outgoing MQTT message type can only be 'bytes' when no KNX addresses are imported. Change your config.")
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

	knxClient := knx.NewClient(ctx, *cfg, knxItems)
	mqttClient := mqtt.NewClient(*cfg)

	// Close upon exiting.
	defer knxClient.Close()
	defer mqttClient.Close()

	bridge := bridge.NewBridge(*cfg, knxItems, knxClient, mqttClient)
	bridge.Start()

	<-ctx.Done()

	stop()
	log.Info().Msg("Shutting down ...")
}
