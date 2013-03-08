// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mywidget

import (
	. "github.com/lxn/walk"
)

type (
	MyImageView struct {
		*CustomWidget
		image Image

		x, y, w, h int
		boundPen   *CosmeticPen
	}
)

func NewMyImageView(parent Container) (*MyImageView, error) {
	iv := &MyImageView{}

	cw, err := NewCustomWidget(parent, 0, func(canvas *Canvas, updateBounds Rectangle) error {
		return iv.drawImage(canvas, updateBounds)
	})
	if err != nil {
		return nil, err
	}

	iv.CustomWidget = cw

	// iv.widget = iv

	iv.SetInvalidatesOnResize(true)

	return iv, nil
}

func (iv *MyImageView) SetImage(value Image) error {
	iv.image = value

	_, isMetafile := value.(*Metafile)
	iv.SetClearsBackground(isMetafile)

	return iv.Invalidate()
}

func (iv *MyImageView) SetBoundary(x, y, w, h int) {
	iv.x = x
	iv.y = y
	iv.w = w
	iv.h = h
}

func (mw *MyImageView) drawBoundary(canvas *Canvas, updateBounds Rectangle) error {
	if mw.w == 0 && mw.h == 0 {
		return nil
	}

	if mw.boundPen == nil {
		mw.boundPen, _ = NewCosmeticPen(PenSolid, RGB(255, 255, 255))
		return nil
	}
	canvas.DrawRectangle(mw.boundPen, Rectangle{mw.x, mw.y, mw.w, mw.h})
	return nil
}

func (iv *MyImageView) drawImage(canvas *Canvas, updateBounds Rectangle) error {
	if iv.image == nil {
		return nil
	}
	canvas.DrawImage(iv.image, Point{0, 0})

	return iv.drawBoundary(canvas, updateBounds)
}
