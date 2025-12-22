package services

import (
	"context"
	"fmt"
	"time"

	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/types"

	"encoding/json"
	"kero-kero/internal/models"
	"kero-kero/internal/repository"
	"kero-kero/internal/whatsapp"
	"kero-kero/pkg/errors"
)

type BusinessService struct {
	waManager   *whatsapp.Manager
	redisClient *repository.RedisClient
}

func NewBusinessService(waManager *whatsapp.Manager, redisClient *repository.RedisClient) *BusinessService {
	return &BusinessService{
		waManager:   waManager,
		redisClient: redisClient,
	}
}

// CreateLabel crea o edita una etiqueta de WhatsApp Business.
// He usado el helper BuildLabelEdit de whatsmeow que genera el patch de AppState necesario.
func (s *BusinessService) CreateLabel(ctx context.Context, instanceID string, req *models.CreateLabelRequest) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	// Generar un ID único para la etiqueta si es nueva, o usar uno existente si lo soportáramos.
	// Por simplicidad, usamos el timestamp actual como ID.
	labelID := fmt.Sprintf("%d", time.Now().Unix())

	patch := appstate.BuildLabelEdit(labelID, req.Name, req.Color, false)

	err := client.WAClient.SendAppState(ctx, patch)
	if err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error sincronizando etiqueta: %v", err))
	}

	return nil
}

// AssignLabel asigna o remueve una etiqueta de un chat.
func (s *BusinessService) AssignLabel(ctx context.Context, instanceID string, req *models.LabelActionRequest) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}

	chatJID, err := types.ParseJID(req.ChatJID)
	if err != nil {
		return errors.ErrBadRequest.WithDetails("JID de chat inválido")
	}

	labeled := req.Action == "add"
	patch := appstate.BuildLabelChat(chatJID, req.LabelID, labeled)
	err = client.WAClient.SendAppState(ctx, patch)
	if err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error al (des)etiquetar chat: %v", err))
	}

	return nil
}

// GetBusinessProfile obtiene la información del perfil de empresa.
func (s *BusinessService) GetBusinessProfile(ctx context.Context, instanceID, jidStr string) (*types.BusinessProfile, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	jid, err := types.ParseJID(jidStr)
	if err != nil {
		jid = client.WAClient.Store.ID.ToNonAD() // Si no pasan JID, usamos el nuestro
	}

	profile, err := client.WAClient.GetBusinessProfile(ctx, jid)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("No se pudo obtener el perfil de empresa: %v", err))
	}

	return profile, nil
}

// SetAutoLabelRules guarda las reglas de etiquetado automático para una instancia.
func (s *BusinessService) SetAutoLabelRules(ctx context.Context, instanceID string, rules []models.AutoLabelRule) error {
	rulesJSON, err := json.Marshal(rules)
	if err != nil {
		return errors.ErrInternalServer.WithDetails("Error codificando reglas")
	}

	return s.redisClient.SetAutoLabelRules(ctx, instanceID, string(rulesJSON))
}

// GetAutoLabelRules obtiene las reglas para una instancia.
func (s *BusinessService) GetAutoLabelRules(ctx context.Context, instanceID string) ([]models.AutoLabelRule, error) {
	rulesJSON, err := s.redisClient.GetAutoLabelRules(ctx, instanceID)
	if err != nil {
		return nil, nil // Opcional: manejar error de redis.Nil
	}

	var rules []models.AutoLabelRule
	if err := json.Unmarshal([]byte(rulesJSON), &rules); err != nil {
		return nil, errors.ErrInternalServer.WithDetails("Error decodificando reglas")
	}

	return rules, nil
}
