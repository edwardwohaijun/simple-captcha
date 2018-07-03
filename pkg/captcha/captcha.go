package captcha

import (
	"math"
	"image"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/draw"
	"math/rand"
	"image/color"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"encoding/base64"
	"os"
	"log"
	"image/png"
	"bytes"
	"time"
	"errors"
	"strings"
	"path/filepath"
	"io/ioutil"
	"github.com/golang/freetype"
	"strconv"
	"path"
	"fmt"
)

var	captchaW = 140
var	captchaH = 40 // final captcha image width/height. If you shrink these 2 values, don't forget to shrink the font size
var ttfList = make([]*truetype.Font, 0, 8)
var charSet = ""

// Initialise read all the true type fonts in the given fontDir, and set the character set(to be randomly chosen from)
func Initialise(fontDir string, characters string) error {
	if len(ttfList) > 0 && len(charSet) >= 6 { // get called already
		return nil
	}
	if len(characters) < 6 {
		return errors.New("character set is too small, it should be at lease 6 characters")
	}
	charSet = characters

	fonts, err := os.Open(fontDir)
	if err != nil {
		return errors.New("failed to open " + fontDir + " folder.")
	}
	defer fonts.Close()

	fontFiles, err := fonts.Readdir(-1)
	if err != nil {
		return errors.New("failed to read " + fontDir + " folder.")
	}
	if len(fontFiles) == 0 {
		return errors.New(fontDir + " is empty.")
	}
	for _, fontFile := range fontFiles {
		if fontFile.IsDir() {
			continue
		}
		if strings.ToLower(filepath.Ext(fontFile.Name())) != ".ttf" {
			log.Println(fontFile.Name() + " is not a true type font, skip.")
			continue
		}

		fontBytes, err := ioutil.ReadFile(path.Join(fontDir, fontFile.Name()))
		if err != nil {
			return errors.New("failed to read font file: " + fontFile.Name())
		}
		f, err := freetype.ParseFont(fontBytes)
		if err != nil {
			return errors.New("failed to parse font file: " + fontFile.Name())
		}
		ttfList = append(ttfList, f)
	}
	if len(ttfList) == 0 {
		return errors.New("no true type font found")
	}
	log.Println(strconv.Itoa(len(ttfList)) + " true type fonts found")
	return nil
}

// rotatePoint rotate the given x/y point by radian around the origin point(x0, y0), and return the new point coordinate
func rotatePoint(x, y, x0, y0, radian float64 ) (int, int){
	sin, cos := math.Sincos(radian)
	dx, dy := x - x0, y - y0
	return int(cos * dx - sin * dy + x0), int(sin * dx + cos * dy + y0)
}

// Character is drawn within 80x80 rectangle, thus there are many white space surrounding them.
// To make them stick together in the final image, I need to crop each character image to fit into the minimum wrapping rectangle.
// minXY, maxXY is the top-left, bottom-right point of this rectangle.
type captchaChar struct {
	character string
	im 		*image.RGBA
	minX, minY, maxX, maxY int
}

// newCharacter create a new captchaChar value(declared by captchaChar struct), and send it through a returned channel.
// During the creation, the character get rotated randomly between +- 20 degree, fontSize varies from 16~32, fontFace chosen randomly.
func newCharacter(character string, result chan<- *captchaChar) {
	f := ttfList[rand.Intn(len(ttfList))]

	fontSize := float64(rand.Intn(16) + 24) // don't set font size bigger than this, otherwise they might not fit inside the final image
	fontColor := color.RGBA{255, 109, 0, 255} // orange
	// fontColor := color.RGBA{233, 30, 99, 255} // pink
	canvasWidth, canvasHeight := 60, 60 // should be 20% bigger than the characterWidth/Height, because when character get transformed, their bounding box will get bigger.

	bgImg := image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))
	draw.Draw(bgImg, bgImg.Bounds(), image.Transparent, image.ZP, draw.Over)

	face := truetype.NewFace(f, &truetype.Options{
		Size: fontSize,
		DPI: 72,
	});

	// pre-calculate the width/height of the to-be-drawn-character
	bound, _ := font.BoundBytes(face, []byte(character))
	charWidth, charHeight := int(bound.Max.X>>6 - bound.Min.X>>6), int(bound.Max.Y>>6 - bound.Min.Y>>6)

	drawingDot := fixed.Point26_6{
		fixed.Int26_6(((canvasWidth - charWidth)/2)<<6),
		fixed.Int26_6(((canvasHeight - charHeight)/2 + charHeight)<<6),
	}

	d := &font.Drawer{
		Dst:  bgImg,
		Src:  image.NewUniform(fontColor),
		Face: face,
		Dot: drawingDot,
	}
	d.DrawString(character)

	charImg := image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))
	imgWidth := charImg.Bounds().Dx()

	rotateDeg := float64(rand.Intn(40) - 20)
	b := charImg.Bounds()
	minX, minY, maxX, maxY := b.Max.X, b.Max.Y, b.Min.X, b.Min.Y // to calculate the minimum wrapping rectangle
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			newX, newY := rotatePoint(float64(x), float64(y), float64(canvasWidth/2), float64(canvasHeight/2), rotateDeg * math.Pi/180)
			if newX < canvasWidth && newX >= 0 && newY < canvasHeight && newY >= 0 {
				if bgImg.Pix[(newY * imgWidth + newX)*4+0] == 0 && bgImg.Pix[(newY * imgWidth + newX)*4+1] == 0 && bgImg.Pix[(newY * imgWidth + newX)*4+2] == 0 {
					continue // skip the transparent area.
				}
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}

				charImg.Pix[(y * imgWidth + x)*4 + 0] = bgImg.Pix[(newY * imgWidth + newX)*4+0]
				charImg.Pix[(y * imgWidth + x)*4 + 1] = bgImg.Pix[(newY * imgWidth + newX)*4+1]
				charImg.Pix[(y * imgWidth + x)*4 + 2] = bgImg.Pix[(newY * imgWidth + newX)*4+2]
				charImg.Pix[(y * imgWidth + x)*4 + 3] = bgImg.Pix[(newY * imgWidth + newX)*4+3]
			}
		}
	}
	// grow the minimum wrapping rectangle by 1 pixel
	if minX > 0 {
		minX--
	}
	if minY > 0 {
		minY--
	}
	if maxX < canvasWidth - 1 {
		maxX++
	}
	if maxY < canvasHeight - 1{
		maxY++
	}

	result<- &captchaChar{
		character:character,
		im: charImg,
		minX: minX,
		minY: minY,
		maxX: maxX,
		maxY: maxY,
	}
}

// Merge pieces together all the individual character images into a final image.
// And position their drawing point randomly
func merge(imgs []*captchaChar) (string, *image.RGBA) {
	var result = image.NewRGBA(image.Rect(0, 0, captchaW, captchaH))
	// todo: set the img's bg here, add some noise, sin/cos curve.

	totalW := 0
	for _, img := range imgs {
		totalW += img.maxX - img.minX
	}
	//var drawAt = image.Rect(0, 0, 0, 0)
	var drawAt = image.Rect(0, 0, captchaW, captchaH)
	if totalW < captchaW {
		drawAt.Min.X = rand.Intn(captchaW - totalW)
	} else {
		fmt.Println("warning: total character width is bigger than final image width, last character might get cut off")
	}

	var text = make([]byte, 0, len(imgs))
	for _, img := range imgs {
		text = append(text, strings.ToLower(img.character)[0])

		if (img.maxY - img.minY) < captchaH {
			drawAt.Min.Y = rand.Intn(captchaH - (img.maxY - img.minY))
		} else {
			drawAt.Min.Y = 0
		}
		drawAt.Max.Y = drawAt.Min.Y + (img.maxY - img.minY)
		draw.Draw(result, drawAt, img.im, image.Pt(img.minX, img.minY), draw.Over)
		drawAt.Min.X += img.maxX - img.minX - (rand.Intn(4) + 3) // make 2 character tie closer
	}
	return string(text), result
}

// New return the text and binary image or base64 string.
// For base64 string, caller must compose the final string for web use, e.g., imgTag := "<img src=\"data:image/png;base64," + base64string + "\" />"
// Caller must call Initialise(fontDir, charSet) first(and once only) before calling New.
func New(isBase64 bool) (text string, img *image.RGBA, base64string string, err error) {
	if ttfList == nil {
		return "", nil, "", errors.New("no true type font found, make sure you have called Initialise(fontDir) first")
	}

	rand.Seed(time.Now().UTC().UnixNano())
	var numChar = rand.Intn(3) + 4
	var result = make(chan *captchaChar, numChar)

	for i :=0; i<numChar; i++ {
		randomChar := charSet[rand.Intn(len(charSet))]
		go newCharacter(string(randomChar), result)
	}
	var charImg = make([]*captchaChar, numChar)
	for i := 0; i<numChar; i++ {
		charImg[i] = <-result
	}
	var captchaTxt, captchaImg = merge(charImg)

	if isBase64 {
		var b bytes.Buffer
		png.Encode(&b, captchaImg)
		base64string = base64.StdEncoding.EncodeToString(b.Bytes())
		return captchaTxt, nil, base64string, nil
	}
	return captchaTxt, captchaImg, "", nil
}
