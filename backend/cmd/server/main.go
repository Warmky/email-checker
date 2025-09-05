package main

import (
	"backend/internal/api"
	"log"
	"net/http"

	"github.com/rs/cors"
)

func main() {
	api.StartProgressBroadcaster() // 启动进度推送协程

	http.HandleFunc("/checkAll", api.CheckAllHandler)
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
