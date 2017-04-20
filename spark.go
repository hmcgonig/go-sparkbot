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
        return sparkMessage, errors.New(fmt.Sprintf("Message request returned status code: %d", int(resp.StatusCode)))
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

    log.Println(string(body))

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
        log.Println(string(body))
        return errors.New(fmt.Sprintf("Message request returned status code: %d; error body: %s", int(resp.StatusCode), string(body)))
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
