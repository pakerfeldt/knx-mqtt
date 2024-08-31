package models

type XmlGroupAddressExport struct {
    GroupRanges []XmlGroupRange `xml:"GroupRange"`
}

type XmlGroupRange struct {
    Name        string        `xml:"Name,attr"`
    GroupRanges []XmlGroupRange  `xml:"GroupRange"`    // For nested GroupRanges
    Addresses   []XmlGroupAddress `xml:"GroupAddress"` // For GroupAddresses
}

type XmlGroupAddress struct {
    Name        string `xml:"Name,attr"`
    Address     string `xml:"Address,attr"`
    DPTs        string `xml:"DPTs,attr,omitempty"`
    Description string `xml:"Description,attr,omitempty"`
}