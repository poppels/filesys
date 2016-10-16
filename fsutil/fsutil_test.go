package fsutil

import (
	"testing"

	"github.com/poppels/filesys/virtual"
)

func TestPutFile(t *testing.T) {
	fs := virtual.NewVirtualFilesys()
	PutFile(fs, "/a/b/c.txt", []byte("Hello"))
	content, err := fs.ReadFile("/a/b/c.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "Hello" {
		t.Fatalf("Expected file content 'Hello', got '%s'", content)
	}
	err = PutFile(fs, "/a/b/d.txt/", []byte("Bye"))
	if err == nil {
		t.Fatal("Error expected when creating a file ending with /")
	}
}
