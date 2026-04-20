package main

import (
	"strings"
	"testing"
)

func TestGetPoster(t *testing.T) {
	ttid := "tt0082674" // ID for "The Evil Dead" (1981)

	posterURL := getPoster(ttid)

	if posterURL == "" {
		t.Fatalf("getPoster(%q) returned an empty string, expected a valid URL", ttid)
	}

	if !strings.HasPrefix(posterURL, "https://") {
		t.Errorf("getPoster(%q) returned %q, expected a URL starting with 'https://'", ttid, posterURL)
	}

	if !strings.Contains(posterURL, "media-amazon.com") {
		t.Logf("Warning: Expected poster URL to typically contain 'media-amazon.com', got: %s", posterURL)
	}
}
