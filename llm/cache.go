package llm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Cache stores the command to service name mappings
type Cache struct {
	mu    sync.RWMutex
	items map[string]string // command -> service name
}

// NewCache creates a new empty cache
func NewCache() *Cache {
	return &Cache{
		items: make(map[string]string),
	}
}

// cachePath returns the full path to the cache file
func cachePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "cache.json"), nil
}

// LoadCache loads the cache from disk
// Returns empty cache if file doesn't exist
func LoadCache() (*Cache, error) {
	path, err := cachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewCache(), nil
		}
		return nil, err
	}

	var items map[string]string
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	return &Cache{items: items}, nil
}

// Save persists the cache to disk
func (c *Cache) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := ensureConfigDir(); err != nil {
		return err
	}

	path, err := cachePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c.items, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// Get retrieves the service name for a command
// Returns empty string if not cached
func (c *Cache) Get(command string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.items[command]
}

// Set stores a command to service name mapping
// Does NOT cache "未知" results
func (c *Cache) Set(command, serviceName string) {
	// Don't cache unknown results
	if serviceName == "未知" {
		return
	}

	c.mu.Lock()
	c.items[command] = serviceName
	c.mu.Unlock()
}

// Has checks if a command is in the cache
func (c *Cache) Has(command string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.items[command]
	return ok
}
