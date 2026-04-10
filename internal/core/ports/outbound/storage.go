package outbound

import (
	"context"
	"time"
)

// StorageService is the port the core uses to talk to object storage.
// The implementation can be S3, Cloudflare R2, or MinIO — the core
// doesn't care, it only calls these two methods.
type StorageService interface {
	// PresignPUT returns a time-limited signed URL that the browser can use
	// to PUT a file directly to the bucket without going through the API.
	// key is the object path inside the bucket (e.g. "avatars/user-123.jpg").
	// ttl is how long the URL is valid (e.g. 5 minutes).
	PresignPUT(ctx context.Context, key string, ttl time.Duration) (uploadURL string, err error)

	// PublicURL returns the permanent public URL for an already-uploaded object.
	// This is stored in the database and returned in API responses.
	PublicURL(key string) string

	// DeleteObject removes an object from the bucket (used when user replaces avatar).
	DeleteObject(ctx context.Context, key string) error
}
