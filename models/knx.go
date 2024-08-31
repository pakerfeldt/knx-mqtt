package models

import (
	"errors"

	"github.com/pakerfeldt/knx-mqtt/utils"
)

type GroupAddress struct {
	FullName  string
	Name      string
	Address   string
	Datapoint string
}

type KNX struct {
	NameToIndex    map[string]int
	GadToIndex     map[string]int
	GroupAddresses []GroupAddress
}

func EmptyKNX() KNX {
	return KNX{
		NameToIndex:    make(map[string]int),
		GadToIndex:     make(map[string]int),
		GroupAddresses: []GroupAddress{},
	}
}

func (k *KNX) AddGroupAddress(groupAddress GroupAddress) {
	k.GroupAddresses = append(k.GroupAddresses, groupAddress)
	index := len(k.GroupAddresses) - 1
	k.NameToIndex[groupAddress.FullName] = index
	k.GadToIndex[groupAddress.Address] = index
}

func (k *KNX) GetGroupAddress(address string) (*GroupAddress, bool) {
	var index int
	var exists bool
	if utils.IsRegularGroupAddress(address) {
		index, exists = k.GadToIndex[address]
	} else {
		index, exists = k.NameToIndex[address]
	}
	if !exists {
		return nil, false
	}
	return &k.GroupAddresses[index], true
}

func (k *KNX) Is(address GroupAddress) error {
	if address.Name == "" {
		return errors.New("address name cannot be empty")
	}
	k.GroupAddresses = append(k.GroupAddresses, address)
	return nil
}

type KNXInterface interface {
	AddGroupAddress(address GroupAddress) error
	GetGroupAddress(name string) (*GroupAddress, error)
}
