package discovery

import (
	"context"
	"fmt"
	"sync"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
	lim "github.com/oracle/oci-go-sdk/v65/limits"
	"golang.org/x/sync/errgroup"
)

func Run(ctx *Context) (*Result, error) {
	configPath := ctx.ConfigDir + "/config"
	configProvider, err := common.ConfigurationProviderFromFileWithProfile(configPath, ctx.Profile, "")
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

	result := &Result{
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
		tenancy, err := discoverTenancy(gctx, identityClient, ctx.TenancyID)
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
		comps, err := discoverCompartments(gctx, identityClient, ctx.TenancyID)
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
		ads, err := discoverADs(gctx, identityClient, ctx.TenancyID)
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
		shapes, err := discoverShapes(gctx, computeClient, ctx.TenancyID)
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
		images, err := discoverImages(gctx, computeClient, ctx.TenancyID)
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
		vcns, err := discoverVCNs(gctx, networkClient, ctx.TenancyID)
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
		limits, err := discoverLimits(gctx, limitsClient, ctx.TenancyID)
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
		volumes, err := discoverBlockVolumes(gctx, blockstorageClient, ctx.TenancyID)
		if err != nil {
			fmt.Printf("    ⚠ Block volume discovery failed (non-fatal): %v\n", err)
			return nil
		}
		mu.Lock()
		result.BlockVolumes = volumes
		mu.Unlock()
		return nil
	})

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
