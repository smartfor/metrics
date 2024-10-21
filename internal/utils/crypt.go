package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

// GenerateSymmetricKey generates a new AES symmetric key
func GenerateSymmetricKey() ([]byte, error) {
	key := make([]byte, 32) // AES-256
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// EncryptWithSymmetricKey encrypts data with a symmetric key
func EncryptWithSymmetricKey(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// EncryptWithPublicKey encrypts data using hybrid encryption (AES + RSA)
func EncryptWithPublicKey(data []byte, publicKey []byte) ([]byte, []byte, error) {
	symmetricKey, err := GenerateSymmetricKey()
	if err != nil {
		return nil, nil, err
	}

	encryptedData, err := EncryptWithSymmetricKey(data, symmetricKey)
	if err != nil {
		return nil, nil, err
	}

	encryptedKey, err := encryptKeyWithPublicKey(symmetricKey, publicKey)
	if err != nil {
		return nil, nil, err
	}

	return encryptedData, encryptedKey, nil
}

// encryptKeyWithPublicKey encrypts a symmetric key with a public key
func encryptKeyWithPublicKey(key []byte, publicKey []byte) ([]byte, error) {
	fmt.Println("publicKey: ", string(publicKey), "key: ", string(key))
	block, _ := pem.Decode(publicKey)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not RSA public key")
	}

	return rsa.EncryptPKCS1v15(rand.Reader, rsaPub, key)
}

// DecryptWithPrivateKey decrypts data using hybrid decryption (AES + RSA)
func DecryptWithPrivateKey(encryptedData []byte, encryptedKey []byte, privateKey []byte) ([]byte, error) {
	symmetricKey, err := decryptKeyWithPrivateKey(encryptedKey, privateKey)
	if err != nil {
		return nil, err
	}

	return DecryptWithSymmetricKey(encryptedData, symmetricKey)
}

// decryptKeyWithPrivateKey decrypts a symmetric key with a private key
func decryptKeyWithPrivateKey(encryptedKey []byte, privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing the key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return rsa.DecryptPKCS1v15(rand.Reader, priv, encryptedKey)
}

// DecryptWithSymmetricKey decrypts data with a symmetric key
func DecryptWithSymmetricKey(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
