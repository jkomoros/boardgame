package main

import (
	"testing"

	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
)

func TestClientConfigForServeUsesProxyForPortOverrides(t *testing.T) {
	c := config.DefaultStarterConfig("")
	original := c.Client(false)
	if original == nil || original.DevHost == "" {
		t.Fatal("starter config should have a non-empty development host")
	}

	proxied := clientConfigForServe(c, false)
	if proxied == nil {
		t.Fatal("proxied client config was nil")
	}
	if proxied.DevHost != "" {
		t.Fatalf("proxied DevHost = %q, want same-origin empty host", proxied.DevHost)
	}
	if proxied.Host != "" {
		t.Fatalf("proxied Host = %q, want same-origin empty host", proxied.Host)
	}
	if original.DevHost == "" || c.Client(false).DevHost == "" {
		t.Fatal("clientConfigForServe mutated the source config")
	}
}

func TestClientConfigForServeUsesProxyInProduction(t *testing.T) {
	c := config.SampleStarterConfig("")

	proxiedProd := clientConfigForServe(c, true)
	if proxiedProd == nil || proxiedProd.Host != "" || proxiedProd.DevHost != "" {
		t.Fatal("production config should use the same-origin proxy with overrides")
	}
}

func TestLocalServeAllowedOriginsUsesBothLoopbackHosts(t *testing.T) {
	got := localServeAllowedOrigins("49152")
	want := "http://localhost:49152,http://127.0.0.1:49152"
	if got != want {
		t.Fatalf("localServeAllowedOrigins() = %q, want %q", got, want)
	}
}
