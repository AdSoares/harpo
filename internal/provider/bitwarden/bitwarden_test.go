package bitwarden

import "testing"

func TestPickItem(t *testing.T) {
	t.Run("no match", func(t *testing.T) {
		if _, err := pickItem(nil, "gitlab"); err == nil {
			t.Fatal("expected error for zero matches")
		}
	})
	t.Run("single match", func(t *testing.T) {
		got, err := pickItem([]bwItem{{ID: "id-1", Name: "gitlab"}}, "gitlab")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != "id-1" {
			t.Fatalf("id = %q, want id-1", got.ID)
		}
	})
	t.Run("ambiguous without exact name", func(t *testing.T) {
		items := []bwItem{{ID: "a", Name: "gitlab dev"}, {ID: "b", Name: "gitlab prod"}}
		if _, err := pickItem(items, "gitlab"); err == nil {
			t.Fatal("expected ambiguity error")
		}
	})
	t.Run("ambiguous resolved by unique exact name", func(t *testing.T) {
		items := []bwItem{{ID: "a", Name: "gitlab"}, {ID: "b", Name: "gitlab prod"}}
		got, err := pickItem(items, "gitlab")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != "a" {
			t.Fatalf("id = %q, want a (exact name match)", got.ID)
		}
	})
}

func TestLooksLikeUUID(t *testing.T) {
	if !looksLikeUUID("3fa85f64-5717-4562-b3fc-2c963f66afa6") {
		t.Fatal("expected a valid UUID to be recognized")
	}
	if looksLikeUUID("gitlab.com | ad | PAT") {
		t.Fatal("a search string must not be treated as a UUID")
	}
}
