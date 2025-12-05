package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// AuthService maneja la autenticación y generación de tokens JWT
type AuthService struct {
	jwtSecret string
	apiKey    string
}

// NewAuthService crea una nueva instancia del servicio de autenticación
func NewAuthService(jwtSecret, apiKey string) *AuthService {
	return &AuthService{
		jwtSecret: jwtSecret,
		apiKey:    apiKey,
	}
}

// JWTClaims representa los claims del token JWT
type JWTClaims struct {
	Subject   string `json:"sub"`           // Usuario o identificador
	IssuedAt  int64  `json:"iat"`           // Tiempo de emisión
	ExpiresAt int64  `json:"exp"`           // Tiempo de expiración
	Type      string `json:"type,omitempty"`// Tipo de token (dashboard, api, etc)
}

// LoginRequest representa la solicitud de login
type LoginRequest struct {
	APIKey string `json:"api_key"`
}

// LoginResponse representa la respuesta del login
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	Type      string `json:"type"`
}

// Login valida las credenciales y genera un token JWT
// Por ahora solo valida la API_KEY, pero se puede extender para usuario/contraseña
func (s *AuthService) Login(apiKey string) (*LoginResponse, error) {
	// Validar que la API key sea correcta
	if apiKey != s.apiKey {
		return nil, fmt.Errorf("credenciales inválidas")
	}

	// Generar token JWT con expiración de 24 horas
	expiresAt := time.Now().Add(24 * time.Hour).Unix()
	
	claims := JWTClaims{
		Subject:   "dashboard",
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: expiresAt,
		Type:      "dashboard",
	}

	token, err := s.generateJWT(claims)
	if err != nil {
		return nil, fmt.Errorf("error generando token: %w", err)
	}

	return &LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		Type:      "Bearer",
	}, nil
}

// ValidateToken valida un token JWT y retorna los claims si es válido
func (s *AuthService) ValidateToken(token string) (*JWTClaims, error) {
	// Remover el prefijo "Bearer " si existe
	token = strings.TrimPrefix(token, "Bearer ")

	// Dividir el token en sus partes: header.payload.signature
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("formato de token inválido")
	}

	// Decodificar el payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("error decodificando payload: %w", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("error parseando claims: %w", err)
	}

	// Verificar que el token no haya expirado
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("token expirado")
	}

	// Verificar la firma
	expectedSignature := s.sign(parts[0] + "." + parts[1])
	if parts[2] != expectedSignature {
		return nil, fmt.Errorf("firma de token inválida")
	}

	return &claims, nil
}

// generateJWT genera un token JWT con los claims proporcionados
func (s *AuthService) generateJWT(claims JWTClaims) (string, error) {
	// Header del JWT
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	// Codificar header
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("error codificando header: %w", err)
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)

	// Codificar payload
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("error codificando claims: %w", err)
	}
	payloadEncoded := base64.RawURLEncoding.EncodeToString(claimsBytes)

	// Crear firma
	message := headerEncoded + "." + payloadEncoded
	signature := s.sign(message)

	// Construir token completo
	return message + "." + signature, nil
}

// sign genera una firma HMAC-SHA256 del mensaje
func (s *AuthService) sign(message string) string {
	h := hmac.New(sha256.New, []byte(s.jwtSecret))
	h.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
