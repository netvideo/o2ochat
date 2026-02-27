package storage

import (
	"fmt"
	"testing"
	"time"
)

func TestCacheManager_BasicOperations(t *testing.T) {
	cache := NewLRUCacheManager(1024 * 1024)
	defer cache.Close()

	key := "test:key"
	value := []byte("test value")
	ttl := time.Hour

	t.Run("Set and Get", func(t *testing.T) {
		err := cache.Set(key, value, ttl)
		if err != nil {
			t.Fatalf("Failed to set cache: %v", err)
		}

		retrieved, err := cache.Get(key)
		if err != nil {
			t.Fatalf("Failed to get cache: %v", err)
		}

		if string(retrieved) != string(value) {
			t.Errorf("Expected %s, got %s", string(value), string(retrieved))
		}
	})

	t.Run("Get Non-existent", func(t *testing.T) {
		_, err := cache.Get("nonexistent")
		if err != ErrCacheNotFound {
			t.Errorf("Expected ErrCacheNotFound, got %v", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := cache.Delete(key)
		if err != nil {
			t.Fatalf("Failed to delete cache: %v", err)
		}

		_, err = cache.Get(key)
		if err != ErrCacheNotFound {
			t.Errorf("Expected ErrCacheNotFound after delete, got %v", err)
		}
	})
}

func TestCacheManager_LRU(t *testing.T) {
	cache := NewLRUCacheManager(10)

	for i := 0; i < 10; i++ {
		key := string(rune('a' + i))
		value := []byte("0123456789") // 10 bytes each
		cache.Set(key, value, 0)
	}

	stats, _ := cache.GetCacheStats()
	if stats.Evictions == 0 {
		t.Error("Expected some evictions due to size limit")
	}
}

func TestCacheManager_TTL(t *testing.T) {
	cache := NewLRUCacheManager(1024 * 1024)
	defer cache.Close()

	key := "ttl:key"
	value := []byte("ttl value")
	ttl := 100 * time.Millisecond

	err := cache.Set(key, value, ttl)
	if err != nil {
		t.Fatalf("Failed to set cache with TTL: %v", err)
	}

	time.Sleep(150 * time.Millisecond)

	_, err = cache.Get(key)
	if err != ErrCacheNotFound {
		t.Errorf("Expected ErrCacheNotFound after TTL expired, got %v", err)
	}
}

func TestCacheManager_Clear(t *testing.T) {
	cache := NewLRUCacheManager(1024 * 1024)
	defer cache.Close()

	for i := 0; i < 5; i++ {
		key := string(rune('a' + i))
		value := []byte(string(rune('0' + i)))
		cache.Set(key, value, 0)
	}

	stats, _ := cache.GetCacheStats()
	if stats.Items != 5 {
		t.Errorf("Expected 5 items, got %d", stats.Items)
	}

	err := cache.Clear()
	if err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}

	stats, _ = cache.GetCacheStats()
	if stats.Items != 0 {
		t.Errorf("Expected 0 items after clear, got %d", stats.Items)
	}
}

func TestCacheManager_Resize(t *testing.T) {
	cache := NewLRUCacheManager(1024)
	defer cache.Close()

	for i := 0; i < 600; i++ {
		key := fmt.Sprintf("key%d", i)
		value := []byte(fmt.Sprintf("value%d", i))
		cache.Set(key, value, 0)
	}

	stats, _ := cache.GetCacheStats()
	initialItems := stats.Items

	err := cache.Resize(512)
	if err != nil {
		t.Fatalf("Failed to resize cache: %v", err)
	}

	stats, _ = cache.GetCacheStats()
	if stats.Size != 512 {
		t.Errorf("Expected size 512, got %d", stats.Size)
	}

	if stats.Evictions == 0 {
		t.Error("Expected some evictions after resize")
	}

	_ = initialItems
}

func TestCacheManager_ConcurrentAccess(t *testing.T) {
	cache := NewLRUCacheManager(1024 * 1024)
	defer cache.Close()

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				key := string(rune('a' + id))
				value := []byte(string(rune('0' + j)))
				cache.Set(key, value, 0)
				cache.Get(key)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestCacheManager_CleanupExpired(t *testing.T) {
	cache := NewLRUCacheManager(1024 * 1024)
	defer cache.Close()

	for i := 0; i < 5; i++ {
		key := string(rune('a' + i))
		value := []byte(string(rune('0' + i)))
		ttl := time.Duration(50+i*10) * time.Millisecond
		cache.Set(key, value, ttl)
	}

	time.Sleep(200 * time.Millisecond)

	cleaned := cache.CleanupExpired()
	if cleaned == 0 {
		t.Error("Expected some expired items to be cleaned up")
	}

	stats, _ := cache.GetCacheStats()
	if stats.Items != 0 {
		t.Errorf("Expected 0 items after cleanup, got %d", stats.Items)
	}
}
