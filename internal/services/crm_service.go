package services

import (
	"context"
	"sync"
	"time"

	"kero-kero/internal/models"
	"kero-kero/pkg/errors"
)

type CRMService struct {
	contacts map[string]*models.CRMContact // Key: instanceID:JID
	mu       sync.RWMutex
}

func NewCRMService() *CRMService {
	return &CRMService{
		contacts: make(map[string]*models.CRMContact),
	}
}

func (s *CRMService) getKey(instanceID, jid string) string {
	return instanceID + ":" + jid
}

// ListContacts lista los contactos del CRM para una instancia
func (s *CRMService) ListContacts(ctx context.Context, instanceID string) ([]*models.CRMContact, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var contacts []*models.CRMContact
	for _, c := range s.contacts {
		if c.InstanceID == instanceID {
			contacts = append(contacts, c)
		}
	}
	return contacts, nil
}

// GetContact obtiene un contacto específico
func (s *CRMService) GetContact(ctx context.Context, instanceID, jid string) (*models.CRMContact, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if contact, ok := s.contacts[s.getKey(instanceID, jid)]; ok {
		return contact, nil
	}
	return nil, errors.ErrNotFound
}

// UpdateContact crea o actualiza un contacto
func (s *CRMService) UpdateContact(ctx context.Context, instanceID string, jid string, req *models.UpdateCRMContactRequest) (*models.CRMContact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.getKey(instanceID, jid)
	contact, exists := s.contacts[key]

	if !exists {
		contact = &models.CRMContact{
			JID:        jid,
			InstanceID: instanceID,
			CreatedAt:  time.Now(),
		}
	}

	// Actualizar campos
	if req.Name != "" {
		contact.Name = req.Name
	}
	if req.Email != "" {
		contact.Email = req.Email
	}
	if req.Notes != "" {
		contact.Notes = req.Notes
	}
	if req.Status != "" {
		contact.Status = req.Status
	}
	if req.Tags != nil {
		contact.Tags = req.Tags
	}

	contact.UpdatedAt = time.Now()
	s.contacts[key] = contact

	return contact, nil
}

// DeleteContact elimina un contacto del CRM
func (s *CRMService) DeleteContact(ctx context.Context, instanceID, jid string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.contacts, s.getKey(instanceID, jid))
	return nil
}
