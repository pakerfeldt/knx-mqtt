package parser

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pakerfeldt/knx-mqtt/internal/models"
	"github.com/pakerfeldt/knx-mqtt/internal/utils"
	"github.com/rs/zerolog/log"
)

var regexpDpt = regexp.MustCompile(`DPST-(\d+)-(\d+)`)

// FileID represents the type of KNX file being processed
type FileID int

const (
	// NA - not applicable, i.e. no file was ID'd
	NA FileID = iota

	// GroupAddressExport represents a KNX group address export file
	GroupAddressExport
	// ETSExport represents an ETS project export file
	ETSExport
)

func ReadGroupsFromFile(filePath string, gaTranslation models.FlatAddressTranslation) (*models.KNX, error) {
	fileid, byteValue, err := readAndIDFile(filePath)
	if err != nil {
		return nil, err
	}

	var export models.XmlGroupAddressExport

	switch fileid {
	case GroupAddressExport:
		log.Info().Msgf("%s identified as a group address export", filePath)
		err = xml.Unmarshal(byteValue, &export)
		if err != nil {
			return nil, err
		}

	case ETSExport:
		log.Info().Msgf("%s identified as a ETS export", filePath)
		var knxfile models.XmlKNX
		err = xml.Unmarshal(byteValue, &knxfile)
		if err != nil {
			return nil, err
		}
		installations := knxfile.Project.Installations
		if len(installations) == 0 {
			return nil, fmt.Errorf("no installations in ETS file")
		} else if len(installations) > 1 {
			log.Info().Msgf("Found multiple installations in ETS file; using the first one: \"%s\"", installations[0].Name)
		}
		export = installations[0].GroupAddresses.GroupRanges

		// Translate DatapointType -> DPTs
		translateDatapointTypeToDPTs(&export)

	default:
		return nil, fmt.Errorf("cannot handle file type %v", fileid)
	}

	knxItems := models.EmptyKNX()
	for _, main := range export.GroupRanges {
		for _, middle := range main.GroupRanges {
			for _, address := range middle.Addresses {
				if address.DPTs == "" {
					log.Warn().Msgf("%s with address %s did not have a DPT specified and will be ignored", address.Name, address.Address)
					continue
				}
				flatAddr, err := models.ParseGroupAddress(address.Address)
				if err != nil {
					return nil, err
				}
				translatedAddr := address.Address
				if utils.IsFlatGroupAddress(address.Address) {
					translatedAddr = models.TranslateFlatAddress(flatAddr, gaTranslation)
				}
				knxItems.AddGroupAddress(models.GroupAddress{
					Name:        address.Name,
					FullName:    fmt.Sprintf("%s/%s/%s", main.Name, middle.Name, replaceSlashInName(address.Name)),
					Address:     translatedAddr,
					FlatAddress: flatAddr,
					Datapoint:   convertDptFormat(address.DPTs),
				})
			}
		}
	}

	return &knxItems, nil
}

func readAndIDFile(filePath string) (FileID, []byte, error) {
	fid := NA
	file, err := os.Open(filePath)
	if err != nil {
		return fid, nil, err
	}
	defer file.Close()

	// bufreader := bufio.NewReader(file)
	magic := make([]byte, 4)
	n, err := file.Read(magic)
	if err != nil {
		return fid, nil, err
	}

	// Reset file position to beginning
	_, err = file.Seek(0, 0)
	if err != nil {
		return fid, nil, err
	}

	var maybeZipReader io.Reader
	maybeZipReader = file

	// Check if magic equals the ZIP signature 50 4b 03 04
	// If so, assume this is a ETS project export and look for 0.xml
	zipSignature := []byte{0x50, 0x4b, 0x03, 0x04}
	if n == 4 && bytes.Equal(magic, zipSignature) {
		log.Debug().Msg("ZIP signature detected, using gzip reader")

		rc, err := findGroupsInZip(file)
		if err != nil {
			return fid, nil, err
		}
		defer rc.Close()
		maybeZipReader = rc
		fid = ETSExport
	} else {
		fid = GroupAddressExport
	}

	byteValue, err := io.ReadAll(maybeZipReader)
	if err != nil {
		return fid, nil, err
	}
	return fid, byteValue, nil
}

func findGroupsInZip(file *os.File) (io.ReadCloser, error) {
	fs, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Create a gzip reader
	zipReader, err := zip.NewReader(file, fs.Size())
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	// Find a file "0.xml" in any path within the ZIP file
	for _, f := range zipReader.File {
		// Check if the file name is exactly "0.xml" at any path level
		if filepath.Base(f.Name) == "0.xml" {
			log.Debug().Msgf("Found 0.xml at path: %s", f.Name)

			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open 0.xml file: %w", err)
			}

			return rc, nil
		}
	}

	return nil, fmt.Errorf("no 0.xml file found in the ZIP archive")
}

func replaceSlashInName(name string) string {
	newName := strings.ReplaceAll(name, "/", "_")
	if newName != name {
		log.Warn().Msgf("%s was replaced with %s to avoid MQTT topic separation.", name, newName)
	}
	return newName
}

func convertDptFormat(dpt string) string {
	match := regexpDpt.FindStringSubmatch(dpt)
	// TODO: This must also match 14.1200
	if match != nil {
		return fmt.Sprintf("%s.%03s", match[1], match[2])
	}
	return ""
}

// translateDatapointTypeToDPTs iterates over all XmlGroupAddress objects nested inside export
// and checks if DPTs is empty. If so, it sets DPTs = DatapointType
func translateDatapointTypeToDPTs(export *models.XmlGroupAddressExport) {
	// Process each top-level GroupRange
	for i := range export.GroupRanges {
		translateGroupRange(&export.GroupRanges[i])
	}
}

// translateGroupRange recursively processes a GroupRange and its nested elements
func translateGroupRange(groupRange *models.XmlGroupRange) {
	// Process addresses at this level
	for i := range groupRange.Addresses {
		address := &groupRange.Addresses[i]
		if address.DPTs == "" && address.DatapointType != "" {
			address.DPTs = address.DatapointType
		}
	}

	// Recursively process nested GroupRanges
	for i := range groupRange.GroupRanges {
		translateGroupRange(&groupRange.GroupRanges[i])
	}
}
