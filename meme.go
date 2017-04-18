package main

import (
    "strings"
    "os"
    "image"

    "github.com/fogleman/gg"
)

func generateMeme(sourcePath string, outputPath string, text string) (string, error) {
    const fontSize = 36

    file, err := os.Open(sourcePath)
    if err != nil {
        return "", err
    }

    // get our image object and create a new context from it
    img, _, err := image.Decode(file)
    r := img.Bounds()
    w := r.Dx()
    h := r.Dy()
    m := gg.NewContext(w, h)
    m.DrawImage(img, 0, 0)
    m.LoadFontFace("Impact.ttf", fontSize)

    // Apply black stroke
    m.SetHexColor("#000")
    strokeSize := 4

    // prep our text for drawing
    var topText = ""
    var bottomText = ""

    // we need to separate the text into top text and bottom text. we can do this by finding a mid-ish-point
    // long words can skew this, but OH WELL, maybe for another day.
    textFields := strings.Fields(text)
    textLength := len(textFields)
    if textLength > 3 {
        // find text midpoint
        midpoint := (textLength/2)

        topText = strings.Join(textFields[:midpoint], " ")
        bottomText = strings.Join(textFields[midpoint:], " ")
    } else {
        bottomText = text
    }

    // wrap our text to the width of the image so that we can read our super funny haha jokes
    wrappedTopText := m.WordWrap(topText, float64(w))
    wrappedBottomText := m.WordWrap(bottomText, float64(w))

    // draw the text shadows
    for dy := -strokeSize; dy <= strokeSize; dy++ {
        for dx := -strokeSize; dx <= strokeSize; dx++ {
            // give it rounded corners
            if dx*dx+dy*dy >= strokeSize*strokeSize {
                continue
            }

            // draw top text shadows
            if len(wrappedTopText) != 0 {
                topX := float64(w/2 + dx)
                topY := float64(fontSize + dy)
                for i := range wrappedTopText { // draw our top text. always draw below the previous line using our font size to calculate the Y pos
                    m.DrawStringAnchored(wrappedTopText[i], topX, topY + float64((i * fontSize)), 0.5, 0.5)
                }
            }

            // draw bottom text shadows
            if len(wrappedBottomText) != 0 {
                bottomX := float64(w/2 + dx)
                bottomY := float64(h - fontSize + dy)
                for i := range wrappedBottomText { // draw our top text. always draw below the previous line using our font size to calculate the Y pos
                    m.DrawStringAnchored(wrappedBottomText[i], bottomX, bottomY - float64(((len(wrappedBottomText) - 1 - i) * fontSize)), 0.5, 0.5)
                }
            }
        }
    }

    // draw the white text fill
    m.SetHexColor("#FFF")
    if len(wrappedTopText) != 0 {
        topX := float64(w/2)
        topY := float64(fontSize)
        for i := range wrappedTopText { // draw our top text. always draw below the previous line using our font size to calculate the Y pos
            m.DrawStringAnchored(wrappedTopText[i], topX, topY + float64((i * fontSize)), 0.5, 0.5)
        }
    }
    if len(wrappedBottomText) != 0 {
        bottomX := float64(w/2)
        bottomY := float64(h - fontSize)
        for i := range wrappedBottomText { // draw our top text. always draw below the previous line using our font size to calculate the Y pos
            m.DrawStringAnchored(wrappedBottomText[i], bottomX, bottomY - float64(((len(wrappedBottomText) - 1 - i) * fontSize)), 0.5, 0.5)
        }
    }
    m.SavePNG(outputPath)

    return outputPath, nil
}
