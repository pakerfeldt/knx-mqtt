package mqtt

import (
	"fmt"

	"github.com/pakerfeldt/knx-mqtt/internal/models"
	"github.com/pakerfeldt/knx-mqtt/internal/msg"
	"github.com/rs/zerolog/log"

	mqttgo "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	cfg      *models.Config
	client   mqttgo.Client
	callback *func(*msg.MQTTMessage)
}

func NewClient(config models.Config) *MQTTClient {
	c := &MQTTClient{
		cfg:      &config,
		client:   nil,
		callback: nil,
	}
	mqttOptions := mqttgo.NewClientOptions()
	if config.MQTT.Username != nil {
		mqttOptions.SetUsername(*config.MQTT.Username)
	}
	if config.MQTT.Password != nil {
		mqttOptions.SetPassword(*config.MQTT.Password)
	}
	if config.MQTT.ClientID != nil {
		mqttOptions.SetClientID(*config.MQTT.ClientID)
	}

	if config.MQTT.TLSCA != nil && config.MQTT.TLSCert != nil && config.MQTT.TLSKey != nil {
		tlsConfig, err := NewTLSConfig(*config.MQTT.TLSCA, *config.MQTT.TLSCert, *config.MQTT.TLSKey)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create TLS configuration")
		} else {
			mqttOptions.SetTLSConfig(tlsConfig)
		}
	}
	mqttOptions.AddBroker(config.MQTT.URL)

	mqttOptions.OnConnectionLost = func(client mqttgo.Client, err error) {
		log.Error().Str("error", fmt.Sprintf("%+v", err)).Msg("Connection to MQTT broker lost")
	}
	mqttOptions.SetOnConnectHandler(c.onConnect)
	c.client = mqttgo.NewClient(mqttOptions)
	return c
}

func (c *MQTTClient) onConnect(client mqttgo.Client) {
	token := c.client.Subscribe(c.cfg.MQTT.TopicPrefix+"+/+/+/+", 0, func(client mqttgo.Client, m mqttgo.Message) {
		if c.callback != nil {
			(*c.callback)(msg.NewMQTT(m))
		}
	})
	token.Wait()
	if token.Error() != nil {
		log.Warn().Msg("Failed to connect to MQTT broker")
	} else {
		log.Info().Msg("Subscribed to MQTT")
	}
}

func (c *MQTTClient) Connect(callback func(*msg.MQTTMessage)) *error {
	c.callback = &callback
	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		err := token.Error()
		return &err
	}
	return nil
}

func (c *MQTTClient) Close() {
	c.client.Disconnect(1)
}

func (c *MQTTClient) Subscribe(callback func(*msg.MQTTMessage)) {
	c.callback = &callback
}

func (c *MQTTClient) Send(message msg.KNXMessage) {
	if !message.IsResolved() && c.cfg.OutgoingMqttMessage.Type == "bytes" && c.cfg.OutgoingMqttMessage.EmitUsingAddress {
		c.client.Publish(c.cfg.MQTT.TopicPrefix+message.Destination(), c.cfg.MQTT.Qos, c.cfg.MQTT.Retain, message.Data())
		return
	} else if !message.IsResolved() {
		log.Info().Str("address", message.Destination()).Msg("Cannot read unknown address, update your KNX XML export")
		return
	}

	var payload interface{}
	if c.cfg.OutgoingMqttMessage.Type == "bytes" {
		payload = message.Data()
	} else {
		var err error
		payload, err = message.ToPayload(c.cfg.OutgoingMqttMessage.EmitValueAsString, c.cfg.OutgoingMqttMessage.Type, &c.cfg.OutgoingMqttMessage.IncludedJsonFields)
		if err != nil {
			log.Warn().Str("address", message.Destination()).Msgf("Could not create message payload for %s for address %s", message.Datapoint(), message.Destination())
			return
		}
	}

	if c.cfg.OutgoingMqttMessage.EmitUsingAddress {
		c.client.Publish(c.cfg.MQTT.TopicPrefix+message.Address(), c.cfg.MQTT.Qos, c.cfg.MQTT.Retain, payload)
	}
	if c.cfg.OutgoingMqttMessage.EmitUsingName {
		c.client.Publish(c.cfg.MQTT.TopicPrefix+message.FullName(), c.cfg.MQTT.Qos, c.cfg.MQTT.Retain, payload)
	}
}
