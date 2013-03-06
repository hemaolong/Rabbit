package main

import (
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/lxn/go-winapi"
	"github.com/lxn/walk"

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
		paintWidget  *walk.CustomWidget
		prevFilePath string

		// Other ui
		uiFrameCnt *walk.NumberEdit
		uiPoseCnt  *walk.LineEdit
	}

	// Image struct
	ImageItem struct {
		fname string
		bm    *walk.Bitmap
		img   image.Image
	}
)

var (
	_iniImgList  [1000]*ImageItem
	imgList      = _iniImgList[0:0]
	currentFrame int
	imageW       int
	imageH       int

	boundary image.Rectangle = image.Rect(0, 0, 0, 0)

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
			if a == 0 {
				_p := image.Point{i, j}
				if !_p.In(boundary) {
					boundary = boundary.Union(image.Rect(i, j, i, j))
				}
			}
		}
	}
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
				imgList = append(imgList)
				fullname := path + "/" + fname
				if bm, err := walk.NewBitmapFromFile(fullname); err == nil {
					newImg := new(ImageItem)
					newImg.fname = fullname
					newImg.bm = bm
					newImg.img, err = readImge(fullname)
					if err == nil {
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
	if err := ReadImageList(folderPath, ".png"); err != nil {
		return
	}
	mw.setImageSize()
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
}

func (mw *MainWindow) drawBoundary(canvas *walk.Canvas, updateBounds walk.Rectangle) error {
	// bmp, _ := walk.NewBitmap(walk.Size{imageW, imageH})
	// canvas, _ := walk.NewCanvasFromImage(bmp)

	rectPen, _ := walk.NewCosmeticPen(walk.PenSolid, walk.RGB(255, 255, 255))
	if boundary.Empty() {
		return nil
	}

	x := boundary.Min.X
	y := boundary.Min.Y
	w := boundary.Dx()
	h := boundary.Dy()

	canvas.DrawRectangle(rectPen, walk.Rectangle{x, y, w, h})

	println(x)
	println(y)
	println(w)
	println(h)

	defer rectPen.Dispose()
	println("draw")
	return nil
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

	// Exit
	exitAction := walk.NewAction()
	exitAction.SetText("E&xit")
	exitAction.Triggered().Attach(func() { walk.App().Exit(0) })
	fileMenu.Actions().Add(exitAction)
}

func (mw *MainWindow) initCanvas() {
	comp, _ := walk.NewComposite(mw)
	// comp.Children().observer = mw
	// Init image view
	iv, _ := walk.NewImageView(comp)
	mw.imageView = iv
	mw.initFrame()

	mw.paintWidget, _ = walk.NewCustomWidget(comp, 0,
		func(canvas *walk.Canvas, updateBounds walk.Rectangle) error {
			return mw.drawBoundary(canvas, updateBounds)
		})
	mw.paintWidget.SetSize(walk.Size{600, 600})
	mw.paintWidget.SetClearsBackground(true)
	mw.paintWidget.SetInvalidatesOnResize(true)
}
func (mw *MainWindow) initOtherBars() {
	sp, _ := walk.NewSplitter(mw)
	sp.SetSize(walk.Size{400, TB_H})
	walk.NewHSpacer(sp)
	walk.NewHSpacer(sp)
	// others
	mw.uiFrameCnt, _ = walk.NewNumberEdit(sp)
	mw.uiFrameCnt.SetSize(walk.Size{32, TB_H})
	mw.uiFrameCnt.SetRange(1, 100)
	mw.uiFrameCnt.SetDecimals(0)

	walk.NewSplitter(sp)

	lab, _ := walk.NewLabel(sp)
	lab.SetSize(walk.Size{16, 30})
	lab.SetText("Pose")

	mw.uiPoseCnt, _ = walk.NewLineEdit(sp)
	mw.uiPoseCnt.SetSize(walk.Size{32, 30})

	// Space

	walk.InitWidget(sp, mw, FREEZEIZE_CLASS,
		winapi.CCS_NORESIZE,
		winapi.WS_EX_TOOLWINDOW|winapi.WS_EX_WINDOWEDGE)

	walk.NewSplitter(mw)
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
