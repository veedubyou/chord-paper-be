package testlib

import . "github.com/onsi/gomega"

func ExpectSuccess[T any](t T, err error) T {
	Expect(err).NotTo(HaveOccurred())
	return t
}
