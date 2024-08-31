package protocols

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/pakerfeldt/knx-mqtt/models"
	"github.com/vapourismo/knx-go/knx/dpt"
)

func TestConstructPayload_ValuePayload(t *testing.T) {
	dpt := dpt.DPT_1001(true)

	payload, err := constructPayload(&dpt, models.ValueType, nil, nil)

	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	if payload != "On" {
		t.Fatalf("Expected %s, but got %s", "On", payload)
	}
}

func TestConstructPayload_ValueWithUnitPayload(t *testing.T) {
	dpt := dpt.DPT_5004(95)

	payload, err := constructPayload(&dpt, models.ValueWithUnitType, nil, nil)

	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	if payload != "95.00%" {
		t.Fatalf("Expected %s, but got %s", "95.00%", payload)
	}
}

func TestConstructPayload_BytesPayload(t *testing.T) {
	dpt := dpt.DPT_1002(true)

	payload, err := constructPayload(&dpt, models.BytesType, nil, nil)

	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	if b, ok := payload.([]byte); ok {
		if !bytes.Equal(b, []byte{0x01}) {
			t.Fatalf("Expected %s, but got %s", "0x01", payload)
		}
	} else {
		t.Fatalf("Expected payload to be []byte")
	}
}

func TestConstructPayload_JsonPayloadFull(t *testing.T) {
	dpt := dpt.DPT_9001(23.4)

	jsonFields := models.IncludedJsonFields{
		IncludeBytes: true,
		IncludeName:  true,
		IncludeValue: true,
		IncludeUnit:  true,
	}
	name := "My device"
	serializedPayload, err := constructPayload(&dpt, models.JsonType, &jsonFields, &name)
	fmt.Printf("String: %s\n", serializedPayload)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	serializedJson, ok := serializedPayload.(string)
	if !ok {
		t.Fatalf("Expected response of type string, got %T", serializedPayload)
		return
	}

	var payload models.OutgoingMqttJson

	err = json.Unmarshal([]byte(serializedJson), &payload)

	if err != nil {
		t.Fatalf("Error unmarshalling payload %s", serializedPayload)
	}

	if payload.Bytes == nil || *payload.Bytes != base64.StdEncoding.EncodeToString(dpt.Pack()) {
		t.Fatalf("Expected Bytes to be '%s', but got '%v'", base64.StdEncoding.EncodeToString(dpt.Pack()), payload.Bytes)
	}

	if payload.Name == nil || *payload.Name != "My device" {
		t.Fatalf("Expected Name to be '%s', but got '%s'", "My device", *payload.Name)
	}

	fmt.Printf("Value: %s\n", *payload.Value)

	if payload.Value == nil || *payload.Value != "23.40" {
		t.Fatalf("Expected Value to be '%s', but got '%s'", "23.40", *payload.Value)
	}

	if payload.Unit == nil || *payload.Unit != "°C" {
		t.Fatalf("Expected Unit to be '%s', but got '%s'", "°C", *payload.Unit)
	}

}
