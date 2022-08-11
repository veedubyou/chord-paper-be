package main

import (
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application"
)

func main() {
	app := application.NewApp()
	if err := app.Start(); err != nil {
		panic(err)
	}
}
