package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"github.com/sisisin/gh-project-sync/lib/logger"
	"google.golang.org/api/iterator"
)

type appBucket struct {
	client *storage.Client
	bucket string
}

func New(ctx context.Context, bucketName string) (*appBucket, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create storage client")
	}
	return &appBucket{
		client: client,
		bucket: bucketName,
	}, nil
}

func (s *appBucket) List(ctx context.Context, path string, body []byte) error {
	it := s.client.Bucket(s.bucket).Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return errors.Wrap(err, "failed to iterate over objects")
		}
		logger.Infof(ctx, "object: %s", attrs.Name)
	}

	return nil
}

func (s *appBucket) GetObjectWriter(ctx context.Context, path string) *storage.Writer {
	o := s.client.Bucket(s.bucket).Object(path)
	w := o.NewWriter(ctx)
	return w
}
