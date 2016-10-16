# Go File System Abstraction

This library enables unit testing of your application's file system operations without actually writing any changes to disk.

It wraps the most common operations from the __os__ package and some from __ioutil__.

### Usage
Create a FileSystem object that is passed around instead of using the __os__ methods directly.
```
	fs := osfilesys.NewOsWrapper()
	convertJSONToXML("sillywords.json", "sillywords.xml", fs)
```

Use the regular functions like you would with __os__ _(error handling omitted in example)_.
```
func convertJSONToXML(inpath, outpath string, fs filesys.FileSystem) error {
	infile, _ := fs.Open(inpath)
	defer infile.Close()
    var wl wordList
    decoder := json.NewDecoder(infile)
    decoder.Decode(&wl);
	(...)
```

Create a virtual file system to use in tests _(error handling omitted in example)_.
```
	fs := virtual.NewVirtualFilesys()
	fs.PutFile("/a/sillywords.json", []byte(testJSON))
	convertJSONToXML("/a/sillywords.json", "/a/sillywords.xml", fs)
```

For a small example, see [example](https://github.com/poppels/filesys/tree/master/example)