package codegen

import (
	"github.com/workfit/tester/assert"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"testing"
)

func TestOutput(t *testing.T) {

	readerOutput, _, err := ProcessReaders("examplepkg/")

	assert.For(t).ThatActual(err).IsNil()

	expectedBytes, err := ioutil.ReadFile("test/expected_auto_reader.txt")

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(strings.TrimSpace(readerOutput)).Equals(strings.TrimSpace(string(expectedBytes))).ThenDiffOnFail()

}

func TestOutputTest(t *testing.T) {

	_, readerTestOutput, err := ProcessReaders("examplepkg/")

	assert.For(t).ThatActual(err).IsNil()

	expectedBytes, err := ioutil.ReadFile("test/expected_auto_reader_test.txt")

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(strings.TrimSpace(readerTestOutput)).Equals(strings.TrimSpace(string(expectedBytes))).ThenDiffOnFail()

}

func TestEnumOutput(t *testing.T) {

	enumOutput, err := ProcessEnums("examplepkg/")
	assert.For(t).ThatActual(err).IsNil()

	expectedBytes, err := ioutil.ReadFile("test/expected_auto_enum.txt")

	assert.For(t).ThatActual(err).IsNil()

	trimmedOut := strings.TrimSpace(enumOutput)
	trimmedExpected := strings.TrimSpace(string(expectedBytes))

	assert.For(t).ThatActual(trimmedOut).Equals(trimmedExpected).ThenDiffOnFail()

}

func TestBuild(t *testing.T) {

	log.Println("WARNING: running this command generates auto output for examplepkg/")

	//Get default options

	output, testOutput, err := ProcessReaders("examplepkg/")

	assert.For(t).ThatActual(err).IsNil()

	enumOutput, err := ProcessEnums("examplepkg/")

	assert.For(t).ThatActual(err).IsNil()

	err = ioutil.WriteFile("examplepkg/auto_reader.go", []byte(output), 0644)
	assert.For(t).ThatActual(err).IsNil()

	err = ioutil.WriteFile("examplepkg/auto_reader_test.go", []byte(testOutput), 0644)
	assert.For(t).ThatActual(err).IsNil()

	err = ioutil.WriteFile("examplepkg/auto_enum.go", []byte(enumOutput), 0644)
	assert.For(t).ThatActual(err).IsNil()

	cmd := exec.Command("go", "test")
	cmd.Dir = "./examplepkg/"

	err = cmd.Run()

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
