package main

import (
	"bytes"
	"github.com/workfit/tester/assert"
	"io/ioutil"
	"testing"
)

func TestOutput(t *testing.T) {
	options := &appOptions{
		PrintToConsole:   true,
		PackageDirectory: "examplepkg/",
		UseReflection:    true,
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	process(options, out, errOut)

	expectedBytes, err := ioutil.ReadFile("test/output.txt")

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(out.String()).Equals(string(expectedBytes)).ThenDiffOnFail()

}
