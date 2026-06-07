package bootstrap

import "testing"

func TestServiceStatusIsDeterministicAndReady(t *testing.T) {
	service := NewService("LoomiDBX", "0.1.0")

	status := service.Status()

	if status.Name != "LoomiDBX" {
		t.Fatalf("Name = %q, want LoomiDBX", status.Name)
	}
	if status.Version != "0.1.0" {
		t.Fatalf("Version = %q, want 0.1.0", status.Version)
	}
	if status.Runtime != "go" {
		t.Fatalf("Runtime = %q, want go", status.Runtime)
	}
	if !status.Ready {
		t.Fatal("Ready = false, want true")
	}
	if status.Message == "" {
		t.Fatal("Message should be developer-readable")
	}
}
