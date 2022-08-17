package testing

import (
	"fmt"
	"strings"
)

func ServerEndpoint(path string) string {
	if !strings.HasPrefix(path, "/") {
		panic("path convention should start with /")
	}

	return fmt.Sprintf("http://localhost%s%s", ServerPort, path)
}
