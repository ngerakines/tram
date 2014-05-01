package storage

type S3Client interface {
	Put(s3object S3Object, content []byte) error
	Get(url string) (S3Object, error)
	Delete(url string) error
	NewContentObject(name, bucket string) (S3Object, error)
	NewMetadataObject(name, bucket string) (S3Object, error)
}

type S3Object interface {
	FileName() string
	Bucket() string
	Url() string
}

type AmazonS3Client struct {
}

type AmazoneS3Object struct {
	name, bucket string
}

func NewAmazonS3Client() S3Client {
	return &AmazonS3Client{}
}

func NewAmazonS3Object(name, bucket string) S3Object {
	return &AmazoneS3Object{name, bucket}
}

func (client *AmazonS3Client) Put(s3object S3Object, content []byte) error {
	return nil
}

func (client *AmazonS3Client) Get(url string) (S3Object, error) {
	return nil, nil
}

func (client *AmazonS3Client) Delete(url string) error {
	return nil
}

func (client *AmazonS3Client) NewContentObject(name, bucket string) (S3Object, error) {
	return NewAmazonS3Object("/content/"+name, bucket), nil
}

func (client *AmazonS3Client) NewMetadataObject(name, bucket string) (S3Object, error) {
	return NewAmazonS3Object("/meta/"+name, bucket), nil
}

func (s3obj *AmazoneS3Object) FileName() string {
	return s3obj.name
}

func (s3obj *AmazoneS3Object) Bucket() string {
	return s3obj.bucket
}

func (s3obj *AmazoneS3Object) Url() string {
	return "http://s3.amazonaws.com/" + s3obj.name
}
