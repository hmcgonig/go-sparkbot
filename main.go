package main

import (
    "net/http"
    "fmt"
    "io/ioutil"
    "encoding/json"
    "log"
    "errors"
    "os"
    "strings"
    "bytes"
    "mime/multipart"
    "path/filepath"
    "io"
    "net/textproto"

    "github.com/fogleman/gg"
    "image"
)

var sparkBearerToken = os.Getenv("SPARK_BEARER_TOKEN")
var sparkBotEmail = os.Getenv("SPARK_BOT_EMAIL")

func getMessage(messageId string) (SparkMessage, error) {
    var sparkMessage SparkMessage

    client := &http.Client{}

    // create the request
    req, err := http.NewRequest("GET", fmt.Sprintf("https://api.ciscospark.com/v1/messages/%s", messageId), nil)
    if err != nil {
        return sparkMessage, err
    }

    // add headers
    req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sparkBearerToken))

    // do the request
    resp, err := client.Do(req)
    defer resp.Body.Close()

    // error handling
    if resp.StatusCode != 200 {
        return sparkMessage, errors.New(fmt.Sprintf("Message request returned status code: %s", string(resp.StatusCode)))
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return sparkMessage, err
    }

    if err := json.Unmarshal(body, &sparkMessage); err != nil {
        return sparkMessage, err
    }

    return sparkMessage, nil
}

func sendMessageToSpark(message *SparkMessage) (error) {
    client := &http.Client{}

    // marshal the message
    body, err := json.Marshal(message)
    if err != nil {
        return err
    }

    log.Printf(string(body))

    // create the req object
    req, err := http.NewRequest("POST", "https://api.ciscospark.com/v1/messages", bytes.NewReader(body))
    if err != nil {
        return err
    }

    // add headers
    req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sparkBearerToken))
    req.Header.Add("Content-Type", "Application/json")

    // do the req
    resp, err := client.Do(req)
    defer resp.Body.Close()

    // error handling
    if resp.StatusCode != 200 {
        body, _ := ioutil.ReadAll(resp.Body)
        log.Printf(string(body))
        return errors.New(fmt.Sprintf("Message request returned status code: %s error body: %s", string(resp.StatusCode), string(body)))
    }

    return nil
}

func sendImageToSpark(roomId string, path string) (error) {
    client := &http.Client{}
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    // sniff the content type of the image using the first 512 bytes
    buffer := make([]byte, 512)
    _, err = file.Read(buffer)
    if err != nil {
        return err
    }

    // reset the read pointer
    file.Seek(0, 0)

    // attach the image the the request as a part of a multipart message
    h := make(textproto.MIMEHeader)
    h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s";`, "files", filepath.Base(path)))
    h.Set("Content-Type", http.DetectContentType(buffer))
    part, err := writer.CreatePart(h)
    if err != nil {
        return err
    }
    _, err = io.Copy(part, file)

    // copy the room id into another field
    writer.WriteField("roomId", roomId)

    // close the writer
    err = writer.Close()
    if err != nil {
        return err
    }

    // make the request object
    req, err := http.NewRequest("POST", "https://api.ciscospark.com/v1/messages", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sparkBearerToken))

    // do the req
    resp, err := client.Do(req)
    defer resp.Body.Close()

    // simple error handling
    if resp.StatusCode != 200 {
        body, _ := ioutil.ReadAll(resp.Body)
        log.Printf(string(body))
        return errors.New(fmt.Sprintf("Message request returned status code: %s error body: %s", string(resp.StatusCode), string(body)))
    }

    return nil
}

func handleMessage(message *SparkMessage) {
    messageContents := message.Text
    messageContents = "Captain asdasdasd woot this is a string"

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

    log.Printf(fmt.Sprintf("Recieved command: %s, input: %s", command, input))

    // handle our various commands
    switch(command) {
        case "good":
            sendMessageToSpark(&SparkMessage{RoomId: message.RoomId, Text:"good job!"})
            break
        case "generate": // generate a meme image using the input and send it to spark
            outputPath := "out.jpg"
            generateMeme(outputPath, input)
            sendImageToSpark(message.RoomId, outputPath)
            break
        default:
            break;
    }
}

func generateMeme(outputPath string, text string) (string, error) {
    path := "picard.jpg"
    fontSize := 36

    file, err := os.Open(path)
    if err != nil {
        return "", err
    }

    img, _, err := image.Decode(file)
    r := img.Bounds()
    w := r.Dx()
    h := r.Dy()

    m := gg.NewContext(w, h)
    m.DrawImage(img, 0, 0)
    m.LoadFontFace("/Library/Fonts/Impact.ttf", fontSize)

    // Apply black stroke
    m.SetHexColor("#000")
    strokeSize := 6
    for dy := -strokeSize; dy <= strokeSize; dy++ {
        for dx := -strokeSize; dx <= strokeSize; dx++ {
            // give it rounded corners
            if dx*dx+dy*dy >= strokeSize*strokeSize {
                continue
            }
            x := float64(w/2 + dx)
            y := float64(h - fontSize + dy)
            m.DrawStringAnchored(text, x, y, 0.5, 0.5)
        }
    }

    // Apply white fill
    m.SetHexColor("#FFF")
    m.DrawStringAnchored(text, float64(w)/2, float64(h)-fontSize, 0.5, 0.5)
    m.SavePNG(outputPath)

    return outputPath, nil
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
    defer r.Body.Close()

    // parse the webhook post data to get the message id.
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Printf(err.Error())
        return
    }

    var sparkWebhook SparkWebhook
    if err := json.Unmarshal(body, &sparkWebhook); err != nil {
        log.Printf(err.Error())
        return
    }

    // use the message id to get a message struct containing more message info
    message, err := getMessage(sparkWebhook.Data.Id)
    if err != nil {
        log.Printf(err.Error())
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
