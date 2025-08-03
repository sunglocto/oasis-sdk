package oasis_sdk

func (chatMsg *ChatMessageBody) ParseReply() {
	//if this isnt a reply we dont need to parse the reply out
	if chatMsg.Reply == nil || chatMsg.Reply.ID == "" || chatMsg.Reply.To == "" {
		chatMsg.CleanedBody = chatMsg.Body
		chatMsg.FallbacksParsed = true
		return
	}

	var replyFallback *Fallback
	for _, fallback := range chatMsg.Fallback {
		if fallback.For == "urn:xmpp:reply:0" {
			replyFallback = &fallback
			break
		}
	}
	if replyFallback == nil {
		chatMsg.CleanedBody = chatMsg.Body
		chatMsg.FallbacksParsed = true
		return
	}

	//parse out reply fallback
	bodyPart := ""
	if replyFallback.Body.Start > 0 {
		bodyPart += (*chatMsg.Body)[:replyFallback.Body.Start]
	}
	replyPart := (*chatMsg.Body)[replyFallback.Body.Start:replyFallback.Body.End]
	bodyPart += (*chatMsg.Body)[replyFallback.Body.End:]
	chatMsg.CleanedBody = &bodyPart
	chatMsg.ReplyFallbackText = &replyPart
}
