package envoy

import (
	"github.com/3scale-sre/marin3r/api/envoy"
	envoy_serializer "github.com/3scale-sre/marin3r/api/envoy/serializer"
)

func Validate(resource string, encoding envoy_serializer.Serialization, version envoy.APIVersion, rType envoy.Type) error {
	decoder := envoy_serializer.NewResourceUnmarshaller(encoding, version)
	generator := NewGenerator(version)
	res := generator.New(rType)
	if err := decoder.Unmarshal(resource, res); err != nil {
		return err
	}

	return nil
}
