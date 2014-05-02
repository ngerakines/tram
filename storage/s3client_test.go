package storage

import (
	"github.com/ngerakines/tram/util"
	"os"
	"testing"
)

var (
	AWS_KEY    = os.Getenv("AWS_KEY")
	AWS_SECRET = os.Getenv("AWS_SECRET")
)

func TestS3Connect(t *testing.T) {
	if !Integration() {
		t.Skip("skipping test when not in integration mode.")
	}

	um := util.NewUidManager()
	bucket := um.GenerateHex()

	s3Client := NewAmazonS3Client(&AmazonS3ClientConfig{AWS_KEY, AWS_SECRET, "http://localhost:9444/s3"})

	err := s3Client.Put(NewAmazonS3Object("hello_world.txt", bucket, "text/plain"), []byte("hello world"))
	if err != nil {
		t.Errorf("Error creating object: %s", err.Error())
	}

	obj, err := s3Client.Get(bucket, "hello_world.txt")
	if err != nil {
		t.Errorf("Error creating object: %s", err.Error())
	}
	if string(obj.Payload()) != "hello world" {
		t.Errorf("Object content is incorrect: %s", string(obj.Payload()))
	}
}
