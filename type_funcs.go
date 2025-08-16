package main

import "fmt"

// Add inserts an entry into the map
func (s *seenStrings) Add(entry string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[entry] = true
}

// Remove uses delete on the entry in the map
func (s *seenStrings) Remove(entry string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, entry)
}

// Len returns the length of the map
func (s *seenStrings) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.m)
}

// String implements the Stringer interface
func (s *seenStrings) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fmt.Sprint(s.m)
}

// True sets the entry to true
func (s *seenStrings) True(entry string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[entry] = true
}

// False sets the entry to false
func (s *seenStrings) False(entry string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, entry)
}

// Exists returns a bool if the map contains the entry
func (s *seenStrings) Exists(entry string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.m[entry]
}
