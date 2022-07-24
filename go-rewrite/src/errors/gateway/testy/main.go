package main

import (
	"fmt"
	"go/constant"
	"go/types"
	"golang.org/x/tools/go/packages"
)

func main() {
	load, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}, ".../chord-paper-be/...")
	if err != nil {
		panic(err)
	}

	for _, l := range load {
		for _, v := range l.TypesInfo.Defs {
			c, ok := v.(*types.Const)
			if !ok {
				continue
			}

			a, ok := c.Type().(*types.Named)
			if !ok {
				continue
			}
			b := a.Obj().Name()

			if b != "ErrorCode" {
				continue
			}

			e := a.Obj().Pkg().Name()

			if e != "api" {
				continue
			}

			if c.Val().Kind() != constant.String {
				continue
			}

			fmt.Println(c.Val())
		}

	}
}
