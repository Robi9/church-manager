package auth

import "testing"

func TestSessionTokensAreRandomAndKeyed(t *testing.T) {
	first, err := newSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	second, err := newSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 43 || first == second {
		t.Fatal("invalid session entropy")
	}
	digest := SessionDigest(first, "secret")
	if len(digest) != 64 || digest == first || digest == SessionDigest(first, "rotated") {
		t.Fatal("session digest does not protect the token or secret rotation")
	}
}
