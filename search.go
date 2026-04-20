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
	AverageRating float64  `json:"averageRating,omitempty"`
	NumVotes      int      `json:"numVotes,omitempty"`
}

type SearchEngine struct {
	movies   []MovieItem
	yearMap  map[int16][]*MovieItem
	genreMap map[string][]*MovieItem
	years    []int16
	genres   []string
}

func NewSearchEngine(movies []TitleBasic, ratings map[string]TitleRating) *SearchEngine {
	engine := &SearchEngine{
		movies:   make([]MovieItem, 0, len(movies)),
		yearMap:  make(map[int16][]*MovieItem),
		genreMap: make(map[string][]*MovieItem),
		years:    make([]int16, 0),
		genres:   make([]string, 0),
	}

	var totalRatingSum float64
	var totalVotesSum int

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

		movie := MovieItem{
			MovieID:       title.TConst,
			Title:         title.PrimaryTitle,
			Year:          year,
			Genres:        title.Genres,
			AverageRating: avgRating,
			NumVotes:      numVotes,
		}

		if numVotes > 0 {
			totalRatingSum += avgRating * float64(numVotes)
			totalVotesSum += numVotes
		}

		engine.movies = append(engine.movies, movie)
	}

	var C float64
	if totalVotesSum > 0 {
		C = totalRatingSum / float64(totalVotesSum)
	}
	m := 1000.0

	sort.Slice(engine.movies, func(i, j int) bool {
		mi := engine.movies[i]
		mj := engine.movies[j]

		vi := float64(mi.NumVotes)
		wi := (mi.AverageRating*vi + C*m) / (vi + m)

		vj := float64(mj.NumVotes)
		wj := (mj.AverageRating*vj + C*m) / (vj + m)

		return wi > wj
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
	}

	return engine
}

// Search returns a list of movies matching the optional year range and genre.
// Use 0 for startYear/endYear to represent no bound, and "" for no genre filter.
func (e *SearchEngine) Search(startYear, endYear int16, genre string, offset, limit int) []*MovieItem {
	if limit <= 0 {
		limit = 32
	}
	if offset < 0 {
		offset = 0
	}

	var allResults []*MovieItem

	if startYear != 0 || endYear != 0 {
		// If no genre but there is a year range, use the year map
		for year, movies := range e.yearMap {
			if (startYear == 0 || year >= startYear) && (endYear == 0 || year <= endYear) {
				for _, m := range movies {
					if genre == "" || containsString(m.Genres, genre) {
						allResults = append(allResults, m)
					}
				}
			}
		}
	} else if genre != "" {
		// If a genre is specified, use the genre index first
		for _, m := range e.genreMap[genre] {
			if (startYear == 0 || m.Year >= startYear) && (endYear == 0 || m.Year <= endYear) {
				allResults = append(allResults, m)
			}
		}
	} else {
		// If no filters are provided, return all movies
		for i := range e.movies {
			allResults = append(allResults, &e.movies[i])
		}
	}

	if offset >= len(allResults) {
		return []*MovieItem{}
	}

	end := offset + limit
	if end > len(allResults) {
		end = len(allResults)
	}

	return allResults[offset:end]
}
