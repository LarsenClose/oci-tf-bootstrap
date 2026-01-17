package discovery

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
	lim "github.com/oracle/oci-go-sdk/v65/limits"
)

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func discoverCompartments(ctx context.Context, client identity.IdentityClient, tenancyID string) ([]Compartment, error) {
	req := identity.ListCompartmentsRequest{
		CompartmentId:          &tenancyID,
		CompartmentIdInSubtree: common.Bool(true),
		LifecycleState:         identity.CompartmentLifecycleStateActive,
	}

	var compartments []Compartment
	for {
		resp, err := client.ListCompartments(ctx, req)
		if err != nil {
			return nil, err
		}

		for _, c := range resp.Items {
			compartments = append(compartments, Compartment{
				ID:          *c.Id,
				Name:        *c.Name,
				Description: safeString(c.Description),
				ParentID:    safeString(c.CompartmentId),
			})
		}

		if resp.OpcNextPage == nil {
			break
		}
		req.Page = resp.OpcNextPage
	}
	return compartments, nil
}

func discoverADs(ctx context.Context, client identity.IdentityClient, tenancyID string) ([]AvailabilityDomain, error) {
	req := identity.ListAvailabilityDomainsRequest{
		CompartmentId: &tenancyID,
	}

	resp, err := client.ListAvailabilityDomains(ctx, req)
	if err != nil {
		return nil, err
	}

	var ads []AvailabilityDomain
	for _, ad := range resp.Items {
		fds, _ := discoverFaultDomains(ctx, client, tenancyID, *ad.Name)
		ads = append(ads, AvailabilityDomain{
			ID:           safeString(ad.Id),
			Name:         *ad.Name,
			FaultDomains: fds,
		})
	}
	return ads, nil
}

func discoverFaultDomains(ctx context.Context, client identity.IdentityClient, tenancyID, adName string) ([]string, error) {
	req := identity.ListFaultDomainsRequest{
		CompartmentId:      &tenancyID,
		AvailabilityDomain: &adName,
	}

	resp, err := client.ListFaultDomains(ctx, req)
	if err != nil {
		return nil, err
	}

	var fds []string
	for _, fd := range resp.Items {
		fds = append(fds, *fd.Name)
	}
	return fds, nil
}

func discoverShapes(ctx context.Context, client core.ComputeClient, compartmentID string) ([]Shape, error) {
	req := core.ListShapesRequest{
		CompartmentId: &compartmentID,
	}

	var shapes []Shape
	seen := make(map[string]bool)

	for {
		resp, err := client.ListShapes(ctx, req)
		if err != nil {
			return nil, err
		}

		for _, s := range resp.Items {
			if seen[*s.Shape] {
				continue
			}
			seen[*s.Shape] = true

			shape := Shape{
				Name:          *s.Shape,
				ProcessorDesc: safeString(s.ProcessorDescription),
			}
			if s.Ocpus != nil {
				shape.OCPUs = *s.Ocpus
			}
			if s.MemoryInGBs != nil {
				shape.MemoryGB = *s.MemoryInGBs
			}
			if s.OcpuOptions != nil {
				shape.IsFlexible = true
				if s.OcpuOptions.Max != nil {
					shape.MaxOCPUs = *s.OcpuOptions.Max
				}
			}
			if s.MemoryOptions != nil && s.MemoryOptions.MaxInGBs != nil {
				shape.MaxMemoryGB = *s.MemoryOptions.MaxInGBs
			}
			shapes = append(shapes, shape)
		}

		if resp.OpcNextPage == nil {
			break
		}
		req.Page = resp.OpcNextPage
	}
	return shapes, nil
}

func discoverImages(ctx context.Context, client core.ComputeClient, compartmentID string) ([]Image, error) {
	osList := []string{"Oracle Linux", "Canonical Ubuntu", "CentOS", "Windows"}
	var images []Image

	for _, osName := range osList {
		req := core.ListImagesRequest{
			CompartmentId:   &compartmentID,
			OperatingSystem: &osName,
			SortBy:          core.ListImagesSortByTimecreated,
			SortOrder:       core.ListImagesSortOrderDesc,
		}

		seenVersions := make(map[string]bool)

		for {
			resp, err := client.ListImages(ctx, req)
			if err != nil {
				break // Skip this OS on error, continue to next
			}

			for _, img := range resp.Items {
				version := safeString(img.OperatingSystemVersion)
				key := osName + "-" + version
				if seenVersions[key] {
					continue
				}
				seenVersions[key] = true

				image := Image{
					ID:          *img.Id,
					DisplayName: safeString(img.DisplayName),
					OS:          safeString(img.OperatingSystem),
					OSVersion:   version,
				}
				if img.TimeCreated != nil {
					image.TimeCreated = img.TimeCreated.String()
				}
				if img.SizeInMBs != nil {
					image.SizeGB = float64(*img.SizeInMBs) / 1024.0
				}
				images = append(images, image)
			}

			if resp.OpcNextPage == nil {
				break
			}
			req.Page = resp.OpcNextPage
		}
	}
	return images, nil
}

func discoverVCNs(ctx context.Context, client core.VirtualNetworkClient, compartmentID string) ([]VCN, error) {
	req := core.ListVcnsRequest{
		CompartmentId: &compartmentID,
	}

	var vcns []VCN
	for {
		resp, err := client.ListVcns(ctx, req)
		if err != nil {
			return nil, err
		}

		for _, v := range resp.Items {
			vcn := VCN{
				ID:            *v.Id,
				DisplayName:   safeString(v.DisplayName),
				CIDRBlock:     safeString(v.CidrBlock),
				CompartmentID: *v.CompartmentId,
				DNSLabel:      safeString(v.DnsLabel),
			}

			// Discover subnets
			subnets, err := discoverSubnets(ctx, client, *v.CompartmentId, *v.Id)
			if err != nil {
				fmt.Printf("    ⚠ Could not list subnets for VCN %s: %v\n", safeString(v.DisplayName), err)
			} else {
				vcn.Subnets = subnets
			}

			// Discover security lists
			secLists, err := discoverSecurityLists(ctx, client, *v.CompartmentId, *v.Id)
			if err != nil {
				fmt.Printf("    ⚠ Could not list security lists for VCN %s: %v\n", safeString(v.DisplayName), err)
			} else {
				vcn.SecurityLists = secLists
			}

			// Discover route tables
			routeTables, err := discoverRouteTables(ctx, client, *v.CompartmentId, *v.Id)
			if err != nil {
				fmt.Printf("    ⚠ Could not list route tables for VCN %s: %v\n", safeString(v.DisplayName), err)
			} else {
				vcn.RouteTables = routeTables
			}

			// Discover internet gateway
			igw, err := discoverInternetGateway(ctx, client, *v.CompartmentId, *v.Id)
			if err != nil {
				fmt.Printf("    ⚠ Could not list internet gateway for VCN %s: %v\n", safeString(v.DisplayName), err)
			} else {
				vcn.InternetGateway = igw
			}

			// Discover NAT gateway
			nat, err := discoverNATGateway(ctx, client, *v.CompartmentId, *v.Id)
			if err != nil {
				fmt.Printf("    ⚠ Could not list NAT gateway for VCN %s: %v\n", safeString(v.DisplayName), err)
			} else {
				vcn.NATGateway = nat
			}

			vcns = append(vcns, vcn)
		}

		if resp.OpcNextPage == nil {
			break
		}
		req.Page = resp.OpcNextPage
	}
	return vcns, nil
}

func discoverSubnets(ctx context.Context, client core.VirtualNetworkClient, compartmentID, vcnID string) ([]Subnet, error) {
	req := core.ListSubnetsRequest{
		CompartmentId: &compartmentID,
		VcnId:         &vcnID,
	}

	var subnets []Subnet
	resp, err := client.ListSubnets(ctx, req)
	if err != nil {
		return nil, err
	}

	for _, s := range resp.Items {
		subnets = append(subnets, Subnet{
			ID:                 *s.Id,
			DisplayName:        safeString(s.DisplayName),
			CIDRBlock:          safeString(s.CidrBlock),
			AvailabilityDomain: safeString(s.AvailabilityDomain),
			IsPublic:           !*s.ProhibitPublicIpOnVnic,
			DNSLabel:           safeString(s.DnsLabel),
		})
	}
	return subnets, nil
}

func discoverSecurityLists(ctx context.Context, client core.VirtualNetworkClient, compartmentID, vcnID string) ([]SecurityList, error) {
	req := core.ListSecurityListsRequest{
		CompartmentId: &compartmentID,
		VcnId:         &vcnID,
	}

	var securityLists []SecurityList
	for {
		resp, err := client.ListSecurityLists(ctx, req)
		if err != nil {
			return nil, err
		}

		for _, sl := range resp.Items {
			secList := SecurityList{
				ID:          *sl.Id,
				DisplayName: safeString(sl.DisplayName),
			}

			for _, rule := range sl.IngressSecurityRules {
				secRule := SecurityRule{
					Protocol: safeString(rule.Protocol),
					Source:   safeString(rule.Source),
				}
				if rule.TcpOptions != nil && rule.TcpOptions.DestinationPortRange != nil {
					secRule.PortMin = *rule.TcpOptions.DestinationPortRange.Min
					secRule.PortMax = *rule.TcpOptions.DestinationPortRange.Max
				}
				if rule.UdpOptions != nil && rule.UdpOptions.DestinationPortRange != nil {
					secRule.PortMin = *rule.UdpOptions.DestinationPortRange.Min
					secRule.PortMax = *rule.UdpOptions.DestinationPortRange.Max
				}
				secRule.Description = safeString(rule.Description)
				secList.IngressRules = append(secList.IngressRules, secRule)
			}

			for _, rule := range sl.EgressSecurityRules {
				secRule := SecurityRule{
					Protocol:    safeString(rule.Protocol),
					Destination: safeString(rule.Destination),
				}
				if rule.TcpOptions != nil && rule.TcpOptions.DestinationPortRange != nil {
					secRule.PortMin = *rule.TcpOptions.DestinationPortRange.Min
					secRule.PortMax = *rule.TcpOptions.DestinationPortRange.Max
				}
				if rule.UdpOptions != nil && rule.UdpOptions.DestinationPortRange != nil {
					secRule.PortMin = *rule.UdpOptions.DestinationPortRange.Min
					secRule.PortMax = *rule.UdpOptions.DestinationPortRange.Max
				}
				secRule.Description = safeString(rule.Description)
				secList.EgressRules = append(secList.EgressRules, secRule)
			}

			securityLists = append(securityLists, secList)
		}

		if resp.OpcNextPage == nil {
			break
		}
		req.Page = resp.OpcNextPage
	}
	return securityLists, nil
}

func discoverRouteTables(ctx context.Context, client core.VirtualNetworkClient, compartmentID, vcnID string) ([]RouteTable, error) {
	req := core.ListRouteTablesRequest{
		CompartmentId: &compartmentID,
		VcnId:         &vcnID,
	}

	var routeTables []RouteTable
	for {
		resp, err := client.ListRouteTables(ctx, req)
		if err != nil {
			return nil, err
		}

		for _, rt := range resp.Items {
			routeTable := RouteTable{
				ID:          *rt.Id,
				DisplayName: safeString(rt.DisplayName),
			}

			for _, rule := range rt.RouteRules {
				routeRule := RouteRule{
					Destination:     safeString(rule.Destination),
					DestinationType: string(rule.DestinationType),
					NetworkEntityID: safeString(rule.NetworkEntityId),
					Description:     safeString(rule.Description),
				}
				routeTable.Routes = append(routeTable.Routes, routeRule)
			}

			routeTables = append(routeTables, routeTable)
		}

		if resp.OpcNextPage == nil {
			break
		}
		req.Page = resp.OpcNextPage
	}
	return routeTables, nil
}

func discoverInternetGateway(ctx context.Context, client core.VirtualNetworkClient, compartmentID, vcnID string) (*InternetGateway, error) {
	req := core.ListInternetGatewaysRequest{
		CompartmentId: &compartmentID,
		VcnId:         &vcnID,
	}

	resp, err := client.ListInternetGateways(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Items) == 0 {
		return nil, nil
	}

	igw := resp.Items[0]
	return &InternetGateway{
		ID:          *igw.Id,
		DisplayName: safeString(igw.DisplayName),
		IsEnabled:   *igw.IsEnabled,
	}, nil
}

func discoverNATGateway(ctx context.Context, client core.VirtualNetworkClient, compartmentID, vcnID string) (*NATGateway, error) {
	req := core.ListNatGatewaysRequest{
		CompartmentId: &compartmentID,
		VcnId:         &vcnID,
	}

	resp, err := client.ListNatGateways(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Items) == 0 {
		return nil, nil
	}

	nat := resp.Items[0]
	return &NATGateway{
		ID:           *nat.Id,
		DisplayName:  safeString(nat.DisplayName),
		PublicIP:     safeString(nat.NatIp),
		BlockTraffic: *nat.BlockTraffic,
	}, nil
}

func discoverBlockVolumes(ctx context.Context, client core.BlockstorageClient, compartmentID string) ([]BlockVolume, error) {
	req := core.ListVolumesRequest{
		CompartmentId: &compartmentID,
	}

	var volumes []BlockVolume
	for {
		resp, err := client.ListVolumes(ctx, req)
		if err != nil {
			return nil, err
		}

		for _, v := range resp.Items {
			vol := BlockVolume{
				ID:                 *v.Id,
				DisplayName:        safeString(v.DisplayName),
				AvailabilityDomain: safeString(v.AvailabilityDomain),
			}
			if v.SizeInGBs != nil {
				vol.SizeGB = *v.SizeInGBs
			}
			if v.VpusPerGB != nil {
				vol.VPUsPerGB = *v.VpusPerGB
			}
			if v.IsHydrated != nil {
				vol.IsHydrated = *v.IsHydrated
			}
			volumes = append(volumes, vol)
		}

		if resp.OpcNextPage == nil {
			break
		}
		req.Page = resp.OpcNextPage
	}
	return volumes, nil
}

func discoverLimits(ctx context.Context, client lim.LimitsClient, tenancyID string) ([]ServiceLimit, error) {
	computeServices := []string{"compute", "compute-core"}
	var limits []ServiceLimit

	for _, svc := range computeServices {
		req := lim.ListLimitValuesRequest{
			CompartmentId: &tenancyID,
			ServiceName:   &svc,
		}

		resp, err := client.ListLimitValues(ctx, req)
		if err != nil {
			continue
		}

		for _, l := range resp.Items {
			if l.Value == nil || *l.Value == 0 {
				continue
			}
			limit := ServiceLimit{
				ServiceName: svc,
				LimitName:   safeString(l.Name),
				Value:       *l.Value,
				Scope:       string(l.ScopeType),
			}
			if l.AvailabilityDomain != nil {
				limit.Scope = *l.AvailabilityDomain
			}
			limits = append(limits, limit)
		}
	}
	return limits, nil
}

func discoverTenancy(ctx context.Context, client identity.IdentityClient, tenancyID string) (TenancyInfo, error) {
	req := identity.GetTenancyRequest{
		TenancyId: &tenancyID,
	}

	resp, err := client.GetTenancy(ctx, req)
	if err != nil {
		return TenancyInfo{ID: tenancyID}, nil
	}

	return TenancyInfo{
		ID:          *resp.Id,
		Name:        safeString(resp.Name),
		Description: safeString(resp.Description),
		HomeRegion:  safeString(resp.HomeRegionKey),
	}, nil
}
