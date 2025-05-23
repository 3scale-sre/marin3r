package controllers

import (
	"context"
	"time"

	"github.com/3scale-sre/marin3r/api/envoy/defaults"
	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("DiscoveryService controller", func() {
	var namespace string
	var ds *operatorv1alpha1.DiscoveryService

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
		ds = &operatorv1alpha1.DiscoveryService{
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

	Context("DiscoveryService", func() {

		It("creates the required resources", func() {

			By("waiting for the root CA DiscoveryServiceCertificate to be created")
			{
				dsc := &operatorv1alpha1.DiscoveryServiceCertificate{}
				Eventually(func() error {
					return k8sClient.Get(
						context.Background(),
						types.NamespacedName{Name: "marin3r-ca-cert-instance", Namespace: namespace},
						dsc,
					)
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())

				Expect(dsc.Spec.SecretRef.Name).To(Equal(ds.GetRootCertificateAuthorityOptions().SecretName))
				Expect(dsc.Spec.ValidFor).To(Equal(int64(ds.GetRootCertificateAuthorityOptions().Duration.Seconds())))
			}

			By("waiting for the server DiscoveryServiceCertificate to be created")
			{
				dsc := &operatorv1alpha1.DiscoveryServiceCertificate{}
				Eventually(func() error {
					return k8sClient.Get(
						context.Background(),
						types.NamespacedName{Name: "marin3r-server-cert-instance", Namespace: namespace},
						dsc,
					)
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())

				Expect(dsc.Spec.SecretRef.Name).To(Equal(ds.GetServerCertificateOptions().SecretName))
				Expect(dsc.Spec.ValidFor).To(Equal(int64(ds.GetServerCertificateOptions().Duration.Seconds())))
			}

			By("waiting for the discovery service ServiceAccount to be created")
			{
				sa := &corev1.ServiceAccount{}
				Eventually(func() error {
					return k8sClient.Get(
						context.Background(),
						types.NamespacedName{Name: "marin3r-instance", Namespace: namespace},
						sa,
					)
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
			}

			By("waiting for the discovery service Role to be created")
			{
				cr := &rbacv1.Role{}
				Eventually(func() error {
					return k8sClient.Get(
						context.Background(),
						types.NamespacedName{Name: "marin3r-instance", Namespace: namespace},
						cr,
					)
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
			}

			By("waiting for the discovery service RoleBinding to be created")
			{
				crb := &rbacv1.RoleBinding{}
				Eventually(func() error {
					return k8sClient.Get(
						context.Background(),
						types.NamespacedName{Name: "marin3r-instance", Namespace: namespace},
						crb,
					)
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
			}

			By("waiting for the discovery service Deployment to be created")
			{
				dep := &appsv1.Deployment{}
				key := types.NamespacedName{Name: "marin3r-instance", Namespace: namespace}
				Eventually(func() bool {
					if err := k8sClient.Get(context.Background(), key, dep); err != nil {
						return false
					}
					hash, ok := dep.Spec.Template.Labels[operatorv1alpha1.DiscoveryServiceCertificateHashLabelKey]
					if !ok || hash == "" {
						return false
					}

					return true
				}, 60*time.Second, 5*time.Second).Should(BeTrue())
			}

			By("waiting for the discovery service Service to be created")
			{
				svc := &corev1.Service{}
				Eventually(func() error {
					return k8sClient.Get(
						context.Background(),
						types.NamespacedName{Name: ds.GetServiceConfig().Name, Namespace: namespace},
						svc,
					)
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
			}

			By("waiting for the client certificate resource to be created")
			{
				eb := &operatorv1alpha1.DiscoveryServiceCertificate{}
				Eventually(func() error {
					return k8sClient.Get(
						context.Background(),
						types.NamespacedName{Name: defaults.SidecarClientCertificate, Namespace: namespace},
						eb,
					)
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
			}
		})
	})

})
