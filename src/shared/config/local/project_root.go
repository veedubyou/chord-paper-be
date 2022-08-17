package local

import (
	"path"
	"runtime"
	"strings"
)

func ProjectRoot() string {
	_, filePath, _, ok := runtime.Caller(0)

	if !ok {
		panic("Failed to call runtime.Caller")
	}

	if !strings.HasSuffix(filePath, "/src/shared/config/local/project_root.go") {
		panic("")
	}

	// huge assumption:
	// this file is currently situated in
	// projectRoot/src/shared/values/local/project_root.go
	for i := 0; i < 5; i++ {
		filePath = path.Dir(filePath)
	}

	return filePath
}
