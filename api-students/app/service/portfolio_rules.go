package service

import "strings"

func NormalizeTechs(techs []string) []string {
	result := make([]string, 0, len(techs))
	seen := make(map[string]struct{}, len(techs))
	for _, t := range techs {
		clean := strings.ToLower(strings.TrimSpace(t))
		if clean == "" {
			continue
		}
		if _, dup := seen[clean]; dup {
			continue
		}
		seen[clean] = struct{}{}
		result = append(result, clean)
	}
	return result
}