package oasis_sdk

import (
	"encoding/xml"
	"fmt"

	"mellium.im/xmlstream"
	"mellium.im/xmpp/stanza"
)

// startServing is an internal function to add an internal handler to the session.
// Most of this is just obtuse things inherited from mellium
func (self *XmppClient) startServing() error {
	err := self.Session.Send(self.Ctx, stanza.Presence{Type: stanza.AvailablePresence}.Wrap(nil))
	if err != nil {
		return err
	}
	return self.Session.Serve(
		self.Multiplexer,
	)
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

func (self *XmppClient) internalHandleDeliveryReceipt(header stanza.Message, t xmlstream.TokenReadEncoder) error {

	//only decode if there is a handler
	if self.deliveryReceiptHandler == nil {
		return nil
	}

	// decode receipt type message
	d := xml.NewTokenDecoder(t)
	receipt := DeliveryReceiptBody{}
	err := d.Decode(&receipt)
	if err != nil {
		return err
	}

	//only one possible field
	id := receipt.Received.ID

	self.deliveryReceiptHandler(self, header.From, id)
	return nil
}

func (self *XmppClient) internalHandleReadReceipt(header stanza.Message, t xmlstream.TokenReadEncoder) error {

	//only decode if there is a handler
	if self.readReceiptHandler == nil {
		return nil
	}

	// decode receipt type message
	d := xml.NewTokenDecoder(t)
	receipt := ReadReceiptBody{}
	err := d.Decode(&receipt)
	if err != nil {
		return err
	}

	//only one possible field
	id := receipt.Displayed.ID

	self.readReceiptHandler(self, header.From, id)
	return nil
}
