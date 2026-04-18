package providers

import (
	"context"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	appconfig "github.com/frdavidh/nyarikos/internal/config"
)

type S3Provider struct {
	s3Client *s3.Client
	tm       *transfermanager.Client
	bucket   string
	endpoint string
}

func NewS3Provider(cfg *appconfig.Config) *S3Provider {
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.AWS.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWS.AccessKey,
			cfg.AWS.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		panic(err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.AWS.S3Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.AWS.S3Endpoint)
			o.UsePathStyle = true
		}
	})

	return &S3Provider{
		s3Client: s3Client,
		tm:       transfermanager.New(s3Client),
		bucket:   cfg.AWS.S3BucketName,
		endpoint: cfg.AWS.S3Endpoint,
	}
}

func (p *S3Provider) UploadFile(file *multipart.FileHeader, path string) (string, error) {
	// log.Printf("upload file %s", path)
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	_, err = p.tm.UploadObject(context.TODO(), &transfermanager.UploadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(path),
		Body:   src,
	})
	if err != nil {
		return "", err
	}

	return path, nil
}

func (p *S3Provider) DeleteFile(path string) error {
	_, err := p.s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(path),
	})

	return err
}
