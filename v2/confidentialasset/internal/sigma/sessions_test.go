package sigma

import "testing"

func TestBCS_sessions_lengths(t *testing.T) {
	t.Parallel()
	var sender, recipient, token [32]byte
	sender[0] = 1
	recipient[0] = 2
	token[0] = 3

	w := BCSWithdrawSession(sender, token, 8, false)
	if len(w) < 70 {
		t.Fatalf("withdraw session len=%d", len(w))
	}
	tr := BCSTransferSession(sender, recipient, token, 8, 4, false, 0)
	if len(tr) < 100 {
		t.Fatalf("transfer session len=%d", len(tr))
	}
	kr := BCSKeyRotationSession(sender, token, 8)
	if len(kr) < 70 {
		t.Fatalf("key rotation session len=%d", len(kr))
	}
}
