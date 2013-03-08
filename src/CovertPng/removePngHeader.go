// Parse the work path, remove the png file header
package main

import (
	"io"
	"os"
	"path/filepath"
	// _ "image/png"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Open folder
	folder, err := os.Open(wd)
	if err != nil {
		panic(err)
	}

	// Read all the png files
	fileList, err := folder.Readdir(-1)
	if err != nil {
		panic(err)
	}

	const headLen int64 = 24
	var head [headLen]byte
	var body [1024 * 1024]byte
	for _, v := range fileList {
		if !v.IsDir() {
			fname := v.Name()
			ext := filepath.Ext(fname)
			if ext == ".png" {
				// Covert png
				f, err := os.OpenFile(wd+"\\"+fname, os.O_RDWR, os.ModePerm)
				defer f.Close()
				if err != nil {
					panic(err)
				}

				f.ReadAt(head[:], 0)
				if head[0] == 0 && head[1] == 0 && head[3] == 8 {
					// Convert
					_info, _ := f.Stat()
					filesize := _info.Size()
					count, err := f.ReadAt(body[:filesize - headLen], headLen)
					if err != nil && err != io.EOF {
						panic(err)
					}
					print("File size ")
					println(count)
					f.Truncate(0)
					_, err = f.WriteAt(body[:count], 0)
					if err != nil {
						panic(err)
					}
					println("Convert file " + fname + " ")
				}
			}
		}
	}
}
