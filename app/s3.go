package app

import (
	"errors"
	"github.com/ngerakines/ketama"
	"log"
	"net/http"
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

func (storageManager *S3StorageManager) Store(contentHash string, payload []byte, urls, aliases []string, callback chan CachedFile) {
	bucket := storageManager.bucketRing.Hash(contentHash)

	contentObject, err := storageManager.s3Client.NewObject(contentHash, bucket, "application/octet-stream")
	if err != nil {
		log.Println(err.Error())
		return
	}

	err = storageManager.s3Client.Put(contentObject, payload)
	if err != nil {
		log.Println(err.Error())
		return
	}

	cachedFile := storageManager.newCachedFile(contentHash, urls, aliases, len(payload), bucket)

	callback <- cachedFile
}

func (storageManager *S3StorageManager) Delete(cachedFile CachedFile) error {
	return nil
}

func (storageManager *S3StorageManager) Serve(cachedFile CachedFile, res http.ResponseWriter, req *http.Request) error {
	bucket, hasBucket := cachedFile.Attributes()["bucket"]
	if !hasBucket {
		log.Println("Could not serve file because bucket attribute not set", cachedFile)
		return errors.New("Invalid bucket attribute.")
	}
	err := storageManager.s3Client.Proxy(bucket, cachedFile.ContentHash(), res)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (storageManager *S3StorageManager) newCachedFile(contentHash string, urls, aliases []string, size int, bucket string) CachedFile {
	attributes := make(map[string]string)
	attributes["bucket"] = bucket
	cachedFile := new(simpleCachedFile)
	cachedFile.InternalContentHash = contentHash
	cachedFile.InternalUrls = urls
	cachedFile.InternalAliases = aliases
	cachedFile.InternalSize = size
	cachedFile.InternalAttributes = attributes
	return cachedFile
}

func (cachedFile *S3CachedFile) Size() int {
	return cachedFile.size
}

func (cachedFile *S3CachedFile) ContentHash() string {
	return cachedFile.contentHash
}

func (cachedFile *S3CachedFile) Urls() []string {
	return cachedFile.urls
}

func (cachedFile *S3CachedFile) Aliases() []string {
	return cachedFile.aliases
}

func (cachedFile *S3CachedFile) Serve(res http.ResponseWriter, req *http.Request) {

}
