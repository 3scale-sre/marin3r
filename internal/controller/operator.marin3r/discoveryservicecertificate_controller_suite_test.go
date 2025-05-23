package controllers

import (
	"context"
	"time"

	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	"github.com/3scale-sre/marin3r/internal/pkg/util/pki"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("DiscoveryServiceCertificate controller", func() {
	var namespace string
	var dsc *operatorv1alpha1.DiscoveryServiceCertificate

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
		Eventually(func() bool {
			err := k8sClient.Get(context.Background(), types.NamespacedName{Name: namespace}, n)

			return err == nil
		}, 60*time.Second, 5*time.Second).Should(BeTrue())

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

	Context("self-signed", func() {

		BeforeEach(func() {
			By("creating a DiscoveryServiceCertificate instance")
			dsc = &operatorv1alpha1.DiscoveryServiceCertificate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dsc",
					Namespace: namespace,
				},
				Spec: operatorv1alpha1.DiscoveryServiceCertificateSpec{
					CommonName: "test",
					ValidFor:   10,
					Hosts:      []string{"example.test"},
					Signer: operatorv1alpha1.DiscoveryServiceCertificateSigner{
						SelfSigned: &operatorv1alpha1.SelfSignedConfig{},
					},
					SecretRef: corev1.SecretReference{Name: "secret"},
				},
			}
			err := k8sClient.Create(context.Background(), dsc)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() error {
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: "dsc", Namespace: namespace}, dsc)
			}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())

			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), types.NamespacedName{Name: "dsc", Namespace: namespace}, dsc)
				Expect(err).ToNot(HaveOccurred())

				return dsc.Status.IsReady()
			}, 60*time.Second, 5*time.Second).Should(BeTrue())
		})

		It("creates a valid certificate within a Secret", func() {

			secret := &corev1.Secret{}
			err := k8sClient.Get(context.Background(), types.NamespacedName{Name: "secret", Namespace: namespace}, secret)
			Expect(err).ToNot(HaveOccurred())
			Expect(secret.Type).To(Equal(corev1.SecretTypeTLS))
			cert, err := pki.LoadX509Certificate(secret.Data["tls.crt"])
			Expect(err).ToNot(HaveOccurred())
			err = pki.Verify(cert, cert)
			Expect(err).ToNot(HaveOccurred())
		})

		It("renews the certificate", func() {
			hash := dsc.Status.GetCertificateHash()
			Eventually(func() string {
				err := k8sClient.Get(context.Background(), types.NamespacedName{Name: "dsc", Namespace: namespace}, dsc)
				Expect(err).ToNot(HaveOccurred())

				return dsc.Status.GetCertificateHash()
			}, 60*time.Second, 5*time.Second).ShouldNot(Equal(hash))
		})

	})

	Context("ca-signed", func() {
		caSecret := &corev1.Secret{}
		certSecret := &corev1.Secret{}

		BeforeEach(func() {
			By("creating a DiscoveryServiceCertificate instance for the CA")
			ca := &operatorv1alpha1.DiscoveryServiceCertificate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ca",
					Namespace: namespace,
				},
				Spec: operatorv1alpha1.DiscoveryServiceCertificateSpec{
					CommonName: "test",
					IsCA:       ptr.To(true),
					ValidFor:   3600,
					Signer: operatorv1alpha1.DiscoveryServiceCertificateSigner{
						SelfSigned: &operatorv1alpha1.SelfSignedConfig{},
					},
					SecretRef: corev1.SecretReference{Name: "ca"},
				},
			}
			err := k8sClient.Create(context.Background(), ca)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() error {
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: "ca", Namespace: namespace}, caSecret)
			}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
		})

		It("creates a valid certificate within a Secret, signed by the ca", func() {

			By("creating a DiscoveryServiceCertificate instance for the certificate")
			dsc = &operatorv1alpha1.DiscoveryServiceCertificate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cert",
					Namespace: namespace,
				},
				Spec: operatorv1alpha1.DiscoveryServiceCertificateSpec{
					CommonName: "test",
					ValidFor:   3600,
					Signer: operatorv1alpha1.DiscoveryServiceCertificateSigner{
						CASigned: &operatorv1alpha1.CASignedConfig{
							SecretRef: corev1.SecretReference{Name: "ca", Namespace: namespace},
						},
					},
					SecretRef: corev1.SecretReference{Name: "cert"},
				},
			}
			err := k8sClient.Create(context.Background(), dsc)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() error {
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: "cert", Namespace: namespace}, certSecret)
			}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())

			ca, err := pki.LoadX509Certificate(caSecret.Data["tls.crt"])
			Expect(err).ToNot(HaveOccurred())

			cert, err := pki.LoadX509Certificate(certSecret.Data["tls.crt"])
			Expect(err).ToNot(HaveOccurred())

			err = pki.Verify(cert, ca)
			Expect(err).ToNot(HaveOccurred())
		})

		It("re-issues the certificate if the CA changes", func() {

			By("creating a DiscoveryServiceCertificate instance for the certificate")
			dsc = &operatorv1alpha1.DiscoveryServiceCertificate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cert",
					Namespace: namespace,
				},
				Spec: operatorv1alpha1.DiscoveryServiceCertificateSpec{
					CommonName: "test",
					ValidFor:   3600,
					Signer: operatorv1alpha1.DiscoveryServiceCertificateSigner{
						CASigned: &operatorv1alpha1.CASignedConfig{
							SecretRef: corev1.SecretReference{Name: "ca", Namespace: namespace},
						},
					},
					SecretRef: corev1.SecretReference{Name: "cert"},
				},
			}
			err := k8sClient.Create(context.Background(), dsc)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), types.NamespacedName{Name: "cert", Namespace: namespace}, dsc)

				return err == nil && dsc.Status.IsReady()
			}, 60*time.Second, 5*time.Second).Should(BeTrue())

			hash := dsc.Status.GetCertificateHash()

			By("forcing recreation of the CA certificate")
			err = k8sClient.Delete(context.Background(), caSecret)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() string {
				err := k8sClient.Get(context.Background(), types.NamespacedName{Name: "cert", Namespace: namespace}, dsc)
				Expect(err).ToNot(HaveOccurred())

				return dsc.Status.GetCertificateHash()
			}, 60*time.Second, 5*time.Second).ShouldNot(Equal(hash))
		})

	})

})
