package virtual

import (
	"io"
	"io/ioutil"
	"testing"
)

func TestRead(t *testing.T) {
	fs := NewVirtualFilesys()
	fs.PutFile("/a/b/c.txt", []byte("Hello"))
	f, err := fs.Open("/a/b/c.txt")
	if err != nil {
		t.Fatal(err)
	}
	result, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	expected := "Hello"
	if string(result) != expected {
		t.Fatalf("Expected content '%s', got '%s'", expected, result)
	}
}

func TestWrite(t *testing.T) {
	fs := NewVirtualFilesys()
	err := fs.MkdirAll("/a/b", 0777)
	if err != nil {
		t.Fatal(err)
	}
	f, err := fs.Create("/a/b/c.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_, err = f.Write([]byte("Hello"))
	if err != nil {
		t.Fatal(err)
	}
	content, err := fs.ReadFile("/a/b/c.txt")
	if err != nil {
		t.Fatal(err)
	}
	expected := "Hello"
	if string(content) != expected {
		t.Fatalf("Expected file content '%s', got '%s'", expected, content)
	}
}

func TestSeek(t *testing.T) {
	fs := NewVirtualFilesys()
	err := fs.MkdirAll("/a/b", 0777)
	if err != nil {
		t.Fatal(err)
	}
	f, err := fs.Create("/a/b/c.txt")
	if err != nil {
		t.Fatalf("Error creating file: %s", err.Error())
	}
	defer f.Close()
	if _, err = f.Write([]byte("Hello")); err != nil {
		t.Fatal(err)
	}

	// Seek from beginning
	var pos int64
	if pos, err = f.Seek(1, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	if pos != 1 {
		t.Fatalf("Expected position 1, got %d", pos)
	}
	if _, err = f.Write([]byte("ipp")); err != nil {
		t.Fatal(err)
	}
	content, err := fs.ReadFile("/a/b/c.txt")
	if err != nil {
		t.Fatal(err)
	}
	expected := "Hippo"
	if string(content) != "Hippo" {
		t.Fatalf("Expected '%s', got '%s'", expected, content)
	}

	// Seek from current position
	if pos, err = f.Seek(-2, io.SeekCurrent); err != nil {
		t.Fatal(err)
	}
	if pos != 2 {
		t.Fatalf("Expected position 2, got %d", pos)
	}
	if _, err = f.Write([]byte("ng")); err != nil {
		t.Fatal(err)
	}
	if content, err = fs.ReadFile("/a/b/c.txt"); err != nil {
		t.Fatal(err)
	}
	expected = "Hingo"
	if string(content) != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, content)
	}

	// Seek from end
	if pos, err = f.Seek(3, io.SeekEnd); err != nil {
		t.Fatal(err)
	}
	if pos != 8 {
		t.Fatalf("Expected position 8, got %d", pos)
	}
	if _, err = f.Write([]byte("Hey")); err != nil {
		t.Fatal(err)
	}
	if content, err = fs.ReadFile("/a/b/c.txt"); err != nil {
		t.Fatal(err)
	}
	expected = "Hingo\x00\x00\x00Hey"
	if string(content) != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, content)
	}
}

func TestFolderReaddir(t *testing.T) {
	fs := NewVirtualFilesys()
	if err := fs.MkdirAll("/a/b/c", 0777); err != nil {
		t.Fatal(err)
	}
	if err := fs.PutFile("/a/b/f.txt", []byte("Hello")); err != nil {
		t.Fatal(err)
	}
	if err := fs.PutFile("/a/b/Z.txt", []byte("Bye")); err != nil {
		t.Fatal(err)
	}

	f, err := fs.Open("/a/b")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// Read all
	infos, err := f.Readdir(-1)
	if err != nil {
		t.Fatal(err)
	}
	for _, info := range infos {
		t.Log(info.Name())
	}
	expected := []string{"Z.txt", "c", "f.txt"}
	if len(infos) != len(expected) {
		t.Fatalf("Expected %d files/folders, got %d", len(expected), len(infos))
	}
	for i := 0; i < len(infos); i++ {
		if expected[i] != infos[i].Name() {
			t.Fatalf("Expected infos[%d] to be '%s', got '%s'", i, expected[i], infos[i].Name())
		}
	}
	f2, err := fs.Open("/a/b")
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()

	// Read some
	if infos, err = f2.Readdir(2); err != nil {
		t.Fatal(err)
	}
	expected = []string{"Z.txt", "c"}
	for _, info := range infos {
		t.Log(info.Name())
	}
	if len(infos) != len(expected) {
		t.Fatalf("Expected %d files/folders, got %d", len(expected), len(infos))
	}
	for i := 0; i < len(infos); i++ {
		if expected[i] != infos[i].Name() {
			t.Fatalf("Expected infos[%d] to be '%s', got '%s'", i, expected[i], infos[i].Name())
		}
	}

	// Read past EOF
	if infos, err = f2.Readdir(4); err == nil {
		t.Fatal("Expected io.EOF error")
	}
	for _, info := range infos {
		t.Log(info.Name())
	}
	expected = []string{"f.txt"}
	if len(infos) != len(expected) {
		t.Fatalf("Expected %d files/folders, got %d", len(expected), len(infos))
	}
	for i := 0; i < len(infos); i++ {
		if expected[i] != infos[i].Name() {
			t.Fatalf("Expected infos[%d] to be '%s', got '%s'", i, expected[i], infos[i].Name())
		}
	}

	// Readdir error for file
	f3, err := fs.Open("/a/b/f.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f3.Close()
	if _, err := f3.Readdir(-1); err == nil {
		t.Fatal("Error expected for Readdir on file")
	}
}
