package main


import(
    "image"
    "os"
    "path/filepath"

)

func ReadImage(path string)(image.Image, string, error) {
    f, err := os.Open(path)
    if  err != nil{
        return nil, "", err
    }
    defer f.Close()
    return image.Decode(f)
}

func ReadImageList(path, ext string) []image.Image {
    // Iterate the path, find all the image files as the ext.
    imgFiles := make(map[string] int)
    filepath.Walk(path,
        func(p string, f os.FileInfo, err error) error{
            if f.IsDir() {
                return nil
            }
            imgFiles[f.Name()] = 1
            return nil
        })


    // Read all the images
    result := make([]image.Image, len (imgFiles))
    return result
}
