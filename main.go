package main

import (
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"ttp.sh/frost/internal/project"

	_ "ttp.sh/frost/internal/handler/composer"
)

func main() {
	log.SetHandler(cli.Default)

	root, _ := os.Getwd()
	p := project.New(root)

	p.Install()
}
