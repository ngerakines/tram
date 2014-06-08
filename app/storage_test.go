package app

import (
	"bytes"
	"testing"
)

func TestDownload(t *testing.T) {
	md := buildMockDownlader()
	sm := buildMockStorageManager()
	Download(md.dedupedDownload(), sm, "http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8", []string{"home"}, make(chan CachedFile))
	if len(sm.storePayload) != 1 {
		t.Error("storage manager request not parsed.")
	}
	mockStorageManagerPayload1 := sm.storePayload[0]
	if mockStorageManagerPayload1.sourceUrl != "http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8" {
		t.Errorf("source url not the same: %s", mockStorageManagerPayload1.sourceUrl)
	}
	if mockStorageManagerPayload1.contentHash != "42099b4af021e53fd8fd4e056c2568d7c2e3ffa8" {
		t.Errorf("content hash not the same: %s", mockStorageManagerPayload1.contentHash)
	}
	if !bytes.Equal(mockStorageManagerPayload1.payload, []byte("/")) {
		t.Errorf("payload not the same: %s", mockStorageManagerPayload1.payload)
	}
	testAliases := []string{"http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8", "e69be504a5157ba8cab8e4633955d50c6a7118b9", "42099b4af021e53fd8fd4e056c2568d7c2e3ffa8", "home"}
	if !stringArrayEquals(mockStorageManagerPayload1.aliases, testAliases) {
		t.Errorf("aliases not the same: %s %s", mockStorageManagerPayload1.aliases, testAliases)
	}
}

func TestDownloadBadUrl(t *testing.T) {
	md := buildMockDownlader()
	sm := buildMockStorageManager()
	Download(md.dedupedDownload(), sm, "http://localhost:3001/d483f1940a34cb5e8d50c778cd186622eea8268a", []string{"invalid-url"}, make(chan CachedFile))
	if len(sm.storePayload) != 0 {
		t.Error("storage manager was called when it shouldn't have been.")
	}
}
