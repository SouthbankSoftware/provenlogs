/*
 * @Author: guiguan
 * @Date:   2019-05-26T10:55:00+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-27T00:24:27+10:00
 */

package provendb

const (
	// DocKeyMetadata represents the key path to the metadata of a ProvenDB doc
	DocKeyMetadata = "_provendb_metadata"
	// DocKeyMetadataMinVersion represents the key path to the min version of a ProvenDB doc
	DocKeyMetadataMinVersion = DocKeyMetadata + ".minVersion"
)

// Metadata represents a ProvenDB document's metadata
type Metadata struct {
	ID         interface{} `json:"_id" bson:"_id"`
	MongoID    interface{} `json:"_mongoId,omitempty" bson:"_mongoId,omitempty"`
	MinVersion int64       `json:"minVersion" bson:"minVersion"`
	Hash       string      `json:"hash,omitempty" bson:"hash,omitempty"`
	MaxVersion int64       `json:"maxVersion,omitempty" bson:"maxVersion,omitempty"`
}
