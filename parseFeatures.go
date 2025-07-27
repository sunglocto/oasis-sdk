package oasis_sdk

func (self *ChatMessageBody) ParseReply() {
	//if this isnt a reply we dont need to parse the reply out
	if self.Reply == nil || self.Reply.ID == "" || self.Reply.To == "" {
		self.CleanedBody = self.Body
		self.FallbacksParsed = true
		return
	}

	var replyFallback *Fallback
	for _, fallback := range self.Fallback {
		if fallback.For == "urn:xmpp:reply:0" {
			replyFallback = &fallback
			break
		}
	}
	if replyFallback == nil {
		self.CleanedBody = self.Body
		self.FallbacksParsed = true
		return
	}

	//parse out reply fallback
	bodyPart := ""
	if replyFallback.Body.Start > 0 {
		bodyPart += (*self.Body)[:replyFallback.Body.Start]
	}
	replyPart := (*self.Body)[replyFallback.Body.Start:replyFallback.Body.End]
	bodyPart += (*self.Body)[replyFallback.Body.End:]
	self.CleanedBody = &bodyPart
	self.ReplyFallbackText = &replyPart
}
