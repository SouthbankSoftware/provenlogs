/*
 * @Author: guiguan
 * @Date:   2019-05-23T21:20:34+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-27T18:16:28+10:00
 */

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SouthbankSoftware/provendb-verify/pkg/proof/anchor"
	"github.com/SouthbankSoftware/provenlogs/pkg/provenlogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	anchor.ShowProgress = false
}

const (
	verifyName              = "verify"
	keyPublicKey            = "public-key"
	viperKeyVerifyPublicKey = verifyName + "." + keyPublicKey
	keyRawLogEntry          = "raw-log-entry"
	viperKeyRawLogEntry     = verifyName + "." + keyRawLogEntry
)

func initVerifyCMD(rootCMD *cobra.Command) {
	cmd := &cobra.Command{
		Use:   verifyName,
		Short: "verify a log entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			raw := viper.GetString(viperKeyRawLogEntry)
			if raw == "" {
				return fmt.Errorf("`%s` must be provided", viperKeyRawLogEntry)
			}

			provenDBURI := viper.GetString(viperKeyProvenDBURI)
			if provenDBURI == "" {
				return fmt.Errorf("`%s` must be provided", viperKeyProvenDBURI)
			}

			ctx := context.Background()

			verifier, err := provenlogs.NewVerifier(
				ctx,
				provenDBURI,
				viper.GetString(viperKeyProvenDBColName),
				viper.GetString(viperKeyVerifyPublicKey),
			)
			if err != nil {
				return err
			}

			fmt.Println("The raw log entry:")
			fmt.Printf("\t%s\n", raw)

			result, err := verifier.Verify(ctx, raw)
			if err != nil {
				fmt.Println("is falsified:")
				fmt.Printf("\t%s\n", err)
				os.Exit(1)
			}

			fmt.Println("is found in the batch with ProvenDB version:")
			fmt.Printf("\t%v\n", result.BatchVersion)
			fmt.Println("is stored in ProvenDB as:")
			fmt.Printf("\t%v\n", result.StoredLog)
			fmt.Println("has its batch RSA signature stored in the last log entry of the batch with hash:")
			fmt.Printf("\t%v\n", result.BatchSigHash)
			fmt.Println("has a valid batch RSA signature that has been verified using the batch data stored in ProvenDB and with the public key:")
			p, _ := filepath.Abs(result.PubKeyPath)
			fmt.Printf("\t%v\n", p)
			fmt.Println("has its batch RSA signature proven to be existed on Bitcoin with this ProvenDB document proof (just verified):")
			fmt.Printf("\t%v\n", result.BatchSigProof)
			fmt.Printf("\twhich is confirmed at: %v\n", result.BatchSigBTCConfirmedTimestamp)
			fmt.Printf("\twhich has transaction ID: %v\n", result.BatchSigBTCTxn)
			fmt.Printf("\twhich resides in block: %v\n", result.BatchSigBTCBlock)
			fmt.Println("is verified")

			return nil
		},
	}

	cmd.Flags().StringP(keyPublicKey, "i", defaultPublicKeyName, "specify the RSA public key (.pem) path")
	viper.BindPFlag(viperKeyVerifyPublicKey, cmd.Flags().Lookup(keyPublicKey))
	cmd.Flags().StringP(keyRawLogEntry, "l", "", "specify the raw log entry from Zap production output to verify")
	viper.BindPFlag(viperKeyRawLogEntry, cmd.Flags().Lookup(keyRawLogEntry))

	rootCMD.AddCommand(cmd)
}
