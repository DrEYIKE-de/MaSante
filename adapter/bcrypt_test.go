package adapter

import "testing"

func TestBcryptHasher_RoundTrip(t *testing.T) {
	h := BcryptHasher{}
	password := "monmotdepasse"

	hash, err := h.Hash(password)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	if hash == password {
		t.Fatal("hash must differ from plaintext")
	}

	if err := h.Verify(hash, password); err != nil {
		t.Fatalf("Verify correct password: %v", err)
	}

	if err := h.Verify(hash, "wrong"); err == nil {
		t.Fatal("Verify wrong password: expected error")
	}
}
