package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {
	titles, err := getTitleBasics()
	if err != nil {
		log.Fatalf("Error getting title basics: %v", err)
	}

	fmt.Printf("Successfully imported %d titles.\n", len(titles))

	ratings, err := getTitleRatings()
	if err != nil {
		log.Fatalf("Error getting title ratings: %v", err)
	}

	fmt.Printf("Successfully imported %d ratings.\n", len(ratings))

	languages, err := getTitleLanguages()
	if err != nil {
		log.Fatalf("Error getting title languages: %v", err)
	}

	fmt.Printf("Successfully imported %d languages.\n", len(languages))

	searchEngine := NewSearchEngine(titles, ratings, languages)
	fmt.Printf("Search engine initialized with %d movies.\n", len(searchEngine.movies))
	fmt.Printf("Search engine has %d unique years.\n", len(searchEngine.years))
	fmt.Printf("Search engine has %d unique genres.\n", len(searchEngine.genres))

	// Free up memory from raw datasets after indexing and trigger Garbage Collection
	titles = nil
	ratings = nil
	languages = nil
	runtime.GC()

	r := gin.Default()

	// Serve static files from the "static" directory
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

	r.GET("/media/:ttid", func(c *gin.Context) {
		ttid := c.Param("ttid")
		posterURL := getPoster(ttid)
		if posterURL == "" {
			c.String(http.StatusNotFound, "Poster not found")
			return
		}

		req, err := http.NewRequest("GET", posterURL, nil)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to prepare image request")
			return
		}

		// Some Amazon/IMDb image servers might require a realistic User-Agent
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			c.String(http.StatusNotFound, "Failed to download image")
			if resp != nil {
				resp.Body.Close()
			}
			return
		}
		defer resp.Body.Close()

		c.Header("Cache-Control", "public, max-age=2592000") // 30 days
		c.DataFromReader(http.StatusOK, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	})

	// API Endpoints
	api := r.Group("/api")
	{
		api.GET("/genre", func(c *gin.Context) {
			c.JSON(http.StatusOK, searchEngine.genres)
		})

		api.GET("/language", func(c *gin.Context) {
			c.JSON(http.StatusOK, searchEngine.languages)
		})

		api.GET("/year", func(c *gin.Context) {
			c.JSON(http.StatusOK, searchEngine.years)
		})

		api.GET("/search", func(c *gin.Context) {
			startYearStr := c.Query("startYear")
			endYearStr := c.Query("endYear")
			genre := c.Query("genre")
			language := c.Query("language")
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

			results := searchEngine.Search(startYear, endYear, genre, language, offset, limit)
			c.JSON(http.StatusOK, results)
		})
	}

	fmt.Println("Server is running on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
