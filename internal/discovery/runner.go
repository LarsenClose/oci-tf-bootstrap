package discovery

import (
	"context"
	"fmt"
	"sync"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/containerengine"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
	lim "github.com/oracle/oci-go-sdk/v65/limits"
	"golang.org/x/sync/errgroup"
)

// Clients holds the OCI API clients used during discovery.
// Callers can inject mock implementations for testing.
type Clients struct {
	Identity        IdentityAPI
	Compute         ComputeAPI
	VirtualNetwork  VirtualNetworkAPI
	Blockstorage    BlockstorageAPI
	Limits          LimitsAPI
	ContainerEngine ContainerEngineAPI
}

// Run creates concrete OCI clients from the config provider and delegates to RunWithClients.
func Run(ctx *Context) (*Result, error) {
	configProvider, err := common.ConfigurationProviderFromFileWithProfile(ctx.ConfigPath, ctx.Profile, "")
	if err != nil {
		return nil, err
	}

	identityClient, err := identity.NewIdentityClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, fmt.Errorf("identity client: %w", err)
	}

	computeClient, err := core.NewComputeClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, fmt.Errorf("compute client: %w", err)
	}

	networkClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, fmt.Errorf("network client: %w", err)
	}

	blockstorageClient, err := core.NewBlockstorageClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, fmt.Errorf("blockstorage client: %w", err)
	}

	limitsClient, err := lim.NewLimitsClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, fmt.Errorf("limits client: %w", err)
	}

	ceClient, err := containerengine.NewContainerEngineClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, fmt.Errorf("containerengine client: %w", err)
	}

	clients := &Clients{
		Identity:        identityClient,
		Compute:         computeClient,
		VirtualNetwork:  networkClient,
		Blockstorage:    blockstorageClient,
		Limits:          limitsClient,
		ContainerEngine: ceClient,
	}

	return RunWithClients(ctx, clients)
}

// RunWithClients runs the full discovery pipeline using the provided clients.
// This enables mock-based testing of the discovery orchestration.
func RunWithClients(ctx *Context, clients *Clients) (*Result, error) {
	result := &Result{
		CompartmentID: ctx.CompartmentID,
		Tenancy: TenancyInfo{
			ID:         ctx.TenancyID,
			HomeRegion: ctx.Region,
		},
	}
	var mu sync.Mutex
	g, gctx := errgroup.WithContext(context.Background())

	fmt.Println("Discovering resources...")

	g.Go(func() error {
		fmt.Println("  → Tenancy Details")
		tenancy, err := discoverTenancy(gctx, clients.Identity, ctx.TenancyID)
		if err != nil {
			return fmt.Errorf("tenancy: %w", err)
		}
		mu.Lock()
		result.Tenancy = tenancy
		result.Tenancy.HomeRegion = ctx.Region
		mu.Unlock()
		return nil
	})

	g.Go(func() error {
		fmt.Println("  → Compartments")
		comps, err := discoverCompartments(gctx, clients.Identity, ctx.TenancyID)
		if err != nil {
			return fmt.Errorf("compartments: %w", err)
		}
		mu.Lock()
		result.Compartments = comps
		mu.Unlock()
		return nil
	})

	g.Go(func() error {
		fmt.Println("  → Availability Domains")
		ads, err := discoverADs(gctx, clients.Identity, ctx.TenancyID)
		if err != nil {
			return fmt.Errorf("availability domains: %w", err)
		}
		mu.Lock()
		result.AvailabilityDomains = ads
		mu.Unlock()
		return nil
	})

	g.Go(func() error {
		fmt.Println("  → Shapes")
		shapes, err := discoverShapes(gctx, clients.Compute, ctx.CompartmentID)
		if err != nil {
			return fmt.Errorf("shapes: %w", err)
		}
		mu.Lock()
		result.Shapes = shapes
		mu.Unlock()
		return nil
	})

	g.Go(func() error {
		fmt.Println("  → Images")
		images, err := discoverImages(gctx, clients.Compute, ctx.CompartmentID)
		if err != nil {
			return fmt.Errorf("images: %w", err)
		}
		mu.Lock()
		result.Images = images
		mu.Unlock()
		return nil
	})

	g.Go(func() error {
		fmt.Println("  → VCNs")
		vcns, err := discoverVCNs(gctx, clients.VirtualNetwork, ctx.CompartmentID)
		if err != nil {
			fmt.Printf("    ⚠ VCN discovery failed (non-fatal): %v\n", err)
			return nil
		}
		mu.Lock()
		result.VCNs = vcns
		mu.Unlock()
		return nil
	})

	g.Go(func() error {
		fmt.Println("  → Service Limits")
		limits, err := discoverLimits(gctx, clients.Limits, ctx.TenancyID)
		if err != nil {
			fmt.Printf("    ⚠ Limits discovery failed (non-fatal): %v\n", err)
			return nil
		}
		mu.Lock()
		result.Limits = limits
		mu.Unlock()
		return nil
	})

	g.Go(func() error {
		fmt.Println("  → Block Volumes")
		volumes, err := discoverBlockVolumes(gctx, clients.Blockstorage, ctx.CompartmentID)
		if err != nil {
			fmt.Printf("    ⚠ Block volume discovery failed (non-fatal): %v\n", err)
			return nil
		}
		mu.Lock()
		result.BlockVolumes = volumes
		mu.Unlock()
		return nil
	})

	// Discover OKE images when explicitly requested or in always-free mode
	if ctx.AlwaysFree || ctx.OKE {
		g.Go(func() error {
			fmt.Println("  → OKE Node Images")
			okeImages, err := discoverOKEImages(gctx, clients.ContainerEngine, ctx.CompartmentID)
			if err != nil {
				fmt.Printf("    ⚠ OKE image discovery failed (non-fatal): %v\n", err)
				return nil
			}
			mu.Lock()
			result.OKEImages = okeImages
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Apply always-free filtering if requested
	if ctx.AlwaysFree {
		result.Shapes = FilterShapesForAlwaysFree(result.Shapes)
		result.Images = FilterImagesForAlwaysFree(result.Images)
	}

	return result, nil
}
