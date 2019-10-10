package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

// _GenerateSSHKeys generates a new private and public keypair
func _GenerateSSHKeys(keyPath string) error {
	// Create new private key of length 2048
	// TODO: Add capability to specify different sizes
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Generate a PEM structure using the private key
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Create the destination private key file
	privateKeyFile, err := os.Create(keyPath)
	defer privateKeyFile.Close()
	if err != nil {
		return err
	}

	// Write out the PEM object to the private key file
	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return err
	}

	// Get our public key part from the private key we generated
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}

	log.Printf("New host key fingerprint: %s", ssh.FingerprintSHA256(publicKey))

	// Write the public key into the provided path
	publicKeyPath := keyPath + ".pub"
	return ioutil.WriteFile(publicKeyPath,
		ssh.MarshalAuthorizedKey(publicKey),
		0644)
}
