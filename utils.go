package main

import (
    "net/http"
    "encoding/json"
    "bytes"
    "io/ioutil"
    "log"
    "fmt"
    "errors"
)

func doRequest(method string, headers map[string]string, url string, expectedResponseCode int, contentStructure interface{}) ([]byte, error){
    client := &http.Client{}

    // marshal the message
    requestBody, err := json.Marshal(contentStructure)
    if err != nil {
        return []byte{}, err
    }

    req, _ := http.NewRequest(method, url, bytes.NewReader(requestBody))

    for k, v := range headers {
        req.Header.Set(k, v)
    }

    // do the req
    resp, err := client.Do(req)
    if err != nil {
        return []byte{}, err
    }
    defer resp.Body.Close()

    // simple error handling
    if resp.StatusCode != expectedResponseCode {
        body, _ := ioutil.ReadAll(resp.Body)
        log.Println(string(body))
        return []byte{}, errors.New(fmt.Sprintf("Error: Request returned status code: %d; error body: %s", int(resp.StatusCode), string(body)))
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return []byte{}, err
    }

    return body, nil
}
