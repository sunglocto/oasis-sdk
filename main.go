package oasis_sdk

import (
	"context"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	"mellium.im/sasl"
	"mellium.im/xmpp"
	"mellium.im/xmpp/dial"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/muc"
	"mellium.im/xmpp/mux"
	"mellium.im/xmpp/stanza"
)

type connectionErrHandler func(err error)

/*
Connect dials the server and starts receiving the events.
If blocking is true, this method will not exit until the xmpp connection is no longer being maintained.
If blocking is false, this method will exit as soon as a connection is created, and errors will be emitted
through the callback onErr
*/
func (self *XmppClient) Connect(blocking bool, onErr connectionErrHandler) error {
	d := dial.Dialer{}

	conn, err := d.DialServer(self.Ctx, "tcp", *self.JID, *self.Server)
	if err != nil {
		return errors.New("Could not connect stage 1 - " + err.Error())
	}

	self.Session, err = xmpp.NewSession(
		self.Ctx,
		self.JID.Domain(),
		*self.JID,
		conn,
		0,
		xmpp.NewNegotiator(func(*xmpp.Session, *xmpp.StreamConfig) xmpp.StreamConfig {
			return xmpp.StreamConfig{
				Lang: "en",
				Features: []xmpp.StreamFeature{
					xmpp.BindResource(),
					xmpp.StartTLS(&tls.Config{
						ServerName: *self.Server,
						MinVersion: tls.VersionTLS12,
					}),
					xmpp.SASL("", self.Login.Password, sasl.ScramSha1Plus, sasl.ScramSha1, sasl.Plain),
				},
				TeeIn:  nil,
				TeeOut: nil,
			}
		},
		))
	if err != nil {
		return errors.New("Could not connect stage 2 - " + err.Error())
	}

	if self.Session == nil {
		panic("session never got set")
	}

	go func() {
		n := len(self.mucsToJoin)
		for i, mucJID := range self.mucsToJoin {
			fmt.Printf("Joining muc %d/%d \"%s\" with nickname \"%s\"\n", i+1, n, mucJID.Bare().String(), mucJID.Resourcepart())
			ch, err := self.MucClient.Join(self.Ctx, mucJID, self.Session)
			if err != nil {
				println(err.Error())
				continue
			}
			self.mucChannels[mucJID.String()] = ch
			fmt.Printf("joined muc %d/%d\n", i+1, n)
		}
	}()

	if blocking {
		return self.startServing()
	} else {
		//serve in a thread
		go func() {
			err := self.startServing()

			//if error, try callback error handler, otherwise panic
			if err != nil {
				if onErr == nil {
					panic(err)
				} else {
					onErr(err)
				}
			}
		}()
	}

	return nil
}

// MarkAsDelivered sends delivery receipt as per https://xmpp.org/extensions/xep-0184.html
func (self *XmppClient) MarkAsDelivered(orignalMSG *XMPPChatMessage) {
	msg := DeliveryReceiptResponse{
		Message: stanza.Message{
			To:   orignalMSG.From.Bare(),
			Type: orignalMSG.Type,
		},
		Received: DeliveryReceipt{
			ID: orignalMSG.ID, // dont send in groupchats, no need to handle
		},
	}
	err := self.Session.Encode(self.Ctx, msg)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// MarkAsRead sends Read receipt as per https://xmpp.org/extensions/xep-0333.html
func (self *XmppClient) MarkAsRead(orignalMSG *XMPPChatMessage) error {

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
	return self.Session.Encode(self.Ctx, msg)
	//err := self.Session.Encode(self.Ctx, msg)
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
}

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

// CreateClient creates the client object using the login info object and returns it
func CreateClient(
	login *LoginInfo,
	dmHandler ChatMessageHandler,
	groupMessageHandler GroupChatMessageHandler,
	chatstateHandler ChatstateHandler,
	deliveryReceiptHandler DeliveryReceiptHandler,
	readReceiptHandler ReadReceiptHandler,
) (*XmppClient, error) {

	mucJIDs := make([]jid.JID, 0, len(login.MucsToJoin))
	for _, jidStr := range login.MucsToJoin {
		//join with default displayname
		j, err := jid.Parse(jidStr + "/" + login.DisplayName)
		if err != nil {
			fmt.Println("Error parsing MUC jid: " + err.Error())
			continue
		}
		mucJIDs = append(mucJIDs, j)
	}

	// create client object
	client := &XmppClient{
		Login:                  login,
		dmHandler:              dmHandler,
		groupMessageHandler:    groupMessageHandler,
		chatstateHandler:       chatstateHandler,
		deliveryReceiptHandler: deliveryReceiptHandler,
		readReceiptHandler:     readReceiptHandler,
		mucsToJoin:             mucJIDs,
		mucChannels:            make(map[string]*muc.Channel),
	}
	client.Ctx, client.CtxCancel = context.WithCancel(context.Background())

	client.MucClient = &muc.Client{}
	messageNS := xml.Name{
		Local: "body",
	}

	// ------ chatstates -------
	composingNS := xml.Name{
		Local: "composing",
	}
	activeNS := xml.Name{
		Local: "active",
	}
	pausedNS := xml.Name{
		Local: "paused",
	}
	inactiveNS := xml.Name{
		Local: "inactive",
	}
	goneNS := xml.Name{
		Local: "gone",
	}
	// ------ chatstates ------

	// ------ receipts --------
	deliveredNS := xml.Name{
		Local: "received",
	}
	displayedNS := xml.Name{
		Local: "displayed",
	}
	// ------ receipts --------

	client.Multiplexer = mux.New(
		"jabber:client",

		//provide object to hold muc state
		muc.HandleClient(client.MucClient),

		//handlers for chat messages
		mux.MessageFunc(stanza.ChatMessage, messageNS, client.internalHandleDM),
		mux.MessageFunc(stanza.GroupChatMessage, messageNS, client.internalHandleGroupMsg),

		// Chat state handlers for direct messages
		mux.MessageFunc(stanza.ChatMessage, activeNS, client.internalActiveChatstateReceiver),
		mux.MessageFunc(stanza.ChatMessage, composingNS, client.internalComposingChatstateReciever),
		mux.MessageFunc(stanza.ChatMessage, pausedNS, client.internalPausedChatstateReceiver),
		mux.MessageFunc(stanza.ChatMessage, inactiveNS, client.internalInactiveChatstateReceiver),
		mux.MessageFunc(stanza.ChatMessage, goneNS, client.internalGoneChatstateReceiver),

		// Receipt handlers for direct messages
		mux.MessageFunc(stanza.ChatMessage, deliveredNS, client.internalHandleDeliveryReceipt),
		mux.MessageFunc(stanza.ChatMessage, displayedNS, client.internalHandleReadReceipt),

		// Chat state handlers for group messages
		mux.MessageFunc(stanza.GroupChatMessage, activeNS, client.internalActiveChatstateReceiver),
		mux.MessageFunc(stanza.GroupChatMessage, composingNS, client.internalComposingChatstateReciever),
		mux.MessageFunc(stanza.GroupChatMessage, pausedNS, client.internalPausedChatstateReceiver),
		mux.MessageFunc(stanza.GroupChatMessage, inactiveNS, client.internalInactiveChatstateReceiver),
		mux.MessageFunc(stanza.GroupChatMessage, goneNS, client.internalGoneChatstateReceiver),

		// Receipt handlers for group messages
		mux.MessageFunc(stanza.GroupChatMessage, displayedNS, client.internalHandleReadReceipt),
	)

	//string to jid object
	j, err := jid.Parse(login.User)
	if err != nil {
		return nil,
			errors.New("Could not parse user JID from `" + login.User + " - " + err.Error())
	}
	server := j.Domainpart()
	client.JID = &j
	client.Server = &server

	return client, nil
}
