/*
 * @Author: guiguan
 * @Date:   2019-05-21T16:22:18+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-27T10:33:59+10:00
 */

package provenlogs

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/SouthbankSoftware/provenlogs/pkg/provendb"
	"github.com/SouthbankSoftware/provenlogs/pkg/rsasig"
	"go.mongodb.org/mongo-driver/bson"
)

// LogEntry represents an internal log entry data structure
type LogEntry struct {
	PDBMetadata  *provendb.Metadata     `json:"_provendb_metadata,omitempty" bson:"_provendb_metadata,omitempty"`
	Timestamp    time.Time              `json:"timestamp" bson:"timestamp"`
	Level        string                 `json:"level" bson:"level"`
	Message      string                 `json:"message" bson:"message"`
	Data         map[string]interface{} `json:"data,omitempty" bson:"data,omitempty"`
	RSASignature string                 `json:"_rsa_sig,omitempty" bson:"_rsa_sig,omitempty"`
}

// Hash returns the LogEntry's SHA256 hash
func (l *LogEntry) Hash() ([]byte, error) {
	meta := l.PDBMetadata

	if meta != nil {
		// if it contains ProvenDB metadata, make sure it only hashes the `_id` and `minVersion` of
		// the metadata
		en := *l
		en.PDBMetadata = &provendb.Metadata{
			ID:         meta.ID,
			MinVersion: meta.MinVersion,
		}
		l = &en
	}

	bA, err := bson.MarshalWithRegistry(DefaultBSONRegistry(), l)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(bA)

	return hash[:], nil
}

// Verify verifies the LogEntry's hash against the hash in its ProvenDB metadata
func (l *LogEntry) Verify() error {
	meta := l.PDBMetadata

	if meta == nil {
		return errors.New("ProvenDB metadata doesn't exist")
	}

	hash, err := l.Hash()
	if err != nil {
		return err
	}

	hashStr := hex.EncodeToString(hash)

	if hashStr != meta.Hash {
		return fmt.Errorf("log entry hash doesn't match. Expected: %s, actual: %s",
			meta.Hash, hashStr)
	}

	return nil
}

// LogEntries represents a slice of log entries
type LogEntries []LogEntry

// Len is the number of LogEntries
func (l LogEntries) Len() int {
	return len(l)
}

// Less reports whether the LogEntry with index i should sort before the LogEntry with index j
func (l LogEntries) Less(i, j int) bool {
	return l[i].Timestamp.Before(l[j].Timestamp)
}

// Swap swaps the LogEntries with indexes i and j
func (l LogEntries) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// Sort sorts the LogEntries in chronological order of its entries' timestamps
func (l LogEntries) Sort() {
	sort.Sort(l)
}

// Hash returns the LogEntries' SHA256 hash
func (l LogEntries) Hash() ([]byte, error) {
	hasher := sha256.New()

	for _, e := range l {
		// ignore ProvenDB metadata in this hash for RSA signing
		e.PDBMetadata = nil

		bA, err := bson.MarshalWithRegistry(DefaultBSONRegistry(), &e)
		if err != nil {
			return nil, err
		}

		_, err = hasher.Write(bA)
		if err != nil {
			return nil, err
		}
	}

	hash := hasher.Sum(nil)
	return hash, nil
}

// Sign signs the LogEntries using the given RSA private key and returns a base64 encoded RSA
// signature
func (l LogEntries) Sign(prv *rsa.PrivateKey) (string, error) {
	hash, err := l.Hash()
	if err != nil {
		return "", err
	}

	return rsasig.Sign(hash, prv)
}

// Verify verifies the LogEntries's base64 encoded RSA signature with the given RSA public key
func (l LogEntries) Verify(sigStr string, pub *rsa.PublicKey) error {
	hash, err := l.Hash()
	if err != nil {
		return err
	}

	return rsasig.Verify(hash, sigStr, pub)
}

// AttachSig attaches the given RSA signature to the last log entry
func (l LogEntries) AttachSig(sigStr string) {
	l[len(l)-1].RSASignature = sigStr
}

// DetachSig detaches the RSA signature from the last log entry and returns it
func (l LogEntries) DetachSig() string {
	e := &l[len(l)-1]
	sigStr := e.RSASignature
	e.RSASignature = ""

	return sigStr
}

// AnyArray returns the `[]interface{}` version of the LogEntries
func (l LogEntries) AnyArray() []interface{} {
	result := make([]interface{}, 0, len(l))

	for i := range l {
		result = append(result, interface{}(&l[i]))
	}

	return result
}
