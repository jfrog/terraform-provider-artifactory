package artifactory

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"
)

const keypairEndPoint = "artifactory/api/security/keypair/"

type KeyPairPayLoad struct {
	PairName    string `hcl:"pair_name" json:"pairName"`
	PairType    string `hcl:"pair_type" json:"pairType"`
	Alias       string `hcl:"alias" json:"alias"`
	PrivateKey  string `hcl:"private_key" json:"privateKey"`
	Passphrase  string `hcl:"passphrase" json:"passphrase"`
	PublicKey   string `hcl:"public_key" json:"publicKey"`
	Unavailable bool   `hcl:"unavailable" json:"unavailable"`
}

func resourceArtifactoryKeyPair() *schema.Resource {
	return &schema.Resource{
		CreateContext: createKeyPair,
		DeleteContext: rmKeyPair,
		ReadContext:   readKeyPair,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manage the GPG signing keys used to sign packages for authentication and the RSA keys used to sign and verify the Alpine Linux Index files\n" +
			"https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-CreateKeyPair",

		Schema: map[string]*schema.Schema{
			"pair_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"pair_type": {
				Type: schema.TypeString,
				// working sample PGP key is checked in but not tested
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"RSA", "GPG"}, false)),
				Required:         true,
				Description:      "Let's RT know what kind of key pair you're supplying. RT also supports GPG, but that's for a later day",
				ForceNew:         true,
			},
			"alias": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Will be used as a filename when retrieving the public key via REST API",
				ForceNew:    true,
			},
			"private_key": {
				Type:             schema.TypeString,
				Sensitive:        true,
				Required:         true,
				DiffSuppressFunc: stripTabs,
				ValidateDiagFunc: validatePrivateKey,
				ForceNew:         true,
			},
			"passphrase": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to decrypt the private key (if applicable). Will be verified server side",
				ForceNew:    true,
			},
			"public_key": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: stripTabs,
				ValidateDiagFunc: validatePublicKey,
				ForceNew:         true,
			},
			"unavailable": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Unknown usage. Returned in the json payload and cannot be set.",
			},
		},
	}
}
func validatePrivateKey(value interface{}, _ cty.Path) diag.Diagnostics {
	stripped := strings.ReplaceAll(value.(string), "\t", "")
	var err error
	// currently can't validate GPG
	if strings.Contains(stripped, "BEGIN PGP PRIVATE KEY BLOCK") {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "usage of GPG can't be validated.",
				Detail:   "Due to limitations of go libraries, your GPG key can't be validated client side",
			},
		}
	}
	privPem, _ := pem.Decode([]byte(stripped))
	if privPem == nil {
		return diag.Errorf("unable to decode private key pem format")
	}
	var privPemBytes []byte
	if privPem.Type != "RSA PRIVATE KEY" {
		return diag.Errorf("RSA private key is of the wrong type. Pem Type: %s", privPem.Type)
	}

	privPemBytes = privPem.Bytes
	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(privPemBytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(privPemBytes); err != nil { // note this returns type `interface{}`
			return diag.FromErr(err)
		}
	}

	_, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return diag.Errorf("unable to cast to RSA private key")
	}
	return nil
}

func validatePublicKey(value interface{}, path cty.Path) diag.Diagnostics {
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

func stripTabs(_, old, new string, _ *schema.ResourceData) bool {
	return old == strings.ReplaceAll(new, "\t", "")
}

func unpackKeyPair(s *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{s}
	result := KeyPairPayLoad{
		PairName:    d.getString("pair_name", false),
		PairType:    d.getString("pair_type", false),
		Alias:       d.getString("alias", false),
		PrivateKey:  strings.ReplaceAll(d.getString("private_key", false), "\t", ""),
		PublicKey:   strings.ReplaceAll(d.getString("public_key", false), "\t", ""),
		Unavailable: d.getBool("unavailable", false),
	}
	return &result, result.PairName, nil
}

func createKeyPair(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	keyPair, key, _ := unpackKeyPair(d)

	_, err := m.(*resty.Client).R().SetBody(keyPair).Post(keypairEndPoint)
	if err != nil {
		return diag.FromErr(err)
	}
	err = universalPack(*keyPair.(*KeyPairPayLoad), d)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(key)
	return nil
}

func readKeyPair(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	data := KeyPairPayLoad{}
	_, err := meta.(*resty.Client).R().SetResult(&data).Get(keypairEndPoint + d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = universalPack(data, d)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
func rmKeyPair(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	_, err := m.(*resty.Client).R().Delete(keypairEndPoint + d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func verifyKeyPair(id string, request *resty.Request) (*resty.Response, error) {
	return request.Head(keypairEndPoint + id)
}

func (kp KeyPairPayLoad) Id() string {
	return kp.PairName
}
