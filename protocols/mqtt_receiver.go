package protocols

import (
	"fmt"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pakerfeldt/knx-mqtt/models"
	"github.com/pakerfeldt/knx-mqtt/utils"
	"github.com/rs/zerolog/log"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
)

func IncomingMqttMessageHandler(topicPrefix string, knxItems *models.KNX, knxClient KnxClient) func(mqttClient mqtt.Client, message mqtt.Message) {
	return func(mqttClient mqtt.Client, message mqtt.Message) {
		incomingMqttMessageHandler(topicPrefix, knxItems, knxClient, message)
	}
}

func incomingMqttMessageHandler(topicPrefix string, knxItems *models.KNX, knxClient KnxClient, msg mqtt.Message) {
	log.Debug().Str("protocol", "mqtt").Str("topic", msg.Topic()).Str("payload", string(msg.Payload())).Msgf("Incoming")
	log.Debug().Msgf("%s, %+v, %+v", msg.Topic(), msg.Qos(), msg.Retained())
	address := strings.TrimPrefix(msg.Topic()[:strings.LastIndex(msg.Topic(), "/")], topicPrefix)
	command := msg.Topic()[strings.LastIndex(msg.Topic(), "/")+1:]

	var event *knx.GroupEvent
	if command == "write" || command == "write-bytes" {
		writeBytes := command == "write-bytes"
		event = createGroupWriteKnxEvent(writeBytes, msg.Payload(), knxItems, address)
		if event == nil {
			return
		}
		if writeBytes {
			log.Debug().Str("protocol", "knx").Str("address", event.Destination.String()).Bool("binary", true).Msg("Outgoing")
		} else {
			log.Debug().Str("protocol", "knx").Str("address", event.Destination.String()).Bool("binary", false).Str("value", string(msg.Payload())).Msg("Outgoing")
		}
	} else if command == "read" {
		event = createGroupReadKnxEvent(knxItems, address)
	} else {
		log.Warn().Str("command", command).Msg("Unknown command")
		return
	}

	err := knxClient.Send(*event)
	if err != nil {
		log.Error().Str("error", fmt.Sprintf("%s", err)).Msgf("Error writing to KNX")
	}

}

func createGroupWriteKnxEvent(writeRawBinary bool, payload []byte, knxItems *models.KNX, addressFromMqtt string) *knx.GroupEvent {
	groupAddress, exists := knxItems.GetGroupAddress(addressFromMqtt)
	isRegularAddress := utils.IsRegularGroupAddress(addressFromMqtt)

	if !isRegularAddress && !exists {
		log.Error().Str("address", addressFromMqtt).Msg("Missing reference to group address")
		return nil
	}

	var destination cemi.GroupAddr
	var err error
	if exists {
		destination, err = cemi.NewGroupAddrString(groupAddress.Address)
	} else {
		destination, err = cemi.NewGroupAddrString(addressFromMqtt)
	}

	if err != nil {
		if exists {
			log.Error().Str("address", groupAddress.Address).Msg("Failed to create native group address")
		} else {
			log.Error().Str("address", addressFromMqtt).Msg("Failed to create native group address")
		}
		return nil
	}

	var packedBytes []byte
	if writeRawBinary {
		packedBytes = payload
	} else if exists {
		packedBytes, err = utils.PackString(groupAddress.Datapoint, string(payload))
	} else {
		log.Error().Str("address", destination.String()).Msg("Missing reference to group address for converting to non-binary payload")
		return nil
	}

	if err != nil {
		log.Error().Str("address", destination.String()).Str("datapoint", groupAddress.Datapoint).Str("payload", string(payload)).Msg("Failed to pack payload")
		return nil
	}

	return &knx.GroupEvent{
		Command:     knx.GroupWrite,
		Destination: destination,
		Data:        packedBytes,
	}

}

func createGroupReadKnxEvent(knxItems *models.KNX, addressFromMqtt string) *knx.GroupEvent {
	groupAddress, exists := knxItems.GetGroupAddress(addressFromMqtt)
	isRegularAddress := utils.IsRegularGroupAddress(addressFromMqtt)

	if !isRegularAddress && !exists {
		log.Error().Str("address", addressFromMqtt).Msg("Missing reference to group address")
		return nil
	}

	var destination cemi.GroupAddr
	var err error
	if exists {
		destination, err = cemi.NewGroupAddrString(groupAddress.Address)
	} else {
		destination, err = cemi.NewGroupAddrString(addressFromMqtt)
	}

	if err != nil {
		if exists {
			log.Error().Str("address", groupAddress.Address).Msg("Failed to create native group address")
		} else {
			log.Error().Str("address", addressFromMqtt).Msg("Failed to create native group address")
		}
		return nil
	}

	return &knx.GroupEvent{
		Command:     knx.GroupRead,
		Destination: destination,
	}
}
