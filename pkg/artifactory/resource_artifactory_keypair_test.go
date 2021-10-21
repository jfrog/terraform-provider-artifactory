package artifactory

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccKeyPair(t *testing.T) {
	id, fqrn, name := mkNames("mykp", "artifactory_keypair")
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
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDestroy(fqrn, verifyKeyPair),
		ProviderFactories: testAccProviders,
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
				),
			},
		},
	})
}
