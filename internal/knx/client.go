package knx

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pakerfeldt/knx-mqtt/internal/models"
	"github.com/pakerfeldt/knx-mqtt/internal/msg"
	"github.com/pakerfeldt/knx-mqtt/internal/utils"
	"github.com/rs/zerolog/log"
	knxgo "github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
	"github.com/vapourismo/knx-go/knx/dpt"
)

type KNXClient struct {
	ctx      context.Context
	cancel   context.CancelFunc
	cfg      *models.Config
	knxItems *models.KNX
	tunnel   *knxgo.GroupTunnel
	router   *knxgo.GroupRouter
}

func NewClient(ctx context.Context, config models.Config, knxItems *models.KNX) *KNXClient {
	childCtx, cancel := context.WithCancel(ctx)
	client := KNXClient{ctx: childCtx, cancel: cancel, cfg: &config, knxItems: knxItems}
	return &client
}

func (c *KNXClient) Connect(callback func(*msg.KNXMessage)) *error {
	err := c.connect()
	if err != nil {
		return err
	}
	c.subscribe(callback)
	return nil
}

func (c *KNXClient) connect() *error {
	if c.cfg.KNX.TunnelMode {
		tunnel, err := knxgo.NewGroupTunnel(c.cfg.KNX.Endpoint, knxgo.DefaultTunnelConfig)
		if err != nil {
			return &err
		}
		c.tunnel = &tunnel
	} else {
		router, err := knxgo.NewGroupRouter(c.cfg.KNX.Endpoint, knxgo.DefaultRouterConfig)
		if err != nil {
			return &err
		}
		c.router = &router
	}
	return nil
}

func (c *KNXClient) newMessage(event knxgo.GroupEvent) *msg.KNXMessage {
	destination := event.Destination.String()
	flatDestination, err := models.ParseGroupAddress(destination)
	if err != nil {
		log.Error().Err(err).Str("address", destination).Msg("Failed to parse KNX group address")
		return msg.NewKNX(event, nil, nil)
	}
	index, exists := c.knxItems.GadToIndex[flatDestination]
	if !exists {
		return msg.NewKNX(event, nil, nil)
	}
	groupAddress := c.knxItems.GroupAddresses[index]
	datapoint, ok := dpt.Produce(groupAddress.Datapoint)
	if !ok {
		log.Error().Msgf("Failed to create datapoint %s", groupAddress.Datapoint)
		return msg.NewKNX(event, nil, nil)
	}
	datapoint.Unpack(event.Data)
	return msg.NewKNX(event, &datapoint, &groupAddress)
}

func (c *KNXClient) subscribe(callback func(*msg.KNXMessage)) {
	go func() {
	Listening:
		for {
			log.Info().Msg("Subscribed to KNX")

		ReadEvent:
			for {
				select {
				case <-c.ctx.Done():
					log.Info().Msg("Stopping KNX subscription...")
					return
				case event, ok := <-c.tunnel.Inbound():
					if !ok {
						break ReadEvent
					}

					message := c.newMessage(event)
					callback(message)
				}
			}

			log.Error().Msg("Lost connection to KNX, trying to reconnect ...")
			for {
				select {
				case <-c.ctx.Done():
					log.Info().Msg("Stopping KNX reconnection...")
					return
				default:
					err := c.connect()
					if err == nil {
						continue Listening
					}
					log.Error().Err(*err).Msg("Failed to connect to KNX, retrying in 5s...")
					time.Sleep(5 * time.Second)
				}
			}
		}
	}()
}

func (c *KNXClient) createWriteEvent(payload []byte, address string, writeRawBinary bool) *knxgo.GroupEvent {
	groupAddress, exists := c.knxItems.GetGroupAddress(address)
	isRegularAddress := utils.IsRegularGroupAddress(address)

	if !isRegularAddress && !exists {
		log.Error().Str("address", address).Msg("Missing reference to group address")
		return nil
	}

	var destination cemi.GroupAddr
	var err error
	if exists {
		destination, err = cemi.NewGroupAddrString(groupAddress.Address)
	} else {
		destination, err = cemi.NewGroupAddrString(address)
	}

	if err != nil {
		if exists {
			log.Error().Str("address", groupAddress.Address).Msg("Failed to create native group address")
		} else {
			log.Error().Str("address", address).Msg("Failed to create native group address")
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

	return &knxgo.GroupEvent{
		Command:     knxgo.GroupWrite,
		Destination: destination,
		Data:        packedBytes,
	}
}

func (c *KNXClient) createReadEvent(address string) *knxgo.GroupEvent {
	groupAddress, exists := c.knxItems.GetGroupAddress(address)
	isRegularAddress := utils.IsRegularGroupAddress(address)

	if !isRegularAddress && !exists {
		log.Error().Str("address", address).Msg("Missing reference to group address")
		return nil
	}

	var destination cemi.GroupAddr
	var err error
	if exists {
		destination, err = cemi.NewGroupAddrString(groupAddress.Address)
	} else {
		destination, err = cemi.NewGroupAddrString(address)
	}

	if err != nil {
		if exists {
			log.Error().Str("address", groupAddress.Address).Msg("Failed to create native group address")
		} else {
			log.Error().Str("address", address).Msg("Failed to create native group address")
		}
		return nil
	}

	return &knxgo.GroupEvent{
		Command:     knxgo.GroupRead,
		Destination: destination,
	}
}

func (c *KNXClient) Send(message msg.MQTTMessage) {
	address := strings.TrimPrefix(message.Topic()[:strings.LastIndex(message.Topic(), "/")], c.cfg.MQTT.TopicPrefix)
	command := message.Topic()[strings.LastIndex(message.Topic(), "/")+1:]

	var event *knxgo.GroupEvent
	if command == "write" || command == "write-bytes" {
		writeBytes := command == "write-bytes"
		event = c.createWriteEvent(message.Bytes(), address, writeBytes)
		if event == nil {
			log.Error().Msg("Failed to create KNX write event")
			return
		}
		if writeBytes {
			log.Debug().Str("protocol", "knx").Str("address", event.Destination.String()).Bool("binary", true).Msg("Outgoing")
		} else {
			log.Debug().Str("protocol", "knx").Str("address", event.Destination.String()).Bool("binary", false).Str("value", string(message.Bytes())).Msg("Outgoing")
		}
	} else if command == "read" {
		event = c.createReadEvent(address)
		if event == nil {
			log.Error().Msg("Failed to create KNX read event")
			return
		}
	} else {
		log.Warn().Str("command", command).Msg("Unknown command")
		return
	}

	err := c.send(*event)
	if err != nil {
		log.Error().Str("error", fmt.Sprintf("%s", err)).Msgf("Error writing to KNX")
	}
}

func (c *KNXClient) Router() *knxgo.GroupRouter {
	return c.router
}

func (c *KNXClient) send(event knxgo.GroupEvent) error {
	if c.tunnel != nil {
		return c.tunnel.Send(event)
	}
	if c.router != nil {
		return c.router.Send(event)
	}
	return fmt.Errorf("no valid KNX client initialized")
}

func (c *KNXClient) Inbound() <-chan knxgo.GroupEvent {
	if c.tunnel != nil {
		return c.tunnel.Inbound()
	}
	if c.router != nil {
		return c.router.Inbound()
	}

	// Should never happen, but just return closed channel
	closedChan := make(chan knxgo.GroupEvent)
	close(closedChan)
	return closedChan
}

func (c *KNXClient) Close() {
	if c.tunnel != nil {
		c.tunnel.Close()
	}
	if c.router != nil {
		c.router.Close()
	}
}
