package oasis_sdk

import (
	"encoding/xml"
	"errors"
	"fmt"

	"mellium.im/xmlstream"
	"mellium.im/xmpp/stanza"
)

// ------ routing namespace --------
var deliveredNS = xml.Name{
	Local: "received",
}
var displayedNS = xml.Name{
	Local: "displayed",
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

// MarkAsDelivered sends delivery receipt as per https://xmpp.org/extensions/xep-0184.html
func (client *XmppClient) MarkAsDelivered(orignalMSG *XMPPChatMessage) {
	msg := DeliveryReceiptResponse{
		Message: stanza.Message{
			To:   orignalMSG.From.Bare(),
			Type: orignalMSG.Type,
		},
		Received: DeliveryReceipt{
			ID: orignalMSG.ID, // dont send in groupchats, no need to handle
		},
	}
	err := client.Session.Encode(client.Ctx, msg)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// MarkAsRead sends Read receipt as per https://xmpp.org/extensions/xep-0333.html
func (client *XmppClient) MarkAsRead(orignalMSG *XMPPChatMessage) error {

	//pull relevant id for type of message
	var id string
	if orignalMSG.Type == stanza.GroupChatMessage {
		stanzaID := orignalMSG.StanzaID
		if stanzaID == nil {
			return errors.New("stanza id is nil")
		}
		if stanzaID.By.String() != orignalMSG.From.Bare().String() {
			return errors.New("stanza id is not set by group host")
		}
		//TODO check if muc advertises stable IDs
		id = stanzaID.ID
	} else {
		id = orignalMSG.ID
	}

	//craft event
	msg := ReadReceiptResponse{
		Message: stanza.Message{
			To:   orignalMSG.From.Bare(),
			Type: orignalMSG.Type,
		},
		Displayed: ReadReceipt{
			ID: id,
		},
	}

	//send
	return client.Session.Encode(client.Ctx, msg)
}
