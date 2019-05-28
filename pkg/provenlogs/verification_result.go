/*
 * @Author: guiguan
 * @Date:   2019-05-25T11:03:30+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-27T18:04:33+10:00
 */

package provenlogs

import "time"

// VerificationResult represents a log entry's verification result
type VerificationResult struct {
	RawLog                        string
	StoredLog                     string
	PubKeyPath                    string
	BatchVersion                  int64
	BatchSigHash                  string
	BatchSigProof                 string
	BatchSigBTCConfirmedTimestamp time.Time
	BatchSigBTCTxn                string
	BatchSigBTCBlock              string
}
