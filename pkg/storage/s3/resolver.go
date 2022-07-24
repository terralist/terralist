package s3

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"terralist/pkg/file/getter"
	"terralist/pkg/file/zipper"
	"terralist/pkg/storage"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// The S3 resolver will download files from the given URL then
// uploads them to an S3 bucket, generating a public download URL

// Resolver is the concrete implementation of storage.Resolver
type Resolver struct {
	CacheDir   string
	BucketName string
	LinkExpire int

	Session *session.Session
}

func (r *Resolver) Store(in *storage.StoreInput) (string, error) {
	var buffer []byte

	if in.Content != nil {
		buffer = in.Content
	} else {
		// Generate a random temporary directory
		tempDir, err := os.MkdirTemp("", "terralist")
		if err != nil {
			return "", fmt.Errorf("could not create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir) // clean up

		if err := getter.Get(in.URL, tempDir); err != nil {
			return "", fmt.Errorf("could not fetch from the URL: %w", err)
		}

		var resultPath string
		if in.Archive {
			archivePath, err := zipper.Zip(tempDir, r.CacheDir)
			if err != nil {
				return "", fmt.Errorf("could not create archive: %w", err)
			}

			resultPath = archivePath
		} else {
			// The tempDir is created in this flow, read dir cannot fail
			files, _ := ioutil.ReadDir(tempDir)
			resultPath = path.Join(tempDir, files[0].Name())
		}

		// If the file is created in this flow, open cannot fail
		f, _ := os.Open(resultPath)
		defer f.Close()
		defer os.Remove(resultPath)

		// Same for stat
		inf, _ := f.Stat()
		size := inf.Size()
		buffer = make([]byte, size)

		_, _ = f.Read(buffer)
	}

	key := fmt.Sprintf("%s/%s", in.KeyPrefix, in.FileName)
	if _, err := s3.New(r.Session).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(r.BucketName),
		Key:                  aws.String(key),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(int64(len(buffer))),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	}); err != nil {
		return "", fmt.Errorf("could not upload archive: %v", err)
	}

	return key, nil
}

func (r *Resolver) Find(key string) (string, error) {
	req, _ := s3.New(r.Session).GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(r.BucketName),
		Key:    aws.String(key),
	})

	url, err := req.Presign(time.Duration(r.LinkExpire) * time.Minute)
	if err != nil {
		return "", fmt.Errorf("could not generate URL for %v: %v", key, err)
	}

	return url, nil
}

func (r *Resolver) Purge(key string) error {
	if _, err := s3.New(r.Session).DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(r.BucketName),
		Key:    aws.String(key),
	}); err != nil {
		return fmt.Errorf("could not purge object: %v", err)
	}

	return nil
}
