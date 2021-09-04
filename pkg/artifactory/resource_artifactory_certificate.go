package artifactory

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const endpoint = "artifactory/api/system/security/certificates/"

// CertificateDetails this type doesn't even exist in the new go client. In fact, the whole API call doesn't
type CertificateDetails struct {
	CertificateAlias string `json:"certificateAlias,omitempty"`
	IssuedTo         string `json:"issuedTo,omitempty"`
	IssuedBy         string `json:"issuedby,omitempty"`
	IssuedOn         string `json:"issuedOn,omitempty"`
	ValidUntil       string `json:"validUntil,omitempty"`
	FingerPrint      string `json:"fingerPrint,omitempty"`
}

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
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
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
				Type:          schema.TypeString,
				Sensitive:     true,
				Optional:      true,
				ExactlyOneOf: []string{"content", "file"},
				ValidateFunc: func(value interface{}, key string) ([]string, []error) {
					var errors []error
					if _, err := os.Stat(value.(string)); err != nil {
						return nil, append(errors,err)
					}
					data, err := ioutil.ReadFile(value.(string))
					if err != nil {
						return nil, append(errors,err)
					}
					_, err = extractCertificate(string(data))
					if err != nil {
						return nil, append(errors,err)
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

		CustomizeDiff: func(d *schema.ResourceDiff, _ interface{}) error {
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
		},
	}
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

func findCertificate(alias string, m interface{}) (*CertificateDetails, error) {
	c := m.(*resty.Client)
	certificates := new([]CertificateDetails)
	_, err := c.R().SetResult(certificates).Get(endpoint)

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

func resourceCertificateCreate(d *schema.ResourceData, m interface{}) error {
	d.SetId(d.Get("alias").(string))
	return resourceCertificateUpdate(d, m)
}

func resourceCertificateRead(d *schema.ResourceData, m interface{}) error {
	cert, err := findCertificate(d.Id(), m)
	if err != nil {
		return err
	}

	if cert != nil {
		setValue := mkLens(d)

		setValue("alias", (*cert).CertificateAlias)
		setValue("fingerprint", (*cert).FingerPrint)
		setValue("issued_by", (*cert).IssuedBy)
		setValue("issued_on", (*cert).IssuedOn)
		setValue("issued_to", (*cert).IssuedTo)
		errors := setValue("valid_until", (*cert).ValidUntil)

		if errors != nil && len(errors) > 0 {
			return fmt.Errorf("failed to pack certificate %q", errors)
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

	if content, ok := d.GetOkExists("content"); ok {
		return content.(string), nil
	}
	if file, ok := d.GetOkExists("file"); ok {
		data, err := ioutil.ReadFile(file.(string))
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return "", fmt.Errorf("mmm, couldn't get content or file. You need either a content or a file")
}

func resourceCertificateUpdate(d *schema.ResourceData, m interface{}) error {
	content, err := getContentFromData(d)
	if err != nil {
		return err
	}

	_, err = m.(*resty.Client).R().SetBody(content).SetHeader("content-type", "text/plain").Post(endpoint + d.Id())

	if err != nil {
		return err
	}

	return resourceCertificateRead(d, m)
}

func resourceCertificateDelete(d *schema.ResourceData, m interface{}) error {
	_, err := m.(*resty.Client).R().Delete(endpoint + d.Id())
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourceCertificateExists(d *schema.ResourceData, m interface{}) (bool, error) {
	cert, err := findCertificate(d.Id(), m)
	return err == nil && cert != nil, err
}
