package main

import (
	"crypto/tls"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

type CertificateResponse struct {
	Certificate string `json:"certificate"`
}

func fetchCertHandler(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Query().Get("host")
	if host == "" {
		http.Error(w, "Missing 'host' parameter", http.StatusBadRequest)
		return
	}

	// Устанавливаем TLS-соединение с указанным хостом
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{
			Timeout: 5 * time.Second,
		},
		Config: &tls.Config{
			InsecureSkipVerify: true, // Пропускаем проверку сертификата
		},
	}

	conn, err := dialer.Dial("tcp", host)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect: %v", err), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		http.Error(w, "Failed to assert tls.Conn", http.StatusInternalServerError)
		return
	}

	// Завершаем TLS-рукопожатие
	if err := tlsConn.Handshake(); err != nil {
		http.Error(w, fmt.Sprintf("TLS handshake failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Получаем сертификаты сервера
	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		http.Error(w, "No certificates found", http.StatusInternalServerError)
		return
	}

	// Кодируем первый сертификат в PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: state.PeerCertificates[0].Raw,
	})

	resp := CertificateResponse{
		Certificate: string(certPEM),
	}

	// Возвращаем сертификат в ответе
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/cert", fetchCertHandler)
	log.Println("cert-fetcher is running on :8080")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
