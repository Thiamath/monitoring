/*
 * Copyright 2019 Nalej
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package asset

import (
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-inventory-go"
	"github.com/nalej/grpc-organization-go"

	"github.com/rs/zerolog/log"
)

// Mapping from Edge Controller ID to relevant Asset Selector
type SelectorMap map[string]*grpc_inventory_go.AssetSelector

func (s SelectorMap) AddAsset(asset *grpc_inventory_go.Asset) {
	ecId := asset.GetEdgeControllerId()
	assetId := asset.GetAssetId()
	selector, found := s[ecId]
	if found {
		selector.AssetIds = append(selector.AssetIds, assetId)
	} else {
		s[ecId] = &grpc_inventory_go.AssetSelector{
			OrganizationId:   asset.GetOrganizationId(),
			EdgeControllerId: ecId,
			AssetIds:         []string{assetId},
		}
	}
}

// Create a mapping based on a user-provided selector. Returns (nil, nil)
// if this function doesn't apply and the next function in the selectorFuncList
// should be called
type selectorFunc func(*grpc_inventory_go.AssetSelector) (SelectorMap, derrors.Error)

type SelectorMapFactory struct {
	funcList []selectorFunc

	assetsClient      grpc_inventory_go.AssetsClient
	controllersClient grpc_inventory_go.ControllersClient
}

func NewSelectorMapFactory(assetsClient grpc_inventory_go.AssetsClient, controllersClient grpc_inventory_go.ControllersClient) *SelectorMapFactory {
	f := &SelectorMapFactory{
		assetsClient:      assetsClient,
		controllersClient: controllersClient,
	}

	f.funcList = []selectorFunc{
		f.getAssetSelectors,
		f.getECSelectors,
		f.getFilteredSelectors,
	}

	return f
}

// SelectorMap will turn a single asset selector into a selector per
// edge controller, so we can sent the request to each edge controller
// that needs it. This is done by creating a list of assets from either
// AssetIds in the source selector or, if none are provided, by retrieving
// all Assets for an OrganizationId or EdgeControllerId. We then filter
// that list to remove the Assets that don't match the groups and labels and
// sort them out by EdgeControllerId.
// If there are no group/label filters, we can just create a selector
// for each edge controller without specific assets, as that will select
// all assets available on an Edge Controller without having to communicate
// a long list.
// We do not filter out disabled assets, as we assume that disabled assets
// ("show" is false) are not sending monitoring data anymore anyway. We still
// want to include when retrieving historic data.
func (f *SelectorMapFactory) SelectorMap(selector *grpc_inventory_go.AssetSelector) (SelectorMap, derrors.Error) {
	var selectors SelectorMap
	var derr derrors.Error

	for _, fn := range f.funcList {
		selectors, derr = fn(selector)
		if derr != nil {
			return nil, derr
		}

		// Got selectors, we're done
		if selectors != nil {
			break
		}
	}

	derr = f.filterECs(selector.GetOrganizationId(), selectors)
	if derr != nil {
		return nil, derr
	}

	return selectors, nil
}

// Filter out disabled and unavailable Edge Controllers
func (f *SelectorMapFactory) filterECs(orgId string, selectors SelectorMap) derrors.Error {
	if selectors == nil || len(selectors) == 0 {
		return nil
	}

	log.Debug().Msg("filtering selectors")

	ctx, cancel := InventoryContext()
	defer cancel()

	ecList, err := f.controllersClient.List(ctx, &grpc_organization_go.OrganizationId{
		OrganizationId: orgId,
	})
	if err != nil {
		return derrors.NewUnavailableError("unable to retrieve edge controllers", err).WithParams(orgId)
	}
	for _, ec := range ecList.GetControllers() {
		ecId := ec.GetEdgeControllerId()
		lastAlive := ec.GetLastAliveTimestamp()

		if !ec.GetShow() {
			log.Debug().Str("edge-controller-id", ecId).Msg("removing disabled edge controller from selectors")
			delete(selectors, ecId)
		} else if time.Now().UTC().Unix()-lastAlive > edgeControllerAliveTimeout {
			log.Debug().Str("edge-controller-id", ecId).Int64("last-alive", lastAlive).Msg("removing unavailable edge controller from selectors")
			delete(selectors, ec.GetEdgeControllerId())
		}
	}

	return nil
}

// If we have explicit assets, that's the minimum set we start from
func (f *SelectorMapFactory) getAssetSelectors(selector *grpc_inventory_go.AssetSelector) (SelectorMap, derrors.Error) {
	log.Debug().Interface("selector", selector).Msg("getAssetSelectors")

	orgId := selector.GetOrganizationId()
	assetIds := selector.GetAssetIds()

	// Check if this is an appropriate selectormap creator for the selector
	if len(assetIds) == 0 {
		return nil, nil
	}

	selectors := make(SelectorMap)

	for _, id := range assetIds {
		ctx, cancel := InventoryContext()
		// Calling cancel manually to avoid stacking up a lot of defers
		asset, err := f.assetsClient.Get(ctx, &grpc_inventory_go.AssetId{
			OrganizationId: orgId,
			AssetId:        id,
		})
		cancel()
		if err != nil {
			return nil, derrors.NewUnavailableError("unable to retrieve asset information", err).WithParams(id)
		}
		if selectedAsset(asset, selector) {
			selectors.AddAsset(asset)
		}
	}

	return selectors, nil
}

// Make a selector for each Edge Controller, without explicit assets
func (f *SelectorMapFactory) getECSelectors(selector *grpc_inventory_go.AssetSelector) (SelectorMap, derrors.Error) {
	log.Debug().Interface("selector", selector).Msg("getECSelectors")

	// Check if this is an appropriate selectormap creator for the selector
	if len(selector.GetLabels()) != 0 || len(selector.GetGroupIds()) != 0 {
		return nil, nil
	}

	selectors := make(SelectorMap)

	orgId := selector.GetOrganizationId()
	ecId := selector.GetEdgeControllerId()

	if ecId != "" {
		// No further selectors and ecId means we just need the
		// already existing selector
		selectors[ecId] = selector
	} else {
		// Selector for each Edge Controller in Organization
		ctx, cancel := InventoryContext()
		defer cancel()

		ecList, err := f.controllersClient.List(ctx, &grpc_organization_go.OrganizationId{
			OrganizationId: orgId,
		})
		if err != nil {
			return nil, derrors.NewUnavailableError("unable to retrieve edge controllers", err).WithParams(orgId)
		}

		for _, ec := range ecList.GetControllers() {
			id := ec.GetEdgeControllerId()
			selectors[id] = &grpc_inventory_go.AssetSelector{
				OrganizationId:   orgId,
				EdgeControllerId: id,
			}
		}
	}

	return selectors, nil
}

// If we have more filters to apply (labels, groups), we need to get
// a set of matching assets to filter. The Edge Controller doesn't
// have this info so we need to do it here and provide an exhaustive
// list of assets to query.
func (f *SelectorMapFactory) getFilteredSelectors(selector *grpc_inventory_go.AssetSelector) (SelectorMap, derrors.Error) {
	log.Debug().Interface("selector", selector).Msg("getFilteredSelectors")

	// Check if this is an appropriate selectormap creator for the selector
	if len(selector.GetLabels()) == 0 && len(selector.GetGroupIds()) == 0 {
		return nil, nil
	}

	selectors := make(SelectorMap)

	var assetList *grpc_inventory_go.AssetList
	var err error

	orgId := selector.GetOrganizationId()
	ecId := selector.GetEdgeControllerId()

	ctx, cancel := InventoryContext()
	defer cancel()

	if ecId != "" {
		// We start with all assets for an Edge Controller
		assetList, err = f.assetsClient.ListControllerAssets(ctx, &grpc_inventory_go.EdgeControllerId{
			OrganizationId:   orgId,
			EdgeControllerId: ecId,
		})
		if err != nil {
			return nil, derrors.NewUnavailableError("unable to retrieve assets for edge controller", err).WithParams(ecId)
		}
	} else {
		// If there's no Edge Controller, we start with all assets for
		// an organization
		assetList, err = f.assetsClient.List(ctx, &grpc_organization_go.OrganizationId{
			OrganizationId: orgId,
		})
		if err != nil {
			return nil, derrors.NewUnavailableError("unable to retrieve assets for organization", err).WithParams(orgId)
		}
	}

	for _, asset := range assetList.GetAssets() {
		if selectedAsset(asset, selector) {
			selectors.AddAsset(asset)
		}
	}

	return selectors, nil
}

func selectedAsset(asset *grpc_inventory_go.Asset, selector *grpc_inventory_go.AssetSelector) bool {
	// Check org
	orgId := selector.GetOrganizationId()
	if asset.GetOrganizationId() != orgId {
		return false
	}

	// Check Edge Controller
	ecId := selector.GetEdgeControllerId()
	if ecId != "" && asset.GetEdgeControllerId() != ecId {
		return false
	}

	// Check labels
	labels := selector.GetLabels()
	if labels != nil {
		assetLabels := asset.GetLabels()
		if assetLabels == nil {
			return false
		}
		for k, v := range labels {
			if assetLabels[k] != v {
				return false
			}
		}
	}

	// All checks succeeded
	return true
}
