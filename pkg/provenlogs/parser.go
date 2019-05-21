/*
 * @Author: guiguan
 * @Date:   2019-05-20T14:45:05+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-21T16:48:17+10:00
 */

package provenlogs

import (
	"encoding/json"
	"time"
)

// Parser represents a simple string parser
type Parser struct{}

// NewParser creates a new parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses the given string into log entry
func (p *Parser) Parse(raw string) LogEntry {
	result := LogEntry{}

	// first try to parse the string as a JSON object
	err := json.Unmarshal([]byte(raw), &result)
	if err != nil {
		// fallback: use raw as message

		result.Timestamp = time.Now()
		result.Message = raw

		return result
	}

	if result.Timestamp.IsZero() {
		result.Timestamp = time.Now()
	}

	return result
}
