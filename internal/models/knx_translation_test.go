package models

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFlatAddressTranslationUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlStr  string
		expected FlatAddressTranslation
		wantErr  bool
	}{
		{
			name:     "Integer 0",
			yamlStr:  "0",
			expected: FAT_None,
			wantErr:  false,
		},
		{
			name:     "Integer 1",
			yamlStr:  "1",
			expected: FAT_2_parts,
			wantErr:  false,
		},
		{
			name:     "Integer 2",
			yamlStr:  "2",
			expected: FAT_3_parts,
			wantErr:  false,
		},
		{
			name:     "Invalid Integer",
			yamlStr:  "3",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "String 'none'",
			yamlStr:  "none",
			expected: FAT_None,
			wantErr:  false,
		},
		{
			name:     "String 'flat'",
			yamlStr:  "flat",
			expected: FAT_None,
			wantErr:  false,
		},
		{
			name:     "String '2-part'",
			yamlStr:  "2-part",
			expected: FAT_2_parts,
			wantErr:  false,
		},
		{
			name:     "String 'two-part'",
			yamlStr:  "two-part",
			expected: FAT_2_parts,
			wantErr:  false,
		},
		{
			name:     "String '3-part'",
			yamlStr:  "3-part",
			expected: FAT_3_parts,
			wantErr:  false,
		},
		{
			name:     "String 'three-part'",
			yamlStr:  "three-part",
			expected: FAT_3_parts,
			wantErr:  false,
		},
		{
			name:     "Case insensitive 'NONE'",
			yamlStr:  "NONE",
			expected: FAT_None,
			wantErr:  false,
		},
		{
			name:     "Case insensitive 'Three-Part'",
			yamlStr:  "Three-Part",
			expected: FAT_3_parts,
			wantErr:  false,
		},
		{
			name:     "Invalid string",
			yamlStr:  "invalid",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fat FlatAddressTranslation
			err := yaml.Unmarshal([]byte(tt.yamlStr), &fat)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && fat != tt.expected {
				t.Errorf("UnmarshalYAML() got = %v, want %v", fat, tt.expected)
			}
		})
	}
}

func TestFlatAddressTranslationInConfig(t *testing.T) {
	// Test with integer value
	yamlWithInt := `
knx:
  translateFlatGroupAddresses: 2
`
	var configInt Config
	err := yaml.Unmarshal([]byte(yamlWithInt), &configInt)
	if err != nil {
		t.Errorf("Failed to unmarshal config with integer: %v", err)
	}
	if configInt.KNX.GaTranslation != FAT_3_parts {
		t.Errorf("Expected FAT_3_parts (2), got %v", configInt.KNX.GaTranslation)
	}

	// Test with string value
	yamlWithString := `
knx:
  translateFlatGroupAddresses: "two-part"
`
	var configString Config
	err = yaml.Unmarshal([]byte(yamlWithString), &configString)
	if err != nil {
		t.Errorf("Failed to unmarshal config with string: %v", err)
	}
	if configString.KNX.GaTranslation != FAT_2_parts {
		t.Errorf("Expected FAT_2_parts (1), got %v", configString.KNX.GaTranslation)
	}
}
