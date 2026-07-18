package s3service

import (
	"fmt"
	"strings"
)

// ResolveEndpoint returns the AWS SDK endpoint URL for the configured host.
// host may be "minio:9000", "s3.amazonaws.com", or already include a scheme.
func ResolveEndpoint(host string, useSSL bool) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	if strings.Contains(host, "://") {
		return host
	}
	scheme := "https"
	if !useSSL {
		scheme = "http"
	}
	return scheme + "://" + host
}

// ResolveLinkPrefix returns the public base URL used to form object links (without trailing slash).
// publicBase, when set, is preferred (e.g. http://localhost:9000 for browser access to MinIO).
func ResolveLinkPrefix(bucket, host, publicBase string, useSSL, forcePathStyle bool) string {
	base := strings.TrimRight(strings.TrimSpace(publicBase), "/")
	if base == "" {
		base = ResolveEndpoint(host, useSSL)
	}
	base = strings.TrimRight(base, "/")
	if forcePathStyle {
		return fmt.Sprintf("%s/%s", base, bucket)
	}
	// Virtual-hosted–style: https://bucket.host
	trimmed := strings.TrimPrefix(strings.TrimPrefix(base, "https://"), "http://")
	scheme := "https"
	if !useSSL || strings.HasPrefix(base, "http://") {
		scheme = "http"
	}
	if strings.HasPrefix(base, "https://") {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s.%s", scheme, bucket, trimmed)
}
