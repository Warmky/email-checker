// package api

// import (
// 	"backend/internal/discover"
// 	"backend/internal/models"
// 	"bufio"
// 	"encoding/csv"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"
// 	"path/filepath"
// 	"strings"
// 	"sync"
// 	"time"
// )

// //上传.csv接口

// // 8.18
// // 上传CSV并返回处理结果文件下载链接
// func UploadCsvAndExportJsonlHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	err := r.ParseMultipartForm(10 << 20) // 限制大小为10MB
// 	if err != nil {
// 		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	file, handler, err := r.FormFile("file")
// 	if err != nil {
// 		http.Error(w, "Failed to retrieve file: "+err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	defer file.Close()

// 	// 保存上传的文件到临时目录
// 	os.MkdirAll("tmp", os.ModePerm)
// 	tmpFilePath := filepath.Join("tmp", handler.Filename)
// 	tmpFile, err := os.Create(tmpFilePath)
// 	if err != nil {
// 		http.Error(w, "Failed to create temp file: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	defer tmpFile.Close()
// 	io.Copy(tmpFile, file)

// 	// 重新打开读取（确保读取内容）
// 	tmpFile.Seek(0, 0)
// 	reader := csv.NewReader(bufio.NewReader(tmpFile))

// 	var domains []string
// 	first := true
// 	for {
// 		record, err := reader.Read()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil || len(record) == 0 {
// 			continue
// 		}
// 		domain := record[0]

// 		// ✅ 移除 BOM（只在第一行第一列出现）
// 		if first {
// 			domain = strings.TrimPrefix(domain, "\uFEFF")
// 			first = false
// 		}

// 		domains = append(domains, domain)
// 	}

// 	// 创建输出目录
// 	os.MkdirAll("downloads", os.ModePerm)
// 	timestamp := time.Now().Format("20060102_150405")
// 	outputFile := filepath.Join("downloads", fmt.Sprintf("result_%s.jsonl", timestamp))
// 	out, err := os.Create(outputFile)
// 	if err != nil {
// 		http.Error(w, "Failed to create result file: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	defer out.Close()

// 	// 并发处理每个域名
// 	var wg sync.WaitGroup
// 	var mu sync.Mutex
// 	for _, domain := range domains {
// 		wg.Add(1)
// 		go func(domain string) {
// 			defer wg.Done()
// 			//email := "info@" + domain
// 			result := discover.ProcessDomain(domain)
// 			bytes, err := json.Marshal(result)
// 			if err != nil {
// 				return
// 			}
// 			mu.Lock()
// 			out.Write(append(bytes, '\n'))
// 			mu.Unlock()
// 		}(domain)
// 	}
// 	wg.Wait()

// 	// 返回结果文件路径
// 	downloadURL := "/downloads/" + filepath.Base(outputFile)
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]string{
// 		"download_url": downloadURL,
// 	})
// }

package api

import (
	"backend/internal/discover"
	"backend/internal/models"
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// UploadCsvAndExportJsonlHandler 处理 CSV 上传并返回 JSONL 下载链接
func UploadCsvAndExportJsonlHandler(w http.ResponseWriter, r *http.Request) {
	// 统一设置 CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 处理 OPTIONS 预检请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent) // 204
		return
	}

	// 只允许 POST 上传
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析 multipart form
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 获取文件
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to retrieve file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 保存到 tmp 目录
	os.MkdirAll("tmp", os.ModePerm)
	tmpFilePath := filepath.Join("tmp", handler.Filename)
	tmpFile, err := os.Create(tmpFilePath)
	if err != nil {
		http.Error(w, "Failed to create temp file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer tmpFile.Close()
	io.Copy(tmpFile, file)

	// 读取 CSV
	tmpFile.Seek(0, 0)
	reader := csv.NewReader(bufio.NewReader(tmpFile))
	var domains []string
	first := true
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(record) == 0 {
			continue
		}
		domain := record[0]
		if first {
			domain = strings.TrimPrefix(domain, "\uFEFF") // 移除 BOM
			first = false
		}
		domains = append(domains, domain)
	}

	// 输出文件
	os.MkdirAll("downloads", os.ModePerm)
	timestamp := time.Now().Format("20060102_150405")
	outputFile := filepath.Join("downloads", fmt.Sprintf("result_%s.jsonl", timestamp))
	out, err := os.Create(outputFile)
	if err != nil {
		http.Error(w, "Failed to create result file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// 并发处理每个域名
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, domain := range domains {
		wg.Add(1)
		go func(domain string) {
			defer wg.Done()
			result := discover.ProcessDomain(domain)
			bytes, err := json.Marshal(result)
			if err != nil {
				return
			}
			mu.Lock()
			out.Write(append(bytes, '\n'))
			mu.Unlock()
		}(domain)
	}
	wg.Wait()

	// 返回下载链接
	downloadURL := "/downloads/" + filepath.Base(outputFile)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"download_url": downloadURL,
	})
}

// func UploadCsvAndExportJsonlHandler(w http.ResponseWriter, r *http.Request) {
// 	// 打开日志文件
// 	logDir := "/home/wzq/email-checker/backend/log"
// 	os.MkdirAll(logDir, os.ModePerm)
// 	logFile, err1 := os.OpenFile("log/upload_csv.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
// 	if err1 != nil {
// 		log.Println("Failed to open log file:", err1)
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
// 	defer logFile.Close()

// 	logger := log.New(logFile, "", log.LstdFlags)

// 	if r.Method != http.MethodPost {
// 		logger.Println("Method not allowed:", r.Method)
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	err := r.ParseMultipartForm(10 << 20) // 限制大小为10MB
// 	if err != nil {
// 		logger.Println("ParseMultipartForm failed:", err)
// 		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	file, handler, err := r.FormFile("file")
// 	if err != nil {
// 		logger.Println("FormFile failed:", err)
// 		http.Error(w, "Failed to retrieve file: "+err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	defer file.Close()

// 	// 保存上传的文件到临时目录
// 	os.MkdirAll("tmp", os.ModePerm)
// 	tmpFilePath := filepath.Join("tmp", handler.Filename)
// 	tmpFile, err := os.Create(tmpFilePath)
// 	if err != nil {
// 		logger.Println("Create temp file failed:", err)
// 		http.Error(w, "Failed to create temp file: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	defer tmpFile.Close()

// 	if _, err := io.Copy(tmpFile, file); err != nil {
// 		logger.Println("Copy file failed:", err)
// 		http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	logger.Println("Uploaded file successfully:", handler.Filename)

// 	// 重新打开读取（确保读取内容）
// 	tmpFile.Seek(0, 0)
// 	reader := csv.NewReader(bufio.NewReader(tmpFile))

// 	var domains []string
// 	first := true
// 	for {
// 		record, err := reader.Read()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil || len(record) == 0 {
// 			logger.Println("CSV read error:", err, "record:", record)
// 			continue
// 		}
// 		domain := record[0]
// 		if first {
// 			domain = strings.TrimPrefix(domain, "\uFEFF")
// 			first = false
// 		}
// 		domains = append(domains, domain)
// 	}

// 	// 创建输出目录
// 	os.MkdirAll("downloads", os.ModePerm)
// 	timestamp := time.Now().Format("20060102_150405")
// 	outputFile := filepath.Join("downloads", fmt.Sprintf("result_%s.jsonl", timestamp))
// 	out, err := os.Create(outputFile)
// 	if err != nil {
// 		logger.Println("Create output file failed:", err)
// 		http.Error(w, "Failed to create result file: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	defer out.Close()

// 	// 并发处理每个域名
// 	var wg sync.WaitGroup
// 	var mu sync.Mutex
// 	for _, domain := range domains {
// 		wg.Add(1)
// 		go func(domain string) {
// 			defer wg.Done()
// 			result := discover.ProcessDomain(domain)
// 			bytes, err := json.Marshal(result)
// 			if err != nil {
// 				logger.Println("JSON marshal failed for domain:", domain, err)
// 				return
// 			}
// 			mu.Lock()
// 			out.Write(append(bytes, '\n'))
// 			mu.Unlock()
// 		}(domain)
// 	}
// 	wg.Wait()

// 	// 返回结果文件路径
// 	downloadURL := "/downloads/" + filepath.Base(outputFile)
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]string{
// 		"download_url": downloadURL,
// 	})
// 	logger.Println("Processed domains:", len(domains), "output:", downloadURL)
// }

func UploadCSVHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 最大 10MB
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 创建临时文件保存上传内容
	tmpPath := filepath.Join("tmp", fmt.Sprintf("upload_%d.csv", time.Now().UnixNano()))
	outFile, err := os.Create(tmpPath)
	if err != nil {
		http.Error(w, "Failed to save uploaded file", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()
	io.Copy(outFile, file)

	log.Printf("Uploaded CSV: %s (%d bytes)", handler.Filename, handler.Size)

	// 处理上传的 CSV 文件并生成结果
	jsonlPath, err := processCSVAndExport(tmpPath)
	if err != nil {
		http.Error(w, "Processing failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回结果文件路径
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"download_url": "/downloads/" + filepath.Base(jsonlPath),
	})
}

func processCSVAndExport(csvPath string) (string, error) {
	f, err := os.Open(csvPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	var domains []string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if len(record) > 0 {
			domains = append(domains, record[0])
		}
	}

	// 结果文件路径
	outputFile := filepath.Join("downloads", fmt.Sprintf("result_%d.jsonl", time.Now().UnixNano()))
	outFile, err := os.Create(outputFile)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	var wg sync.WaitGroup
	var fileMutex sync.Mutex

	for _, domain := range domains {
		wg.Add(1)
		models.Semaphore <- struct{}{}
		go func(domain string) {
			defer wg.Done()
			defer func() { <-models.Semaphore }()

			result := discover.ProcessDomain(domain) // 使用你已有的函数
			jsonBytes, err := json.Marshal(result)
			if err != nil {
				log.Printf("Marshal error for %s: %v", domain, err)
				return
			}

			fileMutex.Lock()
			outFile.Write(append(jsonBytes, '\n'))
			fileMutex.Unlock()
		}(domain)
	}

	wg.Wait()
	return outputFile, nil
}
