package clients

import "testing"

func TestParseClientPrimaryID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want int64
		ok   bool
	}{
		{name: "numeric id", id: "123", want: 123, ok: true},
		{name: "trim spaces", id: " 42 ", want: 42, ok: true},
		{name: "zero", id: "0", ok: false},
		{name: "negative", id: "-1", ok: false},
		{name: "node id", id: "node-1", ok: false},
		{name: "client ip", id: "127.0.0.1", ok: false},
		{name: "empty", id: "", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseClientPrimaryID(tt.id)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("parseClientPrimaryID(%q) = (%d, %v), want (%d, %v)", tt.id, got, ok, tt.want, tt.ok)
			}
		})
	}
}
