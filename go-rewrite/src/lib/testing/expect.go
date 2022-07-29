package testlib

import . "github.com/onsi/gomega"

func ExpectSuccess[T any](t T, err error) T {
	Expect(err).NotTo(HaveOccurred())
	return t
}

func ExpectType[T any](thing interface{}) T {
	Expect(thing).NotTo(BeNil())
	realThing, ok := thing.(T)
	Expect(ok).To(BeTrue())
	return realThing
}
