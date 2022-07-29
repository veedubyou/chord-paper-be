package authtest

import (
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/auth"
	. "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/testing"
	"net/http"
	"net/http/httptest"
)

// to use this shared test, all tests must set the Endpoint in the BeforeEach
// and JSONBody optionally
var (
	Endpoint func(c echo.Context) error
	JSONBody interface{}
)

func ItRejectsUnauthorizedRequests(method string, path string) {
	Describe("Shared user authorization tests", func() {
		var (
			response       *httptest.ResponseRecorder
			requestFactory RequestFactory
		)

		BeforeEach(func() {
			requestFactory = RequestFactory{
				Method:  method,
				Path:    path,
				JSONObj: JSONBody,
			}
		})

		BeforeEach(func() {
			Expect(Endpoint).NotTo(BeNil())
		})

		AfterEach(func() {
			Endpoint = nil
			JSONBody = nil
		})

		JustBeforeEach(func() {
			request := requestFactory.Make()
			response = httptest.NewRecorder()
			c := PrepareEchoContext(request, response)

			Expect(Endpoint).NotTo(BeNil())
			err := Endpoint(c)
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("With no auth header", func() {
			It("fails with the right error code", func() {
				resErr := DecodeJSONError(response)
				Expect(resErr.Code).To(BeEquivalentTo(auth.BadAuthorizationHeaderCode))
			})

			It("fails with the right status code", func() {
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Describe("With a malformed token", func() {
			BeforeEach(func() {
				token := TokenForUserID(PrimaryUser.ID)
				requestFactory.Mods.Add(WithAuthHeader(token))
			})

			It("fails with the right error code", func() {
				resErr := DecodeJSONError(response)
				Expect(resErr.Code).To(BeEquivalentTo(auth.BadAuthorizationHeaderCode))
			})

			It("fails with the right status code", func() {
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Describe("With a Google unauthorized token", func() {
			BeforeEach(func() {
				requestFactory.Mods.Add(WithUserCred(GoogleUnauthorizedUser))
			})

			It("fails with the right error code", func() {
				resErr := DecodeJSONError(response)
				Expect(resErr.Code).To(BeEquivalentTo(auth.NotGoogleAuthorizedCode))
			})

			It("fails with the right status code", func() {
				Expect(response.Code).To(Equal(http.StatusUnauthorized))
			})
		})

		Describe("For a user that's not in the DB", func() {
			BeforeEach(func() {
				requestFactory.Mods.Add(WithUserCred(NoAccountUser))
			})

			It("fails with the right error code", func() {
				resErr := DecodeJSONError(response)
				Expect(resErr.Code).To(BeEquivalentTo(auth.NoAccountCode))
			})

			It("fails with the right status code", func() {
				Expect(response.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})
}
