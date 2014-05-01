package storage

import (
	"os"
	"strings"
	"testing"
)

func TestLocalFile(t *testing.T) {
	directoryManager := newDirectoryManager()
	defer directoryManager.Close()

	sm := NewLocalStorageManager(directoryManager.path)
	callback := make(chan CachedFile)
	go sm.Store([]byte("/"), "http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8", "42099b4af021e53fd8fd4e056c2568d7c2e3ffa8", []string{"http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8", "e69be504a5157ba8cab8e4633955d50c6a7118b9", "42099b4af021e53fd8fd4e056c2568d7c2e3ffa8", "home"}, callback)
	cachedFile := <-callback
	if cachedFile.ContentHash() != "42099b4af021e53fd8fd4e056c2568d7c2e3ffa8" {
		t.Errorf("cached file content hash invalid: %s", cachedFile.ContentHash())
	}
	if !stringArrayEquals(cachedFile.Urls(), []string{"http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8"}) {
		t.Errorf("urls not the same: %s %s", cachedFile.Urls(), []string{"http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8"})
	}
	if cachedFile.LocationType() != CachedFile_Local {
		t.Errorf("invalid location type: %s", cachedFile.LocationType())
	}
	if !strings.HasSuffix(cachedFile.Location(), "42099b4af021e53fd8fd4e056c2568d7c2e3ffa8") {
		t.Errorf("invalid location: %s", cachedFile.Location())
	}
	testAliases := []string{"http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8", "e69be504a5157ba8cab8e4633955d50c6a7118b9", "42099b4af021e53fd8fd4e056c2568d7c2e3ffa8", "home"}
	if !stringArrayEquals(cachedFile.Aliases(), testAliases) {
		t.Errorf("aliases not the same: %s %s", cachedFile.Aliases(), testAliases)
	}
	if cachedFile.Size() != 1 {
		t.Errorf("cached file size invalid: %d", cachedFile.Size())
	}

	path := cachedFile.Location()
	err := cachedFile.Delete()
	if err != nil {
		t.Error(err.Error())
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Errorf("stat error is not isNotExist %s", statErr.Error())
	}

}
