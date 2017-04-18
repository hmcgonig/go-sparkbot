package main

import (
    "encoding/json"
    "fmt"
    "image"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strings"

    "github.com/fogleman/gg"
)

func handleMessage(message *SparkMessage) {
    messageContents := message.Text

    tmp := strings.Fields(messageContents)
    var command string
    var input string
    if (len(tmp) > 1) { // there was a command specified. index 0 will always be the tagged name of the bot.
        command = tmp[1]

        if (len(tmp) > 2) { // there was input after the command. index 1 is always the command.
            // cut the first two words from the message, join the rest back together with spaces
            input = strings.Join(tmp[2:], " ")
        }

    }

    log.Println(fmt.Sprintf("Recieved command: %s, input: %s", command, input))

    // handle our various commands
    switch(command) {
        case "help":
            newMessage := "Available Commands: \n" +
                "help: get a list of commands \n" +
                "good: get some good text \n" +
                "bad: :( \n" +
                "picard: generate a meme! ex. @Captain picard hey, i just made a meme!"
            sendMessageToSpark(&SparkMessage{RoomId: message.RoomId, Text:newMessage})
        case "good":
            sendMessageToSpark(&SparkMessage{RoomId: message.RoomId, Text:"good job!"})
            break
        case "bad":
            sendMessageToSpark(&SparkMessage{RoomId: message.RoomId, Text:"bad :("})
            break
        case "picard": // generate a meme image using the input and send it to spark
            outputPath := "out.jpg"
            generateMeme("picard.jpg", outputPath, input)
            sendImageToSpark(message.RoomId, outputPath)
            break
        default:
            break;
    }
}

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

func handleWebhook(w http.ResponseWriter, r *http.Request) {
    defer r.Body.Close()

    // parse the webhook post data to get the message id.
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Println(err.Error())
        return
    }

    var sparkWebhook SparkWebhook
    if err := json.Unmarshal(body, &sparkWebhook); err != nil {
        log.Println(err.Error())
        return
    }

    // use the message id to get a message struct containing more message info
    message, err := getMessage(sparkWebhook.Data.Id)
    if err != nil {
        log.Println(err.Error())
        return
    }

    // if the message isnt from our bot, handle it
    if message.PersonEmail != sparkBotEmail {
        handleMessage(&message)
    }

    fmt.Fprintf(w, "it wark.")
}

func main() {
    if len(os.Getenv("SPARK_BEARER_TOKEN")) == 0 {
        panic("Please set SPARK_BEARER_TOKEN environment variable.")
    }

    if len(os.Getenv("SPARK_BOT_EMAIL")) == 0 {
        panic("Please set SPARK_BOT_EMAIL environment variable.")
    }

    http.HandleFunc("/", handleWebhook)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", getenv("PORT", "8080")), nil))
}
