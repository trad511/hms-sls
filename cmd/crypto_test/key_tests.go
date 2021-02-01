// MIT License
//
// (C) Copyright [2021] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

func GenerateRsaKeyPair() (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 4096)
	return privateKey, &privateKey.PublicKey
}

func ExportRsaPrivateKeyAsPemStr(privkey *rsa.PrivateKey) string {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privkey)
	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privateKeyBytes,
		},
	)
	return string(privateKeyPEM)
}

func ParseRsaPrivateKeyFromPemStr(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func ExportRsaPublicKeyAsPemStr(publicKey *rsa.PublicKey) (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}
	publicKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKeyBytes,
		},
	)

	return string(publicKeyPEM), nil
}

func ParseRsaPublicKeyFromPemStr(pubPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		break
	}
	return nil, errors.New("key type is not RSA")
}

func EncryptBytes(bytes []byte, publicKey *rsa.PublicKey) []byte {
	encryptedBytes, encryptErr := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, bytes, nil)
	if encryptErr != nil {
		panic(encryptErr)
	}

	return encryptedBytes
}

func DecryptBytes(bytes []byte, privateKey *rsa.PrivateKey) []byte {
	decryptedBytes, decryptErr := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, bytes, nil)
	if decryptErr != nil {
		panic(decryptErr)
	}

	return decryptedBytes
}

func main() {
	// Create the keys.
	privateKey, publicKey := GenerateRsaKeyPair()

	// Export the keys to pem string.
	privPem := ExportRsaPrivateKeyAsPemStr(privateKey)
	pubPem, _ := ExportRsaPublicKeyAsPemStr(publicKey)

	// Import the keys from pem string.
	privateKeyParsed, _ := ParseRsaPrivateKeyFromPemStr(privPem)
	publicKeyParsed, _ := ParseRsaPublicKeyFromPemStr(pubPem)

	// Export the newly imported keys.
	privateKeyParsedPem := ExportRsaPrivateKeyAsPemStr(privateKeyParsed)
	publicKeyParsedPem, _ := ExportRsaPublicKeyAsPemStr(publicKeyParsed)

	fmt.Println(privateKeyParsedPem)
	fmt.Println(publicKeyParsedPem)

	// Check that the exported/imported keys match the original keys.
	if privPem != privateKeyParsedPem || pubPem != publicKeyParsedPem {
		fmt.Println("Failure: Export and Import did not result in same keys!!!")
	} else {
		fmt.Println("Success - export/import see the same keys.")
	}


	// Test encryption/decryption
	inputText := "The quick brown fox jumps over the lazy dog."
	inputBytes := []byte(inputText)
	//inputBytes := []byte(fmt.Sprintf("%v", baz))
	ciphertext := EncryptBytes(inputBytes, publicKeyParsed)
	base64EncryptedStringEncoded := base64.StdEncoding.EncodeToString(ciphertext)
	fmt.Printf("\nBase 64 encoded encrypted string:\n%s\n", base64EncryptedStringEncoded)

	base64EncryptedStringDecoded, _ := base64.StdEncoding.DecodeString(base64EncryptedStringEncoded)
	decryptedBytes := DecryptBytes(base64EncryptedStringDecoded, privateKeyParsed)
	decryptedText := string(decryptedBytes)
	fmt.Printf("\nDecrypted info:\n%s\n", string(decryptedText))
}
