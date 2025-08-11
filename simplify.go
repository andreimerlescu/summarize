package main

// simplify takes a list of strings and reduces duplicates from the slice
func simplify(t []string) []string {
	seen := make(map[string]bool)
	for _, v := range t {
		seen[v] = true
	}
	results := make([]string, len(t))
	for i, v := range t {
		if seen[v] {
			results[i] = v
		}
	}
	return results
}
