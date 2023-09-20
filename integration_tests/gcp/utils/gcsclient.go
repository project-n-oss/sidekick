package utils

import (
	"context"
	"testing"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

func GetSidekickGcsClient(t *testing.T, ctx context.Context) *storage.Client {
	client, err := storage.NewClient(context.Background(), option.WithEndpoint(SidekickURL))
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func GetGoogleGcsClient(t *testing.T, ctx context.Context) *storage.Client {
	client, err := storage.NewClient(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
