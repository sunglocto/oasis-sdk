package oasis_sdk

import (
	"encoding/xml"

	"mellium.im/xmlstream"
	"mellium.im/xmpp/stanza"
)

// PresenceShow represents the possible values for the XMPP presence show element
type PresenceShow int

const (
	// PresenceShowUnknown is any unknown presence state
	PresenceShowUnknown PresenceShow = iota
	// PresenceShowAvailable indicates the entity is available for communication (empty/absent show element)
	PresenceShowAvailable
	// PresenceShowChat indicates the entity is actively interested in chatting
	PresenceShowChat
	// PresenceShowAway indicates the entity is temporarily away
	PresenceShowAway
	// PresenceShowXA indicates the entity is away for an extended period (eXtended Away)
	PresenceShowXA
	// PresenceShowDND indicates the entity is busy and does not want to be disturbed (Do Not Disturb)
	PresenceShowDND
)

type UserPresence struct {
	Indicator PresenceShow
	Status    string
	Type	  stanza.PresenceType
	Body      PresenceBody
	Header    stanza.Presence
}

func (client *XmppClient) internalHandleVanityPresence(header stanza.Presence, t xmlstream.TokenReadEncoder) error {
	d := xml.NewTokenDecoder(t)
	body := PresenceBody{}
	err := d.Decode(&body)
	if err != nil {
		return err
	}
	presence := Presence{
		Presence:     header,
		PresenceBody: body,
	}
	p := UserPresence{
		Status: presence.Status,
		Type: presence.Type,
		Body: body,
		Header: stanza.Presence
	}
	switch presence.Show {
	case "chat":
		p.Indicator = PresenceShowChat
	case "away":
		p.Indicator = PresenceShowAway
	case "xa":
		p.Indicator = PresenceShowXA
	case "dnd":
		p.Indicator = PresenceShowDND
	case "":
		p.Indicator = PresenceShowAvailable
	default:
		p.Indicator = PresenceShowUnknown
	}

	//get handler with lock
	client.handlers.Lock.Lock()
	handler := client.handlers.PresenceHandler
	client.handlers.Lock.Unlock()

	if handler != nil {
		handler(client, presence.From, p)
	}

	return nil
}
