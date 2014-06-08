package app

import (
	"os"
	"strings"
	"testing"
)

type mockS3Client struct {
	putObjects []*putS3Object
}

type putS3Object struct {
	s3Object S3Object
	payload  []byte
}

func (mc *mockS3Client) Put(s3object S3Object, content []byte) error {
	mc.putObjects = append(mc.putObjects, &putS3Object{s3object, content})
	return nil
}

func (mc *mockS3Client) Get(bucket, file string) (S3Object, error) {
	for _, putObject := range mc.putObjects {
		if putObject.s3Object.Bucket() == bucket && putObject.s3Object.FileName() == file {
			return putObject.s3Object, nil
		}
	}
	return nil, StorageError{"No file with that url exists"}
}

func (mc *mockS3Client) Delete(bucket, file string) error {
	return nil
}

func (mc *mockS3Client) NewContentObject(name, bucket, contentType string) (S3Object, error) {
	return NewAmazonS3Object("/content/"+name, bucket, contentType), nil
}

func (mc *mockS3Client) NewMetadataObject(name, bucket, contentType string) (S3Object, error) {
	return NewAmazonS3Object("/meta/"+name, bucket, contentType), nil
}

func newMockS3Client() *mockS3Client {
	mc := new(mockS3Client)
	mc.putObjects = make([]*putS3Object, 0, 0)
	return mc
}

func TestS3File(t *testing.T) {
	directoryManager := newDirectoryManager()
	defer directoryManager.Close()

	s3Client := newMockS3Client()

	sm := NewS3StorageManager([]string{"test"}, s3Client)
	callback := make(chan CachedFile)
	go sm.Store([]byte("/"), "http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8", "42099b4af021e53fd8fd4e056c2568d7c2e3ffa8", []string{"http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8", "e69be504a5157ba8cab8e4633955d50c6a7118b9", "42099b4af021e53fd8fd4e056c2568d7c2e3ffa8", "home"}, callback)
	cachedFile := <-callback
	if cachedFile.ContentHash() != "42099b4af021e53fd8fd4e056c2568d7c2e3ffa8" {
		t.Errorf("cached file content hash invalid: %s", cachedFile.ContentHash())
	}
	if !stringArrayEquals(cachedFile.Urls(), []string{"http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8"}) {
		t.Errorf("urls not the same: %s %s", cachedFile.Urls(), []string{"http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8"})
	}
	if cachedFile.LocationType() != CachedFile_Remote {
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
	err := sm.Delete(cachedFile)
	if err != nil {
		t.Error(err.Error())
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Errorf("stat error is not isNotExist %s", statErr.Error())
	}

}
