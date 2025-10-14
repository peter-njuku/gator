package main

import (
	"fmt"
	"log"

	"github.com/peter-njuku/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Unable to Read config file: %v", err)
	}

	err = cfg.SetUser("Peter")
	if err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}

	fmt.Println("Config updated successfully")

	updatedCfg, err := config.Read()
	if err != nil {
		log.Fatalf("Unable to Read config file: %v", err)
	}
	fmt.Println("Updated Config")
	fmt.Printf("%+v\n", updatedCfg)
}
