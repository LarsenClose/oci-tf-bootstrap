package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oracle/oci-go-sdk/v65/containerengine"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
	lim "github.com/oracle/oci-go-sdk/v65/limits"

	"github.com/larsenclose/oci-tf-bootstrap/internal/discovery"
	"github.com/larsenclose/oci-tf-bootstrap/internal/renderer"
)

// --- Pointer helpers ---

func strPtr(s string) *string   { return &s }
func boolPtr(b bool) *bool      { return &b }
func intPtr(i int64) *int64     { return &i }
func f32Ptr(f float32) *float32 { return &f }

// --- Mock Identity Client ---

type mockIdentityClient struct {
	compartments   []identity.Compartment
	compartmentErr error
	ads            []identity.AvailabilityDomain
	adErr          error
	faultDomains   []identity.FaultDomain
	faultDomainErr error
	tenancy        identity.Tenancy
	tenancyErr     error
}

func (m *mockIdentityClient) ListCompartments(_ context.Context, _ identity.ListCompartmentsRequest) (identity.ListCompartmentsResponse, error) {
	if m.compartmentErr != nil {
		return identity.ListCompartmentsResponse{}, m.compartmentErr
	}
	return identity.ListCompartmentsResponse{Items: m.compartments}, nil
}

func (m *mockIdentityClient) ListAvailabilityDomains(_ context.Context, _ identity.ListAvailabilityDomainsRequest) (identity.ListAvailabilityDomainsResponse, error) {
	if m.adErr != nil {
		return identity.ListAvailabilityDomainsResponse{}, m.adErr
	}
	return identity.ListAvailabilityDomainsResponse{Items: m.ads}, nil
}

func (m *mockIdentityClient) ListFaultDomains(_ context.Context, _ identity.ListFaultDomainsRequest) (identity.ListFaultDomainsResponse, error) {
	if m.faultDomainErr != nil {
		return identity.ListFaultDomainsResponse{}, m.faultDomainErr
	}
	return identity.ListFaultDomainsResponse{Items: m.faultDomains}, nil
}

func (m *mockIdentityClient) GetTenancy(_ context.Context, _ identity.GetTenancyRequest) (identity.GetTenancyResponse, error) {
	if m.tenancyErr != nil {
		return identity.GetTenancyResponse{}, m.tenancyErr
	}
	return identity.GetTenancyResponse{Tenancy: m.tenancy}, nil
}

// --- Mock Compute Client ---

type mockComputeClient struct {
	shapes   []core.Shape
	shapeErr error
	images   []core.Image
	imageErr error
}

func (m *mockComputeClient) ListShapes(_ context.Context, _ core.ListShapesRequest) (core.ListShapesResponse, error) {
	if m.shapeErr != nil {
		return core.ListShapesResponse{}, m.shapeErr
	}
	return core.ListShapesResponse{Items: m.shapes}, nil
}

func (m *mockComputeClient) ListImages(_ context.Context, _ core.ListImagesRequest) (core.ListImagesResponse, error) {
	if m.imageErr != nil {
		return core.ListImagesResponse{}, m.imageErr
	}
	return core.ListImagesResponse{Items: m.images}, nil
}

// --- Mock VirtualNetwork Client ---

type mockVirtualNetworkClient struct {
	vcns             []core.Vcn
	vcnErr           error
	subnets          []core.Subnet
	subnetErr        error
	securityLists    []core.SecurityList
	secListErr       error
	routeTables      []core.RouteTable
	routeTableErr    error
	internetGateways []core.InternetGateway
	igwErr           error
	natGateways      []core.NatGateway
	natErr           error
}

func (m *mockVirtualNetworkClient) ListVcns(_ context.Context, _ core.ListVcnsRequest) (core.ListVcnsResponse, error) {
	if m.vcnErr != nil {
		return core.ListVcnsResponse{}, m.vcnErr
	}
	return core.ListVcnsResponse{Items: m.vcns}, nil
}

func (m *mockVirtualNetworkClient) ListSubnets(_ context.Context, _ core.ListSubnetsRequest) (core.ListSubnetsResponse, error) {
	if m.subnetErr != nil {
		return core.ListSubnetsResponse{}, m.subnetErr
	}
	return core.ListSubnetsResponse{Items: m.subnets}, nil
}

func (m *mockVirtualNetworkClient) ListSecurityLists(_ context.Context, _ core.ListSecurityListsRequest) (core.ListSecurityListsResponse, error) {
	if m.secListErr != nil {
		return core.ListSecurityListsResponse{}, m.secListErr
	}
	return core.ListSecurityListsResponse{Items: m.securityLists}, nil
}

func (m *mockVirtualNetworkClient) ListRouteTables(_ context.Context, _ core.ListRouteTablesRequest) (core.ListRouteTablesResponse, error) {
	if m.routeTableErr != nil {
		return core.ListRouteTablesResponse{}, m.routeTableErr
	}
	return core.ListRouteTablesResponse{Items: m.routeTables}, nil
}

func (m *mockVirtualNetworkClient) ListInternetGateways(_ context.Context, _ core.ListInternetGatewaysRequest) (core.ListInternetGatewaysResponse, error) {
	if m.igwErr != nil {
		return core.ListInternetGatewaysResponse{}, m.igwErr
	}
	return core.ListInternetGatewaysResponse{Items: m.internetGateways}, nil
}

func (m *mockVirtualNetworkClient) ListNatGateways(_ context.Context, _ core.ListNatGatewaysRequest) (core.ListNatGatewaysResponse, error) {
	if m.natErr != nil {
		return core.ListNatGatewaysResponse{}, m.natErr
	}
	return core.ListNatGatewaysResponse{Items: m.natGateways}, nil
}

// --- Mock Blockstorage Client ---

type mockBlockstorageClient struct {
	volumes   []core.Volume
	volumeErr error
}

func (m *mockBlockstorageClient) ListVolumes(_ context.Context, _ core.ListVolumesRequest) (core.ListVolumesResponse, error) {
	if m.volumeErr != nil {
		return core.ListVolumesResponse{}, m.volumeErr
	}
	return core.ListVolumesResponse{Items: m.volumes}, nil
}

// --- Mock Limits Client ---

type mockLimitsClient struct {
	limitValues []lim.LimitValueSummary
	limitErr    error
}

func (m *mockLimitsClient) ListLimitValues(_ context.Context, _ lim.ListLimitValuesRequest) (lim.ListLimitValuesResponse, error) {
	if m.limitErr != nil {
		return lim.ListLimitValuesResponse{}, m.limitErr
	}
	return lim.ListLimitValuesResponse{Items: m.limitValues}, nil
}

// --- Mock ContainerEngine Client ---

type mockContainerEngineClient struct {
	sources []containerengine.NodeSourceOption
	ceErr   error
}

func (m *mockContainerEngineClient) GetNodePoolOptions(_ context.Context, _ containerengine.GetNodePoolOptionsRequest) (containerengine.GetNodePoolOptionsResponse, error) {
	if m.ceErr != nil {
		return containerengine.GetNodePoolOptionsResponse{}, m.ceErr
	}
	return containerengine.GetNodePoolOptionsResponse{
		NodePoolOptions: containerengine.NodePoolOptions{
			Sources: m.sources,
		},
	}, nil
}

// Compile-time interface satisfaction checks.
var (
	_ discovery.IdentityAPI        = (*mockIdentityClient)(nil)
	_ discovery.ComputeAPI         = (*mockComputeClient)(nil)
	_ discovery.VirtualNetworkAPI  = (*mockVirtualNetworkClient)(nil)
	_ discovery.BlockstorageAPI    = (*mockBlockstorageClient)(nil)
	_ discovery.LimitsAPI          = (*mockLimitsClient)(nil)
	_ discovery.ContainerEngineAPI = (*mockContainerEngineClient)(nil)
)

// --- Client builders ---

// buildStandardClients returns mock clients with VCNs present, OKE data, and standard shapes.
func buildStandardClients() *discovery.Clients {
	return &discovery.Clients{
		Identity: &mockIdentityClient{
			compartments: []identity.Compartment{
				{Id: strPtr("comp-net"), Name: strPtr("network"), Description: strPtr("Network compartment"), CompartmentId: strPtr("tenancy-1")},
				{Id: strPtr("comp-app"), Name: strPtr("applications"), Description: strPtr("App compartment"), CompartmentId: strPtr("tenancy-1")},
			},
			ads: []identity.AvailabilityDomain{
				{Id: strPtr("ad-1"), Name: strPtr("GqIf:US-ASHBURN-AD-1")},
			},
			faultDomains: []identity.FaultDomain{
				{Name: strPtr("FAULT-DOMAIN-1")},
				{Name: strPtr("FAULT-DOMAIN-2")},
				{Name: strPtr("FAULT-DOMAIN-3")},
			},
			tenancy: identity.Tenancy{
				Id: strPtr("tenancy-1"), Name: strPtr("test-tenancy"),
				Description: strPtr("Test"), HomeRegionKey: strPtr("IAD"),
			},
		},
		Compute: &mockComputeClient{
			shapes: []core.Shape{
				{Shape: strPtr("VM.Standard.A1.Flex"), Ocpus: f32Ptr(4), MemoryInGBs: f32Ptr(24), OcpuOptions: &core.ShapeOcpuOptions{Max: f32Ptr(80)}, MemoryOptions: &core.ShapeMemoryOptions{MaxInGBs: f32Ptr(512)}},
				{Shape: strPtr("VM.Standard.E2.1.Micro"), Ocpus: f32Ptr(1), MemoryInGBs: f32Ptr(1)},
				{Shape: strPtr("VM.Standard.E4.Flex"), Ocpus: f32Ptr(1), MemoryInGBs: f32Ptr(16), OcpuOptions: &core.ShapeOcpuOptions{Max: f32Ptr(64)}},
			},
			images: []core.Image{
				{Id: strPtr("img-ubuntu"), DisplayName: strPtr("Canonical-Ubuntu-24.04-aarch64-2025.01.15-0"), OperatingSystem: strPtr("Canonical Ubuntu"), OperatingSystemVersion: strPtr("24.04 Minimal aarch64"), SizeInMBs: intPtr(2048)},
				{Id: strPtr("img-ol9"), DisplayName: strPtr("Oracle-Linux-9-2025.01.15-0"), OperatingSystem: strPtr("Oracle Linux"), OperatingSystemVersion: strPtr("9"), SizeInMBs: intPtr(4096)},
			},
		},
		VirtualNetwork: &mockVirtualNetworkClient{
			vcns: []core.Vcn{
				{Id: strPtr("vcn-1"), DisplayName: strPtr("main-vcn"), CidrBlock: strPtr("10.0.0.0/16"), CompartmentId: strPtr("tenancy-1"), DnsLabel: strPtr("main")},
			},
			subnets: []core.Subnet{
				{Id: strPtr("sub-pub"), DisplayName: strPtr("public-subnet"), CidrBlock: strPtr("10.0.0.0/24"), ProhibitPublicIpOnVnic: boolPtr(false), DnsLabel: strPtr("pub")},
				{Id: strPtr("sub-priv"), DisplayName: strPtr("private-subnet"), CidrBlock: strPtr("10.0.1.0/24"), ProhibitPublicIpOnVnic: boolPtr(true), DnsLabel: strPtr("priv")},
			},
			securityLists: []core.SecurityList{
				{Id: strPtr("sl-1"), DisplayName: strPtr("Default Security List for main-vcn")},
			},
			routeTables: []core.RouteTable{
				{Id: strPtr("rt-1"), DisplayName: strPtr("Default Route Table for main-vcn")},
			},
			internetGateways: []core.InternetGateway{
				{Id: strPtr("igw-1"), DisplayName: strPtr("internet-gateway"), IsEnabled: boolPtr(true)},
			},
			natGateways: []core.NatGateway{
				{Id: strPtr("nat-1"), DisplayName: strPtr("nat-gateway"), NatIp: strPtr("129.146.1.1"), BlockTraffic: boolPtr(false)},
			},
		},
		Blockstorage: &mockBlockstorageClient{
			volumes: []core.Volume{
				{Id: strPtr("vol-1"), DisplayName: strPtr("boot-volume"), AvailabilityDomain: strPtr("GqIf:US-ASHBURN-AD-1"), SizeInGBs: intPtr(50), VpusPerGB: intPtr(10), IsHydrated: boolPtr(true)},
			},
		},
		Limits: &mockLimitsClient{
			limitValues: []lim.LimitValueSummary{
				{Name: strPtr("vm-standard-a1-flex-count"), Value: intPtr(4), ScopeType: lim.LimitValueSummaryScopeTypeAd},
			},
		},
		ContainerEngine: &mockContainerEngineClient{
			sources: []containerengine.NodeSourceOption{
				containerengine.NodeSourceViaImageOption{
					SourceName: strPtr("Oracle-Linux-8.10-aarch64-2025.11.20-0-OKE-1.31.10-1345"),
					ImageId:    strPtr("ocid1.image.oc1..oke-aarch64-1.31.10"),
				},
				containerengine.NodeSourceViaImageOption{
					SourceName: strPtr("Oracle-Linux-8.10-2025.11.20-0-OKE-1.31.10-1345"),
					ImageId:    strPtr("ocid1.image.oc1..oke-x86-1.31.10"),
				},
			},
		},
	}
}

// buildAlwaysFreeClients returns mock clients with no VCNs (triggers network.tf).
func buildAlwaysFreeClients() *discovery.Clients {
	return &discovery.Clients{
		Identity: &mockIdentityClient{
			compartments: []identity.Compartment{
				{Id: strPtr("comp-net"), Name: strPtr("network"), Description: strPtr("Network compartment"), CompartmentId: strPtr("tenancy-1")},
			},
			ads: []identity.AvailabilityDomain{
				{Id: strPtr("ad-1"), Name: strPtr("GqIf:US-ASHBURN-AD-1")},
			},
			faultDomains: []identity.FaultDomain{
				{Name: strPtr("FAULT-DOMAIN-1")},
				{Name: strPtr("FAULT-DOMAIN-2")},
				{Name: strPtr("FAULT-DOMAIN-3")},
			},
			tenancy: identity.Tenancy{
				Id: strPtr("tenancy-1"), Name: strPtr("test-tenancy"),
				Description: strPtr("Test"), HomeRegionKey: strPtr("IAD"),
			},
		},
		Compute: &mockComputeClient{
			shapes: []core.Shape{
				{Shape: strPtr("VM.Standard.A1.Flex"), Ocpus: f32Ptr(4), MemoryInGBs: f32Ptr(24), OcpuOptions: &core.ShapeOcpuOptions{Max: f32Ptr(80)}, MemoryOptions: &core.ShapeMemoryOptions{MaxInGBs: f32Ptr(512)}},
				{Shape: strPtr("VM.Standard.E2.1.Micro"), Ocpus: f32Ptr(1), MemoryInGBs: f32Ptr(1)},
				{Shape: strPtr("VM.Standard.E4.Flex"), Ocpus: f32Ptr(1), MemoryInGBs: f32Ptr(16), OcpuOptions: &core.ShapeOcpuOptions{Max: f32Ptr(64)}},
			},
			images: []core.Image{
				{Id: strPtr("img-ubuntu"), DisplayName: strPtr("Canonical-Ubuntu-24.04-aarch64-2025.01.15-0"), OperatingSystem: strPtr("Canonical Ubuntu"), OperatingSystemVersion: strPtr("24.04 Minimal aarch64"), SizeInMBs: intPtr(2048)},
				{Id: strPtr("img-ol9"), DisplayName: strPtr("Oracle-Linux-9-2025.01.15-0"), OperatingSystem: strPtr("Oracle Linux"), OperatingSystemVersion: strPtr("9"), SizeInMBs: intPtr(4096)},
			},
		},
		VirtualNetwork: &mockVirtualNetworkClient{
			vcns: []core.Vcn{},
		},
		Blockstorage: &mockBlockstorageClient{
			volumes: []core.Volume{},
		},
		Limits: &mockLimitsClient{
			limitValues: []lim.LimitValueSummary{
				{Name: strPtr("vm-standard-a1-flex-count"), Value: intPtr(4), ScopeType: lim.LimitValueSummaryScopeTypeAd},
			},
		},
		ContainerEngine: &mockContainerEngineClient{
			sources: []containerengine.NodeSourceOption{
				containerengine.NodeSourceViaImageOption{
					SourceName: strPtr("Oracle-Linux-8.10-aarch64-2025.11.20-0-OKE-1.31.10-1345"),
					ImageId:    strPtr("ocid1.image.oc1..oke-aarch64-1.31.10"),
				},
				containerengine.NodeSourceViaImageOption{
					SourceName: strPtr("Oracle-Linux-8.10-2025.11.20-0-OKE-1.31.10-1345"),
					ImageId:    strPtr("ocid1.image.oc1..oke-x86-1.31.10"),
				},
			},
		},
	}
}

// --- Integration Tests ---

func TestFullPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping pipeline integration test in short mode")
	}

	tofuPath, err := exec.LookPath("tofu")
	if err != nil {
		t.Skip("tofu not found, skipping pipeline integration test")
	}

	// --- Phase 1: Discovery via RunWithClients ---

	dctx := &discovery.Context{
		TenancyID:     "tenancy-1",
		UserID:        "user-1",
		Region:        "us-ashburn-1",
		Profile:       "DEFAULT",
		ConfigPath:    "/tmp/fake-config",
		ConfigDir:     "/tmp",
		AlwaysFree:    false,
		OKE:           true,
		CompartmentID: "tenancy-1",
	}

	clients := buildStandardClients()
	result, err := discovery.RunWithClients(dctx, clients)
	if err != nil {
		t.Fatalf("RunWithClients failed: %v", err)
	}

	// --- Phase 2: Verify Result fields ---

	if result.CompartmentID != "tenancy-1" {
		t.Errorf("expected CompartmentID 'tenancy-1', got %q", result.CompartmentID)
	}

	if result.Tenancy.Name != "test-tenancy" {
		t.Errorf("expected Tenancy.Name 'test-tenancy', got %q", result.Tenancy.Name)
	}

	// RunWithClients overwrites HomeRegion from Context.Region
	if result.Tenancy.HomeRegion != "us-ashburn-1" {
		t.Errorf("expected Tenancy.HomeRegion 'us-ashburn-1', got %q", result.Tenancy.HomeRegion)
	}

	if len(result.Compartments) != 2 {
		t.Errorf("expected 2 compartments, got %d", len(result.Compartments))
	}

	if len(result.AvailabilityDomains) != 1 {
		t.Errorf("expected 1 AD, got %d", len(result.AvailabilityDomains))
	}
	if len(result.AvailabilityDomains) > 0 && len(result.AvailabilityDomains[0].FaultDomains) != 3 {
		t.Errorf("expected 3 fault domains, got %d", len(result.AvailabilityDomains[0].FaultDomains))
	}

	if len(result.Shapes) != 3 {
		t.Errorf("expected 3 shapes, got %d", len(result.Shapes))
	}

	// discoverImages iterates 4 OS types; the mock returns the same 2 images for every OS query.
	// Dedup key is osName (from loop) + "-" + version (from image), so each OS iteration yields
	// up to 2 unique keys. Total is up to 8.
	if len(result.Images) < 1 {
		t.Errorf("expected at least 1 image, got %d", len(result.Images))
	}

	if len(result.VCNs) != 1 {
		t.Errorf("expected 1 VCN, got %d", len(result.VCNs))
	}
	if len(result.VCNs) > 0 {
		if len(result.VCNs[0].Subnets) != 2 {
			t.Errorf("expected 2 subnets, got %d", len(result.VCNs[0].Subnets))
		}
		if result.VCNs[0].InternetGateway == nil {
			t.Error("expected internet gateway to be present")
		}
		if result.VCNs[0].NATGateway == nil {
			t.Error("expected NAT gateway to be present")
		}
	}

	if len(result.BlockVolumes) != 1 {
		t.Errorf("expected 1 block volume, got %d", len(result.BlockVolumes))
	}

	// 2 compute services * 1 non-zero limit each = 2 limits
	if len(result.Limits) < 1 {
		t.Errorf("expected at least 1 limit, got %d", len(result.Limits))
	}

	if len(result.OKEImages) != 2 {
		t.Errorf("expected 2 OKE images, got %d", len(result.OKEImages))
	}

	// --- Phase 3: Render Terraform ---

	tmpDir, err := os.MkdirTemp("", "pipeline-test-standard-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	opts := renderer.Options{AlwaysFree: false}
	if err := renderer.OutputTerraform(result, tmpDir, opts); err != nil {
		t.Fatalf("OutputTerraform failed: %v", err)
	}

	// --- Phase 4: Verify .tf files exist ---

	expectedFiles := []string{"provider.tf", "locals.tf", "data.tf", "instance_example.tf"}
	for _, fname := range expectedFiles {
		path := filepath.Join(tmpDir, fname)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s was not created", fname)
		}
	}

	// network.tf should NOT exist (VCNs were discovered)
	networkPath := filepath.Join(tmpDir, "network.tf")
	if _, err := os.Stat(networkPath); !os.IsNotExist(err) {
		t.Error("network.tf should NOT be created when VCNs already exist")
	}

	// Verify content sanity
	localsContent, err := os.ReadFile(filepath.Join(tmpDir, "locals.tf"))
	if err != nil {
		t.Fatalf("failed to read locals.tf: %v", err)
	}
	localsStr := string(localsContent)

	if !strings.Contains(localsStr, `tenancy_ocid = "tenancy-1"`) {
		t.Error("locals.tf should contain tenancy_ocid")
	}
	if !strings.Contains(localsStr, "vcn_main_vcn") {
		t.Error("locals.tf should contain VCN entry")
	}
	if !strings.Contains(localsStr, "subnet_public_subnet") {
		t.Error("locals.tf should contain public subnet entry")
	}
	if !strings.Contains(localsStr, "OKE Node Images") {
		t.Error("locals.tf should contain OKE Node Images section")
	}
	if !strings.Contains(localsStr, "oke_image_") {
		t.Error("locals.tf should contain OKE image entries")
	}

	// Verify instance_example.tf references public subnet
	instanceContent, err := os.ReadFile(filepath.Join(tmpDir, "instance_example.tf"))
	if err != nil {
		t.Fatalf("failed to read instance_example.tf: %v", err)
	}
	if !strings.Contains(string(instanceContent), "local.subnet_public_subnet") {
		t.Error("instance_example.tf should reference the public subnet")
	}

	// Verify data.tf contains OKE data source
	dataContent, err := os.ReadFile(filepath.Join(tmpDir, "data.tf"))
	if err != nil {
		t.Fatalf("failed to read data.tf: %v", err)
	}
	if !strings.Contains(string(dataContent), "oci_containerengine_node_pool_option") {
		t.Error("data.tf should contain OKE node pool option data source")
	}

	// --- Phase 5: tofu init + validate ---

	initCmd := exec.Command(tofuPath, "init")
	initCmd.Dir = tmpDir
	initOut, err := initCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("tofu init failed: %v\n%s", err, string(initOut))
	}

	valCmd := exec.Command(tofuPath, "validate")
	valCmd.Dir = tmpDir
	valOut, err := valCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("tofu validate failed: %v\n%s", err, string(valOut))
	}

	t.Logf("tofu validate passed: %s", strings.TrimSpace(string(valOut)))
}

func TestFullPipelineAlwaysFree(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping pipeline integration test in short mode")
	}

	tofuPath, err := exec.LookPath("tofu")
	if err != nil {
		t.Skip("tofu not found, skipping pipeline integration test")
	}

	// --- Phase 1: Discovery via RunWithClients ---

	dctx := &discovery.Context{
		TenancyID:     "tenancy-1",
		UserID:        "user-1",
		Region:        "us-ashburn-1",
		Profile:       "DEFAULT",
		ConfigPath:    "/tmp/fake-config",
		ConfigDir:     "/tmp",
		AlwaysFree:    true,
		OKE:           false,
		CompartmentID: "tenancy-1",
	}

	clients := buildAlwaysFreeClients()
	result, err := discovery.RunWithClients(dctx, clients)
	if err != nil {
		t.Fatalf("RunWithClients failed: %v", err)
	}

	// --- Phase 2: Verify always-free filtering ---

	// AlwaysFree mode should filter shapes to only always-free eligible ones
	for _, s := range result.Shapes {
		if !discovery.AlwaysFreeShapes[s.Name] {
			t.Errorf("shape %q should have been filtered out in always-free mode", s.Name)
		}
	}
	if len(result.Shapes) != 2 {
		t.Errorf("expected 2 always-free shapes (A1.Flex + E2.1.Micro), got %d", len(result.Shapes))
	}

	// VCNs should be empty (mock has no VCNs)
	if len(result.VCNs) != 0 {
		t.Errorf("expected 0 VCNs, got %d", len(result.VCNs))
	}

	// AlwaysFree triggers OKE discovery
	if len(result.OKEImages) != 2 {
		t.Errorf("expected 2 OKE images (AlwaysFree triggers OKE discovery), got %d", len(result.OKEImages))
	}

	// Images should be filtered for always-free compatibility
	for _, img := range result.Images {
		name := strings.ToLower(img.DisplayName)
		version := strings.ToLower(img.OSVersion)
		isARM := strings.Contains(name, "aarch64") || strings.Contains(version, "aarch64")
		isMinimal := strings.Contains(name, "minimal") || strings.Contains(version, "minimal")
		if !isARM && !isMinimal {
			t.Errorf("image %q should be ARM or minimal in always-free mode", img.DisplayName)
		}
	}

	// --- Phase 3: Render Terraform ---

	tmpDir, err := os.MkdirTemp("", "pipeline-test-alwaysfree-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	opts := renderer.Options{AlwaysFree: true}
	if err := renderer.OutputTerraform(result, tmpDir, opts); err != nil {
		t.Fatalf("OutputTerraform failed: %v", err)
	}

	// --- Phase 4: Verify .tf files exist ---

	expectedFiles := []string{"provider.tf", "locals.tf", "data.tf", "instance_example.tf", "network.tf"}
	for _, fname := range expectedFiles {
		path := filepath.Join(tmpDir, fname)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s was not created", fname)
		}
	}

	// Verify network.tf was generated (no VCNs -> bootstrap VCN)
	networkContent, err := os.ReadFile(filepath.Join(tmpDir, "network.tf"))
	if err != nil {
		t.Fatalf("failed to read network.tf: %v", err)
	}
	networkStr := string(networkContent)
	if !strings.Contains(networkStr, `resource "oci_core_vcn" "main"`) {
		t.Error("network.tf should contain VCN resource")
	}
	if !strings.Contains(networkStr, "VCN and networking resources are FREE") {
		t.Error("network.tf should contain always-free tier note")
	}

	// Verify locals.tf contains always-free header
	localsContent, err := os.ReadFile(filepath.Join(tmpDir, "locals.tf"))
	if err != nil {
		t.Fatalf("failed to read locals.tf: %v", err)
	}
	localsStr := string(localsContent)
	if !strings.Contains(localsStr, "always-free tier resources only") {
		t.Error("locals.tf should contain always-free tier header")
	}

	// Verify only always-free shapes in locals
	if strings.Contains(localsStr, "vm_standard_e4_flex") {
		t.Error("locals.tf should NOT contain non-free shape E4.Flex")
	}
	if !strings.Contains(localsStr, "vm_standard_a1_flex") {
		t.Error("locals.tf should contain A1.Flex shape")
	}
	if !strings.Contains(localsStr, "vm_standard_e2_1_micro") {
		t.Error("locals.tf should contain E2.1.Micro shape")
	}

	// Verify instance_example.tf uses always-free configuration
	instanceContent, err := os.ReadFile(filepath.Join(tmpDir, "instance_example.tf"))
	if err != nil {
		t.Fatalf("failed to read instance_example.tf: %v", err)
	}
	instanceStr := string(instanceContent)
	if !strings.Contains(instanceStr, "always_free") {
		t.Error("instance_example.tf should use 'always_free' resource name")
	}
	if !strings.Contains(instanceStr, "VM.Standard.A1.Flex") {
		t.Error("instance_example.tf should use A1.Flex shape")
	}
	// instance_example.tf should reference the bootstrap subnet from network.tf
	if !strings.Contains(instanceStr, "oci_core_subnet.public.id") {
		t.Error("instance_example.tf should reference oci_core_subnet.public.id (from generated network.tf)")
	}

	// --- Phase 5: tofu init + validate ---

	initCmd := exec.Command(tofuPath, "init")
	initCmd.Dir = tmpDir
	initOut, err := initCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("tofu init failed: %v\n%s", err, string(initOut))
	}

	valCmd := exec.Command(tofuPath, "validate")
	valCmd.Dir = tmpDir
	valOut, err := valCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("tofu validate failed: %v\n%s", err, string(valOut))
	}

	t.Logf("tofu validate (always-free) passed: %s", strings.TrimSpace(string(valOut)))
}
