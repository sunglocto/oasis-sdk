package oasis_sdk

import (
	"encoding/xml"
	"mellium.im/xmlstream"
	"mellium.im/xmpp/stanza"
)

type ChatState int

const (
	ChatStateActive ChatState = iota
	ChatStateInactive
	ChatStateComposing
	ChatStatePaused
	ChatStateGone
)

type GoneChatstate struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates gone"`
}
type ActiveChatstate struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates active"`
}
type InactiveChatstate struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates inactive"`
}
type ComposingChatstate struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates composing"`
}
type PausedChatstate struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates paused"`
}

func (client *XmppClient) internalComposingChatstateReciever(header stanza.Message, _ xmlstream.TokenReadEncoder) error {
	if client.chatstateHandler != nil {
		client.chatstateHandler(client, header.From, ChatStateComposing)
	}
	return nil
}

func (client *XmppClient) internalActiveChatstateReceiver(header stanza.Message, _ xmlstream.TokenReadEncoder) error {
	if client.chatstateHandler != nil {
		client.chatstateHandler(client, header.From, ChatStateActive)
	}
	return nil
}

func (client *XmppClient) internalPausedChatstateReceiver(header stanza.Message, _ xmlstream.TokenReadEncoder) error {
	if client.chatstateHandler != nil {
		client.chatstateHandler(client, header.From, ChatStatePaused)
	}
	return nil
}

func (client *XmppClient) internalInactiveChatstateReceiver(header stanza.Message, _ xmlstream.TokenReadEncoder) error {
	if client.chatstateHandler != nil {
		client.chatstateHandler(client, header.From, ChatStateInactive)
	}
	return nil
}

func (client *XmppClient) internalGoneChatstateReceiver(header stanza.Message, _ xmlstream.TokenReadEncoder) error {
	if client.chatstateHandler != nil {
		client.chatstateHandler(client, header.From, ChatStateGone)
	}
	return nil
}
