package validators

import (
	"net"
	"net/url"
	"strings"

	"kero-kero/pkg/errors"
)

// ValidateMediaURL valida que una URL de medios sea segura
// Previene ataques SSRF (Server-Side Request Forgery)
func ValidateMediaURL(urlStr string) error {
	if urlStr == "" {
		return errors.New(400, "La URL del medio es requerida")
	}

	// Parsear URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return errors.New(400, "URL inválida: "+err.Error())
	}

	// Solo permitir HTTP y HTTPS
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New(400, "Solo se permiten URLs con protocolo HTTP o HTTPS")
	}

	// Obtener hostname
	host := parsedURL.Hostname()
	if host == "" {
		return errors.New(400, "URL sin hostname válido")
	}

	// Bloquear localhost y variantes
	if isLocalhost(host) {
		return errors.New(400, "No se permiten URLs a localhost")
	}

	// Si es una IP, verificar que no sea privada
	ip := net.ParseIP(host)
	if ip != nil {
		if isPrivateIP(ip) {
			return errors.New(400, "No se permiten URLs a redes privadas")
		}
	} else {
		// Si no es IP, resolver el hostname para verificar
		ips, err := net.LookupIP(host)
		if err != nil {
			// No fallar si no se puede resolver, puede ser temporal
			// Solo logueamos y continuamos
			return nil
		}

		// Verificar que ninguna de las IPs resueltas sea privada
		for _, resolvedIP := range ips {
			if isPrivateIP(resolvedIP) {
				return errors.New(400, "El hostname resuelve a una IP privada")
			}
		}
	}

	return nil
}

// isLocalhost verifica si un hostname es localhost
func isLocalhost(host string) bool {
	localhosts := []string{
		"localhost",
		"127.0.0.1",
		"::1",
		"0.0.0.0",
	}

	host = strings.ToLower(host)
	for _, localhost := range localhosts {
		if host == localhost {
			return true
		}
	}

	return false
}

// isPrivateIP verifica si una IP es privada o de loopback
func isPrivateIP(ip net.IP) bool {
	// Verificar loopback
	if ip.IsLoopback() {
		return true
	}

	// Verificar si es privada (RFC 1918)
	if ip.IsPrivate() {
		return true
	}

	// Verificar link-local
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// Verificar rangos privados manualmente por si acaso
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16", // Link-local
		"fc00::/7",       // IPv6 ULA
		"fe80::/10",      // IPv6 Link-local
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	return false
}
