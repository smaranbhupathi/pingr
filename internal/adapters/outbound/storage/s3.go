// Package storage provides an S3-compatible object storage adapter.
// It works identically with AWS S3, Cloudflare R2, and MinIO — the only
// difference is which endpoint URL and credentials you pass in.
//
// Interview explanation:
//   - S3-compatible means the HTTP API (PutObject, presigned URLs, etc.)
//     is the same across all three providers.
//   - We inject the endpoint at construction time, so switching from MinIO
//     (local dev) to R2 (prod) is a one-line env var change.
package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
)

type s3Store struct {
	presigner *s3.PresignClient
	client    *s3.Client
	bucket    string
	publicURL string // base public URL, e.g. "https://pub-xxx.r2.dev" or "http://localhost:9000/avatars"
}

// Config holds all storage configuration. Every value comes from env vars
// so the same binary runs against MinIO locally and R2 in production.
type Config struct {
	Endpoint        string // "" for real AWS; "http://localhost:9000" for MinIO; "https://<id>.r2.cloudflarestorage.com" for R2
	Region          string // "us-east-1" for MinIO/R2, real region for AWS
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	PublicBaseURL   string // where browsers download uploaded files from
}

func NewS3Store(cfg Config) (outbound.StorageService, error) {
	awsCfg := aws.Config{
		Region: cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		),
	}

	opts := []func(*s3.Options){}

	// If an endpoint is provided (MinIO or R2), override the default AWS endpoint.
	// For real AWS, leave endpoint empty and the SDK resolves it from the region.
	if cfg.Endpoint != "" {
		opts = append(opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			// PathStyleURLEncoding is required for MinIO.
			// R2 and AWS work with virtual-hosted style (default),
			// but path-style is safe for all three.
			o.UsePathStyle = true
		})
	}

	client := s3.NewFromConfig(awsCfg, opts...)
	presigner := s3.NewPresignClient(client)

	return &s3Store{
		presigner: presigner,
		client:    client,
		bucket:    cfg.Bucket,
		publicURL: cfg.PublicBaseURL,
	}, nil
}

// PresignPUT generates a signed PUT URL valid for ttl.
// The browser uses this URL to upload directly — the API server never
// touches the file bytes, which saves bandwidth and compute cost.
func (s *s3Store) PresignPUT(ctx context.Context, key string, ttl time.Duration) (string, error) {
	req, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("presign put: %w", err)
	}
	return req.URL, nil
}

// PublicURL returns the permanent public URL for a stored object.
// This is saved in the DB and returned in profile responses.
func (s *s3Store) PublicURL(key string) string {
	return fmt.Sprintf("%s/%s", s.publicURL, key)
}

// DeleteObject removes an object from the bucket.
// Called when a user replaces their avatar so we don't accumulate stale files.
func (s *s3Store) DeleteObject(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}
