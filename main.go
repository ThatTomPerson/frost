package main

import (
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"ttp.sh/frost/internal/manager"
	_ "ttp.sh/frost/internal/manager/composer"
	_ "ttp.sh/frost/internal/manager/yarn"
)

func main() {
	log.SetHandler(cli.Default)
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}
	manager.Run(path)
}
