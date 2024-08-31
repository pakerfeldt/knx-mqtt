package parser

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/pakerfeldt/knx-mqtt/models"
	"github.com/rs/zerolog/log"
)

var regexpDpt = regexp.MustCompile(`DPST-(\d+)-(\d+)`)

func ParseGroupAddressExport(filePath string) (*models.KNX, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var export models.XmlGroupAddressExport
	err = xml.Unmarshal(byteValue, &export)
	if err != nil {
		return nil, err
	}

	knxItems := models.EmptyKNX()
	for _, main := range export.GroupRanges {
		for _, middle := range main.GroupRanges {
			for _, address := range middle.Addresses {
				if address.DPTs == "" {
					log.Warn().Msgf("%s with address %s did not have a DPT specified and will be ignored", address.Name, address.Address)
					continue
				}
				knxItems.AddGroupAddress(models.GroupAddress{
					Name:      address.Name,
					FullName:  fmt.Sprintf("%s/%s/%s", main.Name, middle.Name, address.Name),
					Address:   address.Address,
					Datapoint: convertDptFormat(address.DPTs),
				})
			}
		}
	}

	return &knxItems, nil
}

func convertDptFormat(dpt string) string {
	match := regexpDpt.FindStringSubmatch(dpt)
	// TODO: This must also match 14.1200
	if match != nil {
		return fmt.Sprintf("%s.%03s", match[1], match[2])
	}
	return ""
}
