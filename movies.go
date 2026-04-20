package main

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
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
