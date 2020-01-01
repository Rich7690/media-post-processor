package constants

import "testing"

func TestConstants(t *testing.T) {
	local := IsLocal

	if local == true {
		t.Error("Local should not be true in tests")
	}
}
