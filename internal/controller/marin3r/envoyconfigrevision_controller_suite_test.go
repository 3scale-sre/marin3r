package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/3scale-sre/marin3r/api/envoy"
	marin3rv1alpha1 "github.com/3scale-sre/marin3r/api/marin3r/v1alpha1"
	xdss "github.com/3scale-sre/marin3r/internal/pkg/discoveryservice/xdss"
	xdss_v3 "github.com/3scale-sre/marin3r/internal/pkg/discoveryservice/xdss/v3"
	k8sutil "github.com/3scale-sre/marin3r/internal/pkg/util/k8s"
	testutil "github.com/3scale-sre/marin3r/internal/pkg/util/test"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_extensions_transport_sockets_tls_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("EnvoyConfigRevision controller", func() {
	var namespace string
	var nodeID string

	BeforeEach(func() {
		// Create a namespace for each block
		namespace = "test-ns-" + nameGenerator.Generate()
		By(fmt.Sprintf("creating a new ns %q", namespace))
		// Create a nodeID for each block
		nodeID = nameGenerator.Generate()
		// Add any setup steps that needs to be executed before each test
		testNamespace := &corev1.Namespace{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		}

		err := k8sClient.Create(context.Background(), testNamespace)
		Expect(err).ToNot(HaveOccurred())

		n := &corev1.Namespace{}
		Eventually(func() error {
			return k8sClient.Get(context.Background(), types.NamespacedName{Name: namespace}, n)
		}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())

	})

	AfterEach(func() {
		// Delete the namespace
		testNamespace := &corev1.Namespace{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		}
		// Add any teardown steps that needs to be executed after each test
		err := k8sClient.Delete(context.Background(), testNamespace, client.PropagationPolicy(metav1.DeletePropagationForeground))
		Expect(err).ToNot(HaveOccurred())

		n := &corev1.Namespace{}
		Eventually(func() bool {
			err := k8sClient.Get(context.Background(), types.NamespacedName{Name: namespace}, n)
			if err != nil && errors.IsNotFound(err) {
				return false
			}

			return true
		}, 60*time.Second, 5*time.Second).Should(BeTrue())
	})

	Context("using v3 envoy API version", func() {
		var ecr *marin3rv1alpha1.EnvoyConfigRevision

		BeforeEach(func() {
			By("creating a v3 EnvoyConfigRevision")
			ecr = &marin3rv1alpha1.EnvoyConfigRevision{
				ObjectMeta: metav1.ObjectMeta{Name: "ecr", Namespace: namespace},
				Spec: marin3rv1alpha1.EnvoyConfigRevisionSpec{
					NodeID:   nodeID,
					Version:  "xxxx",
					EnvoyAPI: ptr.To(envoy.APIv3),
					Resources: []marin3rv1alpha1.Resource{
						{Type: envoy.Endpoint, Value: k8sutil.StringtoRawExtension(`{"cluster_name": "endpoint"}`)},
					}},
			}
			err := k8sClient.Create(context.Background(), ecr)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() error {
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: "ecr", Namespace: namespace}, ecr)
			}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
		})

		When("RevisionPublished condition is false in EnvoyConfigRevision", func() {

			It("should not make changes to the xDS cache", func() {

				_, err := ecrV3Reconciler.XdsCache.GetSnapshot(ecr.Spec.NodeID)
				Expect(err).To(HaveOccurred())
			})
		})

		When("RevisionPublished condition is true in EnvoyConfigRevision", func() {

			It("should update the xDS cache with a new snapshot for the nodeID", func() {

				By("setting ECR RevisionPublished condition to true")
				ecr = &marin3rv1alpha1.EnvoyConfigRevision{}
				Eventually(func() error {
					return k8sClient.Get(context.Background(), types.NamespacedName{Name: "ecr", Namespace: namespace}, ecr)
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())

				patch := client.MergeFrom(ecr.DeepCopy())
				meta.SetStatusCondition(&ecr.Status.Conditions, metav1.Condition{
					Type:   marin3rv1alpha1.RevisionPublishedCondition,
					Status: metav1.ConditionTrue,
					Reason: "test",
				})

				err := k8sClient.Status().Patch(context.Background(), ecr, patch)
				Expect(err).ToNot(HaveOccurred())
				Expect(meta.IsStatusConditionTrue(ecr.Status.Conditions, marin3rv1alpha1.RevisionPublishedCondition)).To(BeTrue())

				By("checking that a snapshot for spec.nodeId exists in the v3 xDS cache")
				var gotV3Snap xdss.Snapshot
				Eventually(func() error {
					gotV3Snap, err = ecrV3Reconciler.XdsCache.GetSnapshot(ecr.Spec.NodeID)

					return err
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())

				wantSnap := xdss_v3.NewSnapshot().SetResources(envoy.Endpoint, []envoy.Resource{
					&envoy_config_endpoint_v3.ClusterLoadAssignment{ClusterName: "endpoint"},
				})
				Expect(testutil.SnapshotsAreEqual(gotV3Snap, wantSnap)).To(BeTrue())

			})

		})
	})

	Context("load certificates from secrets", func() {
		var ecr *marin3rv1alpha1.EnvoyConfigRevision

		BeforeEach(func() {
			By("creating a secret of 'kubernetes.io/tls' type")
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "secret", Namespace: namespace},
				Type:       corev1.SecretTypeTLS,
				Data:       map[string][]byte{"tls.crt": []byte("cert"), "tls.key": []byte("key")},
			}
			err := k8sClient.Create(context.Background(), secret)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() error {
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: "secret", Namespace: namespace}, secret)
			}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())

			By("creating a EnvoyConfigRevision with a reference to the created Secret")
			ecr = &marin3rv1alpha1.EnvoyConfigRevision{
				ObjectMeta: metav1.ObjectMeta{Name: "ecr", Namespace: namespace},
				Spec: marin3rv1alpha1.EnvoyConfigRevisionSpec{
					NodeID:   nodeID,
					Version:  "xxxx",
					EnvoyAPI: ptr.To(envoy.APIv3),
					Resources: []marin3rv1alpha1.Resource{
						{Type: envoy.Secret, GenerateFromTlsSecret: ptr.To("secret")},
					},
				},
			}
			err = k8sClient.Create(context.Background(), ecr)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() error {
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: "ecr", Namespace: namespace}, ecr)
			}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())

			By("settign the EnvoyConfigRevision as published")
			patch := client.MergeFrom(ecr.DeepCopy())
			meta.SetStatusCondition(&ecr.Status.Conditions, metav1.Condition{
				Type:   marin3rv1alpha1.RevisionPublishedCondition,
				Status: metav1.ConditionTrue,
				Reason: "test",
			})
			err = k8sClient.Status().Patch(context.Background(), ecr, patch)
			Expect(err).ToNot(HaveOccurred())
			Expect(meta.IsStatusConditionTrue(ecr.Status.Conditions, marin3rv1alpha1.RevisionPublishedCondition)).To(BeTrue())

			wantSnap := xdss_v3.NewSnapshot().SetResources(envoy.Secret, []envoy.Resource{
				&envoy_extensions_transport_sockets_tls_v3.Secret{
					Name: "secret",
					Type: &envoy_extensions_transport_sockets_tls_v3.Secret_TlsCertificate{
						TlsCertificate: &envoy_extensions_transport_sockets_tls_v3.TlsCertificate{
							PrivateKey: &envoy_config_core_v3.DataSource{
								Specifier: &envoy_config_core_v3.DataSource_InlineBytes{InlineBytes: []byte("key")},
							},
							CertificateChain: &envoy_config_core_v3.DataSource{
								Specifier: &envoy_config_core_v3.DataSource_InlineBytes{InlineBytes: []byte("cert")},
							}}}},
			})

			By("waiting for the envoy resources to be published in the xDS cache")
			Eventually(func() bool {
				gotV3Snap, err := ecrV3Reconciler.XdsCache.GetSnapshot(ecr.Spec.NodeID)
				if err != nil {
					return false
				}
				// spew.Dump(gotV3Snap)
				return testutil.SnapshotsAreEqual(gotV3Snap, wantSnap)
			}, 60*time.Second, 5*time.Second).Should(BeTrue())
		})

		When("Secret changes", func() {

			It("should update the xDS cache with new snapshot for the nodeID", func() {
				By("updating the certificate contained in the Secret resource")
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{Name: "secret", Namespace: namespace},
					Type:       corev1.SecretTypeTLS,
					Data:       map[string][]byte{"tls.crt": []byte("new-cert"), "tls.key": []byte("new-key")},
				}
				err := k8sClient.Update(context.Background(), secret)
				Expect(err).ToNot(HaveOccurred())

				Eventually(func() bool {
					err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "secret", Namespace: namespace}, secret)

					return string(secret.Data["tls.crt"]) == "new-cert"
				}, 60*time.Second, 5*time.Second).Should(BeTrue())

				wantSnap := xdss_v3.NewSnapshot().SetResources(envoy.Secret, []envoy.Resource{
					&envoy_extensions_transport_sockets_tls_v3.Secret{
						Name: "secret",
						Type: &envoy_extensions_transport_sockets_tls_v3.Secret_TlsCertificate{
							TlsCertificate: &envoy_extensions_transport_sockets_tls_v3.TlsCertificate{
								PrivateKey: &envoy_config_core_v3.DataSource{
									Specifier: &envoy_config_core_v3.DataSource_InlineBytes{InlineBytes: []byte("new-key")},
								},
								CertificateChain: &envoy_config_core_v3.DataSource{
									Specifier: &envoy_config_core_v3.DataSource_InlineBytes{InlineBytes: []byte("new-cert")},
								}}}},
				})

				By("checking the new certificate it's in the xDS cache")
				Eventually(func() bool {
					gotV3Snap, err := ecrV3Reconciler.XdsCache.GetSnapshot(ecr.Spec.NodeID)
					if err != nil {
						return false
					}

					return testutil.SnapshotsAreEqual(gotV3Snap, wantSnap)
					// this should occur triggered by an event, before the 60 seconds
					// ECR reconcile period
				}, 10*time.Second, 1*time.Second).Should(BeTrue())

			})

		})
	})

	Context("EnvoyConfigRevision finalizer", func() {
		var ecr *marin3rv1alpha1.EnvoyConfigRevision

		BeforeEach(func() {
			By("creating an EnvoyConfigRevision")
			ecr = &marin3rv1alpha1.EnvoyConfigRevision{
				ObjectMeta: metav1.ObjectMeta{Name: "ecr", Namespace: namespace},
				Spec: marin3rv1alpha1.EnvoyConfigRevisionSpec{
					NodeID:  nodeID,
					Version: "xxxx",
					Resources: []marin3rv1alpha1.Resource{
						{Type: envoy.Endpoint, Value: k8sutil.StringtoRawExtension("{\"cluster_name\": \"endpoint\"}")},
					}},
			}
			err := k8sClient.Create(context.Background(), ecr)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() error {
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: "ecr", Namespace: namespace}, ecr)
			}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
		})

		When("resource is created", func() {

			It("should have a finalizer", func() {
				Eventually(func() bool {
					err := k8sClient.Get(context.Background(), types.NamespacedName{Name: "ecr", Namespace: namespace}, ecr)
					Expect(err).ToNot(HaveOccurred())

					return len(ecr.GetFinalizers()) == 1
				}, 60*time.Second, 5*time.Second).Should(BeTrue())
			})
		})

		When("resource is deleted", func() {

			BeforeEach(func() {
				By("setting the published condition in the EnvoyConfigRevision to force execution of the finalizer code")
				patch := client.MergeFrom(ecr.DeepCopy())
				meta.SetStatusCondition(&ecr.Status.Conditions, metav1.Condition{
					Type:   marin3rv1alpha1.RevisionPublishedCondition,
					Status: metav1.ConditionTrue,
					Reason: "test",
				})
				err := k8sClient.Status().Patch(context.Background(), ecr, patch)
				Expect(err).ToNot(HaveOccurred())

				By("waiting for the EnvoyConfigRevision to get published")
				Eventually(func() error {
					_, err := ecrV3Reconciler.XdsCache.GetSnapshot(ecr.Spec.NodeID)

					return err
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())

				Expect(k8sClient.Delete(context.Background(), ecr)).Should(Succeed())
			})

			Specify("Snapshot for the nodeID should have been cleared in the xDS cache", func() {
				Eventually(func() error {
					_, err := ecrV3Reconciler.XdsCache.GetSnapshot(ecr.Spec.NodeID)

					return err
				}, 60*time.Second, 5*time.Second).Should(HaveOccurred())
			})
		})
	})

	Context("EnvoyConfigRevision taints", func() {
		var ecr *marin3rv1alpha1.EnvoyConfigRevision

		BeforeEach(func() {
			By("creating an EnvoyConfigRevision")
			ecr = &marin3rv1alpha1.EnvoyConfigRevision{
				ObjectMeta: metav1.ObjectMeta{Name: "ecr", Namespace: namespace},
				Spec: marin3rv1alpha1.EnvoyConfigRevisionSpec{
					NodeID:  nodeID,
					Version: "xxxx",
					Resources: []marin3rv1alpha1.Resource{
						{Type: envoy.Endpoint, Value: k8sutil.StringtoRawExtension("{\"cluster_name\": \"endpoint\"}")},
					}},
			}
			err := k8sClient.Create(context.Background(), ecr)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() error {
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: "ecr", Namespace: namespace}, ecr)
			}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
		})

		When("RevisionTainted condition is true", func() {

			BeforeEach(func() {
				By("setting the RevisionTained condition in the EnvoyConfigRevision")
				patch := client.MergeFrom(ecr.DeepCopy())
				meta.SetStatusCondition(&ecr.Status.Conditions, metav1.Condition{
					Type:   marin3rv1alpha1.RevisionTaintedCondition,
					Status: metav1.ConditionTrue,
					Reason: "test",
				})
				err := k8sClient.Status().Patch(context.Background(), ecr, patch)
				Expect(err).ToNot(HaveOccurred())

			})

			Specify("status.tainted should be true", func() {
				Eventually(func() bool {
					err := k8sClient.Get(context.Background(), types.NamespacedName{Name: "ecr", Namespace: namespace}, ecr)
					Expect(err).ToNot(HaveOccurred())

					return ecr.Status.IsTainted()
				}, 60*time.Second, 5*time.Second).Should(BeTrue())
			})

			Specify("status.tainted should be false when condition is cleared", func() {

				By("unsetting the RevisionTained condition in the EnvoyConfigRevision")
				patch := client.MergeFrom(ecr.DeepCopy())
				meta.SetStatusCondition(&ecr.Status.Conditions, metav1.Condition{
					Type:   marin3rv1alpha1.RevisionTaintedCondition,
					Status: metav1.ConditionFalse,
					Reason: "test",
				})
				err := k8sClient.Status().Patch(context.Background(), ecr, patch)
				Expect(err).ToNot(HaveOccurred())

				By("checking the status.Tainded field in the EnvoyConfigCache")
				Eventually(func() bool {
					err := k8sClient.Get(context.Background(), types.NamespacedName{Name: "ecr", Namespace: namespace}, ecr)
					Expect(err).ToNot(HaveOccurred())

					return !ecr.Status.IsTainted()
				}, 60*time.Second, 5*time.Second).Should(BeTrue())
			})
		})

		When("resources cannot be loaded", func() {
			Specify("the EnvoyConfigRevision taints itself", func() {

				By("updating the EnvoyConfigRevision with an incorrect resource")
				ecr = &marin3rv1alpha1.EnvoyConfigRevision{}
				key := types.NamespacedName{Name: "ecr", Namespace: namespace}
				err := k8sClient.Get(context.Background(), key, ecr)
				Expect(err).ToNot(HaveOccurred())
				patch := client.MergeFrom(ecr.DeepCopy())
				ecr.Spec.Resources = []marin3rv1alpha1.Resource{
					{Type: envoy.Cluster, Value: k8sutil.StringtoRawExtension("{\"wrong_key\": \"wrong_value\"}")},
				}
				err = k8sClient.Patch(context.Background(), ecr, patch)
				Expect(err).ToNot(HaveOccurred())

				By("publishing the EnvoyConfigRevision to force it to load the resources")
				patch = client.MergeFrom(ecr.DeepCopy())
				meta.SetStatusCondition(&ecr.Status.Conditions, metav1.Condition{
					Type:   marin3rv1alpha1.RevisionPublishedCondition,
					Status: metav1.ConditionTrue,
					Reason: "test",
				})
				err = k8sClient.Status().Patch(context.Background(), ecr, patch)
				Expect(err).ToNot(HaveOccurred())

				Eventually(func() bool {
					err := k8sClient.Get(context.Background(), key, ecr)
					Expect(err).ToNot(HaveOccurred())

					return meta.IsStatusConditionTrue(ecr.Status.Conditions, marin3rv1alpha1.RevisionTaintedCondition)
				}, 60*time.Second, 5*time.Second).Should(BeTrue())
			})

			Specify("the EnvoyConfigRevision does not taint itself if the resource that failed was a Secret", func() {

				By("updating the EnvoyConfigRevision with an non-existent Secret")
				ecr = &marin3rv1alpha1.EnvoyConfigRevision{}
				key := types.NamespacedName{Name: "ecr", Namespace: namespace}
				err := k8sClient.Get(context.Background(), key, ecr)
				Expect(err).ToNot(HaveOccurred())
				patch := client.MergeFrom(ecr.DeepCopy())
				ecr.Spec.Resources = []marin3rv1alpha1.Resource{{Type: envoy.Secret, GenerateFromTlsSecret: ptr.To("secret")}}
				err = k8sClient.Patch(context.Background(), ecr, patch)
				Expect(err).ToNot(HaveOccurred())

				By("publishing the EnvoyConfigRevision to force it to load the resources")
				patch = client.MergeFrom(ecr.DeepCopy())
				meta.SetStatusCondition(&ecr.Status.Conditions, metav1.Condition{
					Type:   marin3rv1alpha1.RevisionPublishedCondition,
					Status: metav1.ConditionTrue,
					Reason: "test",
				})
				err = k8sClient.Status().Patch(context.Background(), ecr, patch)
				Expect(err).ToNot(HaveOccurred())

				Eventually(func() bool {
					err := k8sClient.Get(context.Background(), key, ecr)
					Expect(err).ToNot(HaveOccurred())

					return meta.IsStatusConditionTrue(ecr.Status.Conditions, marin3rv1alpha1.RevisionPublishedCondition)
				}, 60*time.Second, 5*time.Second).Should(BeTrue())

				Expect(meta.IsStatusConditionTrue(ecr.Status.Conditions, marin3rv1alpha1.ResourcesInSyncCondition)).ToNot(BeTrue())
				Expect(ecr.Status.IsPublished()).ToNot(BeTrue())
			})
		})
	})

})
