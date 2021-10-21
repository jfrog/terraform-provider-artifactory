package artifactory

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

const keypairEndPoint = "artifactory/api/security/keypair/"

type KeyPairPayLoad struct {
	PairName    string `hcl:"pair_name" json:"pairName"`
	PairType    string `hcl:"pair_type" json:"pairType"`
	Alias       string `hcl:"alias" json:"alias"`
	PrivateKey  string `hcl:"private_key" json:"privateKey"`
	PublicKey   string `hcl:"public_key" json:"publicKey"`
	Unavailable bool   `hcl:"unavailable" json:"unavailable"`
}

func resourceArtifactoryKeyPair() *schema.Resource {
	return &schema.Resource{
		CreateContext: createKeyPair,
		DeleteContext: rmKeyPair,
		ReadContext:   readKeyPair,
		UpdateContext: func(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
			return diag.Errorf("please implement me")
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"pair_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"pair_type": {
				Type: schema.TypeString,
				//ValidateDiagFunc: upgrade(validation.StringInSlice([]string{"rsa", "gpg"}, false), "pair_type"),
				Required: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Required: true,
			},
			"private_key": {
				Type:         schema.TypeString,
				Sensitive:    true,
				Optional:     true,
				DiffSuppressFunc: stripTabs,
				ExactlyOneOf: []string{"private_key_file", "private_key"},
			},
			"private_key_file": {
				Type:         schema.TypeString,
				Sensitive:    true,
				Optional:     true,
				ExactlyOneOf: []string{"private_key_file", "private_key"},
			},
			"public_key": {
				Type:     schema.TypeString,
				Required: true,
				DiffSuppressFunc: stripTabs,
			},
			"unavailable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default: false,
			},
		},
	}
}
func stripTabs(_, old, new string, _ *schema.ResourceData) bool {
	return old == strings.ReplaceAll(new,"\t","")
}
func packKeyPair(kp KeyPairPayLoad, d *schema.ResourceData) error {

	setValue := mkLens(d)

	setValue("pair_name", kp.PairName)
	setValue("pair_type", kp.PairType)
	setValue("alias", kp.Alias)
	setValue("unavailable", kp.Unavailable)
	setValue("private_key", strings.ReplaceAll(kp.PrivateKey,"\t",""))
	errors := setValue("public_key", strings.ReplaceAll(kp.PublicKey,"\t",""))

	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed to pack keypair %q", errors)
	}

	return nil
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
	err = packKeyPair(*keyPair.(*KeyPairPayLoad), d)
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
	err = packKeyPair(data, d)
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
