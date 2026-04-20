package main

import (
	"fmt"
	"log"
)

func main() {
	titles, err := getTitleBasics()
	if err != nil {
		log.Fatalf("Error getting title basics: %v", err)
	}

	fmt.Printf("Successfully imported %d titles.\n", len(titles))
}
