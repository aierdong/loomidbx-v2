package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// FileState describes whether an ordinary config file was available on disk.
type FileState string

const (
	// FileStateMissing means the config file does not exist and callers may continue with defaults.
	FileStateMissing FileState = "missing"

	// FileStatePresent means the config file exists and was attempted for parsing.
	FileStatePresent FileState = "present"
)

// ConfigFileStore reads and writes the ordinary JSON user config file.
type ConfigFileStore interface {
	// Read loads a UserConfig from path and reports whether the file was present.
	//
	// A missing file returns a zero UserConfig, FileStateMissing, and nil error so the later loader can
	// continue with default configuration. Existing files with unreadable or invalid JSON content return
	// FileStatePresent and a non-nil error.
	Read(path string) (UserConfig, FileState, error)

	// Save persists the user-editable config fields to path.
	//
	// Save creates the parent directory when needed and writes through a temporary file in the same
	// directory before replacing the target. If writing or replacing fails, the error is returned and any
	// temporary file is removed without intentionally modifying an existing target file.
	Save(path string, config UserConfig) error
}

// JSONConfigFileStore implements ConfigFileStore for ordinary JSON config files.
type JSONConfigFileStore struct {
	// replace optionally overrides the final file replacement step for package-level tests.
	replace func(oldPath string, newPath string) error
}

// Read loads and parses a JSON UserConfig from path.
//
// If path does not exist, Read returns FileStateMissing with no error. If the file exists but cannot be
// read or parsed as UserConfig JSON, Read returns FileStatePresent with the underlying failure wrapped.
func (JSONConfigFileStore) Read(path string) (UserConfig, FileState, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return UserConfig{}, FileStateMissing, nil
		}
		return UserConfig{}, FileStatePresent, fmt.Errorf("read config file: %w", err)
	}

	var config UserConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		return UserConfig{}, FileStatePresent, fmt.Errorf("parse config file: %w", err)
	}

	return config, FileStatePresent, nil
}

// Save writes config as formatted JSON using a temporary file followed by target replacement.
//
// The persisted shape is UserConfig only, so resolved config-file paths, environment override sources,
// and other non-persistent loader state are not written into the ordinary config file.
func (store JSONConfigFileStore) Save(path string, config UserConfig) error {
	parent := filepath.Dir(path)
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return fmt.Errorf("create config parent directory: %w", err)
	}

	raw, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config file: %w", err)
	}
	raw = append(raw, '\n')

	temp, err := os.CreateTemp(parent, ".loomidbx-config-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary config file: %w", err)
	}
	tempName := temp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(tempName)
		}
	}()

	if _, err := temp.Write(raw); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write temporary config file: %w", err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("sync temporary config file: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temporary config file: %w", err)
	}

	if err := store.replaceFile(tempName, path); err != nil {
		return fmt.Errorf("replace config file: %w", err)
	}
	committed = true

	return nil
}

func (store JSONConfigFileStore) replaceFile(oldPath string, newPath string) error {
	if store.replace != nil {
		return store.replace(oldPath, newPath)
	}
	return os.Rename(oldPath, newPath)
}
