package pki

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

// LoadX509Certificate loads a x509.Certificate object from the given bytes
func LoadX509Certificate(cert []byte) (*x509.Certificate, error) {
	cpb, _ := pem.Decode(cert)

	crt, err := x509.ParseCertificate(cpb.Bytes)
	if err != nil {
		return nil, err
	}

	return crt, nil
}

// DecodePrivateKeyBytes will decode a PEM encoded private key into a crypto.Signer.
// It supports RSA private keys only. All other types will return err.
func DecodePrivateKeyBytes(keyBytes []byte) (crypto.Signer, error) {
	// decode the private key pem
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("error decoding private key PEM block")
	}

	switch block.Type {
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing pkcs#8 private key: %s", err.Error())
		}

		signer, ok := key.(crypto.Signer)
		if !ok {
			return nil, errors.New("error parsing pkcs#8 private key: invalid key type")
		}

		return signer, nil
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing rsa private key: %s", err.Error())
		}

		err = key.Validate()
		if err != nil {
			return nil, fmt.Errorf("rsa private key failed validation: %s", err.Error())
		}

		return key, nil

	default:
		return nil, fmt.Errorf("unknown private key type: %s", block.Type)
	}
}
