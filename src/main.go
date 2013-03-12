package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/lxn/go-winapi"
	"github.com/lxn/walk"

	// Self
	selfWidget "mywidget"
)

type (
	ImageExt interface {
		ColorModel() color.Model
		Bounds() image.Rectangle
		At(x, y int) color.Color

		// (x, y, stride int)

		SubImage(image.Rectangle) image.Image
	}
	MainWindow struct {
		*walk.MainWindow
		viewGrid  *walk.GridLayout
		imageView [POSE_CNT_MAX]*selfWidget.MyImageView

		// Other ui
		uiFrameCnt *walk.NumberEdit
		uiPoseCnt  *walk.NumberEdit
		// uiPlayPose      *walk.NumberEdit
		uiConvirm       *walk.PushButton
		uiComposeAction *walk.Action
		mode            int

		refreshTimer *time.Ticker
	}

	// Image struct
	ImageItem struct {
		fname string
		bm    *walk.Bitmap
		img   ImageExt
	}
)

const (
	GRID_CNT     = 10
	TB_H     int = 16 // Tool bar size
	OB_H     int = 12 // Other bar size

	POSE_CNT_MAX int = 8

	FREEZEIZE_CLASS string = "FREEZEIZE_CLASS0"

	ttPosCnt   string = "动作个数"
	ttPlayPose string = "播放动作"

	// MODE
	MODE_COMPOSE int = 0
	MODE_PLAY    int = 1
	MODE_INVALID int = 3
)

var (
	screenW, screenH int

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

func (mw *MainWindow) readPoseImage(path, ext string) {
	mw.resetImageList()
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
			if modelItem != nil {
				modelItem.bm.Dispose()
			}
			modelItem = newImg

		}
	}
}

func (mw *MainWindow) resetImageList() {
	if mw.refreshTimer != nil {
		mw.refreshTimer.Stop()
		mw.refreshTimer = nil
	}

	for i := 0; i < POSE_CNT_MAX; i++ {
		mw.imageView[i].SetImage(nil)
	}

	for _, v := range imgList {
		v.bm.Dispose()
	}
	imgList = _iniImgList[0:0]
}

func (mw *MainWindow) readImageList(path, ext string) error {
	folder, err := os.Open(path)
	if err != nil {
		return err
	}

	// Read all the files in the folder
	fileList, err := folder.Readdir(-1)
	if err != nil {
		return err
	}
	mw.resetImageList()
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

func (mw *MainWindow) openImage(mode int) { //
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

	boundary = image.Rect(-1, -1, -1, -1)

	if fs.IsDir() {
		mw.SetTitle(folderPath)
		mw.readImageList(folderPath, ".png")
		mw.setImageSize()
		mw.refreshToolBar(MODE_COMPOSE)
		mw.initFrame()
		return
	}
	mw.SetTitle(folderPath)
	mw.readPoseImage(folderPath, ".png")
	mw.refreshToolBar(MODE_PLAY)
	mw.initPoseInfo()
	mw.setImageSize()
	// mw.onUiSetFrameCnt()

	mw.initFrame()
}

func (mw *MainWindow) saveImage() {
	title := "Save the image"
	filter := "Image Files (*.png)|*.png"
	path := selfWidget.GetSaveFileName(mw, title, filter)
	if len(path) == 0 {
		return
	}
	println(path)
	if !strings.HasSuffix(path, ".png") {
		path += ".png"
	}
	mw.composeImg(path)
}

func (mw *MainWindow) drawImage() {
	l := len(imgList)
	if l == 0 {
		return
	}

	poseCnt := mw.getPoseCnt()

	f := currentFrame % frameCount
	currentFrame++

	for i := 0; i < poseCnt; i++ {
		curFrame := f + frameCount*i
		if curFrame < len(imgList) {
			mw.imageView[i].SetImage(imgList[curFrame].bm)
		}
	}
}

// Calc the image size, draw the image boundary
func getLineCnt() int {
	w := screenW - 80
	switch {
	case imageW > w/2:
		return 1
	case imageW > w/3:
		return 2
	case imageW > w/4:
		return 3
	default:
		return 4
	}
	return 1
}
func (mw *MainWindow) setImageSize() {
	lc := getLineCnt()
	gridW := 800 / GRID_CNT
	gridH := 600 / GRID_CNT

	w := imageW/gridW + 1
	h := imageH/gridH + 1

	i := 0
	for ; i < mw.getPoseCnt(); i++ {
		mw.imageView[i].SetSize(walk.Size{imageW, imageH})

		mw.imageView[i].SetBoundary(boundary.Min.X, boundary.Min.Y,
			boundary.Dx(), boundary.Dy())

		mw.imageView[i].SetVisible(true)
		x := (i % lc) * w
		y := (i / lc) * h
		mw.viewGrid.SetRange(mw.imageView[i], walk.Rectangle{x, y, w, h})
	}

	for ; i < POSE_CNT_MAX; i++ {
		mw.imageView[i].SetVisible(false)
		mw.imageView[i].Invalidate()
	}

	fmt.Printf("Boundary %v\n", boundary)
}

func (mw *MainWindow) initFrame() {
	mw.uiPoseCnt.SetValue(float64(mw.getPoseCnt()))
	if mw.refreshTimer != nil {
		return
	}
	mw.refreshTimer = time.NewTicker(time.Millisecond * 83)
	go func() {
		for _ = range mw.refreshTimer.C {
			mw.drawImage()
		}

	}()
}

func (mw *MainWindow) getPoseCnt() int {
	l := len(imgList)
	r := l / frameCount
	if r == 0 {
		return 1
	}
	if r > POSE_CNT_MAX {
		return POSE_CNT_MAX
	}
	return r
}

func (mw *MainWindow) initPoseInfo() {
	if modelItem == nil {
		return
	}

	model := modelItem.img
	if model == nil {
		return
	}
	// Get the pose count
	poseCnt := 1
	x := model.Bounds().Min.X
	y := model.Bounds().Min.Y
	w := model.Bounds().Dx()
	h := model.Bounds().Dy()
	for i := 2; i < POSE_CNT_MAX; i++ {
		sh := h / i
		for j := 1; j < i; j++ {
			beginY := sh * j
			// Erase the boundary by 1 pix to handel the neighbor pix
			for z := 1; z < w-1; z++ {
				_, _, _, a := model.At(x+z, y+beginY).RGBA()
				if a != 0 {
					_, _, _, la := model.At(x+z-1, y+beginY).RGBA()
					_, _, _, ra := model.At(x+z+1, y+beginY).RGBA()
					_, _, _, ta := model.At(x+z, y+beginY-1).RGBA()
					_, _, _, da := model.At(x+z, y+beginY+1).RGBA()
					if la != 0 && ra != 0 && ta != 0 && da != 0 {
						fmt.Println("Pose alpha:", x+z, y+beginY, a)
						goto nextPose
					}
				}
			}
		}
		poseCnt = i
		break

	nextPose:
	}

	fmt.Println("Pose count: ", poseCnt)

	// Init the pose list
	imageW = w / frameCount
	imageH = h / poseCnt

	mw.resetImageList()
	boundary = image.Rect(0, 0, imageW, imageH)
	tmpBound := boundary
	// Read all png images
	for i := 0; i < poseCnt; i++ {
		for j := 0; j < frameCount; j++ {
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

}

/*
func (mw *MainWindow) onUiSetFrameCnt() {
	if modelItem == nil {
		return
	}

	// imageW = modelItem.img.Bounds().Dx()
	// imageH = modelItem.img.Bounds().Dy()

	// poseCnt := mw.getPoseCnt()
	playPose = int(mw.uiPlayPose.Value())
	mw.setImageSize()
}*/

func (mw *MainWindow) refreshToolBar(mode int) {
	mw.uiConvirm.SetEnabled(false)
	mw.uiComposeAction.SetEnabled(false)
	mw.uiPoseCnt.SetEnabled(false)

	mw.mode = mode
	if mw.mode == MODE_INVALID {
		return
	}

	if mw.mode == MODE_PLAY {
		return
	}
	if mw.mode == MODE_COMPOSE {
		mw.uiComposeAction.SetEnabled(true)
	}
}

func (mw *MainWindow) getPoseInfo() (int, int) {
	totalFrame := len(imgList)
	poseCnt := mw.getPoseCnt()

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
	case *image.RGBA64:
		result = image.NewRGBA64(_newBound)
	case *image.NRGBA:
		result = image.NewNRGBA(_newBound)
	case *image.NRGBA64:
		result = image.NewNRGBA64(_newBound)
	default:
		fmt.Println("image type: ", reflect.TypeOf(firstImg))
		println("Unsupported image type")
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

func setIcon(ui *walk.Action, fname string) {
	fpath := "./img/" + fname
	_, err := os.Stat(fpath)
	if err != nil {
		fmt.Println(err)
		return
	}
	img, _ := walk.NewBitmapFromFile(fpath)
	ui.SetImage(img)
}

func (mw *MainWindow) initMenu() {
	fileMenu, _ := walk.NewMenu()
	fileMenuAction, _ := mw.Menu().Actions().AddMenu(fileMenu)
	fileMenuAction.SetText("&File")

	imageList, _ := walk.NewImageList(walk.Size{TB_H, TB_H}, 0)
	mw.ToolBar().SetImageList(imageList)

	openAction := walk.NewAction()
	setIcon(openAction, "open.png")
	openAction.SetText("&Open")
	openAction.Triggered().Attach(func() { go mw.openImage(MODE_COMPOSE) })
	fileMenu.Actions().Add(openAction)
	mw.ToolBar().Actions().Add(openAction)

	///
	// Load
	loadAction := walk.NewAction()
	setIcon(loadAction, "load.png")
	loadAction.SetText("&Load")
	loadAction.Triggered().Attach(func() { mw.openImage(MODE_PLAY) })
	fileMenu.Actions().Add(loadAction)
	mw.ToolBar().Actions().Add(loadAction)

	helpMenu, _ := walk.NewMenu()
	helpMenuAction, _ := mw.Menu().Actions().AddMenu(helpMenu)
	helpMenuAction.SetText("&Help")

	aboutAction := walk.NewAction()
	helpMenu.Actions().Add(aboutAction)
	aboutAction.SetText("&About")
	aboutAction.Triggered().Attach(func() {
		walk.MsgBox(mw, "About", "Image composer V0.1\nAuthor:heml",
			walk.MsgBoxOK|walk.MsgBoxIconInformation)
	})

	// Image operations
	// Save
	mw.uiComposeAction = walk.NewAction()
	setIcon(mw.uiComposeAction, "save.png")
	mw.uiComposeAction.SetText("&Save")
	mw.uiComposeAction.Triggered().Attach(func() { go mw.saveImage() })
	fileMenu.Actions().Add(mw.uiComposeAction)
	mw.ToolBar().Actions().Add(mw.uiComposeAction)

	// Exit
	exitAction := walk.NewAction()
	exitAction.SetText("E&xit")
	exitAction.Triggered().Attach(func() { walk.App().Exit(0) })
	fileMenu.Actions().Add(exitAction)
}

func (mw *MainWindow) initCanvas() {
	for i := 0; i < POSE_CNT_MAX; i++ {
		iv, _ := selfWidget.NewMyImageView(mw)
		mw.imageView[i] = iv
	}
}
func (mw *MainWindow) initOtherBars() {
	sp, _ := walk.NewSplitter(mw)
	sp.SetSize(walk.Size{400, 20})

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
		// mw.onUiSetFrameCnt()
	})

	walk.InitWidget(sp, mw, FREEZEIZE_CLASS,
		winapi.CCS_NORESIZE,
		winapi.WS_EX_TOOLWINDOW|winapi.WS_EX_WINDOWEDGE)
}

func newMainWindow() {
	walk.SetPanicOnError(true)
	mainWnd, _ := walk.NewMainWindow()

	mw := &MainWindow{MainWindow: mainWnd}
	mw.viewGrid = walk.NewGridLayout()
	mw.SetLayout(mw.viewGrid)
	mw.viewGrid.SetRowStretchFactor(GRID_CNT, 2)
	mw.viewGrid.SetColumnStretchFactor(GRID_CNT, 2)
	mw.viewGrid.SetMargins(walk.Margins{6, 28, 2, 6})

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
	runtime.GOMAXPROCS(2)

	screenW = int(winapi.GetSystemMetrics(winapi.SM_CXSCREEN))
	screenH = int(winapi.GetSystemMetrics(winapi.SM_CYSCREEN))
}

func main() {
	newMainWindow()
}
