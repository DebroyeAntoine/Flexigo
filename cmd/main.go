package main

import (
	"fmt"
	"log"

	"github.com/DebroyeAntoine/flexigo/internal/config"
)

func main() {
	cfg, err := config.LoadConfig("assets/config.yaml")
	if err != nil {
		log.Fatalf("Erreur de chargement de config: %v", err)
	}

	fmt.Println("== Actions disponibles ==")
	for i, block := range cfg.Blocks {
		fmt.Printf("[%d] %s (%s)\n", i+1, block.Label, block.Type)
	}
}
