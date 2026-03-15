// verify_actor_pool parses internal/service/actor_pool.go, extracts every
// "id, // Name" entry, then hits the TMDB /person/{id} endpoint to confirm
// the actual name matches the comment.
//
// Usage:
//
//	TMDB_API_KEY=<key> go run ./scripts/verify_actor_pool
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	tmdbBaseURL = "https://api.themoviedb.org/3"
	poolFile    = "internal/service/actor_pool.go"
)

// entryRe matches lines like:  1234,  // Some Name
// or                            1234567, // Some Name
var entryRe = regexp.MustCompile(`^\s+(\d+),\s+//\s+(.+)$`)

type entry struct {
	id      int
	comment string
}

type tmdbPerson struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "error: TMDB_API_KEY environment variable is required")
		os.Exit(1)
	}

	entries, err := parsePool(poolFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing pool file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Verifying %d actor IDs against TMDB...\n\n", len(entries))

	client := &http.Client{Timeout: 10 * time.Second}

	var mismatches []string
	var notFound []string

	for _, e := range entries {
		actual, err := fetchPerson(client, apiKey, e.id)
		if err != nil {
			notFound = append(notFound, fmt.Sprintf("  ID %-10d  comment: %q  error: %v", e.id, e.comment, err))
			fmt.Printf("  NOT FOUND  id=%-10d  comment=%q\n", e.id, e.comment)
			continue
		}

		// Trim any trailing comment noise (e.g. "Gary Oldman  // Note: …")
		wantName := strings.TrimSpace(strings.SplitN(e.comment, "//", 2)[0])

		if !namesMatch(wantName, actual.Name) {
			msg := fmt.Sprintf("  MISMATCH   id=%-10d  want=%q  got=%q", e.id, wantName, actual.Name)
			mismatches = append(mismatches, msg)
			fmt.Println(msg)
		} else {
			fmt.Printf("  OK         id=%-10d  %q\n", e.id, actual.Name)
		}

		// Be polite to the TMDB API.
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Printf("\n--- Summary ---\n")
	fmt.Printf("Total:      %d\n", len(entries))
	fmt.Printf("Mismatches: %d\n", len(mismatches))
	fmt.Printf("Not found:  %d\n", len(notFound))

	if len(mismatches) > 0 {
		fmt.Println("\nMismatches:")
		for _, m := range mismatches {
			fmt.Println(m)
		}
	}

	if len(notFound) > 0 {
		fmt.Println("\nNot found:")
		for _, m := range notFound {
			fmt.Println(m)
		}
	}

	if len(mismatches) > 0 || len(notFound) > 0 {
		os.Exit(1)
	}
}

// parsePool reads the actor pool source file and returns all id/comment pairs.
func parsePool(path string) ([]entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []entry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		m := entryRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		id, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		entries = append(entries, entry{id: id, comment: strings.TrimSpace(m[2])})
	}
	return entries, scanner.Err()
}

// fetchPerson calls GET /person/{id} and returns the result.
func fetchPerson(client *http.Client, apiKey string, id int) (tmdbPerson, error) {
	url := fmt.Sprintf("%s/person/%d?api_key=%s", tmdbBaseURL, id, apiKey)
	resp, err := client.Get(url)
	if err != nil {
		return tmdbPerson{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return tmdbPerson{}, fmt.Errorf("404 not found")
	}
	if resp.StatusCode != http.StatusOK {
		return tmdbPerson{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var p tmdbPerson
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return tmdbPerson{}, err
	}
	return p, nil
}

// namesMatch does a case-insensitive comparison, tolerating minor punctuation
// differences (e.g. "Penelope Cruz" vs "Penélope Cruz").
func namesMatch(want, got string) bool {
	normalize := func(s string) string {
		return strings.ToLower(strings.TrimSpace(s))
	}
	return normalize(want) == normalize(got)
}
