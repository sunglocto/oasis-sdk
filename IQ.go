package oasis_sdk

import (
	"encoding/xml"
	"fmt"
	"github.com/google/uuid"
	"mellium.im/xmlstream"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/stanza"
)

// IQMap maps pending IQs to the channel awaiting their result. This way you can send
// off an IQ, then await the channel to send back the result
type IQMap map[string]chan xmlstream.TokenReadEncoder

type preparedIQ struct {
	stanza.IQ
	Payload any
}

// TODO: test
func (client *XmppClient) SendQuery(to jid.JID, iq any, resultObj *any) error {
	id := uuid.New().String()

	// Create and send the IQ stanza
	iqStanza := preparedIQ{
		IQ: stanza.IQ{
			To:   to,
			ID:   id,
			Type: stanza.GetIQ,
		},
		Payload: iq,
	}

	// something to get the result back
	resultChan := make(chan xmlstream.TokenReadEncoder)
	defer close(resultChan)

	// synchronized map of pending IQs
	client.IQMap[id] = resultChan
	defer delete(client.IQMap, id)

	// send request
	err := client.Session.Encode(client.Ctx, iqStanza)
	if err != nil {
		return fmt.Errorf("oasis_sdk: failed to encode IQ after appending header: %w", err)
	}

	//await result stanza
	t := <-resultChan

	// decode result stanza to result object
	d := xml.NewTokenDecoder(t)
	err = d.Decode(resultObj)
	if err != nil {
		return err
	}

	//no error if successful
	return nil
}
