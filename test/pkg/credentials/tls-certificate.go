package credentials

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"time"
)

type TLSCertificate struct {
	certTemplate *x509.Certificate

	pathToCertPEM string
	pathToKeyPEM  string
}

func NewTLSCertificate(options ...tlsCertificateOption) (
	c *TLSCertificate, e error,
) {
	const (
		tempFileDirectory = ""
		tempFilePattern   = "*"

		certValidDuration = time.Hour
		serialNumber      = 1

		certType = "CERTIFICATE"
		keyType  = "PRIVATE KEY"

		filePermission = 0400
	)

	var (
		certPEM *os.File
		keyPEM  *os.File

		option tlsCertificateOption

		key           ed25519.PrivateKey
		keyPEMBlock   *pem.Block
		keyPKCS8Bytes []byte

		publicKey ed25519.PublicKey

		certBytes    []byte
		certPEMBlock *pem.Block
	)

	keyPEM, e = ioutil.TempFile(tempFileDirectory, tempFilePattern)
	if e != nil {
		return
	}

	certPEM, e = ioutil.TempFile(tempFileDirectory, tempFilePattern)
	if e != nil {
		return
	}

	c = &TLSCertificate{
		certTemplate: &x509.Certificate{
			SerialNumber: big.NewInt(serialNumber),
			NotBefore:    time.Now(),
			NotAfter:     time.Now().Add(certValidDuration),
			KeyUsage:     x509.KeyUsageDigitalSignature,
			ExtKeyUsage:  make([]x509.ExtKeyUsage, 0),
			IPAddresses:  make([]net.IP, 0),
		},

		pathToCertPEM: certPEM.Name(),
		pathToKeyPEM:  keyPEM.Name(),
	}

	for _, option = range options {
		e = option(c)
		if e != nil {
			return
		}
	}

	publicKey, key, e = ed25519.GenerateKey(rand.Reader)
	if e != nil {
		return
	}

	keyPKCS8Bytes, e = x509.MarshalPKCS8PrivateKey(key)
	if e != nil {
		return
	}

	keyPEMBlock = &pem.Block{
		Type:  keyType,
		Bytes: keyPKCS8Bytes,
	}

	e = pem.Encode(keyPEM, keyPEMBlock)
	if e != nil {
		return
	}

	e = keyPEM.Chmod(filePermission)
	if e != nil {
		return
	}

	e = keyPEM.Close()
	if e != nil {
		return
	}

	certBytes, e = x509.CreateCertificate(
		rand.Reader,
		c.certTemplate,
		c.certTemplate,
		publicKey,
		key,
	)
	if e != nil {
		return
	}

	certPEMBlock = &pem.Block{
		Type:  certType,
		Bytes: certBytes,
	}

	e = pem.Encode(certPEM, certPEMBlock)
	if e != nil {
		return
	}

	e = certPEM.Close()
	if e != nil {
		return
	}

	return
}

func (c *TLSCertificate) PathToCertPEM() string {
	return c.pathToCertPEM
}

func (c *TLSCertificate) PathToKeyPEM() string {
	return c.pathToKeyPEM
}

func (c *TLSCertificate) Destroy() (e error) {
	e = os.Remove(c.pathToCertPEM)
	if e != nil {
		return
	}

	e = os.Remove(c.pathToKeyPEM)
	if e != nil {
		return
	}

	return
}

type tlsCertificateOption func(*TLSCertificate) error

func WithExtendedKeyUsageForServerAuth() (option tlsCertificateOption) {
	option = func(c *TLSCertificate) (e error) {
		c.certTemplate.ExtKeyUsage = append(c.certTemplate.ExtKeyUsage,
			x509.ExtKeyUsageServerAuth,
		)

		return
	}

	return
}

func WithIPAddress(address string) (option tlsCertificateOption) {
	option = func(c *TLSCertificate) (e error) {
		c.certTemplate.IPAddresses = append(c.certTemplate.IPAddresses,
			net.ParseIP(address),
		)

		return
	}

	return
}
