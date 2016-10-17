# Go File System Abstraction

This library enables unit testing of your application's file system operations without actually writing any changes to disk.

It wraps the most common operations from the _os_ package and some from _ioutil_.

### Usage
There are two ways of using the _filesys_ library.
The first is to set a global singleton FileSystem object.
```
	fs := virtual.NewVirtualFilesys() // or osfilesys.NewOsWrapper()
	filesys.SetGlobalSystem(fs)
```

Then call the _filesys_ methods like you would with _os_. _(error handling omitted in example)_
```
func consumeRSSFeeds(dir string) []RSS {
	fileinfos, _ := filesys.ReadDir(dir)
	feeds := make([]RSS, 0, len(fileinfos))
	for _, fi := range fileinfos {
		fpath := path.Join(dir, fi.Name())
		f, _ := filesys.Open(fpath)
		RSS feed
		rss.NewDecoder(f).Decode(&feed)
		feeds = append(feeds, feed)
		f.Close()
		filesys.Remove(fpath)
	}
	return feeds
}
```

The second way is to pass around the FileSystem objects and use them driectly. They implement the same methods as _filesys_. _(error handling still omitted)_
```
func consumeRSSFeeds(dir string, fs filesys.FileSystem) []RSS {
	fileinfos, _ := fs.ReadDir(dir)
	feeds := make([]RSS, 0, len(fileinfos))
	for _, fi := range fileinfos {
		fpath := path.Join(dir, fi.Name())
		f, _ := fs.Open(fpath)
		RSS feed
		rss.NewDecoder(f).Decode(&feed)
		feeds = append(feeds, feed)
		f.Close()
		fs.Remove(fpath)
	}
	return feeds
}
```

The _fsutil_ package contains some useful functions for unit tests _(error handling omitted)_.
```
	fs := virtual.NewVirtualFilesys()
	fsutil.PutFile(fs, "/a/sillywords.json", []byte(testJSON))
	convertJSONToXML("/a/sillywords.json", "/a/sillywords.xml", fs)
	err := fsutil.VerifyFileContent(fs, "/a/sillywords.xml", []byte(testXML))
```

For a small example, see [example](https://github.com/poppels/filesys/tree/master/example)
