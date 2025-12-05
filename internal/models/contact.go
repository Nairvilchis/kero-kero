package models

// CheckContactRequest solicitud para verificar números
type CheckContactRequest struct {
	Phones []string `json:"phones" validate:"required"`
}

// ContactInfo información detallada de un contacto
type ContactInfo struct {
	JID          string `json:"jid"`
	Phone        string `json:"phone"`
	Found        bool   `json:"found"`
	FirstName    string `json:"first_name,omitempty"`
	FullName     string `json:"full_name,omitempty"`
	PushName     string `json:"push_name,omitempty"`
	BusinessName string `json:"business_name,omitempty"`
	Status       string `json:"status,omitempty"`      // About/Estado (ej: "Disponible")
	PictureURL   string `json:"picture_url,omitempty"` // URL de la foto de perfil
	IsBusiness   bool   `json:"is_business,omitempty"`
	IsBlocked    bool   `json:"is_blocked,omitempty"`
}

// PresenceRequest solicitud para suscribirse a presencia
type PresenceRequest struct {
	Phone string `json:"phone" validate:"required"`
}

// ProfilePictureInfo información de la foto de perfil
type ProfilePictureInfo struct {
	URL  string `json:"url"`
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"` // image o preview
}

// AboutResponse respuesta con el estado/about del contacto
type AboutResponse struct {
	About string `json:"about"`
}

// BlockRequest solicitud para bloquear/desbloquear contacto
type BlockRequest struct {
	Phone string `json:"phone" validate:"required"`
}

// CheckNumbersRequest solicitud para verificar si números están en WhatsApp
type CheckNumbersRequest struct {
	Phones []string `json:"phones" validate:"required,min=1,max=50,dive,required"`
}

// CheckNumberResult resultado de verificación de un número
type CheckNumberResult struct {
	Phone  string  `json:"phone"`
	Exists bool    `json:"exists"`
	JID    *string `json:"jid,omitempty"` // JID completo si existe (ej: 5215512345678@s.whatsapp.net)
}

// CheckNumbersResponse respuesta de verificación de números
type CheckNumbersResponse struct {
	Success bool                `json:"success"`
	Data    []CheckNumberResult `json:"data"`
}
