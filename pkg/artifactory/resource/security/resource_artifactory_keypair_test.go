package security_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccKeyPairFailPrivateCertCheck(t *testing.T) {
	id, fqrn, name := test.MkNames("mykp", "artifactory_keypair")
	keyBasic := fmt.Sprintf(`
		resource "artifactory_keypair" "%s" {
			pair_name  = "%s"
			pair_type = "RSA"
			alias = "foo-alias%d"
			private_key = "not a private key"
			public_key = <<EOF
		-----BEGIN PUBLIC KEY-----
		MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2ymVc24BoaZb0ElXoI3X
		4zUKJGZEetR6F4yT1tJhkPDg7UTmiNoFB5TZJvP6LBrrSwszkpZbxaVOkBrwrGbq
		FUaXPgud8VabfHl0imXvN746zmpjYEMGqJzm+Gh6aBWOlnPdLuHhds/kcanFAEpp
		j5yN0tVWDnqjOJjR7EpxMSdP3TSd6tNAY73LGNLNJQc6tSxh8nMIb4HhSWQSgfof
		+FwcLGvs+mmyBq8Jz5Zy4BSCk1fQFmCnSGyzpyBD0vMd6eLHk2l0tm56DrlonbDM
		X8KGs7e9ZgjANkT5PnipLOaeLJU4H+OWyBZUAT4hl/iRVvLwV81x7g0/O38kmPYJ
		DQIDAQAB
		-----END PUBLIC KEY-----
		EOF
		}
	`, name, name, id)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, security.VerifyKeyPair),
		Steps: []resource.TestStep{
			{
				Config:      keyBasic,
				ExpectError: regexp.MustCompile(".*unable to decode private key pem format.*"),
			},
		},
	})
}

func TestAccKeyPairFailPubCertCheck(t *testing.T) {
	id, fqrn, name := test.MkNames("mykp", "artifactory_keypair")
	keyBasic := fmt.Sprintf(`
		resource "artifactory_keypair" "%s" {
			pair_name  = "%s"
			pair_type = "RSA"
			alias = "foo-alias%d"
			private_key = <<EOF
		-----BEGIN RSA PRIVATE KEY-----
		MIIEpAIBAAKCAQEA2ymVc24BoaZb0ElXoI3X4zUKJGZEetR6F4yT1tJhkPDg7UTm
		iNoFB5TZJvP6LBrrSwszkpZbxaVOkBrwrGbqFUaXPgud8VabfHl0imXvN746zmpj
		YEMGqJzm+Gh6aBWOlnPdLuHhds/kcanFAEppj5yN0tVWDnqjOJjR7EpxMSdP3TSd
		6tNAY73LGNLNJQc6tSxh8nMIb4HhSWQSgfof+FwcLGvs+mmyBq8Jz5Zy4BSCk1fQ
		FmCnSGyzpyBD0vMd6eLHk2l0tm56DrlonbDMX8KGs7e9ZgjANkT5PnipLOaeLJU4
		H+OWyBZUAT4hl/iRVvLwV81x7g0/O38kmPYJDQIDAQABAoIBAFb7wyhEIfuhhlE9
		uryrb2LrGzJlMIq7qBWOouKhLz4SjIM/VGw+c76VkjZGoSU+LeLj+D0W1ie0u2Cw
		gJM8aW22TbK/c5lksWOO5PVFDdPG+ZoRWY3MLGlhlL5E4UhMPgJyy/eeiRjZ3CZM
		pja+UfVAwn1KVNR8UinVZYPt680AvEd1FGxoNLxemIPNug46nNqp6Al86Bn+BnkN
		GXpwyooXVSfo4k+pnFBFIXUdA1dUEXQSVb1AxlTo6g/Ok/+8Gfq8idCdu+5fcZI2
		1eLeC+FAa92rr1SFX/UWeB4cMyuAqwuxbFFIplT22SaUSsNuOUSH4E00nbP/AuCb
		1BqrLmECgYEA7tQKfyF9aiXTsOMdOnSAa5OyEaCfsFtcmLd4ykVrwN8O36NoX005
		VbGuqo87fwIXQIKHU+kOEs/TmaQ8bNcbCD/SfWGTtOnHG4qfIsepJuoMwbQHRhGF
		JeoXh5yEUKg5pcDBY8PENEtEVKmFuL4bPOdn+9FNLGsjftvXpmGWbGUCgYEA6uuQ
		7kzO6WD88OsxdJzlJM11hg2SaSBCh3+5tnOhF1ULOUt4tdYXzh3QI6BPX7tkArYf
		XteVfWoWqn6T7LtCjFm350BqVpPhqfLKnt6fYf1yotsj/cyZXlXquRbxbgakB0n0
		4PrsZaube0TPPVeirzNyOVHyFc+iW+F+IUYD+4kCgYEApDEjBkP/9PoMj4+UiJuP
		rmXcBkJnhtdI0bVRVb5kVjUEBLxTBTISONfvPVM7lBXb5n3Wi9mt00EOOJKw+CLq
		csFt9MUgxz/xov2qaj7aC+bc3k7msUVaRLardpAkZ09AUrQyQGRWf50/XPUu+dO4
		5iYxVu6OH/uIa664k6qDwAECgYEAslS8oomgEL3VhbWkx1dLA5MMggTPfgpFNsMY
		4Y4JXcLrUEUgjzjEvW0YUdMiLhP8qapDSiXxj1D3f9myxWSp8g0xc9UMZEjCZ9at
		RcjNyP8zBLnCKqokSt6B3puyDsnvvrC/ugIBbnTFBOCJSZG7J7CwJx8z3KbQI1ub
		+fpCj7ECgYAd69soLEybUGMjsdI+OijIGoUTUoZGXJm+0VpBt4QJCe7AMnYPfYzA
		JnEmN4D7HLTKUBklQnb/FhP/RuiT2bSAd1l+PNeuU7gYROCBBonzxXQ1wEbNrSYA
		iyoc9g/kvV8HTW8361xEhu7wmuSEEx1gQ/7sdhTkgrTncc8hxVRxuA==
		-----END RSA PRIVATE KEY-----
		EOF
			public_key = "not a key"
		}
	`, name, name, id)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, security.VerifyKeyPair),
		Steps: []resource.TestStep{
			{
				Config:      keyBasic,
				ExpectError: regexp.MustCompile(".*rsa public key not in pem format.*"),
			},
		},
	})
}

func TestAccKeyPairRSA(t *testing.T) {
	id, fqrn, name := test.MkNames("mykp", "artifactory_keypair")
	template := `
	resource "artifactory_keypair" "{{ .name }}" {
		pair_name  = "{{ .name }}"
		pair_type = "RSA"
		alias = "foo-alias{{ .id }}"
		passphrase = "{{ .passphrase }}"
		private_key = <<EOF
	-----BEGIN RSA PRIVATE KEY-----
	MIIEpAIBAAKCAQEA2ymVc24BoaZb0ElXoI3X4zUKJGZEetR6F4yT1tJhkPDg7UTm
	iNoFB5TZJvP6LBrrSwszkpZbxaVOkBrwrGbqFUaXPgud8VabfHl0imXvN746zmpj
	YEMGqJzm+Gh6aBWOlnPdLuHhds/kcanFAEppj5yN0tVWDnqjOJjR7EpxMSdP3TSd
	6tNAY73LGNLNJQc6tSxh8nMIb4HhSWQSgfof+FwcLGvs+mmyBq8Jz5Zy4BSCk1fQ
	FmCnSGyzpyBD0vMd6eLHk2l0tm56DrlonbDMX8KGs7e9ZgjANkT5PnipLOaeLJU4
	H+OWyBZUAT4hl/iRVvLwV81x7g0/O38kmPYJDQIDAQABAoIBAFb7wyhEIfuhhlE9
	uryrb2LrGzJlMIq7qBWOouKhLz4SjIM/VGw+c76VkjZGoSU+LeLj+D0W1ie0u2Cw
	gJM8aW22TbK/c5lksWOO5PVFDdPG+ZoRWY3MLGlhlL5E4UhMPgJyy/eeiRjZ3CZM
	pja+UfVAwn1KVNR8UinVZYPt680AvEd1FGxoNLxemIPNug46nNqp6Al86Bn+BnkN
	GXpwyooXVSfo4k+pnFBFIXUdA1dUEXQSVb1AxlTo6g/Ok/+8Gfq8idCdu+5fcZI2
	1eLeC+FAa92rr1SFX/UWeB4cMyuAqwuxbFFIplT22SaUSsNuOUSH4E00nbP/AuCb
	1BqrLmECgYEA7tQKfyF9aiXTsOMdOnSAa5OyEaCfsFtcmLd4ykVrwN8O36NoX005
	VbGuqo87fwIXQIKHU+kOEs/TmaQ8bNcbCD/SfWGTtOnHG4qfIsepJuoMwbQHRhGF
	JeoXh5yEUKg5pcDBY8PENEtEVKmFuL4bPOdn+9FNLGsjftvXpmGWbGUCgYEA6uuQ
	7kzO6WD88OsxdJzlJM11hg2SaSBCh3+5tnOhF1ULOUt4tdYXzh3QI6BPX7tkArYf
	XteVfWoWqn6T7LtCjFm350BqVpPhqfLKnt6fYf1yotsj/cyZXlXquRbxbgakB0n0
	4PrsZaube0TPPVeirzNyOVHyFc+iW+F+IUYD+4kCgYEApDEjBkP/9PoMj4+UiJuP
	rmXcBkJnhtdI0bVRVb5kVjUEBLxTBTISONfvPVM7lBXb5n3Wi9mt00EOOJKw+CLq
	csFt9MUgxz/xov2qaj7aC+bc3k7msUVaRLardpAkZ09AUrQyQGRWf50/XPUu+dO4
	5iYxVu6OH/uIa664k6qDwAECgYEAslS8oomgEL3VhbWkx1dLA5MMggTPfgpFNsMY
	4Y4JXcLrUEUgjzjEvW0YUdMiLhP8qapDSiXxj1D3f9myxWSp8g0xc9UMZEjCZ9at
	RcjNyP8zBLnCKqokSt6B3puyDsnvvrC/ugIBbnTFBOCJSZG7J7CwJx8z3KbQI1ub
	+fpCj7ECgYAd69soLEybUGMjsdI+OijIGoUTUoZGXJm+0VpBt4QJCe7AMnYPfYzA
	JnEmN4D7HLTKUBklQnb/FhP/RuiT2bSAd1l+PNeuU7gYROCBBonzxXQ1wEbNrSYA
	iyoc9g/kvV8HTW8361xEhu7wmuSEEx1gQ/7sdhTkgrTncc8hxVRxuA==
	-----END RSA PRIVATE KEY-----
	EOF
		public_key = <<EOF
	-----BEGIN PUBLIC KEY-----
	MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2ymVc24BoaZb0ElXoI3X
	4zUKJGZEetR6F4yT1tJhkPDg7UTmiNoFB5TZJvP6LBrrSwszkpZbxaVOkBrwrGbq
	FUaXPgud8VabfHl0imXvN746zmpjYEMGqJzm+Gh6aBWOlnPdLuHhds/kcanFAEpp
	j5yN0tVWDnqjOJjR7EpxMSdP3TSd6tNAY73LGNLNJQc6tSxh8nMIb4HhSWQSgfof
	+FwcLGvs+mmyBq8Jz5Zy4BSCk1fQFmCnSGyzpyBD0vMd6eLHk2l0tm56DrlonbDM
	X8KGs7e9ZgjANkT5PnipLOaeLJU4H+OWyBZUAT4hl/iRVvLwV81x7g0/O38kmPYJ
	DQIDAQAB
	-----END PUBLIC KEY-----
	EOF
	}`

	keyBasic := util.ExecuteTemplate(
		fqrn,
		template,
		map[string]string{
			"id":         fmt.Sprint(id),
			"name":       name,
			"passphrase": "password",
		},
	)

	keyUpdatedPassphrase := util.ExecuteTemplate(
		fqrn,
		template,
		map[string]string{
			"id":         fmt.Sprint(id),
			"name":       name,
			"passphrase": "new-password",
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, security.VerifyKeyPair),
		Steps: []resource.TestStep{
			{
				Config: keyBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "pair_name", name),
					resource.TestCheckResourceAttr(fqrn, "public_key", "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2ymVc24BoaZb0ElXoI3X\n4zUKJGZEetR6F4yT1tJhkPDg7UTmiNoFB5TZJvP6LBrrSwszkpZbxaVOkBrwrGbq\nFUaXPgud8VabfHl0imXvN746zmpjYEMGqJzm+Gh6aBWOlnPdLuHhds/kcanFAEpp\nj5yN0tVWDnqjOJjR7EpxMSdP3TSd6tNAY73LGNLNJQc6tSxh8nMIb4HhSWQSgfof\n+FwcLGvs+mmyBq8Jz5Zy4BSCk1fQFmCnSGyzpyBD0vMd6eLHk2l0tm56DrlonbDM\nX8KGs7e9ZgjANkT5PnipLOaeLJU4H+OWyBZUAT4hl/iRVvLwV81x7g0/O38kmPYJ\nDQIDAQAB\n-----END PUBLIC KEY-----\n"),
					resource.TestCheckResourceAttr(fqrn, "private_key", "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2ymVc24BoaZb0ElXoI3X4zUKJGZEetR6F4yT1tJhkPDg7UTm\niNoFB5TZJvP6LBrrSwszkpZbxaVOkBrwrGbqFUaXPgud8VabfHl0imXvN746zmpj\nYEMGqJzm+Gh6aBWOlnPdLuHhds/kcanFAEppj5yN0tVWDnqjOJjR7EpxMSdP3TSd\n6tNAY73LGNLNJQc6tSxh8nMIb4HhSWQSgfof+FwcLGvs+mmyBq8Jz5Zy4BSCk1fQ\nFmCnSGyzpyBD0vMd6eLHk2l0tm56DrlonbDMX8KGs7e9ZgjANkT5PnipLOaeLJU4\nH+OWyBZUAT4hl/iRVvLwV81x7g0/O38kmPYJDQIDAQABAoIBAFb7wyhEIfuhhlE9\nuryrb2LrGzJlMIq7qBWOouKhLz4SjIM/VGw+c76VkjZGoSU+LeLj+D0W1ie0u2Cw\ngJM8aW22TbK/c5lksWOO5PVFDdPG+ZoRWY3MLGlhlL5E4UhMPgJyy/eeiRjZ3CZM\npja+UfVAwn1KVNR8UinVZYPt680AvEd1FGxoNLxemIPNug46nNqp6Al86Bn+BnkN\nGXpwyooXVSfo4k+pnFBFIXUdA1dUEXQSVb1AxlTo6g/Ok/+8Gfq8idCdu+5fcZI2\n1eLeC+FAa92rr1SFX/UWeB4cMyuAqwuxbFFIplT22SaUSsNuOUSH4E00nbP/AuCb\n1BqrLmECgYEA7tQKfyF9aiXTsOMdOnSAa5OyEaCfsFtcmLd4ykVrwN8O36NoX005\nVbGuqo87fwIXQIKHU+kOEs/TmaQ8bNcbCD/SfWGTtOnHG4qfIsepJuoMwbQHRhGF\nJeoXh5yEUKg5pcDBY8PENEtEVKmFuL4bPOdn+9FNLGsjftvXpmGWbGUCgYEA6uuQ\n7kzO6WD88OsxdJzlJM11hg2SaSBCh3+5tnOhF1ULOUt4tdYXzh3QI6BPX7tkArYf\nXteVfWoWqn6T7LtCjFm350BqVpPhqfLKnt6fYf1yotsj/cyZXlXquRbxbgakB0n0\n4PrsZaube0TPPVeirzNyOVHyFc+iW+F+IUYD+4kCgYEApDEjBkP/9PoMj4+UiJuP\nrmXcBkJnhtdI0bVRVb5kVjUEBLxTBTISONfvPVM7lBXb5n3Wi9mt00EOOJKw+CLq\ncsFt9MUgxz/xov2qaj7aC+bc3k7msUVaRLardpAkZ09AUrQyQGRWf50/XPUu+dO4\n5iYxVu6OH/uIa664k6qDwAECgYEAslS8oomgEL3VhbWkx1dLA5MMggTPfgpFNsMY\n4Y4JXcLrUEUgjzjEvW0YUdMiLhP8qapDSiXxj1D3f9myxWSp8g0xc9UMZEjCZ9at\nRcjNyP8zBLnCKqokSt6B3puyDsnvvrC/ugIBbnTFBOCJSZG7J7CwJx8z3KbQI1ub\n+fpCj7ECgYAd69soLEybUGMjsdI+OijIGoUTUoZGXJm+0VpBt4QJCe7AMnYPfYzA\nJnEmN4D7HLTKUBklQnb/FhP/RuiT2bSAd1l+PNeuU7gYROCBBonzxXQ1wEbNrSYA\niyoc9g/kvV8HTW8361xEhu7wmuSEEx1gQ/7sdhTkgrTncc8hxVRxuA==\n-----END RSA PRIVATE KEY-----\n"),
					resource.TestCheckResourceAttr(fqrn, "alias", fmt.Sprintf("foo-alias%d", id)),
					resource.TestCheckResourceAttr(fqrn, "pair_type", "RSA"),
					resource.TestCheckResourceAttr(fqrn, "unavailable", "false"),
					resource.TestCheckResourceAttr(fqrn, "passphrase", "password"),
				),
			},
			{
				Config: keyUpdatedPassphrase,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "pair_name", name),
					resource.TestCheckResourceAttr(fqrn, "public_key", "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2ymVc24BoaZb0ElXoI3X\n4zUKJGZEetR6F4yT1tJhkPDg7UTmiNoFB5TZJvP6LBrrSwszkpZbxaVOkBrwrGbq\nFUaXPgud8VabfHl0imXvN746zmpjYEMGqJzm+Gh6aBWOlnPdLuHhds/kcanFAEpp\nj5yN0tVWDnqjOJjR7EpxMSdP3TSd6tNAY73LGNLNJQc6tSxh8nMIb4HhSWQSgfof\n+FwcLGvs+mmyBq8Jz5Zy4BSCk1fQFmCnSGyzpyBD0vMd6eLHk2l0tm56DrlonbDM\nX8KGs7e9ZgjANkT5PnipLOaeLJU4H+OWyBZUAT4hl/iRVvLwV81x7g0/O38kmPYJ\nDQIDAQAB\n-----END PUBLIC KEY-----\n"),
					resource.TestCheckResourceAttr(fqrn, "private_key", "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2ymVc24BoaZb0ElXoI3X4zUKJGZEetR6F4yT1tJhkPDg7UTm\niNoFB5TZJvP6LBrrSwszkpZbxaVOkBrwrGbqFUaXPgud8VabfHl0imXvN746zmpj\nYEMGqJzm+Gh6aBWOlnPdLuHhds/kcanFAEppj5yN0tVWDnqjOJjR7EpxMSdP3TSd\n6tNAY73LGNLNJQc6tSxh8nMIb4HhSWQSgfof+FwcLGvs+mmyBq8Jz5Zy4BSCk1fQ\nFmCnSGyzpyBD0vMd6eLHk2l0tm56DrlonbDMX8KGs7e9ZgjANkT5PnipLOaeLJU4\nH+OWyBZUAT4hl/iRVvLwV81x7g0/O38kmPYJDQIDAQABAoIBAFb7wyhEIfuhhlE9\nuryrb2LrGzJlMIq7qBWOouKhLz4SjIM/VGw+c76VkjZGoSU+LeLj+D0W1ie0u2Cw\ngJM8aW22TbK/c5lksWOO5PVFDdPG+ZoRWY3MLGlhlL5E4UhMPgJyy/eeiRjZ3CZM\npja+UfVAwn1KVNR8UinVZYPt680AvEd1FGxoNLxemIPNug46nNqp6Al86Bn+BnkN\nGXpwyooXVSfo4k+pnFBFIXUdA1dUEXQSVb1AxlTo6g/Ok/+8Gfq8idCdu+5fcZI2\n1eLeC+FAa92rr1SFX/UWeB4cMyuAqwuxbFFIplT22SaUSsNuOUSH4E00nbP/AuCb\n1BqrLmECgYEA7tQKfyF9aiXTsOMdOnSAa5OyEaCfsFtcmLd4ykVrwN8O36NoX005\nVbGuqo87fwIXQIKHU+kOEs/TmaQ8bNcbCD/SfWGTtOnHG4qfIsepJuoMwbQHRhGF\nJeoXh5yEUKg5pcDBY8PENEtEVKmFuL4bPOdn+9FNLGsjftvXpmGWbGUCgYEA6uuQ\n7kzO6WD88OsxdJzlJM11hg2SaSBCh3+5tnOhF1ULOUt4tdYXzh3QI6BPX7tkArYf\nXteVfWoWqn6T7LtCjFm350BqVpPhqfLKnt6fYf1yotsj/cyZXlXquRbxbgakB0n0\n4PrsZaube0TPPVeirzNyOVHyFc+iW+F+IUYD+4kCgYEApDEjBkP/9PoMj4+UiJuP\nrmXcBkJnhtdI0bVRVb5kVjUEBLxTBTISONfvPVM7lBXb5n3Wi9mt00EOOJKw+CLq\ncsFt9MUgxz/xov2qaj7aC+bc3k7msUVaRLardpAkZ09AUrQyQGRWf50/XPUu+dO4\n5iYxVu6OH/uIa664k6qDwAECgYEAslS8oomgEL3VhbWkx1dLA5MMggTPfgpFNsMY\n4Y4JXcLrUEUgjzjEvW0YUdMiLhP8qapDSiXxj1D3f9myxWSp8g0xc9UMZEjCZ9at\nRcjNyP8zBLnCKqokSt6B3puyDsnvvrC/ugIBbnTFBOCJSZG7J7CwJx8z3KbQI1ub\n+fpCj7ECgYAd69soLEybUGMjsdI+OijIGoUTUoZGXJm+0VpBt4QJCe7AMnYPfYzA\nJnEmN4D7HLTKUBklQnb/FhP/RuiT2bSAd1l+PNeuU7gYROCBBonzxXQ1wEbNrSYA\niyoc9g/kvV8HTW8361xEhu7wmuSEEx1gQ/7sdhTkgrTncc8hxVRxuA==\n-----END RSA PRIVATE KEY-----\n"),
					resource.TestCheckResourceAttr(fqrn, "alias", fmt.Sprintf("foo-alias%d", id)),
					resource.TestCheckResourceAttr(fqrn, "pair_type", "RSA"),
					resource.TestCheckResourceAttr(fqrn, "unavailable", "false"),
					resource.TestCheckResourceAttr(fqrn, "passphrase", "new-password"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "pair_name"),
				ImportStateVerifyIgnore: []string{"passphrase", "private_key"},
			},
		},
	})
}

func TestAccKeyPairGPG(t *testing.T) {
	id, fqrn, name := test.MkNames("mykp", "artifactory_keypair")
	keyBasic := fmt.Sprintf(`
		resource "artifactory_keypair" "%s" {
			pair_name  = "%s"
			pair_type = "GPG"
			alias = "foo-alias%d"
			passphrase = "password"
			private_key = <<EOF
		-----BEGIN PGP PRIVATE KEY BLOCK-----
		Version: Keybase OpenPGP v1.0.0
		Comment: https://keybase.io/crypto

		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
		zoQkekoL4H4GU8QB6GY=
		=JQxa
		-----END PGP PRIVATE KEY BLOCK-----
		EOF
			public_key = <<EOF
		-----BEGIN PGP PUBLIC KEY BLOCK-----
		Version: Keybase OpenPGP v1.0.0
		Comment: https://keybase.io/crypto

		xo0EYGrVNAEEAPD3YDt0qP8kSV8bnmqVP5XDPoN40gEpUGtDLjAn6d+cRMeNGaru
		6H0bdgwQpND8Gz9Qx2pCNSxlWDZpY1fCvRQ174iGjvO/3527f148cgKNZtwLsKrZ
		laW8z3tB2LuCM2e97ijX+lzRf7YJUXU3pOfoCFWpOPoRg1CHV0NyHl0VABEBAAHN
		FmFsYW4gPGFsYW5uQGpmcm9nLmNvbT7CrQQTAQoAFwUCYGrVNAIbLwMLCQcDFQoI
		Ah4BAheAAAoJENzR2QJlA6glZmsD/iqhnNFy1Elj3hGL0HaEzeb+KDpcSL/L5a/8
		WIGCQFeLcEn9lC+68b/eERKGIoXJ7z8HfPDFNRTKvomKIdAqFiAeDAUUD0B82rsx
		xDf8USnTwJlnd0bPe9nxgXYcrwioEYbPVYGl3jima/KQrbW8XlKyiypy4Nd66Wcn
		TuM6PwRFzo0EYGrVNAEEANVNINyfCQ+y1haaaAJ0uCgx3dW52LwcZfvOP6i798WZ
		dyGA+WSUCEcrklUwZ595E2dNkNKptksftwSeQ0+EH5S1ZlEaq2YUv8fCx32F1ckh
		D3eHaCKRxTPx/zbb96q4ruEGKhOBXceid3o341HbtGVKi8VjBx3XNukskQ+EOvgt
		ABEBAAHCwIMEGAEKAA8FAmBq1TQFCQ8JnAACGy4AqAkQ3NHZAmUDqCWdIAQZAQoA
		BgUCYGrVNAAKCRBddQ63FhKl6NmDBACqxC4lAnsCQERjs02LYAEAwVDhDf0rXxD0
		H+hKDyxQZc80M7WIpXaBHmbs8ekJRnY7JHcer7sizDMdfkR3xB62jNGhc0XiW6nc
		mlwvtWt3+E6AkObmWnocRy5ztTQI0gye0B3cPs2txE2fCs+WD7yLRnM3HqIAh83W
		Cccvh0+uG96dBADlPbZ8g8q6bkeeT72gOi3OCN0A+Y8lUPifhrpSiI9xMpP3aomM
		beZJB6fWjEzNoblQ9jUr/E54bF9jMr6L3uE4OJH9SYJ/HvqcKJC+1TFeQ9lXR7g7
		MTdfxEvhMDhcsd/pYIgrzvDry+B2+jANW10R1yejT/C8QdlWIndDsEsaKs6NBGBq
		1TQBBADJD6gVGTMGb/WsfnSaL5yA3AEMczhPqxD4FDC0vzqGG3XzKgxtmW8cJXls
		NCK80F61daxJ72/UAmfxHbNP1qmHSiosEe+h1YZ/Zo3pN/LODzg9JrOs9A2xwjqE
		nU9mS0jDz5oEQtr9K4+YKOdJvmuaN85ueBizfQ1TYRNuDtmbnQARAQABwsCDBBgB
		CgAPBQJgatU0BQkPCZwAAhsuAKgJENzR2QJlA6glnSAEGQEKAAYFAmBq1TQACgkQ
		ehctgYvYtmh38gP6A9lnQaLuVnTElJLy2XSDTqwWOcy/5J842S/xdQEsWUMXh4I5
		mlotkZwkrdvXp8E/F3P8X7GbxhNAVZX+Xcm95V3g/kmP+Pq7PeUmoZR5LD8ppBfO
		7v6XgaUhraUPAZl6lx4L5pYNCX9JBNUtQAG9xIoap4slvksdz5SN/BwSgV6qqwQA
		tr4YTDXvLyoWwMFB2FjWcw4zwV+7yHwGzogKfGCQy5qVlDoQyWdkwwF1awyk5RIe
		ZxwPZ2SDaiznOmZ+4LjR2NPmjnT96d9RKRtgEjkfW+a19BofrvEalS9wh/jkboea
		d8rDu8wMbLAl77dq1c6dpJDgzoQkekoL4H4GU8QB6GY=
		=fot9
		-----END PGP PUBLIC KEY BLOCK-----
		EOF
		}
	`, name, name, id)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, security.VerifyKeyPair),
		Steps: []resource.TestStep{
			{
				Config: keyBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "pair_name", name),
					resource.TestCheckResourceAttr(fqrn, "public_key", "-----BEGIN PGP PUBLIC KEY BLOCK-----\nVersion: Keybase OpenPGP v1.0.0\nComment: https://keybase.io/crypto\n\nxo0EYGrVNAEEAPD3YDt0qP8kSV8bnmqVP5XDPoN40gEpUGtDLjAn6d+cRMeNGaru\n6H0bdgwQpND8Gz9Qx2pCNSxlWDZpY1fCvRQ174iGjvO/3527f148cgKNZtwLsKrZ\nlaW8z3tB2LuCM2e97ijX+lzRf7YJUXU3pOfoCFWpOPoRg1CHV0NyHl0VABEBAAHN\nFmFsYW4gPGFsYW5uQGpmcm9nLmNvbT7CrQQTAQoAFwUCYGrVNAIbLwMLCQcDFQoI\nAh4BAheAAAoJENzR2QJlA6glZmsD/iqhnNFy1Elj3hGL0HaEzeb+KDpcSL/L5a/8\nWIGCQFeLcEn9lC+68b/eERKGIoXJ7z8HfPDFNRTKvomKIdAqFiAeDAUUD0B82rsx\nxDf8USnTwJlnd0bPe9nxgXYcrwioEYbPVYGl3jima/KQrbW8XlKyiypy4Nd66Wcn\nTuM6PwRFzo0EYGrVNAEEANVNINyfCQ+y1haaaAJ0uCgx3dW52LwcZfvOP6i798WZ\ndyGA+WSUCEcrklUwZ595E2dNkNKptksftwSeQ0+EH5S1ZlEaq2YUv8fCx32F1ckh\nD3eHaCKRxTPx/zbb96q4ruEGKhOBXceid3o341HbtGVKi8VjBx3XNukskQ+EOvgt\nABEBAAHCwIMEGAEKAA8FAmBq1TQFCQ8JnAACGy4AqAkQ3NHZAmUDqCWdIAQZAQoA\nBgUCYGrVNAAKCRBddQ63FhKl6NmDBACqxC4lAnsCQERjs02LYAEAwVDhDf0rXxD0\nH+hKDyxQZc80M7WIpXaBHmbs8ekJRnY7JHcer7sizDMdfkR3xB62jNGhc0XiW6nc\nmlwvtWt3+E6AkObmWnocRy5ztTQI0gye0B3cPs2txE2fCs+WD7yLRnM3HqIAh83W\nCccvh0+uG96dBADlPbZ8g8q6bkeeT72gOi3OCN0A+Y8lUPifhrpSiI9xMpP3aomM\nbeZJB6fWjEzNoblQ9jUr/E54bF9jMr6L3uE4OJH9SYJ/HvqcKJC+1TFeQ9lXR7g7\nMTdfxEvhMDhcsd/pYIgrzvDry+B2+jANW10R1yejT/C8QdlWIndDsEsaKs6NBGBq\n1TQBBADJD6gVGTMGb/WsfnSaL5yA3AEMczhPqxD4FDC0vzqGG3XzKgxtmW8cJXls\nNCK80F61daxJ72/UAmfxHbNP1qmHSiosEe+h1YZ/Zo3pN/LODzg9JrOs9A2xwjqE\nnU9mS0jDz5oEQtr9K4+YKOdJvmuaN85ueBizfQ1TYRNuDtmbnQARAQABwsCDBBgB\nCgAPBQJgatU0BQkPCZwAAhsuAKgJENzR2QJlA6glnSAEGQEKAAYFAmBq1TQACgkQ\nehctgYvYtmh38gP6A9lnQaLuVnTElJLy2XSDTqwWOcy/5J842S/xdQEsWUMXh4I5\nmlotkZwkrdvXp8E/F3P8X7GbxhNAVZX+Xcm95V3g/kmP+Pq7PeUmoZR5LD8ppBfO\n7v6XgaUhraUPAZl6lx4L5pYNCX9JBNUtQAG9xIoap4slvksdz5SN/BwSgV6qqwQA\ntr4YTDXvLyoWwMFB2FjWcw4zwV+7yHwGzogKfGCQy5qVlDoQyWdkwwF1awyk5RIe\nZxwPZ2SDaiznOmZ+4LjR2NPmjnT96d9RKRtgEjkfW+a19BofrvEalS9wh/jkboea\nd8rDu8wMbLAl77dq1c6dpJDgzoQkekoL4H4GU8QB6GY=\n=fot9\n-----END PGP PUBLIC KEY BLOCK-----\n"),
					resource.TestCheckResourceAttr(fqrn, "private_key", "-----BEGIN PGP PRIVATE KEY BLOCK-----\nVersion: Keybase OpenPGP v1.0.0\nComment: https://keybase.io/crypto\n\nxcFGBGBq1TQBBADw92A7dKj/JElfG55qlT+Vwz6DeNIBKVBrQy4wJ+nfnETHjRmq\n7uh9G3YMEKTQ/Bs/UMdqQjUsZVg2aWNXwr0UNe+Iho7zv9+du39ePHICjWbcC7Cq\n2ZWlvM97Qdi7gjNnve4o1/pc0X+2CVF1N6Tn6AhVqTj6EYNQh1dDch5dFQARAQAB\n/gkDCD1IN++hrp7WYJm/QRPGUF3WAddHNpoHWK5bRaW1Zcf2EOp+76SacCOEiOHW\n7VzzVEr/OWym3JZvdqg8K93kHNrwQ1vqCalscti3Cc4MIT3jBUvgzG1HxET3pmVM\nJMkDj15oaEf6bEMuVC61mPa7kmfxdjJeaYjNFdnHSHTqi0gPTqA15vQGCO58AEmX\n5a0hY8jS0pf8CNAWURnYemkrNzy2vwG3x3x7d/M1X3XkpzJVlPR1HaY2V9KJsUBg\naUfv6ydG87T4PYwbOYQJ+wC8KFuylajpdHpUB+5WL5qbMB5nt3TJXcILEb8ALTLi\nQTldl2HZc+GqLG+JnoQRUSXy0ZeRC+qEhjTVnpK2uoJtOtMXCuD0QrlcLwk4mtzn\nzCvEM4uyb8MB/4oEQmPx8iLZ3u4MQEpfUMz5j2nB2XvY1fqrrvdn8Alh8EMsVvK0\nie29qfazy7+fTuJ8p6o3VpJVP10pVZZ/oGIDmn41RsLVULTtZbkF0NzNFmFsYW4g\nPGFsYW5uQGpmcm9nLmNvbT7CrQQTAQoAFwUCYGrVNAIbLwMLCQcDFQoIAh4BAheA\nAAoJENzR2QJlA6glZmsD/iqhnNFy1Elj3hGL0HaEzeb+KDpcSL/L5a/8WIGCQFeL\ncEn9lC+68b/eERKGIoXJ7z8HfPDFNRTKvomKIdAqFiAeDAUUD0B82rsxxDf8USnT\nwJlnd0bPe9nxgXYcrwioEYbPVYGl3jima/KQrbW8XlKyiypy4Nd66WcnTuM6PwRF\nx8FGBGBq1TQBBADVTSDcnwkPstYWmmgCdLgoMd3Vudi8HGX7zj+ou/fFmXchgPlk\nlAhHK5JVMGefeRNnTZDSqbZLH7cEnkNPhB+UtWZRGqtmFL/Hwsd9hdXJIQ93h2gi\nkcUz8f822/equK7hBioTgV3Hond6N+NR27RlSovFYwcd1zbpLJEPhDr4LQARAQAB\n/gkDCOjV8ORMDf1sYMHoCaYCl8atFXxI3WyvMwaFPJVjbEiEWHK1ljCTOSkeXufI\nWBTwdJ11AiEGMdU3pxxueThr5FtcVvfitlmGEYwGbFFwo2iQPOWk3MhfRStrSXmP\n3yaFwRN4brJGdcNUo6HDT+8xpJeneZtuobKDmUE320L8lHEcA1Saj0jDCnbeaU7M\nX22nLj98Tr7cFT1pwTdimgIVW8iHl3Iv4Ytjd0hO6RDSZvS5a/A7v4bg2VndLhH/\n86HAHV2VtLryUTJRH1tDLy6vOaeJ2Fh5xniPIMTXNK09v6lwONrHMC3kHeaOOrEp\nMYVXx7lNaKNLsyMSuQHZvbshiVcrQZjh+GXtJDdJ7G1J3ENFLo2B/OWeGydFj+RX\npfwae6rmYPKQaxe1aK1iSxtDSv/ANJQHfGm2l39NUeEFf1H3rLJj48n3cfcQTW7O\n/ya9Vx8o/EtdvJBPW1Mdh9b0TikHcuPgrS7pxQJ6EhGHRxao0fajPKvCwIMEGAEK\nAA8FAmBq1TQFCQ8JnAACGy4AqAkQ3NHZAmUDqCWdIAQZAQoABgUCYGrVNAAKCRBd\ndQ63FhKl6NmDBACqxC4lAnsCQERjs02LYAEAwVDhDf0rXxD0H+hKDyxQZc80M7WI\npXaBHmbs8ekJRnY7JHcer7sizDMdfkR3xB62jNGhc0XiW6ncmlwvtWt3+E6AkObm\nWnocRy5ztTQI0gye0B3cPs2txE2fCs+WD7yLRnM3HqIAh83WCccvh0+uG96dBADl\nPbZ8g8q6bkeeT72gOi3OCN0A+Y8lUPifhrpSiI9xMpP3aomMbeZJB6fWjEzNoblQ\n9jUr/E54bF9jMr6L3uE4OJH9SYJ/HvqcKJC+1TFeQ9lXR7g7MTdfxEvhMDhcsd/p\nYIgrzvDry+B2+jANW10R1yejT/C8QdlWIndDsEsaKsfBRgRgatU0AQQAyQ+oFRkz\nBm/1rH50mi+cgNwBDHM4T6sQ+BQwtL86hht18yoMbZlvHCV5bDQivNBetXWsSe9v\n1AJn8R2zT9aph0oqLBHvodWGf2aN6Tfyzg84PSazrPQNscI6hJ1PZktIw8+aBELa\n/SuPmCjnSb5rmjfObngYs30NU2ETbg7Zm50AEQEAAf4JAwhQ3Ax+n8w4YWBI6mOO\nZHng6UUrbVOi7EqO8hgifTkOheVRU4QTwKEkuwvHcEQ4g0ZGHxMN6vDkzdZ/QLrQ\nbHP3YWpRgi9alUFt6Q4FNR10vWZXPMMTbxf7KJ9J2Te3/pSAJX5set39k6rYgfNc\n3VMDvsvf17c/yW4TmbP20VlyiTd4cy7jQ+UeLrZCp3SohnhjJwf5ogpiMi09zB/6\nR0koljwtUlGk5Sjo2Q7zJ2hxx6i45OYzhP7cGW8t8voInTZbA5lKPXFYiWVQx5D0\n2UjfNSO1hNKrohacWQcoVjiU95N2QrP2RTQR1XuVjgqb5c1LW0GzzYx31HUxHW0x\n0OzwM6yPdt238SVZ/0WEby6D4YqJIUT6rUbF7oq8CGV+HiuhBx2Ppxky72TrpI0u\nB0ocNhYPvbY0zNJ46d91uYpWlWj7vynUS6jDDRHvoZZWGKO6iAYLYAU8oXsYY6U8\ngEcYzJGvPw1sQuSA9ag+WIuTzo5GO6Y+wsCDBBgBCgAPBQJgatU0BQkPCZwAAhsu\nAKgJENzR2QJlA6glnSAEGQEKAAYFAmBq1TQACgkQehctgYvYtmh38gP6A9lnQaLu\nVnTElJLy2XSDTqwWOcy/5J842S/xdQEsWUMXh4I5mlotkZwkrdvXp8E/F3P8X7Gb\nxhNAVZX+Xcm95V3g/kmP+Pq7PeUmoZR5LD8ppBfO7v6XgaUhraUPAZl6lx4L5pYN\nCX9JBNUtQAG9xIoap4slvksdz5SN/BwSgV6qqwQAtr4YTDXvLyoWwMFB2FjWcw4z\nwV+7yHwGzogKfGCQy5qVlDoQyWdkwwF1awyk5RIeZxwPZ2SDaiznOmZ+4LjR2NPm\njnT96d9RKRtgEjkfW+a19BofrvEalS9wh/jkboead8rDu8wMbLAl77dq1c6dpJDg\nzoQkekoL4H4GU8QB6GY=\n=JQxa\n-----END PGP PRIVATE KEY BLOCK-----\n"),
					resource.TestCheckResourceAttr(fqrn, "alias", fmt.Sprintf("foo-alias%d", id)),
					resource.TestCheckResourceAttr(fqrn, "pair_type", "GPG"),
					resource.TestCheckResourceAttr(fqrn, "unavailable", "false"),
					resource.TestCheckResourceAttr(fqrn, "passphrase", "password"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "pair_name"),
				ImportStateVerifyIgnore: []string{"passphrase", "private_key"},
			},
		},
	})
}
