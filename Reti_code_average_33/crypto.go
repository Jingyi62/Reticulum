package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/tv42/base58"
)

type Keypair struct {
	Public  []byte `json:"public"`  // base58 (x y)
	Private []byte `json:"private"` // d (base58 encoded)
}

func bigJoin(expectedLen int, bigs ...*big.Int) *big.Int {

	bs := []byte{}
	for i, b := range bigs {

		by := b.Bytes()
		dif := expectedLen - len(by)
		if dif > 0 && i != 0 {

			by = append(ArrayOfBytes(dif, 0), by...)
		}

		bs = append(bs, by...)
	}

	b := new(big.Int).SetBytes(bs)

	return b
}

func GenerateECDSAKeyPair(privateKeyPath, publicKeyPath string) (*Keypair, error) {
	// Generate ECDSA key pair using P-224 curve
	privateKey, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	if err != nil {
		return nil, err
	}

	pk, _ := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)

	b := bigJoin(28, pk.PublicKey.X, pk.PublicKey.Y)

	public := base58.EncodeBig([]byte{}, b)
	private := base58.EncodeBig([]byte{}, pk.D)

	kp := &Keypair{Public: public, Private: private}

	// Save private key to file
	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		return nil, err
	}
	defer privateKeyFile.Close()

	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	privateKeyPEM := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return nil, err
	}

	// Save public key to file
	publicKeyFile, err := os.Create(publicKeyPath)
	if err != nil {
		return nil, err
	}
	defer publicKeyFile.Close()

	publicKeyPEM, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyPEM,
	}

	if err := pem.Encode(publicKeyFile, publicKeyBlock); err != nil {
		return nil, err
	}

	// Save Keypair to file
	keyPairFile, err := os.Create("./keypair.json")
	if err != nil {
		return nil, err
	}
	defer keyPairFile.Close()

	keyPairBytes, err := json.Marshal(kp)
	if err != nil {
		return nil, err
	}

	if _, err := keyPairFile.Write(keyPairBytes); err != nil {
		return nil, err
	}

	return kp, nil
}

func splitBig(b *big.Int, parts int) []*big.Int {

	bs := b.Bytes()
	if len(bs)%2 != 0 {
		bs = append([]byte{0}, bs...)
	}

	l := len(bs) / parts
	as := make([]*big.Int, parts)

	for i := range as {

		as[i] = new(big.Int).SetBytes(bs[i*l : (i+1)*l])
	}

	return as

}
func (k *Keypair) Sign(hash []byte) ([]byte, error) {
	d, err := base58.DecodeToBig(k.Private)
	if err != nil {
		return nil, err
	}

	b, _ := base58.DecodeToBig(k.Public)

	pub := splitBig(b, 2)
	x, y := pub[0], pub[1]

	key := ecdsa.PrivateKey{PublicKey: ecdsa.PublicKey{Curve: elliptic.P224(), X: x, Y: y}, D: d}

	r, s, _ := ecdsa.Sign(rand.Reader, &key, hash)

	return base58.EncodeBig([]byte{}, bigJoin(28, r, s)), nil
}

func DecodePublicKey(publicKey []byte) (*ecdsa.PublicKey, error) {
	// Decode the PEM-encoded public key
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	// Parse the DER-encoded public key
	parsedKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DER-encoded public key: %w", err)
	}

	// Check if the parsed key is an ECDSA public key
	ecdsaPublicKey, ok := parsedKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("parsed key is not an ECDSA public key")
	}

	return ecdsaPublicKey, nil
}
func SignatureVerify(publicKey, sig, hash []byte) bool {

	b, _ := base58.DecodeToBig(publicKey)
	publ := splitBig(b, 2)
	x, y := publ[0], publ[1]

	b, _ = base58.DecodeToBig(sig)
	sigg := splitBig(b, 2)
	r, s := sigg[0], sigg[1]

	pub := ecdsa.PublicKey{Curve: elliptic.P224(), X: x, Y: y}

	return ecdsa.Verify(&pub, hash, r, s)
}
