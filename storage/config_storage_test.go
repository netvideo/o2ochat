package storage

import (
	"os"
	"testing"
)

func TestConfigStorage_BasicOperations(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	configStorage := manager.GetConfigStorage()

	t.Run("StoreConfig", func(t *testing.T) {
		err := configStorage.StoreConfig("app", "theme", "dark")
		if err != nil {
			t.Fatalf("Failed to store config: %v", err)
		}

		err = configStorage.StoreConfig("app", "language", "en")
		if err != nil {
			t.Fatalf("Failed to store config: %v", err)
		}

		err = configStorage.StoreConfig("user", "name", "Alice")
		if err != nil {
			t.Fatalf("Failed to store config: %v", err)
		}
	})

	t.Run("GetConfig", func(t *testing.T) {
		value, err := configStorage.GetConfig("app", "theme", "light")
		if err != nil {
			t.Fatalf("Failed to get config: %v", err)
		}

		if value != "dark" {
			t.Errorf("Expected 'dark', got '%v'", value)
		}
	})

	t.Run("GetConfigNotFound", func(t *testing.T) {
		value, err := configStorage.GetConfig("app", "nonexistent", "default")
		if err != ErrConfigNotFound {
			t.Errorf("Expected ErrConfigNotFound, got %v", err)
		}

		if value != "default" {
			t.Errorf("Expected default value 'default', got '%v'", value)
		}
	})

	t.Run("DeleteConfig", func(t *testing.T) {
		err := configStorage.DeleteConfig("app", "theme")
		if err != nil {
			t.Fatalf("Failed to delete config: %v", err)
		}

		value, err := configStorage.GetConfig("app", "theme", "default")
		if err != ErrConfigNotFound {
			t.Errorf("Expected ErrConfigNotFound after delete, got %v", err)
		}

		if value != "default" {
			t.Errorf("Expected default value after delete")
		}
	})
}

func TestConfigStorage_GetAllConfig(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	configStorage := manager.GetConfigStorage()

	configStorage.StoreConfig("database", "host", "localhost")
	configStorage.StoreConfig("database", "port", 5432)
	configStorage.StoreConfig("database", "name", "o2ochat")

	allConfig, err := configStorage.GetAllConfig("database")
	if err != nil {
		t.Fatalf("Failed to get all config: %v", err)
	}

	if len(allConfig) != 3 {
		t.Errorf("Expected 3 configs, got %d", len(allConfig))
	}

	if allConfig["host"] != "localhost" {
		t.Errorf("Expected host 'localhost', got '%v'", allConfig["host"])
	}

	if allConfig["port"] != float64(5432) {
		t.Errorf("Expected port 5432, got '%v'", allConfig["port"])
	}
}

func TestConfigStorage_ExportImport(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	configStorage := manager.GetConfigStorage()

	configStorage.StoreConfig("export", "key1", "value1")
	configStorage.StoreConfig("export", "key2", "value2")

	tmpFile, err := os.CreateTemp("", "config-export-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	t.Run("ExportConfig", func(t *testing.T) {
		err := configStorage.ExportConfig(tmpPath)
		if err != nil {
			t.Fatalf("Failed to export config: %v", err)
		}

		if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
			t.Fatal("Export file should exist")
		}
	})

	t.Run("ImportConfig", func(t *testing.T) {
		configStorage.ResetConfig()

		allConfig, _ := configStorage.GetAllConfig("export")
		if len(allConfig) != 0 {
			t.Error("Config should be empty after reset")
		}

		err := configStorage.ImportConfig(tmpPath)
		if err != nil {
			t.Fatalf("Failed to import config: %v", err)
		}

		allConfig, err = configStorage.GetAllConfig("export")
		if err != nil {
			t.Fatalf("Failed to get all config: %v", err)
		}

		if len(allConfig) != 2 {
			t.Errorf("Expected 2 configs after import, got %d", len(allConfig))
		}
	})
}

func TestConfigStorage_ResetConfig(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	configStorage := manager.GetConfigStorage()

	configStorage.StoreConfig("reset", "key1", "value1")
	configStorage.StoreConfig("reset", "key2", "value2")
	configStorage.StoreConfig("other", "key1", "value1")

	t.Run("ResetConfig", func(t *testing.T) {
		err := configStorage.ResetConfig()
		if err != nil {
			t.Fatalf("Failed to reset config: %v", err)
		}

		allConfig, err := configStorage.GetAllConfig("reset")
		if err != nil {
			t.Fatalf("Failed to get all config: %v", err)
		}

		if len(allConfig) != 0 {
			t.Errorf("Expected 0 configs after reset, got %d", len(allConfig))
		}
	})
}

func TestConfigStorage_UpdateConfig(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	configStorage := manager.GetConfigStorage()

	configStorage.StoreConfig("update", "version", "1.0.0")

	value, _ := configStorage.GetConfig("update", "version", "")
	if value != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got '%v'", value)
	}

	configStorage.StoreConfig("update", "version", "2.0.0")

	value, _ = configStorage.GetConfig("update", "version", "")
	if value != "2.0.0" {
		t.Errorf("Expected version 2.0.0, got '%v'", value)
	}
}
