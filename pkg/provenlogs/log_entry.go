/*
 * @Author: guiguan
 * @Date:   2019-05-21T16:22:18+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-21T16:33:07+10:00
 */

package provenlogs

import "time"

// LogEntry represents an internal log entry data structure
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp" bson:"timestamp"`
	Message   string                 `json:"message" bson:"message"`
	Data      map[string]interface{} `json:"data" bson:"data"`
}
