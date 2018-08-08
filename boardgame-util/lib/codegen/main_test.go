package codegen

import (
	"bytes"
	"flag"
	"github.com/workfit/tester/assert"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"testing"
)

func TestOutput(t *testing.T) {
	options := &appOptions{
		PrintToConsole:   true,
		PackageDirectory: "examplepkg/",
		OutputReader:     true,
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	assert.For(t).ThatActual(errOut.String()).Equals("").ThenDiffOnFail()

	process(options, out, errOut)

	expectedBytes, err := ioutil.ReadFile("test/expected_auto_reader.txt")

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(strings.TrimSpace(out.String())).Equals(strings.TrimSpace(string(expectedBytes))).ThenDiffOnFail()

}

func TestOutputTest(t *testing.T) {
	options := &appOptions{
		PrintToConsole:   true,
		PackageDirectory: "examplepkg/",
		OutputReaderTest: true,
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	assert.For(t).ThatActual(errOut.String()).Equals("").ThenDiffOnFail()

	process(options, out, errOut)

	expectedBytes, err := ioutil.ReadFile("test/expected_auto_reader_test.txt")

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(strings.TrimSpace(out.String())).Equals(strings.TrimSpace(string(expectedBytes))).ThenDiffOnFail()

}

func TestEnumOutput(t *testing.T) {
	options := &appOptions{
		PrintToConsole:   true,
		PackageDirectory: "examplepkg/",
		OutputEnum:       true,
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	process(options, out, errOut)

	assert.For(t).ThatActual(errOut.String()).Equals("").ThenDiffOnFail()

	expectedBytes, err := ioutil.ReadFile("test/expected_auto_enum.txt")

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(strings.TrimSpace(out.String())).Equals(strings.TrimSpace(string(expectedBytes))).ThenDiffOnFail()

}

func TestBuild(t *testing.T) {

	log.Println("WARNING: running this command generates auto output for examplepkg/")

	//Get default options
	options := getOptions(flag.NewFlagSet("test", 0), []string{})

	options.PackageDirectory = "examplepkg/"

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	process(options, out, errOut)

	assert.For(t).ThatActual(errOut.String()).Equals("").ThenDiffOnFail()

	cmd := exec.Command("go", "test")
	cmd.Dir = "./examplepkg/"

	err := cmd.Run()

	if !assert.For(t).ThatActual(err).IsNil().Passed() {
		log.Println(err)
	}

}

func TestOverrideDisplayNames(t *testing.T) {

	tests := []struct {
		input       string
		hasOverride bool
		displayName string
	}{
		{
			`display:"bam"`,
			true,
			"bam",
		},
		{
			" display:\"foo\"\n",
			true,
			"foo",
		},
		{
			`display:"hello \"john\""`,
			true,
			`hello \"john\"`,
		},
		{
			"blarg\n",
			false,
			"",
		},
		{
			`display:""`,
			true,
			"",
		},
	}

	for i, test := range tests {
		hasOverride, displayName := overrideDisplayname(test.input)

		assert.For(t, i).ThatActual(hasOverride).Equals(test.hasOverride)
		assert.For(t, i).ThatActual(displayName).Equals(test.displayName)

	}

}
