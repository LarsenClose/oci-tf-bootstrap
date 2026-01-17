package discovery

import "strings"

// AlwaysFreeShapes lists OCI shapes eligible for always-free tier
// - VM.Standard.A1.Flex: ARM-based, 4 OCPUs / 24GB total free per tenancy
// - VM.Standard.E2.1.Micro: x86, 2 instances free per tenancy
var AlwaysFreeShapes = map[string]bool{
	"VM.Standard.A1.Flex":    true,
	"VM.Standard.E2.1.Micro": true,
}

// AlwaysFreeResources documents the free tier limits
type AlwaysFreeResources struct {
	// ARM (A1.Flex): 4 OCPUs and 24GB memory total across all instances
	A1FlexOCPUs    float32
	A1FlexMemoryGB float32
	// x86 (E2.1.Micro): 2 instances total
	E2MicroInstances int
	// Block storage: 200GB total
	BlockStorageGB int
	// Outbound data: 10TB/month
	OutboundDataTB int
}

// DefaultAlwaysFreeResources returns the current OCI always-free limits
func DefaultAlwaysFreeResources() AlwaysFreeResources {
	return AlwaysFreeResources{
		A1FlexOCPUs:      4,
		A1FlexMemoryGB:   24,
		E2MicroInstances: 2,
		BlockStorageGB:   200,
		OutboundDataTB:   10,
	}
}

// FilterShapesForAlwaysFree returns only shapes eligible for always-free tier
func FilterShapesForAlwaysFree(shapes []Shape) []Shape {
	var filtered []Shape
	for _, s := range shapes {
		if AlwaysFreeShapes[s.Name] {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// FilterImagesForAlwaysFree returns images compatible with always-free shapes
// Prioritizes aarch64 images for A1.Flex (ARM) and includes x86 for E2.1.Micro
func FilterImagesForAlwaysFree(images []Image) []Image {
	var filtered []Image
	seen := make(map[string]bool)

	// First pass: collect aarch64 images (for A1.Flex ARM)
	for _, img := range images {
		if isARM64Image(img) {
			key := img.OS + "-" + img.OSVersion
			if !seen[key] {
				seen[key] = true
				filtered = append(filtered, img)
			}
		}
	}

	// Second pass: add x86 images for E2.1.Micro if not already covered
	// Only add minimal/standard x86 versions, not duplicates
	for _, img := range images {
		if !isARM64Image(img) && isMinimalImage(img) {
			key := img.OS + "-" + img.OSVersion
			if !seen[key] {
				seen[key] = true
				filtered = append(filtered, img)
			}
		}
	}

	return filtered
}

// isARM64Image checks if the image is for ARM64 architecture
func isARM64Image(img Image) bool {
	version := strings.ToLower(img.OSVersion)
	displayName := strings.ToLower(img.DisplayName)
	return strings.Contains(version, "aarch64") || strings.Contains(displayName, "aarch64")
}

// isMinimalImage checks if this is a minimal image variant (smaller, faster boot)
func isMinimalImage(img Image) bool {
	version := strings.ToLower(img.OSVersion)
	displayName := strings.ToLower(img.DisplayName)
	return strings.Contains(version, "minimal") || strings.Contains(displayName, "minimal")
}
