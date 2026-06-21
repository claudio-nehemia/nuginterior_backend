package main

import (
	"log"

	"github.com/claudio-nehemia/interior_backend/internal/bootstrap"
)

func main() {
	app, err := bootstrap.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	if app.Logger != nil {
		defer func() { _ = app.Logger.Sync() }()
	}
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
