package web

import "testing"

func TestLicenseVerifyDoesNotConsume(t *testing.T) {
	store := newAccountStore()
	key, err := store.addLicense(LicenseKey{Key: "TEST-KEY", Active: true, MaxUses: 2})
	if err != nil {
		t.Fatalf("expected license to be created: %v", err)
	}
	if key.MaxUses != 2 {
		t.Fatalf("expected max uses 2, got %d", key.MaxUses)
	}

	verified, err := store.verifyLicense("user-1", "TEST-KEY")
	if err != nil {
		t.Fatalf("expected verify to succeed: %v", err)
	}
	if verified.Uses != 0 || verified.Remaining != 2 {
		t.Fatalf("verify should not consume license, got uses=%d remaining=%d", verified.Uses, verified.Remaining)
	}

	used, err := store.consumeLicense("user-1", "TEST-KEY")
	if err != nil {
		t.Fatalf("expected consume to succeed: %v", err)
	}
	if used.Uses != 1 || used.Remaining != 1 {
		t.Fatalf("consume should decrement remaining, got uses=%d remaining=%d", used.Uses, used.Remaining)
	}
}

func TestRandomReadableKeyUsesCustomPrefix(t *testing.T) {
	key := randomReadableKey("vip")
	if len(key) <= 4 || key[:4] != "VIP-" {
		t.Fatalf("expected generated key to use VIP prefix, got %q", key)
	}
}
