package ksm

import "testing"

func TestLooksLikeKeeperUID(t *testing.T) {
	if !looksLikeKeeperUID("SNzjw8tM1HsXEzXERCJrNQ") {
		t.Fatal("expected a 22-char base64url Keeper UID to be recognized")
	}
	if looksLikeKeeperUID("Stripe API Key") {
		t.Fatal("a record title must not be treated as a UID")
	}
	if looksLikeKeeperUID("short") {
		t.Fatal("a short string must not be treated as a UID")
	}
}

func TestPickRecord(t *testing.T) {
	records := []ksmRecord{
		{UID: "AAAAAAAAAAAAAAAAAAAAAA", Title: "Stripe API Key"},
		{UID: "BBBBBBBBBBBBBBBBBBBBBB", Title: "GitLab PAT"},
		{UID: "CCCCCCCCCCCCCCCCCCCCCC", Title: "GitLab PAT"},
	}

	t.Run("unique title match", func(t *testing.T) {
		r, err := pickRecord(records, "Stripe API Key")
		if err != nil {
			t.Fatal(err)
		}
		if r.UID != "AAAAAAAAAAAAAAAAAAAAAA" {
			t.Fatalf("got %q, want the Stripe record", r.UID)
		}
	})

	t.Run("case-insensitive title match", func(t *testing.T) {
		if _, err := pickRecord(records, "stripe api key"); err != nil {
			t.Fatalf("expected case-insensitive match: %v", err)
		}
	})

	t.Run("no match", func(t *testing.T) {
		if _, err := pickRecord(records, "Nonexistent"); err == nil {
			t.Fatal("expected error for zero matches")
		}
	})

	t.Run("ambiguous title", func(t *testing.T) {
		if _, err := pickRecord(records, "GitLab PAT"); err == nil {
			t.Fatal("expected ambiguity error for duplicate titles")
		}
	})
}
