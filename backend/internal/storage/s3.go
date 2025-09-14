package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/burndler/burndler/internal/config"
)

// S3Storage implements Storage interface using AWS S3 or S3-compatible storage
type S3Storage struct {
	client     *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	bucket     string
	pathPrefix string
}

// NewS3Storage creates a new S3 storage instance
func NewS3Storage(cfg *config.Config) (*S3Storage, error) {
	awsConfig := &aws.Config{
		Endpoint:         aws.String(cfg.S3Endpoint),
		Region:           aws.String(cfg.S3Region),
		DisableSSL:       aws.Bool(!cfg.S3UseSSL),
		S3ForcePathStyle: aws.Bool(true),
	}

	if cfg.S3AccessKeyID != "" && cfg.S3SecretAccessKey != "" {
		awsConfig.Credentials = credentials.NewStaticCredentials(
			cfg.S3AccessKeyID,
			cfg.S3SecretAccessKey,
			"",
		)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	client := s3.New(sess)

	return &S3Storage{
		client:     client,
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
		bucket:     cfg.S3Bucket,
		pathPrefix: cfg.S3PathPrefix,
	}, nil
}

func (s *S3Storage) getFullKey(key string) string {
	return s.pathPrefix + key
}

func (s *S3Storage) Upload(ctx context.Context, key string, reader io.Reader, size int64) (string, error) {
	fullKey := s.getFullKey(key)

	input := &s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
		Body:   reader,
	}

	result, err := s.uploader.UploadWithContext(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return result.Location, nil
}

func (s *S3Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	fullKey := s.getFullKey(key)

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	}

	result, err := s.client.GetObjectWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}

	return result.Body, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	fullKey := s.getFullKey(key)

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	}

	_, err := s.client.DeleteObjectWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

func (s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := s.getFullKey(key)

	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	}

	_, err := s.client.HeadObjectWithContext(ctx, input)
	if err != nil {
		// Check if the error is a not found error
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeNoSuchKey {
			return false, nil
		}
		return false, fmt.Errorf("failed to check S3 object existence: %w", err)
	}

	return true, nil
}

func (s *S3Storage) List(ctx context.Context, prefix string) ([]FileInfo, error) {
	fullPrefix := s.getFullKey(prefix)

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(fullPrefix),
	}

	var files []FileInfo

	err := s.client.ListObjectsV2PagesWithContext(ctx, input,
		func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, obj := range page.Contents {
				files = append(files, FileInfo{
					Key:          *obj.Key,
					Size:         *obj.Size,
					LastModified: *obj.LastModified,
				})
			}
			return !lastPage
		})

	if err != nil {
		return nil, fmt.Errorf("failed to list S3 objects: %w", err)
	}

	return files, nil
}

func (s *S3Storage) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	fullKey := s.getFullKey(key)

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	}

	req, _ := s.client.GetObjectRequest(input)
	url, err := req.Presign(expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url, nil
}