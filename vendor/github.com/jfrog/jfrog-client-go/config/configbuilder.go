package config

import (
	"github.com/jfrog/jfrog-client-go/auth"
)

func NewConfigBuilder() *servicesConfigBuilder {
	configBuilder := &servicesConfigBuilder{}
	configBuilder.threads = 3
	return configBuilder
}

type servicesConfigBuilder struct {
	auth.ServiceDetails
	certificatesPath string
	threads          int
	isDryRun         bool
	insecureTls      bool
}

func (builder *servicesConfigBuilder) SetServiceDetails(artDetails auth.ServiceDetails) *servicesConfigBuilder {
	builder.ServiceDetails = artDetails
	return builder
}

func (builder *servicesConfigBuilder) SetCertificatesPath(certificatesPath string) *servicesConfigBuilder {
	builder.certificatesPath = certificatesPath
	return builder
}

func (builder *servicesConfigBuilder) SetThreads(threads int) *servicesConfigBuilder {
	builder.threads = threads
	return builder
}

func (builder *servicesConfigBuilder) SetDryRun(dryRun bool) *servicesConfigBuilder {
	builder.isDryRun = dryRun
	return builder
}

func (builder *servicesConfigBuilder) SetInsecureTls(insecureTls bool) *servicesConfigBuilder {
	builder.insecureTls = insecureTls
	return builder
}

func (builder *servicesConfigBuilder) Build() (Config, error) {
	c := &servicesConfig{}
	c.ServiceDetails = builder.ServiceDetails
	c.threads = builder.threads
	c.certificatesPath = builder.certificatesPath
	c.dryRun = builder.isDryRun
	c.insecureTls = builder.insecureTls
	return c, nil
}
