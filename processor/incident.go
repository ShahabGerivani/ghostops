package processor

import (
	"slices"
	"time"
)

type Entry struct {
	Time     time.Time
	Severity int
	Service  string
	Message  string
	Text     string
}

func (entry Entry) String() string {
	return entry.Text
}

type IncidentConfig struct {
	MinSeverityLevel     int
	Duration             time.Duration
	MinEntriesBySeverity []int // The indeces represent severity levels and the values represent the minimum number of entries
}

type Incident struct {
	Entries     []Entry
	TimeWindows []IncidentTime
}

type IncidentTime struct {
	Start time.Time
	End   time.Time
}

func addToIncidents(config IncidentConfig, incidents map[string][]Incident, entry Entry) {
	if entry.Severity < config.MinSeverityLevel {
		return
	}

	// No incident for this service yet
	_, ok := incidents[entry.Service]
	if !ok {
		incidents[entry.Service] = []Incident{{[]Entry{entry}, []IncidentTime{{entry.Time, entry.Time}}}}
		return
	}

	newIncident := true
	for i, incident := range incidents[entry.Service] {
		if !entry.Time.Before(incident.TimeWindows[0].Start.Add(-config.Duration)) && entry.Time.Before(incident.TimeWindows[0].End.Add(config.Duration)) {
			if entry.Time.Before(incident.TimeWindows[0].Start) {
				incidents[entry.Service][i].TimeWindows[0].Start = entry.Time
			}
			if !entry.Time.Before(incident.TimeWindows[0].End) {
				incidents[entry.Service][i].TimeWindows[0].End = entry.Time
			}
			incidents[entry.Service][i].Entries = append(incident.Entries, entry)
			newIncident = false
			break
		}
	}
	if newIncident {
		incidents[entry.Service] = append(incidents[entry.Service], Incident{[]Entry{entry}, []IncidentTime{{entry.Time, entry.Time}}})
	}
}

func checkMinEntriesInIncidents(config IncidentConfig, incidents map[string][]Incident) {
	for service := range incidents {
		incidents[service] = slices.DeleteFunc(incidents[service], func(incident Incident) bool {
			entriesBySeverity := make([]int, len(config.MinEntriesBySeverity))
			for _, entry := range incident.Entries {
				entriesBySeverity[entry.Severity]++
			}
			for severity, entryCount := range entriesBySeverity {
				if entryCount >= config.MinEntriesBySeverity[severity] {
					return false
				}
			}
			return true
		})
	}
}

func groupMatchingIncidents(incidents map[string][]Incident) {
	for service, list := range incidents {
		kept := list[:0]
		for _, incident := range list {
			merged := false
			for j := range kept {
				if sameEntries(incident.Entries, kept[j].Entries) {
					kept[j].TimeWindows = append(kept[j].TimeWindows, incident.TimeWindows...)
					merged = true
					break
				}
			}
			if !merged {
				kept = append(kept, incident)
			}
		}
		clear(list[len(kept):]) // drop stale references so they can be GC'd
		incidents[service] = kept
	}
}

func sameEntries(a, b []Entry) bool {
	if len(a) != len(b) {
		return false
	}
	for _, e := range a {
		if !slices.ContainsFunc(b, func(e2 Entry) bool {
			return e.Severity == e2.Severity && e.Message == e2.Message
		}) {
			return false
		}
	}
	return true
}

func sortTimeWindows(incidents map[string][]Incident) {
	for service := range incidents {
		for i := range incidents[service] {
			slices.SortFunc(incidents[service][i].TimeWindows, func(a, b IncidentTime) int {
				return a.Start.Compare(b.Start)
			})
		}
	}
}
