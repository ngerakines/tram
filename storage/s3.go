package storage

import (
	"encoding/json"
	"fmt"
)

func (sm *S3StorageManager) Load() {
}

func (sm *S3StorageManager) Store(payload []byte, sourceUrl string, contentHash string, aliases []string, callback chan CachedFile) {

	remoteUrl, bucket, remoteUrlErr := sm.buildRemoteUrl(contentHash)
	if remoteUrlErr != nil {
		fmt.Println(remoteUrlErr.Error())
		return
	}

	cachedFile := &S3CachedFile{contentHash, remoteUrl, bucket, []string{sourceUrl}, aliases}

	err1 := sm.storePayload(cachedFile, payload)
	if err1 != nil {
		fmt.Println(err1.Error())
		return
	}
	err2 := sm.storeMetadata(cachedFile)
	if err1 != nil {
		fmt.Println(err2.Error())
		return
	}

	callback <- cachedFile
}

func (sm *S3StorageManager) buildRemoteUrl(contentHash string) (string, string, error) {
	return "", "", StorageError{"S3StorageManager.buildRemoteUrl(...) method not implemented yet."}
}

func (sm *S3StorageManager) storePayload(cachedFile *S3CachedFile, payload []byte) error {
	return StorageError{"S3StorageManager.storePayload(...) method not implemented yet."}
}

func (sm *S3StorageManager) storeMetadata(cachedFile *S3CachedFile) error {
	return StorageError{"S3StorageManager.storeMetadata(...) method not implemented yet."}
}

func (cf *S3CachedFile) LocationType() string {
	return CachedFile_Remote
}

func (cf *S3CachedFile) Location() string {
	return cf.remoteUrl
}

func (cf *S3CachedFile) Size() int {
	return 0
}

func (cf *S3CachedFile) Delete() error {
	return StorageError{"S3CachedFile.Delete() not implemented yet."}
}

func (cf *S3CachedFile) ContentHash() string {
	return cf.contentHash
}

func (cf *S3CachedFile) Urls() []string {
	return cf.urls
}

func (cf *S3CachedFile) Aliases() []string {
	return cf.aliases
}

func (cf *S3CachedFile) Serialize() ([]byte, error) {
	data, err := json.Marshal(cf)
	if err != nil {
		return nil, err
	}
	return data, nil
}
