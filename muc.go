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

// MucLegacyHistoryConfig is the configuration to fetch legacy muc history on join
// for most uses this is obsoleted by mam, which can be fetched separately
type MucLegacyHistoryConfig struct {
	Duration *time.Duration
	MaxCount *uint64
	Since    *time.Time
}

// ConnectMuc connects to a Multi-User Chat (MUC) using the provided bookmark and Legacy history configuration.
// It returns the joined MUC channel or an error if the connection fails.
// The function validates the provided bookmark, applies history settings, and manages the client's active MUC channels.
func (client *XmppClient) ConnectMuc(bookmark bookmarks.Channel, histCFG MucLegacyHistoryConfig, ctx context.Context) (*muc.Channel, error) {

	client.AwaitStart()

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

// DisconnectMuc disconnects the client from a specified MUC (Multi-User Chat) using the provided reason and context.
// It locks the mucLock mutex, retrieves the associated MUC channel, and leaves the MUC if found.
// Returns an error if the MUC channel is not found or if a failure occurs while leaving the MUC.
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

// LeaveMuc allows the client to leave a MUC (Multi-User Chat) room specified by `mucStr` and provides a reason for leaving.
// It updates the internal state by disabling autojoin for the room and attempts to leave the room gracefully.
// Returns two errors: the first for updating the autojoin setting, and the second for the action of leaving the MUC.
func (client *XmppClient) LeaveMuc(mucStr string, reason string, ctx context.Context) (error, error) {

	//just hold the error and still try the second part
	err1 := client.ToggleAutojoin(mucStr, false, context.WithoutCancel(ctx))

	client.mucLock.Lock()
	defer client.mucLock.Unlock()

	muc, ok := client.MucChannels[mucStr]
	if !ok {
		//we have both error values
		err2 := fmt.Errorf("muc channel '%s' not found", mucStr)
		return err1, err2
	}

	//try second part of leave and return any possible errors
	err2 := muc.Leave(context.WithoutCancel(ctx), reason)
	return err1, err2
}

// JoinMuc allows the client to join a multi-user chat (MUC) room, optionally publishing a bookmark and fetching history.
// It takes a channel bookmark, legacy history configuration, and context as parameters.
// Returns the joined MUC channel, a potential error from publishing the bookmark, and a potential error from joining the MUC.
func (client *XmppClient) JoinMuc(bookmark bookmarks.Channel, histCFG MucLegacyHistoryConfig, ctx context.Context) (*muc.Channel, error, error) {
	err1 := client.PublishBookmark(bookmark, context.WithoutCancel(ctx))

	muc, err2 := client.ConnectMuc(bookmark, histCFG, context.WithoutCancel(ctx))

	return muc, err1, err2
}
