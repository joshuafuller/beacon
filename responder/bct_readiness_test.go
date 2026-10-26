package responder

import (
	"testing"
)

// Tests for behaviour that Apple's Bonjour Conformance Test (BCT 1.5.4)
// checks. Each names the BCT subtest it mirrors. Helpers live in
// rfc_conformance_test.go.

// BCT II.7/II.8: queries with type ANY (255) must be answered with every
// record of the name (RFC 6762 §6, RFC 1035 §3.2.3).
func TestHandleQuery_ANY_AnswersAllRecordsOfName(t *testing.T) {
	tests := []struct {
		name      string
		wantTypes []uint16
	}{
		{"Alpha._http._tcp.local", []uint16{33, 16}}, // SRV + TXT
		{"myhost.local", []uint16{1}},                // A
		{"_http._tcp.local", []uint16{12}},           // PTR
	}
	for _, tc := range tests {
		r, mt := newConformanceResponder(t, "Alpha")
		_ = r.handleQuery(mt, buildTestQuery(t, 0, testQuestion{tc.name, 255}), mdnsSource, 0)
		msgs, _ := sentMessages(t, mt)
		for _, want := range tc.wantTypes {
			if len(answersOfType(msgs, want)) == 0 {
				t.Errorf("ANY %s: no type %d answer", tc.name, want)
			}
		}
		for _, m := range msgs {
			for _, a := range m.Answers {
				if !namesEqual(a.NAME, tc.name) {
					t.Errorf("ANY %s: answer for other name %s", tc.name, a.NAME)
				}
			}
		}
	}
}
