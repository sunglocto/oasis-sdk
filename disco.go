package oasis_sdk

import (
	"fmt"
	"golang.org/x/net/context"
	"log"
	"mellium.im/xmpp/disco"
	"mellium.im/xmpp/disco/items"
	jid2 "mellium.im/xmpp/jid"
	"strconv"
)

/*
The error result returned by the function controls how WalkItem continues.
If the function returns the special value ErrSkipItem, WalkItem skips the current item.
Otherwise, if the function returns a non-nil error, WalkItem stops entirely and returns that error.
*/

// DiscoServerItem handles a server item being discovered. Implements WalkItemFunc
func (client *XmppClient) DiscoServerItem(level int, item items.Item, err error) error {
	//fmt.Printf(
	//	"discovered server item at level %d, name: %s, jid %s, node %v, err %v\n",
	//	level, item.Name, item.JID.String(), item.Node, err,
	//)

	info, err := disco.GetInfo(context.Background(), "", item.JID, client.Session)
	if err != nil {
		fmt.Printf("Error while getting info about %s, %v\n", item.JID.String(), err)
	}

	identity := info.Identity[0]
	//fmt.Printf("%s: Type %s, Category %s\n", item.JID.String(), identity.Type, identity.Category)

	if identity.Type == "text" && identity.Category == "conference" {
		return disco.ErrSkipItem
	}

	if identity.Type == "file" && identity.Category == "store" {
		httpUploadComponent := HttpUploadComponent{
			Jid: item.JID,
		}

		for _, x := range info.Form {
			v, ok := x.GetString("max-file-size")
			if ok {
				//fmt.Printf("max-file-size: %s\n", v)
				maxFileSize, err := strconv.ParseInt(v, 10, 32)
				if err != nil {
					fmt.Printf("Could not parse max-file-size: %v\n", err)
					maxFileSize = 0
				}
				httpUploadComponent.MaxFileSize = int(maxFileSize)
				break
				//httpUploadComponent.MaxFileSize
			}
		}

		client.HttpUploadComponent = &httpUploadComponent
		fmt.Println(client.HttpUploadComponent)
	}

	return nil
}

func (client *XmppClient) DiscoServicesOnSelf() {
	item := items.Item{
		JID:  *client.JID,
		Name: "self",
	}

	err := disco.WalkItem(context.Background(), item, client.Session, client.DiscoServerItem)
	if err != nil {
		fmt.Printf("Error while walking self items: %v\n", err)
	}
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

	err = disco.WalkItem(context.Background(), item, client.Session, client.DiscoServerItem)
	if err != nil {
		fmt.Printf("Error while walking server items: %v\n", err)
	}

}
