package extraction

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// PromptLoader loads prompts from files with caching
type PromptLoader struct {
	promptDir string
	cache     map[string]string
	mu        sync.RWMutex
}

// NewPromptLoader creates a new prompt loader
// If promptDir is empty, returns nil (signals to use hardcoded prompts)
func NewPromptLoader(promptDir string) *PromptLoader {
	if promptDir == "" {
		return nil
	}
	return &PromptLoader{
		promptDir: promptDir,
		cache:     make(map[string]string),
	}
}

// LoadSystemPrompt loads a system prompt by version (e.g., "v5")
func (l *PromptLoader) LoadSystemPrompt(version string) (string, error) {
	if l == nil {
		return "", fmt.Errorf("prompt loader not configured")
	}

	cacheKey := "system/" + version

	// Check cache first
	l.mu.RLock()
	if cached, ok := l.cache[cacheKey]; ok {
		l.mu.RUnlock()
		return cached, nil
	}
	l.mu.RUnlock()

	// Load from file
	path := filepath.Join(l.promptDir, "system", version+".txt")
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read system prompt %s: %w", path, err)
	}

	// Cache it
	l.mu.Lock()
	l.cache[cacheKey] = string(content)
	l.mu.Unlock()

	return string(content), nil
}

// LoadFewShotPrompt loads a few-shot prompt by domain (e.g., "brf-v4", "mna-v5", "novel")
func (l *PromptLoader) LoadFewShotPrompt(domain string) (string, error) {
	if l == nil {
		return "", fmt.Errorf("prompt loader not configured")
	}

	cacheKey := "fewshot/" + domain

	// Check cache first
	l.mu.RLock()
	if cached, ok := l.cache[cacheKey]; ok {
		l.mu.RUnlock()
		return cached, nil
	}
	l.mu.RUnlock()

	// Load from file
	path := filepath.Join(l.promptDir, "fewshot", domain+".txt")
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read few-shot prompt %s: %w", path, err)
	}

	// Cache it
	l.mu.Lock()
	l.cache[cacheKey] = string(content)
	l.mu.Unlock()

	return string(content), nil
}

// LoadPostProcessPrompt loads a post-processing prompt by name (e.g., "inconsistency", "dedup")
func (l *PromptLoader) LoadPostProcessPrompt(name string) (string, error) {
	if l == nil {
		return "", fmt.Errorf("prompt loader not configured")
	}

	cacheKey := "postprocess/" + name

	// Check cache first
	l.mu.RLock()
	if cached, ok := l.cache[cacheKey]; ok {
		l.mu.RUnlock()
		return cached, nil
	}
	l.mu.RUnlock()

	// Load from file
	path := filepath.Join(l.promptDir, "postprocess", name+".txt")
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read post-process prompt %s: %w", path, err)
	}

	// Cache it
	l.mu.Lock()
	l.cache[cacheKey] = string(content)
	l.mu.Unlock()

	return string(content), nil
}

// IsConfigured returns true if the prompt loader is ready to use
func (l *PromptLoader) IsConfigured() bool {
	return l != nil && l.promptDir != ""
}
