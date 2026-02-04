package main

import (
	"fmt"
	"log"

	"github.com/DebroyeAntoine/flexigo/internal/config"
	"github.com/DebroyeAntoine/flexigo/internal/ui"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error while loading configuration: %v", err)
	}

	fmt.Println("== Actions disponibles ==")
	for i, block := range cfg.Blocks {
		fmt.Printf("[%d] %s (%s)\n", i+1, block.Label, block.Type)
	}
	uiErr := ui.StartUI(cfg)
	if uiErr != nil {
		// Si StartUI échoue (à cause de l'IR par exemple)
		log.Fatalf("Erreur fatale : %v", uiErr)
	}
}
