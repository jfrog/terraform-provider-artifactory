package provider

import (
	"crypto/tls"
	"fmt"
	"strings"
)

type tlsConfigOptions struct {
	ClientCertificatePath    string
	ClientCertificateKeyPath string
	ClientCertificatePEM     string
	ClientPrivateKeyPEM      string
	InsecureSkipVerify       bool
}

func (o *tlsConfigOptions) normalize() {
	o.ClientCertificatePEM = strings.TrimSpace(o.ClientCertificatePEM)
	o.ClientPrivateKeyPEM = strings.TrimSpace(o.ClientPrivateKeyPEM)
}

func buildTLSConfig(opts tlsConfigOptions) (*tls.Config, error) {
	opts.normalize()

	usePath := opts.ClientCertificatePath != "" || opts.ClientCertificateKeyPath != ""
	usePEM := opts.ClientCertificatePEM != "" || opts.ClientPrivateKeyPEM != ""

	if usePath && usePEM {
		return nil, fmt.Errorf("cannot configure both path-based and inline client certificate options")
	}

	var cert tls.Certificate
	var haveCert bool

	if usePEM {
		if opts.ClientCertificatePEM == "" || opts.ClientPrivateKeyPEM == "" {
			return nil, fmt.Errorf("both client_certificate_pem and client_private_key_pem must be provided together")
		}

		parsed, err := tls.X509KeyPair([]byte(opts.ClientCertificatePEM), []byte(opts.ClientPrivateKeyPEM))
		if err != nil {
			return nil, fmt.Errorf("failed to parse client certificate PEM data: %w", err)
		}
		cert = parsed
		haveCert = true
	} else if usePath {
		if opts.ClientCertificatePath == "" || opts.ClientCertificateKeyPath == "" {
			return nil, fmt.Errorf("both client_certificate_path and client_certificate_key_path must be provided together")
		}

		parsed, err := tls.LoadX509KeyPair(opts.ClientCertificatePath, opts.ClientCertificateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate from path: %w", err)
		}
		cert = parsed
		haveCert = true
	}

	if !haveCert && !opts.InsecureSkipVerify {
		return nil, nil
	}

	tlsConfig := &tls.Config{}
	if haveCert {
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	if opts.InsecureSkipVerify {
		tlsConfig.InsecureSkipVerify = true
	}

	return tlsConfig, nil
}
