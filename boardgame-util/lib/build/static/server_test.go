package static

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestViteServerCommandUsesRequestedPorts(t *testing.T) {
	cmd := viteServerCommand("build-root", "19080", "19888")

	wantDir := filepath.Join("build-root", staticSubFolder)
	if cmd.Dir != wantDir {
		t.Fatalf("Dir = %q, want %q", cmd.Dir, wantDir)
	}
	wantArgs := []string{"npx", "vite", "--port", "19080", "--host", "127.0.0.1", "--strictPort"}
	if !reflect.DeepEqual(cmd.Args, wantArgs) {
		t.Fatalf("Args = %#v, want %#v", cmd.Args, wantArgs)
	}
	if !containsString(cmd.Env, "BOARDGAME_STATIC_PORT=19080") {
		t.Fatal("command environment missing BOARDGAME_STATIC_PORT")
	}
	if !containsString(cmd.Env, "BOARDGAME_API_PORT=19888") {
		t.Fatal("command environment missing BOARDGAME_API_PORT")
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
