package s3service

import "testing"

func TestResolveEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		host   string
		useSSL bool
		want   string
	}{
		{name: "https default", host: "s3.amazonaws.com", useSSL: true, want: "https://s3.amazonaws.com"},
		{name: "http minio", host: "minio:9000", useSSL: false, want: "http://minio:9000"},
		{name: "keeps scheme", host: "http://minio:9000", useSSL: true, want: "http://minio:9000"},
		{name: "empty", host: "", useSSL: true, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ResolveEndpoint(tt.host, tt.useSSL)
			if got != tt.want {
				t.Fatalf("ResolveEndpoint(%q, %v) = %q, want %q", tt.host, tt.useSSL, got, tt.want)
			}
		})
	}
}

func TestResolveLinkPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		bucket         string
		host           string
		publicBase     string
		useSSL         bool
		forcePathStyle bool
		want           string
	}{
		{
			name:           "minio path style with public url",
			bucket:         "saberchan",
			host:           "minio:9000",
			publicBase:     "http://localhost:9000",
			useSSL:         false,
			forcePathStyle: true,
			want:           "http://localhost:9000/saberchan",
		},
		{
			name:           "path style without public url",
			bucket:         "saberchan",
			host:           "minio:9000",
			publicBase:     "",
			useSSL:         false,
			forcePathStyle: true,
			want:           "http://minio:9000/saberchan",
		},
		{
			name:           "virtual hosted https",
			bucket:         "mybucket",
			host:           "s3.eu-central-1.amazonaws.com",
			publicBase:     "",
			useSSL:         true,
			forcePathStyle: false,
			want:           "https://mybucket.s3.eu-central-1.amazonaws.com",
		},
		{
			name:           "public url trims slash",
			bucket:         "b",
			host:           "minio:9000",
			publicBase:     "http://localhost:9000/",
			useSSL:         false,
			forcePathStyle: true,
			want:           "http://localhost:9000/b",
		},
		{
			name:           "public url always path-style even if forcePathStyle false",
			bucket:         "saberchan",
			host:           "garage:3900",
			publicBase:     "https://example.com",
			useSSL:         true,
			forcePathStyle: false,
			want:           "https://example.com/saberchan",
		},
		{
			name:           "public url already includes bucket",
			bucket:         "saberchan",
			host:           "garage:3900",
			publicBase:     "https://example.com/saberchan",
			useSSL:         true,
			forcePathStyle: true,
			want:           "https://example.com/saberchan",
		},
		{
			name:           "public media path is full prefix (no bucket append)",
			bucket:         "saberchan",
			host:           "garage:3900",
			publicBase:     "https://example.com/media",
			useSSL:         true,
			forcePathStyle: true,
			want:           "https://example.com/media",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ResolveLinkPrefix(tt.bucket, tt.host, tt.publicBase, tt.useSSL, tt.forcePathStyle)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestObjectPublicURL(t *testing.T) {
	t.Parallel()
	got := ObjectPublicURL("https://example.com/saberchan", "abc.jpg", "http://old/host/saberchan/abc.jpg")
	if got != "https://example.com/saberchan/abc.jpg" {
		t.Fatalf("got %q", got)
	}
	got = ObjectPublicURL("https://example.com/saberchan", "", "http://stored/link.jpg")
	if got != "http://stored/link.jpg" {
		t.Fatalf("fallback got %q", got)
	}
}
