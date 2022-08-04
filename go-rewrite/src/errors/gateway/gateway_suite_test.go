package gateway_test

import (
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/api"
	"go/constant"
	"go/types"
	"golang.org/x/tools/go/packages"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/testing"
)

var allErrorCodes []api.ErrorCode

func TestGateway(t *testing.T) {
	RegisterFailHandler(Fail)

	allErrorCodes = loadAllErrorCodes()
	Expect(allErrorCodes).NotTo(BeEmpty())

	RunSpecs(t, "Gateway Suite")
}

func loadAllErrorCodes() []api.ErrorCode {
	errorCodes := []api.ErrorCode{}

	packages := ExpectSuccess(packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo,
	}, "github.com/veedubyou/chord-paper-be/..."))

	for _, pack := range packages {
		for _, def := range pack.TypesInfo.Defs {
			errorCode, ok := parseDefinitionForErrorCode(def)
			if ok {
				errorCodes = append(errorCodes, errorCode)
			}
		}
	}

	return errorCodes
}

func parseDefinitionForErrorCode(def types.Object) (api.ErrorCode, bool) {
	notFound := func() (api.ErrorCode, bool) {
		return "", false
	}

	constVal, ok := def.(*types.Const)
	if !ok {
		return notFound()
	}

	namedVal, ok := constVal.Type().(*types.Named)
	if !ok {
		return notFound()
	}

	typeName := namedVal.Obj().Name()

	if typeName != "ErrorCode" {
		return notFound()
	}

	typePkg := namedVal.Obj().Pkg().Name()

	if typePkg != "api" {
		return notFound()
	}

	Expect(constVal.Val().Kind()).To(Equal(constant.String))

	stringVal := constant.StringVal(constVal.Val())
	Expect(stringVal).NotTo(BeZero())

	return api.ErrorCode(stringVal), true
}
