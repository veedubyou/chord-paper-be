package gateway_test

import (
	"github.com/cockroachdb/errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/veedubyou/chord-paper-be/server/src/internal/errors/api"
	"github.com/veedubyou/chord-paper-be/server/src/internal/errors/gateway"
	"github.com/veedubyou/chord-paper-be/server/src/internal/errors/gateway/gatewayfakes"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 github.com/labstack/echo/v4.Context

var _ = Describe("Errors", func() {
	Describe("HTTP status code handling for ErrorCodes", func() {
		var (
			fakeEchoContext *gatewayfakes.FakeContext
		)

		BeforeEach(func() {
			fakeEchoContext = &gatewayfakes.FakeContext{}
			fakeEchoContext.JSONReturns(nil)
		})

		for _, errorCode := range allErrorCodes {
			errorCode := errorCode
			It("processes ErrorCode "+string(errorCode), func() {
				apiError := &api.Error{
					ErrorCode:     errorCode,
					UserMessage:   "Something failed",
					InternalError: errors.New("Our DB blew up"),
				}

				runTest := func() {
					gateway.ErrorResponse(fakeEchoContext, apiError)
				}
				Expect(runTest).NotTo(Panic())
			})
		}
	})
})
