package demo

import (
	"errors"
	"fmt"
	"log"
	"time"
)

type Messenger struct {
	user         Contact
	contacts     []Contact
	messages     []Message
	overviewSubs map[chan []OverviewItem]interface{}
	chatSubs     map[chan ChatScreen]ContactId
}

// Data =====================================

type Message struct {
	Id   MessageId
	From ContactId
	To   ContactId
	Text string
	Time int64
}
type MessageId struct {
	Value string
}

type Contact struct {
	Id    ContactId
	Name  string
	Group []ContactId
}

type ContactId struct{ Value string }

type Contacts struct{ Contacts []Contact }

type OverviewItem struct {
	Contact     Contact
	LastMessage Message
}

type ChatScreen struct {
	Contact         Contact
	Messages        []Message
	Messages_add    []Message
	Messages_update []Message
	Messages_remove []MessageId
}

// Constructors =====================================

func NewMessenger() *Messenger {
	return &Messenger{
		user: Contact{ContactId{"0"}, "Me", nil},
		contacts: []Contact{
			{ContactId{"1"}, "Joe", nil},
			{ContactId{"2"}, "Bob", nil},
		},
		messages: []Message{
			{nextMessageId(), ContactId{"1"}, ContactId{"0"}, "yolo", time.Now().UnixMilli()},
			{nextMessageId(), ContactId{"0"}, ContactId{"1"}, "yolo rolf", time.Now().UnixMilli()},
			{nextMessageId(), ContactId{"2"}, ContactId{"0"}, "rolling jerks", time.Now().UnixMilli()},
			{nextMessageId(), ContactId{"0"}, ContactId{"2"}, "rly?", time.Now().UnixMilli()},
		},
		overviewSubs: map[chan []OverviewItem]interface{}{},
		chatSubs:     map[chan ChatScreen]ContactId{},
	}
}

// Methods =====================================

func (m *Messenger) GetOverview(
	reply chan []OverviewItem,
) (error, func()) {
	reply <- m.getOverview()
	log.Println("add overview subscription")
	m.overviewSubs[reply] = reply
	return nil, func() {
		log.Println("remove overview subscription")
		delete(m.overviewSubs, reply)
	}
}

func (m *Messenger) GetContacts(
	reply chan []Contact,
) error {
	reply <- []Contact{
		{Id: ContactId{"1"}, Name: "Joe"},
		{Id: ContactId{"2"}, Name: "Bob"},
	}
	return nil
}
func (m *Messenger) GetMessages(
	id ContactId,
	reply chan ChatScreen,
) (error, func()) {
	var chatContact *Contact
	for _, contact := range m.contacts {
		if contact.Id == id {
			chatContact = &contact
		}
	}
	if chatContact == nil {
		return errors.New(fmt.Sprint("No contact with id", id.Value)), nil
	}
	var chatMessages []Message
	for _, message := range m.messages {
		if message.From == id || message.To == id {
			chatMessages = append(chatMessages, message)
		}
	}
	reply <- ChatScreen{
		Contact:  *chatContact,
		Messages: chatMessages,
	}

	log.Println("add chat subscription")
	m.chatSubs[reply] = id
	return nil, func() {
		log.Println("remove chat subscription")
		delete(m.chatSubs, reply)
	}
}

func (m *Messenger) SendMessage(
	id ContactId,
	text string,
) error {
	// find Contact by ContactId
	var contact *Contact
	for _, c := range m.contacts {
		if c.Id == id {
			contact = &c
		}
	}
	if contact == nil {
		return errors.New(fmt.Sprint("No contact with id", id))
	}

	// append message to cache
	message := Message{
		Id:   nextMessageId(),
		From: m.user.Id,
		To:   id,
		Text: text,
		Time: time.Now().UnixMilli(),
	}
	m.messages = append(m.messages, message)

	// propagate change to chat subscribers
	screen := ChatScreen{
		Contact:      *contact,
		Messages_add: []Message{message},
	}
	for sub, contactId := range m.chatSubs {
		if contactId == id {
			sub <- screen
		}
	}

	// propagate change to overview subscribers
	overview := m.getOverview()
	for sub, _ := range m.overviewSubs {
		sub <- overview
	}

	return nil
}

// Internal  =====================================

func (m *Messenger) getOverview() (items []OverviewItem) {
	for _, contact := range m.contacts {
		for i := len(m.messages) - 1; i >= 0; i-- {
			message := m.messages[i]
			if contact.Id.participate(message) {
				items = append(items, OverviewItem{
					Contact:     contact,
					LastMessage: message,
				})
				break
			}
		}
	}
	return
}

func (id ContactId) participate(message Message) bool {
	return message.From == id || message.To == id
}

func nextMessageId() MessageId {
	message := MessageId{fmt.Sprint(messageCount)}
	messageCount++
	return message
}

var messageCount = 0
