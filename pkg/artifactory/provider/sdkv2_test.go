package provider_test

import (
	"testing"

	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/provider"
)

func TestProvider(t *testing.T) {
	if err := provider.SdkV2().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ = provider.SdkV2()
}
