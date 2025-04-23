package providers

import "context"

// CertificateProvider has methods to manage certificates using a given provider
type CertificateProvider interface {
	CreateCertificate(context.Context) ([]byte, []byte, error)
	GetCertificate(context.Context) ([]byte, []byte, error)
	UpdateCertificate(context.Context) ([]byte, []byte, error)
	VerifyCertificate(context.Context) error
}
