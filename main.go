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

	searchEngine := NewSearchEngine(titles)
	fmt.Printf("Search engine initialized with %d movies.\n", len(searchEngine.movies))
	fmt.Printf("Search engine has %d unique years.\n", len(searchEngine.years))
	fmt.Printf("Search engine has %d unique genres.\n", len(searchEngine.genres))

}
