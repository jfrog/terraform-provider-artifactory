package artifactory

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"

	v1 "github.com/atlassian/go-artifactory/v2/artifactory/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArtifactoryCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceCertificateCreate,
		Read:   resourceCertificateRead,
		Update: resourceCertificateUpdate,
		Delete: resourceCertificateDelete,
		Exists: resourceCertificateExists,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"alias": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"content": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"issued_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"issued_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"issued_to": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"valid_until": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: calculateFingerprint,
	}
}

func calculateFingerprint(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	fingerprint, err := calculateFingerPrint(d.Get("content").(string))
	if err != nil {
		return err
	}
	if d.Get("fingerprint").(string) != fingerprint {
		if err := d.SetNewComputed("fingerprint"); err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func formatFingerPrint(f []byte) string {
	buf := make([]byte, 0, 3*len(f))
	x := buf[1*len(f) : 3*len(f)]
	hex.Encode(x, f)
	for i := 0; i < len(x); i += 2 {
		buf = append(buf, x[i], x[i+1], ':')
	}
	return strings.ToUpper(string(buf[:len(buf)-1]))
}

func extractCertificate(pemData string) (*x509.Certificate, error) {
	block, rest := pem.Decode([]byte(pemData))
	for block != nil {
		if block.Type != "CERTIFICATE" {
			block, rest = pem.Decode(rest)
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}

		return cert, nil
	}

	return nil, fmt.Errorf("no certificate in PEM data")
}

func calculateFingerPrint(pemData string) (string, error) {
	cert, err := extractCertificate(pemData)
	if err != nil {
		return "", err
	}

	fingerprint := sha256.Sum256(cert.Raw)

	return formatFingerPrint(fingerprint[:]), nil
}

func findCertificate(d *schema.ResourceData, m interface{}) (*v1.CertificateDetails, error) {
	c := m.(*ArtClient).ArtOld

	certs, _, err := c.V1.Security.GetCertificates(context.Background())
	if err != nil {
		return nil, err
	}

	// No way other than to loop through each certificate
	for _, cert := range *certs {
		if *cert.CertificateAlias == d.Id() {
			return &cert, nil
		}
	}

	return nil, nil
}

func resourceCertificateCreate(d *schema.ResourceData, m interface{}) error {
	d.SetId(d.Get("alias").(string))
	return resourceCertificateUpdate(d, m)
}

func resourceCertificateRead(d *schema.ResourceData, m interface{}) error {
	cert, err := findCertificate(d, m)
	if err != nil {
		return err
	}

	if cert != nil {
		hasErr := false
		logErr := cascadingErr(&hasErr)

		logErr(d.Set("alias", *cert.CertificateAlias))
		logErr(d.Set("fingerprint", *cert.FingerPrint))
		logErr(d.Set("issued_by", *cert.IssuedBy))
		logErr(d.Set("issued_on", *cert.IssuedOn))
		logErr(d.Set("issued_to", *cert.IssuedTo))
		logErr(d.Set("valid_until", *cert.ValidUntil))

		if hasErr {
			return fmt.Errorf("failed to pack certificate")
		}

		return nil
	}

	d.SetId("")

	return nil
}

func resourceCertificateUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	_, _, err := c.V1.Security.AddCertificate(context.Background(), d.Id(), strings.NewReader(d.Get("content").(string)))
	if err != nil {
		return err
	}

	return resourceCertificateRead(d, m)
}

func resourceCertificateDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	_, _, err := c.V1.Security.DeleteCertificate(context.Background(), d.Id())
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourceCertificateExists(d *schema.ResourceData, m interface{}) (bool, error) {
	cert, err := findCertificate(d, m)
	if err != nil {
		return false, err
	}

	if cert != nil {
		return true, nil
	}

	return false, nil
}
