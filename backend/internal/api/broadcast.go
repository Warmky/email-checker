package api

import (
	"backend/internal/models"
	"log"
	"net/http"
)

// 查询时广播进度接口
// ✅ 启动广播协程
func StartProgressBroadcaster() {
	go func() {
		for update := range models.ProgressBroadcast {
			for conn := range models.ProgressClients {
				payload := map[string]interface{}{
					"type":     "progress",
					"progress": update.Progress,
					"stage":    update.Stage,
					"message":  update.Message,
				}
				if err := conn.WriteJSON(payload); err != nil {
					log.Println("WebSocket write error:", err)
					conn.Close()
					delete(models.ProgressClients, conn)
				}
			}
		}
	}()
}

// ✅ WebSocket Handler
func WSCheckAllProgressHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := models.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to websocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	models.ProgressClients[conn] = true
	log.Println("客户端已连接进度 WebSocket")
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket 关闭:", err)
			delete(models.ProgressClients, conn)
			break
		}
	}
}
