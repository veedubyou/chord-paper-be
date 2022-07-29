package testlib

import (
	. "github.com/onsi/gomega"
	"os"
)

func SetTestEnv() {
	err := os.Setenv("ENVIRONMENT", "test")
	Expect(err).NotTo(HaveOccurred())
}
