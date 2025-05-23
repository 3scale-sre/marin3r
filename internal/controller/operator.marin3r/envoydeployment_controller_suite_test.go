package controllers

import (
	"context"
	"time"

	"github.com/3scale-sre/marin3r/api/envoy"
	"github.com/3scale-sre/marin3r/api/envoy/defaults"
	marin3rv1alpha1 "github.com/3scale-sre/marin3r/api/marin3r/v1alpha1"
	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("EnvoyDeployment controller", func() {
	var namespace string
	var ed *operatorv1alpha1.EnvoyDeployment

	BeforeEach(func() {
		// Create a namespace for each block
		namespace = "test-ns-" + nameGenerator.Generate()

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

		By("creating a DiscoveryService instance")
		ds := &operatorv1alpha1.DiscoveryService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "instance",
				Namespace: namespace,
			},
			Spec: operatorv1alpha1.DiscoveryServiceSpec{
				Image: ptr.To("image"),
			},
		}
		err = k8sClient.Create(context.Background(), ds)
		Expect(err).ToNot(HaveOccurred())
		Eventually(func() error {
			return k8sClient.Get(context.Background(), types.NamespacedName{Name: "instance", Namespace: namespace}, ds)
		}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())

		By("creating an EnvoyConfig instance")
		ec := &marin3rv1alpha1.EnvoyConfig{
			ObjectMeta: metav1.ObjectMeta{Name: "config", Namespace: namespace},
			Spec: marin3rv1alpha1.EnvoyConfigSpec{
				EnvoyAPI:       ptr.To(envoy.APIv3),
				NodeID:         "test-node",
				EnvoyResources: &marin3rv1alpha1.EnvoyResources{},
			},
		}
		err = k8sClient.Create(context.Background(), ec)
		Expect(err).ToNot(HaveOccurred())
		Eventually(func() bool {
			err := k8sClient.Get(context.Background(), types.NamespacedName{Name: "config", Namespace: namespace}, ec)

			return err == nil
		}, 60*time.Second, 5*time.Second).Should(BeTrue())

		By("creating a EnvoyDeployment instance")
		ed = &operatorv1alpha1.EnvoyDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "instance",
				Namespace: namespace,
			},
			Spec: operatorv1alpha1.EnvoyDeploymentSpec{
				DiscoveryServiceRef: ds.GetName(),
				EnvoyConfigRef:      ec.GetName(),
			},
		}
		err = k8sClient.Create(context.Background(), ed)
		Expect(err).ToNot(HaveOccurred())
		Eventually(func() error {
			return k8sClient.Get(context.Background(), types.NamespacedName{Name: "instance", Namespace: namespace}, ed)
		}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())

	})

	Context("EnvoyDeployment", func() {

		It("creates the required resources", func() {

			By("waiting for the client certificate resource to be created")
			{
				eb := &operatorv1alpha1.DiscoveryServiceCertificate{}
				Eventually(func() error {
					return k8sClient.Get(
						context.Background(),
						types.NamespacedName{Name: defaults.DeploymentClientCertificate + "-instance", Namespace: namespace},
						eb,
					)
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
			}

			By("waiting for the envoy Deployment to be created")
			{
				dep := &appsv1.Deployment{}
				key := types.NamespacedName{Name: "marin3r-envoydeployment-instance", Namespace: namespace}
				Eventually(func() error {
					return k8sClient.Get(context.Background(), key, dep)
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
			}
		})

		It("creates HPA resource", func() {

			By("updating the EnvoyDeployment resource to use a dynamic number of replicas")
			{
				err := k8sClient.Get(context.Background(), types.NamespacedName{Name: "instance", Namespace: namespace}, ed)
				Expect(err).ToNot(HaveOccurred())
				patch := client.MergeFrom(ed.DeepCopy())
				ed.Spec.Replicas = &operatorv1alpha1.ReplicasSpec{
					Dynamic: &operatorv1alpha1.DynamicReplicasSpec{
						MaxReplicas: 5,
						Metrics: []autoscalingv2.MetricSpec{
							{
								Type: autoscalingv2.ResourceMetricSourceType,
								Resource: &autoscalingv2.ResourceMetricSource{
									Name: corev1.ResourceCPU,
									Target: autoscalingv2.MetricTarget{
										Type:               autoscalingv2.UtilizationMetricType,
										AverageUtilization: ptr.To(int32(50)),
									},
								},
							},
						},
					},
				}
				err = k8sClient.Patch(context.Background(), ed, patch)
				Expect(err).ToNot(HaveOccurred())
			}

			By("waiting for the envoy HPA to be created")
			{
				hpa := &autoscalingv2.HorizontalPodAutoscaler{}
				key := types.NamespacedName{Name: "marin3r-envoydeployment-instance", Namespace: namespace}
				Eventually(func() error {
					return k8sClient.Get(context.Background(), key, hpa)
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
			}
		})

		It("creates the PDB resource", func() {

			By("updating the EnvoyDeployment resource to use a PodDisruptionBudget")
			{
				err := k8sClient.Get(context.Background(), types.NamespacedName{Name: "instance", Namespace: namespace}, ed)
				Expect(err).ToNot(HaveOccurred())
				patch := client.MergeFrom(ed.DeepCopy())
				ed.Spec.PodDisruptionBudget = &operatorv1alpha1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
				}
				err = k8sClient.Patch(context.Background(), ed, patch)
				Expect(err).ToNot(HaveOccurred())
			}

			By("waiting for the envoy PDB to be created")
			{
				pdb := &policyv1.PodDisruptionBudget{}
				key := types.NamespacedName{Name: "marin3r-envoydeployment-instance", Namespace: namespace}
				Eventually(func() error {
					return k8sClient.Get(context.Background(), key, pdb)
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
			}
		})
	})

})
