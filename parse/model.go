package parse

import (
	"sort"
	"time"
)

type MessageID string

type Message struct {
	ID     MessageID
	Parent *MessageID
	User   string
	Flair  string
	Date   time.Time
	Title  *string
	Body   string
}

type MessageThread map[MessageID]Message

func (t MessageThread) SortedByDate() []Message {
	messages := make([]Message, 0, len(t))

	for _, message := range t {
		messages = append(messages, message)
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Date.Before(messages[j].Date)
	})

	return messages
}
