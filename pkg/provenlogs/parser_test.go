/*
 * @Author: guiguan
 * @Date:   2019-05-20T14:45:05+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-21T17:29:00+10:00
 */

package provenlogs

import (
	"reflect"
	"testing"
	"time"
)

func TestParser_Parse(t *testing.T) {
	type args struct {
		raw string
	}
	tests := []struct {
		name string
		p    *Parser
		args args
		want LogEntry
	}{
		{
			name: "parse string",
			args: args{
				raw: "this is a log message",
			},
			want: LogEntry{
				Message: "this is a log message",
			},
		},
		{
			name: "parse a JSON",
			args: args{
				raw: `{"message": "this is another log message", 
"timestamp": "` + time.Now().Format(time.RFC3339) + `",
"data": {"a": 1, "b": true}}`,
			},
			want: LogEntry{
				Message: "this is another log message",
				Data: map[string]interface{}{
					"a": float64(1),
					"b": true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{}
			if got := p.Parse(tt.args.raw); !reflect.DeepEqual(got.Message, tt.want.Message) ||
				!reflect.DeepEqual(got.Data, tt.want.Data) {
				t.Errorf("Parser.Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
