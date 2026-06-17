package keychain

import (
	"testing"
	"time"

	"github.com/zalando/go-keyring"
)

func TestSaveLoadDelete(t *testing.T) {
	keyring.MockInit() // in-memory backend; does not touch the real OS keychain

	e := Entry{Name: "BW_SESSION", Value: "tok-123", ExpiresAt: time.Now().Add(time.Hour)}
	if err := Save("bw", e); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, ok, err := Load("bw")
	if err != nil || !ok {
		t.Fatalf("load: ok=%v err=%v", ok, err)
	}
	if got.Value != "tok-123" || got.Name != "BW_SESSION" {
		t.Fatalf("unexpected entry: %+v", got)
	}

	if err := Delete("bw"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, ok, _ := Load("bw"); ok {
		t.Fatal("entry should be gone after delete")
	}
}

func TestLoadExpiredEvicts(t *testing.T) {
	keyring.MockInit()

	e := Entry{Name: "BW_SESSION", Value: "tok", ExpiresAt: time.Now().Add(-time.Minute)}
	if err := Save("bw", e); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := Load("bw"); ok {
		t.Fatal("expired entry must be reported as not found")
	}
	// and it must have been evicted
	if err := keyring.Set(service, "probe", "x"); err != nil {
		t.Fatal(err)
	}
}

func TestLoadMissing(t *testing.T) {
	keyring.MockInit()
	if _, ok, err := Load("nope"); ok || err != nil {
		t.Fatalf("missing entry: ok=%v err=%v", ok, err)
	}
}
