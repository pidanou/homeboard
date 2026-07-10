package service

import (
	"testing"
	"time"
)

func TestOIDCHandoffStoreConsumeOnce(t *testing.T) {
	store := NewOIDCHandoffStore()
	code, err := store.Put("jwt-value")
	if err != nil {
		t.Fatalf("put: %v", err)
	}

	got, err := store.Consume(code)
	if err != nil {
		t.Fatalf("first consume: %v", err)
	}
	if got != "jwt-value" {
		t.Errorf("want jwt-value, got %q", got)
	}

	if _, err := store.Consume(code); err != ErrHandoffCodeInvalid {
		t.Errorf("want ErrHandoffCodeInvalid on second consume, got %v", err)
	}
}

func TestOIDCHandoffStoreExpiry(t *testing.T) {
	store := NewOIDCHandoffStore()
	code, err := store.Put("jwt-value")
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	store.entries[code] = handoffEntry{jwt: "jwt-value", expires: time.Now().Add(-time.Second)}

	if _, err := store.Consume(code); err != ErrHandoffCodeInvalid {
		t.Errorf("want ErrHandoffCodeInvalid for expired code, got %v", err)
	}
}

func TestOIDCHandoffStoreUnknownCode(t *testing.T) {
	store := NewOIDCHandoffStore()
	if _, err := store.Consume("does-not-exist"); err != ErrHandoffCodeInvalid {
		t.Errorf("want ErrHandoffCodeInvalid, got %v", err)
	}
}
