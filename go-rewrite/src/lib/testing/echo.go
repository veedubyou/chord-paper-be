package testlib

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/gomega"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/gateway"
	"io"
	"net/http"
	"net/http/httptest"
)

type RequestModifier func(r *http.Request)

type RequestModifiers []RequestModifier

func (r *RequestModifiers) Add(mods ...RequestModifier) {
	*r = append(*r, mods...)
}

func WithAuthHeader(header string) RequestModifier {
	return func(request *http.Request) {
		request.Header.Set("Authorization", header)
	}
}

func WithUserCred(user User) RequestModifier {
	return func(request *http.Request) {
		token := TokenForUserID(user.ID)
		request.Header.Set("Authorization", "Bearer "+token)
	}
}

type RequestFactory struct {
	Method  string
	Path    string
	JSONObj interface{}
	Mods    RequestModifiers
}

func (r RequestFactory) Make() *http.Request {
	var body io.Reader

	if r.JSONObj != nil {
		buf := &bytes.Buffer{}
		err := json.NewEncoder(buf).Encode(r.JSONObj)
		Expect(err).NotTo(HaveOccurred())

		body = buf
	}

	request := httptest.NewRequest(r.Method, r.Path, body)

	isJSONBody := body != nil
	if isJSONBody {
		request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}

	for _, mod := range r.Mods {
		mod(request)
	}

	return request
}

func PrepareEchoContext(request *http.Request, response http.ResponseWriter) echo.Context {
	e := echo.New()
	return e.NewContext(request, response)
}

func DecodeJSON[T any](response *httptest.ResponseRecorder) T {
	t := new(T)
	err := json.NewDecoder(response.Body).Decode(t)
	Expect(err).NotTo(HaveOccurred())

	return *t
}

func DecodeJSONError(response *httptest.ResponseRecorder) gateway.JSONAPIError {
	return DecodeJSON[gateway.JSONAPIError](response)
}
