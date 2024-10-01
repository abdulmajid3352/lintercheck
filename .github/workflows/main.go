package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

type Note struct {
	Label []string `json:"label"`
	Text  string   `json:"text"`
}

type ReleaseNote struct {
	AddonName   string    `json:"addonName"`
	Version     string    `json:"version"`
	ReleaseDate string    `json:"releaseDate"`
	Notes       []Note    `json:"notes"`
}

var allowedLabels = []string{"kind/bug-fix", "N/A", "kind/feature", "kind/upgrade-consideration", "kind/breaking-change", "kind/api-change", "kind/deprecation", "impact/high", "impact/medium"}

func validateJSON(file string) error {
	// Use os.ReadFile instead of ioutil.ReadFile
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("could not read file %s: %v", file, err)
	}

	var releaseNote ReleaseNote
	if err := json.Unmarshal(data, &releaseNote); err != nil {
		return fmt.Errorf("invalid JSON syntax in %s", file)
	}

	if releaseNote.AddonName == "" || releaseNote.Version == "" || releaseNote.ReleaseDate == "" || releaseNote.Notes == nil {
		return fmt.Errorf("missing one or more required fields (addonName, version, releaseDate, notes) in %s", file)
	}

	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !dateRegex.MatchString(releaseNote.ReleaseDate) {
		return fmt.Errorf("invalid date format in %s. Expected YYYY-MM-DD", file)
	}

	if _, err := time.Parse("2006-01-02", releaseNote.ReleaseDate); err != nil {
		return fmt.Errorf("invalid date value in %s: %v", file, err)
	}

	for i, note := range releaseNote.Notes {
		if len(note.Label) == 0 || note.Text == "" {
			return fmt.Errorf("note entry %d in %s has an empty 'label' or 'text'", i, file)
		}

		for _, label := range note.Label {
			if !contains(allowedLabels, label) {
				return fmt.Errorf("invalid label '%s' in note entry %d of %s. Allowed labels are: %v", label, i, file, allowedLabels)
			}
		}
	}

	return nil
}

func validateRaw(file string) error {
	// Use os.ReadFile instead of ioutil.ReadFile
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("could not read file %s: %v", file, err)
	}

	sourceRegex := regexp.MustCompile(`(?m)^source: (https?://.*)`)
	matches := sourceRegex.FindStringSubmatch(string(data))
	if len(matches) == 0 {
		return fmt.Errorf("%s does not contain a valid 'source: <link>' line", file)
	}

	link := matches[1]
	resp, err := http.Get(link)
	if err != nil || resp.StatusCode != 200 {
		return fmt.Errorf("the link %s returned HTTP status %d", link, resp.StatusCode)
	}

	return nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func main() {
	var errorCount int

	err := filepath.Walk("addons", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".json" && filepath.Base(path) == "curated-release-notes.json" {
			if err := validateJSON(path); err != nil {
				log.Println(err)
				errorCount++
			}
		}

		if filepath.Ext(path) == ".txt" && filepath.Base(path) == "raw-release-notes.txt" {
			if err := validateRaw(path); err != nil {
				log.Println(err)
				errorCount++
			}
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	if errorCount > 0 {
		log.Printf("Validation complete. Found %d errors.", errorCount)
		os.Exit(1)
	} else {
		log.Println("All files passed validation checks.")
	}
}
