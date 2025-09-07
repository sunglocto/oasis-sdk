package oasis_sdk

import (
	"fmt"

	"mellium.im/xmpp/bookmarks"
)

func (client *XmppClient) SetBookmarkHandler(reEmit bool, handler BookmarkHandler) {
	//set handler
	client.handlers.Lock.Lock()
	client.handlers.BookmarkHandler = handler
	client.handlers.Lock.Unlock()

	client.bookmarkLock.RLock()
	//emit to the handler
	if reEmit || len(client.bookmarks) == 0 {
		go client.fetchBookmarks(true)
	}
	client.bookmarkLock.RUnlock()
}

// RefreshBookmarks updates the internal bookmark cache by fetching the latest bookmarks from the server, then returns that cache
func (client *XmppClient) RefreshBookmarks(reEmit bool) map[string]bookmarks.Channel {
	client.fetchBookmarks(reEmit)
	return client.BookmarkCache()
}

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

// fetchBookmarks synchronizes the client's bookmarks with the server and updates the local cache efficiently.
// It acquires necessary locks for safe concurrent access and emits bookmarks to a registered handler if available.
func (client *XmppClient) fetchBookmarks(emit bool) {

	client.AwaitStart()

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

	//only emit to handler if we should
	if !emit {
		return
	}

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
