package main

import (
	"fmt"
	"log"

	"github.com/frankielb/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	err = cfg.SetUser("frankie")
	if err != nil {
		log.Fatalf("Failed to set user: %v", err)
	}

	updatedCfg, err := config.Read()
	if err != nil {
		log.Fatalf("Failed to read updated config: %v", err)
	}

	fmt.Println(updatedCfg.DbURL)
	fmt.Println(updatedCfg.CurrentUserName)
}
