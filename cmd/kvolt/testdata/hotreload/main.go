package main

import (
	"github.com/go-kvolt/kvolt"
	"github.com/go-kvolt/kvolt/context"
)

const version = "v1"

func main() {
	app := kvolt.New()
	app.GET("/version", func(c *context.Context) error {
		return c.String(200, version)
	})
	app.Run(":19876")
}
