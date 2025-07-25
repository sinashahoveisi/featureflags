package entity

import (
	"time"
)

type FlagStatus string

const (
	FlagEnabled  FlagStatus = "enabled"
	FlagDisabled FlagStatus = "disabled"
)

// Flag represents the main feature flag entity with business logic
type Flag struct {
	ID           int64       `json:"id" db:"id"`
	Name         string      `json:"name" db:"name"`
	Status       FlagStatus  `json:"status" db:"status"`
	Dependencies []int64     `json:"dependencies,omitempty"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
}

// IsEnabled returns true if the flag is enabled
func (f *Flag) IsEnabled() bool {
	return f.Status == FlagEnabled
}

// IsDisabled returns true if the flag is disabled
func (f *Flag) IsDisabled() bool {
	return f.Status == FlagDisabled
}

// Enable sets the flag status to enabled
func (f *Flag) Enable() {
	f.Status = FlagEnabled
	f.UpdatedAt = time.Now()
}

// Disable sets the flag status to disabled
func (f *Flag) Disable() {
	f.Status = FlagDisabled
	f.UpdatedAt = time.Now()
}

// HasDependencies returns true if the flag has dependencies
func (f *Flag) HasDependencies() bool {
	return len(f.Dependencies) > 0
}

// AddDependency adds a dependency to the flag
func (f *Flag) AddDependency(dependencyID int64) {
	for _, dep := range f.Dependencies {
		if dep == dependencyID {
			return // Already exists
		}
	}
	f.Dependencies = append(f.Dependencies, dependencyID)
}

// RemoveDependency removes a dependency from the flag
func (f *Flag) RemoveDependency(dependencyID int64) {
	for i, dep := range f.Dependencies {
		if dep == dependencyID {
			f.Dependencies = append(f.Dependencies[:i], f.Dependencies[i+1:]...)
			return
		}
	}
} 