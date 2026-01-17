package discovery

import (
	"testing"
)

func TestFilterShapesForAlwaysFree(t *testing.T) {
	tests := []struct {
		name     string
		input    []Shape
		expected []string // expected shape names in output
	}{
		{
			name:     "empty input",
			input:    []Shape{},
			expected: []string{},
		},
		{
			name: "filters to only always-free shapes",
			input: []Shape{
				{Name: "VM.Standard.A1.Flex", IsFlexible: true, OCPUs: 4, MemoryGB: 24},
				{Name: "VM.Standard.E4.Flex", IsFlexible: true, OCPUs: 64, MemoryGB: 1024},
				{Name: "VM.Standard.E2.1.Micro", IsFlexible: false, OCPUs: 1, MemoryGB: 1},
				{Name: "BM.Standard.E4.128", IsFlexible: false, OCPUs: 128, MemoryGB: 2048},
			},
			expected: []string{"VM.Standard.A1.Flex", "VM.Standard.E2.1.Micro"},
		},
		{
			name: "no always-free shapes available",
			input: []Shape{
				{Name: "VM.Standard.E4.Flex", IsFlexible: true},
				{Name: "VM.Standard3.Flex", IsFlexible: true},
			},
			expected: []string{},
		},
		{
			name: "only A1.Flex available",
			input: []Shape{
				{Name: "VM.Standard.A1.Flex", IsFlexible: true},
			},
			expected: []string{"VM.Standard.A1.Flex"},
		},
		{
			name: "only E2.1.Micro available",
			input: []Shape{
				{Name: "VM.Standard.E2.1.Micro", IsFlexible: false},
			},
			expected: []string{"VM.Standard.E2.1.Micro"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterShapesForAlwaysFree(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d shapes, got %d", len(tt.expected), len(result))
				return
			}

			for i, expectedName := range tt.expected {
				if result[i].Name != expectedName {
					t.Errorf("expected shape[%d] to be %s, got %s", i, expectedName, result[i].Name)
				}
			}
		})
	}
}

func TestFilterImagesForAlwaysFree(t *testing.T) {
	tests := []struct {
		name     string
		input    []Image
		expected int // expected number of images
		checkFn  func([]Image) bool
	}{
		{
			name:     "empty input",
			input:    []Image{},
			expected: 0,
			checkFn:  func(imgs []Image) bool { return true },
		},
		{
			name: "prioritizes aarch64 images",
			input: []Image{
				{OS: "Canonical Ubuntu", OSVersion: "24.04", DisplayName: "Ubuntu 24.04"},
				{OS: "Canonical Ubuntu", OSVersion: "24.04 Minimal aarch64", DisplayName: "Ubuntu 24.04 aarch64"},
				{OS: "Canonical Ubuntu", OSVersion: "22.04", DisplayName: "Ubuntu 22.04"},
			},
			expected: 1, // only aarch64 matches (non-aarch64 non-minimal are excluded)
			checkFn: func(imgs []Image) bool {
				// Should contain the aarch64 image
				for _, img := range imgs {
					if img.OSVersion == "24.04 Minimal aarch64" {
						return true
					}
				}
				return false
			},
		},
		{
			name: "includes minimal x86 images",
			input: []Image{
				{OS: "Canonical Ubuntu", OSVersion: "24.04 Minimal", DisplayName: "Ubuntu 24.04 Minimal"},
				{OS: "Canonical Ubuntu", OSVersion: "24.04", DisplayName: "Ubuntu 24.04"},
			},
			expected: 1, // only minimal
			checkFn: func(imgs []Image) bool {
				for _, img := range imgs {
					if img.OSVersion == "24.04 Minimal" {
						return true
					}
				}
				return false
			},
		},
		{
			name: "deduplicates by OS-version key",
			input: []Image{
				{OS: "Canonical Ubuntu", OSVersion: "24.04 Minimal aarch64", DisplayName: "First"},
				{OS: "Canonical Ubuntu", OSVersion: "24.04 Minimal aarch64", DisplayName: "Second"},
			},
			expected: 1,
			checkFn:  func(imgs []Image) bool { return true },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterImagesForAlwaysFree(tt.input)

			if len(result) != tt.expected {
				t.Errorf("expected %d images, got %d", tt.expected, len(result))
				for _, img := range result {
					t.Logf("  - %s %s", img.OS, img.OSVersion)
				}
				return
			}

			if !tt.checkFn(result) {
				t.Error("checkFn failed")
			}
		})
	}
}

func TestIsARM64Image(t *testing.T) {
	tests := []struct {
		image    Image
		expected bool
	}{
		{Image{OSVersion: "24.04 Minimal aarch64"}, true},
		{Image{OSVersion: "24.04", DisplayName: "Ubuntu-aarch64"}, true},
		{Image{OSVersion: "24.04 Minimal"}, false},
		{Image{OSVersion: "24.04"}, false},
		{Image{OSVersion: "AARCH64"}, true}, // case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.image.OSVersion, func(t *testing.T) {
			result := isARM64Image(tt.image)
			if result != tt.expected {
				t.Errorf("isARM64Image(%+v) = %v, expected %v", tt.image, result, tt.expected)
			}
		})
	}
}

func TestIsMinimalImage(t *testing.T) {
	tests := []struct {
		image    Image
		expected bool
	}{
		{Image{OSVersion: "24.04 Minimal"}, true},
		{Image{OSVersion: "24.04", DisplayName: "Ubuntu-Minimal"}, true},
		{Image{OSVersion: "24.04"}, false},
		{Image{OSVersion: "MINIMAL"}, true}, // case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.image.OSVersion, func(t *testing.T) {
			result := isMinimalImage(tt.image)
			if result != tt.expected {
				t.Errorf("isMinimalImage(%+v) = %v, expected %v", tt.image, result, tt.expected)
			}
		})
	}
}

func TestDefaultAlwaysFreeResources(t *testing.T) {
	res := DefaultAlwaysFreeResources()

	if res.A1FlexOCPUs != 4 {
		t.Errorf("expected A1FlexOCPUs = 4, got %f", res.A1FlexOCPUs)
	}
	if res.A1FlexMemoryGB != 24 {
		t.Errorf("expected A1FlexMemoryGB = 24, got %f", res.A1FlexMemoryGB)
	}
	if res.E2MicroInstances != 2 {
		t.Errorf("expected E2MicroInstances = 2, got %d", res.E2MicroInstances)
	}
	if res.BlockStorageGB != 200 {
		t.Errorf("expected BlockStorageGB = 200, got %d", res.BlockStorageGB)
	}
	if res.OutboundDataTB != 10 {
		t.Errorf("expected OutboundDataTB = 10, got %d", res.OutboundDataTB)
	}
}

func TestAlwaysFreeShapesMap(t *testing.T) {
	// Verify the map contains exactly the expected shapes
	expectedShapes := []string{"VM.Standard.A1.Flex", "VM.Standard.E2.1.Micro"}

	if len(AlwaysFreeShapes) != len(expectedShapes) {
		t.Errorf("expected %d shapes in AlwaysFreeShapes, got %d", len(expectedShapes), len(AlwaysFreeShapes))
	}

	for _, shape := range expectedShapes {
		if !AlwaysFreeShapes[shape] {
			t.Errorf("expected %s to be in AlwaysFreeShapes", shape)
		}
	}

	// Ensure no unexpected shapes
	unexpectedShapes := []string{"VM.Standard.E4.Flex", "VM.Standard3.Flex", "BM.Standard.E4.128"}
	for _, shape := range unexpectedShapes {
		if AlwaysFreeShapes[shape] {
			t.Errorf("unexpected shape %s found in AlwaysFreeShapes", shape)
		}
	}
}
