package patchtree

import (
	jd "github.com/josephburnett/jd/lib"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestBasic(t *testing.T) {
	blob, err := JSON("test/a/b")

	assert.For(t).ThatActual(err).IsNil()

	compareBlobTo(t, blob, "expected_a_b.json")
}

func compareBlobTo(t *testing.T, blob []byte, fileName string) {

	node, err := jd.ReadJsonString(string(blob))

	if err != nil {
		t.Error("Got error reading composed blob: " + err.Error())
		return
	}

	fileName = "./test/" + fileName

	expected, err := jd.ReadJsonFile(fileName)

	if err != nil {
		t.Error("Got error loading expected: " + err.Error())
		return
	}

	if !node.Equals(expected) {
		diff := node.Diff(expected)
		t.Error("Nodes don't match: " + diff.Render())
		return
	}
}
