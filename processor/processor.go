package processor

import (
	"bufio"
	"encoding/json"
	"os"
	"slices"
	"strings"
	"time"
)

type ProcessorResult struct {
	TotalLines            int
	ValidLines            int
	MalformedLines        int
	EntryCountsBySeverity map[string]int
	EntryCountsByService  map[string]int
	Incidents             map[string][]Incident
	ValidSeverityLevels   []string
}

func ProcessLogFile(logFilePath string, incidentConfigFilePath string) (ProcessorResult, error) {
	result := ProcessorResult{
		EntryCountsBySeverity: make(map[string]int),
		EntryCountsByService:  make(map[string]int),
		Incidents:             make(map[string][]Incident),
	}

	// Opening the log file
	f, err := os.Open(logFilePath)
	if err != nil {
		return result, err
	}
	defer f.Close()

	// Loading incident config from file
	configData, configErr := os.ReadFile(incidentConfigFilePath)
	if configErr != nil {
		return result, configErr
	}
	incidentConfig := IncidentConfig{}
	if err := json.Unmarshal(configData, &incidentConfig); err != nil {
		return result, err
	}
	result.ValidSeverityLevels = incidentConfig.ValidSeverityLevels

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
		if !slices.Contains(incidentConfig.ValidSeverityLevels, severityText) {
			result.MalformedLines++
			continue
		}

		service := tokens[2]

		message := strings.Join(tokens[3:], " ")

		entry := Entry{entryTime, slices.Index(incidentConfig.ValidSeverityLevels, severityText), service, message, entryText}
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
