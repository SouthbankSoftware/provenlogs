/*
 * @Author: guiguan
 * @Date:   2019-05-15T15:07:19+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-15T17:15:44+10:00
 */

package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"github.com/SouthbankSoftware/provenlogs/pkg/rsakey"
)

func generatePrvPubKeys() {
	prv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	prvPEM := rsakey.ExportPrivateKeyToPEM(prv)

	err = ioutil.WriteFile("key-prv.pem", prvPEM, 0600)
	if err != nil {
		panic(err)
	}

	pubPEM, err := rsakey.ExportPublicKeyToPEM(&prv.PublicKey)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("key-pub.pem", pubPEM, 0644)
	if err != nil {
		panic(err)
	}
}

func signData() {
	data := []byte("this is a test")

	prvPEM, err := ioutil.ReadFile("key-prv.pem")
	if err != nil {
		panic(err)
	}

	prv, err := rsakey.ImportPrivateKeyFromPEM(prvPEM)
	if err != nil {
		panic(err)
	}

	hashed := sha256.Sum256(data)

	sig, err := rsa.SignPSS(rand.Reader, prv, crypto.SHA256, hashed[:], nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(base64.StdEncoding.EncodeToString(sig))
}

func verifySignature() {
	data := []byte("this is a test")

	sig, err := base64.StdEncoding.DecodeString("4R2H6IRjQ+AOOwiw0reInQJjKTuG4eQqWUbWuheNwdTjATEgKB78n2m4yfn9IFMRVwF+S3e0f2rrihV83Sc82Sa7kJ5eG1fjAY9ZPJ7uYXMWsw/m/1NokEoUvqql1qdUpyVO6C6THpNCrjLkleZjX45sbBbM9NYzzWkCiR+RNTGZgIkXQtsG9j10EqQg37Sq2L5ccGQLz8pS9zJMTY4aWDgFCPABbwpm8y7ZkD5QxmNQwZoQyXGkSxVntjWioNHZ7PoE4Ux7Xsc84o2KVyX5K5MVt3AISk79F0e0iGfpl4v969caj63Styfj51N19vdVsYAHo1WjkwatVTMS93+Jtw==")
	if err != nil {
		panic(err)
	}

	pubPEM, err := ioutil.ReadFile("key-pub.pem")
	if err != nil {
		panic(err)
	}

	pub, err := rsakey.ImportPublicKeyFromPEM(pubPEM)
	if err != nil {
		panic(err)
	}

	hashed := sha256.Sum256(data)

	err = rsa.VerifyPSS(pub, crypto.SHA256, hashed[:], sig, nil)
	if err != nil {
		fmt.Printf("%#v\n", err)
		return
	}

	fmt.Println("verified")
}

func main() {
	// generatePrvPubKeys()
	// signData()
	verifySignature()
}
