package renderer

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsenclose/oci-tf-bootstrap/internal/discovery"
)

func TestOutputJSON(t *testing.T) {
	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			Name:       "test-tenancy",
			HomeRegion: "us-phoenix-1",
		},
		Shapes: []discovery.Shape{
			{Name: "VM.Standard.A1.Flex", IsFlexible: true, OCPUs: 4},
		},
	}

	var buf bytes.Buffer
	if err := OutputJSON(result, &buf); err != nil {
		t.Fatalf("OutputJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("OutputJSON produced invalid JSON: %v", err)
	}

	// Check key fields are present
	tenancy, ok := parsed["tenancy"].(map[string]interface{})
	if !ok {
		t.Fatal("expected tenancy object in JSON output")
	}
	if tenancy["id"] != "ocid1.tenancy.oc1..test" {
		t.Errorf("expected tenancy ID to be ocid1.tenancy.oc1..test, got %v", tenancy["id"])
	}
}

func TestOutputTerraform(t *testing.T) {
	// Create temp directory for output
	tmpDir, err := os.MkdirTemp("", "oci-tf-bootstrap-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			Name:       "test-tenancy",
			HomeRegion: "us-phoenix-1",
		},
		AvailabilityDomains: []discovery.AvailabilityDomain{
			{Name: "TEST:AD-1"},
			{Name: "TEST:AD-2"},
		},
		Shapes: []discovery.Shape{
			{Name: "VM.Standard.A1.Flex", IsFlexible: true, OCPUs: 4},
		},
		Images: []discovery.Image{
			{OS: "Canonical Ubuntu", OSVersion: "24.04"},
		},
	}

	opts := Options{AlwaysFree: false}
	if err := OutputTerraform(result, tmpDir, opts); err != nil {
		t.Fatalf("OutputTerraform failed: %v", err)
	}

	// Verify all expected files were created
	expectedFiles := []string{"provider.tf", "locals.tf", "data.tf", "instance_example.tf"}
	for _, fname := range expectedFiles {
		path := filepath.Join(tmpDir, fname)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s was not created", fname)
		}
	}

	// Verify provider.tf content
	providerContent, _ := os.ReadFile(filepath.Join(tmpDir, "provider.tf"))
	if !strings.Contains(string(providerContent), "us-phoenix-1") {
		t.Error("provider.tf should contain region")
	}
	if !strings.Contains(string(providerContent), "oracle/oci") {
		t.Error("provider.tf should contain oracle/oci provider")
	}

	// Verify locals.tf content
	localsContent, _ := os.ReadFile(filepath.Join(tmpDir, "locals.tf"))
	if !strings.Contains(string(localsContent), "tenancy_ocid") {
		t.Error("locals.tf should contain tenancy_ocid")
	}
	if !strings.Contains(string(localsContent), "ocid1.tenancy.oc1..test") {
		t.Error("locals.tf should contain tenancy OCID value")
	}
}

func TestOutputTerraformAlwaysFree(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "oci-tf-bootstrap-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-phoenix-1",
		},
		AvailabilityDomains: []discovery.AvailabilityDomain{
			{Name: "TEST:AD-1"},
		},
		Shapes: []discovery.Shape{
			{Name: "VM.Standard.A1.Flex", IsFlexible: true},
		},
		Images: []discovery.Image{
			{OS: "Canonical Ubuntu", OSVersion: "24.04 Minimal aarch64"},
		},
	}

	opts := Options{AlwaysFree: true}
	if err := OutputTerraform(result, tmpDir, opts); err != nil {
		t.Fatalf("OutputTerraform with AlwaysFree failed: %v", err)
	}

	// Verify locals.tf contains always-free header
	localsContent, _ := os.ReadFile(filepath.Join(tmpDir, "locals.tf"))
	localsStr := string(localsContent)

	expectedStrings := []string{
		"always-free tier resources only",
		"4 OCPUs + 24GB memory",
		"VM.Standard.E2.1.Micro",
		"Block Storage: 200GB",
		"1 Flexible Load Balancer",
		"Bastion service",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(localsStr, expected) {
			t.Errorf("locals.tf should contain %q", expected)
		}
	}

	// Verify instance_example.tf uses always-free config
	instanceContent, _ := os.ReadFile(filepath.Join(tmpDir, "instance_example.tf"))
	instanceStr := string(instanceContent)

	if !strings.Contains(instanceStr, "always_free") {
		t.Error("instance_example.tf should use 'always_free' resource name")
	}
	if !strings.Contains(instanceStr, "VM.Standard.A1.Flex") {
		t.Error("instance_example.tf should use A1.Flex shape")
	}
	if !strings.Contains(instanceStr, "ocpus         = 2") {
		t.Error("instance_example.tf should use 2 OCPUs (half of free allocation)")
	}
	if !strings.Contains(instanceStr, "memory_in_gbs = 12") {
		t.Error("instance_example.tf should use 12GB memory (half of free allocation)")
	}
}

func TestToTFName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"VM.Standard.A1.Flex", "vm_standard_a1_flex"},
		{"Canonical Ubuntu", "canonical_ubuntu"},
		{"Oracle Linux 9", "oracle_linux_9"},
		{"some-dashed-name", "some_dashed_name"},
		{"MixedCase Name", "mixedcase_name"},
		{"multiple   spaces", "multiple___spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toTFName(tt.input)
			if result != tt.expected {
				t.Errorf("toTFName(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestOutputTerraformCreatesDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "oci-tf-bootstrap-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Use a nested path that doesn't exist
	nestedPath := filepath.Join(tmpDir, "nested", "path", "terraform")

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-phoenix-1",
		},
		AvailabilityDomains: []discovery.AvailabilityDomain{{Name: "AD-1"}},
	}

	opts := Options{}
	if err := OutputTerraform(result, nestedPath, opts); err != nil {
		t.Fatalf("OutputTerraform should create nested directories: %v", err)
	}

	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("OutputTerraform should have created the nested directory")
	}
}

// Tests for network.go

func TestWriteNetworkGeneratesWhenNoVCNs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "oci-tf-bootstrap-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-phoenix-1",
		},
		VCNs: []discovery.VCN{}, // No existing VCNs
	}

	opts := Options{AlwaysFree: false}
	if err := writeNetwork(result, tmpDir, opts); err != nil {
		t.Fatalf("writeNetwork failed: %v", err)
	}

	// Verify network.tf was created
	networkPath := filepath.Join(tmpDir, "network.tf")
	content, err := os.ReadFile(networkPath)
	if err != nil {
		t.Fatalf("failed to read network.tf: %v", err)
	}

	contentStr := string(content)

	// Check for VCN resource
	if !strings.Contains(contentStr, `resource "oci_core_vcn" "main"`) {
		t.Error("network.tf should contain VCN resource")
	}

	// Check for internet gateway
	if !strings.Contains(contentStr, `resource "oci_core_internet_gateway" "main"`) {
		t.Error("network.tf should contain internet gateway resource")
	}

	// Check for route table
	if !strings.Contains(contentStr, `resource "oci_core_route_table" "public"`) {
		t.Error("network.tf should contain route table resource")
	}

	// Check for security list
	if !strings.Contains(contentStr, `resource "oci_core_security_list" "public"`) {
		t.Error("network.tf should contain security list resource")
	}

	// Check for subnet
	if !strings.Contains(contentStr, `resource "oci_core_subnet" "public"`) {
		t.Error("network.tf should contain subnet resource")
	}
}

func TestWriteNetworkSkipsWhenVCNsExist(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "oci-tf-bootstrap-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-phoenix-1",
		},
		VCNs: []discovery.VCN{
			{ID: "ocid1.vcn.oc1..existing", DisplayName: "existing-vcn"},
		},
	}

	opts := Options{}
	if err := writeNetwork(result, tmpDir, opts); err != nil {
		t.Fatalf("writeNetwork failed: %v", err)
	}

	// Verify network.tf was NOT created
	networkPath := filepath.Join(tmpDir, "network.tf")
	if _, err := os.Stat(networkPath); !os.IsNotExist(err) {
		t.Error("network.tf should NOT be created when VCNs already exist")
	}
}

func TestWriteNetworkAlwaysFreeMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "oci-tf-bootstrap-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-phoenix-1",
		},
		VCNs: []discovery.VCN{},
	}

	opts := Options{AlwaysFree: true}
	if err := writeNetwork(result, tmpDir, opts); err != nil {
		t.Fatalf("writeNetwork failed: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "network.tf"))
	contentStr := string(content)

	// Check for always-free specific comments
	if !strings.Contains(contentStr, "VCN and networking resources are FREE") {
		t.Error("network.tf should contain always-free tier note")
	}
}

// Tests for data.go

func TestWriteDataSources(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "oci-tf-bootstrap-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-phoenix-1",
		},
		AvailabilityDomains: []discovery.AvailabilityDomain{
			{Name: "TEST:AD-1"},
			{Name: "TEST:AD-2"},
		},
		Images: []discovery.Image{
			{OS: "Canonical Ubuntu", OSVersion: "24.04"},
			{OS: "Oracle Linux", OSVersion: "9"},
		},
	}

	if err := writeDataSources(result, tmpDir); err != nil {
		t.Fatalf("writeDataSources failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "data.tf"))
	if err != nil {
		t.Fatalf("failed to read data.tf: %v", err)
	}

	contentStr := string(content)

	// Check for AD data source
	if !strings.Contains(contentStr, `data "oci_identity_availability_domains" "ads"`) {
		t.Error("data.tf should contain availability domains data source")
	}

	// Check for image data sources
	if !strings.Contains(contentStr, `data "oci_core_images" "canonical_ubuntu_24_04"`) {
		t.Error("data.tf should contain Ubuntu image data source")
	}
	if !strings.Contains(contentStr, `data "oci_core_images" "oracle_linux_9"`) {
		t.Error("data.tf should contain Oracle Linux image data source")
	}

	// Check for outputs
	if !strings.Contains(contentStr, `output "availability_domains"`) {
		t.Error("data.tf should contain availability_domains output")
	}
	if !strings.Contains(contentStr, `output "latest_images"`) {
		t.Error("data.tf should contain latest_images output")
	}
}

func TestWriteDataSourcesDeduplicatesImages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "oci-tf-bootstrap-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-phoenix-1",
		},
		Images: []discovery.Image{
			{OS: "Canonical Ubuntu", OSVersion: "24.04", ID: "first"},
			{OS: "Canonical Ubuntu", OSVersion: "24.04", ID: "second"}, // Duplicate
			{OS: "Canonical Ubuntu", OSVersion: "22.04", ID: "third"},
		},
	}

	if err := writeDataSources(result, tmpDir); err != nil {
		t.Fatalf("writeDataSources failed: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "data.tf"))
	contentStr := string(content)

	// Count occurrences of the data source - should only appear once
	count := strings.Count(contentStr, `data "oci_core_images" "canonical_ubuntu_24_04"`)
	if count != 1 {
		t.Errorf("expected 1 occurrence of canonical_ubuntu_24_04 data source, got %d", count)
	}
}

// Tests for templates.go

func TestWriteProvider(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "oci-tf-bootstrap-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	result := &discovery.Result{
		Tenancy: discovery.TenancyInfo{
			ID:         "ocid1.tenancy.oc1..test",
			HomeRegion: "us-ashburn-1",
		},
	}

	if err := writeProvider(result, tmpDir); err != nil {
		t.Fatalf("writeProvider failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "provider.tf"))
	if err != nil {
		t.Fatalf("failed to read provider.tf: %v", err)
	}

	contentStr := string(content)

	// Check for terraform block
	if !strings.Contains(contentStr, "terraform {") {
		t.Error("provider.tf should contain terraform block")
	}

	// Check for required_providers
	if !strings.Contains(contentStr, "required_providers") {
		t.Error("provider.tf should contain required_providers")
	}

	// Check for OCI provider source
	if !strings.Contains(contentStr, `source  = "oracle/oci"`) {
		t.Error("provider.tf should contain oracle/oci source")
	}

	// Check for region
	if !strings.Contains(contentStr, `region = "us-ashburn-1"`) {
		t.Error("provider.tf should contain the correct region")
	}
}

func TestNameTrackerUnique(t *testing.T) {
	tracker := newNameTracker()

	// First use should return base name
	name1 := tracker.unique("MyCompartment")
	if name1 != "mycompartment" {
		t.Errorf("expected 'mycompartment', got %q", name1)
	}

	// Second use should return with suffix
	name2 := tracker.unique("MyCompartment")
	if name2 != "mycompartment_2" {
		t.Errorf("expected 'mycompartment_2', got %q", name2)
	}

	// Third use should increment suffix
	name3 := tracker.unique("MyCompartment")
	if name3 != "mycompartment_3" {
		t.Errorf("expected 'mycompartment_3', got %q", name3)
	}

	// Different name should not have suffix
	name4 := tracker.unique("OtherCompartment")
	if name4 != "othercompartment" {
		t.Errorf("expected 'othercompartment', got %q", name4)
	}
}
