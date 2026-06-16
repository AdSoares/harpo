package vault

import (
	"strings"
	"testing"
)

func TestKvGetArgs(t *testing.T) {
	got := kvGetArgs("secret/myapp", "password")
	want := []string{"kv", "get", "-field=password", "secret/myapp"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("kvGetArgs = %v, want %v", got, want)
	}
}

func TestKvGetArgsFieldFlagIsSingleArg(t *testing.T) {
	// The field must be passed as a single "-field=NAME" argument so a field
	// name with spaces or special chars cannot be split into separate args.
	got := kvGetArgs("secret/x", "api key")
	if got[2] != "-field=api key" {
		t.Fatalf("expected -field=api key as one arg, got %q", got[2])
	}
}
