package app

import (
	"encoding/json"
	"github.com/ngerakines/ketama"
	"log"
)

type S3StorageManager struct {
	bucketRing ketama.HashRing
	s3Client   S3Client
}

type S3CachedFile struct {
	contentHash string
	remoteUrl   string
	bucket      string
	urls        []string
	aliases     []string
	size        int
}

func NewS3StorageManager(buckets []string, s3Client S3Client) StorageManager {
	hashRing := ketama.NewRing(180)
	for _, bucket := range buckets {
		hashRing.Add(bucket, 1)
	}
	hashRing.Bake()
	return &S3StorageManager{hashRing, s3Client}
}

func (sm *S3StorageManager) Store(payload []byte, sourceUrl string, contentHash string, aliases []string, callback chan CachedFile) {
	bucket := sm.bucketRing.Hash(contentHash)

	contentObject, err := sm.s3Client.NewContentObject(contentHash, bucket, "application/octet-stream")
	if err != nil {
		log.Println(err.Error())
		return
	}

	err = sm.s3Client.Put(contentObject, payload)
	if err != nil {
		log.Println(err.Error())
		return
	}

	cachedFile := &S3CachedFile{contentHash, contentObject.Url(), bucket, []string{sourceUrl}, aliases, len(payload)}

	callback <- cachedFile
}

func (sm *S3StorageManager) Delete(cachedFile CachedFile) error {
	return nil
}

func (cf *S3CachedFile) LocationType() string {
	return CachedFile_Remote
}

func (cf *S3CachedFile) Location() string {
	return cf.remoteUrl
}

func (cf *S3CachedFile) Size() int {
	return cf.size
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
