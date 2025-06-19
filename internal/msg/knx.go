package msg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pakerfeldt/knx-mqtt/internal/models"
	"github.com/pakerfeldt/knx-mqtt/internal/utils"
	"github.com/rs/zerolog/log"
	knxgo "github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/dpt"
)

type KNXMessage struct {
	ge                knxgo.GroupEvent
	resolvedDatapoint *ResolvedDatapoint
}

type ResolvedDatapoint struct {
	datapoint    dpt.Datapoint
	groupAddress models.GroupAddress
}

func NewKNX(ge knxgo.GroupEvent, datapoint *dpt.Datapoint, groupAddress *models.GroupAddress) *KNXMessage {
	var resolvedDatapoint *ResolvedDatapoint
	if datapoint != nil || groupAddress != nil {
		resolvedDatapoint = &ResolvedDatapoint{datapoint: *datapoint, groupAddress: *groupAddress}
	}
	return &KNXMessage{ge: ge, resolvedDatapoint: resolvedDatapoint}
}

func (m KNXMessage) Destination() string {
	return m.ge.Destination.String()
}

func (m KNXMessage) Data() []byte {
	return m.ge.Data
}

func (m KNXMessage) IsResolved() bool {
	return m.resolvedDatapoint != nil
}

func (m KNXMessage) Name() string {
	if m.resolvedDatapoint == nil {
		return "<unresolved>"
	} else {
		return m.resolvedDatapoint.groupAddress.Name
	}
}

func (m KNXMessage) Address() string {
	if m.resolvedDatapoint == nil {
		return "<unresolved>"
	} else {
		return m.resolvedDatapoint.groupAddress.Address
	}
}

func (m KNXMessage) FullName() string {
	if m.resolvedDatapoint == nil {
		return "<unresolved>"
	} else {
		return m.resolvedDatapoint.groupAddress.FullName
	}
}

func (m KNXMessage) String() string {
	if m.resolvedDatapoint == nil {
		return "<unresolved value>"
	} else {
		return m.resolvedDatapoint.datapoint.String()
	}
}

func (m KNXMessage) Datapoint() string {
	if m.resolvedDatapoint == nil {
		return "<unresolved>"
	} else {
		return m.resolvedDatapoint.groupAddress.Datapoint
	}
}

func (m KNXMessage) ToPayload(emitValueAsString bool, messageType string, jsonFields *models.IncludedJsonFields) (interface{}, error) {
	var payload interface{}
	if messageType == models.JsonType {
		outgoingJson := models.OutgoingMqttJson{}
		if jsonFields.IncludeAddress {
			outgoingJson.Address = &m.resolvedDatapoint.groupAddress.Address
		}
		if jsonFields.IncludeName {
			outgoingJson.Name = &m.resolvedDatapoint.groupAddress.Name
		}
		if jsonFields.IncludeBytes {
			base64 := base64.StdEncoding.EncodeToString(m.resolvedDatapoint.datapoint.Pack())
			outgoingJson.Bytes = &base64
		}
		if jsonFields.IncludeDpt {
			outgoingJson.Dpt = m.Datapoint()
		}
		if jsonFields.IncludeValue {
			if emitValueAsString {
				outgoingJson.Value = utils.StringWithoutSuffix(m.resolvedDatapoint.datapoint)
			} else {
				outgoingJson.Value = utils.ExtractDatapointValue(m.resolvedDatapoint.datapoint, m.resolvedDatapoint.groupAddress.Datapoint)
			}
		}
		if jsonFields.IncludeUnit {
			unit := m.resolvedDatapoint.datapoint.Unit()
			outgoingJson.Unit = &unit
		}
		jsonBytes, err := json.Marshal(outgoingJson)
		if err != nil {
			log.Error().Str("error", fmt.Sprintf("%+v", err)).Msg("Failed to create outgoing JSON message")
			return nil, err
		}
		payload = string(jsonBytes)
	} else if messageType == models.ValueType {
		if emitValueAsString {
			payload = utils.StringWithoutSuffix(m.resolvedDatapoint.datapoint)
		} else {
			payload = fmt.Sprintf("%v", utils.ExtractDatapointValue(m.resolvedDatapoint.datapoint, m.resolvedDatapoint.groupAddress.Datapoint))
		}
	} else if messageType == models.ValueWithUnitType {
		payload = m.resolvedDatapoint.datapoint.String()
	} else if messageType == models.BytesType {
		payload = m.resolvedDatapoint.datapoint.Pack()
	}
	return payload, nil
}

func (m KNXMessage) ToPayload3(groupAddress models.GroupAddress, emitValueAsString bool, messageType string, jsonFields *models.IncludedJsonFields, addressName *string) (interface{}, *string, error) {
	datapoint, ok := dpt.Produce(groupAddress.Datapoint)
	if !ok {
		return nil, nil, fmt.Errorf("could not create datapoint %s", groupAddress.Datapoint)
	}
	datapoint.Unpack(m.Data())

	var payload interface{}
	if messageType == models.JsonType {
		outgoingJson := models.OutgoingMqttJson{}
		if jsonFields.IncludeBytes {
			base64 := base64.StdEncoding.EncodeToString(datapoint.Pack())
			outgoingJson.Bytes = &base64
		}
		if jsonFields.IncludeName {
			outgoingJson.Name = addressName
		}
		if jsonFields.IncludeValue {
			if emitValueAsString {
				outgoingJson.Value = utils.StringWithoutSuffix(datapoint)
			} else {
				outgoingJson.Value = utils.ExtractDatapointValue(datapoint, groupAddress.Datapoint)
			}
		}
		if jsonFields.IncludeUnit {
			unit := datapoint.Unit()
			outgoingJson.Unit = &unit
		}
		jsonBytes, err := json.Marshal(outgoingJson)
		if err != nil {
			log.Error().Str("error", fmt.Sprintf("%+v", err)).Msg("Failed to create outgoing JSON message")
			return nil, nil, err
		}
		payload = string(jsonBytes)
	} else if messageType == models.ValueType {
		if emitValueAsString {
			payload = utils.StringWithoutSuffix(datapoint)
		} else {
			payload = fmt.Sprintf("%v", utils.ExtractDatapointValue(datapoint, groupAddress.Datapoint))
		}
	} else if messageType == models.ValueWithUnitType {
		payload = datapoint.String()
	} else if messageType == models.BytesType {
		payload = datapoint.Pack()
	}
	strValue := datapoint.String()
	return payload, &strValue, nil
}
