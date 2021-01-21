package utils

import (
	"io/ioutil"
	"testing"
)

func TestFileDoesntExist(t *testing.T) {
	exist := FileExists("/tmp/somepaththatdoesnotexist.jpg")

	if exist {
		t.Log("File should not have existed")
		t.Fail()
	}
}

func TestFileExists(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "DAT2")
	if err != nil {
		t.Fail()
	}
	defer f.Close()

	exist := FileExists(f.Name())

	if !exist {
		t.Log("File should have existed")
		t.Fail()
	}
}
