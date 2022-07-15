package s3

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"terralist/pkg/file/getter"
	"terralist/pkg/file/zipper"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/mazen160/go-random"
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

func (r *Resolver) Store(url string, archive bool) (string, error) {
	// Generate a random temporary directory
	tempDirName, _ := random.String(32)
	tempDir := filepath.Clean(path.Join(r.CacheDir, tempDirName))
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("could not create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := getter.Get(url, tempDir); err != nil {
		return "", fmt.Errorf("could not fetch from the URL: %w", err)
	}

	var resultPath string
	if archive {
		//
		archivePath, err := zipper.Zip(tempDir, r.CacheDir)
		if err != nil {
			return "", fmt.Errorf("could not create archive: %w", err)
		}

		resultPath = archivePath
	} else {
		files, _ := ioutil.ReadDir(tempDir)
		resultPath = path.Join(tempDir, files[0].Name())
	}

	key := filepath.Base(resultPath)
	if err := r.upload(key, resultPath); err != nil {
		return "", fmt.Errorf("could not upload archive: %v", err)
	}

	if archive {
		_ = os.Remove(resultPath)
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

func (r *Resolver) upload(key, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	inf, _ := f.Stat()
	size := inf.Size()
	buffer := make([]byte, size)

	_, _ = f.Read(buffer)

	if _, err := s3.New(r.Session).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(r.BucketName),
		Key:                  aws.String(key),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	}); err != nil {
		return err
	}

	return nil
}
