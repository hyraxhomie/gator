package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/hyraxhomie/gator/internal/config"
	"github.com/hyraxhomie/gator/internal/database"
	_ "github.com/lib/pq"
)

func main(){
	// get config, initialize state and db
	cfg := config.Read()
	state := State{config: &cfg}
	db, err := sql.Open("postgres", state.config.DbUrl)	
	handleErr(err)

	state.db = database.New(db)

	// commands
	commands := Commands{commands: make(map[string]func(*State, Command) error)}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", handlerAddFeed)
	commands.register("feeds", handlerFeeds)


	// get and validate args
	args := os.Args
	if len(args) < 2 {
		fmt.Println(fmt.Errorf("Not enough arguments were provided"))
		os.Exit(1)
	}

	//do command
	command := Command{name: args[1], args: append([]string{}, args[2:]...)}
	err = commands.run(&state, command)
	handleErr(err)
}

func handleErr(err error, msg ...string){
	if len(msg) > 0 {
		for _, v := range msg {
			fmt.Println(v)
		}
	}
	if err != nil {
		fmt.Println(fmt.Errorf("An error occurred: %w", err))
		os.Exit(1)
	}
}