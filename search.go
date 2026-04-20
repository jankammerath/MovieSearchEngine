package main

import "fmt"

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
	MovieID string   `json:"movieId"`
	Title   string   `json:"title"`
	Year    int16    `json:"year"`
	Genres  []string `json:"genres"`
}

type SearchEngine struct {
	movies   []MovieItem
	yearMap  map[int16][]*MovieItem
	genreMap map[string][]*MovieItem
	years    []int16
	genres   []string
}

func NewSearchEngine(movies []TitleBasic) *SearchEngine {
	engine := &SearchEngine{
		movies:   make([]MovieItem, 0, len(movies)),
		yearMap:  make(map[int16][]*MovieItem),
		genreMap: make(map[string][]*MovieItem),
		years:    make([]int16, 0),
		genres:   make([]string, 0),
	}

	for _, title := range movies {
		if title.TitleType != "movie" {
			continue
		}

		var year int16
		if title.StartYear != "\\N" {
			fmt.Sscanf(title.StartYear, "%d", &year)
		}

		movie := MovieItem{
			MovieID: title.TConst,
			Title:   title.PrimaryTitle,
			Year:    year,
			Genres:  title.Genres,
		}

		engine.movies = append(engine.movies, movie)

		if year != 0 {
			engine.yearMap[year] = append(engine.yearMap[year], &engine.movies[len(engine.movies)-1])
			if !containsInt16(engine.years, year) {
				engine.years = append(engine.years, year)
			}
		}

		for _, genre := range title.Genres {
			engine.genreMap[genre] = append(engine.genreMap[genre], &engine.movies[len(engine.movies)-1])
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
