package models

type OutgoingMqttJson struct {
	Bytes   *string `json:"bytes,omitempty"`
	Address *string `json:"address,omitempty"`
	Name    *string `json:"name,omitempty"`
	Dpt     string  `json:"dpt,omitempty"`
	Value   any     `json:"value,omitempty"`
	Unit    *string `json:"unit,omitempty"`
}
