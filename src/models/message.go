package models

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	FormId   uint   `json:"FormId"`
	TargetId uint   `json:"TargetId"`
	Type     int    `json:"Type"`
	Media    int    `json:"Media"`
	Content  []byte `json:"Content"`
	Pic      string `json:"Pic"`
	Url      string `json:"Url"`
	Desc     string `json:"Desc"`
}

func (s *Message) TableName() string {
	return "message"
}

func MessageFromString(jsonStr string) (Message, error) {
	var msg Message
	err := json.Unmarshal([]byte(jsonStr), &msg)
	return msg, err
}

func (m *Message) String() string {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf("Message{error: %v}", err)
	}
	return string(data)
}
