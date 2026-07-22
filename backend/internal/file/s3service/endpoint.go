package s3service

import (
	"fmt"
	"net/url"
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

// publicBaseHasPath reports whether publicBase includes a URL path other than "/".
// e.g. https://example.com/media → true; https://example.com → false.
func publicBaseHasPath(publicBase string) bool {
	u, err := url.Parse(publicBase)
	if err != nil || u.Scheme == "" || u.Host == "" {
		// Non-URL string with a slash after the first segment — treat as pathful.
		rest := publicBase
		if i := strings.Index(publicBase, "://"); i >= 0 {
			rest = publicBase[i+3:]
		}
		if j := strings.Index(rest, "/"); j >= 0 {
			return strings.Trim(rest[j:], "/") != ""
		}
		return false
	}
	return strings.Trim(u.Path, "/") != ""
}

// ResolveLinkPrefix returns the public base URL used to form object links
// (without trailing slash). Links are always "{prefix}/{objectKey}".
//
// When publicBase is set:
//   - Origin only (https://example.com) → "{public}/{bucket}" (MinIO-style).
//   - Origin + path (https://example.com/media) → publicBase as-is; nginx
//     should map that path onto Garage's /{bucket}/{key} (strip /media, add bucket).
//   - Already ends with /{bucket} → returned as-is.
//
// When publicBase is empty, derive from the SDK host (path-style or virtual-hosted).
func ResolveLinkPrefix(bucket, host, publicBase string, useSSL, forcePathStyle bool) string {
	public := strings.TrimRight(strings.TrimSpace(publicBase), "/")
	if public != "" {
		if bucket != "" && (strings.HasSuffix(public, "/"+bucket) || public == bucket) {
			return public
		}
		if bucket == "" || publicBaseHasPath(public) {
			return public
		}
		return fmt.Sprintf("%s/%s", public, bucket)
	}

	base := strings.TrimRight(ResolveEndpoint(host, useSSL), "/")
	if base == "" {
		return ""
	}
	if forcePathStyle {
		return fmt.Sprintf("%s/%s", base, bucket)
	}
	// Virtual-hosted–style against the SDK host (typical AWS).
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

// ObjectPublicURL builds the browser URL for an object.
// Prefers key + linkPrefix so changing S3_PUBLIC_URL rewrites media without
// a DB migration; falls back to the stored link when key is empty.
func ObjectPublicURL(linkPrefix, key, storedLink string) string {
	key = strings.TrimSpace(key)
	prefix := strings.TrimRight(strings.TrimSpace(linkPrefix), "/")
	if key != "" && prefix != "" {
		return prefix + "/" + key
	}
	return strings.TrimSpace(storedLink)
}
