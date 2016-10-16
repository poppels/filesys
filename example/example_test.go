package main

import (
	"testing"

	"github.com/poppels/filesys/fsutil"
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
	files := map[string][]byte{"/a/json/sillywords.json": []byte(testJSON)}
	folders := []string{"/a/xml", "/a/workingdir"}

	fs := virtual.NewVirtualFilesys()
	if err := fsutil.CreateStructure(fs, files, folders); err != nil {
		t.Fatal(err)
	}

	// This returns a pointer to the same file system,
	// but with the working directory changed
	fs, err := fs.ChangeDir("/a/workingdir")
	if err != nil {
		t.Fatal(err)
	}

	// Run the method we actually want to test
	err = convertJSONToXML("../json/sillywords.json", "../xml/sillywords.xml", fs)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the file content is what we expect it to be
	err = fsutil.VerifyFileContent(fs, "/a/xml/sillywords.xml", []byte(testXML))
	if err != nil {
		t.Fatal(err)
	}
}
