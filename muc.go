package oasis_sdk

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mellium.im/xmpp/bookmarks"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/muc"
)

type MucLegacyHistoryConfig struct {
	Duration *time.Duration
	MaxCount *uint64
	Since    *time.Time
}

func (client *XmppClient) ConnectMuc(bookmark bookmarks.Channel, histCFG MucLegacyHistoryConfig, ctx context.Context) (*muc.Channel, error) {
	fmt.Println("Debug: Connecting to muc", bookmark)

	if bookmark.Nick == "" {
		return nil, errors.New("no nick provided")
	}
	j, err := jid.Parse(fmt.Sprintf("%s/%s", bookmark.JID.String(), bookmark.Nick))
	if err != nil {
		return nil, fmt.Errorf("error parsing jid: %v", err)
	}

	// https://pkg.go.dev/mellium.im/xmpp/muc@v0.22.0#Option
	opts := make([]muc.Option, 0, 5)

	// get nick and pass from the bookmark object
	opts = append(opts, muc.Nick(bookmark.Nick))
	if bookmark.Password != "" {
		opts = append(opts, muc.Password(bookmark.Password))
	}

	//get hist config from histCFG
	if histCFG.Duration != nil {
		opts = append(opts, muc.Duration(*histCFG.Duration))
	}
	if histCFG.MaxCount != nil {
		opts = append(opts, muc.MaxHistory(*histCFG.MaxCount))
	}
	if histCFG.Since != nil {
		opts = append(opts, muc.Since(*histCFG.Since))
	}

	ch, err := client.MucClient.Join(ctx, j, client.Session, opts...)
	if err != nil {
		return nil, fmt.Errorf("mellium unable to join muc %s: %w",
			bookmark.JID.String(), err)
	}

	client.mucLock.Lock()
	defer client.mucLock.Unlock()
	client.MucChannels[bookmark.JID.String()] = ch

	return ch, nil
}

func (client *XmppClient) DisconnectMuc(mucStr string, reason string, ctx context.Context) error {
	client.mucLock.Lock()
	defer client.mucLock.Unlock()

	ch, ok := client.MucChannels[mucStr]
	if !ok {
		return fmt.Errorf("muc channel '%s' not found", mucStr)
	}

	err := ch.Leave(ctx, reason)
	if err != nil {
		return fmt.Errorf("mellium unable to leave muc %s: %w", mucStr, err)
	}

	return nil
}
