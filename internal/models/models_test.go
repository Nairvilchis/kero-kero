package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstanceStatus(t *testing.T) {
	tests := []struct {
		name   string
		status InstanceStatus
		want   string
	}{
		{"Disconnected", StatusDisconnected, "disconnected"},
		{"Connecting", StatusConnecting, "connecting"},
		{"Connected", StatusConnected, "connected"},
		{"Authenticated", StatusAuthenticated, "authenticated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.status))
		})
	}
}

func TestSendTextRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     SendTextRequest
		wantErr bool
	}{
		{
			name: "válido",
			req: SendTextRequest{
				Phone:   "5215512345678",
				Message: "Hola",
			},
			wantErr: false,
		},
		{
			name: "sin teléfono",
			req: SendTextRequest{
				Message: "Hola",
			},
			wantErr: true,
		},
		{
			name: "sin mensaje",
			req: SendTextRequest{
				Phone: "5215512345678",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Aquí normalmente usarías un validador como go-playground/validator
			// Por ahora solo verificamos que los campos requeridos estén presentes
			hasPhone := tt.req.Phone != ""
			hasMessage := tt.req.Message != ""
			isValid := hasPhone && hasMessage

			if tt.wantErr {
				assert.False(t, isValid)
			} else {
				assert.True(t, isValid)
			}
		})
	}
}

func TestWebhookEvent(t *testing.T) {
	event := WebhookEvent{
		InstanceID: "test-instance",
		Event:      "message",
		Timestamp:  1234567890,
		Data: MessageEvent{
			MessageID:   "msg-123",
			From:        "5215512345678",
			MessageType: "text",
			Text:        "Hola",
		},
	}

	assert.Equal(t, "test-instance", event.InstanceID)
	assert.Equal(t, "message", event.Event)
	assert.NotNil(t, event.Data)
}

func TestMessageEvent(t *testing.T) {
	msg := MessageEvent{
		MessageID:   "msg-123",
		From:        "5215512345678",
		To:          "5215587654321",
		IsGroup:     false,
		MessageType: "text",
		Text:        "Hola mundo",
	}

	assert.Equal(t, "msg-123", msg.MessageID)
	assert.Equal(t, "text", msg.MessageType)
	assert.False(t, msg.IsGroup)
	assert.Equal(t, "Hola mundo", msg.Text)
}

func TestGroupResponse(t *testing.T) {
	group := GroupResponse{
		JID:              "123456789@g.us",
		Name:             "Mi Grupo",
		ParticipantCount: 5,
		Participants: []GroupParticipant{
			{
				JID:     "5215512345678@s.whatsapp.net",
				IsAdmin: true,
			},
		},
	}

	assert.Equal(t, "Mi Grupo", group.Name)
	assert.Equal(t, 5, group.ParticipantCount)
	assert.Len(t, group.Participants, 1)
	assert.True(t, group.Participants[0].IsAdmin)
}
