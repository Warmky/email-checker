package api

//少量存储最近查询过的域名
import (
	"backend/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func AddRecentScanWithScore(domain string, score int, grade string) {
	if models.RecentScans == nil {
		models.RecentScans = []models.ScanHistory{}
	}

	filtered := []models.ScanHistory{}
	for _, item := range models.RecentScans {
		if item.Domain != domain {
			filtered = append(filtered, item)
		}
	}
	models.RecentScans = append([]models.ScanHistory{{
		Domain:    domain,
		Timestamp: time.Now(),
		Score:     score,
		Grade:     grade,
	}}, filtered...)

	if len(models.RecentScans) > models.MaxRecent {
		models.RecentScans = models.RecentScans[:models.MaxRecent]
	}
}

func GetRecentScans() []models.ScanHistory {
	return models.RecentScans
}

func HandleRecentScans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if models.RecentScans == nil {
		json.NewEncoder(w).Encode([]models.ScanHistory{})
	} else {
		json.NewEncoder(w).Encode(models.RecentScans)
	}
}

// 接收完整结构
func StoreTempDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	id := fmt.Sprintf("%d", time.Now().UnixNano()) // 唯一 ID
	models.TempDataStore[id] = payload

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

// 获取完整结构
func GetTempDataHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}

	data, ok := models.TempDataStore[id]
	if !ok {
		http.Error(w, "Data not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
