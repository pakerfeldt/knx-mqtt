package models

import (
	"encoding/xml"
)

// // XmlETSRoot represents the root ETS file export
// type XmlETSRoot struct {
// 	KNX XmlKNX `xml:"KNX"`
// }

// XmlKNX represents the root KNX element
type XmlKNX struct {
	XMLName xml.Name   `xml:"KNX"`
	Project XmlProject `xml:"Project"`
}

// XmlProject represents the Project element
type XmlProject struct {
	Installations []XmlInstallation `xml:"Installations>Installation"`
}

// XmlInstallation represents the Installation element
type XmlInstallation struct {
	Name           string `xml:",attr"`
	GroupAddresses XmlGroupAddresses
}

// XmlGroupAddresses represents the GroupAddresses element
type XmlGroupAddresses struct {
	GroupRanges XmlGroupAddressExport `xml:"GroupRanges"`
}

// XmlGroupAddressExport represents the GroupRanges element for direct export format
type XmlGroupAddressExport struct {
	GroupRanges []XmlGroupRange `xml:"GroupRange"`
}

type XmlGroupRange struct {
	Name        string            `xml:"Name,attr"`
	GroupRanges []XmlGroupRange   `xml:"GroupRange"`   // For nested GroupRanges
	Addresses   []XmlGroupAddress `xml:"GroupAddress"` // For GroupAddresses
}

type XmlGroupAddress struct {
	Name          string `xml:"Name,attr"`
	Address       string `xml:"Address,attr"`
	DPTs          string `xml:"DPTs,attr,omitempty"`          // DPTs is the attribute's name in group address exports
	DatapointType string `xml:"DatapointType,attr,omitempty"` // DatapointType is the attribute's name in ETS exports
	Description   string `xml:"Description,attr,omitempty"`
}
