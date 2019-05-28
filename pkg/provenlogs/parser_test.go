/*
 * @Author: guiguan
 * @Date:   2019-05-20T14:45:05+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-24T16:36:24+10:00
 */

package provenlogs

import (
	"reflect"
	"testing"
	"time"
)

func TestParser_ZapProductionParser(t *testing.T) {
	type args struct {
		raw string
	}
	tests := []struct {
		name string
		p    *ZapProductionParser
		args args
		want LogEntry
	}{
		{
			name: "parse string",
			args: args{
				raw: "this is a log message",
			},
			want: LogEntry{
				Level:   "info",
				Message: "this is a log message",
				Data:    map[string]interface{}{},
			},
		},
		{
			name: "parse a JSON",
			args: args{
				raw: `{"level":"info","ts":1558485057.335567,"caller":"playground/playground.go:145","msg":"test","a":{"a1": 1, "a2": true}}`,
			},
			want: LogEntry{
				Timestamp: time.Unix(1558485057, 336000000),
				Level:     "info",
				Message:   "test",
				Data: map[string]interface{}{
					"caller": "playground/playground.go:145",
					"a": map[string]interface{}{
						"a1": float64(1),
						"a2": true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewZapProductionParser()

			got := p.Parse(tt.args.raw)

			if tt.want.Timestamp.IsZero() {
				tt.want.Timestamp = got.Timestamp
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
