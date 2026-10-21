package processor

import (
	"bufio"
	"os"
	"slices"
	"strings"
	"time"
)

var ValidSeverityLevels = []string{"INFO", "WARN", "ERROR", "FATAL"}

type ProcessorResult struct {
	TotalLines            int
	ValidLines            int
	MalformedLines        int
	EntryCountsBySeverity map[string]int
	EntryCountsByService  map[string]int
	Incidents             map[string][]Incident
}

func ProcessLogFile(filePath string) (ProcessorResult, error) {
	result := ProcessorResult{
		EntryCountsBySeverity: make(map[string]int),
		EntryCountsByService:  make(map[string]int),
		Incidents:             make(map[string][]Incident),
	}

	f, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer f.Close()

	incidentConfig := IncidentConfig{1, time.Minute, []int{1, 3, 2, 1}}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		entryText := scanner.Text()
		tokens := strings.Fields(entryText)
		result.TotalLines++

		entryTime, timeErr := time.Parse(time.RFC3339, tokens[0])
		if timeErr != nil {
			result.MalformedLines++
			continue
		}

		severityText := tokens[1]
		if !slices.Contains(ValidSeverityLevels, severityText) {
			result.MalformedLines++
			continue
		}

		service := tokens[2]

		message := strings.Join(tokens[3:], " ")

		entry := Entry{entryTime, slices.Index(ValidSeverityLevels, severityText), service, message, entryText}
		addToIncidents(incidentConfig, result.Incidents, entry)

		result.ValidLines++
		result.EntryCountsBySeverity[severityText]++
		result.EntryCountsByService[service]++
	}
	checkMinEntriesInIncidents(incidentConfig, result.Incidents)
	groupMatchingIncidents(result.Incidents)
	sortTimeWindows(result.Incidents)
	err = scanner.Err()
	return result, err
}
