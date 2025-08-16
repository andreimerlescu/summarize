package main

// simplify takes a list of strings and reduces duplicates from the slice
func simplify(t []string) []string {
	seen := make(map[string]bool)
	results := make([]string, 0)
	for _, v := range t {
		if !seen[v] {
			seen[v] = true
			results = append(results, v)
		}
	}
	return results
}
