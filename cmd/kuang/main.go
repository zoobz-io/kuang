// Package main is the entrypoint for a Kuang server with all the bells & whistles.
package main

import (
	"log"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/modules/github"
	"github.com/zoobzio/kuang/modules/matrix"
)

func main() {
	err := core.Run(github.Module(), matrix.Module())
	if err != nil {
		log.Fatal(err)
	}
}
