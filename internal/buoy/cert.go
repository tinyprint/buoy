package buoy

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

func domainToFileName(domain string) string {
	return strings.ReplaceAll(domain, ".", "-")
}

func getCertPath(configDir string, domain string) string {
	return path.Join(configDir, domainToFileName(domain)+"-cert.pem")
}

func getKeyPath(configDir string, domain string) string {
	return path.Join(configDir, domainToFileName(domain)+"-key.pem")
}

// checkCert checks if a certificate exists for the given domain and returns the path to the
// certificate and key
func checkCert(configDir string, domain string) (string, string, bool) {
	certFile := getCertPath(configDir, domain)
	keyFile := getKeyPath(configDir, domain)

	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		return "", "", false
	}

	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return "", "", false
	}

	return certFile, keyFile, true
}

// generateCert generates a self-signed certificate for the given domain and returns the path to the
// certificate and key
func generateCert(configDir string, domain string) (string, string, error) {
	priv, generatePrivateKeyError := rsa.GenerateKey(rand.Reader, 2048)
	if generatePrivateKeyError != nil {
		return "", "", fmt.Errorf("failed to generate private key: %s", generatePrivateKeyError)
	}

	keyUsage := x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, genSerialError := rand.Int(rand.Reader, serialNumberLimit)
	if genSerialError != nil {
		return "", "", fmt.Errorf("failed to generate serial number: %s", genSerialError)
	}

	dnsNames := []string{domain, "*." + domain}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"buoy"},
			CommonName:   domain,
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		DNSNames: dnsNames,
	}

	certBytes, createCertError := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&priv.PublicKey,
		priv,
	)
	if createCertError != nil {
		return "", "", fmt.Errorf("failed to create certificate: %s", createCertError)
	}

	certOut, err := os.Create(getCertPath(configDir, domain))
	if err != nil {
		return "", "", fmt.Errorf("failed to create temporary file for cert: %s", err)
	}

	fmt.Printf("# cert - writing cert file %s...", certOut.Name())

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return "", "", fmt.Errorf("failed to write data to %s: %s", certOut.Name(), err)
	}
	if err := certOut.Close(); err != nil {
		return "", "", fmt.Errorf("error closing %s: %s", certOut.Name(), err)
	}

	fmt.Println("done")

	keyOut, err := os.OpenFile(getKeyPath(configDir, domain), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", "", fmt.Errorf("failed to create temporary file for cert: %s", err)
	}

	fmt.Printf("# cert - creating key file %s...", keyOut.Name())

	privBytes, privMarshalError := x509.MarshalPKCS8PrivateKey(priv)
	if privMarshalError != nil {
		return "", "", fmt.Errorf("unable to marshal private key: %s", privMarshalError)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return "", "", fmt.Errorf("failed to write data to %s: %s", keyOut.Name(), err)
	}
	if err := keyOut.Close(); err != nil {
		return "", "", fmt.Errorf("error closing %s: %s", keyOut.Name(), err)
	}

	fmt.Println("done")

	return certOut.Name(), keyOut.Name(), nil
}

func trustCert(certPath string, sudoPassword string) error {
	cmd := exec.Command(
		"sudo",
		"-S",
		"security",
		"add-trusted-cert",
		"-d",
		"-r", "trustRoot",
		"-k", "/Library/Keychains/System.keychain",
		certPath,
	)
	cmd.Stdin = strings.NewReader(sudoPassword)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add trusted cert: %s", err)
	}

	return nil
}

// GetOrCreateCert gets or creates a certificate for the given domain and returns the path to the
// certificate and key. If the certificate is created, it is also added to the system keychain.
func GetOrCreateCert(
	configDir string,
	domain string,
	getPassword func() string,
) (string, string, error) {
	certPath, keyPath, certExists := checkCert(configDir, domain)
	if certExists {
		return certPath, keyPath, nil
	}

	certPath, keyPath, certGenError := generateCert(configDir, domain)
	if certGenError != nil {
		return "", "", fmt.Errorf("failed to generate cert: %s", certGenError)
	}

	fmt.Println("# cert - trusting cert, this will require your sudo password...")
	sudoPassword := getPassword()
	trustCertError := trustCert(certPath, sudoPassword)
	if trustCertError != nil {
		return "", "", fmt.Errorf("failed to trust cert: %s", trustCertError)
	}

	return certPath, keyPath, nil
}
