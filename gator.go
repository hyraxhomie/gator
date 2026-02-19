package main

import (
	"fmt"

	"github.com/hyraxhomie/gator/internal/config"
)

type State struct{
	config *config.Config
}

type Commands struct{
	commands map[string]func(*State, Command) error
}

func (c *Commands) run(s *State, cmd Command) error{
	err := c.commands[cmd.name](s, cmd)
	return err
}

func (c *Commands) register(name string, f func(*State, Command) error) {
	c.commands[name] = f
}

type Command struct{
	name string
	args []string
}

func handlerLogin(s *State, cmd Command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Not Enough Args. A username is required.")
	}
	err := s.config.SetUser(cmd.args[0])
	if err != nil {
		return err
	} else {
		fmt.Println("User has been set.")
	}
	return nil
}