package credentials

import (
	"crypto/ecdsa"
    "math/big"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"net"
	"os"
	"time"
)

type TLSCertificate struct {
	privateKeyPEM  *os.File
	certificatePEM *os.File
}

func NewTLSCertificateForIPAddress(ip net.IP) (c *TLSCertificate, e error) {
	const (
		directory = ""
		pattern   = "*"

		certificateType = "CERTIFICATE"
		privateKeyType  = "PRIVATE KEY"

		filePermission = 0400

        serialNumber = 1

		certValidDuration = time.Hour
	)

	var (
		privateKey           *ecdsa.PrivateKey
		privateKeyPEMBlock   *pem.Block
		privateKeyPKCS8Bytes []byte

		certificateBytes    []byte
		certificatePEMBlock *pem.Block
		certificateTemplate *x509.Certificate
	)

	privateKey, e = ecdsa.GenerateKey(
		elliptic.P384(),
		rand.Reader,
	)
	if e != nil {
		return
	}

	privateKeyPKCS8Bytes, e = x509.MarshalPKCS8PrivateKey(privateKey)
	if e != nil {
		return
	}

	c = &TLSCertificate{}

	c.privateKeyPEM, e = ioutil.TempFile(directory, pattern)
	if e != nil {
		return
	}

	privateKeyPEMBlock = &pem.Block{
		Type:  privateKeyType,
		Bytes: privateKeyPKCS8Bytes,
	}

	e = pem.Encode(c.privateKeyPEM, privateKeyPEMBlock)
	if e != nil {
		return
	}

	e = c.privateKeyPEM.Chmod(filePermission)
	if e != nil {
		return
	}

	e = c.privateKeyPEM.Close()
	if e != nil {
		return
	}

	certificateTemplate = &x509.Certificate{
		SerialNumber: big.NewInt(serialNumber),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(certValidDuration),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{ip},
	}

	certificateBytes, e = x509.CreateCertificate(
		rand.Reader,
		certificateTemplate,
		certificateTemplate,
		&privateKey.PublicKey,
		privateKey,
	)
	if e != nil {
		return
	}

	c.certificatePEM, e = ioutil.TempFile(directory, pattern)
	if e != nil {
		return
	}

	certificatePEMBlock = &pem.Block{
		Type:  certificateType,
		Bytes: certificateBytes,
	}

	e = pem.Encode(c.certificatePEM, certificatePEMBlock)
	if e != nil {
		return
	}

	e = c.certificatePEM.Close()
	if e != nil {
		return
	}

	return
}

func (c *TLSCertificate) PathToCertificatePEM() string {
	return c.certificatePEM.Name()
}

func (c *TLSCertificate) PathToPrivateKeyPEM() string {
	return c.privateKeyPEM.Name()
}

func (c *TLSCertificate) Remove() (e error) {
	e = os.Remove(
		c.privateKeyPEM.Name(),
	)
	if e != nil {
		return
	}

	e = os.Remove(
		c.certificatePEM.Name(),
	)
	if e != nil {
		return
	}

	return
}
