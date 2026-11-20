package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := Default("sample")
	if err := Save(dir, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != want.Name || got.Version != want.Version || !got.Stacks.Go || !got.Stacks.Java || !got.Stacks.Python || !got.Stacks.Redis {
		t.Fatalf("unexpected config round trip: %#v", got)
	}
}
func TestLoadRejectsMissingName(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "devforge.yaml"), []byte("version: '1'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); err == nil {
		t.Fatal("expected missing name error")
	}
}
