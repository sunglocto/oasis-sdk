package oasis_sdk

import (
	"encoding/xml"
	"fmt"
	"mellium.im/xmlstream"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/stanza"
	"strings"
)

// SendText sends a plain message with `body` (type string) to `to` JID
func (self *XmppClient) SendText(to jid.JID, body string) error {
	msg := XMPPChatMessage{
		Message: stanza.Message{
			To:   to,
			Type: stanza.ChatMessage,
		},
		ChatMessageBody: ChatMessageBody{
			Body: &body,
		},
	}
	err := self.Session.Encode(self.Ctx, msg)
	return err
}

// ReplyToEvent replies to a message event with body as per https://xmpp.org/extensions/xep-0461.html
func (self *XmppClient) ReplyToEvent(originalMsg *XMPPChatMessage, body string) error {
	//pull out JIDs as per https://xmpp.org/extensions/xep-0461.html#usecases
	replyTo := originalMsg.From
	to := replyTo.Bare()

	//name to include in fallback
	var readableReplyTo string
	if originalMsg.Type == stanza.ChatMessage {
		readableReplyTo = to.String()
	} else if originalMsg.Type == stanza.GroupChatMessage {
		readableReplyTo = replyTo.Resourcepart()
	}

	timeAgo := "TODO ago"

	originalBody := *originalMsg.CleanedBody
	quoteOriginalBody := readableReplyTo + " | " + timeAgo + "\n> " + strings.ReplaceAll(originalBody, "\n", "\n> ") + "\n"

	//ID to use in reply as per https://xmpp.org/extensions/xep-0461.html#business-id
	var replyToID string
	if originalMsg.Type == stanza.GroupChatMessage {
		//TODO check if room advertizes unique ids, if not cannot reply in groupchat even if id is present
		if originalMsg.StanzaID == nil || originalMsg.StanzaID.By.String() != to.String() {
			return self.SendText(to, quoteOriginalBody+body)
		}
		replyToID = originalMsg.StanzaID.ID
	} else if originalMsg.OriginID != nil {
		replyToID = originalMsg.OriginID.ID
	} else {
		replyToID = originalMsg.ID
	}

	// <reply> as per https://xmpp.org/extensions/xep-0461.html#usecases
	replyStanza := Reply{
		To: replyTo.String(),
		ID: replyToID,
	}

	// <fallback> as per https://xmpp.org/extensions/xep-0461.html#compat
	replyFallback := Fallback{
		For: "urn:xmpp:reply:0",
		Body: FallbackBody{
			Start: 0,
			End:   len(quoteOriginalBody) - 1,
		},
	}

	b := quoteOriginalBody + body
	msg := XMPPChatMessage{
		Message: stanza.Message{
			To:   to,
			Type: originalMsg.Type,
		},
		ChatMessageBody: ChatMessageBody{
			Body:     &b,
			Reply:    &replyStanza,
			Fallback: []Fallback{replyFallback},
		},
	}
	return self.Session.Encode(self.Ctx, msg)
}

func (self *XmppClient) internalHandleDM(header stanza.Message, t xmlstream.TokenReadEncoder) error {
	//nothing to do if theres no handler
	if self.dmHandler == nil {
		return nil
	}

	//decode remaining parts to decode
	d := xml.NewTokenDecoder(t)
	body := &ChatMessageBody{}
	err := d.Decode(body)
	if err != nil {
		return err
	}
	msg := &XMPPChatMessage{
		Message:         header,
		ChatMessageBody: *body,
	}

	//mark as received if requested, and not group chat as per https://xmpp.org/extensions/xep-0184.html#when-groupchat
	if msg.RequestingDeliveryReceipt() {
		go self.MarkAsDelivered(msg)
	}

	msg.ParseReply()

	//call handler and return to connection
	self.dmHandler(self, msg)
	return nil
}

func (self *XmppClient) internalHandleGroupMsg(header stanza.Message, t xmlstream.TokenReadEncoder) error {
	//nothing to do if theres no handler
	if self.groupMessageHandler == nil {
		return nil
	}

	//decode remaining parts to decode
	d := xml.NewTokenDecoder(t)
	body := &ChatMessageBody{}
	err := d.Decode(body)
	if err != nil {
		return err
	}
	msg := &XMPPChatMessage{
		Message:         header,
		ChatMessageBody: *body,
	}

	ch := self.mucChannels[msg.From.Bare().String()]

	fmt.Printf("groupchat %s, found channel: %t\n", msg.From.String(), ch == nil)

	//no delivery receipt as per https://xmpp.org/extensions/xep-0184.html#when-groupchat

	msg.ParseReply()

	//call handler and return to connection
	self.groupMessageHandler(self, ch, msg)
	return nil
}
