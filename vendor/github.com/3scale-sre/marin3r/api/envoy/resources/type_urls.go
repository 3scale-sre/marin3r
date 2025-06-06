package envoy

import (
	"github.com/3scale-sre/marin3r/api/envoy"
	envoy_resources_v3 "github.com/3scale-sre/marin3r/api/envoy/resources/v3"
)

func TypeURL(rType envoy.Type, version envoy.APIVersion) string {
	return envoy_resources_v3.Mappings()[rType]
}
