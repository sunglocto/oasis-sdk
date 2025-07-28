package oasis_sdk

import (
	"encoding/xml"
	"mellium.im/xmlstream"
	"mellium.im/xmpp/stanza"
)

// startServing is an internal function to add an internal handler to the session.
// Most of this is just obtuse things inherited from mellium
func (client *XmppClient) startServing() error {
	err := client.Session.Send(client.Ctx, stanza.Presence{Type: stanza.AvailablePresence}.Wrap(nil))
	if err != nil {
		return err
	}
	return client.Session.Serve(
		client.Multiplexer,
	)
}

func (client *XmppClient) internalHandleDeliveryReceipt(header stanza.Message, t xmlstream.TokenReadEncoder) error {

	//only decode if there is a handler
	if client.deliveryReceiptHandler == nil {
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

	client.deliveryReceiptHandler(client, header.From, id)
	return nil
}

func (client *XmppClient) internalHandleReadReceipt(header stanza.Message, t xmlstream.TokenReadEncoder) error {

	//only decode if there is a handler
	if client.readReceiptHandler == nil {
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

	client.readReceiptHandler(client, header.From, id)
	return nil
}
