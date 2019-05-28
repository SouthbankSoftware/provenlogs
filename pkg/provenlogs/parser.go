/*
 * @Author: guiguan
 * @Date:   2019-05-20T14:45:05+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-25T10:41:55+10:00
 */

package provenlogs

import (
	"encoding/json"
	"math"
	"time"
)

// Parser represents a log parser
type Parser interface {
	Parse(raw string) LogEntry
}

// ZapProductionParser represents a parser for production logs from Zap logger
type ZapProductionParser struct {
	TimestampKey,
	LevelKey,
	MessageKey string
}

// NewZapProductionParser creates a new parser for production Zap
func NewZapProductionParser() *ZapProductionParser {
	return &ZapProductionParser{
		TimestampKey: "ts",
		LevelKey:     "level",
		MessageKey:   "msg",
	}
}

// Parse parses the given raw log string into a log entry
func (p *ZapProductionParser) Parse(raw string) LogEntry {
	result := LogEntry{}
	rawJSON := map[string]interface{}{}

	// try to parse from JSON and ignore errors
	json.Unmarshal([]byte(raw), &rawJSON)

	if ts, ok := rawJSON[p.TimestampKey].(float64); ok {
		// use millisecond precision as `time.Time` is persisted as `ISODate` in MongoDB, which has
		// millisecond precision
		result.Timestamp = time.Unix(0, int64(math.Round(ts*1e3))*1e6)
		delete(rawJSON, p.TimestampKey)
	} else {
		result.Timestamp = time.Now()
	}

	if lvl, ok := rawJSON[p.LevelKey].(string); ok {
		result.Level = lvl
		delete(rawJSON, p.LevelKey)
	} else {
		result.Level = "info"
	}

	if msg, ok := rawJSON[p.MessageKey].(string); ok {
		result.Message = msg
		delete(rawJSON, p.MessageKey)
	} else {
		result.Message = raw
	}

	// the rest goes to data
	result.Data = rawJSON

	return result
}
