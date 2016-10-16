package main

import (
	"testing"

	"github.com/poppels/filesys/virtual"
)

const testJSON = `{
	"Words": [
		"Popsicle",
		"Wabbit",
		"Pantaloons",
		"Skedaddle",
		"Flabbergast"
	]
}`

const testXML = `<wordList>
  <word>Popsicle</word>
  <word>Wabbit</word>
  <word>Pantaloons</word>
  <word>Skedaddle</word>
  <word>Flabbergast</word>
</wordList>`

func TestConvertJsonToXml(t *testing.T) {
	fs := virtual.NewVirtualFilesys()

	// PutFile is a convenience method that also creating intermediate directories
	err := fs.PutFile("/a/json/sillywords.json", []byte(testJSON))
	if err != nil {
		t.Fatal(err)
	}

	// Note: MkdirAll ignores the permission flag in the virtual file system
	err = fs.MkdirAll("/a/xml", 0777)
	if err != nil {
		t.Fatal(err)
	}

	err = fs.MkdirAll("/a/workingdir", 0777)
	if err != nil {
		t.Fatal(err)
	}

	// This returns a pointer to the same file system,
	// but with the working directory changed
	fs, err = fs.ChangeDir("/a/workingdir")
	if err != nil {
		t.Fatal(err)
	}

	// Run the method we actually want to test
	err = convertJSONToXML("../json/sillywords.json", "../xml/sillywords.xml", fs)
	if err != nil {
		t.Fatal(err)
	}

	// Check that the file content is what we expect it to be
	content, err := fs.ReadFile("/a/xml/sillywords.xml")
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != testXML {
		t.Fatal("Expected xml does not match actual xml")
	}
}
