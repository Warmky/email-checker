package scoring

import (
	"backend/internal/models"
	"strings"
)

// 评估模块
// 添加防御能力评估模块 6.17
// func evaluateSecurityDefense(uri string, certInfo CertInfo, config string,  dnsRecord *DNSRecord, score map[string]int, connectScores map[string]interface{}) map[string]int {
func EvaluateSecurityDefense(uri string, certInfo models.CertInfo, config string, score map[string]int, connectScores map[string]interface{}) map[string]int {
	defenseScores := make(map[string]int)

	//1.攻击者具有监听能力
	defenseScores["sniffing_defense"] = evaluateEncryption(connectScores) //应该以实际能否明文连接为主

	//2.攻击者中间人篡改
	defenseScores["tampering_defense"] = evaluateTamperingDefense(uri, certInfo) //能否直接通过HTTP连接返回配置信息

	//3.域名接管能力
	defenseScores["domain_takeover_defense"] = evaluateDomaininValidity(connectScores)

	//4.伪造证书能力，证书链是否可信
	defenseScores["fake_cert_defense"] = evaluateCertificateForgery(certInfo)

	//5.DNS劫持
	defenseScores["dns_hijack_defense"] = evaluateDNSHijack_at(connectScores)
	return defenseScores
}

func evaluateEncryption(connectScores map[string]interface{}) int {
	if warnings, ok := connectScores["warnings"].(string); ok {
		if strings.Contains(warnings, "plaintext connections") {
			return 0 //设定的是如果标准加密端口实际上可以明文连接上，则为0分
		}
	}

	if grade, ok := connectScores["Connection_Grade"].(string); ok {
		switch grade {
		case "A+", "A":
			return 100 // 强加密，TLS-only
		case "B":
			return 70 // 部分服务支持加密，但存在明文或 STARTTLS
		case "C":
			return 40 // 弱加密，明文传输多
		case "F":
			return 0 // 明文传输，完全可监听
		default:
			return 50 // 未知评级时折中
		}
	}
	return 0 // 无法评估
}

func evaluateTamperingDefense(uri string, certInfo models.CertInfo) int {
	if strings.HasPrefix(uri, "http://") {
		return 0 // 明文 HTTP 完全无法防篡改
	}

	// HTTPS 的情况，开始细化分析
	score := 100

	if certInfo.IsSelfSigned {
		score -= 30 // 自签名容易伪造
	}
	if !certInfo.IsTrusted {
		score -= 30 // 非可信 CA
	}
	if certInfo.IsExpired {
		score -= 20 // 证书过期
	}
	if !certInfo.IsHostnameMatch {
		score -= 20 // 域名不匹配，容易被伪造证书冒充
	}
	if certInfo.IsInOrder != "yes" {
		score -= 10 // 证书链结构不规范
	}

	if score < 0 {
		score = 0
	}
	return score
}

func evaluateDomaininValidity(connectScores map[string]interface{}) int {
	if warnings, ok := connectScores["warnings"].(string); ok {
		if strings.Contains(warnings, "It may be expired or misconfigured") {
			return 0 // 实际上应该进一步调用aliyun api看是否过期
		}
	}
	return 100
}

func evaluateCertificateForgery(certInfo models.CertInfo) int {
	// 自签名证书最容易被伪造
	if certInfo.IsSelfSigned {
		return 10
	}

	// 不被信任，可能是伪造的根或中间证书
	if !certInfo.IsTrusted {
		return 30
	}

	// 有效期已过，可能被中间人利用旧证书伪造
	if certInfo.IsExpired {
		return 50
	}

	// 主机名不匹配，容易出现泛滥证书欺骗
	if !certInfo.IsHostnameMatch {
		return 70
	}

	// 证书链不规范
	if certInfo.IsInOrder != "yes" {
		return 80
	}

	// 一切正常
	return 100
}

func evaluateDNSHijack_at(connectScores map[string]interface{}) int {
	if warnings, ok := connectScores["warnings"].(string); ok {
		if strings.Contains(warnings, "It may be expired or misconfigured") {
			return 0 // 不能被解析的
		}
	}
	if connectScores["Overall_Connection_Score"] == 0 {
		return 0
	}
	return 100
}

func evaluateDNSHijack(dnsRecord *models.DNSRecord) int {
	if dnsRecord == nil {
		return 20 // 无DNS信息，极易被劫持
	}

	// 检查是否启用 DNSSEC 的 AD bit
	hasDNSSEC := false
	adBits := []*bool{
		dnsRecord.ADbit_imap,
		dnsRecord.ADbit_imaps,
		dnsRecord.ADbit_pop3,
		dnsRecord.ADbit_pop3s,
		dnsRecord.ADbit_smtp,
		dnsRecord.ADbit_smtps,
	}
	for _, bit := range adBits {
		if bit != nil && *bit {
			hasDNSSEC = true
			break
		}
	}

	// 评分依据：
	// - 启用 DNSSEC：+50
	// - SOA、NS 等记录存在：+30
	// - 其他情况：视为弱防护
	score := 0
	if hasDNSSEC {
		score += 50
	}
	if dnsRecord.NS != "" || dnsRecord.SOA != "" {
		score += 30
	}
	if score == 0 {
		return 20 // 没有任何信息的情况最低评分
	}
	if score > 100 {
		return 100
	}
	return score
}

func EvaluateSecurityDefense_SRV(dnsrecord *models.DNSRecord, score map[string]int, connectScores map[string]interface{}) map[string]int {
	defenseScores := make(map[string]int)
	//1.攻击者具有监听能力
	defenseScores["sniffing_defense"] = evaluateEncryption(connectScores) //应该以实际能否明文连接为主

	// //2.攻击者中间人篡改
	// defenseScores["tampering_defense"] = evaluateTamperingDefense(uri, certInfo) //能否直接通过HTTP连接返回配置信息

	//3.域名接管能力
	defenseScores["domain_takeover_defense"] = evaluateDomaininValidity(connectScores)

	//4.伪造证书能力，证书链是否可信
	// defenseScores["fake_cert_defense"] = evaluateCertificateForgery(certInfo)

	//5.DNS劫持
	defenseScores["dns_hijack_defense"] = evaluateDNSHijack(dnsrecord)
	return defenseScores
}
