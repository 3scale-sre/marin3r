package test

import (
	envoy_resources "github.com/3scale-sre/marin3r/api/envoy/resources"
	envoy_resources_v3 "github.com/3scale-sre/marin3r/api/envoy/resources/v3"
	xdss "github.com/3scale-sre/marin3r/internal/pkg/discoveryservice/xdss"
)

func SnapshotsAreEqual(x xdss.Snapshot, y xdss.Snapshot) bool {

	rTypesV3 := envoy_resources_v3.Mappings()
	for rType := range rTypesV3 {
		if !envoy_resources.ResourcesEqual(x.GetResources(rType), y.GetResources(rType)) {
			return false
		}
		if x.GetVersion(rType) != y.GetVersion(rType) {
			return false
		}
	}
	return true
}
