/*
 * @Author: guiguan
 * @Date:   2019-05-16T13:06:33+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-27T17:37:42+10:00
 */

package provendb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ProvenDB represents a ProvenDB database instance
type ProvenDB struct {
	db *mongo.Database
}

// NewProvenDB creates a new ProvenDB functional wrapper around a `*mongo.Database`
func NewProvenDB(db *mongo.Database) *ProvenDB {
	return &ProvenDB{
		db: db,
	}
}

// Database returns the wrapped `*mongo.Database`
func (p *ProvenDB) Database() *mongo.Database {
	return p.db
}

// ShowMetaData controls the display of `_provendb_metadata` information document for find commands
// on user collections. ProvenDB metadata defines the version numbers for which a given document is
// valid, and contains the document hash value which is used to pin the document to the blockchain.
func (p *ProvenDB) ShowMetaData(
	ctx context.Context,
	show bool,
) error {
	return p.db.RunCommand(ctx, bson.D{
		{"showMetadata", show},
	}).Err()
}

// GetVersionResult represents the result of a `GetVersion` command
type GetVersionResult struct {
	Response string `json:"reponse" bson:"reponse"`
	Version  int64  `json:"version" bson:"version"`
	Status   string `json:"status" bson:"status"`
}

// GetVersion retrieves the active version for the current session. If no version has been set by
// the user using the setVersion command, the current version from the database will be returned
func (p *ProvenDB) GetVersion(
	ctx context.Context,
) (*GetVersionResult, error) {
	result := GetVersionResult{}

	err := p.db.RunCommand(ctx, bson.D{
		{"getVersion", 1},
	}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// SubmitProofResult represents the result of a `SubmitProof` command
type SubmitProofResult struct {
	DateTime time.Time `json:"dateTime" bson:"dateTime"`
	Hash     string    `json:"hash" bson:"hash"`
	ProofID  string    `json:"proofId" bson:"proofId"`
	Status   string    `json:"status" bson:"status"`
}

// SubmitProof submits a proof to the blockchain for the specified version and returns a receipt
func (p *ProvenDB) SubmitProof(
	ctx context.Context,
	version int64,
) (*SubmitProofResult, error) {
	result := SubmitProofResult{}

	err := p.db.RunCommand(ctx, bson.D{
		{"submitProof", version},
	}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// VerifyProofResult represents the result of a `VerifyProof` command
type VerifyProofResult struct {
	Version        int64       `json:"version" bson:"version"`
	DateTime       time.Time   `json:"dateTime" bson:"dateTime"`
	ProofID        string      `json:"proofId" bson:"proofId"`
	ProofStatus    string      `json:"proofStatus" bson:"proofStatus"`
	BTCTransaction string      `json:"btcTransaction" bson:"btcTransaction"`
	BTCBlockNumber int64       `json:"btcBlockNumber" bson:"btcBlockNumber"`
	Proof          interface{} `json:"proof" bson:"proof"`
}

// VerifyProof verifies a proof by recalculating the root hash and comparing that to the hash found
// in the Chainpoint receipt and verifying receipt on the blockchain
func (p *ProvenDB) VerifyProof(
	ctx context.Context,
	proofID string,
	format string,
) (*VerifyProofResult, error) {
	result := VerifyProofResult{}

	err := p.db.RunCommand(ctx, bson.D{
		{"verifyProof", proofID},
		{"format", format},
	}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DocumentProof represents a ProvenDB document proof
type DocumentProof struct {
	Collection     string      `json:"collection" bson:"collection"`
	Version        int64       `json:"version" bson:"version"`
	DocumentID     interface{} `json:"documentId" bson:"documentId"`
	VersionProofID string      `json:"versionProofId" bson:"versionProofId"`
	Status         string      `json:"status" bson:"status"`
	StatusMsg      string      `json:"errmsg" bson:"errmsg"`
	DocumentHash   string      `json:"documentHash" bson:"documentHash"`
	VersionHash    string      `json:"versionHash" bson:"versionHash"`
	Proof          interface{} `json:"proof" bson:"proof"`
}

// GetDocumentProofResult represents the result of a `GetDocumentProof` command
type GetDocumentProofResult struct {
	Proofs []DocumentProof `json:"proofs" bson:"proofs"`
}

// GetDocumentProof returns a structured chainpoint format receipt, which cryptographically proves
// that a document for a specific version is included within that versions hash. This can be used to
// prove an individual document is included within a version without having to access other
// documents in a version
func (p *ProvenDB) GetDocumentProof(
	ctx context.Context,
	collection string,
	filter interface{},
	version int64,
	format string,
) (*GetDocumentProofResult, error) {
	result := GetDocumentProofResult{}

	err := p.db.RunCommand(ctx, bson.D{
		{"getDocumentProof", bson.D{
			{"collection", collection},
			{"filter", filter},
			{"version", version},
			{"format", format},
		}},
	}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// VersionProofDetailsProtocol represents the protocol in a VersionProofDetails
type VersionProofDetailsProtocol struct {
	Name               string `json:"name" bson:"name"`
	URI                string `json:"uri" bson:"uri"`
	HashIDNode         string `json:"hashIdNode" bson:"hashIdNode"`
	ChainpointLocation string `json:"chainpointLocation" bson:"chainpointLocation"`
}

// VersionProofDetails represents the details in a VersionProof
type VersionProofDetails struct {
	Protocol        VersionProofDetailsProtocol `json:"protocol" bson:"protocol"`
	BTCTxn          string                      `json:"btcTxn" bson:"btcTxn"`
	BTCTxnReceived  string                      `json:"btcTxnReceived" bson:"btcTxnReceived"`
	BTCTxnConfirmed string                      `json:"btcTxnConfirmed" bson:"btcTxnConfirmed"`
	BTCBlock        string                      `json:"btcBlock" bson:"btcBlock"`
}

// VersionProof represents a ProvenDB version proof
type VersionProof struct {
	ProofID   string              `json:"proofId" bson:"proofId"`
	Version   int64               `json:"version" bson:"version"`
	Submitted time.Time           `json:"submitted" bson:"submitted"`
	Type      string              `json:"type" bson:"type"`
	Hash      string              `json:"hash" bson:"hash"`
	Status    string              `json:"status" bson:"status"`
	Details   VersionProofDetails `json:"details" bson:"details"`
	Proof     interface{}         `json:"proof" bson:"proof"`
}

// GetProofResult represents the result of a `GetProof` command
type GetProofResult struct {
	Proofs []VersionProof `json:"proofs" bson:"proofs"`
}

// GetProof returns a structured document similar to a chainpoint receipt, which cryptographically
// proves that a version proof is valid and on the blockchain
func (p *ProvenDB) GetProof(
	ctx context.Context,
	proofIDOrVersion interface{},
	format string,
	listCollections bool,
) (*GetProofResult, error) {
	result := GetProofResult{}

	err := p.db.RunCommand(ctx, bson.D{
		{"getProof", proofIDOrVersion},
		{"format", format},
		{"listCollections", listCollections},
	}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
