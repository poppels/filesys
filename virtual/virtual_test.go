package virtual

import (
	"testing"
	"time"

	"github.com/poppels/filesys/fsutil"
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
	err = fs.WriteFile("/a/b/c.txt", []byte{}, 0666)
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
	err = fs.WriteFile("/a/b/c.txt", []byte{}, 0666)
	if err != nil {
		t.Fatal(err)
	}
	err = fs.MkdirAll("/a/b/c.txt", 0777)
	if err == nil {
		t.Fatal("File overwritten with directory")
	}
}

func TestReadDir(t *testing.T) {
	folders := []string{"/a/b/c"}
	files := map[string][]byte{
		"/a/b/f.txt": []byte("Hello"),
		"/a/b/Z.txt": []byte("Bye")}

	fs := NewVirtualFilesys()
	if err := fsutil.CreateStructure(fs, files, folders); err != nil {
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
	folders := []string{"/a/b/c", "/a/d"}
	files := map[string][]byte{
		"/a/b/f.txt": []byte("Hello"),
		"/a/b/Z.txt": []byte("Bye"),
		"/a/k/f.txt": []byte("Yo"),
		"/a/k/h.txt": []byte("Hey")}

	fs := NewVirtualFilesys()
	if err := fsutil.CreateStructure(fs, files, folders); err != nil {
		t.Fatal(err)
	}

	// Move directory with content
	if err := fs.Rename("/a/b", "/a/d/e"); err != nil {
		t.Fatal(err)
	}

	fi, err := fs.Stat("/a/d/e")
	if err != nil {
		t.Fatal(err)
	} else if fi.Name() != "e" {
		t.Fatalf("Expected folder name 'e', got '%s'", fi.Name())
	} else if _, err := fs.Stat("/a/b"); err == nil {
		t.Fatal("Folder /a/b should not exist anymore")
	}

	infos, err := fs.ReadDir("/a/d/e")
	if err != nil {
		t.Fatal(err)
	} else if len(infos) != 3 {
		t.Fatal("Expected /a/d/e to contain 3 children, got %d", len(infos))
	}

	// Move file
	if err := fs.Rename("/a/k/f.txt", "/a/d/e/g.txt"); err != nil {
		t.Fatal(err)
	}
	fi, err = fs.Stat("/a/d/e/g.txt")
	if err != nil {
		t.Fatal(err)
	} else if fi.Name() != "g.txt" {
		t.Fatalf("Expected file name 'g.txt', got %s", fi.Name())
	}

	// Test forbidden overwrites
	if err := fs.Rename("/a/d/e/g.txt", "/a/d/e/c"); err == nil {
		t.Fatal("Should not be able to overwrite folder with file")
	}
	if err := fs.Rename("/a/k", "/a/d/e/c"); err == nil {
		t.Fatal("Should not be able to overwrite folder with another folder")
	}
	if err := fs.Rename("/a/k", "/a/d/e/g.txt"); err == nil {
		t.Fatal("Should not be able to overwrite file with folder")
	}

	// Overwrite file
	if err := fs.Rename("/a/k/h.txt", "/a/d/e/g.txt"); err != nil {
		t.Fatal(err)
	}
	fi, err = fs.Stat("/a/d/e/g.txt")
	if err != nil {
		t.Fatal(err)
	} else if fi.Name() != "g.txt" {
		t.Fatalf("Expected file name 'g.txt', got '%s'", fi.Name())
	}
	err = fsutil.VerifyFileContent(fs, "/a/d/e/g.txt", []byte("Hey"))
	if err != nil {
		t.Fatal(err)
	}

	// Move to non existing path
	if err := fs.Rename("/a/k", "/a/j/o"); err == nil {
		t.Fatal("Expected os.ErrNotExist")
	}

	// Moving directory into itself or one of its subfolders
	if err := fs.Rename("/a/d", "/a/d/d"); err == nil {
		t.Fatal("Should not be able to move folder into itself")
	}
	if err := fs.Rename("/a/d", "/a/d/e/d"); err == nil {
		t.Fatal("Should not be able to move folder to its own subdirectory")
	}
}

func TestChtimes(t *testing.T) {
	fs := NewVirtualFilesys()
	fsutil.PutFile(fs, "/a/b/c.txt", []byte("Hello"))
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
