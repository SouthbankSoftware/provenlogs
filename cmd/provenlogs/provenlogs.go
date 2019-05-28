/*
 * @Author: guiguan
 * @Date:   2019-05-15T15:38:02+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-24T15:57:30+10:00
 */

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/SouthbankSoftware/provenlogs/pkg/provenlogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	cmdVersion = "0.0.0"
)

const (
	name                    = "provenlogs"
	viperKeyDebug           = "debug"
	viperKeyProvenDBURI     = "provendb.uri"
	viperKeyProvenDBColName = "provendb.col"
	viperKeyPrivateKey      = "private-key"
	viperKeyBatchTime       = "batch.time"
	viperKeyBatchSize       = "batch.size"
	defaultPrivateKeyName   = "key-prv.pem"
	defaultPublicKeyName    = "key-pub.pem"
	defaultProvenDBColName  = "logs"
)

func initLogger(debug bool) (*zap.Logger, error) {
	if debug {
		return zap.NewDevelopment()
	}

	return zap.NewProduction()
}

func initRootCMD() *cobra.Command {
	rootCMD := &cobra.Command{
		Use:     name,
		Short:   "Prove logs on Blockchain with ProvenDB",
		Version: cmdVersion,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger, err := initLogger(viper.GetBool(viperKeyDebug))
			if err != nil {
				return err
			}

			provenDBURI := viper.GetString(viperKeyProvenDBURI)
			if provenDBURI == "" {
				return fmt.Errorf("`%s` must be provided", viperKeyProvenDBURI)
			}

			svc := provenlogs.NewServer(
				provenDBURI,
				viper.GetString(viperKeyProvenDBColName),
				logger,
				viper.GetString(viperKeyPrivateKey),
				viper.GetDuration(viperKeyBatchTime),
				viper.GetInt(viperKeyBatchSize),
			)
			return svc.Run(context.Background())
		},
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix(name)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

	rootCMD.PersistentFlags().BoolP(viperKeyDebug, "d", true, "whether to show debug messages")
	viper.BindPFlag(viperKeyDebug, rootCMD.PersistentFlags().Lookup(viperKeyDebug))
	rootCMD.PersistentFlags().StringP(viperKeyProvenDBURI, "u", "", "specify the ProvenDB URI with the database name")
	viper.BindPFlag(viperKeyProvenDBURI, rootCMD.PersistentFlags().Lookup(viperKeyProvenDBURI))
	rootCMD.PersistentFlags().StringP(viperKeyProvenDBColName, "c", defaultProvenDBColName, "specify the ProvenDB collection name")
	viper.BindPFlag(viperKeyProvenDBColName, rootCMD.PersistentFlags().Lookup(viperKeyProvenDBColName))
	rootCMD.Flags().StringP(viperKeyPrivateKey, "i", defaultPrivateKeyName, "specify the RSA private key (.pem) path")
	viper.BindPFlag(viperKeyPrivateKey, rootCMD.Flags().Lookup(viperKeyPrivateKey))
	rootCMD.Flags().Duration(viperKeyBatchTime, 5*time.Minute, "specify the maximum time each batch can buffer up to")
	viper.BindPFlag(viperKeyBatchTime, rootCMD.Flags().Lookup(viperKeyBatchTime))
	rootCMD.Flags().Int(viperKeyBatchSize, 500, "specify the maximum size each batch can buffer up to")
	viper.BindPFlag(viperKeyBatchSize, rootCMD.Flags().Lookup(viperKeyBatchSize))

	return rootCMD
}

func main() {
	rootCMD := initRootCMD()

	initVerifyCMD(rootCMD)
	initInitCMD(rootCMD)

	if err := rootCMD.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
