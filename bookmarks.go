package oasis_sdk

import (
	"fmt"

	"mellium.im/xmpp/bookmarks"
)

// BookmarkCache returns a thread-safe copy of the client's bookmarks cache as a map of JID strings to Channel objects.
func (client *XmppClient) BookmarkCache() map[string]bookmarks.Channel {
	client.bookmarkLock.RLock()
	defer client.bookmarkLock.RUnlock()
	res := make(map[string]bookmarks.Channel)
	for jidStr, channel := range client.bookmarks {
		res[jidStr] = channel
	}
	return res
}

func (client *XmppClient) FetchBookmarks() {

	//fetch
	iter := bookmarks.Fetch(client.Ctx, client.Session)

	//clear cache and prepare to write
	client.bookmarkLock.Lock()
	client.bookmarks = make(map[string]bookmarks.Channel)

	//scan
	for iter.Next() {
		//get this bookmark
		bookmark := iter.Bookmark()
		fmt.Println("bookmark", bookmark)
		client.bookmarks[bookmark.JID.String()] = bookmark
	}
	err := iter.Close()
	fmt.Println("bookmark closing", err)

	//done writing
	client.bookmarkLock.Unlock()

	//get bookmark handler
	client.handlers.Lock.Lock()
	handler := client.handlers.BookmarkHandler
	client.handlers.Lock.Unlock()
	if handler == nil {
		return
	}

	//switch to reading lock
	client.bookmarkLock.RLock()
	defer client.bookmarkLock.RUnlock()

	//emit every bookmark
	for _, channel := range client.bookmarks {
		handler(client, channel)
	}
}
