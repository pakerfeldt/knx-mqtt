package protocols

import (
	"bytes"
	"testing"

	"github.com/pakerfeldt/knx-mqtt/models"
	"github.com/pakerfeldt/knx-mqtt/utils"
	"github.com/vapourismo/knx-go/knx/dpt"
)

func TestCreateGroupWriteKnxEvent_StringPayload(t *testing.T) {

	items := models.EmptyKNX()
	items.AddGroupAddress(models.GroupAddress{
		FullName:  "Human/Readable/Device",
		Name:      "Device",
		Address:   "2/5/10",
		Datapoint: "9.001",
	})

	payload := "20.56"
	event := createGroupWriteKnxEvent(false, []byte(payload), &items, "2/5/10")
	if event == nil {
		t.Fatal("Expected event, but got nil")
	}

	datapoint, _ := dpt.Produce("9.001")
	datapoint.Unpack(event.Data)

	if payload != utils.StringWithoutSuffix(datapoint) {
		t.Fatalf("Expected %s, but got %s", payload, utils.StringWithoutSuffix(datapoint))
	}
}

func TestCreateGroupWriteKnxEvent_BinaryPayload(t *testing.T) {

	items := models.EmptyKNX()

	payload := []byte{0xDE, 0xAD}
	event := createGroupWriteKnxEvent(true, payload, &items, "1/2/3")
	if event == nil {
		t.Fatal("Expected event, but got nil")
	}

	if !bytes.Equal(payload, event.Data) {
		t.Fatalf("Expected %#v, but got %#v", payload, event.Data)
	}
}

func TestCreateGroupWriteKnxEvent_AddressByNameWithoutReference(t *testing.T) {

	items := models.EmptyKNX()

	event := createGroupWriteKnxEvent(false, []byte("20.5"), &items, "Address/By/Name")
	if event != nil {
		t.Error("Expected creation to fail and event to be nil")
	}
}

func TestCreateGroupWriteKnxEvent_AddressByNameWithoutReferenceAndBinaryPayload(t *testing.T) {

	items := models.EmptyKNX()

	event := createGroupWriteKnxEvent(true, []byte{0x00}, &items, "Address/By/Name")
	if event != nil {
		t.Error("Expected creation to fail and event to be nil")
	}
}
