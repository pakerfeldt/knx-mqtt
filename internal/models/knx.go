package models

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/pakerfeldt/knx-mqtt/internal/utils"
	"gopkg.in/yaml.v3"
)

type FlatGroupAddress uint16

type GroupAddress struct {
	FullName    string
	Name        string
	Address     string
	FlatAddress FlatGroupAddress
	Datapoint   string
}

type KNX struct {
	NameToIndex    map[string]int
	GadToIndex     map[FlatGroupAddress]int
	GroupAddresses []GroupAddress
}

func EmptyKNX() KNX {
	return KNX{
		NameToIndex:    make(map[string]int),
		GadToIndex:     make(map[FlatGroupAddress]int),
		GroupAddresses: []GroupAddress{},
	}
}

func (k *KNX) AddGroupAddress(groupAddress GroupAddress) {
	k.GroupAddresses = append(k.GroupAddresses, groupAddress)
	index := len(k.GroupAddresses) - 1
	k.NameToIndex[groupAddress.FullName] = index
	k.GadToIndex[groupAddress.FlatAddress] = index
}

func (k *KNX) GetGroupAddress(address string) (*GroupAddress, bool) {
	var index int
	var exists bool
	if utils.IsRegularOrFlatGroupAddress(address) {
		flatAddr, _ := ParseGroupAddress(address)
		index, exists = k.GadToIndex[flatAddr]
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

// ParseGroupAddress parses a group-address in either plain notation (i.e. 0-65535),
// 2-part notation ("31/2047") or 3-part notation ("31/7/255")
// and returns it as a flat uint16 address and an error if parsing fails
func ParseGroupAddress(ga string) (FlatGroupAddress, error) {
	if ga == "" {
		return 0, fmt.Errorf("group address cannot be empty")
	}

	// Check if the address contains slashes (indicating 2-part or 3-part format)
	if strings.Contains(ga, "/") {
		parts := strings.Split(ga, "/")

		// Handle invalid number of parts
		if len(parts) != 2 && len(parts) != 3 {
			return 0, fmt.Errorf("invalid group address format: expected 2 or 3 parts, got %d parts", len(parts))
		}

		// Parse the first part (leftmost 5 bits) - common to both formats
		main, err := strconv.ParseUint(parts[0], 10, 16)
		if err != nil {
			return 0, fmt.Errorf("invalid main part '%s': %w", parts[0], err)
		}

		// Validate main part (5 bits: 0-31)
		if main > 31 {
			return 0, fmt.Errorf("main part '%d' exceeds maximum value of 31 (5 bits)", main)
		}

		// Shift the main part to the leftmost 5 bits position
		mainPart := uint16(main << 11)

		if len(parts) == 2 {
			// 2-part representation: main/sub (5 bits / 11 bits)
			sub, err := strconv.ParseUint(parts[1], 10, 16)
			if err != nil {
				return 0, fmt.Errorf("invalid sub part '%s': %w", parts[1], err)
			}

			// Validate sub part (11 bits: 0-2047)
			if sub > 2047 {
				return 0, fmt.Errorf("sub part '%d' exceeds maximum value of 2047 (11 bits)", sub)
			}

			return FlatGroupAddress(mainPart | uint16(sub)), nil
		} else if len(parts) == 3 {
			// 3-part representation: main/middle/sub (5 bits / 3 bits / 8 bits)
			middle, err := strconv.ParseUint(parts[1], 10, 16)
			if err != nil {
				return 0, fmt.Errorf("invalid middle part '%s': %w", parts[1], err)
			}

			// Validate middle part (3 bits: 0-7)
			if middle > 7 {
				return 0, fmt.Errorf("middle part '%d' exceeds maximum value of 7 (3 bits)", middle)
			}

			sub, err := strconv.ParseUint(parts[2], 10, 16)
			if err != nil {
				return 0, fmt.Errorf("invalid sub part '%s': %w", parts[2], err)
			}

			// Validate sub part (8 bits: 0-255)
			if sub > 255 {
				return 0, fmt.Errorf("sub part '%d' exceeds maximum value of 255 (8 bits)", sub)
			}

			// Shift middle to its position (bits 8-10) and combine with main and sub
			middlePart := uint16(middle << 8)
			return FlatGroupAddress(mainPart | middlePart | uint16(sub)), nil
		}
	} else {
		// Direct plain address format
		flatAddr, err := strconv.ParseUint(ga, 10, 16)
		if err != nil {
			return 0, fmt.Errorf("invalid plain address format '%s': %w", ga, err)
		}

		// Ensure the value fits in uint16
		if flatAddr > 65535 {
			return 0, fmt.Errorf("plain address '%d' exceeds maximum value of 65535 (16 bits)", flatAddr)
		}

		return FlatGroupAddress(flatAddr), nil
	}
	return 0, fmt.Errorf("invalid group address format")
}

type FlatAddressTranslation int

const (
	FAT_None FlatAddressTranslation = iota
	FAT_2_parts
	FAT_3_parts
)

// UnmarshalYAML implements the yaml.Unmarshaler interface for FlatAddressTranslation
func (fat *FlatAddressTranslation) UnmarshalYAML(value *yaml.Node) error {
	// Try to unmarshal as an integer first (for backward compatibility)
	var intValue int
	if err := value.Decode(&intValue); err == nil {
		switch intValue {
		case 0:
			*fat = FAT_None
		case 1:
			*fat = FAT_2_parts
		case 2:
			*fat = FAT_3_parts
		default:
			return fmt.Errorf("invalid FlatAddressTranslation value: %d", intValue)
		}
		return nil
	}

	// If not an integer, try to unmarshal as a string
	var stringValue string
	if err := value.Decode(&stringValue); err != nil {
		return fmt.Errorf("FlatAddressTranslation must be an integer or a string: %w", err)
	}

	// Convert string to lowercase for case-insensitive comparison
	switch strings.ToLower(stringValue) {
	case "none", "flat":
		*fat = FAT_None
	case "2-part", "two-part":
		*fat = FAT_2_parts
	case "3-part", "three-part":
		*fat = FAT_3_parts
	default:
		return fmt.Errorf("invalid FlatAddressTranslation string value: %s", stringValue)
	}

	return nil
}

func TranslateFlatAddress(fga FlatGroupAddress, translation FlatAddressTranslation) string {
	// Convert the flat address to uint16 for bit manipulation
	flatAddr := uint16(fga)

	switch translation {
	case FAT_None:
		// Return the flat address as a string without any translation
		return strconv.FormatUint(uint64(flatAddr), 10)

	case FAT_2_parts:
		// Extract the main part (5 leftmost bits)
		main := (flatAddr >> 11) & 0x1F // 0x1F = 31 (5 bits)

		// Extract the sub part (11 rightmost bits)
		sub := flatAddr & 0x7FF // 0x7FF = 2047 (11 bits)

		// Format as "main/sub"
		return fmt.Sprintf("%d/%d", main, sub)

	case FAT_3_parts:
		// Extract the main part (5 leftmost bits)
		main := (flatAddr >> 11) & 0x1F // 0x1F = 31 (5 bits)

		// Extract the middle part (3 bits after main)
		middle := (flatAddr >> 8) & 0x07 // 0x07 = 7 (3 bits)

		// Extract the sub part (8 rightmost bits)
		sub := flatAddr & 0xFF // 0xFF = 255 (8 bits)

		// Format as "main/middle/sub"
		return fmt.Sprintf("%d/%d/%d", main, middle, sub)

	default:
		// For any unrecognized translation, return the flat address
		return strconv.FormatUint(uint64(flatAddr), 10)
	}
}
