// Copyright 2020 rvazquez@redhat.com
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package v1

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/3scale-sre/marin3r/api/envoy/defaults"
	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	envoy_container "github.com/3scale-sre/marin3r/internal/pkg/envoy/container"
	"github.com/3scale-sre/marin3r/internal/pkg/envoy/container/shutdownmanager"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	marin3rAnnotationsDomain = "marin3r.3scale.net"

	// parameter names
	paramNodeID               = "node-id"
	paramClusterID            = "cluster-id"
	paramContainerName        = "container-name"
	paramPorts                = "ports"
	paramHostPortMapings      = "host-port-mappings"
	paramImage                = "envoy-image"
	paramConfigVolume         = "config-volume"
	paramTLSVolume            = "tls-volume"
	paramClientCertificate    = "client-certificate"
	paramEnvoyExtraArgs       = "envoy-extra-args"
	paramEnvoyAPIVersion      = "envoy-api-version"
	paramDiscoveryServiceName = "discovery-service.name"

	// Annotations to allow configuration of Envoy's admin api
	paramEnvoyAdminPort          = "admin.port"
	paramEnvoyAdminBindAddress   = "admin.bind-address"
	paramEnvoyAdminAccessLogPath = "admin.access-log-path"

	// Annotations to allow definition of container resource requests
	// and limits for CPU and Memory. For this annitations, a non-existing
	// annotation key means no request/limit is enforced. An existing
	// annotation key defined with the empty string will make the operator
	// return an error
	paramResourceRequestsCPU    = "resources.requests.cpu"
	paramResourceRequestsMemory = "resources.requests.memory"
	paramResourceLimitsCPU      = "resources.limits.cpu"
	paramResourceLimitsMemory   = "resources.limits.memory"

	// Annotations to allow configuration of the shutdown manager for
	// Envoy sidecards (to allow graceful termination of the envoy server)
	paramShtdnMgrEnabled             = "shutdown-manager.enabled"
	paramShtdnMgrServerPort          = "shutdown-manager.port"
	paramShtdnMgrImage               = "shutdown-manager.image"
	paramShtdnMgrExtraLifecycleHooks = "shutdown-manager.extra-lifecycle-hooks"
	paramShtdnMgrDrainTime           = "shutdown-manager.drain-time"
	paramShtdnMgrDrainStrategy       = "shutdown-manager.drain-strategy"

	// Annotations to allow configuration of the init manager for
	// Envoy sidecards
	paramInitMgrImage = "init-manager.image"
)

type envoySidecarConfig struct {
	generator envoy_container.ContainerConfig
}

func (esc *envoySidecarConfig) PopulateFromAnnotations(ctx context.Context, clnt client.Client, namespace string, annotations map[string]string) error {
	var err error

	esc.generator.Name = getStringParam(paramContainerName, annotations)
	esc.generator.Image = getStringParam(paramImage, annotations)

	esc.generator.Ports, err = getContainerPorts(annotations)
	if err != nil {
		return err
	}

	esc.generator.ConfigBasePath = defaults.EnvoyConfigBasePath
	esc.generator.ConfigFileName = defaults.EnvoyConfigFileName
	esc.generator.ConfigVolume = getStringParam(paramConfigVolume, annotations)
	esc.generator.TLSBasePath = defaults.EnvoyTLSBasePath
	esc.generator.TLSVolume = getStringParam(paramTLSVolume, annotations)
	esc.generator.NodeID = getNodeID(annotations)
	esc.generator.ClusterID = getStringParam(paramClusterID, annotations)
	esc.generator.ClientCertSecret = getStringParam(paramClientCertificate, annotations)
	esc.generator.ExtraArgs = func() []string {
		extraArgs := getStringParam(paramEnvoyExtraArgs, annotations)
		if extraArgs != "" {
			return strings.Split(extraArgs, " ")
		}

		return nil
	}()

	esc.generator.Resources, err = getContainerResourceRequirements(annotations)
	if err != nil {
		return err
	}

	esc.generator.AdminBindAddress = getStringParam(paramEnvoyAdminBindAddress, annotations)
	esc.generator.AdminPort = getPortOrDefault(paramEnvoyAdminPort, annotations, defaults.EnvoyAdminPort)
	esc.generator.AdminAccessLogPath = getStringParam(paramEnvoyAdminAccessLogPath, annotations)

	esc.generator.LivenessProbe = operatorv1alpha1.ProbeSpec{
		InitialDelaySeconds: defaults.LivenessInitialDelaySeconds,
		TimeoutSeconds:      defaults.LivenessTimeoutSeconds,
		PeriodSeconds:       defaults.LivenessPeriodSeconds,
		SuccessThreshold:    defaults.LivenessSuccessThreshold,
		FailureThreshold:    defaults.LivenessFailureThreshold,
	}
	esc.generator.ReadinessProbe = operatorv1alpha1.ProbeSpec{
		InitialDelaySeconds: defaults.ReadinessProbeInitialDelaySeconds,
		TimeoutSeconds:      defaults.ReadinessProbeTimeoutSeconds,
		PeriodSeconds:       defaults.ReadinessProbePeriodSeconds,
		SuccessThreshold:    defaults.ReadinessProbeSuccessThreshold,
		FailureThreshold:    defaults.ReadinessProbeFailureThreshold,
	}
	esc.generator.ShutdownManagerEnabled = isShtdnMgrEnabled(annotations)
	esc.generator.ShutdownManagerPort = getPortOrDefault(paramShtdnMgrServerPort, annotations, defaults.ShtdnMgrDefaultServerPort)
	esc.generator.ShutdownManagerImage = getStringParam(paramShtdnMgrImage, annotations)
	esc.generator.ShutdownManagerDrainSeconds = getInt64Param(paramShtdnMgrDrainTime, annotations)
	esc.generator.ShutdownManagerDrainStrategy = getDrainStrategy(annotations)

	esc.generator.InitManagerImage = getStringParam(paramInitMgrImage, annotations)

	xdssHost, xdssPort, err := getDiscoveryServiceAddress(ctx, clnt, namespace, annotations)
	if err != nil {
		return err
	}

	esc.generator.XdssHost = xdssHost
	esc.generator.XdssPort = xdssPort
	esc.generator.APIVersion = getStringParam(paramEnvoyAPIVersion, annotations)

	return nil
}

func getDiscoveryServiceAddress(ctx context.Context, clnt client.Client, namespace string, annotations map[string]string) (string, int, error) {
	if dsName := getStringParam(paramDiscoveryServiceName, annotations); dsName != "" {
		// Get the address of the DiscoveryService instance
		ds := &operatorv1alpha1.DiscoveryService{}
		dsKey := types.NamespacedName{Name: dsName, Namespace: namespace}

		if err := clnt.Get(ctx, dsKey, ds); err != nil {
			return "", -1, err
		}

		return fmt.Sprintf("%s.%s.%s", ds.GetServiceConfig().Name, namespace, "svc"), int(ds.GetXdsServerPort()), nil
	}

	// If discovery service name is not specified, assume there is only one DiscoveryService in the namespace
	dsList := &operatorv1alpha1.DiscoveryServiceList{}
	if err := clnt.List(ctx, dsList, client.InNamespace(namespace)); err != nil {
		return "", -1, err
	}

	if n := len(dsList.Items); n != 1 {
		return "", -1, fmt.Errorf("expected just one DiscoveryService, got %d", n)
	}

	return fmt.Sprintf("%s.%s.%s", dsList.Items[0].GetServiceConfig().Name, namespace, "svc"), int(dsList.Items[0].GetXdsServerPort()), nil
}

func lookupMarin3rAnnotation(key string, annotations map[string]string) (string, bool) {
	value, ok := annotations[fmt.Sprintf("%s/%s", marin3rAnnotationsDomain, key)]

	return value, ok
}

func getStringParam(key string, annotations map[string]string) string {
	var defaults = map[string]string{
		paramContainerName:           defaults.SidecarContainerName,
		paramImage:                   defaults.Image,
		paramConfigVolume:            defaults.SidecarConfigVolume,
		paramTLSVolume:               defaults.SidecarTLSVolume,
		paramClientCertificate:       defaults.SidecarClientCertificate,
		paramEnvoyExtraArgs:          defaults.EnvoyExtraArgs,
		paramEnvoyAPIVersion:         defaults.EnvoyAPIVersion,
		paramShtdnMgrEnabled:         "false",
		paramShtdnMgrImage:           defaults.ShtdnMgrImage(),
		paramDiscoveryServiceName:    "",
		paramEnvoyAdminBindAddress:   defaults.EnvoyAdminBindAddress,
		paramEnvoyAdminAccessLogPath: defaults.EnvoyAdminAccessLogPath,
		paramInitMgrImage:            defaults.InitMgrImage(),
	}

	// return the value specified in the corresponding annotation, if any
	if value, ok := lookupMarin3rAnnotation(key, annotations); ok {
		return value
	}

	// the default for cluster-id is to be set to the node-id value, which
	// has no default but is always set by the user, otherwise we won't get
	// this far as there are previous checks that validate that node-id is set
	if key == "cluster-id" {
		return getNodeID(annotations)
	}

	// return a default value
	return defaults[key]
}

func getInt64Param(key string, annotations map[string]string) int64 {
	var defaults = map[string]int64{
		paramShtdnMgrDrainTime: defaults.GracefulShutdownTimeoutSeconds,
	}

	if s, ok := lookupMarin3rAnnotation(key, annotations); ok {
		i, err := strconv.Atoi(s)
		if err == nil {
			return int64(i)
		}
	}

	// return default value if annotation not present or error during parsing
	return defaults[key]
}

func getDrainStrategy(annotations map[string]string) defaults.DrainStrategy {
	if value, ok := lookupMarin3rAnnotation(paramShtdnMgrDrainStrategy, annotations); ok {
		if value == string(defaults.DrainStrategyGradual) || value == string(defaults.DrainStrategyImmediate) {
			return defaults.DrainStrategy(value)
		}
	}

	return defaults.GracefulShutdownStrategy
}

func getNodeID(annotations map[string]string) string {
	// the node-id annotation is always present, otherwise
	// mutation won't even be triggered
	res, _ := lookupMarin3rAnnotation(paramNodeID, annotations)

	return res
}

func getContainerResourceRequirements(annotations map[string]string) (corev1.ResourceRequirements, error) {
	var res corev1.ResourceRequirements

	strCPURequests, okCPURequests := lookupMarin3rAnnotation(paramResourceRequestsCPU, annotations)
	strMemoryRequests, okMemoryRequests := lookupMarin3rAnnotation(paramResourceRequestsMemory, annotations)
	strCPULimits, okCPULimits := lookupMarin3rAnnotation(paramResourceLimitsCPU, annotations)
	strMemoryLimits, okMemoryLimits := lookupMarin3rAnnotation(paramResourceLimitsMemory, annotations)

	if okCPURequests || okMemoryRequests {
		res.Requests = corev1.ResourceList{}
	}

	if okCPULimits || okMemoryLimits {
		res.Limits = corev1.ResourceList{}
	}

	if okCPURequests {
		cpuRequests, err := resource.ParseQuantity(strCPURequests)
		if err != nil {
			return corev1.ResourceRequirements{}, err
		}

		res.Requests[corev1.ResourceCPU] = cpuRequests
	}

	if okMemoryRequests {
		memoryRequests, err := resource.ParseQuantity(strMemoryRequests)
		if err != nil {
			return corev1.ResourceRequirements{}, err
		}

		res.Requests[corev1.ResourceMemory] = memoryRequests
	}

	if okCPULimits {
		cpuLimits, err := resource.ParseQuantity(strCPULimits)
		if err != nil {
			return corev1.ResourceRequirements{}, err
		}

		res.Limits[corev1.ResourceCPU] = cpuLimits
	}

	if okMemoryLimits {
		memoryLimits, err := resource.ParseQuantity(strMemoryLimits)
		if err != nil {
			return corev1.ResourceRequirements{}, err
		}

		res.Limits[corev1.ResourceMemory] = memoryLimits
	}

	return res, nil
}

func getPortOrDefault(key string, annotations map[string]string, defaultPort uint32) int32 {
	s, ok := lookupMarin3rAnnotation(key, annotations)
	if ok {
		p, err := portNumber(s)
		if err != nil {
			return int32(defaultPort)
		}

		return p
	}

	return int32(defaultPort)
}

// port spec format is "name:port[:protocol],name:port[:protcol]"
func getContainerPorts(annotations map[string]string) ([]corev1.ContainerPort, error) {
	plist := []corev1.ContainerPort{}

	if ports, ok := lookupMarin3rAnnotation(paramPorts, annotations); ok {
		for _, containerPort := range strings.Split(ports, ",") {
			p, err := parsePortSpec(containerPort)
			if err != nil {
				return []corev1.ContainerPort{}, err
			}

			hp, err := hostPortMapping(p.Name, annotations)
			if err != nil {
				return []corev1.ContainerPort{}, err
			}

			if hp != 0 {
				p.HostPort = hp
			}

			plist = append(plist, *p)
		}
	} else {
		// no ports defined for envoy sidecar
		return []corev1.ContainerPort{}, nil
	}

	return plist, nil
}

func parsePortSpec(spec string) (*corev1.ContainerPort, error) {
	var port corev1.ContainerPort

	params := strings.Split(spec, ":")
	// Each port spec should at least contain name and port number
	if len(params) < 2 {
		return nil, errors.New("incorrect format, the por specification format for the envoy sidecar container is 'name:port[:protocol]'")
	}

	port.Name = params[0]

	p, err := portNumber(params[1])
	if err != nil {
		return nil, err
	}

	port.ContainerPort = p

	// Check that protocol matches one of the allowed protocols
	if len(params) == 3 {
		if params[2] == "TCP" || params[2] == "UDP" || params[2] == "SCTP" {
			port.Protocol = corev1.Protocol(params[2])
		} else {
			return nil, fmt.Errorf("unsupported port protocol '%s'", params[2])
		}
	}

	return &port, nil
}

func hostPortMapping(portName string, annotations map[string]string) (int32, error) {
	if specs, ok := lookupMarin3rAnnotation(paramHostPortMapings, annotations); ok {
		for _, spec := range strings.Split(specs, ",") {
			params := strings.Split(spec, ":")
			if len(params) != 2 {
				return 0, fmt.Errorf("incorrect number of params in host-port-mapping spec '%v'", spec)
			}

			if params[0] == portName {
				p, err := portNumber(params[1])
				if err != nil {
					return 0, err
				}

				return p, nil
			}
		}
	}

	return 0, nil
}

func portNumber(sport string) (int32, error) {
	// Parse and validate the port number
	iport, err := strconv.Atoi(sport)
	if err != nil {
		return 0, fmt.Errorf("%v doesn't look like a port number, check your port specs", sport)
	}
	// Check if port is within allowed ranges. Privileged ports are not allowed
	if iport < 1024 || iport > 65535 {
		return 0, fmt.Errorf("port number %v is not in the range 1024-65535", iport)
	}

	return int32(iport), nil
}

func isShtdnMgrEnabled(annotations map[string]string) bool {
	b, err := strconv.ParseBool(getStringParam(paramShtdnMgrEnabled, annotations))
	if err != nil {
		return false
	}

	return b
}

func (esc *envoySidecarConfig) containers() []corev1.Container {
	return esc.generator.Containers()
}

func (esc *envoySidecarConfig) initContainers() []corev1.Container {
	return esc.generator.InitContainers()
}

func (esc *envoySidecarConfig) volumes() []corev1.Volume {
	return esc.generator.Volumes()
}

func parseExtraLifecycleHooksAnnotation(annotations map[string]string) []string {
	c := getStringParam(paramShtdnMgrExtraLifecycleHooks, annotations)
	if c == "" {
		return []string{}
	}

	hooks := strings.Split(c, ",")

	return hooks
}

func getContainerByName(name string, containers []corev1.Container) (corev1.Container, int, error) {
	for pos, c := range containers {
		if c.Name == name {
			return c, pos, nil
		}
	}

	return corev1.Container{}, -1, fmt.Errorf("container '%s' specified in the 'shutdown-manager.extra-lifecycle-hooks' annotation was not found", name)
}

func (esc *envoySidecarConfig) addExtraLifecycleHooks(containers []corev1.Container, annotations map[string]string) ([]corev1.Container, error) {
	names := parseExtraLifecycleHooksAnnotation(annotations)
	for _, name := range names {
		_, pos, err := getContainerByName(name, containers)
		if err != nil {
			return nil, err
		}

		containers[pos].Lifecycle = &corev1.Lifecycle{
			PreStop: &corev1.LifecycleHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   shutdownmanager.DrainEndpoint,
					Port:   intstr.FromInt(int(esc.generator.ShutdownManagerPort)),
					Scheme: corev1.URISchemeHTTP,
				},
			},
		}
	}

	return containers, nil
}
