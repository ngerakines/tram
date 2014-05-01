package storage

import (
	"github.com/ngerakines/tram/util"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type directoryManager struct {
	path string
}

func (dm *directoryManager) Close() {
	log.Println("Removing temp path", dm.path)
	os.RemoveAll(dm.path)
}

func newDirectoryManager() *directoryManager {
	um := util.NewUidManager()
	pwd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	cacheDirectory := filepath.Join(pwd, um.GenerateHex())
	os.MkdirAll(cacheDirectory, 00777)
	return &directoryManager{cacheDirectory}
}

type mockDownloader struct {
	payloads map[string][]byte
	counts   map[string]int
	mu       sync.Mutex
}

type mockStorageManagerPayload struct {
	payload     []byte
	sourceUrl   string
	contentHash string
	aliases     []string
	callback    chan CachedFile
}

type mockStorageManager struct {
	storePayload []mockStorageManagerPayload
}

func (sm *mockStorageManager) Store(payload []byte, sourceUrl string, contentHash string, aliases []string, callback chan CachedFile) {
	sm.storePayload = append(sm.storePayload, mockStorageManagerPayload{payload, sourceUrl, contentHash, aliases, callback})
}

func (sm *mockStorageManager) Load(callback chan CachedFile) {

}

func (sm *mockStorageManager) Delete(cachedFile CachedFile) error {
	return nil
}

type StringError struct {
	message string
}

func (err StringError) Error() string {
	return err.message
}

func (md *mockDownloader) download(url string) ([]byte, error) {
	payload, hasPayload := md.payloads[url]
	if hasPayload {
		md.mu.Lock()
		value, hasValue := md.counts[url]
		if !hasValue {
			value = 0
		}
		value += 1
		md.counts[url] = value
		md.mu.Unlock()
		return payload, nil
	}
	return nil, StringError{"No url in mock downloader."}
}

func (md *mockDownloader) dedupedDownload() util.RemoteFileFetcher {
	return util.DedupeWrapDownloader(md.download)
}

func buildMockDownlader() *mockDownloader {
	md := new(mockDownloader)
	md.mu = sync.Mutex{}
	md.counts = make(map[string]int)
	md.payloads = make(map[string][]byte)
	md.payloads["http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8"] = []byte("/")
	md.payloads["http://localhost:3001/ef090dcea7b507772498cd2e67f2b148ae2609f6"] = []byte("/tram")
	md.payloads["http://localhost:3001/a11f846da74df08c2e93ede56beefdde735ccc05"] = []byte("/tram-chef-cookbook")
	md.payloads["http://localhost:3001/974a0682b121adda8d8bb4503d07672c6d65319c"] = []byte("/commitment")
	md.payloads["http://localhost:3001/f6cd811a6b41ae88b1aa5d0e7ab8e16cda349917"] = []byte("/erlang_twitter")
	md.payloads["http://localhost:3001/72a57722a7d426ca21bf52348f3a83c96a3cc72b"] = []byte("/erlang_protobuffs")
	md.payloads["http://localhost:3001/fca08fb51c04e2cf57d2d78e229a28ff10a2c2d7"] = []byte("/erlang_couchdb")
	md.payloads["http://localhost:3001/db28cadbe22a8dcbd65659c833ea91b8ad1246e5"] = []byte("/etap")
	md.payloads["http://localhost:3001/81720de9afb950d0e5f04d01c9e99e03e6425af0"] = []byte("/wallbase-cc-downloader")
	md.payloads["http://localhost:3001/6ce8a69121f4fdcb156772ff00c3828ae542f00b"] = []byte("/miniature-dubstep")
	md.payloads["http://localhost:3001/ffdefcd0d443d73f4058e33bec27c5cabb2ac1c1"] = []byte("/elasticservices")
	return md
}

func buildMockStorageManager() *mockStorageManager {
	return &mockStorageManager{make([]mockStorageManagerPayload, 0, 0)}
}

func stringArrayEquals(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
