package api

import (
	"encoding/json"
	"net/http"
)

// handleDevices returns all tracked devices.
func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		devices := s.mgr.GetDevices()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(devices)
		return
	}

	if r.Method == http.MethodPut {
		var req struct {
			IP   string `json:"ip"`
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := s.mgr.SetDeviceName(req.IP, req.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleDeviceHistory returns connection history for devices.
func (s *Server) handleDeviceHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	deviceIP := r.URL.Query().Get("device_ip")
	proxyID := r.URL.Query().Get("proxy_id")
	limit := 100 // default limit

	var history interface{}
	var err error

	if deviceIP != "" {
		// Get history for specific device
		history, err = s.mgr.GetConnectionHistory(deviceIP, limit)
	} else {
		// Get all history, optionally filtered by proxy
		history, err = s.mgr.GetAllConnectionHistory(proxyID, limit)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(history)
}
