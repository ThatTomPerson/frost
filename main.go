package main

import (
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"ttp.sh/frost/internal/manager"
	"ttp.sh/frost/internal/manager/composer"
)

func main() {
	log.SetHandler(cli.Default)
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}
	m := manager.New(path)
	m.Add(composer.New(path))

	m.Run()
}
