package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/hyraxhomie/gator/internal/config"
	"github.com/hyraxhomie/gator/internal/database"
)

type State struct{
	db *database.Queries
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
	user, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("User does not exist.\n%w",err)
	} 
	err = s.config.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("Unable to set user config.\n%w",err)
	} else {
		fmt.Println("User has been set.")
	}
	return nil
}

func handlerRegister(s *State, cmd Command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Not Enough Args. A username is required.")
	}
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{ID: uuid.New(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(), Name: cmd.args[0]})
	if err != nil{
		fmt.Println(fmt.Errorf("An error occurred.\n%w", err))
		os.Exit(1)
	}
	s.config.SetUser(user.Name)
	fmt.Printf("A new (%s) user was registered.\n", user.Name)
	fmt.Println(user)
	return nil
}

func handlerReset(s *State, _ Command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Error resetting users table.\n%w",err)
	} else {
		fmt.Println("User table reset.")
	}
	return nil
}

func handlerUsers(s *State, _ Command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Error getting users.\n%w",err)
	} 

	for _, v := range users {
		fmt.Printf("* %s", v.Name)
		if s.config.CurrentUserName == v.Name {
			fmt.Print(" (current)")
		}
		fmt.Println()
	}
	return nil
}