package virtual

import (
	"testing"
	"time"
)

func TestMkdir(t *testing.T) {
	fs := NewVirtualFilesys()
	err := fs.Mkdir("/a", 0777)
	if err != nil {
		t.Fatal(err)
	}
	_, err = fs.Stat("/a")
	if err != nil {
		t.Fatal(err)
	}
	err = fs.Mkdir("/a/b", 0777)
	if err != nil {
		t.Fatal(err)
	}
	err = fs.Mkdir("/a/b/c/d", 0777)
	if err == nil {
		t.Fatal("Directory created in non existing path")
	}
	err = fs.PutFile("/a/b/c.txt", []byte{})
	if err != nil {
		t.Fatal(err)
	}
	err = fs.Mkdir("/a/b/c.txt", 0777)
	if err == nil {
		t.Fatal("File overwritten with directory")
	}
}

func TestMkdirAll(t *testing.T) {
	fs := NewVirtualFilesys()
	err := fs.MkdirAll("/a/b", 0777)
	if err != nil {
		t.Fatal(err)
	}
	_, err = fs.Stat("/a/b/")
	if err != nil {
		t.Fatal(err)
	}
	err = fs.PutFile("/a/b/c.txt", []byte{})
	if err != nil {
		t.Fatal(err)
	}
	err = fs.MkdirAll("/a/b/c.txt", 0777)
	if err == nil {
		t.Fatal("File overwritten with directory")
	}
}

func TestPutFile(t *testing.T) {
	fs := NewVirtualFilesys()
	fs.PutFile("/a/b/c.txt", []byte("Hello"))
	content, err := fs.ReadFile("/a/b/c.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "Hello" {
		t.Fatalf("Expected file content 'Hello', got '%s'", content)
	}
	err = fs.PutFile("/a/b/d.txt/", []byte("Bye"))
	if err == nil {
		t.Fatal("Error expected when creating a directory ending with /")
	}
}

func TestReadDir(t *testing.T) {
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

	infos, err := fs.ReadDir("/a/b")
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
}

func TestRename(t *testing.T) {
	fs := NewVirtualFilesys()
	if err := fs.MkdirAll("/a/b/c", 0777); err != nil {
		t.Fatal(err)
	}
	if err := fs.MkdirAll("/a/d", 0777); err != nil {
		t.Fatal(err)
	}
	if err := fs.Rename("/a/b", "/a/d/e"); err != nil {
		t.Fatal(err)
	}
	fi, err := fs.Stat("/a/d/e/c")
	if err != nil {
		t.Fatal(err)
	}
	if fi.Name() != "c" {
		t.Fatalf("Expected folder name 'c', got '%s'", fi.Name())
	}
	if _, err := fs.Stat("/a/b/c"); err == nil {
		t.Fatal("Folder /a/b/c should not exist anymore")
	}
	if err := fs.PutFile("/a/k/f.txt", []byte("Yo")); err != nil {
		t.Fatal(err)
	}
	if err := fs.Rename("/a/k/f.txt", "/a/d/e/g.txt"); err != nil {
		t.Fatal(err)
	}
	fi, err = fs.Stat("/a/d/e/g.txt")
	if err != nil {
		t.Fatal(err)
	}
	if fi.Name() != "g.txt" {
		t.Fatalf("Expected file name 'g.txt', got %s", fi.Name())
	}
	if err := fs.Rename("/a/d/e/g.txt", "/a/d/e/c"); err == nil {
		t.Fatal("Should not be able to overwrite folder with file")
	}
	if err := fs.Rename("/a/k", "/a/d/e/c"); err == nil {
		t.Fatal("Should not be able to overwrite folder with another folder")
	}
	if err := fs.Rename("/a/k", "/a/d/e/g.txt"); err == nil {
		t.Fatal("Should not be able to overwrite file with folder")
	}
	err = fs.PutFile("/a/k/h.txt", []byte("Hey"))
	if err != nil {
		t.Fatal(err)
	}
	if err := fs.Rename("/a/k/h.txt", "/a/d/e/g.txt"); err != nil {
		t.Fatal(err)
	}
	fi, err = fs.Stat("/a/d/e/g.txt")
	if err != nil {
		t.Fatal(err)
	}
	if fi.Name() != "g.txt" {
		t.Fatalf("Expected file name 'g.txt', got '%s'", fi.Name())
	}
	content, err := fs.ReadFile("/a/d/e/g.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "Hey" {
		t.Fatalf("Expected file content 'Hey', got '%s'", content)
	}
	if err := fs.Rename("/a/k", "/a/j/o"); err == nil {
		t.Fatal("Expected os.ErrNotExist")
	}
	if err := fs.Rename("/a/d", "/a/d/d"); err == nil {
		t.Fatal("Should not be able to move folder into itself")
	}
	if err := fs.Rename("/a/d", "/a/d/e/d"); err == nil {
		t.Fatal("Should not be able to move folder to its own subdirectory")
	}
}

func TestChtimes(t *testing.T) {
	fs := NewVirtualFilesys()
	fs.PutFile("/a/b/c.txt", []byte("Hello"))
	mtime := time.Now().Add(-10 * time.Hour)
	err := fs.Chtimes("/a/b/c.txt", time.Now(), mtime)
	if err != nil {
		t.Fatal(err)
	}
	fi, err := fs.Stat("/a/b/c.txt")
	if err != nil {
		t.Fatal(err)
	}
	if fi.ModTime() != mtime {
		layout := "2006-01-02 15-04-05"
		t.Fatalf("Expected time '%s', got '%s'", mtime.Format(layout), fi.ModTime().Format(layout))
	}
}

func TestCurrentDir(t *testing.T) {
	fs := NewVirtualFilesys()
	path := "/abra/kadabra/sim/salabim/€ß"
	err := fs.MkdirAll(path, 0777)
	if err != nil {
		t.Fatal(err)
	}
	fs, err = fs.ChangeDir(path)
	if err != nil {
		t.Fatal(err)
	}
	cd := fs.CurrentDir()
	if cd != path {
		t.Fatalf("Expected path '%s', got '%s'", path, cd)
	}
}

func BenchmarkCurrentDir(b *testing.B) {
	fs := NewVirtualFilesys()
	err := fs.MkdirAll("/abra/kadabra/sim/salabim/€ß", 0777)
	if err != nil {
		b.Fatal(err)
	}
	fs, err = fs.ChangeDir("/abra/kadabra/sim/salabim/€ß")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.CurrentDir()
	}
}
