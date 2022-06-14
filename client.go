package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/textproto"
	"time"
)

type GcpStorageClient interface {
	ListFiles(ctx context.Context, pageSize int) ([]*storage.ObjectAttrs, error)
	UploadWriter(ctx context.Context, file multipart.File, fileHeader textproto.MIMEHeader, object string) error
	UploadReader(ctx context.Context, object string) ([]byte, error)
}

type gcpStorageClient struct {
	client     *storage.Client
	bucketName string
}

func NewGCPBucketClient(ctx context.Context, bucketName string) (GcpStorageClient, error) {
	client, err := storage.NewClient(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		return nil, err
	}

	return &gcpStorageClient{client, bucketName}, nil
}

// UploadWriter Upload an object with storage.Writer.
func (c *gcpStorageClient) UploadWriter(ctx context.Context, file multipart.File, fileHeader textproto.MIMEHeader, object string) error {
	sw := c.client.Bucket(c.bucketName).Object(object).NewWriter(ctx)
	sw.ContentType = fileHeader.Get("Content-Type")

	if _, err := io.Copy(sw, file); err != nil {
		return err
	}
	defer sw.Close()

	return nil
}

// UploadReader Get an object with storage.Reader.
func (c *gcpStorageClient) UploadReader(ctx context.Context, object string) ([]byte, error) {
	sr, err := c.client.Bucket(c.bucketName).Object(object).NewReader(ctx)
	if err != nil {
		return nil, err
	}

	ioRead, err := ioutil.ReadAll(sr)
	if err != nil {
		return nil, err
	}

	if err := sr.Close(); err != nil {
		return nil, err
	}

	return ioRead, nil
}

func (c *gcpStorageClient) ListFiles(ctx context.Context, pageSize int) ([]*storage.ObjectAttrs, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	var objectAtt []*storage.ObjectAttrs
	it := c.client.Bucket(c.bucketName).Objects(ctx, nil)
	var cursor int
	for {
		attrs, err := it.Next()
		if err == iterator.Done || cursor > pageSize {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Bucket(%q).Objects: %v", c.bucketName, err)
		}
		cursor++
		objectAtt = append(objectAtt, attrs)
	}

	return objectAtt, nil
}
