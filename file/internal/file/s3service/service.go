package s3service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/Fodro/saberchan/file/config"
	"github.com/Fodro/saberchan/file/internal/database"
	"github.com/Fodro/saberchan/file/internal/file"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

type service struct {
	bucket       string
	shouldExpire bool
	linkPrefix   string
	expires      time.Duration
	svc          *s3.S3

	repo database.Repository
}

func (s *service) ClearFilesForPost(ctx context.Context, postID uuid.UUID) error {
	files, err := s.repo.GetFilesForPost(postID)
	if err != nil {
		return err
	}
	for _, f := range files {
		_, err := s.svc.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(f.Key),
		})
		if err != nil {
			log.Printf("failed to delete file %s: %s", f.Key, err)
		}
	}

	return s.repo.DeleteFilesForPost(postID)
}

func (s *service) UploadFile(ctx context.Context, f *file.FileReq) (*file.FileResp, error) {
	id := uuid.New()
	key := id.String() + "-" + f.Name

	var expires *time.Time
	if s.shouldExpire {
		expires = aws.Time(time.Now().Add(s.expires))
	} else {
		expires = nil
	}

	body, err := base64.URLEncoding.DecodeString(f.Body)
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

	err = s.repo.AddFile(&database.File{
		ID:     id,
		PostID: f.PostID,
		Key:    key,
	})

	if err != nil {
		return nil, err
	}

	return &file.FileResp{
		Link: fmt.Sprintf("%s/%s", s.linkPrefix, key),
	}, nil
}

func NewService(conf *config.Config, repo database.Repository) file.Service {
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
		repo:         repo,
	}
}
