package main

import (
	"context"
	"database/sql"
	"fmt"
	"main/internal/config"
	"main/internal/database"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	dbURL := "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"

	config, err := config.Read()
	if err != nil {
		fmt.Printf("Error reading config file %s", err)
	}

	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)
	if err != nil {
		fmt.Printf("Error connecting to the database: %s", err)
	}

	appState := state{cfg: &config, db: dbQueries}
	cmds := commands{availibleCommands: map[string]func(*state, command) error{}}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerListUsers)
	cmds.register("agg", handlerAggregator)
	if len(os.Args) < 2 {
		err := fmt.Errorf("no arguments passed, exiting")
		fmt.Println(err)
		os.Exit(1)
	}
	cmd := command{name: os.Args[1], args: os.Args[2:]}
	err = cmds.run(&appState, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type state struct {
	cfg *config.Config
	db  *database.Queries
}

type command struct {
	name string
	args []string
}

type commands struct {
	availibleCommands map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	runCommand, ok := c.availibleCommands[cmd.name]
	if !ok {
		err := fmt.Errorf("%s command does not exist", cmd.name)
		return err
	}

	return runCommand(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.availibleCommands[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if cmd.name == "login" {
		if len(cmd.args) == 0 {
			err := fmt.Errorf("login command requires at least one argument")
			return err
		}
		user, err := s.db.GetUser(context.Background(), cmd.args[0])
		if err != nil {
			return err
		}
		err = s.cfg.SetUser(user.Name)
		fmt.Printf("The user has been set to: %s", user.Name)
	}
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if cmd.name == "register" {
		if len(cmd.args) == 0 {
			err := fmt.Errorf("register command requires at least 2 argument")
			return err
		}
		params := database.CreateUserParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.args[0]}
		user, err := s.db.CreateUser(context.Background(), params)
		if err != nil {
			return err
		}
		err = s.cfg.SetUser(user.Name)
		if err != nil {
			return err
		}
		fmt.Printf("The user has been set to: %s", user.Name)
	}
	return nil
}

func handlerListUsers(s *state, cmd command) error {
	if cmd.name == "users" {
		users, err := s.db.GetUsers(context.Background())
		if err != nil {
			return err
		}
		for _, value := range users {
			if value.Name == s.cfg.CurrentUserName {
				fmt.Printf("* %s (current)\n", value.Name)
			} else {
				fmt.Printf("* %s\n", value.Name)
			}
		}
	}
	return nil
}

func handlerAggregator(s *state, cmd command) error {
	url := "https://www.wagslane.dev/index.xml"
	if cmd.name == "agg" {
		fetchFeed(context.Background(), url)
	}
	return nil
}

func handlerReset(s *state, cmd command) error {
	if cmd.name == "reset" {
		err := s.db.DeleteUsers(context.Background())
		if err != nil {
			return err
		}
		fmt.Print("The users table has been reset")
	}
	return nil
}
