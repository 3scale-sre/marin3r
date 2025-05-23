package generators

import (
	"fmt"
	"time"

	"github.com/3scale-sre/marin3r/api/envoy/defaults"
	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type GeneratorOptions struct {
	InstanceName                      string
	Namespace                         string
	RootCertificateNamePrefix         string
	RootCertificateCommonNamePrefix   string
	RootCertificateDuration           time.Duration
	ServerCertificateNamePrefix       string
	ServerCertificateCommonNamePrefix string
	ServerCertificateDuration         time.Duration
	ClientCertificateDuration         time.Duration
	XdsServerPort                     int32
	MetricsServerPort                 int32
	ProbePort                         int32
	ServiceType                       operatorv1alpha1.ServiceType
	DeploymentImage                   string
	DeploymentResources               corev1.ResourceRequirements
	Debug                             bool
	PodPriorityClass                  *string
	Affinity                          *corev1.Affinity
}

func (cfg *GeneratorOptions) labels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "marin3r",
		"app.kubernetes.io/managed-by": "marin3r-operator",
		"app.kubernetes.io/component":  "discovery-service",
		"app.kubernetes.io/instance":   cfg.InstanceName,
	}
}

func (cfg *GeneratorOptions) RootCertName() string {
	return fmt.Sprintf("%s-%s", cfg.RootCertificateNamePrefix, cfg.InstanceName)
}

func (cfg *GeneratorOptions) ServerCertName() string {
	return fmt.Sprintf("%s-%s", cfg.ServerCertificateNamePrefix, cfg.InstanceName)
}

func (cfg *GeneratorOptions) ClientCertName() string {
	return defaults.SidecarClientCertificate
}

func (cfg *GeneratorOptions) ResourceName() string {
	return fmt.Sprintf("%s-%s", "marin3r", cfg.InstanceName)
}
