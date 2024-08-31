package protocols

import (
	"fmt"

	"github.com/vapourismo/knx-go/knx"
)

type KnxClient struct {
	tunnel *knx.GroupTunnel
	router *knx.GroupRouter
}

func (kc *KnxClient) Router() *knx.GroupRouter {
	return kc.router
}

func (kc *KnxClient) Send(event knx.GroupEvent) error {
	if kc.tunnel != nil {
		return kc.tunnel.Send(event)
	}
	if kc.router != nil {
		return kc.router.Send(event)
	}
	return fmt.Errorf("no valid KNX client initialized")
}

func (kc *KnxClient) Inbound() <-chan knx.GroupEvent {
	if kc.tunnel != nil {
		return kc.tunnel.Inbound()
	}
	if kc.router != nil {
		return kc.router.Inbound()
	}

	// Should never happen, but just return closed channel
	closedChan := make(chan knx.GroupEvent)
	close(closedChan)
	return closedChan
}

func (kc *KnxClient) Close() {
	if kc.tunnel != nil {
		kc.tunnel.Close()
	}
	if kc.router != nil {
		kc.router.Close()
	}
}

func NewKnxClient(gt *knx.GroupTunnel, gr *knx.GroupRouter) KnxClient {
	return KnxClient{
		tunnel: gt,
		router: gr,
	}
}
