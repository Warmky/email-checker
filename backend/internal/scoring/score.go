package scoring

//score评分相关函数
import (
	"backend/internal/models"
	"backend/internal/utils"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/etree"
)

func calculateConnectScores(config string) (map[string]interface{}, []models.ConnectDetail) {
	var allConnectionDetails []models.ConnectDetail
	scores := make(map[string]interface{})
	doc := etree.NewDocument()
	if err := doc.ReadFromString(config); err != nil {
		scores["error"] = 0
		return scores, nil
	}
	root := doc.SelectElement("Autodiscover")
	if root == nil {
		scores["error"] = 0
		return scores, nil
	}
	responseElem := root.SelectElement("Response")
	if responseElem == nil {
		scores["error"] = 0
		return scores, nil
	}
	accountElem := responseElem.SelectElement("Account")
	if accountElem == nil {
		scores["error"] = 0
		return scores, nil
	}
	accountTypeElem := accountElem.SelectElement("AccountType")
	if accountTypeElem == nil || accountTypeElem.Text() != "email" {
		scores["error"] = 0
		return scores, nil
	}
	actionElem := accountElem.SelectElement("Action")
	if actionElem == nil || actionElem.Text() != "settings" {
		scores["error"] = 0
		return scores, nil
	}

	// 加密端口列表（不应该允许 plain）
	plaintextNotAllowedPorts := map[string]bool{
		"993": true, // IMAPS
		"995": true, // POP3S
		"465": true, // SMTPS
	}

	// 统计项
	successTLS := 0
	successPlain := 0
	successStartTLS := 0
	totalProtocols := 0

	// 用于记录警告
	warnings := []string{}

	for _, protocolElem := range accountElem.SelectElements("Protocol") {
		protocolType := ""
		port := ""
		hostname := ""
		if typeElem := protocolElem.SelectElement("Type"); typeElem != nil {
			protocolType = strings.ToLower(typeElem.Text())
		}
		if portElem := protocolElem.SelectElement("Port"); portElem != nil {
			port = portElem.Text()
		}
		if serverElem := protocolElem.SelectElement("Server"); serverElem != nil {
			hostname = serverElem.Text()
		}

		if protocolType == "" || port == "" || hostname == "" {
			continue
		}
		totalProtocols++

		// canConnectPlain, errPlain := RunZGrab2(protocolType, hostname, port, "plain")
		// canConnectStartTLS, errStartTLS := RunZGrab2(protocolType, hostname, port, "starttls")
		// canConnectTLS, errTLS := RunZGrab2(protocolType, hostname, port, "ssl") //8.12
		canConnectPlain, plainInfo, errPlain := utils.RunZGrab2WithResult(protocolType, hostname, port, "plain")
		canConnectStartTLS, starttlsInfo, errStartTLS := utils.RunZGrab2WithResult(protocolType, hostname, port, "starttls")
		canConnectTLS, tlsInfo, errTLS := utils.RunZGrab2WithResult(protocolType, hostname, port, "ssl")
		// 将结果附加
		if plainInfo == nil {
			plainInfo = &models.ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if starttlsInfo == nil {
			starttlsInfo = &models.ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if tlsInfo == nil {
			tlsInfo = &models.ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}

		detail := models.ConnectDetail{
			Type:     protocolType,
			Host:     hostname,
			Port:     port,
			Plain:    *plainInfo,
			StartTLS: *starttlsInfo,
			TLS:      *tlsInfo,
		}
		allConnectionDetails = append(allConnectionDetails, detail)

		if errPlain != nil {
			fmt.Println("Plain err:", errPlain)
		}
		if errStartTLS != nil {
			fmt.Println("StartTLS err:", errStartTLS)
		}
		if errTLS != nil {
			fmt.Println("TLS err:", errTLS)
		}

		if utils.IsNoSuchHostError(errPlain) || utils.IsNoSuchHostError(errStartTLS) || utils.IsNoSuchHostError(errTLS) {
			warningMsg := fmt.Sprintf("Warning: The hostname '%s' cannot be resolved. It may be expired or misconfigured.", hostname)
			warnings = append(warnings, warningMsg)
		}

		if canConnectTLS {
			successTLS++
		}
		if canConnectStartTLS {
			successStartTLS++
		}
		if canConnectPlain {
			successPlain++
			// ⚠️ 如果该端口是加密端口，却允许明文连接，添加警告
			if plaintextNotAllowedPorts[port] {
				warningMsg := fmt.Sprintf("Warning: Port %s on %s (type %s) allows plaintext connections, which is insecure.", port, hostname, protocolType)
				warnings = append(warnings, warningMsg)
			}
		}
	}

	if totalProtocols == 0 {
		scores["Connection_Grade"] = "T"
		scores["Overall_Connection_Score"] = 0
		scores["timestamp"] = time.Now().Format(time.RFC3339)
		scores["warnings"] = warnings
		return scores, nil
	}

	tlsScore := (successTLS * 100) / totalProtocols
	starttlsScore := (successStartTLS * 100) / totalProtocols
	plainScore := (successPlain * 100) / totalProtocols
	//overall := (tlsScore*3 + starttlsScore*2 + plainScore*1) / 6
	overall := (tlsScore*3+starttlsScore*2)/5 - plainScore/2
	if overall < 0 {
		overall = 0
	}
	if overall > 100 {
		overall = 100
	}

	grade := "F"
	switch {
	case tlsScore == 100 && plainScore == 0:
		grade = "A+"
	case tlsScore >= 80 && plainScore == 0:
		grade = "A"
	case tlsScore >= 50 && plainScore == 0:
		grade = "B"
	case plainScore > 0:
		grade = "C" // 任何明文可连上，直接降级
	default:
		grade = "F"
	}
	//9.11_2 修改后端对实际连接测试的评分标准，降低可以明文连接的比重

	scores["TLS_Connections"] = tlsScore
	scores["Plaintext_Connections"] = plainScore
	scores["STARTTLS_Connections"] = starttlsScore
	scores["Overall_Connection_Score"] = overall
	scores["Connection_Grade"] = grade
	scores["timestamp"] = time.Now().Format(time.RFC3339)
	scores["warnings"] = warnings

	return scores, allConnectionDetails
}

func ScoreConfig(config string, certInfo models.CertInfo) (map[string]int, map[string]interface{}, []models.ConnectDetail, []models.PortUsageDetail) {
	scores := make(map[string]int)
	//score_connect_Detail := make(map[string]interface{})
	// 计算端口评分
	portScores, PortsUsage := calculatePortScores(config)
	scores["encrypted_ports"] = portScores["encrypted_ports"]
	scores["standard_ports"] = portScores["standard_ports"]

	// 计算证书评分
	certScores := calculateCertScores(certInfo)
	scores["cert_score"] = certScores["cert"]

	//计算实际连接测试评分
	connectScores, ConnectDetails := calculateConnectScores(config)
	// if overall, ok := connectScores["Overall_Connection_Score"].(int); ok {
	// 	scores["connect_score"] = overall
	// } else {
	// 	scores["connect_score"] = 0
	// }
	// score_connect_Detail["connection"] = connectScores
	// 计算最终评分（例如加权平均）
	scores["connect_score"] = connectScores["Overall_Connection_Score"].(int)
	scores["overall"] = (scores["encrypted_ports"] + scores["standard_ports"] + scores["cert_score"] + scores["connect_score"]) / 4

	//return scores, score_connect_Detail
	return scores, connectScores, ConnectDetails, PortsUsage
}

func ScoreConfig_Autoconfig(config string, certInfo models.CertInfo) (map[string]int, map[string]interface{}, []models.ConnectDetail, []models.PortUsageDetail) {
	scores := make(map[string]int)
	//score_connect_Detail := make(map[string]interface{})
	// 计算端口评分
	portScores, PortsUsage := calculatePortScores_Autoconfig(config)
	scores["encrypted_ports"] = portScores["encrypted_ports"]
	scores["standard_ports"] = portScores["standard_ports"]

	// 计算证书评分
	certScores := calculateCertScores(certInfo)
	scores["cert_score"] = certScores["cert"]

	//计算实际连接测试评分
	connectScores, ConnectDetails := calculateConnectScores_Autoconfig(config)
	// if overall, ok := connectScores["Overall_Connection_Score"].(int); ok {
	// 	scores["connect_score"] = overall
	// } else {
	// 	scores["connect_score"] = 0
	// }
	// score_connect_Detail["connection"] = connectScores
	// 计算最终评分（例如加权平均）
	scores["connect_score"] = connectScores["Overall_Connection_Score"].(int)
	scores["overall"] = (scores["encrypted_ports"] + scores["standard_ports"] + scores["cert_score"] + scores["connect_score"]) / 4

	//return scores, score_connect_Detail
	return scores, connectScores, ConnectDetails, PortsUsage
}

func calculatePortScores_Autoconfig(config string) (map[string]int, []models.PortUsageDetail) {
	scores := make(map[string]int)

	doc := etree.NewDocument()
	if err := doc.ReadFromString(config); err != nil {
		scores["error"] = 0
		return scores, nil
	}
	//这里是评分规则
	root := doc.SelectElement("clientConfig")
	if root == nil {
		scores["error"] = 0
		return scores, nil
	}
	emailProviderElem := root.SelectElement("emailProvider")
	if emailProviderElem == nil {
		scores["error"] = 0
		return scores, nil
	}
	var portsUsage []models.PortUsageDetail
	// 记录使用的端口情况
	securePorts := map[string]bool{
		"SMTP": false,
		"IMAP": false,
		"POP3": false,
	}
	insecurePorts := map[string]bool{
		"SMTP": false,
		"IMAP": false,
		"POP3": false,
	}
	nonStandardPorts := map[string]bool{
		"SMTP": false,
		"IMAP": false,
		"POP3": false,
	}
	//var protocols []ProtocolInfo
	for _, protocolElem := range emailProviderElem.SelectElements("incomingServer") {
		//protocol := ProtocolInfo{}
		protocolType := ""
		port := ""
		host := ""
		ssl := ""
		// 检查每个子元素是否存在再获取其内容
		if typeELem := protocolElem.SelectAttr("type"); typeELem != nil {
			protocolType = typeELem.Value //? type属性 -> <Type>
		}
		if serverElem := protocolElem.SelectElement("hostname"); serverElem != nil {
			host = serverElem.Text() //<hostname> -> <Server>
		}
		if portElem := protocolElem.SelectElement("port"); portElem != nil {
			port = portElem.Text()
		}
		if sslElem := protocolElem.SelectElement("socketType"); sslElem != nil {
			ssl = sslElem.Text()
		} else {
			ssl = "N/A"
		} //7.27
		status := "nonstandard"
		// 分类端口
		switch protocolType {
		case "smtp":
			if port == "465" {
				status = "secure"
				securePorts["SMTP"] = true
			} else if port == "25" || port == "587" {
				status = "insecure"
				insecurePorts["SMTP"] = true
			} else {
				nonStandardPorts["SMTP"] = true
			}
		case "imap":
			if port == "993" {
				status = "secure"
				securePorts["IMAP"] = true
			} else if port == "143" {
				status = "insecure"
				insecurePorts["IMAP"] = true
			} else {
				nonStandardPorts["IMAP"] = true
			}
		case "pop3":
			if port == "995" {
				status = "secure"
				securePorts["POP3"] = true
			} else if port == "110" {
				status = "insecure"
				insecurePorts["POP3"] = true
			} else {
				nonStandardPorts["POP3"] = true
			}
		}
		if protocolType != "" && port != "" {
			portsUsage = append(portsUsage, models.PortUsageDetail{
				Protocol: strings.ToTitle(protocolType),
				Port:     port,
				Status:   status,
				Host:     host,
				SSL:      ssl,
			})
		} //全部记录到新增结构中
	}

	for _, protocolElem := range emailProviderElem.SelectElements("outgoingServer") {
		//protocol := ProtocolInfo{}
		protocolType := ""
		port := ""
		host := ""
		ssl := ""
		// 检查每个子元素是否存在再获取其内容
		if typeELem := protocolElem.SelectAttr("type"); typeELem != nil {
			protocolType = typeELem.Value //? type属性 -> <Type>
		}
		if serverElem := protocolElem.SelectElement("hostname"); serverElem != nil {
			host = serverElem.Text() //<hostname> -> <Server>
		}
		if portElem := protocolElem.SelectElement("port"); portElem != nil {
			port = portElem.Text()
		}
		if sslElem := protocolElem.SelectElement("socketType"); sslElem != nil {
			ssl = sslElem.Text()
		} else {
			ssl = "N/A"
		}
		status := "nonstandard"
		// 分类端口
		switch protocolType {
		case "smtp":
			if port == "465" {
				status = "secure"
				securePorts["SMTP"] = true
			} else if port == "25" || port == "587" {
				status = "insecure"
				insecurePorts["SMTP"] = true
			} else {
				nonStandardPorts["SMTP"] = true
			}
		case "imap":
			if port == "993" {
				status = "secure"
				securePorts["IMAP"] = true
			} else if port == "143" {
				status = "insecure"
				insecurePorts["IMAP"] = true
			} else {
				nonStandardPorts["IMAP"] = true
			}
		case "pop3":
			if port == "995" {
				status = "secure"
				securePorts["POP3"] = true
			} else if port == "110" {
				status = "insecure"
				insecurePorts["POP3"] = true
			} else {
				nonStandardPorts["POP3"] = true
			}
		}
		if protocolType != "" && port != "" {
			portsUsage = append(portsUsage, models.PortUsageDetail{
				Protocol: strings.ToTitle(protocolType),
				Port:     port,
				Status:   status,
				Host:     host,
				SSL:      ssl,
			})
		} //全部记录到新增结构中
	}

	// 计算加密端口评分
	secureCount := 0
	insecureCount := 0
	nonStandardCount := 0

	for _, v := range securePorts {
		if v {
			secureCount++
		}
	}
	for _, v := range insecurePorts {
		if v {
			insecureCount++
		}
	}
	for _, v := range nonStandardPorts {
		if v {
			nonStandardCount++
		}
	}

	// 评分逻辑
	secureOnly := insecureCount == 0
	secureAndInsecure := secureCount > 0 && insecureCount > 0
	onlyInsecure := secureCount == 0
	hasNonStandard := nonStandardCount > 0

	var encryptionScore int
	if secureOnly {
		encryptionScore = 100
	} else if secureAndInsecure {
		encryptionScore = 60
	} else if onlyInsecure {
		encryptionScore = 10
	} else {
		encryptionScore = 0
	}
	var standardScore int
	if hasNonStandard {
		if len(nonStandardPorts) == 1 {
			standardScore = 80
		} else if len(nonStandardPorts) == 2 {
			standardScore = 60
		} else {
			standardScore = 50
		}
	} else {
		standardScore = 100
	}
	scores["encrypted_ports"] = encryptionScore
	scores["standard_ports"] = standardScore
	return scores, portsUsage
}

func calculateConnectScores_Autoconfig(config string) (map[string]interface{}, []models.ConnectDetail) {
	var allConnectionDetails []models.ConnectDetail
	scores := make(map[string]interface{})
	//这里为了方便又将config解析了一遍，后面应该和之前的端口评分合并
	doc := etree.NewDocument()
	if err := doc.ReadFromString(config); err != nil {
		scores["error"] = 0
		return scores, nil
	}
	root := doc.SelectElement("clientConfig")
	if root == nil {
		scores["error"] = 0
		return scores, nil
	}
	emailProviderElem := root.SelectElement("emailProvider")
	if emailProviderElem == nil {
		scores["error"] = 0
		return scores, nil
	}
	// 加密端口列表（不应该允许 plain）
	plaintextNotAllowedPorts := map[string]bool{
		"993": true, // IMAPS
		"995": true, // POP3S
		"465": true, // SMTPS
	}

	// 遍历每个 Protocol 进行连接测试(三种模式都会尝试)
	successTLS := 0
	successPlain := 0
	successStartTLS := 0
	totalProtocols := 0
	// 用于记录警告
	warnings := []string{}

	for _, protocolElem := range emailProviderElem.SelectElements("incomingServer") {
		//遍历每个protocol来获取连接测试需要的协议类型、端口、主机名
		protocolType := ""
		port := ""
		hostname := ""
		if typeElem := protocolElem.SelectAttr("type"); typeElem != nil {
			protocolType = strings.ToLower(typeElem.Value)
			fmt.Println("ProtocolType:", protocolType)
		}
		if portElem := protocolElem.SelectElement("port"); portElem != nil {
			port = portElem.Text()
			fmt.Println("Port:", port)
		}
		if serverElem := protocolElem.SelectElement("hostname"); serverElem != nil {
			hostname = serverElem.Text()
			fmt.Println("Hostname:", hostname)
		}
		// 确保数据完整
		if protocolType == "" || port == "" || hostname == "" {
			continue
		}

		totalProtocols++
		// canConnectPlain, errPlain := RunZGrab2(protocolType, hostname, port, "plain") //
		// canConnectStartTLS, errStartTLS := RunZGrab2(protocolType, hostname, port, "starttls")
		// canConnectTLS, errTLS := RunZGrab2(protocolType, hostname, port, "ssl") //
		canConnectPlain, plainInfo, errPlain := utils.RunZGrab2WithResult(protocolType, hostname, port, "plain")
		canConnectStartTLS, starttlsInfo, errStartTLS := utils.RunZGrab2WithResult(protocolType, hostname, port, "starttls")
		canConnectTLS, tlsInfo, errTLS := utils.RunZGrab2WithResult(protocolType, hostname, port, "ssl")
		// 将结果附加
		if plainInfo == nil {
			plainInfo = &models.ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if starttlsInfo == nil {
			starttlsInfo = &models.ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if tlsInfo == nil {
			tlsInfo = &models.ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}

		detail := models.ConnectDetail{
			Type:     protocolType,
			Host:     hostname,
			Port:     port,
			Plain:    *plainInfo,
			StartTLS: *starttlsInfo,
			TLS:      *tlsInfo,
		}
		allConnectionDetails = append(allConnectionDetails, detail)

		if utils.IsNoSuchHostError(errPlain) || utils.IsNoSuchHostError(errStartTLS) || utils.IsNoSuchHostError(errTLS) {
			warningMsg := fmt.Sprintf("Warning: The hostname '%s' cannot be resolved. It may be expired or misconfigured.", hostname)
			warnings = append(warnings, warningMsg)
		}
		// 统计连接成功情况
		if canConnectTLS {
			successTLS++
		}
		if canConnectStartTLS {
			successStartTLS++
		}
		if canConnectPlain {
			successPlain++
			// ⚠️ 如果该端口是加密端口，却允许明文连接，添加警告
			if plaintextNotAllowedPorts[port] {
				warningMsg := fmt.Sprintf("Warning: Port %s on %s (type %s) allows plaintext connections, which is insecure.", port, hostname, protocolType)
				warnings = append(warnings, warningMsg)
			}
		}

	}

	for _, protocolElem := range emailProviderElem.SelectElements("outgoingServer") {
		//遍历每个protocol来获取连接测试需要的协议类型、端口、主机名
		protocolType := ""
		port := ""
		hostname := ""
		if typeElem := protocolElem.SelectAttr("type"); typeElem != nil {
			protocolType = strings.ToLower(typeElem.Value)
			fmt.Println("ProtocolType:", protocolType)
		}
		if portElem := protocolElem.SelectElement("port"); portElem != nil {
			port = portElem.Text()
			fmt.Println("Port:", port)
		}
		if serverElem := protocolElem.SelectElement("hostname"); serverElem != nil {
			hostname = serverElem.Text()
			fmt.Println("Hostname:", hostname)
		}
		// 确保数据完整
		if protocolType == "" || port == "" || hostname == "" {
			continue
		}

		totalProtocols++
		// canConnectPlain, errPlain := RunZGrab2(protocolType, hostname, port, "plain")
		// canConnectStartTLS, errStartTLS := RunZGrab2(protocolType, hostname, port, "starttls")
		// canConnectTLS, errTLS := RunZGrab2(protocolType, hostname, port, "ssl")
		canConnectPlain, plainInfo, errPlain := utils.RunZGrab2WithResult(protocolType, hostname, port, "plain")
		canConnectStartTLS, starttlsInfo, errStartTLS := utils.RunZGrab2WithResult(protocolType, hostname, port, "starttls")
		canConnectTLS, tlsInfo, errTLS := utils.RunZGrab2WithResult(protocolType, hostname, port, "ssl")
		// 将结果附加
		if plainInfo == nil {
			plainInfo = &models.ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if starttlsInfo == nil {
			starttlsInfo = &models.ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if tlsInfo == nil {
			tlsInfo = &models.ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}

		detail := models.ConnectDetail{
			Type:     protocolType,
			Host:     hostname,
			Port:     port,
			Plain:    *plainInfo,
			StartTLS: *starttlsInfo,
			TLS:      *tlsInfo,
		}
		allConnectionDetails = append(allConnectionDetails, detail)
		if utils.IsNoSuchHostError(errPlain) || utils.IsNoSuchHostError(errStartTLS) || utils.IsNoSuchHostError(errTLS) {
			warningMsg := fmt.Sprintf("Warning: The hostname '%s' cannot be resolved. It may be expired or misconfigured.", hostname)
			warnings = append(warnings, warningMsg)
		}
		// 统计连接成功情况
		if canConnectTLS {
			successTLS++
		}
		if canConnectStartTLS {
			successStartTLS++
		}
		if canConnectPlain {
			successPlain++
			// ⚠️ 如果该端口是加密端口，却允许明文连接，添加警告
			if plaintextNotAllowedPorts[port] {
				warningMsg := fmt.Sprintf("Warning: Port %s on %s (type %s) allows plaintext connections, which is insecure.", port, hostname, protocolType)
				warnings = append(warnings, warningMsg)
			}

		}

	}
	if totalProtocols == 0 {
		scores["Connection_Grade"] = "T"
		scores["Overall_Connection_Score"] = 0
		scores["timestamp"] = time.Now().Format(time.RFC3339)
		scores["warnings"] = warnings
		return scores, nil
	}
	// 计算评分
	tlsScore := (successTLS * 100) / totalProtocols // 100 分制
	starttlsScore := (successStartTLS * 100) / totalProtocols
	plainScore := (successPlain * 100) / totalProtocols
	//overall := (tlsScore*3 + starttlsScore*2 + plainScore*1) / 6
	overall := (tlsScore*3+starttlsScore*2)/5 - plainScore/2
	if overall < 0 {
		overall = 0
	}
	if overall > 100 {
		overall = 100
	}

	grade := "F"
	switch {
	case tlsScore == 100 && plainScore == 0:
		grade = "A+"
	case tlsScore >= 80 && plainScore == 0:
		grade = "A"
	case tlsScore >= 50 && plainScore == 0:
		grade = "B"
	case plainScore > 0:
		grade = "C" // 任何明文可连上，直接降级
	default:
		grade = "F"
	}
	//9.11_2
	scores["TLS_Connections"] = tlsScore
	scores["Plaintext_Connections"] = plainScore
	scores["STARTTLS_Connections"] = starttlsScore
	scores["Overall_Connection_Score"] = overall
	scores["Connection_Grade"] = grade
	scores["timestamp"] = time.Now().Format(time.RFC3339)
	scores["warnings"] = warnings
	return scores, allConnectionDetails
}

func ScoreConfig_SRV(result models.SRVResult) (map[string]int, map[string]interface{}, []models.ConnectDetail, []models.PortUsageDetail) {
	scores := make(map[string]int)
	//score_connect_Detail := make(map[string]interface{})
	// 计算端口评分
	portScores, PortsUsage := calculatePortScores_SRV(result)
	scores["encrypted_ports"] = portScores["encrypted_ports"]
	scores["standard_ports"] = portScores["standard_ports"]

	//计算DNS记录的DNSSEC评分
	dnssecScores := calculateDNSSECScores_SRV(result)
	scores["dnssec_score"] = dnssecScores

	//计算实际连接测试评分
	connectScores, ConnectDetails := calculateConnectScores_SRV(result)

	// 计算最终评分（例如加权平均）
	scores["connect_score"] = connectScores["Overall_Connection_Score"].(int)
	scores["overall"] = (scores["encrypted_ports"] + scores["standard_ports"] + scores["cert_score"] + scores["connect_score"]) / 4

	return scores, connectScores, ConnectDetails, PortsUsage
}

func calculatePortScores_SRV(result models.SRVResult) (map[string]int, []models.PortUsageDetail) {
	scores := make(map[string]int)

	securePorts := map[string]bool{}
	insecurePorts := map[string]bool{}
	nonStandardPorts := map[string]bool{}

	standardEncrypted := map[uint16]bool{993: true, 995: true, 465: true}
	standardInsecure := map[uint16]bool{143: true, 110: true, 25: true, 587: true}
	var portsUsage []models.PortUsageDetail
	allRecords := append(result.RecvRecords, result.SendRecords...)
	for _, record := range allRecords {
		port := record.Port
		status := Identify_Port_Status(record)
		fmt.Print(status)
		if standardEncrypted[port] {
			securePorts[record.Service] = true
		} else if standardInsecure[port] {
			insecurePorts[record.Service] = true
		} else {
			nonStandardPorts[record.Service] = true
		}
		portsUsage = append(portsUsage, models.PortUsageDetail{
			Protocol: normalizeProtocol(record.Service),
			Port:     strconv.Itoa(int(port)),
			Status:   status,
		})
	}

	secureCount := len(securePorts)
	insecureCount := len(insecurePorts)
	nonStandardCount := len(nonStandardPorts)

	var encryptionScore int
	if secureCount > 0 && insecureCount == 0 {
		encryptionScore = 100
	} else if secureCount > 0 && insecureCount > 0 {
		encryptionScore = 60
	} else if secureCount == 0 && insecureCount > 0 {
		encryptionScore = 10
	} else {
		encryptionScore = 0
	}

	var standardScore int
	if nonStandardCount == 0 {
		standardScore = 100
	} else if nonStandardCount == 1 {
		standardScore = 80
	} else if nonStandardCount == 2 {
		standardScore = 60
	} else {
		standardScore = 50
	}

	scores["encrypted_ports"] = encryptionScore
	scores["standard_ports"] = standardScore
	fmt.Print(portsUsage)
	return scores, portsUsage
}

func normalizeProtocol(service string) string {
	if strings.HasPrefix(service, "_submission") || strings.HasPrefix(service, "_submissions") {
		return "SMTP"
	} else if strings.HasPrefix(service, "_imap") || strings.HasPrefix(service, "_imaps") {
		return "IMAP"
	} else if strings.HasPrefix(service, "_pop3") || strings.HasPrefix(service, "_pop3s") {
		return "POP3"
	}
	return "OTHER"
}

func Identify_Port_Status(record models.SRVRecord) string {
	port := record.Port
	service_prefix := strings.Split(record.Service, ".")[0]
	var status string
	switch service_prefix {
	case "_submissions":
		if port == 465 {
			status = "secure"
		} else {
			status = "nonstandard"
		}
	case "_submission":
		if port == 25 || port == 587 {
			status = "insecure"
		} else {
			status = "nonstandard"
		}
	case "_imaps":
		if port == 993 {
			status = "secure"
		} else {
			status = "nonstandard"
		}
	case "_imap":
		if port == 143 {
			status = "insecure"
		} else {
			status = "nonstandard"
		}
	case "_pop3s":
		if port == 995 {
			status = "secure"
		} else {
			status = "nonstandard"
		}
	case "_pop3":
		if port == 110 {
			status = "insecure"
		} else {
			status = "nonstandard"
		}
	}
	return status
}

func calculateDNSSECScores_SRV(result models.SRVResult) int {
	if result.DNSRecord == nil {
		return 0
	}
	trueCount := 0
	total := 0
	adBits := []*bool{
		result.DNSRecord.ADbit_imap,
		result.DNSRecord.ADbit_imaps,
		result.DNSRecord.ADbit_pop3,
		result.DNSRecord.ADbit_pop3s,
		result.DNSRecord.ADbit_smtp,
		result.DNSRecord.ADbit_smtps,
	}
	for _, bit := range adBits {
		if bit != nil {
			total++
			if *bit {
				trueCount++
			}
		}
	}
	if total == 0 {
		return 0
	}
	return int(float64(trueCount) / float64(total) * 100)
}

func calculateConnectScores_SRV(result models.SRVResult) (map[string]interface{}, []models.ConnectDetail) {
	var allConnectionDetails []models.ConnectDetail
	scores := make(map[string]interface{})
	// 加密端口列表（不应该允许 plain）
	plaintextNotAllowedPorts := map[string]bool{
		"993": true, // IMAPS
		"995": true, // POP3S
		"465": true, // SMTPS
	}
	successTLS := 0
	successPlain := 0
	successStartTLS := 0
	totalProtocols := 0
	// 用于记录警告
	warnings := []string{}
	testRecords := append(result.RecvRecords, result.SendRecords...)
	for _, record := range testRecords {
		protocol := detectProtocolFromService(record.Service)
		if protocol == "" || record.Port == 0 || record.Target == "" {
			continue
		}

		portStr := fmt.Sprintf("%d", record.Port)
		hostname := record.Target

		totalProtocols++

		// canConnectPlain, _ := RunZGrab2(protocol, hostname, portStr, "plain")
		// canConnectStartTLS, _ := RunZGrab2(protocol, hostname, portStr, "starttls")
		// canConnectTLS, _ := RunZGrab2(protocol, hostname, portStr, "tls")
		// canConnectPlain, _ := RunZGrab2(protocol, hostname, portStr, "plain")
		// canConnectStartTLS, _ := RunZGrab2(protocol, hostname, portStr, "starttls")
		// canConnectTLS, _ := RunZGrab2(protocol, hostname, portStr, "ssl")
		canConnectPlain, plainInfo, errPlain := utils.RunZGrab2WithResult(protocol, hostname, portStr, "plain")
		canConnectStartTLS, starttlsInfo, errStartTLS := utils.RunZGrab2WithResult(protocol, hostname, portStr, "starttls")
		canConnectTLS, tlsInfo, errTLS := utils.RunZGrab2WithResult(protocol, hostname, portStr, "ssl")
		// 将结果附加
		if plainInfo == nil {
			plainInfo = &models.ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if starttlsInfo == nil {
			starttlsInfo = &models.ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if tlsInfo == nil {
			tlsInfo = &models.ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}

		detail := models.ConnectDetail{
			Type:     protocol,
			Host:     hostname,
			Port:     portStr,
			Plain:    *plainInfo,
			StartTLS: *starttlsInfo,
			TLS:      *tlsInfo,
		}
		allConnectionDetails = append(allConnectionDetails, detail)

		if errPlain != nil {
			fmt.Println("Plain err:", errPlain)
		}
		if errStartTLS != nil {
			fmt.Println("StartTLS err:", errStartTLS)
		}
		if errTLS != nil {
			fmt.Println("TLS err:", errTLS)
		}

		if utils.IsNoSuchHostError(errPlain) || utils.IsNoSuchHostError(errStartTLS) || utils.IsNoSuchHostError(errTLS) {
			warningMsg := fmt.Sprintf("Warning: The hostname '%s' cannot be resolved. It may be expired or misconfigured.", hostname)
			warnings = append(warnings, warningMsg)
		}

		if canConnectTLS {
			successTLS++
		}
		if canConnectStartTLS {
			successStartTLS++
		}
		if canConnectPlain {
			successPlain++
			// ⚠️ 如果该端口是加密端口，却允许明文连接，添加警告
			if plaintextNotAllowedPorts[portStr] {
				warningMsg := fmt.Sprintf("Warning: Port %s on %s (type %s) allows plaintext connections, which is insecure.", portStr, hostname, protocol)
				warnings = append(warnings, warningMsg)
			}
		}
	}

	if totalProtocols == 0 {
		scores["Connection_Grade"] = "T"
		scores["Overall_Connection_Score"] = 0
		scores["timestamp"] = time.Now().Format(time.RFC3339)
		scores["warnings"] = warnings
		return scores, nil
	}

	tlsScore := (successTLS * 100) / totalProtocols
	starttlsScore := (successStartTLS * 100) / totalProtocols
	plainScore := (successPlain * 100) / totalProtocols
	//overall := (tlsScore*3 + starttlsScore*2 + plainScore*1) / 6
	overall := (tlsScore*3+starttlsScore*2)/5 - plainScore/2
	if overall < 0 {
		overall = 0
	}
	if overall > 100 {
		overall = 100
	}

	grade := "F"
	switch {
	case tlsScore == 100 && plainScore == 0:
		grade = "A+"
	case tlsScore >= 80 && plainScore == 0:
		grade = "A"
	case tlsScore >= 50 && plainScore == 0:
		grade = "B"
	case plainScore > 0:
		grade = "C" // 任何明文可连上，直接降级
	default:
		grade = "F"
	}
	//9.11_2

	scores["TLS_Connections"] = tlsScore
	scores["Plaintext_Connections"] = plainScore
	scores["STARTTLS_Connections"] = starttlsScore
	scores["Overall_Connection_Score"] = overall
	scores["Connection_Grade"] = grade
	scores["timestamp"] = time.Now().Format(time.RFC3339)
	scores["warnings"] = warnings
	return scores, allConnectionDetails
}
func detectProtocolFromService(service string) string {
	// 将常见 SRV 服务名转换为 zgrab2 的协议名
	service = strings.ToLower(service)
	switch {
	case strings.Contains(service, "imap"):
		return "imap"
	case strings.Contains(service, "pop3"):
		return "pop3"
	case strings.Contains(service, "submission"):
		return "smtp"
	}
	return ""
}

func ScoreConfig_Guess(guessed []string) (map[string]interface{}, []models.ConnectDetail, []models.PortUsageDetail) {

	connectScores, ConnectDetails, portsUsage := calculateConnectScores_Guess(guessed)
	return connectScores, ConnectDetails, portsUsage

}

func calculateConnectScores_Guess(guessed []string) (map[string]interface{}, []models.ConnectDetail, []models.PortUsageDetail) {
	var allConnectionDetails []models.ConnectDetail
	scores := make(map[string]interface{})
	var portsUsage []models.PortUsageDetail
	// 加密端口列表（不应该允许 plain）
	plaintextNotAllowedPorts := map[string]bool{
		"993": true, // IMAPS
		"995": true, // POP3S
		"465": true, // SMTPS
	}
	// 遍历每个 Protocol 进行连接测试(三种模式都会尝试)
	successTLS := 0
	successPlain := 0
	successStartTLS := 0
	totalProtocols := 0
	// 用于记录警告
	warnings := []string{}
	for _, hostPort := range guessed {
		parts := strings.Split(hostPort, ":")
		if len(parts) != 2 {
			continue
		}
		hostname := parts[0]
		portStr := parts[1]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}
		protocolType := ""
		status := ""
		if port == 465 {
			protocolType = "smtp"
			status = "secure"
		} else if port == 587 {
			protocolType = "smtp"
			status = "insecure"
		} else if port == 110 {
			protocolType = "pop3"
			status = "insecure"
		} else if port == 995 {
			protocolType = "pop3"
			status = "secure"
		} else if port == 143 {
			protocolType = "imap"
			status = "insecure"
		} else if port == 993 {
			protocolType = "imap"
			status = "secure"
		}
		if protocolType == "" || portStr == "" || hostname == "" {
			continue
		}
		portsUsage = append(portsUsage, models.PortUsageDetail{
			Protocol: protocolType,
			Port:     portStr,
			Status:   status,
			Host:     hostname,
		})
		totalProtocols++
		// canConnectPlain, errPlain := RunZGrab2(protocolType, hostname, portStr, "plain") //
		// canConnectStartTLS, errStartTLS := RunZGrab2(protocolType, hostname, portStr, "starttls")
		// canConnectTLS, errTLS := RunZGrab2(protocolType, hostname, portStr, "ssl") //
		canConnectPlain, plainInfo, errPlain := utils.RunZGrab2WithResult(protocolType, hostname, portStr, "plain")
		canConnectStartTLS, starttlsInfo, errStartTLS := utils.RunZGrab2WithResult(protocolType, hostname, portStr, "starttls")
		canConnectTLS, tlsInfo, errTLS := utils.RunZGrab2WithResult(protocolType, hostname, portStr, "ssl")
		// 将结果附加
		if plainInfo == nil {
			plainInfo = &models.ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if starttlsInfo == nil {
			starttlsInfo = &models.ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if tlsInfo == nil {
			tlsInfo = &models.ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}

		detail := models.ConnectDetail{
			Type:     protocolType,
			Host:     hostname,
			Port:     portStr,
			Plain:    *plainInfo,
			StartTLS: *starttlsInfo,
			TLS:      *tlsInfo,
		}
		allConnectionDetails = append(allConnectionDetails, detail)

		if errPlain != nil {
			fmt.Println("Plain err:", errPlain)
		}
		if errStartTLS != nil {
			fmt.Println("StartTLS err:", errStartTLS)
		}
		if errTLS != nil {
			fmt.Println("TLS err:", errTLS)
		}
		if utils.IsNoSuchHostError(errPlain) || utils.IsNoSuchHostError(errStartTLS) || utils.IsNoSuchHostError(errTLS) {
			warningMsg := fmt.Sprintf("Warning: The hostname '%s' cannot be resolved. It may be expired or misconfigured.", hostname)
			warnings = append(warnings, warningMsg)
		}
		// 统计连接成功情况
		if canConnectTLS {
			successTLS++
		}
		if canConnectStartTLS {
			successStartTLS++
		}
		if canConnectPlain {
			successPlain++
			// ⚠️ 如果该端口是加密端口，却允许明文连接，添加警告
			if plaintextNotAllowedPorts[portStr] {
				warningMsg := fmt.Sprintf("Warning: Port %s on %s (type %s) allows plaintext connections, which is insecure.", portStr, hostname, protocolType)
				warnings = append(warnings, warningMsg)
			}
		}
	}
	if totalProtocols == 0 {
		scores["Connection_Grade"] = "T"
		scores["Overall_Connection_Score"] = 0
		scores["timestamp"] = time.Now().Format(time.RFC3339)
		scores["warnings"] = warnings
		return scores, nil, portsUsage
	}
	// 计算评分
	tlsScore := (successTLS * 100) / totalProtocols // 100 分制
	starttlsScore := (successStartTLS * 100) / totalProtocols
	plainScore := (successPlain * 100) / totalProtocols
	//overall := (tlsScore*3 + starttlsScore*2 + plainScore*1) / 6
	overall := (tlsScore*3+starttlsScore*2)/5 - plainScore/2
	if overall < 0 {
		overall = 0
	}
	if overall > 100 {
		overall = 100
	}

	grade := "F"
	switch {
	case tlsScore == 100 && plainScore == 0:
		grade = "A+"
	case tlsScore >= 80 && plainScore == 0:
		grade = "A"
	case tlsScore >= 50 && plainScore == 0:
		grade = "B"
	case plainScore > 0:
		grade = "C" // 任何明文可连上，直接降级
	default:
		grade = "F"
	}
	//9.11_2

	scores["TLS_Connections"] = tlsScore
	scores["Plaintext_Connections"] = plainScore
	scores["STARTTLS_Connections"] = starttlsScore
	scores["Overall_Connection_Score"] = overall
	scores["Connection_Grade"] = grade
	scores["timestamp"] = time.Now().Format(time.RFC3339)
	scores["warnings"] = warnings
	return scores, allConnectionDetails, portsUsage

}

func calculateCertScores(cert models.CertInfo) map[string]int {
	score := 100 // 最高分
	scores := make(map[string]int)
	// 1. 证书可信度
	if !cert.IsTrusted {
		score -= 30
	}

	// 2. 证书主机名匹配
	if !cert.IsHostnameMatch {
		score -= 20
	}

	// 3. 证书是否过期
	if cert.IsExpired {
		score -= 40
	}

	// 4. 证书是否自签名
	if cert.IsSelfSigned {
		score -= 30
	}

	// 5. TLS 版本检查
	switch cert.TLSVersion {
	case 0x304: // TLS 1.3
		score += 10
	case 0x303: // TLS 1.2
		// 不加分，默认
	case 0x302: // TLS 1.1
		score -= 40
	case 0x301: // TLS 1.0
		score -= 60
	default: // 低于 TLS 1.0 或未知
		score -= 80
	}

	// 限制最低分为 0
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}
	scores["cert"] = score
	return scores
}

func calculatePortScores(config string) (map[string]int, []models.PortUsageDetail) { //增加了一个返回参数[]PortUsageDetail
	scores := make(map[string]int)

	doc := etree.NewDocument()
	if err := doc.ReadFromString(config); err != nil {
		scores["error"] = 0
		return scores, nil
	}
	//这里是评分规则
	root := doc.SelectElement("Autodiscover")
	if root == nil {
		scores["error"] = 0
		return scores, nil
	}
	responseElem := root.SelectElement("Response")
	if responseElem == nil {
		scores["error"] = 0
		return scores, nil
	}
	accountElem := responseElem.SelectElement("Account")
	if accountElem == nil {
		scores["error"] = 0
		return scores, nil
	}
	accountTypeElem := accountElem.SelectElement("AccountType")
	if accountTypeElem == nil || accountTypeElem.Text() != "email" {
		scores["error"] = 0
		return scores, nil
	}
	actionElem := accountElem.SelectElement("Action")
	if actionElem == nil || actionElem.Text() != "settings" {
		scores["error"] = 0
		return scores, nil
	}

	var portsUsage []models.PortUsageDetail
	// 记录使用的端口情况
	securePorts := map[string]bool{
		"SMTP": false,
		"IMAP": false,
		"POP3": false,
	}
	insecurePorts := map[string]bool{
		"SMTP": false,
		"IMAP": false,
		"POP3": false,
	}
	nonStandardPorts := map[string]bool{
		"SMTP": false,
		"IMAP": false,
		"POP3": false,
	}
	//var protocols []ProtocolInfo
	for _, protocolElem := range accountElem.SelectElements("Protocol") {
		//protocol := ProtocolInfo{}
		protocolType := ""
		port := ""
		host := ""
		ssl := ""
		// 检查每个子元素是否存在再获取其内容
		if typeElem := protocolElem.SelectElement("Type"); typeElem != nil {
			protocolType = typeElem.Text()
		}
		if serverElem := protocolElem.SelectElement("Server"); serverElem != nil {
			host = serverElem.Text() //7.27
		}
		if portElem := protocolElem.SelectElement("Port"); portElem != nil {
			port = portElem.Text()
		}
		if encElem := protocolElem.SelectElement("Encryption"); encElem != nil {
			ssl = encElem.Text()
		} else if sslElem := protocolElem.SelectElement("SSL"); sslElem != nil {
			ssl = sslElem.Text()
		} else {
			ssl = "N/A"
		} //7.27
		// if protocol.SSL != "SSL" {
		// 	scores["SSL"] = "HHH"
		// 	//return scores
		// }
		// if protocol.Type == "SMTP" && protocol.Port == "465" {
		// 	scores["SMTPS"] = "yes"
		// }
		// if protocol.Type == "IMAP" && protocol.Port == "993" {
		// 	scores["IMAPS"] = "yes"
		// }
		status := "nonstandard"
		// 分类端口
		switch protocolType {
		case "SMTP":
			if port == "465" {
				status = "secure"
				securePorts["SMTP"] = true
			} else if port == "25" || port == "587" {
				status = "insecure"
				insecurePorts["SMTP"] = true
			} else {
				nonStandardPorts["SMTP"] = true
			}
		case "IMAP":
			if port == "993" {
				status = "secure"
				securePorts["IMAP"] = true
			} else if port == "143" {
				status = "insecure"
				insecurePorts["IMAP"] = true
			} else {
				nonStandardPorts["IMAP"] = true
			}
		case "POP3":
			if port == "995" {
				status = "secure"
				securePorts["POP3"] = true
			} else if port == "110" {
				status = "insecure"
				insecurePorts["POP3"] = true
			} else {
				nonStandardPorts["POP3"] = true
			}
		}
		if protocolType != "" && port != "" {
			portsUsage = append(portsUsage, models.PortUsageDetail{
				Protocol: protocolType,
				Port:     port,
				Status:   status,
				Host:     host,
				SSL:      ssl,
			})
		} //全部记录到新增结构中
	}
	// 计算加密端口评分
	secureCount := 0
	insecureCount := 0
	nonStandardCount := 0

	for _, v := range securePorts {
		if v {
			secureCount++
		}
	}
	for _, v := range insecurePorts {
		if v {
			insecureCount++
		}
	}
	for _, v := range nonStandardPorts {
		if v {
			nonStandardCount++
		}
	}

	// 评分逻辑
	secureOnly := insecureCount == 0 //&&secureCount!=0? TODO
	secureAndInsecure := secureCount > 0 && insecureCount > 0
	onlyInsecure := secureCount == 0
	hasNonStandard := nonStandardCount > 0

	var encryptionScore int //其实只设定了这四种分数，或许可以划分更细致点
	if secureOnly {
		encryptionScore = 100
	} else if secureAndInsecure {
		encryptionScore = 60
	} else if onlyInsecure {
		encryptionScore = 10
	} else {
		encryptionScore = 0
	}
	var standardScore int
	if hasNonStandard {
		if len(nonStandardPorts) == 1 {
			standardScore = 80
		} else if len(nonStandardPorts) == 2 {
			standardScore = 60
		} else {
			standardScore = 50
		}
	} else {
		standardScore = 100
	}
	scores["encrypted_ports"] = encryptionScore
	scores["standard_ports"] = standardScore
	return scores, portsUsage
}
