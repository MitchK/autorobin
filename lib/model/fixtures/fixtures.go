package fixtures

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

// ExamplePVCSVStr ExamplePVCSVStr
func ExamplePVCSVStr(t *testing.T) string {
	path := filepath.Join("fixtures", "test.csv")
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(buf)
}
