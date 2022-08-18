package testing

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/onsi/gomega"
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
	Target  string
	JSONObj interface{}
	Mods    RequestModifiers
}

func (r RequestFactory) make(reqMaker func(string, string, io.Reader) *http.Request) *http.Request {
	var body io.Reader

	if r.JSONObj != nil {
		buf := &bytes.Buffer{}
		err := json.NewEncoder(buf).Encode(r.JSONObj)
		gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())

		body = buf
	}

	request := reqMaker(r.Method, r.Target, body)

	isJSONBody := body != nil
	if isJSONBody {
		request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}

	for _, mod := range r.Mods {
		mod(request)
	}

	return request
}

func (r RequestFactory) MakeFake() *http.Request {
	return r.make(httptest.NewRequest)
}

func (r RequestFactory) Do() (*http.Response, error) {
	makeRealRequest := func(method string, target string, body io.Reader) *http.Request {
		return ExpectSuccess(http.NewRequest(method, target, body))
	}

	req := r.make(makeRealRequest)
	return http.DefaultClient.Do(req)
}
