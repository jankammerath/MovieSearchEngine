package main

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
