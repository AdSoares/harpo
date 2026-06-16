package onepassword

import "testing"

func TestBuildReference(t *testing.T) {
	cases := []struct {
		name  string
		ref   string
		field string
		want  string
	}{
		{"vault/item path", "Private/GitLab", "password", "op://Private/GitLab/password"},
		{"already has scheme", "op://Private/GitLab", "password", "op://Private/GitLab/password"},
		{"trailing slash", "Private/GitLab/", "password", "op://Private/GitLab/password"},
		{"with section", "Private/AWS/section", "access_key", "op://Private/AWS/section/access_key"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := buildReference(c.ref, c.field)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Fatalf("buildReference(%q,%q) = %q, want %q", c.ref, c.field, got, c.want)
			}
		})
	}
}

func TestBuildReferenceErrors(t *testing.T) {
	if _, err := buildReference("", "password"); err == nil {
		t.Fatal("expected error for empty ref")
	}
	if _, err := buildReference("Private/GitLab", ""); err == nil {
		t.Fatal("expected error for empty field")
	}
}
