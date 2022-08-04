package testlib

import . "github.com/onsi/gomega"

// compare most of the fields, except last saved at
func ExpectJSONEqualExceptLastSavedAt(a map[string]interface{}, b map[string]interface{}) {
	a["lastSavedAt"] = nil
	b["lastSavedAt"] = nil
	Expect(a).To(Equal(b))
}
