package file

import (
	"io"
	"testing"
)

func TestContentTypeRewindsSmallFiles(t *testing.T) {
	f := NewInMemoryFile("small.txt", []byte("small file"))

	if got := ContentType(f); got == "" {
		t.Fatal("ContentType returned an empty content type")
	}

	body, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}

	if string(body) != "small file" {
		t.Fatalf("expected file content to remain readable after ContentType, got %q", string(body))
	}
}
