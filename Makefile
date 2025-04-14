# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 0.13.2-alpha.5

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "candidate,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=candidate,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="candidate,fast,stable")
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# IMAGE_TAG_BASE defines the docker.io namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
#
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
# quay.io/3scale/marin3r-bundle:$VERSION and quay.io/3scale/marin3r-catalog:$VERSION.
IMAGE_TAG_BASE ?= quay.io/3scale/marin3r

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)

# BUNDLE_GEN_FLAGS are the flags passed to the operator-sdk generate bundle command
BUNDLE_GEN_FLAGS ?= -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)

# USE_IMAGE_DIGESTS defines if images are resolved via tags or digests
# You can enable this value if you would like to use SHA Based Digests
# To enable set flag to true
USE_IMAGE_DIGESTS ?= false
ifeq ($(USE_IMAGE_DIGESTS), true)
    BUNDLE_GEN_FLAGS += --use-image-digests
endif

# Image URL to use all building/pushing image targets
IMG ?= $(IMAGE_TAG_BASE):v$(VERSION)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= podman

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...


.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

.PHONY: lint-config
lint-config: golangci-lint ## Verify golangci-lint linter configuration
	$(GOLANGCI_LINT) config verify

.PHONY: go-generate
go-generate: gen-pkg-version gen-pkg-image gen-pkg-envoy-proto
	IMAGE=$(IMG) VERSION=$(VERSION) PATH=$$PATH:$$PWD/bin go generate ./...

.PHONY=gen-pkg-envoy-proto
gen-pkg-envoy-proto: export TARGET_PATH = $(PWD)/bin
gen-pkg-envoy-proto: ## builds the gen-pkg-envoy-proto binary
	 cd generators/pkg-envoy-proto && go build -o $${TARGET_PATH}/gen-pkg-envoy-proto main.go

.PHONY=gen-pkg-version
gen-pkg-version: export TARGET_PATH = $(PWD)/bin
gen-pkg-version: ## builds the gen-pkg-version binary
	 cd generators/pkg-version && go build -o $${TARGET_PATH}/gen-pkg-version main.go

.PHONY=gen-pkg-image
gen-pkg-image: export TARGET_PATH = $(PWD)/bin
gen-pkg-image: ## builds the gen-pkg-image binary
	 cd generators/pkg-image && go build -o $${TARGET_PATH}/gen-pkg-image main.go

##@ Test

TEST_PKG = ./api/... ./internal/...
COVERPROFILE = coverprofile.out

.PHONY: test-new
test: manifests generate fmt vet envtest ginkgo ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" \
		$(GINKGO) -p -procs=$(shell nproc) -coverprofile=$(COVERPROFILE) -coverpkg=$(COVERPKGS) $(TEST_PKG)
	$(MAKE) fix-cover && go tool cover -func=$(COVERPROFILE) | awk '/total/{print $$3}'

.PHONY: fix-cover
fix-cover:
	tmpfile=$$(mktemp) && grep -v "_generated.deepcopy.go" $(COVERPROFILE) > $${tmpfile} && cat $${tmpfile} > $(COVERPROFILE) && rm -f $${tmpfile}

.PHONY: e2e-test
e2e-test: export KUBECONFIG = $(PWD)/kubeconfig
e2e-test: kind-create ## Runs e2e test suite
	$(MAKE) e2e-envtest-suite
	$(MAKE) kind-delete

.PHONY: e2e-envtest-suite
e2e-envtest-suite: export KUBECONFIG = $(PWD)/kubeconfig
e2e-envtest-suite: container-build kind-load-image manifests ginkgo deploy-test
	$(GINKGO) -r -p ./test/e2e

##@ Build

CONTAINER_TOOL ?= podman


.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/main.go operator

# MULTI-PLATFORM BUILD/PUSH FUNCTIONS
# NOTE IF USING DOCKER (https://docs.docker.com/build/building/multi-platform/#prerequisites):
#   The "classic" image store of the Docker Engine does not support multi-platform images. 
#   Switching to the containerd image store ensures that your Docker Engine can push, pull,
#   and build multi-platform images.

# container-build-multiplatform will build a multiarch image using the defined container tool
# $1 - image tag
# $2 - container tool: docker/podman
# $3 - dockerfile path
# $4 - build context path
# $5 - platforms
define container-build-multiplatform
@{\
set -e; \
echo "Building $1 for $5 using $2"; \
if [ "$2" = "docker" ]; then \
	docker buildx build --platform $5 -f $3 --tag $1 $4; \
elif [ "$2" = "podman" ]; then \
	podman build --platform $5 -f $3 --manifest $1 $4; \
else \
	echo "unknown container tool $2"; exit -1; \
fi \
}
endef

# container-push-multiplatform will push a multiarch image using the defined container tool
# $1 - image tag
# $2 - container tool: docker/podman
define container-push-multiplatform
@{\
set -e; \
echo "Pushing $1 using $2"; \
if [ "$2" = "docker" ]; then \
	docker push $1; \
elif [ "$2" = "podman" ]; then \
	podman manifest push --all $1; \
else \
	echo "unknown container tool $2"; exit -1; \
fi \
}
endef

# LOCAL PLATFORM BUILD

.PHONY: container-build
container-build:
	$(call container-build-multiplatform,$(IMG),$(CONTAINER_TOOL),Dockerfile,.,$(shell go env GOARCH))
	$(CONTAINER_TOOL) tag $(IMG) $(IMAGE_TAG_BASE):test

.PHONY: container-push
container-push:
	$(call container-push-multiplatform,$(IMG),$(CONTAINER_TOOL))

# MULTIPLATFORM BUILD

# PLATFORMS defines the target platforms for mult-platform build.
PLATFORMS ?= linux/arm64,linux/amd64

.PHONY: container-buildx
container-buildx: ## cross-platfrom build
	$(call container-build-multiplatform,$(IMG),$(CONTAINER_TOOL),Dockerfile,.,$(PLATFORMS))
	
.PHONY: container-pushx
container-pushx:
	$(call container-push-multiplatform,$(IMG),$(CONTAINER_TOOL))

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	cd config/webhook && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

.PHONY: undeploy
undeploy: kustomize ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -


deploy-test: manifests kustomize ## Deploy controller (test configuration) to the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/test | $(KUBECTL) apply -f -

undeploy-test: manifests kustomize ## Undeploy controller (test configuration) from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/test | $(KUBECTL) delete -f -

deploy-cert-manager: ## Deployes cert-manager in the K8s cluster specified in ~/.kube/config.
	$(KUBECTL) apply -f https://github.com/jetstack/cert-manager/releases/download/v1.7.3/cert-manager.yaml
	$(KUBECTL) -n cert-manager wait --timeout=300s --for=condition=Available deployments --all

##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GINKGO ?= $(LOCALBIN)/ginkgo
GOCOVMERGE ?= $(LOCALBIN)/gocovmerge
CRD_REFDOCS ?= $(LOCALBIN)/crd-ref-docs
KIND ?= $(LOCALBIN)/kind
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

## Tool Versions
KUSTOMIZE_VERSION ?= v5.6.0
CONTROLLER_TOOLS_VERSION ?= v0.17.1
#ENVTEST_VERSION is the version of controller-runtime release branch to fetch the envtest setup script (i.e. release-0.20)
ENVTEST_VERSION ?= $(shell go list -m -f "{{ .Version }}" sigs.k8s.io/controller-runtime | awk -F'[v.]' '{printf "release-%d.%d", $$2, $$3}')
#ENVTEST_K8S_VERSION is the version of Kubernetes to use for setting up ENVTEST binaries (i.e. 1.31)
ENVTEST_K8S_VERSION ?= $(shell go list -m -f "{{ .Version }}" k8s.io/api | awk -F'[v.]' '{printf "1.%d", $$3}')
GOLANGCI_LINT_VERSION ?= v1.63.4
GINKGO_VERSION ?= v2.23.3
CRD_REFDOCS_VERSION ?= v0.0.8
KIND_VERSION ?= v0.27.0

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: setup-envtest
setup-envtest: envtest ## Download the binaries required for ENVTEST in the local bin directory.
	@echo "Setting up envtest binaries for Kubernetes version $(ENVTEST_K8S_VERSION)..."
	@$(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path || { \
		echo "Error: Failed to set up envtest binaries for version $(ENVTEST_K8S_VERSION)."; \
		exit 1; \
	}

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: ginkgo
ginkgo: $(GINKGO) ## Download ginkgo locally if necessary
$(GINKGO):
	$(call go-install-tool,$(GINKGO),github.com/onsi/ginkgo/v2/ginkgo,$(GINKGO_VERSION))

.PHONY: crd-ref-docs
crd-ref-docs: ## Download crd-ref-docs locally if necessary
	$(call go-install-tool,$(CRD_REFDOCS),github.com/elastic/crd-ref-docs,$(CRD_REFDOCS_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

.PHONY: kind
KIND = $(shell pwd)/bin/kind
kind: $(KIND) ## Download kind locally if necessary
$(KIND):
	$(call go-install-tool,$(KIND),sigs.k8s.io/kind,$(KIND_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef

##@ Operator SDK related targets

.PHONY: operator-sdk
OPERATOR_SDK = bin/operator-sdk-$(OPERATOR_SDK_RELEASE)
OPERATOR_SDK_RELEASE = v1.39.0
operator-sdk: ## Download operator-sdk locally if necessary.
ifeq (,$(wildcard $(OPERATOR_SDK)))
ifeq (,$(shell which $(OPERATOR_SDK) 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPERATOR_SDK)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPERATOR_SDK) https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_RELEASE}/operator-sdk_$${OS}_$${ARCH};\
	chmod +x $(OPERATOR_SDK) ;\
	}
else
OPERATOR_SDK = $(shell which $(OPERATOR_SDK))
endif
endif

.PHONY: bundle
bundle: manifests kustomize operator-sdk ## Generate bundle manifests and metadata, then validate generated files.
	$(OPERATOR_SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	cd config/webhook && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle $(BUNDLE_GEN_FLAGS)
	$(OPERATOR_SDK) bundle validate ./bundle


.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	$(call container-build-multiplatform,$(BUNDLE_IMG),$(CONTAINER_TOOL),bundle.Dockerfile,.,$(shell go env GOARCH))

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(call container-push-multiplatform,$(BUNDLE_IMG),$(CONTAINER_TOOL))

.PHONY: opm
OPM = $(LOCALBIN)/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.23.0/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

.PHONY: catalog-validate
catalog-validate: ## Validate the file based catalog.
	$(OPM) validate catalog/marin3r

.PHONY: catalog-build
catalog-build: opm catalog-validate ## Build the file based catalog image.
	$(call container-build-multiplatform,$(CATALOG_IMG),$(CONTAINER_TOOL),catalog/marin3r.Dockerfile,catalog/,$(PLATFORMS))

.PHONY: catalog-run
catalog-run: catalog-build ## Run the catalog image locally.
	$(CONTAINER_TOOL) run --rm -p 50051:50051 $(CATALOG_IMG)

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(call container-push-multiplatform,$(CATALOG_IMG),$(CONTAINER_TOOL))

.PHONY: catalog-add-bundle-to-alpha
catalog-add-bundle-to-alpha: opm ## Adds a bundle to a file based catalog
	$(OPM) render $(BUNDLE_IMGS) -oyaml > catalog/marin3r/objects/marin3r.v$(VERSION).clusterserviceversion.yaml
	yq -i '.entries += {"name": "marin3r.v$(VERSION)","replaces":"$(shell yq '.entries[-1].name' catalog/marin3r/alpha-channel.yaml)"}' catalog/marin3r/alpha-channel.yaml

.PHONY: catalog-add-bundle-to-stable
catalog-add-bundle-to-stable: opm ## Adds a bundle to a file based catalog
	$(OPM) render $(BUNDLE_IMGS) -oyaml > catalog/marin3r/objects/marin3r.v$(VERSION).clusterserviceversion.yaml
	yq -i '.entries += {"name": "marin3r.v$(VERSION)","replaces":"$(shell yq '.entries[-1].name' catalog/marin3r/alpha-channel.yaml)"}' catalog/marin3r/alpha-channel.yaml
	yq -i '.entries += {"name": "marin3r.v$(VERSION)","replaces":"$(shell yq '.entries[-1].name' catalog/marin3r/stable-channel.yaml)"}' catalog/marin3r/stable-channel.yaml

##@ Kind Deployment

export KIND_EXPERIMENTAL_PROVIDER=$(CONTAINER_TOOL)

kind-create: export KUBECONFIG = $(PWD)/kubeconfig
kind-create: container-build kind ## Runs a k8s kind cluster with a local registry in "localhost:5000" and ports 1080 and 1443 exposed to the host
	$(KIND) create cluster --wait 5m --config test/kind.yaml --image kindest/node:v1.27.10
	$(MAKE) deploy-cert-manager
	$(MAKE) kind-load-image

kind-deploy: export KUBECONFIG = $(PWD)/kubeconfig
kind-deploy: manifests kustomize ## Deploy operator to the Kind K8s cluster
	$(KUSTOMIZE) build config/test | $(KUBECTL) apply -f -

kind-undeploy: export KUBECONFIG = $(PWD)/kubeconfig
kind-undeploy: ## Undeploy controller from the Kind K8s cluster
	$(KUSTOMIZE) build config/test | $(KUBECTL) delete -f -

kind-load-image: export KUBECONFIG = $(PWD)/kubeconfig
kind-load-image: kind ## Load the marin3r:test image into the cluster
	tmpfile=$$(mktemp) && \
		$(CONTAINER_TOOL) save -o $${tmpfile}  $(IMAGE_TAG_BASE):test && \
		$(KIND) load image-archive $${tmpfile} --name kind && \
		rm $${tmpfile}

kind-refresh-image: export KUBECONFIG = ${PWD}/kubeconfig
kind-refresh-image: manifests kind container-build ## Reloads the image into the K8s cluster and deletes the old pods
	$(MAKE) kind-load-image
	$(KUBECTL) -n marin3r-system delete pod -l control-plane=controller-manager
	$(KUBECTL) -n marin3r-system delete pod -l control-plane=controller-webhook
	$(KUBECTL) -n default delete pod -l app.kubernetes.io/component=discovery-service

kind-delete: ## Deletes the kind cluster and the registry
kind-delete: kind
	$(KIND) delete cluster

##@ Release

prepare-alpha-release: generate fmt vet manifests go-generate bundle ## Generates bundle manifests for alpha channel release

prepare-stable-release: generate fmt vet manifests go-generate bundle refdocs ## Generates bundle manifests for stable channel release
	$(MAKE) bundle CHANNELS=alpha,stable DEFAULT_CHANNEL=stable

bundle-publish: container-build container-push bundle-build bundle-push ## Builds and pushes operator and bundle images

catalog-publish: catalog-build catalog-push catalog-retag-latest ## Builds and pushes the catalog image

get-new-release: ## Checks if a release with the name $(VERSION) already exists in https://github.com/3scale-sre/marin3r/releases
	@hack/new-release.sh v$(VERSION)

catalog-retag-latest:
	$(CONTAINER_TOOL) tag $(CATALOG_IMG) $(IMAGE_TAG_BASE)-catalog:latest
	$(call container-push-multiplatform,$(IMAGE_TAG_BASE)-catalog:latest,$(CONTAINER_TOOL))

##@ Run components locally
$(TMP)/certs:
	hack/gen-certs.sh

ENVOY_VERSION ?= v1.23.2

run-ds: ## locally starts a discovery service
run-ds: manifests generate fmt vet go-generate $(TMP)/certs
	WATCH_NAMESPACE="default" go run main.go \
		discovery-service \
		--server-certificate-path $(TMP)/certs/server \
		--ca-certificate-path $(TMP)/certs/ca \
		--client-certificate-path $(TMP)/certs/client \
		--metrics-bind-address :8383
		--debug

run-envoy: ## runs an envoy process in a container that will try to connect to a local discovery service
run-envoy: $(TMP)/certs
	docker run -ti --rm \
		--network=host \
		--add-host marin3r.default.svc:127.0.0.1 \
		-v $$(pwd)/$(TMP)/certs/client:/etc/envoy/tls \
		-v $$(pwd)/examples/local:/config \
		envoyproxy/envoy:$(ENVOY_VERSION) \
		envoy -c /config/envoy-client-bootstrap.yaml $(ARGS)

##@ Other

refdocs: ## Generates api reference documentation from code
refdocs: crd-ref-docs
	$(CRD_REFDOCS) \
		--source-path=apis \
		--config=docs/api-reference/config.yaml \
		--templates-dir=docs/api-reference/templates/asciidoctor \
		--renderer=asciidoctor \
		--output-path=docs/api-reference/reference.asciidoc

TMP = tmp
$(TMP): ## Create project local tmp directory
	mkdir tmp

clean: ## Clean project directory
	rm -rf $(TMP) $(LOCALBIN) $(COVERPROFILE) kubeconfig