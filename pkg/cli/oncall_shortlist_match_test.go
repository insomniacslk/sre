package cli

import (
	"testing"

	"github.com/insomniacslk/sre/pkg/config"
)

func TestSelectShortlistEntries(t *testing.T) {
	entries := []config.OncallShortlistEntry{
		{Name: "Kubernetes / Platform", Component: "kubernetes"},
		{Name: "Bare metal / Provisioning", Component: "bare-metal"},
		{Name: "Security", Component: "security"},
		{Name: "Data", Component: "data", Aliases: []string{"db", "database"}},
	}

	tests := []struct {
		name          string
		query         string
		exact         bool
		caseSensitive bool
		wantComps     []string
	}{
		{name: "empty matches all", query: "", wantComps: []string{"kubernetes", "bare-metal", "security", "data"}},
		{name: "synonym k8s -> kubernetes", query: "k8s", wantComps: []string{"kubernetes"}},
		{name: "synonym kubernetes -> k8s label too", query: "kubernetes", wantComps: []string{"kubernetes"}},
		{name: "separator insensitive baremetal", query: "baremetal", wantComps: []string{"bare-metal"}},
		{name: "separator insensitive bare_metal", query: "bare_metal", wantComps: []string{"bare-metal"}},
		{name: "substring fuzzy sec", query: "sec", wantComps: []string{"security"}},
		{name: "alias match db", query: "db", wantComps: []string{"data"}},
		{name: "alias match database", query: "database", wantComps: []string{"data"}},
		{name: "exact rejects substring", query: "sec", exact: true, wantComps: nil},
		{name: "exact accepts full term", query: "security", exact: true, wantComps: []string{"security"}},
		{name: "exact still honors synonyms", query: "k8s", exact: true, wantComps: []string{"kubernetes"}},
		{name: "case-insensitive default", query: "SECURITY", wantComps: []string{"security"}},
		{name: "case-sensitive misses", query: "SECURITY", caseSensitive: true, wantComps: nil},
		{name: "no match", query: "storage", wantComps: nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := selectShortlistEntries(entries, tc.query, nil, tc.exact, tc.caseSensitive)
			var gotComps []string
			for _, e := range got {
				gotComps = append(gotComps, e.Component)
			}
			if len(gotComps) != len(tc.wantComps) {
				t.Fatalf("query %q: got %v, want %v", tc.query, gotComps, tc.wantComps)
			}
			for i := range gotComps {
				if gotComps[i] != tc.wantComps[i] {
					t.Fatalf("query %q: got %v, want %v", tc.query, gotComps, tc.wantComps)
				}
			}
		})
	}
}

func TestSelectShortlistEntriesConfiguredSynonyms(t *testing.T) {
	entries := []config.OncallShortlistEntry{
		{Name: "Storage", Component: "storage"},
	}
	// A configured synonym group makes "disk" match the storage entry.
	got := selectShortlistEntries(entries, "disk", [][]string{{"storage", "disk", "nvme"}}, false, false)
	if len(got) != 1 || got[0].Component != "storage" {
		t.Fatalf("configured synonym: got %v, want [storage]", got)
	}
}
