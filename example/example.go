package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/poppels/filesys"
	"github.com/poppels/filesys/osfilesys"
)

type wordList struct {
	Words []string `xml:"word"`
}

func main() {
	fs := osfilesys.NewOsWrapper()
	err := convertJSONToXML("sillywords.json", "sillywords.xml", fs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func convertJSONToXML(inpath, outpath string, fs filesys.FileSystem) error {
	infile, err := fs.Open(inpath)
	if err != nil {
		return err
	}
	defer infile.Close()

	var wl wordList
	decoder := json.NewDecoder(infile)
	if err = decoder.Decode(&wl); err != nil {
		return err
	}

	outfile, err := fs.Create(outpath)
	if err != nil {
		return err
	}
	defer outfile.Close()

	encoder := xml.NewEncoder(outfile)
	encoder.Indent("", "  ")
	err = encoder.Encode(wl)
	return err
}
