package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

	r := gin.Default()

	// Serve static files from the "static" directory
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

	// API Endpoints
	api := r.Group("/api")
	{
		api.GET("/genre", func(c *gin.Context) {
			c.JSON(http.StatusOK, searchEngine.genres)
		})

		api.GET("/year", func(c *gin.Context) {
			c.JSON(http.StatusOK, searchEngine.years)
		})

		api.GET("/search", func(c *gin.Context) {
			startYearStr := c.Query("startYear")
			endYearStr := c.Query("endYear")
			genre := c.Query("genre")
			offsetStr := c.Query("offset")
			limitStr := c.Query("limit")

			var startYear, endYear int16
			var offset, limit int

			if startYearStr != "" {
				v, _ := strconv.Atoi(startYearStr)
				startYear = int16(v)
			}
			if endYearStr != "" {
				v, _ := strconv.Atoi(endYearStr)
				endYear = int16(v)
			}
			if offsetStr != "" {
				offset, _ = strconv.Atoi(offsetStr)
			}
			if limitStr != "" {
				limit, _ = strconv.Atoi(limitStr)
			}

			results := searchEngine.Search(startYear, endYear, genre, offset, limit)
			c.JSON(http.StatusOK, results)
		})
	}

	fmt.Println("Server is running on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
