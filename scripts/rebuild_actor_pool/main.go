// rebuild_actor_pool searches TMDB for every actor by name, resolves the
// correct person ID from the top result, and rewrites
// internal/service/actor_pool.go with accurate IDs.
//
// Usage:
//
//	TMDB_API_KEY=<key> go run ./scripts/rebuild_actor_pool
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"
)

const (
	tmdbBaseURL = "https://api.themoviedb.org/3"
	outputFile  = "internal/service/actor_pool.go"
)

// wantedActors is the canonical list of names, grouped by category.
// Edit this list when adding new actors; the script resolves IDs automatically.
var wantedActors = []section{
	{
		label: "A-list headliners",
		names: []string{
			"Tom Hanks",
			"Brad Pitt",
			"Robert De Niro",
			"Tom Cruise",
			"Jack Nicholson",
			"Morgan Freeman",
			"Al Pacino",
			"Dustin Hoffman",
			"Denzel Washington",
			"Samuel L. Jackson",
			"Leonardo DiCaprio",
			"Matt Damon",
			"Keanu Reeves",
			"Kevin Bacon",
			"Will Smith",
			"Julia Roberts",
			"Meryl Streep",
			"Hugh Jackman",
			"Scarlett Johansson",
			"Natalie Portman",
			"Johnny Depp",
			"Ben Affleck",
			"Bruce Willis",
			"Charlize Theron",
			"Russell Crowe",
			"Nicole Kidman",
			"Clint Eastwood",
			"Harrison Ford",
			"Anthony Hopkins",
			"Helen Mirren",
		},
	},
	{
		label: "Classic Hollywood",
		names: []string{
			"Paul Newman",
			"Marlon Brando",
			"Robert Redford",
			"Gene Hackman",
			"Jeff Bridges",
			"Sean Connery",
			"Christopher Walken",
			"John Malkovich",
			"Harvey Keitel",
			"Joe Pesci",
			"Sean Penn",
			"James Earl Jones",
			"Gene Wilder",
			"Robert Duvall",
			"Michael Douglas",
			"Kevin Costner",
			"Ethan Hawke",
			"Jake Gyllenhaal",
			"Quentin Tarantino",
			"Woody Allen",
		},
	},
	{
		label: "Comedy",
		names: []string{
			"Ian McKellen",
			"Patrick Stewart",
			"Steve Martin",
			"Robin Williams",
			"Danny DeVito",
			"Richard Pryor",
			"Billy Crystal",
			"Jim Carrey",
			"Bill Murray",
			"Alec Baldwin",
			"John Candy",
			"Ryan Reynolds",
			"Adam Sandler",
			"Mike Myers",
			"Gwyneth Paltrow",
			"Jonah Hill",
			"Seth Rogen",
			"Tina Fey",
			"Melissa McCarthy",
		},
	},
	{
		label: "Action / Blockbuster",
		names: []string{
			"Sylvester Stallone",
			"Arnold Schwarzenegger",
			"Mel Gibson",
			"Eddie Murphy",
			"Dwayne Johnson",
			"Jason Statham",
			"Wesley Snipes",
			"Nicolas Cage",
			"Gary Busey",
			"Liam Neeson",
			"Jessica Alba",
			"Jean-Claude Van Damme",
			"Gary Oldman",
			"Tommy Lee Jones",
			"Idris Elba",
			"John Goodman",
		},
	},
	{
		label: "Bond Universe",
		names: []string{
			"Pierce Brosnan",
			"Daniel Craig",
			"Judi Dench",
			"Timothy Dalton",
			"Roger Moore",
			"Jeremy Irons",
			"Ralph Fiennes",
			"Rami Malek",
		},
	},
	{
		label: "Marvel / DC",
		names: []string{
			"Robert Downey Jr.",
			"Chris Evans",
			"Chris Hemsworth",
			"Chris Pratt",
			"Mark Ruffalo",
			"Jeremy Renner",
			"Paul Rudd",
			"Chadwick Boseman",
			"Don Cheadle",
			"Zoe Saldana",
			"Michael Fassbender",
			"James McAvoy",
			"Sigourney Weaver",
			"Ryan Gosling",
			"Brie Larson",
		},
	},
	{
		label: "Prestige Drama",
		names: []string{
			"Tim Robbins",
			"Kevin Spacey",
			"Steve Buscemi",
			"William H. Macy",
			"Angelina Jolie",
			"George Clooney",
			"Michael Caine",
			"Emily Blunt",
			"Winona Ryder",
			"Jeff Goldblum",
			"Richard Gere",
			"John Cusack",
			"Philip Seymour Hoffman",
			"Javier Bardem",
			"Antonio Banderas",
			"Daniel Day-Lewis",
			"Julianne Moore",
			"Laura Linney",
			"John C. Reilly",
			"Philip Baker Hall",
			"Terrence Howard",
			"Cuba Gooding Jr.",
		},
	},
	{
		label: "Awards Circuit",
		names: []string{
			"Emma Stone",
			"Jennifer Lawrence",
			"Viola Davis",
			"Cate Blanchett",
			"Anne Hathaway",
			"Amy Adams",
			"Jessica Chastain",
			"Lupita Nyong'o",
			"Saoirse Ronan",
		},
	},
	{
		label: "International Stars",
		names: []string{
			"Adrien Brody",
			"Benicio del Toro",
			"Jude Law",
			"Ewan McGregor",
			"Colin Firth",
			"Penélope Cruz",
			"Salma Hayek",
			"Catherine Zeta-Jones",
			"Clive Owen",
			"Paul Bettany",
			"Dev Patel",
			"Keira Knightley",
		},
	},
	{
		label: "Character Actors",
		names: []string{
			"John Hurt",
			"William Hurt",
			"Michael Shannon",
			"Mahershala Ali",
			"Chiwetel Ejiofor",
		},
	},
	{
		label: "TV Crossover",
		names: []string{
			"Bryan Cranston",
			"James Gandolfini",
			"Jon Hamm",
			"Oscar Isaac",
			"Matthew McConaughey",
		},
	},
	{
		label: "Women in Film",
		names: []string{
			"Sandra Bullock",
			"Reese Witherspoon",
			"Halle Berry",
			"Kate Winslet",
			"Jodie Foster",
			"Marisa Tomei",
			"Uma Thurman",
			"Hilary Swank",
			"Diane Keaton",
			"Jessica Lange",
			"Sissy Spacek",
		},
	},
	{
		label: "New Generation",
		names: []string{
			"Timothée Chalamet",
			"Daniel Kaluuya",
			"Florence Pugh",
			"Zendaya",
			"Tom Holland",
			"Ezra Miller",
			"Jacob Elordi",
			"Austin Butler",
			"Barry Keoghan",
		},
	},
}

// -----------------------------------------------------------------------

type section struct {
	label string
	names []string
}

type resolvedEntry struct {
	ID   int
	Name string // actual TMDB name
}

type resolvedSection struct {
	Label   string
	Entries []resolvedEntry
}

type tmdbPerson struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Popularity float64 `json:"popularity"`
}

type tmdbSearchResp struct {
	Results []tmdbPerson `json:"results"`
}

func main() {
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "error: TMDB_API_KEY environment variable is required")
		os.Exit(1)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	var resolved []resolvedSection
	var failures []string

	for _, sec := range wantedActors {
		fmt.Printf("[%s]\n", sec.label)
		rs := resolvedSection{Label: sec.label}

		for _, name := range sec.names {
			person, err := searchPerson(client, apiKey, name)
			if err != nil {
				failures = append(failures, fmt.Sprintf("  FAILED  %q: %v", name, err))
				fmt.Printf("  FAILED  %q: %v\n", name, err)
				continue
			}
			rs.Entries = append(rs.Entries, resolvedEntry{ID: person.ID, Name: person.Name})
			fmt.Printf("  OK  %-30s  id=%-10d  tmdb_name=%q\n", name, person.ID, person.Name)
			time.Sleep(50 * time.Millisecond)
		}

		resolved = append(resolved, rs)
	}

	if len(failures) > 0 {
		fmt.Printf("\n%d lookup(s) failed — they will be omitted from the output:\n", len(failures))
		for _, f := range failures {
			fmt.Println(f)
		}
	}

	if err := writePoolFile(outputFile, resolved); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", outputFile, err)
		os.Exit(1)
	}

	fmt.Printf("\nWrote %s\n", outputFile)
}

// searchPerson queries /search/person and returns the highest-popularity result.
func searchPerson(client *http.Client, apiKey, name string) (tmdbPerson, error) {
	endpoint := fmt.Sprintf("%s/search/person?query=%s&api_key=%s",
		tmdbBaseURL, url.QueryEscape(name), apiKey)

	resp, err := client.Get(endpoint)
	if err != nil {
		return tmdbPerson{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return tmdbPerson{}, fmt.Errorf("status %d", resp.StatusCode)
	}

	var sr tmdbSearchResp
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return tmdbPerson{}, err
	}
	if len(sr.Results) == 0 {
		return tmdbPerson{}, fmt.Errorf("no results")
	}

	// Pick the result with the highest popularity score.
	best := sr.Results[0]
	for _, r := range sr.Results[1:] {
		if r.Popularity > best.Popularity {
			best = r
		}
	}
	return best, nil
}

// poolFileTemplate is the template for the generated actor_pool.go.
var poolFileTemplate = template.Must(template.New("pool").Parse(`package service

// dailyActorPool is a curated set of prolific, well-connected TMDB actor IDs.
// Every actor here has a large filmography, making a valid chain between any
// two almost always achievable.
//
// To add more actors: add the name to scripts/rebuild_actor_pool/main.go and
// re-run:  TMDB_API_KEY=<key> go run ./scripts/rebuild_actor_pool
// Find an actor's TMDB ID at https://www.themoviedb.org/person/<id>.
var dailyActorPool = []int{
{{- range .}}
	// --- {{.Label}} ---
	{{- range .Entries}}
	{{.ID}}, // {{.Name}}
	{{- end}}
{{end}}}
`))

func writePoolFile(path string, sections []resolvedSection) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Use a builder first so we can clean up extra blank lines from the template.
	var sb strings.Builder
	if err := poolFileTemplate.Execute(&sb, sections); err != nil {
		return err
	}

	_, err = fmt.Fprint(f, sb.String())
	return err
}
