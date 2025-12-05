package whatsapp

import (
	"context"
	"fmt"
	"sync"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type Client struct {
	WAClient   *whatsmeow.Client
	EvtHandler func(interface{})
	Log        waLog.Logger

	qrMu        sync.RWMutex
	LastQR      string
	SyncHistory bool
}

func NewClient(device *store.Device, log waLog.Logger) *Client {
	if log == nil {
		log = waLog.Stdout("Client", "DEBUG", true)
	}

	client := whatsmeow.NewClient(device, log)

	c := &Client{
		WAClient: client,
		Log:      log,
	}

	client.AddEventHandler(c.eventHandler)
	return c
}

func (c *Client) Connect() error {
	err := c.WAClient.Connect()
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Disconnect() {
	c.WAClient.Disconnect()
}

func (c *Client) RegisterHandler(handler func(interface{})) {
	c.EvtHandler = handler
}

func (c *Client) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.QR:
		c.qrMu.Lock()
		c.LastQR = v.Codes[0]
		c.qrMu.Unlock()
		fmt.Println("QR Code received:", v.Codes)
	}

	if c.EvtHandler != nil {
		c.EvtHandler(evt)
	}
}

func (c *Client) GetLastQR() string {
	c.qrMu.RLock()
	defer c.qrMu.RUnlock()
	return c.LastQR
}

// ResolveJID intenta resolver un LID a un JID de teléfono si es posible
func (c *Client) ResolveJID(jid types.JID) types.JID {
	if jid.Server != types.HiddenUserServer {
		return jid
	}

	// Usar el store de LIDs de whatsmeow para obtener el número de teléfono asociado
	if c.WAClient.Store.LIDs != nil {
		if pn, err := c.WAClient.Store.LIDs.GetPNForLID(context.Background(), jid); err == nil && !pn.IsEmpty() {
			return pn
		}
	}

	return jid
}
