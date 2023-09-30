package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateECDSAKeyPair(t *testing.T) {
	privateKeyPath := "./private_key.pem"
	publicKeyPath := "./public_key.pem"

	_, err := GenerateECDSAKeyPair(privateKeyPath, publicKeyPath)
	assert.NoError(t, err)

	// Verify that private key file exists and is valid
	privateKeyPEM, err := ioutil.ReadFile(privateKeyPath)
	assert.NoError(t, err)

	block, _ := pem.Decode(privateKeyPEM)
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	assert.NoError(t, err)

	// Verify that public key file exists and is valid
	publicKeyPEM, err := ioutil.ReadFile(publicKeyPath)
	assert.NoError(t, err)

	block, _ = pem.Decode(publicKeyPEM)
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	assert.NoError(t, err)

	ecdsaPublicKey, ok := publicKey.(*ecdsa.PublicKey)
	assert.True(t, ok)

	assert.True(t, privateKey.PublicKey.X.Cmp(ecdsaPublicKey.X) == 0)
	assert.True(t, privateKey.PublicKey.Y.Cmp(ecdsaPublicKey.Y) == 0)

	assert.True(t, ok)
}

func TestSignAndVerify(t *testing.T) {
	// Generate a key pair
	privateKeyPath := "./private_key.pem"
	publicKeyPath := "./public_key.pem"
	_, err := GenerateECDSAKeyPair(privateKeyPath, publicKeyPath)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	// Read the key pair from file
	keyPairFile, err := os.Open("./keypair.json")
	if err != nil {
		t.Fatalf("failed to open keypair file: %v", err)
	}
	defer keyPairFile.Close()

	var kp Keypair
	err = json.NewDecoder(keyPairFile).Decode(&kp)
	if err != nil {
		t.Fatalf("failed to decode keypair file: %v", err)
	}

	// Sign a message
	message := []byte("hello world")
	signature, err := kp.Sign(message)
	if err != nil {
		t.Fatalf("failed to sign message: %v", err)
	}

	// Verify the signature
	valid := SignatureVerify(kp.Public, signature, message)
	if !valid {
		t.Fatalf("signature verification failed")
	}
}
