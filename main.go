package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strings"
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
                "picard <message>: generate a meme! ex. @Captain picard hey, i just made a meme!"
            sendMessageToSpark(&SparkMessage{RoomId: message.RoomId, Text:newMessage})
        case "good":
            sendMessageToSpark(&SparkMessage{RoomId: message.RoomId, Text:"good job!"})
            break
        case "bad":
            sendMessageToSpark(&SparkMessage{RoomId: message.RoomId, Text:"bad :("})
            break
        case "picard": // generate a meme image using the input and send it to spark
            outputPath := "resources/out.jpg"
            generateMeme("resources/picard.jpg", outputPath, input)
            sendImageToSpark(message.RoomId, outputPath)
            break
        default:
            break;
    }
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

    port := os.Getenv("PORT")
    if len(port) == 0 { // default the port to 8080 for local testing
        port = "8080"
    }

    http.HandleFunc("/", handleWebhook)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
