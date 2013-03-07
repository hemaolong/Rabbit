// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mywidget

import (
    "github.com/lxn/walk"
)

type ImageView struct {
	*walk.CustomWidget
	image Image
}

func NewImageView(parent Container) (*ImageView, error) {
	iv := &ImageView{}

	cw, err := NewCustomWidget(parent, 0, func(canvas *Canvas, updateBounds Rectangle) error {
		return iv.drawImage(canvas, updateBounds)
	})
	if err != nil {
		return nil, err
	}

	iv.CustomWidget = cw

	iv.widget = iv

	iv.SetInvalidatesOnResize(true)

	return iv, nil
}

func (iv *ImageView) SetImage(value Image) error {
	iv.image = value

	_, isMetafile := value.(*Metafile)
	iv.SetClearsBackground(isMetafile)

	return iv.Invalidate()
}

func (iv *ImageView) drawImage(canvas *Canvas, updateBounds Rectangle) error {
	if iv.image == nil {
		return nil
	}

	bounds := iv.ClientBounds()

	return canvas.DrawImageStretched(iv.image, bounds)
}
