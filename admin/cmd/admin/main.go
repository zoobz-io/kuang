// Package main is the entry point for the kuang admin API.
package main

import (
	"log"

	_ "github.com/mattn/go-sqlite3"

	"github.com/zoobz-io/kuang/admin"
)

func main() {
	if err := admin.Run(); err != nil {
		log.Fatal(err)
	}
}
