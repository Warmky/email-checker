package api

import (
	"backend/internal/models"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// configViewPage中的实时连接接口
func WSConnectHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := models.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to websocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	query := r.URL.Query()
	host := query.Get("host")
	portStr := query.Get("port")
	protocol := query.Get("protocol")
	mode := query.Get("mode") // ✅ 新增：加密方式

	port, err := strconv.Atoi(portStr)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Invalid port"))
		return
	}

	// LiveTestConnection(host, port, protocol, conn)
	LiveTestConnection(host, port, protocol, mode, conn)
}

// 8.20
func SendLog(conn *websocket.Conn, msg string) {
	payload := map[string]interface{}{
		"type":    "log",
		"content": msg,
	}
	b, _ := json.Marshal(payload)
	conn.WriteMessage(websocket.TextMessage, b)
}

func LiveTestConnection(host string, port int, protocol string, mode string, conn *websocket.Conn) {
	SendLog(conn, fmt.Sprintf("🔍 开始连接测试：%s:%d [%s, %s]", host, port, protocol, mode))
	time.Sleep(300 * time.Millisecond)

	SendLog(conn, "⚙️ 执行连接测试脚本...")
	time.Sleep(300 * time.Millisecond)

	log.Printf("Start testing host=%s, port=%d, protocol=%s, mode=%s", host, port, protocol, mode)

	_, result, err := RunZGrab2WithProgress(protocol, host, fmt.Sprintf("%d", port), mode, conn) // ✅ 使用 mode
	// if err != nil {
	// 	SendLog(conn, fmt.Sprintf("❌ 测试失败：%v", err))
	// 	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "测试失败"))
	// 	return
	// }

	// SendLog(conn, "✅ 测试完成，准备返回结构化结果")
	if err != nil {
		// 即使失败，也构造 result
		result = &models.ConnectInfo{
			Success: false,
			Info:    nil,
			Error:   err.Error(),
		}
		SendLog(conn, fmt.Sprintf("❌ 测试失败：%v", err))
	}

	payload := map[string]interface{}{
		"type":   "result",
		"result": result,
	}
	finalJSON, _ := json.Marshal(payload)
	conn.WriteMessage(websocket.TextMessage, finalJSON)
}

func RunZGrab2WithProgress(protocol, hostname, port, mode string, conn *websocket.Conn) (bool, *models.ConnectInfo, error) {
	SendLog(conn, "📦 正在执行 TLS 检测脚本...")

	pythonPath := "python" // 视你的系统为 python 或 python3
	scriptPath := "/home/wzq/email-checker/backend/tlscheck/test_tls.py"

	cmd := exec.Command(pythonPath, scriptPath,
		"--protocol", protocol,
		"--host", hostname,
		"--port", port,
		"--mode", mode,
	)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return false, nil, fmt.Errorf("stdout pipe error: %v", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return false, nil, fmt.Errorf("stderr pipe error: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return false, nil, fmt.Errorf("cmd start error: %v", err)
	}

	var outputBuf bytes.Buffer
	var wg sync.WaitGroup

	wg.Add(2)

	// 读取 stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			outputBuf.WriteString(line + "\n")
			SendLog(conn, "📄 "+line)
		}
	}()

	// 读取 stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			SendLog(conn, "⚠️ "+line)
		}
	}()

	// 等待命令执行完成
	if err := cmd.Wait(); err != nil {
		wg.Wait() // 等待读取完成
		return false, nil, fmt.Errorf("执行失败：%v", err)
	}

	wg.Wait() // 确保所有输出已读完

	// 解析 JSON 输出
	var result models.ConnectInfo
	if err := json.Unmarshal(outputBuf.Bytes(), &result); err != nil {
		return false, nil, fmt.Errorf("结果解析失败（非 JSON）：%v", err)
	}

	if !result.Success {
		return false, nil, fmt.Errorf("TLS 测试失败：%s", result.Error)
	}

	return true, &result, nil
}
