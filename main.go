package main

import (
	"fmt"
	"log"
	"os"

	"github.com/frankielb/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	state := config.State{
		CurrentConfig: &cfg,
	}
	commands := config.Commands{
		Handlers: make(map[string]func(*config.State, config.Command) error),
	}
	commands.Register("login", config.HandlerLogin)

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Error: not enough arguments")
		os.Exit(1)
	}
	cmdName := args[1]
	cmdArgs := args[2:]

	cmd := config.Command{
		Name: cmdName,
		Args: cmdArgs,
	}
	err = commands.Run(&state, cmd)
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}

}
