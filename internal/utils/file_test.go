package utils

import (
	"os"
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
	f, err := os.Create("/tmp/dat2")
	if err != nil {
		t.Fail()
	}
	defer f.Close()

	exist := FileExists("/tmp/dat2")

	if !exist {
		t.Log("File should have existed")
		t.Fail()
	}

}
