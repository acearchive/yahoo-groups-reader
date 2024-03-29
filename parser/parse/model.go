package parse

import (
	"github.com/acearchive/yahoo-groups-reader/body"
	"sort"
	"time"
)

type MessageID string

type MessageBody struct {
	Tokens []body.Token
	Html   string
}

type Message struct {
	ID     MessageID
	Parent *MessageID
	User   string
	Flair  string
	Date   time.Time
	Title  *string
	Body   MessageBody
}

type MessageThread map[MessageID]Message

func (t MessageThread) SortedByDate() ([]Message, map[MessageID]int) {
	messages := make([]Message, 0, len(t))
	messageIndices := make(map[MessageID]int, len(t))

	for _, message := range t {
		messages = append(messages, message)
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Date.Before(messages[j].Date)
	})

	for i, message := range messages {
		messageIndices[message.ID] = i
	}

	return messages, messageIndices
}
