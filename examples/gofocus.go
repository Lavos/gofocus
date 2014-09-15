package main

import (
	"github.com/Lavos/gofocus"
	"github.com/kelseyhightower/envconfig"
)

func main() {
	c := &gofocus.Configuration{}
	envconfig.Process("GOFOCUS", c)

	a := gofocus.NewApplication(c)
	a.Run()
}
