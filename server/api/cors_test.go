package api

import "testing"

func TestCORSOriginsNormalizesDocumentedCommaDelimitedFormat(t *testing.T) {
	for _, test := range []struct {
		name string
		in   string
		want string
	}{
		{"single", "http://localhost:8080", "http://localhost:8080"},
		{"without spaces", "http://localhost:8080,http://127.0.0.1:8080", "http://localhost:8080, http://127.0.0.1:8080"},
		{"mixed whitespace", " http://localhost:8080,  http://127.0.0.1:8080,", "http://localhost:8080, http://127.0.0.1:8080"},
		{"wildcard", "*", "*"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := corsOrigins(test.in); got != test.want {
				t.Fatalf("corsOrigins(%q) = %q, want %q", test.in, got, test.want)
			}
		})
	}
}
