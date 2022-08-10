package testlib

import (
	. "github.com/onsi/gomega"
	"os"
)

func SetTestEnv() {
	err := os.Setenv("ENVIRONMENT", "test")
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
}
