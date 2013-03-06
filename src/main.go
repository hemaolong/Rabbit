// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"image"
	"os"
	"path/filepath"
	"time"

	"github.com/lxn/walk"
	"github.com/lxn/go-winapi"

	// Self
	selfWidget "mywidget"
)

const (
	TB_H int = 16 // Tool bar size
	OB_H int = 12 // Other bar size

	FREEZEIZE_CLASS string = "FREEZEIZE_CLASS0"
)

type (
	MainWindow struct {
		*walk.MainWindow
		imageView    *walk.ImageView
		prevFilePath string
		frameTimer   int

		// Other ui
		uiFrameCnt *walk.NumberEdit
		uiPoseCnt *walk.LineEdit
	}

	// Image struct
	ImageItem struct {
		fname string
		bm    *walk.Bitmap
	}
)

var (
	_iniImgList  [1000]*ImageItem
	imgList      = _iniImgList[0:0]
	currentFrame int
	imageW       int
	imageH       int

	// ui
	frameCount int = 8 // The frame count of a pose

)

//////////////Image Operation
func readImge(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	return img, err
}

func ReadImageList(path, ext string) error {
	folder, err := os.Open(path)
	if err != nil {
		return err
	}

	// Read all the files in the folder
	fileList, err := folder.Readdir(-1)
	if err != nil {
		return err
	}

	imgList = _iniImgList[0:0]
	// Read all png images

	for _, v := range fileList {
		if !v.IsDir() {
			fname := v.Name()
			curExt := filepath.Ext(fname)
			if curExt == ext {
				//img, err := readImge(path + "/" + fname)
				//if err == nil {
				//	fileMap[fname] = img
				//}
				imgList = append(imgList)
				fullname := path + "/" + fname
				if bm, err := walk.NewBitmapFromFile(fullname); err == nil {
					newImg := new(ImageItem)
					newImg.fname = fullname
					newImg.bm = bm
					imgList = append(imgList, newImg)

					imageW = bm.Size().Width
					imageH = bm.Size().Height
				}

			}
		}
	}
	return nil
}

/////////////End image opration

func (mw *MainWindow) openImage() {
	//dlg := &walk.FileDialog{}

	//dlg.FilePath = mw.prevFilePath
	//dlg.Filter = "Image Files (*.emf;*.bmp;*.exif;*.gif;*.jpeg;*.jpg;*.png;*.tiff)|*.emf;*.bmp;*.exif;*.gif;*.jpeg;*.jpg;*.png;*.tiff"
	//dlg.Title = "Select an Image"

	//if ok, _ := dlg.ShowOpen(mw); !ok {
	//	return
	//}
	folderPath := selfWidget.GetPath(0, 0)
	if err := ReadImageList(folderPath, ".png"); err != nil {
		return
	}
	mw.setImageSize()

	//mw.prevFilePath = folderPath// dlg.FilePath

	// img, _ := walk.NewImageFromFile(dlg.FilePath)

	//page, _ := walk.NewTabPage()
	// page.SetTitle(path.Base(strings.Replace(dlg.FilePath, "\\", "/", -1)))
	//page.SetLayout(walk.NewHBoxLayout())

	//var succeeded bool
	//defer func() {
	//	if !succeeded {
	// page.Dispose()
	//	}
	//}()

	// imageView.SetImage(img)
	// mw.tabWidget.Pages().Add(page)
	// mw.tabWidget.SetCurrentIndex(mw.tabWidget.Pages().Len() - 1)

	//succeeded = true
}

func (mw *MainWindow) drawImage() {
	l := len(imgList)
	if l == 0 {
		return
	}

	f := currentFrame % l
	currentFrame++
	mw.imageView.SetImage(imgList[f].bm)
}

func (mw *MainWindow) setImageSize() {
	mw.imageView.SetSize(walk.Size{imageW, imageH})
}

func (mw *MainWindow) initFrame() {
	timer := time.NewTicker(time.Millisecond * 83)
	go func() {
		for _ = range timer.C {
			// <-timer.C
			// fmt.Println(t)
			mw.drawImage()
		}

	}()
}

func (mw *MainWindow) composeImg() {
	println("Save OK")
}

func (mw *MainWindow) initMenu() {
	fileMenu, _ := walk.NewMenu()
	fileMenuAction, _ := mw.Menu().Actions().AddMenu(fileMenu)
	fileMenuAction.SetText("&File")

	//openBmp, _ := walk.NewBitmapFromFile("../img/open.png")
	imageList, _ := walk.NewImageList(walk.Size{TB_H, TB_H}, 0)
	mw.ToolBar().SetImageList(imageList)

	openAction := walk.NewAction()
	// openAction.SetImage(openBmp)
	openAction.SetText("&Open")
	openAction.Triggered().Attach(func() { mw.openImage() })
	fileMenu.Actions().Add(openAction)
	mw.ToolBar().Actions().Add(openAction)

	exitAction := walk.NewAction()
	exitAction.SetText("E&xit")
	exitAction.Triggered().Attach(func() { walk.App().Exit(0) })
	fileMenu.Actions().Add(exitAction)

	helpMenu, _ := walk.NewMenu()
	helpMenuAction, _ := mw.Menu().Actions().AddMenu(helpMenu)
	helpMenuAction.SetText("&Help")

	aboutAction := walk.NewAction()
	aboutAction.SetText("&About")
	aboutAction.Triggered().Attach(func() {
		walk.MsgBox(mw, "About", "Image composer", walk.MsgBoxOK|walk.MsgBoxIconInformation)
	})
	helpMenu.Actions().Add(aboutAction)

	// Image operations
	// Save
	composeAction := walk.NewAction()
	composeAction.SetText("&Save")
	composeAction.Triggered().Attach(func() { mw.composeImg() })
	fileMenu.Actions().Add(composeAction)
	mw.ToolBar().Actions().Add(composeAction)
}

func (mw *MainWindow) initCanvas() {
	// Init image view
	iv, _ := walk.NewImageView(mw)
	mw.imageView = iv
	mw.initFrame()
}
func (mw *MainWindow) initOtherBars() {
	sp, _ := walk.NewSplitter(mw)
	sp.SetSize(walk.Size{800, TB_H})
	walk.NewHSpacer(sp)
	walk.NewHSpacer(sp)
	// others
	mw.uiFrameCnt, _ = walk.NewNumberEdit(sp)
	mw.uiFrameCnt.SetSize(walk.Size{32, TB_H})
	mw.uiFrameCnt.SetRange(1, 100)


	lab, _ := walk.NewLineEdit(sp)
	lab.SetSize(walk.Size{32, 30})
	lab.SetReadOnly(true)

	mw.uiPoseCnt, _ = walk.NewLineEdit(sp)
	mw.uiPoseCnt.SetSize(walk.Size{32, 30})

	// Space


	walk.InitWidget(sp, mw, FREEZEIZE_CLASS,
					winapi.CCS_NORESIZE,
					winapi.WS_EX_TOOLWINDOW | winapi.WS_EX_WINDOWEDGE)
}

func newMainWindow() {
	walk.SetPanicOnError(true)
	mainWnd, _ := walk.NewMainWindow()

	mw := &MainWindow{MainWindow: mainWnd}
	mw.SetLayout(walk.NewVBoxLayout())
	mw.SetTitle("Image composer")
	mw.SetMinMaxSize(walk.Size{800, 600}, walk.Size{})
	mw.SetSize(walk.Size{800, 600})

	mw.initMenu()
	mw.initOtherBars()
	mw.initCanvas()
	mw.Show()
	mw.Run()
}

func init() {
    walk.MustRegisterWindowClass(FREEZEIZE_CLASS)
}

func main() {
	newMainWindow()
}