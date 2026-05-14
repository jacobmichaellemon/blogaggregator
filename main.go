package main

import (
	"fmt"
	"main/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("Error reading config file %s", err)
	}

	cfg.SetUser("Lovely Lemon")

	cfg, err = config.Read()
	if err != nil {
		fmt.Printf("Error reading config file %s", err)
	}
}
