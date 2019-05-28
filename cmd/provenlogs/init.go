/*
 * @Author: guiguan
 * @Date:   2019-05-23T21:54:44+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-24T15:47:54+10:00
 */

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io/ioutil"

	"github.com/SouthbankSoftware/provenlogs/pkg/rsakey"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	initName            = "init"
	keySize             = "public-key"
	viperKeyInitKeySize = initName + "." + keySize
)

func initInitCMD(rootCMD *cobra.Command) {
	cmd := &cobra.Command{
		Use:   initName,
		Short: "init ProvenLogs by generating a RSA public/private key pair in PEM format in current working directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			prv, err := rsa.GenerateKey(rand.Reader, viper.GetInt(viperKeyInitKeySize))
			if err != nil {
				return err
			}

			prvPEM := rsakey.ExportPrivateKeyToPEM(prv)

			err = ioutil.WriteFile(defaultPrivateKeyName, prvPEM, 0600)
			if err != nil {
				return err
			}

			pubPEM, err := rsakey.ExportPublicKeyToPEM(&prv.PublicKey)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(defaultPublicKeyName, pubPEM, 0644)
			if err != nil {
				return err
			}

			fmt.Println("successfully inited ProvenLogs")

			return nil
		},
	}

	cmd.Flags().IntP(keySize, "s", 2048, "specify the RSA key size in number of bits")
	viper.BindPFlag(viperKeyInitKeySize, cmd.Flags().Lookup(keySize))

	rootCMD.AddCommand(cmd)
}
