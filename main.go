package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ShahabGerivani/ghostops/processor"
)

func main() {
	o := flag.String("o", "", "JSON Output File Path")
	s := flag.Bool("s", false, "Show short answer in terminal (only analytics, no incidents. Doesn't affect JSON output)")
	t := flag.Bool("t", false, "Show timestamps for incidents in terminal (Doesn't affect JSON output)")
	flag.Parse()

	logFilePath := flag.Arg(0)
	if logFilePath == "" {
		log.Fatal("usage: app [-o report.json] [-s] [-t] <input file>")
	}

	result, err := processor.ProcessLogFile(logFilePath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total Number of Lines: %d\n", result.TotalLines)
	fmt.Printf("Number of Valid Lines: %d\n", result.ValidLines)
	fmt.Printf("Number of Malformed Lines: %d\n", result.MalformedLines)
	fmt.Printf("Number of Entries by Severity:\n")
	for severity, count := range result.EntryCountsBySeverity {
		fmt.Printf("\t%s: %d\n", severity, count)
	}
	fmt.Printf("Number of Entries by Service:\n")
	for service, count := range result.EntryCountsByService {
		fmt.Printf("\t%s: %d\n", service, count)
	}
	fmt.Println("Number of Incidents by Service:")
	for service, incidents := range result.Incidents {
		fmt.Printf("\t%s: %d\n", service, len(incidents))
	}
	if !*s {
		fmt.Println("Incidents:")
		for service, incidents := range result.Incidents {
			fmt.Printf("\tIncidents of %s:\n", service)
			for i, incident := range incidents {
				fmt.Printf("\t\tIncident %d:\n", i+1)
				fmt.Printf("\t\t\tCount: %d\n", len(incident.TimeWindows))
				fmt.Printf("\t\t\tEntries:\n")
				for _, entry := range incident.Entries {
					fmt.Printf("\t\t\t\t%s\n", strings.Join([]string{processor.ValidSeverityLevels[entry.Severity], entry.Message}, " "))
				}
				if !*t {
					continue
				}
				fmt.Printf("\t\t\tTimestamps:\n")
				for _, timeWindow := range incident.TimeWindows {
					fmt.Printf("\t\t\t\t%s\n", timeWindow.Start)
				}
			}
		}
	}

	if *o != "" {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(*o, data, 0644); err != nil {
			log.Fatal(err)
		}
	}
}
