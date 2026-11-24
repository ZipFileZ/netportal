/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package webui

import (
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"unsafe"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

var errNoQRCode = errors.New("no QR code found")

func decodeQR(img image.Image) (string, error) {
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", err
	}
	hints := map[gozxing.DecodeHintType]interface{}{gozxing.DecodeHintType_TRY_HARDER: true}
	res, err := qrcode.NewQRCodeReader().Decode(bmp, hints)
	if err != nil {
		return "", errNoQRCode
	}
	return res.GetText(), nil
}

func decodeQRFromFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return "", err
	}
	return decodeQR(img)
}

// captureScreen grabs the whole virtual desktop (all monitors).
func captureScreen() (*image.RGBA, error) {
	x := systemMetric(smXVirtualScreen)
	y := systemMetric(smYVirtualScreen)
	w := systemMetric(smCxVirtualScreen)
	h := systemMetric(smCyVirtualScreen)
	if w <= 0 || h <= 0 {
		return nil, errors.New("unable to determine screen size")
	}
	screenDC, _, _ := procGetDC.Call(0)
	if screenDC == 0 {
		return nil, errors.New("GetDC failed")
	}
	defer procReleaseDC.Call(0, screenDC)
	memDC, _, _ := procCreateCompatibleDC.Call(screenDC)
	if memDC == 0 {
		return nil, errors.New("CreateCompatibleDC failed")
	}
	defer procDeleteDC.Call(memDC)
	bmp, _, _ := procCreateCompatibleBitmap.Call(screenDC, uintptr(w), uintptr(h))
	if bmp == 0 {
		return nil, errors.New("CreateCompatibleBitmap failed")
	}
	defer procDeleteObject.Call(bmp)
	old, _, _ := procSelectObject.Call(memDC, bmp)
	r, _, _ := procBitBlt.Call(memDC, 0, 0, uintptr(w), uintptr(h), screenDC, uintptr(x), uintptr(y), srccopy|captureblt)
	procSelectObject.Call(memDC, old)
	if r == 0 {
		return nil, errors.New("BitBlt failed")
	}
	hdr := bitmapInfoHeader{
		Size:     uint32(unsafe.Sizeof(bitmapInfoHeader{})),
		Width:    w,
		Height:   -h, // top-down
		Planes:   1,
		BitCount: 32,
	}
	// BITMAPINFO = header + one RGBQUAD; reserve room for it.
	bi := struct {
		hdr    bitmapInfoHeader
		colors [4]uint32
	}{hdr: hdr}
	img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	if r, _, _ := procGetDIBits.Call(memDC, bmp, 0, uintptr(h), uintptr(unsafe.Pointer(&img.Pix[0])), uintptr(unsafe.Pointer(&bi)), dibRgbColors); r == 0 {
		return nil, errors.New("GetDIBits failed")
	}
	// BGRA -> RGBA
	for i := 0; i+3 < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+2] = img.Pix[i+2], img.Pix[i]
		img.Pix[i+3] = 0xff
	}
	return img, nil
}

func decodeQRFromScreen() (string, error) {
	img, err := captureScreen()
	if err != nil {
		return "", err
	}
	return decodeQR(img)
}
