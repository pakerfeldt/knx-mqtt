package models

import (
	"testing"
)

func TestParseGroupAddress(t *testing.T) {
	// Test cases for valid inputs
	tests := []struct {
		name    string
		address string
		want    FlatGroupAddress
	}{
		// Direct plain address format
		{
			name:    "Direct plain address",
			address: "12345",
			want:    12345,
		},
		{
			name:    "Direct plain address zero",
			address: "0",
			want:    0,
		},
		// 2-part representation (main/sub)
		{
			name:    "2-part: main=0, sub=0",
			address: "0/0",
			want:    0,
		},
		{
			name:    "2-part: main=1, sub=0",
			address: "1/0",
			want:    1 << 11, // 2048
		},
		{
			name:    "2-part: main=0, sub=1",
			address: "0/1",
			want:    1,
		},
		{
			name:    "2-part: main=15, sub=2047",
			address: "15/2047",
			want:    FlatGroupAddress((15 << 11) | 2047), // 32767
		},
		// 3-part representation (main/middle/sub)
		{
			name:    "3-part: main=0, middle=0, sub=0",
			address: "0/0/0",
			want:    0,
		},
		{
			name:    "3-part: main=1, middle=0, sub=0",
			address: "1/0/0",
			want:    1 << 11, // 2048
		},
		{
			name:    "3-part: main=0, middle=1, sub=0",
			address: "0/1/0",
			want:    1 << 8, // 256
		},
		{
			name:    "3-part: main=0, middle=0, sub=1",
			address: "0/0/1",
			want:    1,
		},
		{
			name:    "3-part: main=15, middle=7, sub=255",
			address: "15/7/255",
			want:    FlatGroupAddress((15 << 11) | (7 << 8) | 255), // 32767
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGroupAddress(tt.address)
			if err != nil {
				t.Errorf("ParseGroupAddress(%q) error = %v, want no error", tt.address, err)
				return
			}
			if got != tt.want {
				t.Errorf("ParseGroupAddress(%q) = %v, want %v", tt.address, got, tt.want)
			}
		})
	}

	// Test cases for invalid inputs
	errorTests := []struct {
		name    string
		address string
	}{
		{
			name:    "Empty string",
			address: "",
		},
		{
			name:    "Invalid format - non-numeric",
			address: "abc",
		},
		{
			name:    "Invalid 2-part - non-numeric main",
			address: "a/123",
		},
		{
			name:    "Invalid 2-part - non-numeric sub",
			address: "1/abc",
		},
		{
			name:    "Invalid 3-part - non-numeric main",
			address: "a/1/123",
		},
		{
			name:    "Invalid 3-part - non-numeric middle",
			address: "1/a/123",
		},
		{
			name:    "Invalid 3-part - non-numeric sub",
			address: "1/2/abc",
		},
		{
			name:    "Invalid 2-part - main too large",
			address: "32/123", // main > 31 (5 bits)
		},
		{
			name:    "Invalid 2-part - sub too large",
			address: "1/2048", // sub > 2047 (11 bits)
		},
		{
			name:    "Invalid 3-part - main too large",
			address: "32/1/123", // main > 31 (5 bits)
		},
		{
			name:    "Invalid 3-part - middle too large",
			address: "1/8/123", // middle > 7 (3 bits)
		},
		{
			name:    "Invalid 3-part - sub too large",
			address: "1/2/256", // sub > 255 (8 bits)
		},
		{
			name:    "Too many parts",
			address: "1/2/3/4",
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGroupAddress(tt.address)
			if err == nil {
				t.Errorf("ParseGroupAddress(%q) = %v, want error", tt.address, got)
			}
		})
	}
}
