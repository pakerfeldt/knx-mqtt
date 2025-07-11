package bridge

import (
	"github.com/pakerfeldt/knx-mqtt/internal/knx"
	"github.com/pakerfeldt/knx-mqtt/internal/models"
	"github.com/pakerfeldt/knx-mqtt/internal/mqtt"
	"github.com/pakerfeldt/knx-mqtt/internal/msg"
	"github.com/rs/zerolog/log"
)

type Bridge struct {
	cfg        *models.Config
	knxItems   *models.KNX
	knxClient  *knx.KNXClient
	mqttClient *mqtt.MQTTClient
}

func NewBridge(config models.Config, knxItems *models.KNX, knxClient *knx.KNXClient, mqttClient *mqtt.MQTTClient) *Bridge {
	return &Bridge{
		cfg:        &config,
		knxItems:   knxItems,
		knxClient:  knxClient,
		mqttClient: mqttClient,
	}
}

func (b *Bridge) Start() {
	log.Info().Msg("Starting bridge ...")
	err := b.mqttClient.Connect(b.handleMQTTMessage)
	if err != nil {
		log.Fatal().Err(*err).Msg("Failed to establish connection to MQTT broker")
	}
	err = b.knxClient.Connect(b.handleKNXMessage)
	if err != nil {
		log.Fatal().Err(*err).Msg("Error connecting to KNX endpoint")
	}
}

func (b *Bridge) handleKNXMessage(message *msg.KNXMessage) {
	logMessage := log.Debug().Str("protocol", "knx").Str("address", message.Destination()).Str("command", message.Command())
	if message.IsResolved() {
		logMessage.Str("name", message.Name())
		if message.String() != "" {
			logMessage.Str("value", message.String())
		}
	}
	logMessage.Msg("Incoming")
	b.mqttClient.Send(*message)
}

func (b *Bridge) handleMQTTMessage(message *msg.MQTTMessage) {
	log.Debug().Str("protocol", "mqtt").Str("topic", message.Topic()).Str("payload", string(message.Bytes())).Msgf("Incoming")
	b.knxClient.Send(*message)
}
