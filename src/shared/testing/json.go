package testing

import (
	"encoding/json"
	"github.com/onsi/gomega"
	"github.com/veedubyou/chord-paper-be/src/server/api_error"
	"io"
)

func DecodeJSON[T any](jsonBody io.Reader) T {
	t := new(T)
	err := json.NewDecoder(jsonBody).Decode(t)
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())

	return *t
}

func DecodeJSONError(jsonBody io.Reader) api_error.JSONAPIError {
	return DecodeJSON[api_error.JSONAPIError](jsonBody)
}
