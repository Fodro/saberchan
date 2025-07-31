package s3service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"time"

	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/Fodro/saberchan/config"
	"github.com/Fodro/saberchan/internal/file"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type service struct {
	bucket       string
	shouldExpire bool
	linkPrefix   string
	expires      time.Duration
	svc          *s3.S3
}

func (s *service) UploadFile(ctx context.Context, f *file.FileReq) (*file.FileResp, error) {
	id := nanoid.Must(5)
	key := id+"-"+url.QueryEscape(f.Name)

	var expires *time.Time
	if s.shouldExpire {
		expires = aws.Time(time.Now().Add(s.expires))
	} else {
		expires = nil
	}

	body, err := base64.StdEncoding.DecodeString(f.Body)
	if err != nil {
		log.Printf("failed to decode body: %s", err)
		return nil, err
	}

	_, err = s.svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:  aws.String(s.bucket),
		Key:     aws.String(key),
		Body:    bytes.NewReader(body),
		Expires: expires,
	})
	if err != nil {
		log.Printf("failed to upload file: %s", err)
		return nil, err
	}

	return &file.FileResp{
		Link: fmt.Sprintf("%s/%s", s.linkPrefix, key),
	}, nil
}

func NewService(conf *config.Config) file.Service {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(conf.S3.AccessKey, conf.S3.SecretKey, ""),
		Endpoint:    aws.String("https://" + conf.S3.Url),
		Region:      aws.String(conf.S3.Region),
	})
	if err != nil {
		log.Fatalf("failed to connect to s3: %s", err)
	}
	svc := s3.New(sess)

	return &service{
		bucket:       conf.S3.Bucket,
		shouldExpire: conf.S3.EnableExpriration,
		expires:      conf.S3.FileExpire,
		svc:          svc,
		linkPrefix:   fmt.Sprintf("https://%s.%s", conf.S3.Bucket, conf.S3.Url),
	}
}
