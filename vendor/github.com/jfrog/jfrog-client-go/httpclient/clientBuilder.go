package httpclient

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/jfrog/jfrog-client-go/auth/cert"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

func ClientBuilder() *httpClientBuilder {
	return &httpClientBuilder{}
}

type httpClientBuilder struct {
	certificatesDirPath string
	clientCertPath      string
	clientCertKeyPath   string
	insecureTls         bool
}

func (builder *httpClientBuilder) SetCertificatesPath(certificatesPath string) *httpClientBuilder {
	builder.certificatesDirPath = certificatesPath
	return builder
}

func (builder *httpClientBuilder) SetClientCertPath(certificatePath string) *httpClientBuilder {
	builder.clientCertPath = certificatePath
	return builder
}

func (builder *httpClientBuilder) SetClientCertKeyPath(certificatePath string) *httpClientBuilder {
	builder.clientCertKeyPath = certificatePath
	return builder
}

func (builder *httpClientBuilder) SetInsecureTls(insecureTls bool) *httpClientBuilder {
	builder.insecureTls = insecureTls
	return builder
}

func (builder *httpClientBuilder) AddClientCertToTransport(transport *http.Transport) error {
	if builder.clientCertPath != "" {
		cert, err := tls.LoadX509KeyPair(builder.clientCertPath, builder.clientCertKeyPath)
		if err != nil {
			return errorutils.CheckError(errors.New("Failed loading client certificate: " + err.Error()))
		}
		transport.TLSClientConfig.Certificates = []tls.Certificate{cert}
	}

	return nil
}

func (builder *httpClientBuilder) Build() (*HttpClient, error) {
	if builder.certificatesDirPath == "" {
		transport := createDefaultHttpTransport()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: builder.insecureTls}
		err := builder.AddClientCertToTransport(transport)
		if err != nil {
			return nil, err
		}
		return &HttpClient{Client: &http.Client{Transport: transport}}, nil
	}

	transport, err := cert.GetTransportWithLoadedCert(builder.certificatesDirPath, builder.insecureTls, createDefaultHttpTransport())
	if err != nil {
		return nil, errorutils.CheckError(errors.New("Failed creating HttpClient: " + err.Error()))
	}
	err = builder.AddClientCertToTransport(transport)
	if err != nil {
		return nil, err
	}
	return &HttpClient{Client: &http.Client{Transport: transport}}, nil
}

func createDefaultHttpTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 20 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
