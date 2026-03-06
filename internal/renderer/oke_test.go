package renderer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsenclose/oci-tf-bootstrap/internal/discovery"
)

func TestWriteOKEExampleMixedArchMultipleVersions(t *testing.T) {
	tmpDir := t.TempDir()

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-phoenix-1",
		},
		AvailabilityDomains: []discovery.AvailabilityDomain{
			{Name: "TEST:AD-1"},
		},
		OKEImages: []discovery.OKEImage{
			{ID: "ocid1.image.oc1..v131-arm", SourceName: "Oracle-Linux-8.10-aarch64-2025.11.20-0-OKE-1.31.10-1345", KubernetesVersion: "v1.31.10", Architecture: "aarch64"},
			{ID: "ocid1.image.oc1..v131-x86", SourceName: "Oracle-Linux-8.10-2025.11.20-0-OKE-1.31.10-1345", KubernetesVersion: "v1.31.10", Architecture: "x86_64"},
			{ID: "ocid1.image.oc1..v132-arm", SourceName: "Oracle-Linux-8.10-aarch64-2025.11.20-0-OKE-1.32.10-1345", KubernetesVersion: "v1.32.10", Architecture: "aarch64"},
			{ID: "ocid1.image.oc1..v132-x86", SourceName: "Oracle-Linux-8.10-2025.11.20-0-OKE-1.32.10-1345", KubernetesVersion: "v1.32.10", Architecture: "x86_64"},
		},
	}

	opts := Options{AlwaysFree: false}
	if err := writeOKEExample(result, tmpDir, opts); err != nil {
		t.Fatalf("writeOKEExample failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "oke_example.tf"))
	if err != nil {
		t.Fatalf("failed to read oke_example.tf: %v", err)
	}
	s := string(content)

	// Latest version (v1.32.10) should be uncommented
	if !strings.Contains(s, "Kubernetes v1.32.10 (latest)") {
		t.Error("should mark v1.32.10 as latest")
	}

	// Latest version ARM pool should be uncommented
	if !strings.Contains(s, `resource "oci_containerengine_node_pool" "arm_pool_v1_32_10"`) {
		t.Error("should contain uncommented ARM pool for v1.32.10")
	}
	if !strings.Contains(s, `resource "oci_containerengine_node_pool" "x86_pool_v1_32_10"`) {
		t.Error("should contain uncommented x86 pool for v1.32.10")
	}

	// Older version (v1.31.10) should be commented out
	if !strings.Contains(s, `# resource "oci_containerengine_node_pool" "arm_pool_v1_31_10"`) {
		t.Error("should contain commented-out ARM pool for v1.31.10")
	}
	if !strings.Contains(s, `# resource "oci_containerengine_node_pool" "x86_pool_v1_31_10"`) {
		t.Error("should contain commented-out x86 pool for v1.31.10")
	}

	// Check ARM shape
	if !strings.Contains(s, `node_shape = "VM.Standard.A1.Flex"`) {
		t.Error("ARM pool should use VM.Standard.A1.Flex shape")
	}

	// Check x86 shape
	if !strings.Contains(s, `node_shape = "VM.Standard.E4.Flex"`) {
		t.Error("x86 pool should use VM.Standard.E4.Flex shape")
	}

	// Check correct local references for latest version
	if !strings.Contains(s, "local.oke_image_v1_32_10_aarch64") {
		t.Error("should reference local.oke_image_v1_32_10_aarch64 for ARM pool")
	}
	if !strings.Contains(s, "local.oke_image_v1_32_10_x86_64") {
		t.Error("should reference local.oke_image_v1_32_10_x86_64 for x86 pool")
	}

	// Check correct local references for older version
	if !strings.Contains(s, "local.oke_image_v1_31_10_aarch64") {
		t.Error("should reference local.oke_image_v1_31_10_aarch64 for older ARM pool")
	}
	if !strings.Contains(s, "local.oke_image_v1_31_10_x86_64") {
		t.Error("should reference local.oke_image_v1_31_10_x86_64 for older x86 pool")
	}

	// Check for helpful footer about adding more pools
	if !strings.Contains(s, "Adding More Node Pools") {
		t.Error("should contain guidance about adding more node pools")
	}
}

func TestWriteOKEExampleARMOnly(t *testing.T) {
	tmpDir := t.TempDir()

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-phoenix-1",
		},
		AvailabilityDomains: []discovery.AvailabilityDomain{
			{Name: "TEST:AD-1"},
		},
		OKEImages: []discovery.OKEImage{
			{ID: "ocid1.image.oc1..v132-arm", SourceName: "Oracle-Linux-8.10-aarch64-2025.11.20-0-OKE-1.32.10-1345", KubernetesVersion: "v1.32.10", Architecture: "aarch64"},
		},
	}

	opts := Options{AlwaysFree: false}
	if err := writeOKEExample(result, tmpDir, opts); err != nil {
		t.Fatalf("writeOKEExample failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "oke_example.tf"))
	if err != nil {
		t.Fatalf("failed to read oke_example.tf: %v", err)
	}
	s := string(content)

	// Should have ARM pool
	if !strings.Contains(s, `resource "oci_containerengine_node_pool" "arm_pool_v1_32_10"`) {
		t.Error("should contain ARM pool resource")
	}
	if !strings.Contains(s, `node_shape = "VM.Standard.A1.Flex"`) {
		t.Error("ARM pool should use A1.Flex shape")
	}

	// Should NOT have x86 pool
	if strings.Contains(s, "x86_pool") {
		t.Error("should NOT contain x86 pool when only ARM images available")
	}
	if strings.Contains(s, "VM.Standard.E4.Flex") {
		t.Error("should NOT reference E4.Flex shape when only ARM images available")
	}
}

func TestWriteOKEExampleX86Only(t *testing.T) {
	tmpDir := t.TempDir()

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-phoenix-1",
		},
		AvailabilityDomains: []discovery.AvailabilityDomain{
			{Name: "TEST:AD-1"},
		},
		OKEImages: []discovery.OKEImage{
			{ID: "ocid1.image.oc1..v132-x86", SourceName: "Oracle-Linux-8.10-2025.11.20-0-OKE-1.32.10-1345", KubernetesVersion: "v1.32.10", Architecture: "x86_64"},
		},
	}

	opts := Options{AlwaysFree: false}
	if err := writeOKEExample(result, tmpDir, opts); err != nil {
		t.Fatalf("writeOKEExample failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "oke_example.tf"))
	if err != nil {
		t.Fatalf("failed to read oke_example.tf: %v", err)
	}
	s := string(content)

	// Should have x86 pool
	if !strings.Contains(s, `resource "oci_containerengine_node_pool" "x86_pool_v1_32_10"`) {
		t.Error("should contain x86 pool resource")
	}
	if !strings.Contains(s, `node_shape = "VM.Standard.E4.Flex"`) {
		t.Error("x86 pool should use E4.Flex shape")
	}

	// Should NOT have ARM pool
	if strings.Contains(s, "arm_pool") {
		t.Error("should NOT contain ARM pool when only x86 images available")
	}
	if strings.Contains(s, "VM.Standard.A1.Flex") {
		t.Error("should NOT reference A1.Flex shape when only x86 images available")
	}
}

func TestWriteOKEExampleEmptyImages(t *testing.T) {
	tmpDir := t.TempDir()

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-phoenix-1",
		},
		AvailabilityDomains: []discovery.AvailabilityDomain{
			{Name: "TEST:AD-1"},
		},
		OKEImages: []discovery.OKEImage{},
	}

	opts := Options{AlwaysFree: false}
	if err := writeOKEExample(result, tmpDir, opts); err != nil {
		t.Fatalf("writeOKEExample failed: %v", err)
	}

	// oke_example.tf should NOT be created
	okePath := filepath.Join(tmpDir, "oke_example.tf")
	if _, err := os.Stat(okePath); !os.IsNotExist(err) {
		t.Error("oke_example.tf should NOT be created when OKEImages is empty")
	}
}

func TestWriteOKEExampleNilImages(t *testing.T) {
	tmpDir := t.TempDir()

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-phoenix-1",
		},
		AvailabilityDomains: []discovery.AvailabilityDomain{
			{Name: "TEST:AD-1"},
		},
		OKEImages: nil,
	}

	opts := Options{AlwaysFree: false}
	if err := writeOKEExample(result, tmpDir, opts); err != nil {
		t.Fatalf("writeOKEExample failed: %v", err)
	}

	// oke_example.tf should NOT be created
	okePath := filepath.Join(tmpDir, "oke_example.tf")
	if _, err := os.Stat(okePath); !os.IsNotExist(err) {
		t.Error("oke_example.tf should NOT be created when OKEImages is nil")
	}
}

func TestOutputTerraformNoOKEExample(t *testing.T) {
	tmpDir := t.TempDir()

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-phoenix-1",
		},
		AvailabilityDomains: []discovery.AvailabilityDomain{
			{Name: "TEST:AD-1"},
		},
		Images: []discovery.Image{
			{OS: "Canonical Ubuntu", OSVersion: "24.04"},
		},
		OKEImages: nil,
	}

	opts := Options{AlwaysFree: false}
	if err := OutputTerraform(result, tmpDir, opts); err != nil {
		t.Fatalf("OutputTerraform failed: %v", err)
	}

	// oke_example.tf should NOT be created when there are no OKE images
	okePath := filepath.Join(tmpDir, "oke_example.tf")
	if _, err := os.Stat(okePath); !os.IsNotExist(err) {
		t.Error("oke_example.tf should NOT be created when OKEImages is nil")
	}

	// Other files should still exist
	for _, name := range []string{"provider.tf", "locals.tf", "data.tf", "instance_example.tf"} {
		if _, err := os.Stat(filepath.Join(tmpDir, name)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", name)
		}
	}
}

func TestOutputTerraformWithOKEExample(t *testing.T) {
	tmpDir := t.TempDir()

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-phoenix-1",
		},
		AvailabilityDomains: []discovery.AvailabilityDomain{
			{Name: "TEST:AD-1"},
		},
		Images: []discovery.Image{
			{OS: "Canonical Ubuntu", OSVersion: "24.04"},
		},
		OKEImages: []discovery.OKEImage{
			{ID: "ocid1.image.oc1..oke-arm", SourceName: "OKE-ARM", KubernetesVersion: "v1.32.10", Architecture: "aarch64"},
			{ID: "ocid1.image.oc1..oke-x86", SourceName: "OKE-x86", KubernetesVersion: "v1.32.10", Architecture: "x86_64"},
		},
	}

	opts := Options{AlwaysFree: false}
	if err := OutputTerraform(result, tmpDir, opts); err != nil {
		t.Fatalf("OutputTerraform failed: %v", err)
	}

	// oke_example.tf should exist
	okePath := filepath.Join(tmpDir, "oke_example.tf")
	if _, err := os.Stat(okePath); os.IsNotExist(err) {
		t.Fatal("oke_example.tf should be created when OKEImages is present")
	}
}

func TestWriteOKEExampleAlwaysFreeMode(t *testing.T) {
	tmpDir := t.TempDir()

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-phoenix-1",
		},
		AvailabilityDomains: []discovery.AvailabilityDomain{
			{Name: "TEST:AD-1"},
		},
		OKEImages: []discovery.OKEImage{
			{ID: "ocid1.image.oc1..v132-arm", SourceName: "OKE-ARM", KubernetesVersion: "v1.32.10", Architecture: "aarch64"},
			{ID: "ocid1.image.oc1..v132-x86", SourceName: "OKE-x86", KubernetesVersion: "v1.32.10", Architecture: "x86_64"},
		},
	}

	opts := Options{AlwaysFree: true}
	if err := writeOKEExample(result, tmpDir, opts); err != nil {
		t.Fatalf("writeOKEExample failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "oke_example.tf"))
	if err != nil {
		t.Fatalf("failed to read oke_example.tf: %v", err)
	}
	s := string(content)

	// Should contain always-free specific comments
	if !strings.Contains(s, "Always-Free Tier Notes") {
		t.Error("should contain Always-Free Tier Notes header")
	}
	if !strings.Contains(s, "OKE control plane is FREE") {
		t.Error("should mention OKE control plane is free")
	}
	if !strings.Contains(s, "4 OCPUs + 24GB total") {
		t.Error("should mention free tier ARM limits")
	}

	// ARM pool should have free-tier-aware sizing comments
	if !strings.Contains(s, "Half of 4 free OCPUs") {
		t.Error("ARM pool should have free tier OCPU comment")
	}
	if !strings.Contains(s, "Half of 24GB free memory") {
		t.Error("ARM pool should have free tier memory comment")
	}
	if !strings.Contains(s, "free tier limits") {
		t.Error("ARM pool should mention free tier limits in node count comment")
	}
}

func TestWriteOKEExampleStandardMode(t *testing.T) {
	tmpDir := t.TempDir()

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-phoenix-1",
		},
		AvailabilityDomains: []discovery.AvailabilityDomain{
			{Name: "TEST:AD-1"},
		},
		OKEImages: []discovery.OKEImage{
			{ID: "ocid1.image.oc1..v132-arm", SourceName: "OKE-ARM", KubernetesVersion: "v1.32.10", Architecture: "aarch64"},
		},
	}

	opts := Options{AlwaysFree: false}
	if err := writeOKEExample(result, tmpDir, opts); err != nil {
		t.Fatalf("writeOKEExample failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "oke_example.tf"))
	if err != nil {
		t.Fatalf("failed to read oke_example.tf: %v", err)
	}
	s := string(content)

	// Should NOT contain always-free specific comments
	if strings.Contains(s, "Always-Free Tier Notes") {
		t.Error("standard mode should NOT contain Always-Free Tier Notes")
	}
	if strings.Contains(s, "Half of 4 free OCPUs") {
		t.Error("standard mode should NOT contain free tier OCPU comment")
	}
}

func TestWriteOKEExampleLocalNamesMatchLocals(t *testing.T) {
	tmpDir := t.TempDir()

	okeImages := []discovery.OKEImage{
		{ID: "ocid1.image.oc1..v132-arm", SourceName: "Oracle-Linux-8.10-aarch64-2025.11.20-0-OKE-1.32.10-1345", KubernetesVersion: "v1.32.10", Architecture: "aarch64"},
		{ID: "ocid1.image.oc1..v132-x86", SourceName: "Oracle-Linux-8.10-2025.11.20-0-OKE-1.32.10-1345", KubernetesVersion: "v1.32.10", Architecture: "x86_64"},
		{ID: "ocid1.image.oc1..v131-arm", SourceName: "Oracle-Linux-8.10-aarch64-2025.11.20-0-OKE-1.31.10-1345", KubernetesVersion: "v1.31.10", Architecture: "aarch64"},
	}

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-phoenix-1",
		},
		AvailabilityDomains: []discovery.AvailabilityDomain{
			{Name: "TEST:AD-1"},
		},
		Images: []discovery.Image{
			{OS: "Canonical Ubuntu", OSVersion: "24.04"},
		},
		OKEImages: okeImages,
	}

	opts := Options{AlwaysFree: false}

	// Generate both locals.tf and oke_example.tf
	if err := OutputTerraform(result, tmpDir, opts); err != nil {
		t.Fatalf("OutputTerraform failed: %v", err)
	}

	localsContent, err := os.ReadFile(filepath.Join(tmpDir, "locals.tf"))
	if err != nil {
		t.Fatalf("failed to read locals.tf: %v", err)
	}
	okeContent, err := os.ReadFile(filepath.Join(tmpDir, "oke_example.tf"))
	if err != nil {
		t.Fatalf("failed to read oke_example.tf: %v", err)
	}

	localsStr := string(localsContent)
	okeStr := string(okeContent)

	// For each OKE image, verify the local name referenced in oke_example.tf
	// is defined in locals.tf
	expectedLocals := []string{
		"oke_image_v1_32_10_aarch64",
		"oke_image_v1_32_10_x86_64",
		"oke_image_v1_31_10_aarch64",
	}

	for _, localName := range expectedLocals {
		// Check it's defined in locals.tf (as an assignment)
		if !strings.Contains(localsStr, localName+" =") {
			t.Errorf("locals.tf should define %s", localName)
		}

		// Check it's referenced in oke_example.tf (as local.<name>)
		if !strings.Contains(okeStr, "local."+localName) {
			t.Errorf("oke_example.tf should reference local.%s", localName)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int // -1, 0, 1
	}{
		{"v1.32.10", "v1.31.10", 1},
		{"v1.31.10", "v1.32.10", -1},
		{"v1.32.10", "v1.32.10", 0},
		{"v1.34.2", "v1.33.1", 1},
		{"v1.32.10", "v1.32.2", 1},
		{"v1.9.0", "v1.10.0", -1},
		{"1.32.10", "1.31.10", 1}, // Without v prefix
		{"v2.0.0", "v1.99.99", 1}, // Major version difference
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			got := compareVersions(tt.a, tt.b)
			if (tt.want < 0 && got >= 0) || (tt.want > 0 && got <= 0) || (tt.want == 0 && got != 0) {
				t.Errorf("compareVersions(%q, %q) = %d, want sign %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestGroupOKEImagesByVersion(t *testing.T) {
	images := []discovery.OKEImage{
		{ID: "id1", KubernetesVersion: "v1.31.10", Architecture: "aarch64"},
		{ID: "id2", KubernetesVersion: "v1.31.10", Architecture: "x86_64"},
		{ID: "id3", KubernetesVersion: "v1.32.10", Architecture: "aarch64"},
		{ID: "id4", KubernetesVersion: "v1.32.10", Architecture: "x86_64"},
		{ID: "id5", KubernetesVersion: "v1.34.2", Architecture: "aarch64"},
	}

	groups := groupOKEImagesByVersion(images)

	if len(groups) != 3 {
		t.Fatalf("expected 3 version groups, got %d", len(groups))
	}

	// Should be sorted descending: v1.34.2, v1.32.10, v1.31.10
	if groups[0].version != "v1.34.2" {
		t.Errorf("first group should be v1.34.2, got %s", groups[0].version)
	}
	if groups[1].version != "v1.32.10" {
		t.Errorf("second group should be v1.32.10, got %s", groups[1].version)
	}
	if groups[2].version != "v1.31.10" {
		t.Errorf("third group should be v1.31.10, got %s", groups[2].version)
	}

	// v1.34.2 should have only ARM
	if groups[0].arm == nil {
		t.Error("v1.34.2 should have ARM image")
	}
	if groups[0].x86 != nil {
		t.Error("v1.34.2 should NOT have x86 image")
	}

	// v1.32.10 should have both
	if groups[1].arm == nil || groups[1].x86 == nil {
		t.Error("v1.32.10 should have both ARM and x86 images")
	}
}

func TestOKELocalName(t *testing.T) {
	tests := []struct {
		img      discovery.OKEImage
		expected string
	}{
		{
			discovery.OKEImage{KubernetesVersion: "v1.32.10", Architecture: "aarch64"},
			"oke_image_v1_32_10_aarch64",
		},
		{
			discovery.OKEImage{KubernetesVersion: "v1.31.10", Architecture: "x86_64"},
			"oke_image_v1_31_10_x86_64",
		},
		{
			discovery.OKEImage{KubernetesVersion: "v1.34.2", Architecture: "aarch64"},
			"oke_image_v1_34_2_aarch64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := okeLocalName(&tt.img)
			if got != tt.expected {
				t.Errorf("okeLocalName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestWriteOKEExamplePlaceholders(t *testing.T) {
	tmpDir := t.TempDir()

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-phoenix-1",
		},
		AvailabilityDomains: []discovery.AvailabilityDomain{
			{Name: "TEST:AD-1"},
		},
		OKEImages: []discovery.OKEImage{
			{ID: "ocid1.image.oc1..v132-arm", SourceName: "OKE-ARM", KubernetesVersion: "v1.32.10", Architecture: "aarch64"},
		},
	}

	opts := Options{AlwaysFree: false}
	if err := writeOKEExample(result, tmpDir, opts); err != nil {
		t.Fatalf("writeOKEExample failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "oke_example.tf"))
	if err != nil {
		t.Fatalf("failed to read oke_example.tf: %v", err)
	}
	s := string(content)

	// Should contain placeholder values with clear instructions
	if !strings.Contains(s, "PLACEHOLDER_CLUSTER_OCID") {
		t.Error("should contain PLACEHOLDER_CLUSTER_OCID")
	}
	if !strings.Contains(s, "Replace with your OKE cluster OCID") {
		t.Error("should contain instruction to replace cluster OCID")
	}
	if !strings.Contains(s, "PLACEHOLDER_SUBNET_OCID") {
		t.Error("should contain PLACEHOLDER_SUBNET_OCID")
	}
	if !strings.Contains(s, "Replace with your worker subnet OCID") {
		t.Error("should contain instruction to replace subnet OCID")
	}

	// Should contain header explaining this is a reference config
	if !strings.Contains(s, "REFERENCE configurations") {
		t.Error("should note these are reference configurations")
	}
}
