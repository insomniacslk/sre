package cli

import (
	"strings"

	"github.com/insomniacslk/sre/pkg/config"
)

// termSeparators are stripped during normalization so that, e.g., "bare-metal",
// "bare_metal", "bare metal" and "baremetal" all compare equal.
var termSeparators = strings.NewReplacer("-", "", "_", "", " ", "", "/", "", ".", "")

// normalizeTerm canonicalizes a term for matching: trims, optionally lowercases
// (unless caseSensitive), and removes common separators.
func normalizeTerm(s string, caseSensitive bool) string {
	s = strings.TrimSpace(s)
	if !caseSensitive {
		s = strings.ToLower(s)
	}
	return termSeparators.Replace(s)
}

// expandSynonyms returns the set of normalized terms equivalent to the query:
// the query itself plus every member of any synonym group it belongs to.
func expandSynonyms(query string, groups [][]string, caseSensitive bool) map[string]struct{} {
	nq := normalizeTerm(query, caseSensitive)
	out := map[string]struct{}{nq: {}}
	for _, g := range groups {
		normalized := make([]string, 0, len(g))
		member := false
		for _, m := range g {
			nm := normalizeTerm(m, caseSensitive)
			normalized = append(normalized, nm)
			if nm == nq {
				member = true
			}
		}
		if member {
			for _, nm := range normalized {
				out[nm] = struct{}{}
			}
		}
	}
	return out
}

// shortlistEntryMatches reports whether a shortlist entry matches any of the
// (already synonym-expanded, normalized) query terms. Targets are the entry's
// component, name and aliases. In exact mode a target must equal a query term;
// otherwise a bidirectional substring match (fuzzy) is enough.
func shortlistEntryMatches(e config.OncallShortlistEntry, queryTerms map[string]struct{}, exact, caseSensitive bool) bool {
	targets := make([]string, 0, 2+len(e.Aliases))
	targets = append(targets, e.Component, e.Name)
	targets = append(targets, e.Aliases...)
	for _, t := range targets {
		if strings.TrimSpace(t) == "" {
			continue
		}
		nt := normalizeTerm(t, caseSensitive)
		for qt := range queryTerms {
			if qt == "" {
				continue
			}
			if exact {
				if nt == qt {
					return true
				}
			} else if strings.Contains(nt, qt) || strings.Contains(qt, nt) {
				return true
			}
		}
	}
	return false
}

// selectShortlistEntries returns the entries matching the query. An empty query
// matches everything. Synonym groups (from `oncall.synonyms` in the config) let
// a search term match entries tagged with an equivalent term.
func selectShortlistEntries(entries []config.OncallShortlistEntry, query string, synonyms [][]string, exact, caseSensitive bool) []config.OncallShortlistEntry {
	if strings.TrimSpace(query) == "" {
		return entries
	}
	queryTerms := expandSynonyms(query, synonyms, caseSensitive)

	selected := make([]config.OncallShortlistEntry, 0, len(entries))
	for _, e := range entries {
		if shortlistEntryMatches(e, queryTerms, exact, caseSensitive) {
			selected = append(selected, e)
		}
	}
	return selected
}
