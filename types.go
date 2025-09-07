package oasis_sdk

import (
	"context"
	"encoding/xml"
	"sync"
	"sync/atomic"

	"mellium.im/xmpp"
	"mellium.im/xmpp/bookmarks"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/muc"
	"mellium.im/xmpp/mux"
	"mellium.im/xmpp/stanza"
)

// LoginInfo is a struct of the information required to log into the xmpp  client
type LoginInfo struct {
	Host        string   `json:"Host"`
	User        string   `json:"User"`
	Password    string   `json:"Password"`
	DisplayName string   `json:"DisplayName"`
	TLSoff      bool     `json:"NoTLS"`
	StartTLS    bool     `json:"StartTLS"`
	MucsToJoin  []string `json:"Mucs"`
}

type FallbackBody struct {
	Start int `xml:"start,attr"`
	End   int `xml:"end,attr"`
}

type Fallback struct {
	XMLName xml.Name     `xml:"urn:xmpp:fallback:0 fallback"`
	For     string       `xml:"for,attr"`
	Body    FallbackBody `xml:"body"`
}

type Reply struct {
	XMLName xml.Name `xml:"urn:xmpp:reply:0 reply"`
	ID      string   `xml:"id,attr"`
	To      string   `xml:"to,attr"`
}

// OriginID provided by XEP-0359: Unique and Stable Stanza IDs
type OriginID struct {
	XMLName xml.Name `xml:"urn:xmpp:sid:0 origin-id"`
	ID      string   `xml:"id,attr"`
}

// DeliveryReceiptRequest provided by XEP-0333: Displayed Markers
type DeliveryReceiptRequest struct {
	XMLName xml.Name `xml:"urn:xmpp:receipts request"`
}

type DeliveryReceipt struct {
	XMLName xml.Name `xml:"urn:xmpp:receipts received"`
	ID      string   `xml:"id,attr"`
}

type DeliveryReceiptBody struct {
	Received DeliveryReceipt `xml:"received"`
}

type DeliveryReceiptResponse struct {
	stanza.Message
	Received DeliveryReceipt `xml:"received"`
}

// ReadReceiptRequest provided by XEP-0184: Message Delivery Receipts
type ReadReceiptRequest struct {
	XMLName xml.Name `xml:"urn:xmpp:chat-markers:0 markable"`
}

type ReadReceipt struct {
	XMLName xml.Name `xml:"urn:xmpp:chat-markers:0 displayed"`
	ID      string   `xml:"id,attr"`
}

type ReadReceiptBody struct {
	Displayed ReadReceipt `xml:"displayed"`
}

type ReadReceiptResponse struct {
	stanza.Message
	Displayed ReadReceipt `xml:"displayed"`
}

// OutOfBandMedia implements https://xmpp.org/extensions/xep-0066.html#x-oob
type OutOfBandMedia struct {
	XMLName     xml.Name `xml:"jabber:x:oob x"`
	URL         string   `xml:"url"`
	Description *string  `xml:"desc"` //optional
}

type UnknownElement struct {
	XMLName xml.Name
	Content string     `xml:",innerxml"`
	Attrs   []xml.Attr `xml:",any,attr"`
}

type ChatMessageBody struct {
	Body               *string                 `xml:"body"`
	OriginID           *OriginID               `xml:"origin-id"`
	StanzaID           *stanza.ID              `xml:"stanza-id"`
	Reply              *Reply                  `xml:"reply"`
	Fallback           []Fallback              `xml:"fallback"`
	Request            *DeliveryReceiptRequest `xml:"request"`
	Markable           *ReadReceiptRequest     `xml:"markable"`
	GoneChatState      *GoneChatstate          `xml:"gone"`
	ActiveChatState    *ActiveChatstate        `xml:"active"`
	InactiveChatState  *InactiveChatstate      `xml:"inactive"`
	ComposingChatState *ComposingChatstate     `xml:"composing"`
	PausedChatState    *PausedChatstate        `xml:"paused"`
	OutOfBandMedia     *OutOfBandMedia         `xml:"jabber:x:oob x"`
	Unknown            []UnknownElement        `xml:",any"`
	FallbacksParsed    bool                    `xml:"-"`
	CleanedBody        *string                 `xml:"-"`
	ReplyFallbackText  *string                 `xml:"-"`
}

func (chatMsg *ChatMessageBody) RequestingDeliveryReceipt() bool {
	return chatMsg.Request != nil
}

func (chatMsg *ChatMessageBody) RequestingReadReceipt() bool {
	return chatMsg.Markable != nil
}

/*
XMPPChatMessage struct is a representation of the stanza such that it's contextual items
such as room, as well as abstract methods such as .reply()
*/
type XMPPChatMessage struct {
	stanza.Message
	ChatMessageBody
}

type HttpUploadComponent struct {
	Jid         jid.JID
	MaxFileSize int
}

type ChatMessageHandler func(client *XmppClient, message *XMPPChatMessage)
type GroupChatMessageHandler func(client *XmppClient, channel *muc.Channel, message *XMPPChatMessage)
type ChatstateHandler func(client *XmppClient, from jid.JID, state ChatState)
type DeliveryReceiptHandler func(client *XmppClient, from jid.JID, id string)
type ReadReceiptHandler func(client *XmppClient, from jid.JID, id string)
type BookmarkHandler func(client *XmppClient, bookmark bookmarks.Channel)

type handlerMap struct {
	Lock                   sync.Mutex
	DmHandler              ChatMessageHandler
	GroupMessageHandler    GroupChatMessageHandler
	ChatstateHandler       ChatstateHandler
	DeliveryReceiptHandler DeliveryReceiptHandler
	ReadReceiptHandler     ReadReceiptHandler
	BookmarkHandler        BookmarkHandler
}

// XmppClient is the end xmpp client object from which everything else works around
type XmppClient struct {
	isStartedLock       sync.Mutex
	Ctx                 context.Context
	CtxCancel           context.CancelFunc
	Login               *LoginInfo
	JID                 *jid.JID
	Server              *string
	Session             *xmpp.Session
	Multiplexer         *mux.ServeMux
	AutojoinLevel       atomic.Int32
	HttpUploadComponent *HttpUploadComponent
	MucClient           *muc.Client
	mucsToJoin          []jid.JID
	MucChannels         map[string]*muc.Channel
	handlers            handlerMap
	bookmarks           map[string]bookmarks.Channel
	bookmarkLock        sync.RWMutex
}

// AwaitStart locks and unlocks the isStarted lock to safely await the client being started before executing things.
func (client *XmppClient) AwaitStart() {
	client.isStartedLock.Lock()
	defer client.isStartedLock.Unlock()
}
