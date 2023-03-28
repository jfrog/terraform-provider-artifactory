package security

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/jfrog/terraform-provider-shared/util"
)

const CertificateEndpoint = "artifactory/api/system/security/certificates/"

// CertificateDetails this type doesn't even exist in the new go client. In fact, the whole API call doesn't
type CertificateDetails struct {
	CertificateAlias string `json:"certificateAlias,omitempty"`
	IssuedTo         string `json:"issuedTo,omitempty"`
	IssuedBy         string `json:"issuedby,omitempty"`
	IssuedOn         string `json:"issuedOn,omitempty"`
	ValidUntil       string `json:"validUntil,omitempty"`
	FingerPrint      string `json:"fingerPrint,omitempty"`
}

func ResourceArtifactoryCertificate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCertificateCreate,
		ReadContext:   resourceCertificateRead,
		UpdateContext: resourceCertificateUpdate,
		DeleteContext: resourceCertificateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"alias": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"content": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				ValidateFunc: validation.All(
					validation.StringIsNotEmpty,
					func(value interface{}, key string) ([]string, []error) {
						_, err := extractCertificate(value.(string))
						if err != nil {
							return nil, []error{err}
						}
						return nil, nil
					},
				),
			},
			"file": {
				Type:         schema.TypeString,
				Sensitive:    true,
				Optional:     true,
				ExactlyOneOf: []string{"content", "file"},
				ValidateFunc: func(value interface{}, key string) ([]string, []error) {
					var errors []error
					if _, err := os.Stat(value.(string)); err != nil {
						return nil, append(errors, err)
					}
					data, err := ioutil.ReadFile(value.(string))
					if err != nil {
						return nil, append(errors, err)
					}
					_, err = extractCertificate(string(data))
					if err != nil {
						return nil, append(errors, err)
					}
					return nil, nil
				},
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
	content, err := getContentFromDiff(d)
	fingerprint, err := calculateFingerPrint(content)
	if err != nil {
		return err
	}
	if d.Get("fingerprint").(string) != fingerprint {
		if err = d.SetNewComputed("fingerprint"); err != nil {
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

func FindCertificate(alias string, m interface{}) (*CertificateDetails, error) {
	c := m.(util.ProvderMetadata).Client
	certificates := new([]CertificateDetails)
	_, err := c.R().SetResult(certificates).Get(CertificateEndpoint)

	if err != nil {
		return nil, err
	}

	// No way other than to loop through each certificate
	for _, cert := range *certificates {
		if cert.CertificateAlias == alias {
			return &cert, nil
		}
	}

	return nil, nil
}

func resourceCertificateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId(d.Get("alias").(string))
	return resourceCertificateUpdate(ctx, d, m)
}

func resourceCertificateRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cert, err := FindCertificate(d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}

	if cert != nil {
		setValue := util.MkLens(d)

		setValue("alias", (*cert).CertificateAlias)
		setValue("fingerprint", (*cert).FingerPrint)
		setValue("issued_by", (*cert).IssuedBy)
		setValue("issued_on", (*cert).IssuedOn)
		setValue("issued_to", (*cert).IssuedTo)
		errors := setValue("valid_until", (*cert).ValidUntil)

		if errors != nil && len(errors) > 0 {
			return diag.Errorf("failed to pack certificate %q", errors)
		}

		return nil
	}

	d.SetId("")

	return nil
}

func getContentFromDiff(d *schema.ResourceDiff) (string, error) {
	content, contentExists := d.GetOkExists("content")
	file, fileExists := d.GetOkExists("file")

	if contentExists == fileExists {
		return "", fmt.Errorf("you must define 'content' as the contents of the pem file, OR set 'file' to the path of your pem file ")
	}

	if contentExists {
		return content.(string), nil
	}
	if fileExists {
		data, err := ioutil.ReadFile(file.(string))
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return "", fmt.Errorf("mmm, couldn't get content or file. You need either a content or a file")
}

func getContentFromData(d *schema.ResourceData) (string, error) {

	if content, ok := d.GetOk("content"); ok {
		return content.(string), nil
	}
	if file, ok := d.GetOk("file"); ok {
		data, err := ioutil.ReadFile(file.(string))
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return "", fmt.Errorf("mmm, couldn't get content or file. You need either a content or a file")
}

func resourceCertificateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	content, err := getContentFromData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = m.(util.ProvderMetadata).Client.R().SetBody(content).SetHeader("content-type", "text/plain").Post(CertificateEndpoint + d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceCertificateRead(ctx, d, m)
}

func resourceCertificateDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	_, err := m.(util.ProvderMetadata).Client.R().Delete(CertificateEndpoint + d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
