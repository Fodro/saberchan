package s3service

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/Fodro/saberchan/config"
	"github.com/Fodro/saberchan/internal/file"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	nanoid "github.com/matoous/go-nanoid/v2"
)

type service struct {
	bucket       string
	shouldExpire bool
	linkPrefix   string
	expires      time.Duration
	svc          *s3.Client
}

func (s *service) UploadFile(ctx context.Context, f *file.FileReq) (*file.FileResp, error) {
	fileArray := strings.Split(f.Name, ".")
	fileExt := fileArray[len(fileArray)-1]
	id := nanoid.Must(15) + "." + fileExt
	key := id

	var expires *time.Time
	if s.shouldExpire {
		t := time.Now().Add(s.expires)
		expires = &t
	}

	if len(f.Data) == 0 {
		return nil, fmt.Errorf("empty file body")
	}

	put := &s3.PutObjectInput{
		Bucket:             aws.String(s.bucket),
		Key:                aws.String(key),
		Body:               bytes.NewReader(f.Data),
		Expires:            expires,
		CacheControl:       aws.String("max-age=31536000"),
		ContentDisposition: aws.String(fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(f.Name))),
	}
	if f.Type != "" {
		put.ContentType = aws.String(f.Type)
	}
	_, err := s.svc.PutObject(ctx, put)
	if err != nil {
		log.Printf("failed to upload file: %s", err)
		return nil, err
	}

	return &file.FileResp{
		Link: fmt.Sprintf("%s/%s", s.linkPrefix, key),
		Key:  key,
	}, nil
}

func (s *service) DeleteFile(ctx context.Context, key string) error {
	if key == "" {
		return nil
	}
	_, err := s.svc.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Printf("failed to delete file %s: %s", key, err)
		return err
	}
	return nil
}

func NewService(conf *config.Config) file.Service {
	endpoint := ResolveEndpoint(conf.S3.Url, conf.S3.UseSSL)

	awsCfg := aws.Config{
		Region:      conf.S3.Region,
		Credentials: credentials.NewStaticCredentialsProvider(conf.S3.AccessKey, conf.S3.SecretKey, ""),
	}

	svc := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = conf.S3.ForcePathStyle
	})

	return &service{
		bucket:       conf.S3.Bucket,
		shouldExpire: conf.S3.EnableExpriration,
		expires:      conf.S3.FileExpire,
		svc:          svc,
		linkPrefix: ResolveLinkPrefix(
			conf.S3.Bucket,
			conf.S3.Url,
			conf.S3.PublicURL,
			conf.S3.UseSSL,
			conf.S3.ForcePathStyle,
		),
	}
}
