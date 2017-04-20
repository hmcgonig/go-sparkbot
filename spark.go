package main

import (
    "log"
    "net/http"
    "bytes"
    "fmt"
    "io/ioutil"
    "os"
    "mime/multipart"
    "net/textproto"
    "path/filepath"
    "io"
    "errors"
    "encoding/json"
)

var sparkBearerToken = os.Getenv("SPARK_BEARER_TOKEN")
var sparkBotEmail = os.Getenv("SPARK_BOT_EMAIL")

func getMessage(messageId string) (SparkMessage, error) {
    var sparkMessage SparkMessage

    responseBody, err := doRequest("GET", map[string]string { "Content-type": "application/json", "Authorization": fmt.Sprintf("Bearer %s", sparkBearerToken) }, fmt.Sprintf("https://api.ciscospark.com/v1/messages/%s", messageId), 200, nil)
    if err != nil {
        return sparkMessage, err
    }

    if err := json.Unmarshal(responseBody, &sparkMessage); err != nil {
        return sparkMessage, err
    }

    return sparkMessage, nil
}

func sendMessageToSpark(message *SparkMessage) (error) {
    _, err := doRequest("POST", map[string]string { "Content-type": "application/json", "Authorization": fmt.Sprintf("Bearer %s", sparkBearerToken) }, "https://api.ciscospark.com/v1/messages", 200, message)
    if err != nil {
        return err
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
        log.Println(string(body))
        return errors.New(fmt.Sprintf("Message request returned status code: %d; error body: %s", int(resp.StatusCode), string(body)))
    }

    return nil
}
