package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/lxn/go-winapi"
	"github.com/lxn/walk"

	// Self
	selfWidget "mywidget"
)

type ImageExt interface {
	ColorModel() color.Model
	Bounds() image.Rectangle
	At(x, y int) color.Color

	// (x, y, stride int)

	SubImage(image.Rectangle) image.Image
}

const (
	TB_H int = 16 // Tool bar size
	OB_H int = 12 // Other bar size

	FREEZEIZE_CLASS string = "FREEZEIZE_CLASS0"

	ttPosCnt   string = "动作个数"
	ttPlayPose string = "播放动作"

	// MODE
	MODE_COMPOSE int = 0
	MODE_PLAY    int = 1
	MODE_INVALID int = 3
)

type (
	MainWindow struct {
		*walk.MainWindow
		imageView *selfWidget.MyImageView

		// Other ui
		uiFrameCnt *walk.NumberEdit
		uiPoseCnt  *walk.NumberEdit
		uiConvirm  *walk.PushButton
		mode       int

		refreshTimer *time.Ticker
	}

	// Image struct
	ImageItem struct {
		fname string
		bm    *walk.Bitmap
		img   ImageExt
	}
)

var (
	modelItem    *ImageItem
	_iniImgList  [1000]*ImageItem
	imgList      = _iniImgList[0:0]
	currentFrame int
	imageW       int
	imageH       int

	boundary image.Rectangle = image.Rect(-1, -1, -1, -1)

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

func parseImgBoundary(img image.Image) {
	minX, maxX := img.Bounds().Min.X, img.Bounds().Max.X
	minY, maxY := img.Bounds().Min.Y, img.Bounds().Max.Y
	for i := minX; i < maxX; i++ {
		for j := minY; j < maxY; j++ {
			_, _, _, a := img.At(i, j).RGBA()
			if a != 0 {
				if boundary.Empty() && boundary.Min.X == -1 {
					boundary = image.Rect(i, j, i, j)
				} else {
					_p := image.Point{i, j}
					if !_p.In(boundary) {
						boundary = boundary.Union(image.Rect(i, j, i, j))
					}
				}
			}
		}
	}

	// Should Make the midline of boundary and the img
	l := boundary.Min.X
	r := imageW - boundary.Max.X
	if l > r {
		boundary.Min.X = r
	} else if l < r {
		boundary.Max.X = imageW - l
	}
}

func readPoseImage(path, ext string) {
	imgList = _iniImgList[0:0]
	// Read all png images

	curExt := filepath.Ext(path)
	if curExt == ext {
		if bm, err := walk.NewBitmapFromFile(path); err == nil {
			newImg := new(ImageItem)
			newImg.fname = path
			newImg.bm = bm
			_img, err := readImge(path)
			if err == nil {
				newImg.img = _img.(ImageExt)
				imgList = append(imgList, newImg)

				imageW = bm.Size().Width
				imageH = bm.Size().Height
				parseImgBoundary(newImg.img)
			}
			modelItem = newImg

		}
	}
}

func readImageList(path, ext string) error {
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
				imgList = append(imgList)
				fullname := path + "/" + fname
				if bm, err := walk.NewBitmapFromFile(fullname); err == nil {
					newImg := new(ImageItem)
					newImg.fname = fullname
					newImg.bm = bm
					_img, err := readImge(fullname)
					if err == nil {
						newImg.img = _img.(ImageExt)
						imgList = append(imgList, newImg)

						imageW = bm.Size().Width
						imageH = bm.Size().Height
						parseImgBoundary(newImg.img)
					}
				}

			}
		}
	}
	return nil
}

/////////////End image opration

func (mw *MainWindow) openImage(mode int) {
	var folderPath string
	if mode == MODE_COMPOSE {
		folderPath = selfWidget.GetPath(mw, "Select the folder")
	} else {
		title := "Select an image"
		filter := "Image Files (*.png)|*.png"
		folderPath = selfWidget.GetOpenFileName(mw, title, filter)
	}
	f, err := os.Open(folderPath)
	if err != nil {
		return
	}

	fs, err := f.Stat()
	if err != nil {
		return
	}

	if mw.refreshTimer != nil {
		mw.refreshTimer.Stop()
	}

	boundary = image.Rect(-1, -1, -1, -1)

	if fs.IsDir() {
		readImageList(folderPath, ".png")
		mw.setImageSize()
		mw.refreshToolBar(MODE_COMPOSE)
		mw.initFrame()
		return
	}
	readPoseImage(folderPath, ".png")
	mw.setImageSize()
	mw.refreshToolBar(MODE_PLAY)
	mw.onClickSave()

	mw.initFrame()
}

func (mw *MainWindow) saveImage() {
	title := "Save the image"
	filter := "Image Files (*.png)|*.png"
	path := selfWidget.GetSaveFileName(mw, title, filter)
	if len(path) == 0 {
		return
	}
	mw.composeImg(path)
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

// Calc the image size, draw the image boundary
func (mw *MainWindow) setImageSize() {
	mw.imageView.SetSize(walk.Size{imageW, imageH})

	fmt.Printf("Boundary %v\n", boundary)
	mw.imageView.SetBoundary(boundary.Min.X, boundary.Min.Y,
		boundary.Dx(), boundary.Dy())
}

func (mw *MainWindow) initFrame() {
	mw.refreshTimer = time.NewTicker(time.Millisecond * 83)
	go func() {
		for _ = range mw.refreshTimer.C {
			mw.drawImage()
		}

	}()

	mw.Invalidate()
}

func (mw *MainWindow) onClickSave() {
	if modelItem == nil {
		return
	}

	imageW = modelItem.img.Bounds().Dx()
	imageH = modelItem.img.Bounds().Dy()

	poseCnt := int(mw.uiPoseCnt.Value())
	if poseCnt <= 0 {
		poseCnt = 1
	}
	frame := 8

	imageW /= frame
	imageH /= poseCnt

	imgList = _iniImgList[0:0]
	boundary = image.Rect(0, 0, imageW, imageH)
	tmpBound := boundary
	// Read all png images
	for i := 0; i < poseCnt; i++ {
		for j := 0; j < frame; j++ {
			deltaX := imageW * j
			deltaY := imageH * i
			tmpBound = boundary.Add(image.Point{deltaX, deltaY})

			newImg := new(ImageItem)
			newImg.fname = ""
			newImg.img = modelItem.img.SubImage(tmpBound).(ImageExt)
			newImg.bm, _ = walk.NewBitmapFromImage(newImg.img)

			imgList = append(imgList, newImg)
		}
	}
	mw.setImageSize()
}

func (mw *MainWindow) refreshToolBar(mode int) {
	mw.uiConvirm.SetEnabled(false)

	mw.mode = mode
	if mw.mode == MODE_INVALID {
		return
	}

	if mw.mode == MODE_PLAY {
		mw.uiConvirm.SetEnabled(true)
		return
	}
	if mw.mode == MODE_COMPOSE {
	}
}

func (mw *MainWindow) getPoseInfo() (int, int) {
	totalFrame := len(imgList)
	poseCnt := int(mw.uiPoseCnt.Value())
	if poseCnt <= 0 {
		poseCnt = 1
	}

	if poseCnt >= totalFrame {
		return 1, totalFrame
	}
	if totalFrame%poseCnt != 0 {
		return 1, totalFrame
	}
	return poseCnt, totalFrame / poseCnt
}

func (mw *MainWindow) composeImg(fullname string) {
	poseCnt, frame := mw.getPoseInfo()
	if frame == 0 {
		return
	}

	var result draw.Image
	sw := boundary.Dx()
	sh := boundary.Dy()

	//var rgba bool
	_newBound := image.Rect(0, 0, sw*frame, sh*poseCnt)
	firstImg := imgList[0].img
	switch firstImg.(type) {
	case *image.RGBA:
		result = image.NewRGBA(_newBound)
		//rgba = true
	case *image.NRGBA:
		result = image.NewNRGBA(_newBound)
		//rgba = false
	default:
		return
	}

	singleBound := image.Rect(0, 0, sw, sh)
	for i, _img := range imgList {
		_subImg := _img.img.SubImage(boundary)
		col := i % frame
		row := i / frame
		drawBound := singleBound.Add(image.Point{sw * col, sh * row})
		draw.Draw(result, drawBound, _subImg, _subImg.Bounds().Min, draw.Src)
	}
	// Modify stride

	/*
		if rgba {
			result.(*image.RGBA).Stride = 8
			println(result.(*image.RGBA).Stride)
		} else {
			result.(*image.NRGBA).Stride = 8
			println(result.(*image.NRGBA).Stride)
		}
	*/

	f, err := os.OpenFile(fullname, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
		return
	}
	defer f.Close()
	f.Truncate(0)
	png.Encode(f, result)
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
	openAction.Triggered().Attach(func() { mw.openImage(MODE_COMPOSE) })
	fileMenu.Actions().Add(openAction)
	mw.ToolBar().Actions().Add(openAction)

	///
	// Load
	loadAction := walk.NewAction()
	// openAction.SetImage(openBmp)
	loadAction.SetText("&Load")
	loadAction.Triggered().Attach(func() { mw.openImage(MODE_PLAY) })
	fileMenu.Actions().Add(loadAction)
	mw.ToolBar().Actions().Add(loadAction)

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
	composeAction.Triggered().Attach(func() { mw.saveImage() })
	fileMenu.Actions().Add(composeAction)
	mw.ToolBar().Actions().Add(composeAction)

	// Exit
	exitAction := walk.NewAction()
	exitAction.SetText("E&xit")
	exitAction.Triggered().Attach(func() { walk.App().Exit(0) })
	fileMenu.Actions().Add(exitAction)
}

func (mw *MainWindow) initCanvas() {
	walk.NewHSpacer(mw)
	iv, _ := selfWidget.NewMyImageView(mw)
	mw.imageView = iv
}
func (mw *MainWindow) initOtherBars() {
	sp, _ := walk.NewSplitter(mw)
	sp.SetSize(walk.Size{400, 20})
	sp.SetOrientation(walk.Horizontal)

	walk.NewHSpacer(sp)

	lab, _ := walk.NewLabel(sp)
	lab.SetSize(walk.Size{16, 30})
	// lab.SetText("Pose")

	// others
	mw.uiFrameCnt, _ = walk.NewNumberEdit(sp)
	//mw.uiFrameCnt.SetSize(walk.Size{42, TB_H})
	mw.uiFrameCnt.SetRange(1, 100)
	mw.uiFrameCnt.SetDecimals(0)
	mw.uiFrameCnt.SetValue(8)
	mw.uiFrameCnt.SetEnabled(false)
	mw.uiFrameCnt.SetToolTipText(ttPlayPose)

	// lab, _ := walk.NewLabel(sp)
	// lab.SetSize(walk.Size{16, 30})
	// lab.SetText("Pose")
	mw.uiPoseCnt, _ = walk.NewNumberEdit(sp)
	//mw.uiPoseCnt.SetSize(walk.Size{42, TB_H})
	mw.uiPoseCnt.SetRange(1, 100)
	mw.uiPoseCnt.SetValue(1)
	mw.uiPoseCnt.SetDecimals(0)
	mw.uiPoseCnt.SetToolTipText(ttPosCnt)

	mw.uiConvirm, _ = walk.NewPushButton(sp)
	mw.uiConvirm.SetText("OK")
	mw.uiConvirm.Clicked().Attach(func() {
		// Get some fresh data.
		mw.onClickSave()
	})

	walk.InitWidget(sp, mw, FREEZEIZE_CLASS,
		winapi.CCS_NORESIZE,
		winapi.WS_EX_TOOLWINDOW|winapi.WS_EX_WINDOWEDGE)
}

func newMainWindow() {
	walk.SetPanicOnError(true)
	mainWnd, _ := walk.NewMainWindow()

	mw := &MainWindow{MainWindow: mainWnd}
	mw.SetLayout(walk.NewVBoxLayout())
	mw.SetTitle("Image composer")

	mw.initMenu()
	mw.initOtherBars()
	mw.initCanvas()

	mw.SetMinMaxSize(walk.Size{800, 600}, walk.Size{})
	mw.SetSize(walk.Size{800, 600})

	mw.refreshToolBar(MODE_INVALID)
	mw.Show()
	mw.Run()
}

func init() {
	walk.MustRegisterWindowClass(FREEZEIZE_CLASS)
}

func main() {
	newMainWindow()
}
