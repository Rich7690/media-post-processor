package helper

import (
	"github.com/gin-gonic/gin"
	"os"
	"testing"
)

// This function is used for setup before executing the test functions
func TestMain(m *testing.M) {
	//Set Gin to Test Mode
	//fmt.Println("Setting gin to test mode")
	gin.SetMode(gin.TestMode)

	// Run the other tests
	os.Exit(m.Run())
}
