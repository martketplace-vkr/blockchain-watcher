package trongrid

import "testing"

func TestNormalizeEventAddress(t *testing.T) {
	t.Parallel()

	got, err := normalizeEventAddress("0x64d21d4f0ef0c230a4b7627d4e0e1ad1b8eb8fc7")
	if err != nil {
		t.Fatal(err)
	}

	const want = "TKAJHFbvPQhbTbYBiPxuKbdfyUKAcY94dz"
	if got != want {
		t.Fatalf("unexpected address: got %q, want %q", got, want)
	}
}

func TestNormalizeEventAddressKeepsBase58(t *testing.T) {
	t.Parallel()

	const address = "TKAJHFbvPQhbTbYBiPxuKbdfyUKAcY94dz"
	got, err := normalizeEventAddress(address)
	if err != nil {
		t.Fatal(err)
	}
	if got != address {
		t.Fatalf("unexpected address: got %q, want %q", got, address)
	}
}
