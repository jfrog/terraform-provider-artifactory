package auth

import (
	"bytes"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/xanzy/ssh-agent"
	"golang.org/x/crypto/ssh"
)

func SshAuthentication(url, sshKeyPath, sshPassphrase string) (sshAuthHeaders map[string]string, newUrl string, err error) {
	_, host, port, err := parseUrl(url)
	if err != nil {
		return nil, "", err
	}

	var sshAuth ssh.AuthMethod
	log.Info("Performing SSH authentication...")
	log.Info("Trying to authenticate via SSH-Agent...")

	// Try authenticating via agent. If failed, try authenticating via key.
	sshAuth, err = sshAuthAgent()
	if err == nil {
		sshAuthHeaders, newUrl, err = getSshHeaders(sshAuth, host, port)
	}
	if err != nil {
		log.Info("Authentication via SSH-Agent failed. Error:\n", err)
		log.Info("Trying to authenticate via SSH Key...")

		// Check if key specified
		if len(sshKeyPath) <= 0 {
			log.Info("Authentication via SSH key failed.")
			return nil, "", errorutils.CheckError(fmt.Errorf("SSH key not specified."))
		}

		// Read key and passphrase
		var sshKey, sshPassphraseBytes []byte
		sshKey, sshPassphraseBytes, err = readSshKeyAndPassphrase(sshKeyPath, sshPassphrase)
		if err != nil {
			log.Info("Authentication via SSH key failed.")
			return nil, "", err
		}

		// Verify key and get ssh headers
		sshAuth, err = sshAuthPublicKey(sshKey, sshPassphraseBytes)
		if err == nil {
			sshAuthHeaders, newUrl, err = getSshHeaders(sshAuth, host, port)
		}
		if err != nil {
			log.Info("Authentication via SSH Key failed.")
			return nil, "", err
		}
	}

	// If successful, return headers
	log.Info("SSH authentication successful.")
	return sshAuthHeaders, newUrl, nil
}

func getSshHeaders(sshAuth ssh.AuthMethod, host string, port int) (map[string]string, string, error) {
	sshConfig := &ssh.ClientConfig{
		User: "admin",
		Auth: []ssh.AuthMethod{
			sshAuth,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	hostAndPort := host + ":" + strconv.Itoa(port)
	connection, err := ssh.Dial("tcp", hostAndPort, sshConfig)
	if errorutils.CheckError(err) != nil {
		return nil, "", err
	}
	defer connection.Close()
	session, err := connection.NewSession()
	if errorutils.CheckError(err) != nil {
		return nil, "", err
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if errorutils.CheckError(err) != nil {
		return nil, "", err
	}

	if err = session.Run("jfrog-authenticate"); err != nil && err != io.EOF {
		return nil, "", errorutils.CheckError(err)
	}
	var buf bytes.Buffer
	io.Copy(&buf, stdout)

	var result SshAuthResult
	if err = json.Unmarshal(buf.Bytes(), &result); errorutils.CheckError(err) != nil {
		return nil, "", err
	}
	url := utils.AddTrailingSlashIfNeeded(result.Href)
	sshAuthHeaders := result.Headers
	return sshAuthHeaders, url, nil
}

func readSshKeyAndPassphrase(sshKeyPath, sshPassphrase string) ([]byte, []byte, error) {
	sshKey, err := ioutil.ReadFile(utils.ReplaceTildeWithUserHome(sshKeyPath))
	if err != nil {
		return nil, nil, errorutils.CheckError(err)
	}
	if len(sshPassphrase) == 0 {
		encryptedKey, err := IsEncrypted(sshKey)
		if err != nil {
			return nil, nil, errorutils.CheckError(err)
		}
		// If key is encrypted but no passphrase specified
		if encryptedKey {
			return nil, nil, errorutils.CheckError(errors.New("SSH Key is encrypted but no passphrase was specified."))
		}
	}

	return sshKey, []byte(sshPassphrase), err
}

func IsEncrypted(buffer []byte) (bool, error) {
	block, _ := pem.Decode(buffer)
	if block == nil {
		return false, errors.New("SSH: no key found")
	}
	return strings.Contains(block.Headers["Proc-Type"], "ENCRYPTED"), nil
}

func parseUrl(url string) (protocol, host string, port int, err error) {
	pattern1 := "^(.+)://(.+):([0-9].+)/$"
	pattern2 := "^(.+)://(.+)$"

	var r *regexp.Regexp
	r, err = regexp.Compile(pattern1)
	if errorutils.CheckError(err) != nil {
		return
	}
	groups := r.FindStringSubmatch(url)
	if len(groups) == 4 {
		protocol = groups[1]
		host = groups[2]
		port, err = strconv.Atoi(groups[3])
		if err != nil {
			err = errorutils.CheckError(errors.New("URL: " + url + " is invalid. Expecting ssh://<host>:<port> or http(s)://..."))
		}
		return
	}

	r, err = regexp.Compile(pattern2)
	err = errorutils.CheckError(err)
	if err != nil {
		return
	}
	groups = r.FindStringSubmatch(url)
	if len(groups) == 3 {
		protocol = groups[1]
		host = groups[2]
		port = 80
	}
	return
}

func sshAuthPublicKey(sshKey, sshPassphrase []byte) (ssh.AuthMethod, error) {
	key, err := ssh.ParsePrivateKeyWithPassphrase(sshKey, sshPassphrase)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	return ssh.PublicKeys(key), nil
}

func sshAuthAgent() (ssh.AuthMethod, error) {
	sshAgent, _, err := sshagent.New()
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	cbk := sshAgent.Signers
	authMethod := ssh.PublicKeysCallback(cbk)
	return authMethod, nil
}

type SshAuthResult struct {
	Href    string
	Headers map[string]string
}
