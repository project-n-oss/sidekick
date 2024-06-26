package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApp_RequestIsCrunchedFile(t *testing.T) {
	testCases := []struct {
		path     string
		expected bool
	}{
		{path: "/foo/bar", expected: false},
		{path: "/foo/bar/", expected: false},
		{path: "foo/bar/", expected: false},
		{path: "/foo/bar/myfile.parquet", expected: false},
		{path: "/foo/bar/myfile.c00.zstd.parquet", expected: false},
		{path: "/foo/bar/myfile.parquet", expected: false},
		{path: "/foo/bar/myfile.0.gr.parquet", expected: true},
		{path: "/foo/bar/myfile.zstd.0.gr.parquet", expected: true},
		{path: "/foo/bar/0.gr.myfile", expected: true},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			assert.Equal(t, tc.expected, isCrunchedFile(tc.path))
		})
	}
}

func TestApp_RequestMakeCrunchFilePath(t *testing.T) {
	testCases := []struct {
		path     string
		bucket   string
		expected string
	}{
		{path: "/foo/bar", bucket: "foo", expected: "0.gr.bar"},
		{path: "/foo/bar/", bucket: "foo", expected: "bar/0.gr."},
		{path: "/foo/bar/myfile.parquet", bucket: "foo", expected: "bar/myfile.0.gr.parquet"},
		{path: "/foo/bar/myfile.c00.zstd.parquet", bucket: "foo", expected: "bar/myfile.c00.zstd.0.gr.parquet"},
		{path: "/foo/bar/myfile.c00.zstd.parquet", bucket: "foo", expected: "bar/myfile.c00.zstd.0.gr.parquet"},
		{path: "bar", bucket: "foo", expected: "0.gr.bar"},
		{path: "bar.parquet", bucket: "foo", expected: "bar.0.gr.parquet"},
		{path: "bar/", bucket: "foo", expected: "bar/0.gr."},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			assert.Equal(t, tc.expected, makeCrunchFilePath(tc.bucket, tc.path))
		})
	}
}
