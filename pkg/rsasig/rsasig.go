/*
 * @Author: guiguan
 * @Date:   2019-05-23T22:15:59+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-24T13:38:42+10:00
 */

package rsasig

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
)

// Sign generates the RSA signature with the given data's SHA256 hash and RSA private key
func Sign(hash []byte, prv *rsa.PrivateKey) (string, error) {
	sig, err := rsa.SignPSS(rand.Reader, prv, crypto.SHA256, hash, nil)
	if err != nil {
		return "", err
	}

	sigStr := base64.StdEncoding.EncodeToString(sig)
	return sigStr, nil
}

// Verify verifies the RSA signature with the given data's SHA256 hash and RSA public key
func Verify(hash []byte, sigStr string, pub *rsa.PublicKey) error {
	sig, err := base64.StdEncoding.DecodeString(sigStr)
	if err != nil {
		return err
	}

	return rsa.VerifyPSS(pub, crypto.SHA256, hash, sig, nil)
}
