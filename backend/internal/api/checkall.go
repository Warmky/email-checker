package api

import (
	"backend/internal/discover"
	"backend/internal/models"
	"backend/internal/scoring"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// 主要的查询几种机制配置的接口
func CheckAllHandler(w http.ResponseWriter, r *http.Request) { //5.19新增以使得用户无需手动选择机制，三种都查询一遍
	models.ProgressBroadcast <- models.ProgressUpdate{Progress: 0, Stage: "start", Message: "开始检测"}
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email parameter is required", http.StatusBadRequest)
		return
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}
	domain := parts[1]

	//从三种机制的结果中选出最优的放在recently seen 5.19
	var bestScore int
	var bestGrade string

	// AUTODISCOVER 查询
	//var autodiscoverResp map[string]interface{}//7.22
	models.ProgressBroadcast <- models.ProgressUpdate{Progress: 10, Stage: "autodiscover", Message: "开始 Autodiscover 检测"}
	adResults := discover.QueryAutodiscover(domain, email)
	var validResults []map[string]interface{} //7.22
	for _, result := range adResults {
		if result.Config != "" && !strings.HasPrefix(result.Config, "Errorcode") &&
			!strings.HasPrefix(result.Config, "Non-valid") &&
			!strings.HasPrefix(result.Config, "Bad response") {
			score, connectScores, ConnectDetails, PortsUsage := scoring.ScoreConfig(result.Config, *result.CertInfo)
			securityDefense := scoring.EvaluateSecurityDefense(result.URI, *result.CertInfo, result.Config, score, connectScores)
			validResults = append(validResults, map[string]interface{}{ //7.22
				"index":  result.Index,
				"uri":    result.URI,
				"method": result.Method,
				"config": result.Config,
				"score":  score,
				"score_detail": map[string]interface{}{
					"connection":            connectScores,
					"defense":               securityDefense,
					"actualconnect_details": ConnectDetails,
					"ports_usage":           PortsUsage,
				},
				"cert_info": result.CertInfo,
				"redirects": result.Redirects, // ⭐ 加上 redirects 字段9.15_4
			})

			if score["overall"] > bestScore {
				bestScore = score["overall"]
				bestGrade = connectScores["Connection_Grade"].(string)
			}
		}
	}
	var autodiscoverResp map[string]interface{}
	if len(validResults) > 0 {
		best := validResults[0]

		// 类型断言
		scoreDetail, ok := best["score_detail"].(map[string]interface{})
		if !ok {
			log.Println("❌ 类型断言失败: score_detail 不是 map[string]interface{}")
			return
		}

		autodiscoverResp = map[string]interface{}{
			"config": best["config"],
			"score":  best["score"],
			"score_detail": map[string]interface{}{
				"connection":            scoreDetail["connection"],
				"defense":               scoreDetail["defense"],
				"actualconnect_details": scoreDetail["actualconnect_details"],
				"ports_usage":           scoreDetail["ports_usage"],
			},
			"cert_info": best["cert_info"],
			"all":       validResults, // ⭐ 前端可用该字段展示所有路径
		}
	} //7.22
	models.ProgressBroadcast <- models.ProgressUpdate{Progress: 30, Stage: "autodiscover", Message: "Autodiscover 检测完成"}
	// AUTOCONFIG 查询
	//var autoconfigResp map[string]interface{}
	models.ProgressBroadcast <- models.ProgressUpdate{Progress: 40, Stage: "autoconfig", Message: "开始 Autoconfig 检测"}
	acResults := discover.QueryAutoconfig(domain, email)
	var validacResults []map[string]interface{}
	for _, result := range acResults {
		if result.Config != "" {
			score, connectScores, ConnectDetails, PortsUsage := scoring.ScoreConfig_Autoconfig(result.Config, *result.CertInfo)
			securityDefense := scoring.EvaluateSecurityDefense(result.URI, *result.CertInfo, result.Config, score, connectScores)
			validacResults = append(validacResults, map[string]interface{}{ //7.22
				"index":  result.Index,
				"uri":    result.URI,
				"method": result.Method,
				"config": result.Config,
				"score":  score,
				"score_detail": map[string]interface{}{
					"connection":            connectScores,
					"defense":               securityDefense,
					"actualconnect_details": ConnectDetails,
					"ports_usage":           PortsUsage,
				},
				"cert_info": result.CertInfo,
				"redirects": result.Redirects, // ⭐ 同样加上9.15_4
			})

			if score["overall"] > bestScore {
				bestScore = score["overall"]
				bestGrade = connectScores["Connection_Grade"].(string)
			}
			// break
		}
	}
	var autoconfigResp map[string]interface{}
	if len(validacResults) > 0 {
		best := validacResults[0]

		// 类型断言
		scoreDetail, ok := best["score_detail"].(map[string]interface{})
		if !ok {
			log.Println("❌ 类型断言失败: score_detail 不是 map[string]interface{}")
			return
		}

		autoconfigResp = map[string]interface{}{
			"config": best["config"],
			"score":  best["score"],
			"score_detail": map[string]interface{}{
				"connection":            scoreDetail["connection"],
				"defense":               scoreDetail["defense"],
				"actualconnect_details": scoreDetail["actualconnect_details"],
				"ports_usage":           scoreDetail["ports_usage"],
			},
			"cert_info": best["cert_info"],
			"all":       validacResults, // ⭐ 前端可用该字段展示所有路径
		}
	} //7.22
	models.ProgressBroadcast <- models.ProgressUpdate{Progress: 60, Stage: "autoconfig", Message: "Autoconfig 检测完成"}
	// SRV 查询
	models.ProgressBroadcast <- models.ProgressUpdate{Progress: 70, Stage: "srv", Message: "开始 SRV 记录检测"}
	srvResult := discover.QuerySRV(domain)
	var srvResp map[string]interface{}
	// var srvScore map[string]int //
	// var srvConnectScores map[string]interface{} //

	if len(srvResult.RecvRecords) > 0 || len(srvResult.SendRecords) > 0 {
		srvScore, srvConnectScores, ConnectDetails, srvPortsUsage := scoring.ScoreConfig_SRV(srvResult)
		securityDefense := scoring.EvaluateSecurityDefense_SRV(srvResult.DNSRecord, srvScore, srvConnectScores)
		srvResp = map[string]interface{}{
			"score": srvScore,
			"score_detail": map[string]interface{}{
				"connection":            srvConnectScores,
				"defense":               securityDefense,
				"actualconnect_details": ConnectDetails,
				"ports_usage":           srvPortsUsage, //
			},
			"srv_records": map[string]interface{}{
				"recv": srvResult.RecvRecords,
				"send": srvResult.SendRecords,
			},
			"dns_record": srvResult.DNSRecord,
		}
		if srvScore["overall"] > bestScore {
			bestScore = srvScore["overall"]
			// 确保字段存在后再断言类型
			if grade, ok := srvConnectScores["Connection_Grade"].(string); ok {
				bestGrade = grade
			}
		}
	}
	// else {
	// 	srvResp = map[string]interface{}{
	// 		"message": "No SRV records found",
	// 	}
	// }  //5.22
	models.ProgressBroadcast <- models.ProgressUpdate{Progress: 85, Stage: "srv", Message: "SRV 检测完成"}
	//8.10本地添加GUESS部分
	models.ProgressBroadcast <- models.ProgressUpdate{Progress: 90, Stage: "guess", Message: "尝试猜测邮件服务器"}
	var guessResp map[string]interface{}
	guessed := discover.GuessMailServer(domain, 2*time.Second, 20)
	if len(guessed) > 0 {
		fmt.Print(guessed)
		connectScores, ConnectDetails, PortsUsage := scoring.ScoreConfig_Guess(guessed)
		//defenseScore := evaluateSecurityDefense_Guess(guessed, guessScore, connectScores)

		guessResp = map[string]interface{}{
			"results": guessed,
			//"score":   guessScore,
			"score_detail": map[string]interface{}{
				"connection": connectScores,
				//"defense":     defenseScore,
				"actualconnect_details": ConnectDetails,
				"ports_usage":           PortsUsage,
			},
		}

		if grade, ok := connectScores["Connection_Grade"].(string); ok {
			// GUESS 没有 overall 分数，可以用 0 或不更新 score，只更新 grade
			if bestScore < 0 {
				bestScore = 0
				bestGrade = grade
			}
		}

		// if guessScore["overall"] > bestScore {
		// 	bestScore = guessScore["overall"]
		// 	if grade, ok := connectScores["Connection_Grade"].(string); ok {
		// 		bestGrade = grade
		// 	}
		// }
	} else {
		guessResp = map[string]interface{}{
			"message": "No reachable common mail host/port combination found.",
		}
	}
	models.ProgressBroadcast <- models.ProgressUpdate{Progress: 98, Stage: "guess", Message: "猜测完成"}

	// 添加到最近记录
	AddRecentScanWithScore(domain, bestScore, bestGrade) //5.19
	models.ProgressBroadcast <- models.ProgressUpdate{Progress: 100, Stage: "done", Message: "检测完成"}
	// 返回统一结构
	response := map[string]interface{}{
		"autodiscover":  autodiscoverResp,
		"autoconfig":    autoconfigResp,
		"srv":           srvResp,
		"guess":         guessResp,
		"recentResults": GetRecentScans(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// // 处理 Autodiscover 查询请求 4.22
// func autodiscoverHandler(w http.ResponseWriter, r *http.Request) {
// 	email := r.URL.Query().Get("email")
// 	if email == "" {
// 		http.Error(w, "Email parameter is required", http.StatusBadRequest)
// 		return
// 	}
// 	// TODO: 这里调用 Autodiscover 查询逻辑
// 	//首先由用户输入的邮件用户名得到domain
// 	parts := strings.Split(email, "@")
// 	if len(parts) != 2 {
// 		http.Error(w, "Invalid email format", http.StatusBadRequest)
// 		return
// 	}
// 	domain := parts[1]
// 	// atIndex := strings.LastIndex(email, "@")
// 	// if atIndex == -1 {
// 	// 	http.Error(w, "Invalid email format", http.StatusBadRequest)
// 	// 	return
// 	// }
// 	// domain := email[atIndex+1:] //这样更能保证即使用户名里出现了@也可以解析虽然实际上不行

// 	//fmt.Fprintf(w, "Processing Autodiscover for: %s", email)
// 	results := discover.QueryAutodiscover(domain, email)
// 	for _, result := range results {
// 		if result.Config != "" && !strings.HasPrefix(result.Config, "Errorcode") && !strings.HasPrefix(result.Config, "Non-valid") && !strings.HasPrefix(result.Config, "Bad response") { //AUtodiscover机制中找到一个有效的配置就停止了
// 			//fmt.Fprint(w, result.Config)
// 			// 解析 config 并评分
// 			score, connectScores, _, _ := scoring.ScoreConfig(result.Config, *result.CertInfo)
// 			securityDefense := scoring.EvaluateSecurityDefense(result.URI, *result.CertInfo, result.Config, score, connectScores)
// 			// 构造返回 JSON
// 			response := map[string]interface{}{
// 				"config": result.Config, // 这里也可以选择不返回原始 XML，避免前端解析麻烦
// 				"score":  score,
// 				"score_detail": map[string]interface{}{
// 					"connection": connectScores,
// 					"defense":    securityDefense,
// 				},
// 				"cert_info": result.CertInfo,
// 			}
// 			// 添加到最近记录
// 			api.AddRecentScanWithScore(domain, score["overall"], connectScores["Connection_Grade"].(string)) //5.19
// 			// 返回 JSON
// 			w.Header().Set("Content-Type", "application/json")
// 			json.NewEncoder(w).Encode(response)
// 			return
// 		}
// 	}
// 	// 如果没有有效的结果，返回错误信息
// 	http.Error(w, "No valid Autodiscover configuration found", http.StatusNotFound)
// 	//fmt.Fprintf(w, "Processing Autodiscover for: %s", email)
// }

// // 处理 Autoconfig 查询请求
// func autoconfigHandler(w http.ResponseWriter, r *http.Request) {
// 	email := r.URL.Query().Get("email")
// 	if email == "" {
// 		http.Error(w, "Email parameter is required", http.StatusBadRequest)
// 		return
// 	}
// 	parts := strings.Split(email, "@")
// 	if len(parts) != 2 {
// 		http.Error(w, "Invalid email format", http.StatusBadRequest)
// 		return
// 	}
// 	domain := parts[1]

// 	results := discover.QueryAutoconfig(domain, email)
// 	for _, result := range results {
// 		if result.Config != "" {
// 			// 如果评分逻辑和 Autodiscover 不一样,可以另写一个 scoreAutoconfig 函数
// 			score, connectScores, _, _ := scoring.ScoreConfig_Autoconfig(result.Config, *result.CertInfo)

// 			response := map[string]interface{}{
// 				"config": result.Config,
// 				"score":  score,
// 				"score_detail": map[string]interface{}{
// 					"connection": connectScores,
// 				},
// 				"cert_info": result.CertInfo,
// 			}
// 			// 添加到最近记录
// 			api.AddRecentScanWithScore(domain, score["overall"], connectScores["Connection_Grade"].(string)) //5.19
// 			w.Header().Set("Content-Type", "application/json")
// 			json.NewEncoder(w).Encode(response)
// 			return
// 		}
// 	}
// 	http.Error(w, "No valid Autoconfig configuration found", http.StatusNotFound)
// }

// func srvHandler(w http.ResponseWriter, r *http.Request) {
// 	email := r.URL.Query().Get("email")
// 	if email == "" {
// 		http.Error(w, "Email parameter is required", http.StatusBadRequest)
// 		return
// 	}
// 	parts := strings.Split(email, "@")
// 	if len(parts) != 2 {
// 		http.Error(w, "Invalid email format", http.StatusBadRequest)
// 		return
// 	}
// 	domain := parts[1]

// 	result := discover.QuerySRV(domain)
// 	if len(result.RecvRecords) > 0 || len(result.SendRecords) > 0 {
// 		score, connectScores, _, _ := scoring.ScoreConfig_SRV(result)

// 		response := map[string]interface{}{
// 			"score":        score,
// 			"score_detail": map[string]interface{}{"connection": connectScores},
// 			"srv_records": map[string]interface{}{
// 				"recv": result.RecvRecords,
// 				"send": result.SendRecords,
// 			},
// 			"dns_record": result.DNSRecord,
// 		}
// 		// 添加到最近记录
// 		api.AddRecentScanWithScore(domain, score["overall"], connectScores["Connection_Grade"].(string)) //5.19
// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(response)
// 		return
// 	}

// 	// // 返回空记录的结构
// 	// w.Header().Set("Content-Type", "application/json")
// 	// json.NewEncoder(w).Encode(map[string]interface{}{
// 	// 	"message": "No SRV records found",
// 	// })
// 	http.Error(w, "No valid SRV configuration found", http.StatusNotFound)
// }
