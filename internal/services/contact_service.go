package services

import (
	"context"
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"kero-kero/internal/models"
	"kero-kero/internal/whatsapp"
	"kero-kero/pkg/errors"
)

type ContactService struct {
	waManager *whatsapp.Manager
}

func NewContactService(waManager *whatsapp.Manager) *ContactService {
	return &ContactService{waManager: waManager}
}

// CheckContacts verifica si los números están registrados en WhatsApp
func (s *ContactService) CheckContacts(ctx context.Context, instanceID string, phones []string) ([]models.ContactInfo, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	result := make([]models.ContactInfo, 0, len(phones))

	for _, phone := range phones {
		phone = strings.TrimPrefix(phone, "+")

		jids, err := client.WAClient.IsOnWhatsApp(ctx, []string{phone})
		if err != nil {
			continue
		}

		if len(jids) > 0 {
			info := jids[0]
			contact := models.ContactInfo{
				JID:   info.JID.String(),
				Found: info.VerifiedName != nil || info.JID.User == phone,
			}

			if storeContact, err := client.WAClient.Store.Contacts.GetContact(ctx, info.JID); err == nil && storeContact.Found {
				contact.FirstName = storeContact.FirstName
				contact.FullName = storeContact.FullName
				contact.PushName = storeContact.PushName
				contact.BusinessName = storeContact.BusinessName
			}

			result = append(result, contact)
		} else {
			result = append(result, models.ContactInfo{
				JID:   phone,
				Found: false,
			})
		}
	}

	return result, nil
}

// GetProfilePicture obtiene la URL de la foto de perfil
func (s *ContactService) GetProfilePicture(ctx context.Context, instanceID, phone string) (*models.ProfilePictureInfo, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	jid := types.NewJID(phone, types.DefaultUserServer)

	info, err := client.WAClient.GetProfilePictureInfo(ctx, jid, &whatsmeow.GetProfilePictureParams{
		Preview: false,
	})
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error obteniendo foto: %v", err))
	}
	if info == nil {
		return nil, errors.New(404, "Foto de perfil no encontrada")
	}

	return &models.ProfilePictureInfo{
		URL: info.URL,
		ID:  info.ID,
	}, nil
}

// GetContacts lista los contactos sincronizados
func (s *ContactService) GetContacts(ctx context.Context, instanceID string) ([]models.ContactInfo, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	contacts, err := client.WAClient.Store.Contacts.GetAllContacts(ctx)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error obteniendo contactos: %v", err))
	}

	result := make([]models.ContactInfo, 0, len(contacts))
	for jid, c := range contacts {
		result = append(result, models.ContactInfo{
			JID:          jid.String(),
			Found:        true,
			FirstName:    c.FirstName,
			FullName:     c.FullName,
			PushName:     c.PushName,
			BusinessName: c.BusinessName,
		})
	}

	return result, nil
}

// SubscribePresence suscribe a actualizaciones de presencia de un usuario
func (s *ContactService) SubscribePresence(ctx context.Context, instanceID, phone string) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	jid := types.NewJID(phone, types.DefaultUserServer)

	if err := client.WAClient.SubscribePresence(ctx, jid); err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error suscribiendo presencia: %v", err))
	}

	return nil
}

// BlockContact bloquea un contacto
func (s *ContactService) BlockContact(ctx context.Context, instanceID, phone string) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	jid := types.NewJID(phone, types.DefaultUserServer)

	if _, err := client.WAClient.UpdateBlocklist(ctx, jid, events.BlocklistChangeActionBlock); err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error bloqueando contacto: %v", err))
	}

	return nil
}

// UnblockContact desbloquea un contacto
func (s *ContactService) UnblockContact(ctx context.Context, instanceID, phone string) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	jid := types.NewJID(phone, types.DefaultUserServer)

	if _, err := client.WAClient.UpdateBlocklist(ctx, jid, events.BlocklistChangeActionUnblock); err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error desbloqueando contacto: %v", err))
	}

	return nil
}

// CheckNumbers verifica si números de teléfono están registrados en WhatsApp
func (s *ContactService) CheckNumbers(ctx context.Context, instanceID string, req *models.CheckNumbersRequest) (*models.CheckNumbersResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	results := make([]models.CheckNumberResult, 0, len(req.Phones))

	// Procesar cada número
	for _, phone := range req.Phones {
		// Limpiar el número
		phone = strings.TrimPrefix(phone, "+")

		// Verificar si el número está en WhatsApp usando IsOnWhatsApp
		jids, err := client.WAClient.IsOnWhatsApp(ctx, []string{phone})
		if err != nil {
			// Si hay error, marcar como no existente
			results = append(results, models.CheckNumberResult{
				Phone:  phone,
				Exists: false,
				JID:    nil,
			})
			continue
		}

		// Verificar si se encontró el número
		if len(jids) > 0 && jids[0].IsIn {
			jidStr := jids[0].JID.String()
			results = append(results, models.CheckNumberResult{
				Phone:  phone,
				Exists: true,
				JID:    &jidStr,
			})
		} else {
			results = append(results, models.CheckNumberResult{
				Phone:  phone,
				Exists: false,
				JID:    nil,
			})
		}
	}

	return &models.CheckNumbersResponse{
		Success: true,
		Data:    results,
	}, nil
}

// GetContactInfo obtiene información detallada de un contacto
func (s *ContactService) GetContactInfo(ctx context.Context, instanceID, phone string) (*models.ContactInfo, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	jid, err := types.ParseJID(phone)
	if err != nil {
		jid, err = types.ParseJID(phone + "@s.whatsapp.net")
		if err != nil {
			return nil, errors.ErrBadRequest.WithDetails("Número de teléfono inválido")
		}
	}

	// 1. Obtener info básica del store
	contact, err := client.WAClient.Store.Contacts.GetContact(ctx, jid)
	if err != nil {
		// Si no está en el store, usar valores por defecto
		contact = types.ContactInfo{Found: false}
	}

	// 2. Obtener estado (About) - OMITIDO POR AHORA
	var status string
	// about, err := client.WAClient.GetUserProfile(ctx, []types.JID{jid})
	// if err == nil && len(about) > 0 {
	// 	status = about[0].Status
	// }

	// 3. Obtener foto de perfil
	var picURL string
	picInfo, err := client.WAClient.GetProfilePictureInfo(ctx, jid, &whatsmeow.GetProfilePictureParams{Preview: true})
	if err == nil && picInfo != nil {
		picURL = picInfo.URL
	}

	// 4. Verificar si es business (aproximación por JID o info)
	isBusiness := strings.Contains(jid.String(), "business") || contact.BusinessName != ""

	// 5. Verificar si está bloqueado
	isBlocked := false
	blocklist, err := client.WAClient.GetBlocklist(ctx)
	if err == nil {
		for _, blocked := range blocklist.JIDs {
			if blocked.User == jid.User {
				isBlocked = true
				break
			}
		}
	}

	return &models.ContactInfo{
		JID:          jid.String(),
		Phone:        jid.User,
		Found:        contact.Found,
		FirstName:    contact.FirstName,
		FullName:     contact.FullName,
		PushName:     contact.PushName,
		BusinessName: contact.BusinessName,
		Status:       status,
		PictureURL:   picURL,
		IsBusiness:   isBusiness,
		IsBlocked:    isBlocked,
	}, nil
}

// GetAbout obtiene el estado/about de un contacto
func (s *ContactService) GetAbout(ctx context.Context, instanceID, phone string) (*models.AboutResponse, error) {
	// Como no tenemos método confiable para GetAbout en esta versión, retornamos no implementado o vacío
	return &models.AboutResponse{About: ""}, nil
}
