package whatsapp

import (
	"strings"

	"go.mau.fi/whatsmeow/types"
)

// GetChatType determina el tipo de chat basándose en el JID
func GetChatType(jid string) string {
	parsedJID, err := types.ParseJID(jid)
	if err != nil {
		return "unknown"
	}

	switch {
	case parsedJID.Server == types.DefaultUserServer:
		return "private"
	case parsedJID.Server == types.GroupServer:
		return "group"
	case parsedJID.Server == types.NewsletterServer:
		return "channel"
	case parsedJID.Server == types.BroadcastServer || strings.Contains(jid, "status@broadcast"):
		return "status"
	default:
		return "unknown"
	}
}
