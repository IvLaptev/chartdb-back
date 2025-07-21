package s3client

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

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
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Bucket          string `yaml:"bucket"`
}

type Client interface {
	SaveContent(ctx context.Context, key string, content string) error
	GetContent(ctx context.Context, key string) (string, error)
}

type ClientImpl struct {
	awsClient *s3.Client
	bucket    string
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
