package discovery

import (
	"context"

	"github.com/oracle/oci-go-sdk/v65/containerengine"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
	lim "github.com/oracle/oci-go-sdk/v65/limits"
)

// IdentityAPI abstracts the identity client methods used by discovery.
type IdentityAPI interface {
	ListCompartments(ctx context.Context, request identity.ListCompartmentsRequest) (identity.ListCompartmentsResponse, error)
	ListAvailabilityDomains(ctx context.Context, request identity.ListAvailabilityDomainsRequest) (identity.ListAvailabilityDomainsResponse, error)
	ListFaultDomains(ctx context.Context, request identity.ListFaultDomainsRequest) (identity.ListFaultDomainsResponse, error)
	GetTenancy(ctx context.Context, request identity.GetTenancyRequest) (identity.GetTenancyResponse, error)
}

// ComputeAPI abstracts the compute client methods used by discovery.
type ComputeAPI interface {
	ListShapes(ctx context.Context, request core.ListShapesRequest) (core.ListShapesResponse, error)
	ListImages(ctx context.Context, request core.ListImagesRequest) (core.ListImagesResponse, error)
}

// VirtualNetworkAPI abstracts the virtual network client methods used by discovery.
type VirtualNetworkAPI interface {
	ListVcns(ctx context.Context, request core.ListVcnsRequest) (core.ListVcnsResponse, error)
	ListSubnets(ctx context.Context, request core.ListSubnetsRequest) (core.ListSubnetsResponse, error)
	ListSecurityLists(ctx context.Context, request core.ListSecurityListsRequest) (core.ListSecurityListsResponse, error)
	ListRouteTables(ctx context.Context, request core.ListRouteTablesRequest) (core.ListRouteTablesResponse, error)
	ListInternetGateways(ctx context.Context, request core.ListInternetGatewaysRequest) (core.ListInternetGatewaysResponse, error)
	ListNatGateways(ctx context.Context, request core.ListNatGatewaysRequest) (core.ListNatGatewaysResponse, error)
}

// BlockstorageAPI abstracts the blockstorage client methods used by discovery.
type BlockstorageAPI interface {
	ListVolumes(ctx context.Context, request core.ListVolumesRequest) (core.ListVolumesResponse, error)
}

// LimitsAPI abstracts the limits client methods used by discovery.
type LimitsAPI interface {
	ListLimitValues(ctx context.Context, request lim.ListLimitValuesRequest) (lim.ListLimitValuesResponse, error)
}

// ContainerEngineAPI abstracts the container engine client methods used by discovery.
type ContainerEngineAPI interface {
	GetNodePoolOptions(ctx context.Context, request containerengine.GetNodePoolOptionsRequest) (containerengine.GetNodePoolOptionsResponse, error)
}

// Compile-time interface satisfaction checks.
var (
	_ IdentityAPI        = identity.IdentityClient{}
	_ ComputeAPI         = core.ComputeClient{}
	_ VirtualNetworkAPI  = core.VirtualNetworkClient{}
	_ BlockstorageAPI    = core.BlockstorageClient{}
	_ LimitsAPI          = lim.LimitsClient{}
	_ ContainerEngineAPI = containerengine.ContainerEngineClient{}
)
