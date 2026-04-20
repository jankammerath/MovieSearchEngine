package main

import "fmt"

type MovieItem struct {
	movieId string
	title   string
	year    int16
	genres  []string
}

type SearchEngine struct {
	movies   []MovieItem
	yearMap  map[int16][]*MovieItem
	genreMap map[string][]*MovieItem
}

func NewSearchEngine(movies []TitleBasic) *SearchEngine {
	engine := &SearchEngine{
		movies:   make([]MovieItem, 0, len(movies)),
		yearMap:  make(map[int16][]*MovieItem),
		genreMap: make(map[string][]*MovieItem),
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
			movieId: title.TConst,
			title:   title.PrimaryTitle,
			year:    year,
			genres:  title.Genres,
		}

		engine.movies = append(engine.movies, movie)

		if year != 0 {
			engine.yearMap[year] = append(engine.yearMap[year], &engine.movies[len(engine.movies)-1])
		}

		for _, genre := range title.Genres {
			engine.genreMap[genre] = append(engine.genreMap[genre], &engine.movies[len(engine.movies)-1])
		}
	}

	return engine
}
