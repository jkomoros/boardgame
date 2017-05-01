package main

import (
	"bytes"
	"github.com/workfit/tester/assert"
	"io/ioutil"
	"log"
	"os/exec"
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

func TestNonReflectOutput(t *testing.T) {
	options := &appOptions{
		PrintToConsole:   true,
		PackageDirectory: "examplepkg/",
		UseReflection:    false,
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	process(options, out, errOut)

	expectedBytes, err := ioutil.ReadFile("test/non_reflect_output.txt")

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(out.String()).Equals(string(expectedBytes)).ThenDiffOnFail()

}

func TestBuild(t *testing.T) {

	log.Println("WARNING: running this command builds and `go install`s autoreader")

	//Make sure a recent version of us is built
	cmd := exec.Command("go", "install")

	err := cmd.Run()

	if !assert.For(t).ThatActual(err).IsNil().Passed() {
		log.Println(err)
	}

	cmd = exec.Command("go", "generate")
	cmd.Dir = "examplepkg/"

	err = cmd.Run()

	if !assert.For(t).ThatActual(err).IsNil().Passed() {
		log.Println(err)
	}

	cmd = exec.Command("go", "test")
	cmd.Dir = "./examplepkg/"

	err = cmd.Run()

	if !assert.For(t).ThatActual(err).IsNil().Passed() {
		log.Println(err)
	}

}
