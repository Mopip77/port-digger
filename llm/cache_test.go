package llm

import (
	"testing"
)

func TestNewCache(t *testing.T) {
	cache := NewCache()
	if cache == nil {
		t.Fatal("NewCache() returned nil")
	}
	if cache.items == nil {
		t.Fatal("NewCache() items map is nil")
	}
}

func TestCache_GetSet(t *testing.T) {
	cache := NewCache()

	// Test get on empty cache
	if got := cache.Get("some-command"); got != "" {
		t.Errorf("Get() on empty cache = %v, want empty string", got)
	}

	// Test set and get
	cache.Set("node /path/to/app.js", "my-app")
	if got := cache.Get("node /path/to/app.js"); got != "my-app" {
		t.Errorf("Get() = %v, want my-app", got)
	}

	// Test Has
	if !cache.Has("node /path/to/app.js") {
		t.Error("Has() returned false for existing key")
	}
	if cache.Has("non-existent") {
		t.Error("Has() returned true for non-existent key")
	}
}

func TestCache_SkipsUnknown(t *testing.T) {
	cache := NewCache()

	// Set "未知" should be a no-op
	cache.Set("some-command", "未知")

	if cache.Has("some-command") {
		t.Error("Cache should not store '未知' results")
	}

	if got := cache.Get("some-command"); got != "" {
		t.Errorf("Get() after setting '未知' = %v, want empty string", got)
	}
}

func TestCache_ThreadSafe(t *testing.T) {
	cache := NewCache()
	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 100; i++ {
		go func(n int) {
			cache.Set("cmd", "service")
			cache.Get("cmd")
			cache.Has("cmd")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}
