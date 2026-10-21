// Package resource provides intelligent Kubernetes resource reconciliation capabilities.
//
// This package implements a sophisticated system for managing Kubernetes resources with
// fine-grained control over which fields are reconciled. It's designed to eliminate the
// common problem of controllers fighting over resource fields by allowing precise
// specification of which properties should be managed and which should be ignored.
//
// # Core Concepts
//
// Template System: Resources are defined using Templates that implement TemplateInterface.
// Templates consist of two phases:
//   - TemplateBuilder: Creates the base desired state (no API access)
//   - TemplateMutations: Modify the desired state using live cluster data (with API access)
//
// Property-based Reconciliation: Instead of replacing entire resources, this package
// reconciles individual properties specified via JSONPath expressions. This allows:
//   - Ensuring specific fields match desired state (e.g., "metadata.labels", "spec.replicas")
//   - Ignoring fields managed by other controllers (e.g., "spec.clusterIP", "status")
//   - Avoiding unnecessary updates when irrelevant fields change
//
// Normalization: Before comparison, both desired and live states are "normalized" by:
//   - Extracting only the properties specified in the 'ensure' configuration
//   - Removing properties specified in the 'ignore' configuration
//   - Creating comparable objects that contain only the relevant fields
//
// # Operation Types
//
// The package supports two modification operations for applying changes to Kubernetes resources:
//
// Update Operations (ModifyOpUpdate): Uses client.Update() to replace the entire resource.
//   - Replaces the complete object in the Kubernetes API
//   - Requires the most recent resource version to avoid conflicts
//   - Higher chance of conflict with other controllers making concurrent changes
//   - More network bandwidth due to sending the complete object
//
// Patch Operations (ModifyOpPatch): Uses client.Patch() with strategic merge patches.
//   - Sends only the changed fields to the Kubernetes API
//   - Uses strategic merge patch for intelligent field merging
//   - Lower chance of conflicts due to field-level granularity
//   - More efficient in terms of network bandwidth and API server processing
//   - Leverages controller-runtime's client.MergeFrom for robust patch generation
//
// # Resource Reconciliation Workflow
//
// The following ASCII diagram illustrates the complete reconciliation flow:
//
//	┌─────────────────────────────────────────────────────────────────────────────────┐
//	│                          CreateOrModify() Entry Point                           │
//	│                  (ModifyOp determined from template config)                     │
//	└───────────────────────────────┬─────────────────────────────────────────────────┘
//	                                │
//	┌───────────────────────────────▼─────────────────────────────────────────────────┐
//	│ 1. BUILD DESIRED STATE                                                           │
//	│                                                                                  │
//	│    Template.Build()                                                              │
//	│    ├─ Execute TemplateBuilder() ────────► Base Object                           │
//	│    └─ Apply TemplateMutations[] ────────► Final Desired State                   │
//	│                                           (with live cluster data)              │
//	└───────────────────────────────┬─────────────────────────────────────────────────┘
//	                                │
//	┌───────────────────────────────▼─────────────────────────────────────────────────┐
//	│ 2. GET LIVE STATE                                                                │
//	│                                                                                  │
//	│    client.Get(key, live) ──────────────► Live Object from K8s API               │
//	│                                                                                  │
//	│    ┌─ Resource Not Found? ────────────► Create Resource & Return                │
//	│    ├─ Template Disabled? ─────────────► Delete Resource & Return                │
//	│    └─ Continue to normalization                                                  │
//	└───────────────────────────────┬─────────────────────────────────────────────────┘
//	                                │
//	┌───────────────────────────────▼─────────────────────────────────────────────────┐
//	│ 3. NORMALIZE STATES                                                              │
//	│                                                                                  │
//	│    Desired State                          Live State                            │
//	│         │                                     │                                 │
//	│         ▼                                     ▼                                 │
//	│    normalize()                           normalize()                            │
//	│    ├─ Extract 'ensure' properties        ├─ Extract 'ensure' properties         │
//	│    ├─ Remove 'ignore' properties         ├─ Remove 'ignore' properties          │
//	│    └─ Return normalized object           └─ Return normalized object            │
//	│         │                                     │                                 │
//	│         ▼                                     ▼                                 │
//	│    Normalized Desired ◄───────┬─────────► Normalized Live                      │
//	└───────────────────────────────┼─────────────────────────────────────────────────┘
//	                                │
//	┌───────────────────────────────▼─────────────────────────────────────────────────┐
//	│ 4. COMPARE STATES                                                                │
//	│                                                                                  │
//	│    equality.Semantic.DeepEqual(normalizedLive, normalizedDesired)               │
//	│                                                                                  │
//	│    ┌─ States Equal? ───────────────────► No Update Needed & Return              │
//	│    └─ States Different? ─────────────────┐                                      │
//	└──────────────────────────────────────────┼──────────────────────────────────────┘
//	                                           │
//	┌──────────────────────────────────────────▼──────────────────────────────────────┐
//	│ 5. RECONCILE PROPERTIES                                                          │
//	│                                                                                  │
//	│    For each property in 'ensure' list:                                          │
//	│    ├─ property.reconcile(live, desired)                                         │
//	│    ├─ Determine delta: missing/present in desired vs live                       │
//	│    ├─ Apply action: add/remove/update property value                            │
//	│    └─ Continue to next property                                                  │
//	│                                                                                  │
//	│    Property Actions:                                                             │
//	│    • missingInBoth ──────────────────────► No action needed                     │
//	│    • missingFromDesired + presentInLive ─► Delete property from live            │
//	│    • presentInDesired + missingFromLive ─► Add property to live                 │
//	│    • presentInBoth + values differ ──────► Update property value in live       │
//	└───────────────────────────────┬─────────────────────────────────────────────────┘
//	                                │
//	┌───────────────────────────────▼─────────────────────────────────────────────────┐
//	│ 6. APPLY CHANGES                                                                 │
//	│                                                                                  │
//	│    ┌─ ModifyOpUpdate ──► client.Update(ctx, modifiedLiveObject)                 │
//	│    │                     • Replaces entire resource                             │
//	│    │                     • Requires latest resource version                     │
//	│    │                     • Higher conflict probability                          │
//	│    │                                                                             │
//	│    └─ ModifyOpPatch ───► client.Patch(ctx, live, client.MergeFrom(original))   │
//	│                          • Strategic merge patch with only changed fields      │
//	│                          • Lower conflict probability                           │
//	│                          • More efficient bandwidth usage                      │
//	│                                                                                  │
//	│    Return ObjectReference for tracking                                           │
//	└──────────────────────────────────────────────────────────────────────────────────┘
//
// # Configuration System
//
// Resource reconciliation behavior is controlled by two types of JSONPath properties:
//
// Ensure Properties: Fields that should be reconciled from desired to live state.
// Examples:
//   - "metadata.labels" - ensure all labels match
//   - "spec.replicas" - ensure replica count matches
//   - "spec.template.spec.containers[0].image" - ensure specific container image
//
// Ignore Properties: Fields that should be excluded from reconciliation.
// Examples:
//   - "spec.clusterIP" - let Kubernetes manage cluster IP
//   - "metadata.annotations['deployment.kubernetes.io/revision']" - let deployment controller manage
//   - "status" - status is typically managed by controllers
//
// Configuration can be specified:
//  1. Per-template via Template.EnsureProperties and Template.IgnoreProperties
//  2. Globally via the config package for specific GroupVersionKinds
//  3. With fallback to default configuration for unknown resource types
//
// # Key Types
//
// TemplateInterface: Defines how resources should be built and reconciled.
// - Build(): Creates desired state by executing builder + mutations
// - Enabled(): Controls whether resource should exist
// - GetReconcileOptions(): Returns reconciliation configuration (properties, operation type)
// - GetGVK(): Returns the GroupVersionKind for the template's resource type
//
// Template[T]: Concrete implementation of TemplateInterface with fluent API.
// - Supports method chaining for easy configuration
// - Separates static template building from dynamic mutations
// - Automatically infers GVK from the generic type parameter T
// - Configures modification operation (Update or Patch) per template
//
// Property: JSONPath-based field specification for fine-grained control.
// - Supports complex path expressions like "spec.template.spec.containers[0].env[1].value"
// - reconcile() method handles individual property synchronization
//
// # Scheme Management
//
// The package uses a convenient shared default scheme, managed by the runtimeconfig package, for
// GVK inference when no explicit scheme is provided to template constructors. This defaults to the
// standard Kubernetes scheme but can be overridden globally:
//
//	import (
//	    myscheme "github.com/myorg/myoperator/pkg/scheme"
//	    "github.com/3scale-sre/basereconciler/runtimeconfig"
//	)
//
//	func init() {
//		runtimeconfig.SetDefaultScheme(myscheme.Scheme)
//	}
//
// This eliminates the need to pass scheme parameters to every template constructor while still
// allowing per-call overrides when needed.
//
// # Usage Examples
//
// Basic Template using default scheme (most common):
//
//	template := resource.NewTemplateFromObjectFunction(func() *corev1.Service {
//	    return &corev1.Service{...}
//	}).WithEnsureProperties([]resource.Property{"metadata.labels", "spec.ports"})
//
// Template with mutations and default scheme:
//
//	template := resource.NewTemplate(builder).
//	    WithMutation(mutators.SetServiceLiveValues()).
//	    WithIgnoreProperties([]resource.Property{"spec.clusterIP"}).
//	    WithModifyOp(resource.ModifyOpUpdate)
//
// Template with explicit scheme override (when default scheme is insufficient):
//
//	template := resource.NewTemplate(builder, customScheme).
//	    WithEnsureProperties([]resource.Property{"metadata.labels", "spec.ports"}).
//	    WithModifyOp(resource.ModifyOpPatch)
//
// Setting up custom default scheme globally (in init function):
//
//	import myscheme "github.com/myorg/myoperator/pkg/scheme"
//
//	func init() {
//		runtimeconfig.SetDefaultScheme(myscheme.Scheme)  // Now all templates use this by default
//	}
//
// Reconcile Resource (operation type determined by template configuration):
//
//	ref, err := resource.CreateOrModify(ctx, client, scheme, owner, template)
//
// # Advanced Features
//
// Default Scheme Management: Global default scheme via runtimeconfig eliminates the need to pass
// scheme parameters to every constructor while still allowing per-call overrides.
//
// Automatic GVK Inference: Templates automatically determine their resource type from
// the generic type parameter, reducing boilerplate and preventing mismatches.
//
// Live Value Preservation: Mutations can query the Kubernetes API to preserve
// values that shouldn't be overridden (like auto-generated clusterIPs).
//
// Conditional Resource Management: Templates can be dynamically enabled/disabled
// to create or delete resources based on runtime conditions.
//
// Multi-Controller Cooperation: Ignore properties allow multiple controllers
// to manage different aspects of the same resource without conflicts.
//
// Change Detection: Only updates resources when normalized states actually differ,
// avoiding unnecessary API calls and potential race conditions.
package resource
