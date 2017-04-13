package main

type SparkWebhook struct {
    Data struct {
             Id string `json:"id"`
             PersonEmail string `json:"personEmail"`
         }
}

type SparkMessage struct {
    Id string `json:"id,omitempty"`
    RoomId string `json:"roomId"`
    RoomType string `json:"roomType,omitempty"`
    Text string `json:"text,omitempty"`
    Markdown string `json:"markdown,omitempty"`
    PersonId string `json:"personId,omitempty"`
    PersonEmail string `json:"personEmail,omitempty"`
    Html string `json:"html,omitempty"`
    MentionedPeople []string `json:"mentionedPeople,omitempty"`
    Created string `json:"created,omitempty"`
}