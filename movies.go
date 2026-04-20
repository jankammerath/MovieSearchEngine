package main

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// TitleBasic represents a row in the title.basics.tsv dataset
type TitleBasic struct {
	TConst         string
	TitleType      string
	PrimaryTitle   string
	OriginalTitle  string
	IsAdult        bool
	StartYear      string // Using string to keep \N as is, could be parsed to int if needed
	EndYear        string
	RuntimeMinutes string
	Genres         []string
}

// TitleRating represents a row in the title.ratings.tsv dataset
type TitleRating struct {
	TConst        string
	AverageRating string
	NumVotes      string
}

func getPoster(ttid string) string {
	url := fmt.Sprintf("https://pro.imdb.com/title/%s/", ttid)

	println("Fetching poster from:", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}

	// Set a realistic User-Agent as IMDb often blocks defaults
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to fetch %s: %v\n", url, err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	println("Response body length:", len(body))

	// Makes the quotes optional and captures until the next quote, space, or closing bracket >
	re := regexp.MustCompile(`property="og:image"[^>]*content=["']?([^"'\s>]+)["']?`)
	matches := re.FindSubmatch(body)
	if len(matches) > 1 {
		return string(matches[1])
	}
	return ""
}

func getTitleBasics() ([]TitleBasic, error) {
	url := "https://datasets.imdbws.com/title.basics.tsv.gz"
	fmt.Printf("Downloading dataset from %s...\n", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Bad status: %s", resp.Status)
	}

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		log.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	csvReader := csv.NewReader(gzReader)
	csvReader.Comma = '\t'
	csvReader.LazyQuotes = true    // Some IMDB titles might have unescaped quotes
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields if malformed

	// Read and discard header
	_, err = csvReader.Read()
	if err != nil {
		log.Fatalf("Failed to read header: %v", err)
	}

	var titles []TitleBasic

	fmt.Println("Parsing dataset...")
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Skip malformed rows
			continue
		}

		// Ensure we have enough columns for the format
		if len(record) < 9 {
			continue
		}

		isAdult := record[4] == "1"

		var genres []string
		if record[8] != "\\N" && record[8] != "" {
			genres = strings.Split(record[8], ",")
		}

		title := TitleBasic{
			TConst:         record[0],
			TitleType:      record[1],
			PrimaryTitle:   record[2],
			OriginalTitle:  record[3],
			IsAdult:        isAdult,
			StartYear:      record[5],
			EndYear:        record[6],
			RuntimeMinutes: record[7],
			Genres:         genres,
		}

		titles = append(titles, title)
	}

	return titles, nil
}

func getTitleRatings() (map[string]TitleRating, error) {
	url := "https://datasets.imdbws.com/title.ratings.tsv.gz"
	fmt.Printf("Downloading dataset from %s...\n", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Bad status: %s", resp.Status)
	}

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		log.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	csvReader := csv.NewReader(gzReader)
	csvReader.Comma = '\t'
	csvReader.LazyQuotes = true
	csvReader.FieldsPerRecord = -1

	// Read and discard header
	_, err = csvReader.Read()
	if err != nil {
		log.Fatalf("Failed to read header: %v", err)
	}

	ratings := make(map[string]TitleRating)

	fmt.Println("Parsing ratings dataset...")
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if len(record) < 3 {
			continue
		}

		ratings[record[0]] = TitleRating{
			TConst:        record[0],
			AverageRating: record[1],
			NumVotes:      record[2],
		}
	}

	return ratings, nil
}

func getTitleLanguages() (map[string][]string, error) {
	url := "https://datasets.imdbws.com/title.akas.tsv.gz"
	fmt.Printf("Downloading dataset from %s...\n", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Bad status: %s", resp.Status)
	}

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		log.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	csvReader := csv.NewReader(gzReader)
	csvReader.Comma = '\t'
	csvReader.LazyQuotes = true
	csvReader.FieldsPerRecord = -1

	// Read and discard header
	_, err = csvReader.Read()
	if err != nil {
		log.Fatalf("Failed to read header: %v", err)
	}

	languages := make(map[string][]string)

	fmt.Println("Parsing languages dataset...")
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if len(record) < 5 {
			continue
		}

		titleId := record[0]
		language := record[4]

		if language != "\\N" && language != "" {
			found := false
			for _, l := range languages[titleId] {
				if l == language {
					found = true
					break
				}
			}
			if !found {
				languages[titleId] = append(languages[titleId], language)
			}
		}
	}

	return languages, nil
}
