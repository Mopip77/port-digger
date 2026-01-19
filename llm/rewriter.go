package llm

import (
	"sync"
)

// Rewriter orchestrates LLM-based process name rewriting with caching
type Rewriter struct {
	config  *Config
	client  *Client
	cache   *Cache
	pending sync.Map // tracks in-flight requests to avoid duplicates
}

// NewRewriter creates a new rewriter instance
func NewRewriter() (*Rewriter, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	cache, err := LoadCache()
	if err != nil {
		return nil, err
	}

	var client *Client
	if config.LLM.Enabled {
		client = NewClient(&config.LLM)
	}

	return &Rewriter{
		config: config,
		client: client,
		cache:  cache,
	}, nil
}

// IsEnabled returns whether LLM rewriting is enabled
func (r *Rewriter) IsEnabled() bool {
	return r.config != nil && r.config.LLM.Enabled && r.client != nil
}

// GetServiceName returns the cached service name for a command
// Returns empty string if not cached or if LLM is disabled
func (r *Rewriter) GetServiceName(command string) string {
	if !r.IsEnabled() {
		return ""
	}
	return r.cache.Get(command)
}

// TriggerRewrite starts an async background rewrite for the given command
// If the command is already cached or a request is in-flight, this is a no-op
func (r *Rewriter) TriggerRewrite(command string) {
	if !r.IsEnabled() {
		return
	}

	// Already cached
	if r.cache.Has(command) {
		return
	}

	// Check if already in-flight
	if _, loaded := r.pending.LoadOrStore(command, true); loaded {
		return
	}

	// Start async rewrite
	go func() {
		defer r.pending.Delete(command)

		serviceName, err := r.client.RewriteProcessName(command)
		if err != nil {
			// Log error but don't fail
			println("LLM rewrite error:", err.Error())
			return
		}

		// Cache the result (cache.Set handles skipping "未知")
		r.cache.Set(command, serviceName)

		// Persist cache
		if err := r.cache.Save(); err != nil {
			println("Failed to save cache:", err.Error())
		}
	}()
}
