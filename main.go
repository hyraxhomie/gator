package main

import (
	"fmt"
	"os"

	"github.com/hyraxhomie/gator/internal/config"
)

func main(){
	cfg := config.Read()
	state := State{config: &cfg}
	commands := Commands{commands: make(map[string]func(*State, Command) error)}
	commands.register("login", handlerLogin)
	args := os.Args
	if len(args) < 2 {
		fmt.Println(fmt.Errorf("Not enough arguments were provided"))
		os.Exit(1)
	}
	command := Command{name: args[1], args: append([]string{}, args[2:]...)}
	err := commands.run(&state, command)
	if err != nil {
		fmt.Println(fmt.Errorf("An error occurred: %w", err))
		os.Exit(1)
	}
}
