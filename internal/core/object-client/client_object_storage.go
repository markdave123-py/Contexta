package objectclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	cfg "github.com/markdave123-py/Contexta/internal/config"
	"github.com/markdave123-py/Contexta/internal/core"
)

type S3Client struct {
	client *s3.Client
	region string
	bucket string
}

func NewS3Client(ctx context.Context, cfg *cfg.Config) (core.ObjectClient, error) {
	if cfg.AwsAccessKey == "" || cfg.AwsSecretKey == "" {
		return nil, fmt.Errorf("AWS credentials not set")
	}
	if cfg.AwsRegion == "" {
		return nil, fmt.Errorf("AWS_REGION not set")
	}
	if cfg.BucketName == "" {
		return nil, fmt.Errorf("S3 bucket name not set")
	}

	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(cfg.AwsRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AwsAccessKey, cfg.AwsSecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	log.Println("Connected to AWS S3 successfully")

	return &S3Client{
		client: client,
		region: cfg.AwsRegion,
		bucket: cfg.BucketName,
	}, nil
}

// UploadFile uploads a file to S3 and returns the public URL.
func (c *S3Client) UploadFile(ctx context.Context, bucket, key string, data []byte, contentType string) (string, error) {
	uploader := manager.NewUploader(c.client)

	input := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	ctxUpload, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	_, err := uploader.Upload(ctxUpload, input)
	if err != nil {
		return "", fmt.Errorf("s3 upload failed: %w", err)
	}

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, c.region, key)
	return url, nil
}

func (c *S3Client) DeleteFile(ctx context.Context, bucket, key string) error {
	ctxDel, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err := c.client.DeleteObject(ctxDel, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3 delete failed: %w", err)
	}
	return nil
}

func (c *S3Client) GetFile(ctx context.Context, bucket, key string) ([]byte, error) {
	ctxGet, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := c.client.GetObject(ctxGet, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 get failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return body, nil
}


func (c *S3Client) GetObjectReader(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	ctxGet, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	resp, err := c.client.GetObject(ctxGet, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 get failed: %w", err)
	}

	return resp.Body, nil
}

