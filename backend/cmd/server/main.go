package main

import (
	"backend/internal/api"
	"backend/internal/cache"
	"log"
	"net/http"
	"time"

	"github.com/rs/cors"
)

var redisCache *cache.RedisCache

func initRedis() {
	var err error
	redisCache, err = cache.NewRedisCache("127.0.0.1:6379", "", 0, 1*time.Hour)
	if err != nil {
		log.Fatalf("❌ Redis 连接失败: %v", err)
	}
	log.Println("✅ Redis 已连接")
}

func main() {
	initRedis()

	api.StartProgressBroadcaster() // 启动进度推送协程

	//http.HandleFunc("/checkAll", api.CheckAllHandler)
	http.HandleFunc("/checkAll", func(w http.ResponseWriter, r *http.Request) {
		api.CheckAllHandler(w, r, redisCache)
	})
	http.HandleFunc("/api/recent", api.HandleRecentScans)
	// 新增临时数据接口
	http.HandleFunc("/store-temp-data", api.StoreTempDataHandler)
	http.HandleFunc("/get-temp-data", api.GetTempDataHandler)

	http.HandleFunc("/api/uploadCsvAndExportJsonl", api.UploadCsvAndExportJsonlHandler)

	http.HandleFunc("/ws/testconnect", api.WSConnectHandler)
	http.HandleFunc("/ws/checkall-progress", api.WSCheckAllProgressHandler)

	// 静态文件托管
	http.Handle("/downloads/", http.StripPrefix("/downloads/", http.FileServer(http.Dir("downloads"))))

	// 启用 CORS
	corsHandler := cors.Default().Handler(http.DefaultServeMux)

	log.Println("Server is running on :8081")
	log.Fatal(http.ListenAndServe("[::]:8081", corsHandler))
}
