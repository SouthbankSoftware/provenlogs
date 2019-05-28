/*
 * @Author: guiguan
 * @Date:   2019-05-20T15:53:10+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-27T18:05:59+10:00
 */

package provenlogs

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/SouthbankSoftware/provendb-verify/pkg/proof"
	veriStatus "github.com/SouthbankSoftware/provendb-verify/pkg/proof/status"
	"github.com/SouthbankSoftware/provenlogs/pkg/provendb"
	"github.com/SouthbankSoftware/provenlogs/pkg/rsakey"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Verifier represents a ProvenLogs Verifier instance
type Verifier struct {
	pubKeyPath string
	pubKey     *rsa.PublicKey
	logCol     *mongo.Collection
	pdb        *provendb.ProvenDB
	parser     Parser
}

// NewVerifier creates a new ProvenLogs Verifier
func NewVerifier(
	ctx context.Context,
	provenDBURI string,
	provenDBColName string,
	pubKeyPath string,
) (*Verifier, error) {
	pubPEM, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return nil, err
	}

	pub, err := rsakey.ImportPublicKeyFromPEM(pubPEM)
	if err != nil {
		return nil, err
	}

	logCol, err := getLogCol(ctx, provenDBURI, provenDBColName)
	if err != nil {
		return nil, err
	}

	return &Verifier{
		pubKeyPath: pubKeyPath,
		pubKey:     pub,
		logCol:     logCol,
		pdb:        provendb.NewProvenDB(logCol.Database()),
		parser:     NewZapProductionParser(),
	}, nil
}

// Verify verifies the given raw log string
func (v *Verifier) Verify(
	ctx context.Context,
	raw string,
) (result *VerificationResult, er error) {
	result = &VerificationResult{
		RawLog:     raw,
		PubKeyPath: v.pubKeyPath,
	}

	entry := v.parser.Parse(raw)

	// locate the raw log entry in ProvenDB
	err := v.logCol.FindOne(ctx, &entry).Decode(&entry)
	if err != nil {
		er = fmt.Errorf("failed to locate the log entry in ProvenDB: %s", err)
		return
	}

	jsonStr, err := bson.MarshalExtJSONWithRegistry(DefaultBSONRegistry(), entry, true, true)
	if err != nil {
		er = err
		return
	}

	result.StoredLog = string(jsonStr)
	result.BatchVersion = entry.PDBMetadata.MinVersion

	// retrieve the batch the stored log entry belongs to
	cur, err := v.logCol.Find(ctx, bson.D{
		{provendb.DocKeyMetadataMinVersion, result.BatchVersion},
	}, options.Find().SetSort(bson.D{
		{docKeyTimestamp, 1},
	}))
	if err != nil {
		er = err
		return
	}
	defer cur.Close(ctx)

	entries := LogEntries{}

	for cur.Next(ctx) {
		ent := LogEntry{}
		err := cur.Decode(&ent)
		if err != nil {
			er = err
			return
		}

		entries = append(entries, ent)
	}

	// verify the last log entry's hash
	lastEntry := entries[len(entries)-1]
	err = lastEntry.Verify()
	if err != nil {
		er = fmt.Errorf("the last log entry that contains the batch signature is falsified: %s", err)
		return
	}
	result.BatchSigHash = lastEntry.PDBMetadata.Hash

	// verify the batch signature
	sig := entries.DetachSig()
	if sig == "" {
		er = errors.New("batch signature is not found in the last log entry")
		return
	}

	err = entries.Verify(sig, v.pubKey)
	if err != nil {
		er = fmt.Errorf("invalid batch RSA signature: %s", err)
		return
	}

	// verify the batch signature existence via ProvenDB
	gDPR, err := v.pdb.GetDocumentProof(
		ctx,
		v.logCol.Name(),
		bson.D{
			{docKeyID, lastEntry.PDBMetadata.ID},
		},
		result.BatchVersion,
		"binary",
	)
	if err != nil {
		er = fmt.Errorf("failed to get document proof for the last log entry that contains the batch signature: %s", err)
		return
	}

	if len(gDPR.Proofs) == 0 {
		er = errors.New("no document proof is returned for the last log entry that contains the batch signature")
		return
	}

	var validProof *provendb.DocumentProof

	for _, pf := range gDPR.Proofs {
		if pf.Status == "Valid" {
			validProof = &pf
			break
		}
	}

	if validProof == nil {
		er = errors.New("no valid document proof is returned yet for the last log entry that contains the batch signature. The valid proof should be on the way. Try again later")
		return
	}

	proofBA, ok := validProof.Proof.(primitive.Binary)
	if !ok {
		er = errors.New("cannot convert document proof to `primitive.Binary`")
		return
	}

	result.BatchSigProof = base64.StdEncoding.EncodeToString(proofBA.Data)

	// retrieve the BTC confirmed timestamp
	gPR, err := v.pdb.GetProof(
		ctx,
		validProof.VersionProofID,
		"binary",
		false,
	)
	if err != nil {
		er = fmt.Errorf("failed to retrieve the Bitcoin confirmed timestamp: %s", err)
		return
	}

	if len(gPR.Proofs) == 0 {
		er = errors.New("no version proof is returned when retrieving the Bitcoin confirmed timestamp")
		return
	}

	versionProof := gPR.Proofs[0]
	versionProofDetails := versionProof.Details

	ts, err := time.Parse(time.RFC3339, versionProofDetails.BTCTxnConfirmed)
	if err != nil {
		er = fmt.Errorf("failed to parse the Bitcoin confirmed timestamp: %s", ts)
		return
	}

	ts = ts.Local()

	result.BatchSigBTCConfirmedTimestamp = ts
	result.BatchSigBTCTxn = versionProofDetails.BTCTxn
	result.BatchSigBTCBlock = versionProofDetails.BTCBlock

	// verify the document proof
	vST, err := proof.Verify(ctx, bytes.NewReader([]byte(result.BatchSigProof)))
	if vST == veriStatus.VerificationStatusUnverifiable && err != nil {
		er = fmt.Errorf("unable to verify the document proof for the last log entry that contains the batch signature: %s", err)
		return
	} else if vST == veriStatus.VerificationStatusFalsified && err != nil {
		er = fmt.Errorf("the document proof for the last log entry that contains the batch signature is falsified: %s", err)
		return
	}

	return
}
