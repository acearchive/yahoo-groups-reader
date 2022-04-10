package parse

import (
	"time"
)

const (
	FirstMessage MessageNumber = 1
	NullParent   MessageNumber = -1
)

type MessageNumber int

type Message struct {
	Number MessageNumber
	Parent MessageNumber
	User   string
	Flair  string
	Date   time.Time
	Title  string
	Body   string
}

type MessageThread map[MessageNumber]Message

func (t MessageThread) Messages() []Message {
	messages := make([]Message, 0, len(t))

	for number := FirstMessage; int(number) <= len(t); number++ {
		messages = append(messages, t[number])
	}

	return messages
}
