package provider

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testClientCertificatePEM = `-----BEGIN CERTIFICATE-----
MIICUjCCAbugAwIBAgIJALRDng3rGeQvMA0GCSqGSIb3DQEBCwUAMEIxCzAJBgNV
BAYTAlhYMRUwEwYDVQQHDAxEZWZhdWx0IENpdHkxHDAaBgNVBAoME0RlZmF1bHQg
Q29tcGFueSBMdGQwHhcNMTkwNTE3MTAwMzI2WhcNMjkwNTE0MTAwMzI2WjBCMQsw
CQYDVQQGEwJYWDEVMBMGA1UEBwwMRGVmYXVsdCBDaXR5MRwwGgYDVQQKDBNEZWZh
dWx0IENvbXBhbnkgTHRkMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDVBRt7
Ua3j7K2htVRu1tw629ZZZQI35RGm/53ffF/QUUFXk35at+IiwYZGGQbOGuN1pdji
gki9/Qit/WO/3uadSkGelKOUYD0DIemlhcZt6iPMQq8mYlUkMPZz5Qlj0ldKI3g+
Q8Tc/6vEeBv/9jrm9Efg/uwc0DjD8B4Ny6xMHQIDAQABo1AwTjAdBgNVHQ4EFgQU
VrBaHnYLayO2lKIUde8etG0H6owwHwYDVR0jBBgwFoAUVrBaHnYLayO2lKIUde8e
tG0H6owwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOBgQA4VBFCrbuOsKtY
uNlSQCBkTXg907iXihZ+Of/2rerS2gfDCUHdz0xbYdlttNjoGVCA+0alt7ugfYpl
fy5aAfCHLXEgYrlhe6oDtCMSskbkKFTEI/bRqwGMDb+9NO/yh2KLbNueKJz9Vs5V
GV9pUrgW6c7kLrC9vpHP+47iyQEbnw==
-----END CERTIFICATE-----`

	testClientPrivateKeyPEM = `-----BEGIN PRIVATE KEY-----
MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBANUFG3tRrePsraG1
VG7W3Drb1lllAjflEab/nd98X9BRQVeTflq34iLBhkYZBs4a43Wl2OKCSL39CK39
Y7/e5p1KQZ6Uo5RgPQMh6aWFxm3qI8xCryZiVSQw9nPlCWPSV0ojeD5DxNz/q8R4
G//2Oub0R+D+7BzQOMPwHg3LrEwdAgMBAAECgYAxWA6GoWQDcRbDZ6qYRkMbi0L6
0DAUXIabRYj/dOMI8VmOfMb/IqtKW8PLxw5Rfd8EqJc12PIauFtjWlfZ4TtP9erQ
1imw2SpVMAWt4HLUw7oONKgNMnBtVQBCoXLuXcnJbCxeRiV1oJtvrddUJPOtUc+y
t5gGTyx/zUAXzPzT7QJBAOvu4CH0Xc+1GdXFUFLzF8B3SFwnOFRERJxFq43dw4t3
tXcON/UyegYcQz2JqKcofwRhM4+uXGnWE+9oOOnxL8sCQQDnI1QtMv+tZcqIcmk6
1ykyNa530eCfoqAvVTRwPIsAD/DZLC4HJNSQauPXC4Unt1tqmOmUoZmgzYQlVsGO
ISa3AkB2xWpPrZUMWz8GPq6RE4+BdIsY2SWiRjvD787NPDaUn07bAG1rIl4LdW7k
K8ibXeeTbNtoGX6sSPkALJd6LdDBAkEA5FAhdgRKSh2iUeWxzE18g/xCuli2aPlb
AWZIxhUHuKgGYH8jeCsJTR5IsMLQZMrZohIpqId4GT7oqXlo99wHQQJBAOvX+5z6
iCooatRyMnwUV6sJ225ZawuJ4sXFt6CA7aOZQ+G5zvG694ONxG9qeF2YnySQp1HH
V87CqqFaUigTzmI=
-----END PRIVATE KEY-----`
)

func TestBuildTLSConfigWithPEM(t *testing.T) {
	cfg, err := buildTLSConfig(tlsConfigOptions{
		ClientCertificatePEM: testClientCertificatePEM,
		ClientPrivateKeyPEM:  testClientPrivateKeyPEM,
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if cfg == nil {
		t.Fatal("expected TLS config when PEM data is provided")
	}
	if len(cfg.Certificates) != 1 {
		t.Fatalf("expected exactly one certificate, got %d", len(cfg.Certificates))
	}
}

func TestBuildTLSConfigWithPaths(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "client-cert.pem")
	keyPath := filepath.Join(dir, "client-key.pem")

	if err := os.WriteFile(certPath, []byte(testClientCertificatePEM), 0o600); err != nil {
		t.Fatalf("failed writing cert file: %s", err)
	}
	if err := os.WriteFile(keyPath, []byte(testClientPrivateKeyPEM), 0o600); err != nil {
		t.Fatalf("failed writing key file: %s", err)
	}

	cfg, err := buildTLSConfig(tlsConfigOptions{
		ClientCertificatePath:    certPath,
		ClientCertificateKeyPath: keyPath,
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if cfg == nil {
		t.Fatal("expected TLS config when certificate paths are provided")
	}
	if len(cfg.Certificates) != 1 {
		t.Fatalf("expected exactly one certificate, got %d", len(cfg.Certificates))
	}
}

func TestBuildTLSConfigConflictingOptions(t *testing.T) {
	_, err := buildTLSConfig(tlsConfigOptions{
		ClientCertificatePath:    "cert.pem",
		ClientCertificateKeyPath: "key.pem",
		ClientCertificatePEM:     testClientCertificatePEM,
		ClientPrivateKeyPEM:      testClientPrivateKeyPEM,
	})
	if err == nil || !strings.Contains(err.Error(), "cannot configure both path-based and inline") {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func TestBuildTLSConfigMissingPathPair(t *testing.T) {
	_, err := buildTLSConfig(tlsConfigOptions{
		ClientCertificatePath: "cert.pem",
	})
	if err == nil || !strings.Contains(err.Error(), "client_certificate_path and client_certificate_key_path") {
		t.Fatalf("expected missing key path error, got %v", err)
	}
}

func TestBuildTLSConfigMissingPEMPair(t *testing.T) {
	_, err := buildTLSConfig(tlsConfigOptions{
		ClientCertificatePEM: testClientCertificatePEM,
	})
	if err == nil || !strings.Contains(err.Error(), "client_certificate_pem and client_private_key_pem") {
		t.Fatalf("expected missing key PEM error, got %v", err)
	}
}

func TestBuildTLSConfigBypassOnly(t *testing.T) {
	cfg, err := buildTLSConfig(tlsConfigOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if cfg == nil {
		t.Fatal("expected TLS config when bypassing verification")
	}
	if !cfg.InsecureSkipVerify {
		t.Fatal("expected InsecureSkipVerify to be true")
	}
	if len(cfg.Certificates) != 0 {
		t.Fatalf("expected zero certificates, got %d", len(cfg.Certificates))
	}
}

func TestBuildTLSConfigReturnsNilWhenNoOptions(t *testing.T) {
	cfg, err := buildTLSConfig(tlsConfigOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if cfg != nil {
		t.Fatalf("expected nil config, got %#v", cfg)
	}
}
