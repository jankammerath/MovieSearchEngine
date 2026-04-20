package main

import (
	"fmt"
	"sort"
	"strconv"
)

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsInt16(slice []int16, item int16) bool {
	for _, i := range slice {
		if i == item {
			return true
		}
	}
	return false
}

type MovieItem struct {
	MovieID       string   `json:"movieId"`
	Title         string   `json:"title"`
	Year          int16    `json:"year"`
	Genres        []string `json:"genres"`
	Languages     []string `json:"languages"`
	AverageRating float64  `json:"averageRating,omitempty"`
	NumVotes      int      `json:"numVotes,omitempty"`
}

type SearchEngine struct {
	movies      []MovieItem
	yearMap     map[int16][]*MovieItem
	genreMap    map[string][]*MovieItem
	languageMap map[string][]*MovieItem
	years       []int16
	genres      []string
	languages   []string
}

func NewSearchEngine(movies []TitleBasic, ratings map[string]TitleRating, languages map[string][]string) *SearchEngine {
	engine := &SearchEngine{
		movies:      make([]MovieItem, 0, len(movies)),
		yearMap:     make(map[int16][]*MovieItem),
		genreMap:    make(map[string][]*MovieItem),
		languageMap: make(map[string][]*MovieItem),
		years:       make([]int16, 0),
		genres:      make([]string, 0),
		languages:   make([]string, 0),
	}

	for _, title := range movies {
		if title.TitleType != "movie" {
			continue
		}

		var year int16
		if title.StartYear != "\\N" {
			fmt.Sscanf(title.StartYear, "%d", &year)
		}

		var avgRating float64
		var numVotes int
		if rating, ok := ratings[title.TConst]; ok {
			avgRating, _ = strconv.ParseFloat(rating.AverageRating, 64)
			numVotes, _ = strconv.Atoi(rating.NumVotes)
		}

		movieLangs := languages[title.TConst]

		movie := MovieItem{
			MovieID:       title.TConst,
			Title:         title.PrimaryTitle,
			Year:          year,
			Genres:        title.Genres,
			Languages:     movieLangs,
			AverageRating: avgRating,
			NumVotes:      numVotes,
		}

		engine.movies = append(engine.movies, movie)
	}

	sort.Slice(engine.movies, func(i, j int) bool {
		return engine.movies[i].NumVotes > engine.movies[j].NumVotes
	})

	for i := range engine.movies {
		m := &engine.movies[i]
		year := m.Year

		if year != 0 {
			engine.yearMap[year] = append(engine.yearMap[year], m)
			if !containsInt16(engine.years, year) {
				engine.years = append(engine.years, year)
			}
		}

		for _, genre := range m.Genres {
			engine.genreMap[genre] = append(engine.genreMap[genre], m)
			if !containsString(engine.genres, genre) {
				engine.genres = append(engine.genres, genre)
			}
		}

		for _, lang := range m.Languages {
			engine.languageMap[lang] = append(engine.languageMap[lang], m)
			if !containsString(engine.languages, lang) {
				engine.languages = append(engine.languages, lang)
			}
		}
	}

	return engine
}

// Search returns a list of movies matching the optional year range, genre and language.
func (e *SearchEngine) Search(startYear, endYear int16, genre string, language string, offset, limit int) []*MovieItem {
	if limit <= 0 {
		limit = 32
	}
	if offset < 0 {
		offset = 0
	}

	results := make([]*MovieItem, 0, limit)
	skipped := 0

	// Helper function to check filters and manage offset/limit
	// Returns true if we have reached our limit and should stop searching.
	addIfMatch := func(m *MovieItem) bool {
		if (startYear != 0 && m.Year < startYear) || (endYear != 0 && m.Year > endYear) {
			return false
		}
		if genre != "" && !containsString(m.Genres, genre) {
			return false
		}
		if language != "" && !containsString(m.Languages, language) {
			return false
		}

		if skipped < offset {
			skipped++
			return false
		}

		results = append(results, m)
		return len(results) == limit
	}

	// Determine the smallest slice to iterate through to minimize pointer chasing.
	// Since all these slices were built from e.movies (which is sorted by NumVotes),
	// iterating them sequentially automatically preserves the NumVotes ranking order!
	var searchSet []*MovieItem

	if genre != "" && language != "" {
		// Pick the smaller subset
		if len(e.genreMap[genre]) < len(e.languageMap[language]) {
			searchSet = e.genreMap[genre]
		} else {
			searchSet = e.languageMap[language]
		}
	} else if genre != "" {
		searchSet = e.genreMap[genre]
	} else if language != "" {
		searchSet = e.languageMap[language]
	}

	// If we determined a specific map index is best:
	if searchSet != nil {
		for _, m := range searchSet {
			if addIfMatch(m) {
				break
			}
		}
	} else {
		// If only years are specified, or no filters are specified,
		// iterate the main array to preserve global NumVotes sorting.
		for i := range e.movies {
			if addIfMatch(&e.movies[i]) {
				break
			}
		}
	}

	return results
}
