package oasis_sdk

import (
	"encoding/xml"
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
