package keeper

import "testing"

func TestLooksLikeKeeperUID(t *testing.T) {
	if !looksLikeKeeperUID("rvwIBG_ban2VTH64OsnzLn") {
		t.Fatal("expected a 22-char base64url Keeper UID to be recognized")
	}
	if looksLikeKeeperUID("office/Zoom") {
		t.Fatal("a record path must not be treated as a UID")
	}
	if looksLikeKeeperUID("short") {
		t.Fatal("a short string must not be treated as a UID")
	}
}

func TestExtractField(t *testing.T) {
	rec := []byte(`{
	  "record_uid": "rvwIBG_ban2VTH64OsnzLn",
	  "title": "GitLab PAT",
	  "type": "login",
	  "fields": [
	    {"type": "login", "value": ["ad@example.com"]},
	    {"type": "password", "value": ["s3cret-token"]}
	  ],
	  "custom": [
	    {"type": "text", "label": "API Token", "value": ["abc123"]}
	  ]
	}`)

	t.Run("standard field by type", func(t *testing.T) {
		v, err := extractField(rec, "password")
		if err != nil {
			t.Fatal(err)
		}
		if v != "s3cret-token" {
			t.Fatalf("got %q, want s3cret-token", v)
		}
	})

	t.Run("custom field by label (case-insensitive)", func(t *testing.T) {
		v, err := extractField(rec, "api token")
		if err != nil {
			t.Fatal(err)
		}
		if v != "abc123" {
			t.Fatalf("got %q, want abc123", v)
		}
	})

	t.Run("missing field", func(t *testing.T) {
		if _, err := extractField(rec, "totp"); err == nil {
			t.Fatal("expected error for a field not present")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		if _, err := extractField([]byte("not json"), "password"); err == nil {
			t.Fatal("expected error for malformed json")
		}
	})
}
