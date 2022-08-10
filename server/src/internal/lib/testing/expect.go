package testlib

import . "github.com/onsi/gomega"

func ExpectSuccess[T any](t T, err error) T {
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return t
}

func ExpectType[T any](thing interface{}) T {
	ExpectWithOffset(1, thing).NotTo(BeNil())
	realThing, ok := thing.(T)
	ExpectWithOffset(1, ok).To(BeTrue())
	return realThing
}
