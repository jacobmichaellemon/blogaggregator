package main

import (
	"fmt"
	"main/internal/config"
	"os"
)

func main() {
	config, err := config.Read()
	if err != nil {
		fmt.Printf("Error reading config file %s", err)
	}

	appState := state{cfg: &config}
	cmds := commands{availibleCommands: map[string]func(*state, command) error{}}
	cmds.register("login", handlerLogin)
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
		err := s.cfg.SetUser(cmd.args[0])
		if err != nil {
			return err
		}
		fmt.Printf("The user has been set to: %s", cmd.args[0])
	}
	return nil
}
