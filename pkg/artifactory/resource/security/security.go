package security

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func VerifyKeyPair(id string, request *resty.Request) (*resty.Response, error) {
	return request.Head(KeypairEndPoint + id)
}

func validatePublicKey(value interface{}, _ cty.Path) diag.Diagnostics {
	var err error

	stripped := strings.ReplaceAll(value.(string), "\t", "")
	// currently can't validate GPG
	if strings.Contains(stripped, "BEGIN PGP PUBLIC KEY BLOCK") {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "usage of GPG can't be validated.",
				Detail:   "Due to limitations of go libraries, your GPG key can't be validated client side",
			},
		}
	}
	pubPem, _ := pem.Decode([]byte(stripped))
	if pubPem == nil {
		return diag.Errorf("rsa public key not in pem format")
	}
	if !strings.Contains(pubPem.Type, "PUBLIC KEY") {
		return diag.Errorf("RSA public key is of the wrong type and must container the header 'PUBLIC KEY': Pem Type: %s ", pubPem.Type)
	}
	var parsedKey interface{}

	if parsedKey, err = x509.ParsePKIXPublicKey(pubPem.Bytes); err != nil {
		return diag.Errorf("unable to parse RSA public key")
	}

	if _, ok := parsedKey.(*rsa.PublicKey); !ok {
		return diag.Errorf("unable to cast to RSA public key")
	}

	return nil
}

func stripTabs(val interface{}) string {
	return strings.ReplaceAll(val.(string), "\t", "")
}
