package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"

	"kero-kero/internal/models"
	"kero-kero/internal/whatsapp"
	"kero-kero/pkg/errors"
)

type GroupService struct {
	waManager *whatsapp.Manager
}

func NewGroupService(waManager *whatsapp.Manager) *GroupService {
	return &GroupService{waManager: waManager}
}

// CreateGroup crea un nuevo grupo
func (s *GroupService) CreateGroup(ctx context.Context, instanceID string, req *models.CreateGroupRequest) (*models.GroupResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	participants := make([]types.JID, len(req.Participants))
	for i, p := range req.Participants {
		participants[i] = types.NewJID(p, types.DefaultUserServer)
	}

	resp, err := client.WAClient.CreateGroup(ctx, whatsmeow.ReqCreateGroup{
		Name:         req.Name,
		Participants: participants,
	})
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error creando grupo: %v", err))
	}

	return &models.GroupResponse{
		JID:  resp.JID.String(),
		Name: req.Name,
	}, nil
}

// ListGroups lista los grupos a los que pertenece la instancia
func (s *GroupService) ListGroups(ctx context.Context, instanceID string) ([]*models.GroupResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	groups, err := client.WAClient.GetJoinedGroups(ctx)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error obteniendo grupos: %v", err))
	}

	var result []*models.GroupResponse
	for _, g := range groups {
		result = append(result, &models.GroupResponse{
			JID:              g.JID.String(),
			Name:             g.Name,
			OwnerJID:         g.OwnerJID.String(),
			Topic:            g.Topic,
			CreationTime:     g.GroupCreated.Unix(),
			ParticipantCount: len(g.Participants),
		})
	}

	return result, nil
}

// GetGroupInfo obtiene información detallada de un grupo
func (s *GroupService) GetGroupInfo(ctx context.Context, instanceID, groupJID string) (*models.GroupResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("JID de grupo inválido")
	}

	info, err := client.WAClient.GetGroupInfo(ctx, jid)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error obteniendo info del grupo: %v", err))
	}

	participants := make([]models.GroupParticipant, len(info.Participants))
	for i, p := range info.Participants {
		participants[i] = models.GroupParticipant{
			JID:          p.JID.String(),
			IsAdmin:      p.IsAdmin,
			IsSuperAdmin: p.IsSuperAdmin,
		}
	}

	return &models.GroupResponse{
		JID:              info.JID.String(),
		Name:             info.Name,
		OwnerJID:         info.OwnerJID.String(),
		Topic:            info.Topic,
		CreationTime:     info.GroupCreated.Unix(),
		ParticipantCount: len(info.Participants),
		Participants:     participants,
	}, nil
}

// UpdateParticipants actualiza los participantes de un grupo
func (s *GroupService) UpdateParticipants(ctx context.Context, instanceID, groupJID string, req *models.GroupParticipantRequest) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return errors.ErrBadRequest.WithDetails("JID de grupo inválido")
	}

	participants := make([]types.JID, len(req.Participants))
	for i, p := range req.Participants {
		participants[i] = types.NewJID(p, types.DefaultUserServer)
	}

	var waAction whatsmeow.ParticipantChange
	switch req.Action {
	case "add":
		waAction = whatsmeow.ParticipantChangeAdd
	case "remove":
		waAction = whatsmeow.ParticipantChangeRemove
	case "promote":
		waAction = whatsmeow.ParticipantChangePromote
	case "demote":
		waAction = whatsmeow.ParticipantChangeDemote
	default:
		return errors.ErrBadRequest.WithDetails("Acción inválida")
	}

	_, err = client.WAClient.UpdateGroupParticipants(ctx, jid, participants, waAction)
	if err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error actualizando participantes: %v", err))
	}

	return nil
}

// LeaveGroup sale de un grupo
func (s *GroupService) LeaveGroup(ctx context.Context, instanceID, groupJID string) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return errors.ErrBadRequest.WithDetails("JID de grupo inválido")
	}

	if err := client.WAClient.LeaveGroup(ctx, jid); err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error saliendo del grupo: %v", err))
	}

	return nil
}

// AddParticipants agrega participantes al grupo
func (s *GroupService) AddParticipants(ctx context.Context, instanceID, groupJID string, req *models.AddParticipantsRequest) (*models.GroupActionResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("JID de grupo inválido")
	}

	participants := make([]types.JID, len(req.Participants))
	for i, p := range req.Participants {
		participants[i] = types.NewJID(p, types.DefaultUserServer)
	}

	resp, err := client.WAClient.UpdateGroupParticipants(ctx, jid, participants, whatsmeow.ParticipantChangeAdd)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error agregando participantes: %v", err))
	}

	// Procesar respuesta para ver qué participantes fallaron
	var failed []string
	for _, r := range resp {
		if r.Error != 0 {
			failed = append(failed, r.JID.User)
		}
	}

	return &models.GroupActionResponse{
		Success: len(failed) < len(req.Participants),
		Message: fmt.Sprintf("Agregados %d de %d participantes", len(req.Participants)-len(failed), len(req.Participants)),
		Failed:  failed,
	}, nil
}

// RemoveParticipants remueve participantes del grupo
func (s *GroupService) RemoveParticipants(ctx context.Context, instanceID, groupJID string, req *models.RemoveParticipantsRequest) (*models.GroupActionResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("JID de grupo inválido")
	}

	participants := make([]types.JID, len(req.Participants))
	for i, p := range req.Participants {
		participants[i] = types.NewJID(p, types.DefaultUserServer)
	}

	resp, err := client.WAClient.UpdateGroupParticipants(ctx, jid, participants, whatsmeow.ParticipantChangeRemove)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error removiendo participantes: %v", err))
	}

	var failed []string
	for _, r := range resp {
		if r.Error != 0 {
			failed = append(failed, r.JID.User)
		}
	}

	return &models.GroupActionResponse{
		Success: len(failed) < len(req.Participants),
		Message: fmt.Sprintf("Removidos %d de %d participantes", len(req.Participants)-len(failed), len(req.Participants)),
		Failed:  failed,
	}, nil
}

// PromoteToAdmin promueve participantes a administradores
func (s *GroupService) PromoteToAdmin(ctx context.Context, instanceID, groupJID string, req *models.PromoteAdminRequest) (*models.GroupActionResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("JID de grupo inválido")
	}

	participants := make([]types.JID, len(req.Participants))
	for i, p := range req.Participants {
		participants[i] = types.NewJID(p, types.DefaultUserServer)
	}

	resp, err := client.WAClient.UpdateGroupParticipants(ctx, jid, participants, whatsmeow.ParticipantChangePromote)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error promoviendo participantes: %v", err))
	}

	var failed []string
	for _, r := range resp {
		if r.Error != 0 {
			failed = append(failed, r.JID.User)
		}
	}

	return &models.GroupActionResponse{
		Success: len(failed) < len(req.Participants),
		Message: fmt.Sprintf("Promovidos %d de %d participantes a admin", len(req.Participants)-len(failed), len(req.Participants)),
		Failed:  failed,
	}, nil
}

// DemoteAdmin remueve permisos de administrador
func (s *GroupService) DemoteAdmin(ctx context.Context, instanceID, groupJID string, req *models.DemoteAdminRequest) (*models.GroupActionResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("JID de grupo inválido")
	}

	participants := make([]types.JID, len(req.Participants))
	for i, p := range req.Participants {
		participants[i] = types.NewJID(p, types.DefaultUserServer)
	}

	resp, err := client.WAClient.UpdateGroupParticipants(ctx, jid, participants, whatsmeow.ParticipantChangeDemote)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error degradando admins: %v", err))
	}

	var failed []string
	for _, r := range resp {
		if r.Error != 0 {
			failed = append(failed, r.JID.User)
		}
	}

	return &models.GroupActionResponse{
		Success: len(failed) < len(req.Participants),
		Message: fmt.Sprintf("Degradados %d de %d admins", len(req.Participants)-len(failed), len(req.Participants)),
		Failed:  failed,
	}, nil
}

// UpdateGroupInfo actualiza nombre y/o descripción del grupo
func (s *GroupService) UpdateGroupInfo(ctx context.Context, instanceID, groupJID string, req *models.UpdateGroupRequest) (*models.GroupActionResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("JID de grupo inválido")
	}

	// Actualizar nombre si se proporcionó
	if req.Name != nil && *req.Name != "" {
		err = client.WAClient.SetGroupName(ctx, jid, *req.Name)
		if err != nil {
			return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error actualizando nombre: %v", err))
		}
	}

	// Actualizar descripción si se proporcionó
	if req.Description != nil {
		// SetGroupTopic(ctx, jid, previousID, newID, topic)
		err = client.WAClient.SetGroupTopic(ctx, jid, "", "", *req.Description)
		if err != nil {
			return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error actualizando descripción: %v", err))
		}
	}

	return &models.GroupActionResponse{
		Success: true,
		Message: "Información del grupo actualizada correctamente",
	}, nil
}

// UpdateGroupSettings actualiza configuración del grupo (locked, announce, ephemeral)
func (s *GroupService) UpdateGroupSettings(ctx context.Context, instanceID, groupJID string, req *models.UpdateGroupSettingsRequest) (*models.GroupActionResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("JID de grupo inválido")
	}

	// Actualizar "locked" (solo admins pueden editar info)
	if req.IsLocked != nil {
		err = client.WAClient.SetGroupLocked(ctx, jid, *req.IsLocked)
		if err != nil {
			return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error actualizando locked: %v", err))
		}
	}

	// Actualizar "announce" (solo admins pueden enviar mensajes)
	if req.IsAnnounce != nil {
		err = client.WAClient.SetGroupAnnounce(ctx, jid, *req.IsAnnounce)
		if err != nil {
			return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error actualizando announce: %v", err))
		}
	}

	// Actualizar mensajes temporales
	if req.DisappearingTimer != nil {
		timer := time.Duration(*req.DisappearingTimer) * time.Second
		err = client.WAClient.SetDisappearingTimer(ctx, jid, timer, time.Time{})
		if err != nil {
			return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error actualizando timer: %v", err))
		}
	}

	return &models.GroupActionResponse{
		Success: true,
		Message: "Configuración del grupo actualizada correctamente",
	}, nil
}

// GetInviteLink obtiene el link de invitación del grupo
func (s *GroupService) GetInviteLink(ctx context.Context, instanceID, groupJID string) (*models.InviteLinkResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("JID de grupo inválido")
	}

	code, err := client.WAClient.GetGroupInviteLink(ctx, jid, false)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error obteniendo link de invitación: %v", err))
	}

	return &models.InviteLinkResponse{
		Success:    true,
		InviteLink: "https://chat.whatsapp.com/" + code,
		InviteCode: code,
	}, nil
}

// RevokeInviteLink invalida el link de invitación actual y genera uno nuevo
func (s *GroupService) RevokeInviteLink(ctx context.Context, instanceID, groupJID string) (*models.InviteLinkResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("JID de grupo inválido")
	}

	// Revocar link actual y obtener uno nuevo
	code, err := client.WAClient.GetGroupInviteLink(ctx, jid, true)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error revocando link: %v", err))
	}

	return &models.InviteLinkResponse{
		Success:    true,
		InviteLink: "https://chat.whatsapp.com/" + code,
		InviteCode: code,
	}, nil
}

// JoinGroup se une a un grupo mediante un link o código de invitación.
// Es una función muy útil para automatizar la entrada a grupos de soporte o ventas.
func (s *GroupService) JoinGroup(ctx context.Context, instanceID string, req *models.JoinGroupRequest) (*models.GroupResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	// Extraer el código si pasaron un link completo.
	code := req.InviteCode
	if len(code) > 26 && code[:26] == "https://chat.whatsapp.com/" {
		code = code[26:]
	}

	log.Info().Str("instance_id", instanceID).Str("code", code).Msg("Intentando unirse a grupo vía link")

	jid, err := client.WAClient.JoinGroupWithLink(ctx, code)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("No se pudo unir al grupo (el link podría estar expirado): %v", err))
	}

	// Una vez dentro, intentamos obtener la info del grupo para devolver algo más que un JID.
	info, _ := client.WAClient.GetGroupInfo(ctx, jid)
	name := "Grupo Unido"
	if info != nil {
		name = info.Name
	}

	return &models.GroupResponse{
		JID:  jid.String(),
		Name: name,
	}, nil
}

// UpdateGroupPicture actualiza la foto de perfil de un grupo.
// Soporta tanto una URL externa como una imagen directamente en Base64.
func (s *GroupService) UpdateGroupPicture(ctx context.Context, instanceID, groupJID string, req *models.UpdateGroupPictureRequest) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return errors.ErrBadRequest.WithDetails("JID de grupo inválido")
	}

	var avatar []byte

	// Prioridad 1: Imagen en Base64
	if req.Image != nil && *req.Image != "" {
		data, err := base64.StdEncoding.DecodeString(*req.Image)
		if err != nil {
			return errors.ErrBadRequest.WithDetails("Formato Base64 inválido")
		}
		avatar = data
	} else if req.ImageURL != nil && *req.ImageURL != "" {
		// Prioridad 2: URL de imagen
		resp, err := http.Get(*req.ImageURL)
		if err != nil {
			return errors.ErrBadRequest.WithDetails(fmt.Sprintf("No se pudo descargar la imagen de la URL: %v", err))
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return errors.ErrBadRequest.WithDetails(fmt.Sprintf("La URL devolvió un estado no exitoso: %d", resp.StatusCode))
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.ErrInternalServer.WithDetails("Error leyendo el cuerpo de la imagen")
		}
		avatar = data
	} else {
		return errors.ErrBadRequest.WithDetails("Se debe proporcionar una imagen en Base64 o una URL")
	}

	log.Info().Str("instance_id", instanceID).Str("group_jid", groupJID).Msg("Actualizando foto de grupo")

	_, err = client.WAClient.SetGroupPhoto(ctx, jid, avatar)
	if err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("WhatsApp rechazó la imagen (podría ser muy grande o formato incorrecto): %v", err))
	}

	return nil
}
