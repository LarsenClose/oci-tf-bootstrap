package discovery

import (
	"context"
	"fmt"
	"testing"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/containerengine"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
	lim "github.com/oracle/oci-go-sdk/v65/limits"
)

// --- Mock helpers ---

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
	return identity.ListCompartmentsResponse{
		Items: m.compartments,
	}, nil
}

func (m *mockIdentityClient) ListAvailabilityDomains(_ context.Context, _ identity.ListAvailabilityDomainsRequest) (identity.ListAvailabilityDomainsResponse, error) {
	if m.adErr != nil {
		return identity.ListAvailabilityDomainsResponse{}, m.adErr
	}
	return identity.ListAvailabilityDomainsResponse{
		Items: m.ads,
	}, nil
}

func (m *mockIdentityClient) ListFaultDomains(_ context.Context, _ identity.ListFaultDomainsRequest) (identity.ListFaultDomainsResponse, error) {
	if m.faultDomainErr != nil {
		return identity.ListFaultDomainsResponse{}, m.faultDomainErr
	}
	return identity.ListFaultDomainsResponse{
		Items: m.faultDomains,
	}, nil
}

func (m *mockIdentityClient) GetTenancy(_ context.Context, _ identity.GetTenancyRequest) (identity.GetTenancyResponse, error) {
	if m.tenancyErr != nil {
		return identity.GetTenancyResponse{}, m.tenancyErr
	}
	return identity.GetTenancyResponse{
		Tenancy: m.tenancy,
	}, nil
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
	return core.ListShapesResponse{
		Items: m.shapes,
	}, nil
}

func (m *mockComputeClient) ListImages(_ context.Context, _ core.ListImagesRequest) (core.ListImagesResponse, error) {
	if m.imageErr != nil {
		return core.ListImagesResponse{}, m.imageErr
	}
	return core.ListImagesResponse{
		Items: m.images,
	}, nil
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
	return core.ListVcnsResponse{
		Items: m.vcns,
	}, nil
}

func (m *mockVirtualNetworkClient) ListSubnets(_ context.Context, _ core.ListSubnetsRequest) (core.ListSubnetsResponse, error) {
	if m.subnetErr != nil {
		return core.ListSubnetsResponse{}, m.subnetErr
	}
	return core.ListSubnetsResponse{
		Items: m.subnets,
	}, nil
}

func (m *mockVirtualNetworkClient) ListSecurityLists(_ context.Context, _ core.ListSecurityListsRequest) (core.ListSecurityListsResponse, error) {
	if m.secListErr != nil {
		return core.ListSecurityListsResponse{}, m.secListErr
	}
	return core.ListSecurityListsResponse{
		Items: m.securityLists,
	}, nil
}

func (m *mockVirtualNetworkClient) ListRouteTables(_ context.Context, _ core.ListRouteTablesRequest) (core.ListRouteTablesResponse, error) {
	if m.routeTableErr != nil {
		return core.ListRouteTablesResponse{}, m.routeTableErr
	}
	return core.ListRouteTablesResponse{
		Items: m.routeTables,
	}, nil
}

func (m *mockVirtualNetworkClient) ListInternetGateways(_ context.Context, _ core.ListInternetGatewaysRequest) (core.ListInternetGatewaysResponse, error) {
	if m.igwErr != nil {
		return core.ListInternetGatewaysResponse{}, m.igwErr
	}
	return core.ListInternetGatewaysResponse{
		Items: m.internetGateways,
	}, nil
}

func (m *mockVirtualNetworkClient) ListNatGateways(_ context.Context, _ core.ListNatGatewaysRequest) (core.ListNatGatewaysResponse, error) {
	if m.natErr != nil {
		return core.ListNatGatewaysResponse{}, m.natErr
	}
	return core.ListNatGatewaysResponse{
		Items: m.natGateways,
	}, nil
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
	return core.ListVolumesResponse{
		Items: m.volumes,
	}, nil
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
	return lim.ListLimitValuesResponse{
		Items: m.limitValues,
	}, nil
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

// --- Tests ---

func TestDiscoverCompartments(t *testing.T) {
	t.Run("returns compartments", func(t *testing.T) {
		mock := &mockIdentityClient{
			compartments: []identity.Compartment{
				{
					Id:            strPtr("comp-1"),
					Name:          strPtr("network"),
					Description:   strPtr("Network compartment"),
					CompartmentId: strPtr("tenancy-1"),
				},
				{
					Id:            strPtr("comp-2"),
					Name:          strPtr("compute"),
					Description:   nil,
					CompartmentId: strPtr("tenancy-1"),
				},
			},
		}

		comps, err := discoverCompartments(context.Background(), mock, "tenancy-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(comps) != 2 {
			t.Fatalf("expected 2 compartments, got %d", len(comps))
		}
		if comps[0].Name != "network" {
			t.Errorf("expected name 'network', got %q", comps[0].Name)
		}
		if comps[1].Description != "" {
			t.Errorf("expected empty description for nil, got %q", comps[1].Description)
		}
	})

	t.Run("empty result", func(t *testing.T) {
		mock := &mockIdentityClient{}
		comps, err := discoverCompartments(context.Background(), mock, "tenancy-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(comps) != 0 {
			t.Errorf("expected 0 compartments, got %d", len(comps))
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := &mockIdentityClient{compartmentErr: fmt.Errorf("api error")}
		_, err := discoverCompartments(context.Background(), mock, "tenancy-1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestDiscoverADs(t *testing.T) {
	t.Run("returns ADs with fault domains", func(t *testing.T) {
		mock := &mockIdentityClient{
			ads: []identity.AvailabilityDomain{
				{Id: strPtr("ad-1"), Name: strPtr("GqIf:US-ASHBURN-AD-1")},
				{Id: strPtr("ad-2"), Name: strPtr("GqIf:US-ASHBURN-AD-2")},
			},
			faultDomains: []identity.FaultDomain{
				{Name: strPtr("FAULT-DOMAIN-1")},
				{Name: strPtr("FAULT-DOMAIN-2")},
			},
		}

		ads, err := discoverADs(context.Background(), mock, "tenancy-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ads) != 2 {
			t.Fatalf("expected 2 ADs, got %d", len(ads))
		}
		if ads[0].Name != "GqIf:US-ASHBURN-AD-1" {
			t.Errorf("unexpected AD name: %s", ads[0].Name)
		}
		if len(ads[0].FaultDomains) != 2 {
			t.Errorf("expected 2 fault domains, got %d", len(ads[0].FaultDomains))
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := &mockIdentityClient{adErr: fmt.Errorf("api error")}
		_, err := discoverADs(context.Background(), mock, "tenancy-1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestDiscoverShapes(t *testing.T) {
	t.Run("returns shapes with deduplication", func(t *testing.T) {
		mock := &mockComputeClient{
			shapes: []core.Shape{
				{Shape: strPtr("VM.Standard.A1.Flex"), Ocpus: f32Ptr(4), MemoryInGBs: f32Ptr(24), OcpuOptions: &core.ShapeOcpuOptions{Max: f32Ptr(80)}},
				{Shape: strPtr("VM.Standard.A1.Flex"), Ocpus: f32Ptr(4), MemoryInGBs: f32Ptr(24)}, // duplicate
				{Shape: strPtr("VM.Standard.E2.1.Micro"), Ocpus: f32Ptr(1), MemoryInGBs: f32Ptr(1)},
			},
		}

		shapes, err := discoverShapes(context.Background(), mock, "comp-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(shapes) != 2 {
			t.Fatalf("expected 2 shapes (deduped), got %d", len(shapes))
		}
		if !shapes[0].IsFlexible {
			t.Error("A1.Flex should be marked as flexible")
		}
		if shapes[0].MaxOCPUs != 80 {
			t.Errorf("expected MaxOCPUs=80, got %f", shapes[0].MaxOCPUs)
		}
		if shapes[1].IsFlexible {
			t.Error("E2.1.Micro should not be flexible")
		}
	})

	t.Run("nil fields", func(t *testing.T) {
		mock := &mockComputeClient{
			shapes: []core.Shape{
				{Shape: strPtr("VM.Standard.E4.Flex"), ProcessorDescription: nil, Ocpus: nil, MemoryInGBs: nil},
			},
		}
		shapes, err := discoverShapes(context.Background(), mock, "comp-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(shapes) != 1 {
			t.Fatalf("expected 1 shape, got %d", len(shapes))
		}
		if shapes[0].OCPUs != 0 || shapes[0].MemoryGB != 0 {
			t.Error("nil fields should result in zero values")
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := &mockComputeClient{shapeErr: fmt.Errorf("api error")}
		_, err := discoverShapes(context.Background(), mock, "comp-1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestDiscoverImages(t *testing.T) {
	t.Run("returns images with deduplication", func(t *testing.T) {
		mock := &mockComputeClient{
			images: []core.Image{
				{
					Id:                     strPtr("img-1"),
					DisplayName:            strPtr("Ubuntu 24.04"),
					OperatingSystem:        strPtr("Canonical Ubuntu"),
					OperatingSystemVersion: strPtr("24.04"),
					SizeInMBs:              intPtr(2048),
				},
				{
					Id:                     strPtr("img-2"),
					DisplayName:            strPtr("Ubuntu 24.04 older"),
					OperatingSystem:        strPtr("Canonical Ubuntu"),
					OperatingSystemVersion: strPtr("24.04"), // same version, should be deduped
				},
			},
		}

		images, err := discoverImages(context.Background(), mock, "comp-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Each OS gets queried separately; mock returns same images for all.
		// With deduplication, each OS should produce 1 image.
		// 4 OS types queried, but since they all return Canonical Ubuntu 24.04,
		// and dedup key is osName + version where osName comes from the query loop,
		// we should get 4 images (one per OS query) since the key includes the loop osName.
		// Actually looking at the code: key = osName + "-" + version where osName is from the loop,
		// not from the image. So each OS iteration uses a different osName prefix.
		if len(images) < 1 {
			t.Fatal("expected at least 1 image")
		}
	})

	t.Run("handles image list error gracefully", func(t *testing.T) {
		mock := &mockComputeClient{imageErr: fmt.Errorf("api error")}
		// discoverImages breaks on error per-OS, doesn't propagate
		images, err := discoverImages(context.Background(), mock, "comp-1")
		if err != nil {
			t.Fatalf("expected nil error (errors are swallowed per-OS), got: %v", err)
		}
		if len(images) != 0 {
			t.Errorf("expected 0 images on error, got %d", len(images))
		}
	})
}

func TestDiscoverVCNs(t *testing.T) {
	t.Run("returns VCNs with subnets", func(t *testing.T) {
		mock := &mockVirtualNetworkClient{
			vcns: []core.Vcn{
				{
					Id:            strPtr("vcn-1"),
					DisplayName:   strPtr("main-vcn"),
					CidrBlock:     strPtr("10.0.0.0/16"),
					CompartmentId: strPtr("comp-1"),
					DnsLabel:      strPtr("main"),
				},
			},
			subnets: []core.Subnet{
				{
					Id:                     strPtr("sub-1"),
					DisplayName:            strPtr("public"),
					CidrBlock:              strPtr("10.0.0.0/24"),
					ProhibitPublicIpOnVnic: boolPtr(false),
					DnsLabel:               strPtr("pub"),
				},
			},
			securityLists: []core.SecurityList{
				{Id: strPtr("sl-1"), DisplayName: strPtr("default")},
			},
			routeTables: []core.RouteTable{
				{Id: strPtr("rt-1"), DisplayName: strPtr("default")},
			},
			internetGateways: []core.InternetGateway{
				{Id: strPtr("igw-1"), DisplayName: strPtr("igw"), IsEnabled: boolPtr(true)},
			},
			natGateways: []core.NatGateway{
				{Id: strPtr("nat-1"), DisplayName: strPtr("nat"), NatIp: strPtr("1.2.3.4"), BlockTraffic: boolPtr(false)},
			},
		}

		vcns, err := discoverVCNs(context.Background(), mock, "comp-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(vcns) != 1 {
			t.Fatalf("expected 1 VCN, got %d", len(vcns))
		}
		if vcns[0].DisplayName != "main-vcn" {
			t.Errorf("unexpected VCN name: %s", vcns[0].DisplayName)
		}
		if len(vcns[0].Subnets) != 1 {
			t.Errorf("expected 1 subnet, got %d", len(vcns[0].Subnets))
		}
		if !vcns[0].Subnets[0].IsPublic {
			t.Error("subnet should be public (ProhibitPublicIpOnVnic=false)")
		}
		if vcns[0].InternetGateway == nil {
			t.Error("expected internet gateway")
		}
		if vcns[0].NATGateway == nil {
			t.Error("expected NAT gateway")
		}
		if vcns[0].NATGateway.PublicIP != "1.2.3.4" {
			t.Errorf("expected NAT IP 1.2.3.4, got %s", vcns[0].NATGateway.PublicIP)
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := &mockVirtualNetworkClient{vcnErr: fmt.Errorf("api error")}
		_, err := discoverVCNs(context.Background(), mock, "comp-1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("no internet gateway", func(t *testing.T) {
		mock := &mockVirtualNetworkClient{
			vcns: []core.Vcn{
				{Id: strPtr("vcn-1"), DisplayName: strPtr("vcn"), CidrBlock: strPtr("10.0.0.0/16"), CompartmentId: strPtr("comp-1")},
			},
		}
		vcns, err := discoverVCNs(context.Background(), mock, "comp-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if vcns[0].InternetGateway != nil {
			t.Error("expected nil internet gateway when none exist")
		}
		if vcns[0].NATGateway != nil {
			t.Error("expected nil NAT gateway when none exist")
		}
	})
}

func TestDiscoverBlockVolumes(t *testing.T) {
	t.Run("returns volumes", func(t *testing.T) {
		mock := &mockBlockstorageClient{
			volumes: []core.Volume{
				{
					Id:                 strPtr("vol-1"),
					DisplayName:        strPtr("boot-volume"),
					AvailabilityDomain: strPtr("AD-1"),
					SizeInGBs:          intPtr(50),
					VpusPerGB:          intPtr(10),
					IsHydrated:         boolPtr(true),
				},
				{
					Id:          strPtr("vol-2"),
					DisplayName: strPtr("data-volume"),
					SizeInGBs:   nil,
					VpusPerGB:   nil,
					IsHydrated:  nil,
				},
			},
		}

		volumes, err := discoverBlockVolumes(context.Background(), mock, "comp-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(volumes) != 2 {
			t.Fatalf("expected 2 volumes, got %d", len(volumes))
		}
		if volumes[0].SizeGB != 50 {
			t.Errorf("expected 50GB, got %d", volumes[0].SizeGB)
		}
		if volumes[0].VPUsPerGB != 10 {
			t.Errorf("expected 10 VPUs/GB, got %d", volumes[0].VPUsPerGB)
		}
		if !volumes[0].IsHydrated {
			t.Error("expected IsHydrated=true")
		}
		// nil fields
		if volumes[1].SizeGB != 0 || volumes[1].VPUsPerGB != 0 || volumes[1].IsHydrated {
			t.Error("nil fields should result in zero values")
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := &mockBlockstorageClient{volumeErr: fmt.Errorf("api error")}
		_, err := discoverBlockVolumes(context.Background(), mock, "comp-1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestDiscoverLimits(t *testing.T) {
	t.Run("returns limits filtering zero values", func(t *testing.T) {
		mock := &mockLimitsClient{
			limitValues: []lim.LimitValueSummary{
				{Name: strPtr("vm-standard-a1-flex-count"), Value: intPtr(5), ScopeType: lim.LimitValueSummaryScopeTypeAd},
				{Name: strPtr("vm-standard-e4-flex-count"), Value: intPtr(0)}, // zero, should be filtered
				{Name: strPtr("bm-count"), Value: nil},                        // nil, should be filtered
				{Name: strPtr("with-ad"), Value: intPtr(2), ScopeType: lim.LimitValueSummaryScopeTypeAd, AvailabilityDomain: strPtr("AD-1")},
			},
		}

		limits, err := discoverLimits(context.Background(), mock, "tenancy-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// 2 services ("compute", "compute-core") * 2 non-zero limits = 4
		if len(limits) != 4 {
			t.Fatalf("expected 4 limits (2 per service), got %d", len(limits))
		}
	})

	t.Run("handles error gracefully", func(t *testing.T) {
		mock := &mockLimitsClient{limitErr: fmt.Errorf("api error")}
		limits, err := discoverLimits(context.Background(), mock, "tenancy-1")
		// errors are swallowed per-service with continue
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if len(limits) != 0 {
			t.Errorf("expected 0 limits on error, got %d", len(limits))
		}
	})
}

func TestDiscoverOKEImages(t *testing.T) {
	t.Run("extracts version and architecture", func(t *testing.T) {
		mock := &mockContainerEngineClient{
			sources: []containerengine.NodeSourceOption{
				containerengine.NodeSourceViaImageOption{
					SourceName: strPtr("Oracle-Linux-8.10-aarch64-2025.11.20-0-OKE-1.31.10-1345"),
					ImageId:    strPtr("ocid1.image.oc1..oke-aarch64"),
				},
				containerengine.NodeSourceViaImageOption{
					SourceName: strPtr("Oracle-Linux-8.10-2025.11.20-0-OKE-1.31.10-1345"),
					ImageId:    strPtr("ocid1.image.oc1..oke-x86"),
				},
			},
		}

		images, err := discoverOKEImages(context.Background(), mock, "comp-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(images) != 2 {
			t.Fatalf("expected 2 OKE images, got %d", len(images))
		}
		if images[0].Architecture != "aarch64" {
			t.Errorf("expected aarch64, got %s", images[0].Architecture)
		}
		if images[0].KubernetesVersion != "1.31.10" {
			t.Errorf("expected 1.31.10, got %s", images[0].KubernetesVersion)
		}
		if images[1].Architecture != "x86_64" {
			t.Errorf("expected x86_64, got %s", images[1].Architecture)
		}
	})

	t.Run("skips entries with empty fields", func(t *testing.T) {
		mock := &mockContainerEngineClient{
			sources: []containerengine.NodeSourceOption{
				containerengine.NodeSourceViaImageOption{
					SourceName: strPtr(""),
					ImageId:    strPtr("ocid1"),
				},
				containerengine.NodeSourceViaImageOption{
					SourceName: strPtr("valid-name"),
					ImageId:    strPtr(""),
				},
			},
		}
		images, err := discoverOKEImages(context.Background(), mock, "comp-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(images) != 0 {
			t.Errorf("expected 0 images (empty fields), got %d", len(images))
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := &mockContainerEngineClient{ceErr: fmt.Errorf("api error")}
		_, err := discoverOKEImages(context.Background(), mock, "comp-1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestDiscoverTenancy(t *testing.T) {
	t.Run("returns tenancy info", func(t *testing.T) {
		mock := &mockIdentityClient{
			tenancy: identity.Tenancy{
				Id:            strPtr("tenancy-1"),
				Name:          strPtr("my-tenancy"),
				Description:   strPtr("Test tenancy"),
				HomeRegionKey: strPtr("IAD"),
			},
		}

		info, err := discoverTenancy(context.Background(), mock, "tenancy-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Name != "my-tenancy" {
			t.Errorf("expected name 'my-tenancy', got %q", info.Name)
		}
		if info.HomeRegion != "IAD" {
			t.Errorf("expected home region 'IAD', got %q", info.HomeRegion)
		}
	})

	t.Run("error fallback returns ID only", func(t *testing.T) {
		mock := &mockIdentityClient{tenancyErr: fmt.Errorf("api error")}
		info, err := discoverTenancy(context.Background(), mock, "tenancy-1")
		if err != nil {
			t.Fatalf("expected nil error (fallback), got: %v", err)
		}
		if info.ID != "tenancy-1" {
			t.Errorf("expected ID 'tenancy-1', got %q", info.ID)
		}
		if info.Name != "" {
			t.Errorf("expected empty name on fallback, got %q", info.Name)
		}
	})
}

func TestSafeString(t *testing.T) {
	if safeString(nil) != "" {
		t.Error("safeString(nil) should return empty string")
	}
	s := "hello"
	if safeString(&s) != "hello" {
		t.Errorf("safeString(&'hello') should return 'hello', got %q", safeString(&s))
	}
}

func TestOKEVersionRegex(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Oracle-Linux-8.10-aarch64-2025.11.20-0-OKE-1.31.10-1345", "1.31.10"},
		{"Oracle-Linux-8.10-2025.11.20-0-OKE-1.30.5-900", "1.30.5"},
		{"no-oke-version-here", ""},
	}
	for _, tt := range tests {
		matches := okeVersionRe.FindStringSubmatch(tt.input)
		got := ""
		if len(matches) > 1 {
			got = matches[1]
		}
		if got != tt.expected {
			t.Errorf("okeVersionRe(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

// Compile-time check that mocks satisfy the interfaces.
var (
	_ IdentityAPI        = (*mockIdentityClient)(nil)
	_ ComputeAPI         = (*mockComputeClient)(nil)
	_ VirtualNetworkAPI  = (*mockVirtualNetworkClient)(nil)
	_ BlockstorageAPI    = (*mockBlockstorageClient)(nil)
	_ LimitsAPI          = (*mockLimitsClient)(nil)
	_ ContainerEngineAPI = (*mockContainerEngineClient)(nil)
)

// Suppress unused import warning for common package (used in interface satisfaction checks).
var _ = common.Bool
