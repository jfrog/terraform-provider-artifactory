package httpclient

import (
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/httpclient"
)

func ArtifactoryClientBuilder() *artifactoryHttpClientBuilder {
	return &artifactoryHttpClientBuilder{}
}

type artifactoryHttpClientBuilder struct {
	certificatesDirPath string
	insecureTls         bool
	ServiceDetails      *auth.ServiceDetails
}

func (builder *artifactoryHttpClientBuilder) SetCertificatesPath(certificatesPath string) *artifactoryHttpClientBuilder {
	builder.certificatesDirPath = certificatesPath
	return builder
}

func (builder *artifactoryHttpClientBuilder) SetInsecureTls(insecureTls bool) *artifactoryHttpClientBuilder {
	builder.insecureTls = insecureTls
	return builder
}

func (builder *artifactoryHttpClientBuilder) SetServiceDetails(rtDetails *auth.ServiceDetails) *artifactoryHttpClientBuilder {
	builder.ServiceDetails = rtDetails
	return builder
}

func (builder *artifactoryHttpClientBuilder) Build() (rtHttpClient *ArtifactoryHttpClient, err error) {
	rtHttpClient = &ArtifactoryHttpClient{ArtDetails: builder.ServiceDetails}
	rtHttpClient.httpClient, err = httpclient.ClientBuilder().
		SetCertificatesPath(builder.certificatesDirPath).
		SetInsecureTls(builder.insecureTls).
		SetClientCertPath((*rtHttpClient.ArtDetails).GetClientCertPath()).
		SetClientCertKeyPath((*rtHttpClient.ArtDetails).GetClientCertKeyPath()).
		Build()
	return
}
