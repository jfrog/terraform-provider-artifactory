package cert

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

func loadCertificates(caCertPool *x509.CertPool, certificatesDirPath string) error {
	if !fileutils.IsPathExists(certificatesDirPath, false) {
		return nil
	}
	files, err := ioutil.ReadDir(certificatesDirPath)
	err = errorutils.CheckError(err)
	if err != nil {
		return err
	}
	for _, file := range files {
		caCert, err := ioutil.ReadFile(filepath.Join(certificatesDirPath, file.Name()))
		err = errorutils.CheckError(err)
		if err != nil {
			return err
		}
		caCertPool.AppendCertsFromPEM(caCert)
	}
	return nil
}

func GetTransportWithLoadedCert(certificatesDirPath string, insecureTls bool, transport *http.Transport) (*http.Transport, error) {
	// Remove once SystemCertPool supports windows
	caCertPool, err := loadSystemRoots()
	err = errorutils.CheckError(err)
	if err != nil {
		return nil, err
	}
	err = loadCertificates(caCertPool, certificatesDirPath)
	if err != nil {
		return nil, err
	}
	// Setup HTTPS client
	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		ClientSessionCache: tls.NewLRUClientSessionCache(1),
		InsecureSkipVerify: insecureTls,
	}
	tlsConfig.BuildNameToCertificate()
	transport.TLSClientConfig = tlsConfig

	return transport, nil
}
