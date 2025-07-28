package s3client

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var (
	ErrContentNotFound = errors.New("content not found")
)

type S3Config struct {
	Region          string `yaml:"region"`
	URL             string `yaml:"url"`
	AccessKeyID     string `yaml:"access_key_id" env:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `yaml:"secret_access_key" env:"AWS_SECRET_ACCESS_KEY"`
	Bucket          string `yaml:"bucket"`
}

type Client interface {
	SaveContent(ctx context.Context, key string, content string) error
	GetContent(ctx context.Context, key string) (string, error)

	ListObjects(ctx context.Context, nextPageToken *string) (*ObjectList, error)
	BatchDeleteObjects(ctx context.Context, keys []string) error
}

type ClientImpl struct {
	awsClient *s3.Client
	bucket    string
}

type Object struct {
	Key          string    `json:"key"`
	LastModified time.Time `json:"last_modified"`
}

type ObjectList struct {
	Objects       []*Object `json:"objects"`
	NextPageToken *string   `json:"next_page_token"`
}

func (c *ClientImpl) ListObjects(ctx context.Context, nextPageToken *string) (*ObjectList, error) {
	objects, err := c.awsClient.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:            aws.String(c.bucket),
		ContinuationToken: nextPageToken,
	})
	if err != nil {
		return nil, fmt.Errorf("list objects v2: %w", err)
	}

	resultObjects := make([]*Object, 0, len(objects.Contents))
	for _, object := range objects.Contents {
		if object.Key != nil && object.LastModified != nil {
			resultObjects = append(resultObjects, &Object{
				Key:          *object.Key,
				LastModified: *object.LastModified,
			})
		}
	}

	return &ObjectList{
		Objects:       resultObjects,
		NextPageToken: objects.NextContinuationToken,
	}, nil
}

func (c *ClientImpl) SaveContent(ctx context.Context, key string, content string) error {
	body := []byte(content)

	h := sha256.New()
	h.Write(body)
	hash := base64.URLEncoding.EncodeToString(h.Sum(nil))

	_, err := c.awsClient.PutObject(ctx, &s3.PutObjectInput{
		Bucket:            aws.String(c.bucket),
		Key:               aws.String(key),
		Body:              bytes.NewReader(body),
		ChecksumAlgorithm: types.ChecksumAlgorithmSha256,
		ChecksumSHA256:    &hash,
	})
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}

	return nil
}

func (c *ClientImpl) GetContent(ctx context.Context, key string) (string, error) {
	object, err := c.awsClient.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFoundErr *types.NoSuchKey
		if errors.As(err, &notFoundErr) {
			return "", ErrContentNotFound
		}

		return "", fmt.Errorf("get object: %w", err)
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(object.Body)
	if err != nil {
		return "", fmt.Errorf("read from: %w", err)
	}

	return buf.String(), nil
}

func (c *ClientImpl) BatchDeleteObjects(ctx context.Context, keys []string) error {
	objects := make([]types.ObjectIdentifier, 0, len(keys))
	for _, key := range keys {
		objects = append(objects, types.ObjectIdentifier{Key: aws.String(key)})
	}

	_, err := c.awsClient.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(c.bucket),
		Delete: &types.Delete{
			Objects: objects,
		},
	})
	if err != nil {
		return fmt.Errorf("delete objects: %w", err)
	}

	return nil
}

func NewS3Client(ctx context.Context, config S3Config) (*ClientImpl, error) {
	awsCfg, err := aconfig.LoadDefaultConfig(
		ctx,
		aconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.SecretAccessKey, ""),
		),
		aconfig.WithRegion(config.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("load default confug: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(config.URL)
	})

	return &ClientImpl{
		awsClient: client,
		bucket:    config.Bucket,
	}, nil
}
