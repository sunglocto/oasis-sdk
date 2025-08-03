package oasis_sdk

import (
	"fmt"
	"golang.org/x/net/context"
	"log"
	"mellium.im/xmpp/disco"
	"mellium.im/xmpp/disco/items"
	jid2 "mellium.im/xmpp/jid"
)

/*
The error result returned by the function controls how WalkItem continues.
If the function returns the special value ErrSkipItem, WalkItem skips the current item.
Otherwise, if the function returns a non-nil error, WalkItem stops entirely and returns that error.
*/

// DiscoServerItem handles a server item being discovered. Implements WalkItemFunc
func (client *XmppClient) DiscoServerItem(level int, item items.Item, err error) error {
	fmt.Printf(
		"discovered server item at level %d, name: %s, jid %s, node %v, err %v\n",
		level, item.Name, item.JID.String(), item.Node, err,
	)

	info, err := disco.GetInfo(context.Background(), "", item.JID, client.Session)

	fmt.Printf("%s: %v", item.JID.String(), info)

	return nil
}

func (client *XmppClient) DiscoServicesOnSelf() {
	item := items.Item{
		JID:  *client.JID,
		Name: "self",
	}

	disco.WalkItem(context.Background(), item, client.Session, client.DiscoServerItem)
}

func (client *XmppClient) DiscoServicesOnServer() {
	jid, err := jid2.Parse(*client.Server)
	if err != nil {
		log.Fatalf("server string \"%s\" not a valid JID, %v", *client.Server, err)
	}

	item := items.Item{
		JID:  jid,
		Name: *client.Server,
	}

	disco.WalkItem(context.Background(), item, client.Session, client.DiscoServerItem)

}
