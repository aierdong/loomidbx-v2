package main

import "testing"

func TestMeetsMinimumVersion(t *testing.T) {
	cases := []struct {
		name    string
		version string
		minimum string
		want    bool
	}{
		{name: "exact go version", version: "go version go1.25.0 windows/amd64", minimum: "1.25", want: true},
		{name: "newer go version", version: "go version go1.26.2 windows/amd64", minimum: "1.25", want: true},
		{name: "older go version", version: "go version go1.24.9 windows/amd64", minimum: "1.25", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := meetsMinimumVersion(tc.version, tc.minimum); got != tc.want {
				t.Fatalf("meetsMinimumVersion(%q, %q) = %v, want %v", tc.version, tc.minimum, got, tc.want)
			}
		})
	}
}
