package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"time"
	"image/color"

	"github.com/lxn/go-winapi"
	"github.com/lxn/walk"

	// Self
	selfWidget "mywidget"
)

type ImageExt interface {
    ColorModel() color.Model
	// Bounds returns the domain for which At can return non-zero color.
	// The bounds do not necessarily contain the point (0, 0).
	Bounds() image.Rectangle
	// At returns the color of the pixel at (x, y).
	// At(Bounds().Min.X, Bounds().Min.Y) returns the upper-left pixel of the grid.
	// At(Bounds().Max.X-1, Bounds().Max.Y-1) returns the lower-right one.
	At(x, y int) color.Color

	Set(x, y int, c color.Color)

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
)

type (
	MainWindow struct {
		*walk.MainWindow
		imageView    *selfWidget.MyImageView
		prevFilePath string

		// Other ui
		uiFrameCnt *walk.NumberEdit
		uiPoseCnt  *walk.NumberEdit
		mode       int
	}

	// Image struct
	ImageItem struct {
		fname string
		bm    *walk.Bitmap
		img   ImageExt
	}
)

var (
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

func (mw *MainWindow) openImage() {
	folderPath := selfWidget.GetPath(0, 0)
	f, err := os.Open(folderPath)
	if err != nil {
		return
	}

	fs, err := f.Stat()
	if err != nil {
		return
	}

	if fs.IsDir() {
		readImageList(folderPath, ".png")
		mw.setImageSize()
		return
	}
	readPoseImage(folderPath, ".png")
	mw.setImageSize()
}

func (mw *MainWindow) saveImage(){
    path := selfWidget.GetSavePath(0)
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

	mw.imageView.SetBoundary(boundary.Min.X, boundary.Min.Y,
		boundary.Dx(), boundary.Dy())
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

func (mw *MainWindow) getPoseInfo() (int, int) {
	totalFrame := len(imgList)
	poseCnt := int(mw.uiPoseCnt.Value())
	if poseCnt <= 0 {
		return 1, totalFrame
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
	fmt.Printf("pose: %v, per frame %v\n", poseCnt, frame)

	var result draw.Image
	sw := boundary.Dx()
	sh := boundary.Dy()

	_newBound := boundary
	_newBound = _newBound.Add(image.Point{sw * (frame - 1), sh * poseCnt})
	firstImg := imgList[0].img
	switch firstImg.(type) {
	case *image.RGBA:
		result = image.NewRGBA(_newBound)
	case *image.NRGBA:
		result = image.NewNRGBA(_newBound)
	default:
		return
	}

	for _, _img := range imgList {
		_subImg := _img.img.SubImage(boundary)
		draw.Draw(result, boundary, _subImg, _subImg.Bounds().Min, draw.Src)
	}

	f, err := os.OpenFile(fullname, os.O_RDWR, os.ModePerm)
	if err != nil {
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
	openAction.Triggered().Attach(func() { mw.openImage() })
	fileMenu.Actions().Add(openAction)
	mw.ToolBar().Actions().Add(openAction)

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
	iv, _ := selfWidget.NewMyImageView(mw)
	mw.imageView = iv
	mw.initFrame()
}
func (mw *MainWindow) initOtherBars() {
	sp, _ := walk.NewSplitter(mw)
	sp.SetSize(walk.Size{400, TB_H})
	sp.SetOrientation(walk.Horizontal)

	walk.NewHSpacer(sp)

	lab, _ := walk.NewLabel(sp)
	lab.SetSize(walk.Size{16, 30})
	// lab.SetText("Pose")

	// others
	mw.uiFrameCnt, _ = walk.NewNumberEdit(sp)
	mw.uiFrameCnt.SetSize(walk.Size{42, TB_H})
	mw.uiFrameCnt.SetRange(1, 100)
	mw.uiFrameCnt.SetDecimals(0)
	mw.uiFrameCnt.SetToolTipText(ttPlayPose)

	// lab, _ := walk.NewLabel(sp)
	// lab.SetSize(walk.Size{16, 30})
	// lab.SetText("Pose")
	mw.uiPoseCnt, _ = walk.NewNumberEdit(sp)
	mw.uiPoseCnt.SetSize(walk.Size{42, TB_H})
	mw.uiPoseCnt.SetRange(1, 100)
	mw.uiPoseCnt.SetDecimals(0)
	mw.uiPoseCnt.SetToolTipText(ttPosCnt)

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
	mw.Show()
	mw.Run()
}

func init() {
	walk.MustRegisterWindowClass(FREEZEIZE_CLASS)
}

func main() {
	newMainWindow()
}
