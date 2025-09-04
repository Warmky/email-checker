// package main

// import (
// 	"bytes"
// 	"crypto/dsa"
// 	"crypto/ecdsa"
// 	"crypto/rsa"
// 	"crypto/tls"
// 	"crypto/x509"
// 	"encoding/base64"
// 	"encoding/csv"
// 	"encoding/json"
// 	"encoding/xml"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"net/url"
// 	"os"
// 	"os/exec"
// 	"sort"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/beevik/etree"
// 	"github.com/miekg/dns"
// 	"github.com/rs/cors"
// 	"github.com/zakjan/cert-chain-resolver/certUtil"
// 	"golang.org/x/net/publicsuffix"
// )

// var (
// 	//msg       = new(dns.Msg)
// 	dnsServer = "8.8.8.8:53"
// 	//client    = new(dns.Client)
// )

// type ProtocolInfo struct {
// 	Type           string `json:"Type"`
// 	Server         string `json:"Server"`
// 	Port           string `json:"Port"`
// 	DomainRequired string `json:"DomainRequired,omitempty"`
// 	SPA            string `json:"SPA,omitempty"`
// 	SSL            string `json:"SSL,omitempty"` //
// 	AuthRequired   string `json:"AuthRequired,omitempty"`
// 	Encryption     string `json:"Encryption,omitempty"`
// 	UsePOPAuth     string `json:"UsePOPAuth,omitempty"`
// 	SMTPLast       string `json:"SMTPLast,omitempty"`
// 	TTL            string `json:"TTL,omitempty"`
// 	SingleCheck    string `json:"SingleCheck"`        //          // Status 用于标记某个Method(Autodiscover/Autoconfig/SRV)的单个Protocol检查结果
// 	Priority       string `json:"Priority,omitempty"` //SRV
// 	Weight         string `json:"Weight,omitempty"`
// }
// type DomainResult struct {
// 	Domain_id     int                  `json:"id"`
// 	Domain        string               `json:"domain"`
// 	CNAME         []string             `json:"cname,omitempty"`
// 	Autodiscover  []AutodiscoverResult `json:"autodiscover"`
// 	Autoconfig    []AutoconfigResult   `json:"autoconfig"`
// 	SRV           SRVResult            `json:"srv"`
// 	Timestamp     string               `json:"timestamp"`
// 	ErrorMessages []string             `json:"errors"`
// }

// type AutoconfigResponse struct {
// 	XMLName xml.Name `xml:"clientConfig"`
// }

// type AutodiscoverResponse struct {
// 	XMLName  xml.Name `xml:"http://schemas.microsoft.com/exchange/autodiscover/responseschema/2006 Autodiscover"`
// 	Response Response `xml:"http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a Response"` //3.13原 是规范的写法，但是有的配置中没有命名空间，导致解析不到Response直接算作成功获取配置信息了
// }

// type Response struct {
// 	User    User    `xml:"User"`
// 	Account Account `xml:"Account"`
// 	Error   *Error  `xml:"Error,omitempty"`
// }

// type User struct {
// 	AutoDiscoverSMTPAddress string `xml:"AutoDiscoverSMTPAddress"`
// 	DisplayName             string `xml:"DisplayName"`
// 	LegacyDN                string `xml:"LegacyDN"`
// 	DeploymentId            string `xml:"DeploymentId"`
// }

// type Account struct {
// 	AccountType     string   `xml:"AccountType"`
// 	Action          string   `xml:"Action"`
// 	MicrosoftOnline string   `xml:"MicrosoftOnline"`
// 	ConsumerMailbox string   `xml:"ConsumerMailbox"`
// 	Protocol        Protocol `xml:"Protocol"`
// 	RedirectAddr    string   `xml:"RedirectAddr"`
// 	RedirectUrl     string   `xml:"RedirectUrl"`
// }

// type Protocol struct{}

// type Error struct {
// 	Time      string `xml:"Time,attr"`
// 	Id        string `xml:"Id,attr"`
// 	DebugData string `xml:"DebugData"`
// 	ErrorCode int    `xml:"ErrorCode"`
// 	Message   string `xml:"Message"`
// }

// type CertInfo struct {
// 	IsTrusted       bool
// 	VerifyError     string
// 	IsHostnameMatch bool
// 	IsInOrder       string
// 	IsExpired       bool
// 	IsSelfSigned    bool
// 	SignatureAlg    string
// 	AlgWarning      string
// 	TLSVersion      uint16
// 	Subject         string
// 	Issuer          string
// 	RawCert         []byte
// }

// // AutodiscoverResult 保存每次Autodiscover查询的结果
// type AutodiscoverResult struct {
// 	Domain            string                   `json:"domain"`
// 	AutodiscoverCNAME []string                 `json:"autodiscovercname,omitempty"`
// 	Method            string                   `json:"method"` // 查询方法，如 POST, GET, SRV
// 	Index             int                      `json:"index"`
// 	URI               string                   `json:"uri"`       // 查询的 URI
// 	Redirects         []map[string]interface{} `json:"redirects"` // 重定向链
// 	Config            string                   `json:"config"`    // 配置信息
// 	CertInfo          *CertInfo                `json:"cert_info"`
// 	Error             string                   `json:"error"` // 错误信息（如果有）
// }

// // AutoconfigResult 保存每次Autoconfig查询的结果
// type AutoconfigResult struct {
// 	Domain    string                   `json:"domain"`
// 	Method    string                   `json:"method"`
// 	Index     int                      `json:"index"`
// 	URI       string                   `json:"uri"`
// 	Redirects []map[string]interface{} `json:"redirects"`
// 	Config    string                   `json:"config"`
// 	CertInfo  *CertInfo                `json:"cert_info"`
// 	Error     string                   `json:"error"`
// }

// type SRVRecord struct {
// 	Service  string
// 	Priority uint16
// 	Weight   uint16
// 	Port     uint16
// 	Target   string
// }

// type DNSRecord struct {
// 	Domain      string `json:"domain"`
// 	SOA         string `json:"SOA,omitempty"`
// 	NS          string `json:"NS,omitempty"`
// 	ADbit_imap  *bool  `json:"ADbit_imap,omitempty"`
// 	ADbit_imaps *bool  `json:"ADbit_imaps,omitempty"`
// 	ADbit_pop3  *bool  `json:"ADbit_pop3,omitempty"`
// 	ADbit_pop3s *bool  `json:"ADbit_pop3s,omitempty"`
// 	ADbit_smtp  *bool  `json:"ADbit_smtp,omitempty"`
// 	ADbit_smtps *bool  `json:"ADbit_smtps,omitempty"`
// }

// type SRVResult struct {
// 	Domain      string      `json:"domain"`
// 	RecvRecords []SRVRecord `json:"recv_records,omitempty"` // 收件服务 (IMAP/POP3)
// 	SendRecords []SRVRecord `json:"send_records,omitempty"` // 发件服务 (SMTP)
// 	DNSRecord   *DNSRecord  `json:"dns_record,omitempty"`
// }

// // 处理 Autodiscover 查询请求
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
// 	//fmt.Fprintf(w, "Processing Autodiscover for: %s", email)
// 	results := queryAutodiscover(domain, email)
// 	for _, result := range results {
// 		if result.Config != "" && !strings.HasPrefix(result.Config, "Errorcode") && !strings.HasPrefix(result.Config, "Non-valid") && !strings.HasPrefix(result.Config, "Bad response") {
// 			//fmt.Fprint(w, result.Config)
// 			// 解析 config 并评分
// 			score, connectScores := scoreConfig(result.Config, *result.CertInfo)

// 			// 构造返回 JSON
// 			response := map[string]interface{}{
// 				"config": result.Config, // 这里也可以选择不返回原始 XML，避免前端解析麻烦
// 				"score":  score,
// 				"score_detail": map[string]interface{}{
// 					"connection": connectScores,
// 				},
// 				"cert_info": result.CertInfo,
// 			}

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

// 	results := queryAutoconfig(domain, email)
// 	for _, result := range results {
// 		if result.Config != "" {
// 			// 如果评分逻辑和 Autodiscover 不一样，你可以另写一个 scoreAutoconfig 函数
// 			score, connectScores := scoreConfig_Autoconfig(result.Config, *result.CertInfo)

// 			response := map[string]interface{}{
// 				"config": result.Config,
// 				"score":  score,
// 				"score_detail": map[string]interface{}{
// 					"connection": connectScores,
// 				},
// 				"cert_info": result.CertInfo,
// 			}

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

// 	result := querySRV(domain)
// 	if len(result.RecvRecords) > 0 || len(result.SendRecords) > 0 {
// 		score, connectScores := scoreConfig_SRV(result)

// 		response := map[string]interface{}{
// 			"score":        score,
// 			"score_detail": map[string]interface{}{"connection": connectScores},
// 			"srv_records": map[string]interface{}{
// 				"recv": result.RecvRecords,
// 				"send": result.SendRecords,
// 			},
// 			"dns_record": result.DNSRecord,
// 		}

// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(response)
// 		return
// 	}

// 	// 返回空记录的结构
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]interface{}{
// 		"message": "No SRV records found",
// 	})
// }

// // func main() {
// // 	http.HandleFunc("/autodiscover", autodiscoverHandler)
// // 	http.HandleFunc("/autoconfig", autoconfigHandler)
// // 	http.HandleFunc("/srv", srvHandler)
// // 	// 启用 CORS
// // 	corsHandler := cors.Default().Handler(http.DefaultServeMux)

// //		log.Println("Server is running on :8081")
// //		log.Fatal(http.ListenAndServe("0.0.0.0:8081", corsHandler))
// //	}
// func main() {
// 	// 正确定位到 frontend/build（相对于 backend 目录）
// 	fs := http.FileServer(http.Dir("../frontend/build"))
// 	http.Handle("/static/", fs)
// 	http.Handle("/favicon.ico", fs)
// 	http.Handle("/manifest.json", fs)

// 	// 支持 React 单页路由：其他路径 fallback 到 index.html
// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		http.ServeFile(w, r, "../frontend/build/index.html")
// 	})

// 	// API 路由
// 	http.HandleFunc("/autodiscover", autodiscoverHandler)
// 	http.HandleFunc("/autoconfig", autoconfigHandler)
// 	http.HandleFunc("/srv", srvHandler)

// 	// 启用 CORS
// 	corsHandler := cors.Default().Handler(http.DefaultServeMux)

// 	log.Println("Server is running on :8081")
// 	log.Fatal(http.ListenAndServe(":8081", corsHandler))
// }

// func queryAutodiscover(domain string, email string) []AutodiscoverResult {
// 	var results []AutodiscoverResult
// 	// //查询autodiscover.example.com的cname记录
// 	// autodiscover_prefixadd := "autodiscover." + domain
// 	// autodiscover_cnameRecords, _ := lookupCNAME(autodiscover_prefixadd)
// 	// method1:直接通过text manipulation，直接发出post请求
// 	uris := []string{
// 		fmt.Sprintf("http://%s/autodiscover/autodiscover.xml", domain),
// 		fmt.Sprintf("https://autodiscover.%s/autodiscover/autodiscover.xml", domain),
// 		fmt.Sprintf("http://autodiscover.%s/autodiscover/autodiscover.xml", domain),
// 		fmt.Sprintf("https://%s/autodiscover/autodiscover.xml", domain),
// 	}
// 	for i, uri := range uris {
// 		index := i + 1
// 		flag1, flag2, flag3, redirects, config, certinfo, err := getAutodiscoverConfig(domain, uri, email, "post", index, 0, 0, 0) //getAutodiscoverConfig照常
// 		fmt.Printf("flag1: %d\n", flag1)
// 		fmt.Printf("flag2: %d\n", flag2)
// 		fmt.Printf("flag3: %d\n", flag3)

// 		result := AutodiscoverResult{
// 			Domain:    domain,
// 			Method:    "POST",
// 			Index:     index,
// 			URI:       uri,
// 			Redirects: redirects,
// 			Config:    config,
// 			CertInfo:  certinfo,
// 		}
// 		if err != nil {
// 			result.Error = err.Error()
// 		}
// 		results = append(results, result)
// 	}

// 	//method2:通过dns找到server,再post请求
// 	service := "_autodiscover._tcp." + domain
// 	uriDNS, _, err := lookupSRVWithAD_autodiscover(domain) //
// 	if err != nil {
// 		result_srv := AutodiscoverResult{
// 			Domain: domain,
// 			Method: "srv-post",
// 			Index:  0,
// 			Error:  fmt.Sprintf("Failed to lookup SRV records for %s: %v", service, err),
// 		}
// 		results = append(results, result_srv)
// 	} else {
// 		//record_ADbit_SRV_autodiscover("autodiscover_record_ad_srv.txt", domain, adBit)
// 		_, _, _, redirects, config, certinfo, err1 := getAutodiscoverConfig(domain, uriDNS, email, "srv-post", 0, 0, 0, 0)
// 		result_srv := AutodiscoverResult{
// 			Domain:    domain,
// 			Method:    "srv-post",
// 			Index:     0,
// 			Redirects: redirects,
// 			Config:    config,
// 			CertInfo:  certinfo,
// 			//AutodiscoverCNAME: autodiscover_cnameRecords,
// 		}
// 		if err1 != nil {
// 			result_srv.Error = err1.Error()
// 		}
// 		results = append(results, result_srv)
// 	}

// 	//method3：先GET找到server，再post请求
// 	getURI := fmt.Sprintf("http://autodiscover.%s/autodiscover/autodiscover.xml", domain) //是通过这个getURI得到server的uri，然后再进行post请求10.26
// 	redirects, config, certinfo, err := GET_AutodiscoverConfig(domain, getURI, email)     //一开始的get请求返回的不是重定向的没有管
// 	result_GET := AutodiscoverResult{
// 		Domain:    domain,
// 		Method:    "get-post",
// 		Index:     0,
// 		URI:       getURI,
// 		Redirects: redirects,
// 		Config:    config,
// 		CertInfo:  certinfo,
// 		//AutodiscoverCNAME: autodiscover_cnameRecords,
// 	}
// 	if err != nil {
// 		result_GET.Error = err.Error()
// 	} //TODO:len(redirect)>0?
// 	results = append(results, result_GET)

// 	//method4:增加几条直接GET请求的路径
// 	direct_getURIs := []string{
// 		fmt.Sprintf("http://%s/autodiscover/autodiscover.xml", domain),               //uri1
// 		fmt.Sprintf("https://autodiscover.%s/autodiscover/autodiscover.xml", domain), //2
// 		fmt.Sprintf("http://autodiscover.%s/autodiscover/autodiscover.xml", domain),  //3
// 		fmt.Sprintf("https://%s/autodiscover/autodiscover.xml", domain),              //4
// 	}
// 	for i, direct_getURI := range direct_getURIs {
// 		index := i + 1
// 		_, _, _, redirects, config, certinfo, err := direct_GET_AutodiscoverConfig(domain, direct_getURI, email, "get", index, 0, 0, 0)
// 		result := AutodiscoverResult{
// 			Domain:    domain,
// 			Method:    "direct_get",
// 			Index:     index,
// 			URI:       direct_getURI,
// 			Redirects: redirects,
// 			Config:    config,
// 			CertInfo:  certinfo,
// 			//AutodiscoverCNAME: autodiscover_cnameRecords,
// 		}
// 		if err != nil {
// 			result.Error = err.Error()
// 		}
// 		results = append(results, result)
// 	}

// 	return results
// }

// func getAutodiscoverConfig(origin_domain string, uri string, email_add string, method string, index int, flag1 int, flag2 int, flag3 int) (int, int, int, []map[string]interface{}, string, *CertInfo, error) {
// 	xmlRequest := fmt.Sprintf(`
// 		<Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/outlook/requestschema/2006">
// 			<Request>
// 				<EMailAddress>%s</EMailAddress>
// 				<AcceptableResponseSchema>http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a</AcceptableResponseSchema>
// 			</Request>
// 		</Autodiscover>`, email_add)

// 	req, err := http.NewRequest("POST", uri, bytes.NewBufferString(xmlRequest))
// 	if err != nil {
// 		fmt.Printf("Error creating request for %s: %v\n", uri, err)
// 		return flag1, flag2, flag3, []map[string]interface{}{}, "", nil, fmt.Errorf("failed to create request: %v", err)
// 	}
// 	req.Header.Set("Content-Type", "text/xml")
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			TLSClientConfig: &tls.Config{
// 				InsecureSkipVerify: true,
// 				MinVersion:         tls.VersionTLS10,
// 			},
// 		},
// 		CheckRedirect: func(req *http.Request, via []*http.Request) error {
// 			return http.ErrUseLastResponse // 禁止重定向
// 		},
// 		Timeout: 15 * time.Second,
// 	}

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		fmt.Printf("Error sending request to %s: %v\n", uri, err)
// 		return flag1, flag2, flag3, []map[string]interface{}{}, "", nil, fmt.Errorf("failed to send request: %v", err)
// 	}

// 	redirects := getRedirects(resp) // 获取当前重定向链
// 	defer resp.Body.Close()         //
// 	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
// 		// 处理重定向
// 		flag1 = flag1 + 1
// 		fmt.Printf("flag1now:%d\n", flag1)
// 		location := resp.Header.Get("Location")
// 		fmt.Printf("Redirect to: %s\n", location)
// 		if location == "" {
// 			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("missing Location header in redirect")
// 		} else if flag1 > 10 { //12.27限制重定向次数
// 			//saveXMLToFile_autodiscover("./location.xml", origin_domain, email_add)
// 			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("too many redirect times")
// 		}

// 		newURI, err := url.Parse(location)
// 		if err != nil {
// 			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("failed to parse redirect URL: %s", location)
// 		}

// 		// 递归调用并合并重定向链
// 		newflag1, newflag2, newflag3, nextRedirects, result, certinfo, err := getAutodiscoverConfig(origin_domain, newURI.String(), email_add, method, index, flag1, flag2, flag3)
// 		//return append(redirects, nextRedirects...), result, err //12.27原
// 		return newflag1, newflag2, newflag3, append(redirects, nextRedirects...), result, certinfo, err
// 	} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
// 		// 处理成功响应
// 		body, err := io.ReadAll(resp.Body)
// 		if err != nil {
// 			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("failed to read response body: %v", err)
// 		}

// 		var autodiscoverResp AutodiscoverResponse
// 		err = xml.Unmarshal(body, &autodiscoverResp)
// 		//这里先记录下unmarshal就不成功的xml
// 		if err != nil {
// 			// if (strings.HasPrefix(strings.TrimSpace(string(body)), `<?xml version="1.0"`) || strings.HasPrefix(strings.TrimSpace(string(body)), `<Autodiscover`)) && !strings.Contains(strings.TrimSpace(string(body)), `<html`) && !strings.Contains(strings.TrimSpace(string(body)), `<item`) && !strings.Contains(strings.TrimSpace(string(body)), `lastmod`) && !strings.Contains(strings.TrimSpace(string(body)), `lt`) {
// 			// 	//if !strings.Contains(strings.TrimSpace(string(body)), `<html`) && !strings.Contains(strings.TrimSpace(string(body)), `<item`) && !strings.Contains(strings.TrimSpace(string(body)), `lastmod`) && !strings.Contains(strings.TrimSpace(string(body)), `lt`) {
// 			// 	//saveno_XMLToFile("no_autodiscover_config.xml", string(body), email_add)
// 			// } //记录错误格式的xml
// 			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("failed to unmarshal XML: %v", err)
// 		}

// 		// 处理 redirectAddr 和 redirectUrl
// 		if autodiscoverResp.Response.Account.Action == "redirectAddr" {
// 			flag2 = flag2 + 1
// 			newEmail := autodiscoverResp.Response.Account.RedirectAddr
// 			//record_filename := filepath.Join("./autodiscover/records", "ReAddr.xml")
// 			//saveXMLToFile_with_ReAdrr_autodiscover(record_filename, string(body), email_add)
// 			if newEmail != "" && flag2 <= 10 {
// 				newflag1, newflag2, newflag3, nextRedirects, result, certinfo, err := getAutodiscoverConfig(origin_domain, uri, newEmail, method, index, flag1, flag2, flag3)
// 				return newflag1, newflag2, newflag3, append(redirects, nextRedirects...), result, certinfo, err
// 			} else if newEmail != "" { //12.27
// 				//saveXMLToFile_autodiscover("./flag2.xml", origin_domain, email_add)
// 				return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("too many RedirectAddr")
// 			} else {
// 				return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("nil ReAddr")
// 			}
// 		} else if autodiscoverResp.Response.Account.Action == "redirectUrl" {
// 			flag3 = flag3 + 1
// 			newUri := autodiscoverResp.Response.Account.RedirectUrl
// 			//record_filename := filepath.Join("./autodiscover/records", "Reurl.xml")
// 			//saveXMLToFile_with_Reuri_autodiscover(record_filename, string(body), email_add)
// 			if newUri != "" && flag3 <= 10 {
// 				newflag1, newflag2, newflag3, nextRedirects, result, certinfo, err := getAutodiscoverConfig(origin_domain, newUri, email_add, method, index, flag1, flag2, flag3)
// 				return newflag1, newflag2, newflag3, append(redirects, nextRedirects...), result, certinfo, err
// 			} else if newUri != "" {
// 				//saveXMLToFile_autodiscover("./flag3.xml", origin_domain, email_add)
// 				return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("too many RedirectUrl")
// 			} else {
// 				return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("nil Reuri")
// 			}
// 		} else if autodiscoverResp.Response.Account.Action == "settings" { //这才是我们需要的
// 			// 记录并返回成功配置(3.13修改，因为会将Response命名空间不合规的也解析到这里)
// 			// outputfile := fmt.Sprintf("./autodiscover/autodiscover_%s_%d_config.xml", method, index)
// 			// saveXMLToFile_autodiscover(outputfile, string(body), email_add)

// 			//只在可以直接返回xml配置的时候记录证书信息
// 			var certInfo CertInfo
// 			// 提取证书信息
// 			if resp.TLS != nil {
// 				var encodedData []byte
// 				goChain := resp.TLS.PeerCertificates
// 				endCert := goChain[0]

// 				// 证书验证
// 				dnsName := resp.Request.URL.Hostname()
// 				var VerifyError error
// 				certInfo.IsTrusted, VerifyError = verifyCertificate(goChain, dnsName)
// 				if VerifyError != nil {
// 					certInfo.VerifyError = VerifyError.Error()
// 				} else {
// 					certInfo.VerifyError = ""
// 				}

// 				certInfo.IsExpired = endCert.NotAfter.Before(time.Now())
// 				certInfo.IsHostnameMatch = verifyHostname(endCert, dnsName)
// 				certInfo.IsSelfSigned = IsSelfSigned(endCert)
// 				certInfo.IsInOrder = isChainInOrder(goChain)
// 				certInfo.TLSVersion = resp.TLS.Version

// 				// 提取证书的其他信息
// 				certInfo.Subject = endCert.Subject.CommonName
// 				certInfo.Issuer = endCert.Issuer.String()
// 				certInfo.SignatureAlg = endCert.SignatureAlgorithm.String()
// 				certInfo.AlgWarning = algWarnings(endCert)

// 				// 将证书编码为 base64 格式
// 				for _, cert := range goChain {
// 					encoded := base64.StdEncoding.EncodeToString(cert.Raw)
// 					encodedData = append(encodedData, []byte(encoded)...)
// 				}
// 				certInfo.RawCert = encodedData
// 			}
// 			return flag1, flag2, flag3, redirects, string(body), &certInfo, nil
// 		} else if autodiscoverResp.Response.Error != nil {
// 			//fmt.Printf("Error: %s\n", string(body))
// 			// 处理错误响应
// 			errorConfig := fmt.Sprintf("Errorcode:%d-%s\n", autodiscoverResp.Response.Error.ErrorCode, autodiscoverResp.Response.Error.Message)
// 			//outputfile := fmt.Sprintf("./autodiscover/autodiscover_%s_%d_Errorconfig.txt", method, index)
// 			//saveXMLToFile_autodiscover(outputfile, errorConfig, email_add)
// 			return flag1, flag2, flag3, redirects, errorConfig, nil, nil
// 		} else {
// 			//fmt.Printf("Response element not valid:%s\n", string(body))
// 			//处理Response可能本身就不正确的响应,同时也会存储不合规的xml(unmarshal的时候合规但Response不合规)
// 			alsoErrorConfig := fmt.Sprintf("Non-valid Response element for %s\n:", email_add)
// 			//outputfile := fmt.Sprintf("./autodiscover/autodiscover_%s_%d_AlsoErrorConfig.xml", method, index)
// 			//saveXMLToFile_autodiscover(outputfile, string(body), email_add)
// 			return flag1, flag2, flag3, redirects, alsoErrorConfig, nil, nil
// 		}
// 	} else {
// 		// 处理非成功响应
// 		//outputfile := fmt.Sprintf("./autodiscover/autodiscover_%s_%d_badresponse.txt", method, index)
// 		badResponse := fmt.Sprintf("Bad response for %s: %d\n", email_add, resp.StatusCode)
// 		//saveXMLToFile_autodiscover(outputfile, badResponse, email_add)
// 		return flag1, flag2, flag3, redirects, badResponse, nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
// 	}
// }
// func GET_AutodiscoverConfig(origin_domain string, uri string, email_add string) ([]map[string]interface{}, string, *CertInfo, error) { //使用先get后post方法
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			TLSClientConfig: &tls.Config{
// 				InsecureSkipVerify: true,
// 				MinVersion:         tls.VersionTLS10,
// 			},
// 		},
// 		CheckRedirect: func(req *http.Request, via []*http.Request) error {
// 			return http.ErrUseLastResponse // 禁止重定向
// 		},
// 		Timeout: 15 * time.Second,
// 	}
// 	resp, err := client.Get(uri)
// 	if err != nil {
// 		return []map[string]interface{}{}, "", nil, fmt.Errorf("failed to send request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	redirects := getRedirects(resp) // 获取当前重定向链

// 	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently { //仅通过get请求获取重定向地址
// 		location := resp.Header.Get("Location")
// 		fmt.Printf("Redirect to: %s\n", location)
// 		if location == "" {
// 			return nil, "", nil, fmt.Errorf("missing Location header in redirect")
// 		}
// 		newURI, err := url.Parse(location)
// 		if err != nil {
// 			return nil, "", nil, fmt.Errorf("failed to parse redirect URL: %s", location)
// 		}

// 		// 递归调用并合并重定向链
// 		_, _, _, nextRedirects, result, certinfo, err := getAutodiscoverConfig(origin_domain, newURI.String(), email_add, "get_post", 0, 0, 0, 0)
// 		return append(redirects, nextRedirects...), result, certinfo, err
// 	} else {
// 		return nil, "", nil, fmt.Errorf("not find Redirect Statuscode")
// 	}
// }

// func direct_GET_AutodiscoverConfig(origin_domain string, uri string, email_add string, method string, index int, flag1 int, flag2 int, flag3 int) (int, int, int, []map[string]interface{}, string, *CertInfo, error) { //一路get请求
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			TLSClientConfig: &tls.Config{
// 				InsecureSkipVerify: true,
// 				MinVersion:         tls.VersionTLS10,
// 			},
// 		},
// 		CheckRedirect: func(req *http.Request, via []*http.Request) error {
// 			return http.ErrUseLastResponse // 禁止重定向
// 		},
// 		Timeout: 15 * time.Second, // 设置请求超时时间
// 	}
// 	resp, err := client.Get(uri)
// 	if err != nil {
// 		return flag1, flag2, flag3, []map[string]interface{}{}, "", nil, fmt.Errorf("failed to send request: %v", err)
// 	}

// 	redirects := getRedirects(resp)
// 	defer resp.Body.Close() //

// 	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
// 		flag1 = flag1 + 1
// 		location := resp.Header.Get("Location")
// 		fmt.Printf("Redirect to: %s\n", location)
// 		if location == "" {
// 			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("missing Location header in redirect")
// 		} else if flag1 > 10 {
// 			//saveXMLToFile_autodiscover("./location2.xml", origin_domain, email_add)
// 			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("too many redirect times")
// 		}

// 		newURI, err := url.Parse(location)
// 		if err != nil {
// 			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("failed to parse redirect URL: %s", location)
// 		}

// 		// 递归调用并合并重定向链
// 		newflag1, newflag2, newflag3, nextRedirects, result, certinfo, err := direct_GET_AutodiscoverConfig(origin_domain, newURI.String(), email_add, method, index, flag1, flag2, flag3)
// 		return newflag1, newflag2, newflag3, append(redirects, nextRedirects...), result, certinfo, err
// 	} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
// 		body, err := io.ReadAll(resp.Body)
// 		if err != nil {
// 			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("failed to read response body: %v", err)
// 		}
// 		var autodiscoverResp AutodiscoverResponse
// 		err = xml.Unmarshal(body, &autodiscoverResp)
// 		if err != nil {
// 			// if (strings.HasPrefix(strings.TrimSpace(string(body)), `<?xml version="1.0"`) || strings.HasPrefix(strings.TrimSpace(string(body)), `<Autodiscover`)) && !strings.Contains(strings.TrimSpace(string(body)), `<html`) && !strings.Contains(strings.TrimSpace(string(body)), `<item`) && !strings.Contains(strings.TrimSpace(string(body)), `lastmod`) && !strings.Contains(strings.TrimSpace(string(body)), `lt`) {
// 			// 	//if !strings.Contains(strings.TrimSpace(string(body)), `<html`) && !strings.Contains(strings.TrimSpace(string(body)), `<item`) && !strings.Contains(strings.TrimSpace(string(body)), `lastmod`) && !strings.Contains(strings.TrimSpace(string(body)), `lt`) {
// 			// 	saveno_XMLToFile("no_autodiscover_config_directget.xml", string(body), email_add)
// 			// } //记录错误格式的xml
// 			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("failed to unmarshal XML: %v", err)
// 		}
// 		if autodiscoverResp.Response.Account.Action == "redirectAddr" {
// 			flag2 = flag2 + 1
// 			newEmail := autodiscoverResp.Response.Account.RedirectAddr
// 			//outputfile := fmt.Sprintf("./autodiscover/autodiscover_%s_%d_redirectAddr_config.xml", method, index)
// 			//saveXMLToFile_autodiscover(outputfile, string(body), email_add)
// 			if newEmail != "" {
// 				return flag1, flag2, flag3, redirects, string(body), nil, nil //TODO, 这里直接返回带redirect_email了
// 			} else {
// 				return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("nil ReAddr")
// 			}
// 		} else if autodiscoverResp.Response.Account.Action == "redirectUrl" {
// 			flag3 = flag3 + 1
// 			newUri := autodiscoverResp.Response.Account.RedirectUrl
// 			//record_filename := filepath.Join("./autodiscover/records", "Reurl_dirGET.xml")
// 			//saveXMLToFile_with_Reuri_autodiscover(record_filename, string(body), email_add) //记录redirecturi,是否会出现继续reUri?
// 			if newUri != "" && flag3 <= 10 {
// 				newflag1, newflag2, newflag3, nextRedirects, result, certinfo, err := direct_GET_AutodiscoverConfig(origin_domain, newUri, email_add, method, index, flag1, flag2, flag3)
// 				return newflag1, newflag2, newflag3, append(redirects, nextRedirects...), result, certinfo, err
// 			} else if newUri != "" {
// 				//saveXMLToFile_autodiscover("./flag32.xml", origin_domain, email_add)
// 				return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("too many RedirectUrl")
// 			} else {
// 				return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("nil Reurl")
// 			}
// 		} else if autodiscoverResp.Response.Account.Action == "settings" {
// 			//outputfile := fmt.Sprintf("./autodiscover/autodiscover_%s_%d_config.xml", method, index)
// 			//saveXMLToFile_autodiscover(outputfile, string(body), email_add)
// 			//只在可以直接返回xml配置的时候记录证书信息
// 			var certInfo CertInfo
// 			// 提取证书信息
// 			if resp.TLS != nil {
// 				var encodedData []byte
// 				goChain := resp.TLS.PeerCertificates
// 				endCert := goChain[0]

// 				// 证书验证
// 				dnsName := resp.Request.URL.Hostname()

// 				var VerifyError error
// 				certInfo.IsTrusted, VerifyError = verifyCertificate(goChain, dnsName)
// 				if VerifyError != nil {
// 					certInfo.VerifyError = VerifyError.Error()
// 				} else {
// 					certInfo.VerifyError = ""
// 				}
// 				certInfo.IsExpired = endCert.NotAfter.Before(time.Now())
// 				certInfo.IsHostnameMatch = verifyHostname(endCert, dnsName)
// 				certInfo.IsSelfSigned = IsSelfSigned(endCert)
// 				certInfo.IsInOrder = isChainInOrder(goChain)
// 				certInfo.TLSVersion = resp.TLS.Version

// 				// 提取证书的其他信息
// 				certInfo.Subject = endCert.Subject.CommonName
// 				certInfo.Issuer = endCert.Issuer.String()
// 				certInfo.SignatureAlg = endCert.SignatureAlgorithm.String()
// 				certInfo.AlgWarning = algWarnings(endCert)

// 				// 将证书编码为 base64 格式
// 				for _, cert := range goChain {
// 					encoded := base64.StdEncoding.EncodeToString(cert.Raw)
// 					encodedData = append(encodedData, []byte(encoded)...)
// 				}
// 				certInfo.RawCert = encodedData
// 			}
// 			return flag1, flag2, flag3, redirects, string(body), &certInfo, nil
// 		} else if autodiscoverResp.Response.Error != nil {
// 			//fmt.Printf("Error: %s\n", string(body))
// 			// 处理错误响应
// 			errorConfig := fmt.Sprintf("Errorcode:%d-%s\n", autodiscoverResp.Response.Error.ErrorCode, autodiscoverResp.Response.Error.Message)
// 			//outputfile := fmt.Sprintf("./autodiscover/autodiscover_%s_%d_Errorconfig.txt", method, index)
// 			//saveXMLToFile_autodiscover(outputfile, errorConfig, email_add)
// 			return flag1, flag2, flag3, redirects, errorConfig, nil, nil
// 		} else {
// 			//fmt.Printf("Response element not valid:%s\n", string(body))
// 			//处理Response可能本身就不正确的响应,同时也会存储不合规的xml(unmarshal的时候合规但Response不合规)
// 			alsoErrorConfig := fmt.Sprintf("Non-valid Response element for %s\n:", email_add)
// 			//outputfile := fmt.Sprintf("./autodiscover/autodiscover_%s_%d_AlsoErrorConfig.xml", method, index)
// 			//saveXMLToFile_autodiscover(outputfile, string(body), email_add)
// 			return flag1, flag2, flag3, redirects, alsoErrorConfig, nil, nil
// 		}
// 	} else {
// 		//outputfile := fmt.Sprintf("./autodiscover/autodiscover_%s_%d_badresponse.txt", method, index)
// 		bad_response := fmt.Sprintf("Bad response for %s:%d\n", email_add, resp.StatusCode)
// 		//saveXMLToFile_autodiscover(outputfile, bad_response, email_add)
// 		return flag1, flag2, flag3, redirects, bad_response, nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode) //同时也想记录请求发送失败时的状态码
// 	}
// }

// // 查询Autoconfig部分
// func queryAutoconfig(domain string, email string) []AutoconfigResult {
// 	var results []AutoconfigResult
// 	//method1 直接通过url发送get请求得到config
// 	urls := []string{
// 		fmt.Sprintf("https://autoconfig.%s/mail/config-v1.1.xml?emailaddress=%s", domain, email),             //uri1
// 		fmt.Sprintf("https://%s/.well-known/autoconfig/mail/config-v1.1.xml?emailaddress=%s", domain, email), //uri2
// 		fmt.Sprintf("http://autoconfig.%s/mail/config-v1.1.xml?emailaddress=%s", domain, email),              //uri3
// 		fmt.Sprintf("http://%s/.well-known/autoconfig/mail/config-v1.1.xml?emailaddress=%s", domain, email),  //uri4
// 	}
// 	for i, url := range urls {
// 		index := i + 1
// 		config, redirects, certinfo, err := Get_autoconfig_config(domain, url, "directurl", index)

// 		result := AutoconfigResult{
// 			Domain:    domain,
// 			Method:    "directurl",
// 			Index:     index,
// 			URI:       url,
// 			Redirects: redirects,
// 			Config:    config,
// 			CertInfo:  certinfo,
// 		}
// 		if err != nil {
// 			result.Error = err.Error()
// 		}
// 		results = append(results, result)
// 	}

// 	//method2 ISPDB
// 	ISPurl := fmt.Sprintf("https://autoconfig.thunderbird.net/v1.1/%s", domain)
// 	config, redirects, certinfo, err := Get_autoconfig_config(domain, ISPurl, "ISPDB", 0)
// 	result_ISPDB := AutoconfigResult{
// 		Domain:    domain,
// 		Method:    "ISPDB",
// 		Index:     0,
// 		URI:       ISPurl,
// 		Redirects: redirects,
// 		Config:    config,
// 		CertInfo:  certinfo,
// 	}
// 	if err != nil {
// 		result_ISPDB.Error = err.Error()
// 	}
// 	results = append(results, result_ISPDB)

// 	//method3 MX查询
// 	mxHost, err := ResolveMXRecord(domain)
// 	if err != nil {
// 		result_MX := AutoconfigResult{
// 			Domain: domain,
// 			Method: "MX",
// 			Index:  0,
// 			Error:  fmt.Sprintf("Resolve MX Record error for %s: %v", domain, err),
// 		}
// 		results = append(results, result_MX)
// 	} else {
// 		mxFullDomain, mxMainDomain, err := extractDomains(mxHost)
// 		if err != nil {
// 			result_MX := AutoconfigResult{
// 				Domain: domain,
// 				Method: "MX",
// 				Index:  0,
// 				Error:  fmt.Sprintf("extract domain from mxHost error for %s: %v", domain, err),
// 			}
// 			results = append(results, result_MX)
// 		} else {
// 			if mxFullDomain == mxMainDomain {
// 				urls := []string{
// 					fmt.Sprintf("https://autoconfig.%s/mail/config-v1.1.xml?emailaddress=%s", mxFullDomain, email), //1
// 					fmt.Sprintf("https://autoconfig.thunderbird.net/v1.1/%s", mxFullDomain),                        //3
// 				}
// 				for i, url := range urls {
// 					config, redirects, certinfo, err := Get_autoconfig_config(domain, url, "MX_samedomain", i*2+1)
// 					result := AutoconfigResult{
// 						Domain:    domain,
// 						Method:    "MX_samedomain",
// 						Index:     i*2 + 1,
// 						URI:       url,
// 						Redirects: redirects,
// 						Config:    config,
// 						CertInfo:  certinfo,
// 					}
// 					if err != nil {
// 						result.Error = err.Error()
// 					}
// 					results = append(results, result)
// 				}
// 			} else {
// 				urls := []string{
// 					fmt.Sprintf("https://autoconfig.%s/mail/config-v1.1.xml?emailaddress=%s", mxFullDomain, email), //1
// 					fmt.Sprintf("https://autoconfig.%s/mail/config-v1.1.xml?emailaddress=%s", mxMainDomain, email), //2
// 					fmt.Sprintf("https://autoconfig.thunderbird.net/v1.1/%s", mxFullDomain),                        //3
// 					fmt.Sprintf("https://autoconfig.thunderbird.net/v1.1/%s", mxMainDomain),                        //4
// 				}
// 				for i, url := range urls {
// 					config, redirects, certinfo, err := Get_autoconfig_config(domain, url, "MX", i+1)
// 					result := AutoconfigResult{
// 						Domain:    domain,
// 						Method:    "MX",
// 						Index:     i + 1,
// 						URI:       url,
// 						Redirects: redirects,
// 						Config:    config,
// 						CertInfo:  certinfo,
// 					}
// 					if err != nil {
// 						result.Error = err.Error()
// 					}
// 					results = append(results, result)
// 				}
// 			}
// 		}

// 	}
// 	return results

// }

// func Get_autoconfig_config(domain string, url string, method string, index int) (string, []map[string]interface{}, *CertInfo, error) {
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			TLSClientConfig: &tls.Config{
// 				InsecureSkipVerify: true,
// 				MinVersion:         tls.VersionTLS10,
// 			},
// 		},
// 		Timeout: 15 * time.Second,
// 	}
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return "", []map[string]interface{}{}, nil, err
// 	}

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return "", []map[string]interface{}{}, nil, err
// 	}
// 	// 获取重定向历史记录
// 	redirects := getRedirects(resp)
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", redirects, nil, fmt.Errorf("failed to read response body: %v", err)
// 	}
// 	var autoconfigResp AutoconfigResponse
// 	err = xml.Unmarshal(body, &autoconfigResp)
// 	if err != nil {
// 		// if (strings.HasPrefix(strings.TrimSpace(string(body)), `<?xml version="1.0"`) || strings.HasPrefix(strings.TrimSpace(string(body)), `<clientConfig`)) && !strings.Contains(strings.TrimSpace(string(body)), `<html`) && !strings.Contains(strings.TrimSpace(string(body)), `<item`) && !strings.Contains(strings.TrimSpace(string(body)), `lastmod`) && !strings.Contains(strings.TrimSpace(string(body)), `lt`) {
// 		// 	saveno_XMLToFile("no_autoconfig_config.xml", string(body), domain)
// 		// }
// 		return "", redirects, nil, fmt.Errorf("failed to unmarshal XML: %v", err)
// 	} else {
// 		var certInfo CertInfo
// 		// 提取证书信息
// 		if resp.TLS != nil {
// 			var encodedData []byte
// 			goChain := resp.TLS.PeerCertificates
// 			endCert := goChain[0]

// 			// 证书验证
// 			dnsName := resp.Request.URL.Hostname()
// 			var VerifyError error
// 			certInfo.IsTrusted, VerifyError = verifyCertificate(goChain, dnsName)
// 			if VerifyError != nil {
// 				certInfo.VerifyError = VerifyError.Error()
// 			} else {
// 				certInfo.VerifyError = ""
// 			}
// 			certInfo.IsExpired = endCert.NotAfter.Before(time.Now())
// 			certInfo.IsHostnameMatch = verifyHostname(endCert, dnsName)
// 			certInfo.IsSelfSigned = IsSelfSigned(endCert)
// 			certInfo.IsInOrder = isChainInOrder(goChain)
// 			certInfo.TLSVersion = resp.TLS.Version

// 			// 提取证书的其他信息
// 			certInfo.Subject = endCert.Subject.CommonName
// 			certInfo.Issuer = endCert.Issuer.String()
// 			certInfo.SignatureAlg = endCert.SignatureAlgorithm.String()
// 			certInfo.AlgWarning = algWarnings(endCert)

// 			// 将证书编码为 base64 格式
// 			for _, cert := range goChain {
// 				encoded := base64.StdEncoding.EncodeToString(cert.Raw)
// 				encodedData = append(encodedData, []byte(encoded)...)
// 			}
// 			certInfo.RawCert = encodedData

// 		}

// 		config := string(body)
// 		// outputfile := fmt.Sprintf("./autoconfig/autoconfig_%s_%d.xml", method, index) //12.18 用Index加以区分
// 		// err = saveXMLToFile_autoconfig(outputfile, config, domain)
// 		// if err != nil {
// 		// 	return "", redirects, &certInfo, err
// 		// }
// 		return config, redirects, &certInfo, nil
// 	}
// }

// // 获取MX记录
// func ResolveMXRecord(domain string) (string, error) {
// 	//创建DNS客户端并设置超时时间
// 	client := &dns.Client{
// 		Timeout: 15 * time.Second, // 设置超时时间
// 	}

// 	// 创建DNS消息
// 	msg := new(dns.Msg)
// 	msg.SetQuestion(dns.Fqdn(domain), dns.TypeMX)
// 	//发送DNS查询
// 	response, _, err := client.Exchange(msg, dnsServer)
// 	if err != nil {
// 		fmt.Printf("Failed to query DNS for %s: %v\n", domain, err)
// 		return "", err
// 	}

// 	//处理响应
// 	if response.Rcode != dns.RcodeSuccess {
// 		fmt.Printf("DNS query failed with Rcode %d\n", response.Rcode)
// 		return "", fmt.Errorf("DNS query failed with Rcode %d", response.Rcode)
// 	}

// 	var mxRecords []*dns.MX
// 	for _, ans := range response.Answer {
// 		if mxRecord, ok := ans.(*dns.MX); ok {
// 			fmt.Printf("MX record for %s: %s, the priority is %d\n", domain, mxRecord.Mx, mxRecord.Preference)
// 			mxRecords = append(mxRecords, mxRecord)
// 		}
// 	}
// 	if len(mxRecords) == 0 {
// 		return "", fmt.Errorf("no MX Record")
// 	}

// 	// 根据Preference字段排序，Preference值越小优先级越高
// 	sort.Slice(mxRecords, func(i, j int) bool {
// 		return mxRecords[i].Preference < mxRecords[j].Preference
// 	})
// 	highestMX := mxRecords[0]
// 	return highestMX.Mx, nil

// }

// // 提取%MXFULLDOMAIN%和%MXMAINDOMAIN%
// func extractDomains(mxHost string) (string, string, error) {
// 	mxHost = strings.TrimSuffix(mxHost, ".")

// 	// 获取%MXFULLDOMAIN%
// 	parts := strings.Split(mxHost, ".")
// 	if len(parts) < 2 {
// 		return "", "", fmt.Errorf("invalid MX Host name: %s", mxHost)
// 	}
// 	mxFullDomain := strings.Join(parts[1:], ".")
// 	fmt.Println("fulldomain:", mxFullDomain)

// 	// 获取%MXMAINDOMAIN%（提取第二级域名）
// 	mxMainDomain, err := publicsuffix.EffectiveTLDPlusOne(mxHost)
// 	if err != nil {
// 		return "", "", fmt.Errorf("cannot extract maindomain: %v", err)
// 	}
// 	fmt.Println("maindomain:", mxMainDomain)

// 	return mxFullDomain, mxMainDomain, nil
// }

// func querySRV(domain string) SRVResult {
// 	var dnsrecord DNSRecord
// 	dnsManager, isSOA, err := queryDNSManager(domain)
// 	if err != nil {
// 		fmt.Printf("Failed to query DNS manager for %s: %v\n", domain, err)
// 	} else {
// 		if isSOA {
// 			dnsrecord = DNSRecord{
// 				Domain: domain,
// 				SOA:    dnsManager,
// 			}
// 		} else {
// 			dnsrecord = DNSRecord{
// 				Domain: domain,
// 				NS:     dnsManager,
// 			}
// 		}
// 	}

// 	// 定义要查询的服务标签
// 	recvServices := []string{
// 		"_imap._tcp." + domain,
// 		"_imaps._tcp." + domain,
// 		"_pop3._tcp." + domain,
// 		"_pop3s._tcp." + domain,
// 	}
// 	sendServices := []string{
// 		"_submission._tcp." + domain,
// 		"_submissions._tcp." + domain,
// 	}

// 	var recvRecords, sendRecords []SRVRecord

// 	// 查询(IMAP/POP3)
// 	for _, service := range recvServices {
// 		records, adBit, err := lookupSRVWithAD_srv(service)
// 		//record_ADbit_SRV(service, "SRV_record_ad_srv.txt", domain, adBit)

// 		if err != nil || len(records) == 0 {
// 			fmt.Printf("Failed to query SRV for %s or no records found: %v\n", service, err)
// 			continue
// 		}

// 		// 更新 DNSRecord 的 AD 位
// 		if strings.HasPrefix(service, "_imaps") {
// 			dnsrecord.ADbit_imaps = &adBit
// 		} else if strings.HasPrefix(service, "_imap") {
// 			dnsrecord.ADbit_imap = &adBit
// 		} else if strings.HasPrefix(service, "_pop3s") {
// 			dnsrecord.ADbit_pop3s = &adBit
// 		} else if strings.HasPrefix(service, "_pop3") {
// 			dnsrecord.ADbit_pop3 = &adBit
// 		}

// 		// 添加 SRV 记录
// 		for _, record := range records {
// 			if record.Target == "." {
// 				continue
// 			}
// 			recvRecords = append(recvRecords, SRVRecord{
// 				Service:  service,
// 				Priority: record.Priority,
// 				Weight:   record.Weight,
// 				Port:     record.Port,
// 				Target:   record.Target,
// 			})
// 		}
// 	}

// 	// 查询 (SMTP)
// 	for _, service := range sendServices {
// 		records, adBit, err := lookupSRVWithAD_srv(service)
// 		//record_ADbit_SRV(service, "SRV_record_ad_srv.txt", domain, adBit)

// 		if err != nil || len(records) == 0 {
// 			fmt.Printf("Failed to query SRV for %s or no records found: %v\n", service, err)
// 			continue
// 		}

// 		// 更新 DNSRecord 的 AD 位
// 		if strings.HasPrefix(service, "_submissions") {
// 			dnsrecord.ADbit_smtps = &adBit
// 		} else if strings.HasPrefix(service, "_submission") {
// 			dnsrecord.ADbit_smtp = &adBit
// 		}

// 		// 添加 SRV 记录
// 		for _, record := range records {
// 			if record.Target == "." {
// 				continue
// 			}
// 			sendRecords = append(sendRecords, SRVRecord{
// 				Service:  service,
// 				Priority: record.Priority,
// 				Weight:   record.Weight,
// 				Port:     record.Port,
// 				Target:   record.Target,
// 			})
// 		}
// 	}

// 	// 对收件服务和发件服务进行排序
// 	sort.Slice(recvRecords, func(i, j int) bool {
// 		if recvRecords[i].Priority == recvRecords[j].Priority {
// 			return recvRecords[i].Weight > recvRecords[j].Weight
// 		}
// 		return recvRecords[i].Priority < recvRecords[j].Priority
// 	})

// 	sort.Slice(sendRecords, func(i, j int) bool {
// 		if sendRecords[i].Priority == sendRecords[j].Priority {
// 			return sendRecords[i].Weight > sendRecords[j].Weight
// 		}
// 		return sendRecords[i].Priority < sendRecords[j].Priority
// 	})

// 	// 返回组合后的结果
// 	return SRVResult{
// 		Domain:      domain,
// 		DNSRecord:   &dnsrecord,
// 		RecvRecords: recvRecords,
// 		SendRecords: sendRecords,
// 	}
// }

// func queryDNSManager(domain string) (string, bool, error) {
// 	resolverAddr := "8.8.8.8:53" // Google Public DNS
// 	timeout := 15 * time.Second  // DNS 查询超时时间

// 	client := &dns.Client{
// 		Net:     "udp",
// 		Timeout: timeout,
// 	}

// 	// 查询 SOA 记录
// 	msg := new(dns.Msg)
// 	msg.SetQuestion(dns.Fqdn(domain), dns.TypeSOA)
// 	response, _, err := client.Exchange(msg, resolverAddr)
// 	if err != nil {
// 		return "", false, fmt.Errorf("SOA query failed: %v", err)
// 	}

// 	// 提取 SOA 记录的管理者信息
// 	for _, ans := range response.Answer {
// 		if soa, ok := ans.(*dns.SOA); ok {
// 			return soa.Ns, true, nil // SOA 记录中的权威 DNS 服务器名称
// 		}
// 	}

// 	// 若 SOA 查询无结果，尝试查询 NS 记录
// 	msg.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
// 	response, _, err = client.Exchange(msg, resolverAddr)
// 	if err != nil {
// 		return "", false, fmt.Errorf("NS query failed: %v", err)
// 	}

// 	var nsRecords []string
// 	for _, ans := range response.Answer {
// 		if ns, ok := ans.(*dns.NS); ok {
// 			nsRecords = append(nsRecords, ns.Ns)
// 		}
// 	}

// 	if len(nsRecords) > 0 {
// 		return strings.Join(nsRecords, ", "), false, nil // 返回 NS 记录列表
// 	}

// 	return "", false, fmt.Errorf("no SOA or NS records found for domain: %s", domain)
// }

// func lookupSRVWithAD_srv(service string) ([]*dns.SRV, bool, error) {
// 	// DNS Resolver configuration
// 	resolverAddr := "8.8.8.8:53" // Google Public DNS
// 	timeout := 15 * time.Second  // Timeout for DNS query

// 	// Create a DNS client
// 	client := &dns.Client{
// 		Net:     "udp", //
// 		Timeout: timeout,
// 	}
// 	// Create the SRV query
// 	msg := new(dns.Msg)
// 	msg.SetQuestion(dns.Fqdn(service), dns.TypeSRV)
// 	msg.RecursionDesired = true // Enable recursion
// 	msg.SetEdns0(4096, true)    // true 表示启用 DO 位，支持 DNSSEC

// 	// Perform the DNS query
// 	response, _, err := client.Exchange(msg, resolverAddr)
// 	if err != nil {
// 		return nil, false, fmt.Errorf("DNS query failed: %v", err)
// 	}

// 	// Check the AD bit in the DNS response flags
// 	adBit := response.AuthenticatedData
// 	// 解析 SRV 记录
// 	var srvRecords []*dns.SRV
// 	for _, ans := range response.Answer {
// 		if srv, ok := ans.(*dns.SRV); ok {
// 			srvRecords = append(srvRecords, srv)
// 		}
// 	}
// 	fmt.Printf("service:%s, adBit:%v\n", service, adBit)
// 	return srvRecords, adBit, nil
// }

// func getRedirects(resp *http.Response) (history []map[string]interface{}) {
// 	for resp != nil {
// 		req := resp.Request
// 		status := resp.StatusCode
// 		entry := map[string]interface{}{
// 			"URL":    req.URL.String(),
// 			"Status": status,
// 		}
// 		history = append(history, entry)
// 		resp = resp.Request.Response
// 	}
// 	if len(history) >= 1 {
// 		for l, r := 0, len(history)-1; l < r; l, r = l+1, r-1 {
// 			history[l], history[r] = history[r], history[l]
// 		}
// 	}
// 	return history
// }

// func lookupSRVWithAD_autodiscover(domain string) (string, bool, error) {
// 	// DNS Resolver configuration
// 	resolverAddr := "8.8.8.8:53" // Google Public DNS
// 	timeout := 5 * time.Second   // Timeout for DNS query

// 	// Create a DNS client
// 	client := &dns.Client{
// 		Net:     "udp", //
// 		Timeout: timeout,
// 	}

// 	// Create the SRV query
// 	service := "_autodiscover._tcp." + domain
// 	msg := new(dns.Msg)
// 	msg.SetQuestion(dns.Fqdn(service), dns.TypeSRV)
// 	msg.RecursionDesired = true // Enable recursion
// 	msg.SetEdns0(4096, true)    // true 表示启用 DO 位，支持 DNSSEC

// 	// Perform the DNS query
// 	response, _, err := client.Exchange(msg, resolverAddr)
// 	if err != nil {
// 		return "", false, fmt.Errorf("DNS query failed: %v", err)
// 	}

// 	// Check the AD bit in the DNS response flags
// 	adBit := response.AuthenticatedData

// 	var srvRecords []*dns.SRV
// 	for _, ans := range response.Answer {
// 		if srv, ok := ans.(*dns.SRV); ok {
// 			srvRecords = append(srvRecords, srv)
// 		}
// 	}
// 	var uriDNS string
// 	if len(srvRecords) > 0 {
// 		sort.Slice(srvRecords, func(i, j int) bool {
// 			if srvRecords[i].Priority == srvRecords[j].Priority {
// 				return srvRecords[i].Weight > srvRecords[j].Weight
// 			}
// 			return srvRecords[i].Priority < srvRecords[j].Priority
// 		})

// 		hostname := srvRecords[0].Target
// 		port := srvRecords[0].Port
// 		if hostname != "." {
// 			if port == 443 {
// 				uriDNS = fmt.Sprintf("https://%s/autodiscover/autodiscover.xml", hostname)
// 			} else if port == 80 {
// 				uriDNS = fmt.Sprintf("http://%s/autodiscover/autodiscover.xml", hostname)
// 			} else {
// 				uriDNS = fmt.Sprintf("https://%s:%d/autodiscover/autodiscover.xml", hostname, port)
// 			}
// 		} else {
// 			return "", adBit, fmt.Errorf("hostname == '.'")
// 		}
// 	} else {
// 		return "", adBit, fmt.Errorf("no srvRecord found")
// 	}

// 	return uriDNS, adBit, nil
// }

// func verifyCertificate(chain []*x509.Certificate, domain string) (bool, error) {
// 	if len(chain) == 1 {
// 		temp_chain, err := certUtil.FetchCertificateChain(chain[0])
// 		if err != nil {
// 			//log.Println("failed to fetch certificate chain")
// 			return false, fmt.Errorf("failed to fetch certificate chain:%v", err)
// 		}
// 		chain = temp_chain
// 	}

// 	intermediates := x509.NewCertPool()
// 	for i := 1; i < len(chain); i++ {
// 		intermediates.AddCert(chain[i])
// 	}

// 	certPool := x509.NewCertPool()
// 	pemFile := "IncludedRootsPEM313.txt" //修改获取roots的途径
// 	pem, err := os.ReadFile(pemFile)
// 	if err != nil {
// 		//log.Println("failed to read root certificate")
// 		return false, fmt.Errorf("failed to read root certificate:%v", err)
// 	}
// 	ok := certPool.AppendCertsFromPEM(pem)
// 	if !ok {
// 		//log.Println("failed to import root certificate")
// 		return false, fmt.Errorf("failed to import root certificate:%v", err)
// 	}

// 	opts := x509.VerifyOptions{
// 		Roots:         certPool,
// 		Intermediates: intermediates,
// 		DNSName:       domain,
// 	}

// 	if _, err := chain[0].Verify(opts); err != nil {
// 		//fmt.Println(err)
// 		return false, fmt.Errorf("certificate verify failed: %v", err)
// 	}

// 	return true, nil
// }

// func verifyHostname(cert *x509.Certificate, domain string) bool {
// 	return cert.VerifyHostname(domain) == nil
// }

// // Ref to: https://github.com/izolight/certigo/blob/v1.10.0/lib/encoder.go#L445
// func IsSelfSigned(cert *x509.Certificate) bool {
// 	if bytes.Equal(cert.RawIssuer, cert.RawSubject) {
// 		return true
// 	} //12.25
// 	return cert.CheckSignatureFrom(cert) == nil
// }

// // Ref to: https://github.com/google/certificate-transparency-go/blob/master/ctutil/sctcheck/sctcheck.go
// func isChainInOrder(chain []*x509.Certificate) string {
// 	// var issuer *x509.Certificate
// 	leaf := chain[0]
// 	for i := 1; i < len(chain); i++ {
// 		c := chain[i]
// 		if bytes.Equal(c.RawSubject, leaf.RawIssuer) && c.CheckSignature(leaf.SignatureAlgorithm, leaf.RawTBSCertificate, leaf.Signature) == nil {
// 			// issuer = c
// 			if i > 1 {
// 				return "not"
// 			}
// 			break
// 		}
// 	}
// 	if len(chain) < 1 {
// 		return "single"
// 	}
// 	return "yes"
// }

// var algoName = [...]string{
// 	x509.MD2WithRSA:      "MD2-RSA",
// 	x509.MD5WithRSA:      "MD5-RSA",
// 	x509.SHA1WithRSA:     "SHA1-RSA",
// 	x509.SHA256WithRSA:   "SHA256-RSA",
// 	x509.SHA384WithRSA:   "SHA384-RSA",
// 	x509.SHA512WithRSA:   "SHA512-RSA",
// 	x509.DSAWithSHA1:     "DSA-SHA1",
// 	x509.DSAWithSHA256:   "DSA-SHA256",
// 	x509.ECDSAWithSHA1:   "ECDSA-SHA1",
// 	x509.ECDSAWithSHA256: "ECDSA-SHA256",
// 	x509.ECDSAWithSHA384: "ECDSA-SHA384",
// 	x509.ECDSAWithSHA512: "ECDSA-SHA512",
// }

// var badSignatureAlgorithms = [...]x509.SignatureAlgorithm{
// 	x509.MD2WithRSA,
// 	x509.MD5WithRSA,
// 	x509.SHA1WithRSA,
// 	x509.DSAWithSHA1,
// 	x509.ECDSAWithSHA1,
// }

// func algWarnings(cert *x509.Certificate) (warning string) {
// 	alg, size := decodeKey(cert.PublicKey)
// 	if (alg == "RSA" || alg == "DSA") && size < 2048 {
// 		// warnings = append(warnings, fmt.Sprintf("Size of %s key should be at least 2048 bits", alg))
// 		warning = fmt.Sprintf("Size of %s key should be at least 2048 bits", alg)
// 	}
// 	if alg == "ECDSA" && size < 224 {
// 		warning = fmt.Sprintf("Size of %s key should be at least 224 bits", alg)
// 	}

// 	for _, alg := range badSignatureAlgorithms {
// 		if cert.SignatureAlgorithm == alg {
// 			warning = fmt.Sprintf("Signed with %s, which is an outdated signature algorithm", algString(alg))
// 		}
// 	}

// 	if alg == "RSA" {
// 		key := cert.PublicKey.(*rsa.PublicKey)
// 		if key.E < 3 {
// 			warning = "Public key exponent in RSA key is less than 3"
// 		}
// 		if key.N.Sign() != 1 {
// 			warning = "Public key modulus in RSA key appears to be zero/negative"
// 		}
// 	}

// 	return warning
// }

// // decodeKey returns the algorithm and key size for a public key.
// func decodeKey(publicKey interface{}) (string, int) {
// 	switch publicKey.(type) {
// 	case *dsa.PublicKey:
// 		return "DSA", publicKey.(*dsa.PublicKey).P.BitLen()
// 	case *ecdsa.PublicKey:
// 		return "ECDSA", publicKey.(*ecdsa.PublicKey).Curve.Params().BitSize
// 	case *rsa.PublicKey:
// 		return "RSA", publicKey.(*rsa.PublicKey).N.BitLen()
// 	default:
// 		return "", 0
// 	}
// }

// func algString(algo x509.SignatureAlgorithm) string {
// 	if 0 < algo && int(algo) < len(algoName) {
// 		return algoName[algo]
// 	}
// 	return strconv.Itoa(int(algo))
// }

// func calculatePortScores(config string) map[string]int {
// 	scores := make(map[string]int)

// 	doc := etree.NewDocument()
// 	if err := doc.ReadFromString(config); err != nil {
// 		scores["error"] = 0
// 		return scores
// 	}
// 	//这里是评分规则
// 	root := doc.SelectElement("Autodiscover")
// 	if root == nil {
// 		scores["error"] = 0
// 		return scores
// 	}
// 	responseElem := root.SelectElement("Response")
// 	if responseElem == nil {
// 		scores["error"] = 0
// 		return scores
// 	}
// 	accountElem := responseElem.SelectElement("Account")
// 	if accountElem == nil {
// 		scores["error"] = 0
// 		return scores
// 	}
// 	accountTypeElem := accountElem.SelectElement("AccountType")
// 	if accountTypeElem == nil || accountTypeElem.Text() != "email" {
// 		scores["error"] = 0
// 		return scores
// 	}
// 	actionElem := accountElem.SelectElement("Action")
// 	if actionElem == nil || actionElem.Text() != "settings" {
// 		scores["error"] = 0
// 		return scores
// 	}
// 	// 记录使用的端口情况
// 	securePorts := map[string]bool{
// 		"SMTP": false,
// 		"IMAP": false,
// 		"POP3": false,
// 	}
// 	insecurePorts := map[string]bool{
// 		"SMTP": false,
// 		"IMAP": false,
// 		"POP3": false,
// 	}
// 	nonStandardPorts := map[string]bool{
// 		"SMTP": false,
// 		"IMAP": false,
// 		"POP3": false,
// 	}
// 	//var protocols []ProtocolInfo
// 	for _, protocolElem := range accountElem.SelectElements("Protocol") {
// 		//protocol := ProtocolInfo{}
// 		protocolType := ""
// 		port := ""
// 		// 检查每个子元素是否存在再获取其内容
// 		if typeElem := protocolElem.SelectElement("Type"); typeElem != nil {
// 			protocolType = typeElem.Text()
// 		}
// 		// if serverElem := protocolElem.SelectElement("Server"); serverElem != nil {
// 		// 	protocol.Server = serverElem.Text()
// 		// }
// 		if portElem := protocolElem.SelectElement("Port"); portElem != nil {
// 			port = portElem.Text()
// 		}
// 		// if domainRequiredElem := protocolElem.SelectElement("DomainRequired"); domainRequiredElem != nil {
// 		// 	protocol.DomainRequired = domainRequiredElem.Text()
// 		// }
// 		// if spaElem := protocolElem.SelectElement("SPA"); spaElem != nil {
// 		// 	protocol.SPA = spaElem.Text()
// 		// }
// 		// if sslElem := protocolElem.SelectElement("SSL"); sslElem != nil {
// 		// 	protocol.SSL = sslElem.Text()
// 		// }
// 		// if authRequiredElem := protocolElem.SelectElement("AuthRequired"); authRequiredElem != nil {
// 		// 	protocol.AuthRequired = authRequiredElem.Text()
// 		// }
// 		// if encryptionElem := protocolElem.SelectElement("Encryption"); encryptionElem != nil {
// 		// 	protocol.Encryption = encryptionElem.Text()
// 		// }
// 		// if usePOPAuthElem := protocolElem.SelectElement("UsePOPAuth"); usePOPAuthElem != nil {
// 		// 	protocol.UsePOPAuth = usePOPAuthElem.Text()
// 		// }
// 		// if smtpLastElem := protocolElem.SelectElement("SMTPLast"); smtpLastElem != nil {
// 		// 	protocol.SMTPLast = smtpLastElem.Text()
// 		// }
// 		// if ttlElem := protocolElem.SelectElement("TTL"); ttlElem != nil {
// 		// 	protocol.TTL = ttlElem.Text()
// 		// }
// 		// if protocol.SSL != "SSL" {
// 		// 	scores["SSL"] = "HHH"
// 		// 	//return scores
// 		// }
// 		// if protocol.Type == "SMTP" && protocol.Port == "465" {
// 		// 	scores["SMTPS"] = "yes"
// 		// }
// 		// if protocol.Type == "IMAP" && protocol.Port == "993" {
// 		// 	scores["IMAPS"] = "yes"
// 		// }
// 		// 分类端口
// 		switch protocolType {
// 		case "SMTP":
// 			if port == "465" {
// 				securePorts["SMTP"] = true
// 			} else if port == "25" || port == "587" {
// 				insecurePorts["SMTP"] = true
// 			} else {
// 				nonStandardPorts["SMTP"] = true
// 			}
// 		case "IMAP":
// 			if port == "993" {
// 				securePorts["IMAP"] = true
// 			} else if port == "143" {
// 				insecurePorts["IMAP"] = true
// 			} else {
// 				nonStandardPorts["IMAP"] = true
// 			}
// 		case "POP3":
// 			if port == "995" {
// 				securePorts["POP3"] = true
// 			} else if port == "110" {
// 				insecurePorts["POP3"] = true
// 			} else {
// 				nonStandardPorts["POP3"] = true
// 			}
// 		}
// 	}
// 	// 计算加密端口评分
// 	secureCount := 0
// 	insecureCount := 0
// 	nonStandardCount := 0

// 	for _, v := range securePorts {
// 		if v {
// 			secureCount++
// 		}
// 	}
// 	for _, v := range insecurePorts {
// 		if v {
// 			insecureCount++
// 		}
// 	}
// 	for _, v := range nonStandardPorts {
// 		if v {
// 			nonStandardCount++
// 		}
// 	}

// 	// 评分逻辑
// 	secureOnly := insecureCount == 0
// 	secureAndInsecure := secureCount > 0 && insecureCount > 0
// 	onlyInsecure := secureCount == 0
// 	hasNonStandard := nonStandardCount > 0

// 	var encryptionScore int
// 	if secureOnly {
// 		encryptionScore = 100
// 	} else if secureAndInsecure {
// 		encryptionScore = 60
// 	} else if onlyInsecure {
// 		encryptionScore = 10
// 	} else {
// 		encryptionScore = 0
// 	}
// 	var standardScore int
// 	if hasNonStandard {
// 		if len(nonStandardPorts) == 1 {
// 			standardScore = 80
// 		} else if len(nonStandardPorts) == 2 {
// 			standardScore = 60
// 		} else {
// 			standardScore = 50
// 		}
// 	} else {
// 		standardScore = 100
// 	}
// 	scores["encrypted_ports"] = encryptionScore
// 	scores["standard_ports"] = standardScore
// 	return scores
// }

// func calculateCertScores(cert CertInfo) map[string]int {
// 	score := 100 // 最高分
// 	scores := make(map[string]int)
// 	// 1. 证书可信度
// 	if !cert.IsTrusted {
// 		score -= 30
// 	}

// 	// 2. 证书主机名匹配
// 	if !cert.IsHostnameMatch {
// 		score -= 20
// 	}

// 	// 3. 证书是否过期
// 	if cert.IsExpired {
// 		score -= 40
// 	}

// 	// 4. 证书是否自签名
// 	if cert.IsSelfSigned {
// 		score -= 30
// 	}

// 	// 5. TLS 版本检查
// 	switch cert.TLSVersion {
// 	case 0x304: // TLS 1.3
// 		score += 10
// 	case 0x303: // TLS 1.2
// 		// 不加分，默认
// 	case 0x302: // TLS 1.1
// 		score -= 40
// 	case 0x301: // TLS 1.0
// 		score -= 60
// 	default: // 低于 TLS 1.0 或未知
// 		score -= 80
// 	}

// 	// 限制最低分为 0
// 	if score < 0 {
// 		score = 0
// 	} else if score > 100 {
// 		score = 100
// 	}
// 	scores["cert"] = score
// 	return scores
// }

// // 生成 CSV 文件，zgrab2 从该文件读取输入
// func writeCSV(hostname string) (string, error) {
// 	// 创建临时文件
// 	tmpFile, err := os.CreateTemp("", "zgrab2_*.csv") // 生成唯一的临时文件
// 	if err != nil {
// 		return "", fmt.Errorf("无法创建临时文件: %v", err)
// 	}
// 	defer tmpFile.Close() // 关闭文件
// 	fmt.Printf("CSV 文件已创建: %s\n", tmpFile.Name())
// 	// 写入 CSV 内容
// 	writer := csv.NewWriter(tmpFile)
// 	defer writer.Flush() // 确保数据写入磁盘

// 	// // zgrab2 需要的 CSV 结构，通常包含 "ip" 或 "domain" 列
// 	// err = writer.Write([]string{"domain"}) // 设置 CSV 头部
// 	// if err != nil {
// 	// 	return "", fmt.Errorf("写入 CSV 头部失败: %v", err)
// 	// }
// 	//会额外读取domain为domain的，所以删去

// 	err = writer.Write([]string{hostname}) // 写入数据
// 	if err != nil {
// 		return "", fmt.Errorf("写入 CSV 数据失败: %v", err)
// 	}

// 	// 返回文件路径
// 	return tmpFile.Name(), nil
// }
// func RunZGrab2(protocoltype, hostname, port string, tlsMode string) (bool, error) {
// 	// 生成临时 CSV 文件
// 	csvfile, err := writeCSV(hostname)
// 	if err != nil {
// 		fmt.Println("CSV 文件未创建")
// 		return false, fmt.Errorf("fail to create csv file: %v", err)
// 	}
// 	defer os.Remove(csvfile) // 开发阶段保留也可以先注释掉这行

// 	// 构造命令参数
// 	args := []string{protocoltype, "--port", port, "-f", csvfile}
// 	if tlsMode == "starttls" {
// 		args = append(args, "--starttls")
// 	} else if tlsMode == "tls" {
// 		// 注意协议后缀变化
// 		tlsFlag := fmt.Sprintf("--%ss", protocoltype)
// 		args = append(args, tlsFlag)
// 	}

// 	// 执行 zgrab2 命令
// 	zgrabPath := "./zgrab2"
// 	cmd := exec.Command(zgrabPath, args...)

// 	var out bytes.Buffer
// 	var stderr bytes.Buffer
// 	cmd.Stdout = &out
// 	cmd.Stderr = &stderr

// 	fmt.Println("Running command:", cmd.String())

// 	err = cmd.Run()
// 	if err != nil {
// 		fmt.Println("Error running command:", err)
// 		fmt.Println("Stderr:", stderr.String())
// 		fmt.Println("Stdout:", out.String())
// 		return false, fmt.Errorf("zgrab2 执行失败: %v\nStderr: %s", err, stderr.String())
// 	}

// 	fmt.Println("Raw JSON Output:", out.String())

// 	var result map[string]interface{}
// 	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
// 		return false, fmt.Errorf("解析 JSON 失败: %v", err)
// 	}

// 	return checkConnectSuccess(result, protocoltype)
// }

// func checkConnectSuccess(result map[string]interface{}, protoType string) (bool, error) {
// 	data, ok := result["data"].(map[string]interface{})
// 	if !ok {
// 		return false, fmt.Errorf("data 字段缺失或类型错误")
// 	}

// 	protoData, ok := data[protoType].(map[string]interface{})
// 	if !ok {
// 		return false, fmt.Errorf("%s 字段缺失或类型错误", protoType)
// 	}

// 	status, ok := protoData["status"].(string)
// 	if !ok {
// 		return false, fmt.Errorf("status 字段缺失或类型错误")
// 	}

// 	if status == "success" {
// 		return true, nil
// 	}
// 	return false, nil
// }

// func calculateConnectScores(config string) map[string]interface{} {
// 	scores := make(map[string]interface{})
// 	//这里为了方便又将config解析了一遍，后面应该和之前的端口评分合并
// 	doc := etree.NewDocument()
// 	if err := doc.ReadFromString(config); err != nil {
// 		scores["error"] = 0
// 		return scores
// 	}
// 	root := doc.SelectElement("Autodiscover")
// 	if root == nil {
// 		scores["error"] = 0
// 		return scores
// 	}
// 	responseElem := root.SelectElement("Response")
// 	if responseElem == nil {
// 		scores["error"] = 0
// 		return scores
// 	}
// 	accountElem := responseElem.SelectElement("Account")
// 	if accountElem == nil {
// 		scores["error"] = 0
// 		return scores
// 	}
// 	accountTypeElem := accountElem.SelectElement("AccountType")
// 	if accountTypeElem == nil || accountTypeElem.Text() != "email" {
// 		scores["error"] = 0
// 		return scores
// 	}
// 	actionElem := accountElem.SelectElement("Action")
// 	if actionElem == nil || actionElem.Text() != "settings" {
// 		scores["error"] = 0
// 		return scores
// 	}

// 	// 遍历每个 Protocol 进行连接测试(三种模式都会尝试)
// 	successTLS := 0
// 	successPlain := 0
// 	successStartTLS := 0
// 	totalProtocols := 0

// 	for _, protocolElem := range accountElem.SelectElements("Protocol") {
// 		//遍历每个protocol来获取连接测试需要的协议类型、端口、主机名
// 		protocolType := ""
// 		port := ""
// 		hostname := ""
// 		if typeElem := protocolElem.SelectElement("Type"); typeElem != nil {
// 			protocolType = strings.ToLower(typeElem.Text())
// 			fmt.Println("ProtocolType:", protocolType)
// 		}
// 		if portElem := protocolElem.SelectElement("Port"); portElem != nil {
// 			port = portElem.Text()
// 			fmt.Println("Port:", port)
// 		}
// 		if serverElem := protocolElem.SelectElement("Server"); serverElem != nil {
// 			hostname = serverElem.Text()
// 			fmt.Println("Hostname:", hostname)
// 		}
// 		// 确保数据完整
// 		if protocolType == "" || port == "" || hostname == "" {
// 			continue
// 		}

// 		totalProtocols++
// 		canConnectPlain, _ := RunZGrab2(protocolType, hostname, port, "plain")
// 		canConnectStartTLS, _ := RunZGrab2(protocolType, hostname, port, "starttls")
// 		canConnectTLS, _ := RunZGrab2(protocolType, hostname, port, "tls")

// 		// 统计连接成功情况
// 		if canConnectTLS {
// 			successTLS++
// 		}
// 		if canConnectStartTLS {
// 			successStartTLS++
// 		}
// 		if canConnectPlain {
// 			successPlain++
// 		}

// 	}
// 	if totalProtocols == 0 {
// 		scores["Connection_Grade"] = "T"
// 		scores["Overall_Connection_Score"] = 0
// 		return scores
// 	}
// 	// 计算评分
// 	tlsScore := (successTLS * 100) / totalProtocols // 100 分制
// 	starttlsScore := (successStartTLS * 100) / totalProtocols
// 	plainScore := (successPlain * 100) / totalProtocols
// 	overall := (tlsScore*3 + starttlsScore*2 + plainScore*1) / 6

// 	//等级判断逻辑
// 	grade := "F"
// 	switch {
// 	case (tlsScore == 100 || starttlsScore == 100) && plainScore == 0:
// 		grade = "A+"
// 	case tlsScore >= 80 || starttlsScore >= 80:
// 		grade = "A"
// 	case tlsScore >= 50 || starttlsScore >= 50:
// 		grade = "B"
// 	case plainScore >= 50:
// 		grade = "C"
// 	case tlsScore == 0 && starttlsScore == 0 && plainScore == 0:
// 		grade = "F"
// 	}
// 	scores["TLS_Connections"] = tlsScore
// 	scores["Plaintext_Connections"] = plainScore
// 	scores["STARTTLS_Connections"] = starttlsScore
// 	scores["Overall_Connection_Score"] = overall
// 	scores["Connection_Grade"] = grade
// 	return scores
// }

// func scoreConfig(config string, certInfo CertInfo) (map[string]int, map[string]interface{}) {
// 	scores := make(map[string]int)
// 	//score_connect_Detail := make(map[string]interface{})
// 	// 计算端口评分
// 	portScores := calculatePortScores(config)
// 	scores["encrypted_ports"] = portScores["encrypted_ports"]
// 	scores["standard_ports"] = portScores["standard_ports"]

// 	// 计算证书评分
// 	certScores := calculateCertScores(certInfo)
// 	scores["cert_score"] = certScores["cert"]

// 	//计算实际连接测试评分
// 	connectScores := calculateConnectScores(config)
// 	// if overall, ok := connectScores["Overall_Connection_Score"].(int); ok {
// 	// 	scores["connect_score"] = overall
// 	// } else {
// 	// 	scores["connect_score"] = 0
// 	// }
// 	// score_connect_Detail["connection"] = connectScores
// 	// 计算最终评分（例如加权平均）
// 	scores["connect_score"] = connectScores["Overall_Connection_Score"].(int)
// 	scores["overall"] = (scores["encrypted_ports"] + scores["standard_ports"] + scores["cert_score"] + scores["connect_score"]) / 4

// 	//return scores, score_connect_Detail
// 	return scores, connectScores
// }

// func scoreConfig_Autoconfig(config string, certInfo CertInfo) (map[string]int, map[string]interface{}) {
// 	scores := make(map[string]int)
// 	//score_connect_Detail := make(map[string]interface{})
// 	// 计算端口评分
// 	portScores := calculatePortScores_Autoconfig(config)
// 	scores["encrypted_ports"] = portScores["encrypted_ports"]
// 	scores["standard_ports"] = portScores["standard_ports"]

// 	// 计算证书评分
// 	certScores := calculateCertScores(certInfo)
// 	scores["cert_score"] = certScores["cert"]

// 	//计算实际连接测试评分
// 	connectScores := calculateConnectScores_Autoconfig(config)
// 	// if overall, ok := connectScores["Overall_Connection_Score"].(int); ok {
// 	// 	scores["connect_score"] = overall
// 	// } else {
// 	// 	scores["connect_score"] = 0
// 	// }
// 	// score_connect_Detail["connection"] = connectScores
// 	// 计算最终评分（例如加权平均）
// 	scores["connect_score"] = connectScores["Overall_Connection_Score"].(int)
// 	scores["overall"] = (scores["encrypted_ports"] + scores["standard_ports"] + scores["cert_score"] + scores["connect_score"]) / 4

// 	//return scores, score_connect_Detail
// 	return scores, connectScores
// }

// func calculatePortScores_Autoconfig(config string) map[string]int {
// 	scores := make(map[string]int)

// 	doc := etree.NewDocument()
// 	if err := doc.ReadFromString(config); err != nil {
// 		scores["error"] = 0
// 		return scores
// 	}
// 	//这里是评分规则
// 	root := doc.SelectElement("clientConfig")
// 	if root == nil {
// 		scores["error"] = 0
// 		return scores
// 	}
// 	emailProviderElem := root.SelectElement("emailProvider")
// 	if emailProviderElem == nil {
// 		scores["error"] = 0
// 		return scores
// 	}

// 	// 记录使用的端口情况
// 	securePorts := map[string]bool{
// 		"SMTP": false,
// 		"IMAP": false,
// 		"POP3": false,
// 	}
// 	insecurePorts := map[string]bool{
// 		"SMTP": false,
// 		"IMAP": false,
// 		"POP3": false,
// 	}
// 	nonStandardPorts := map[string]bool{
// 		"SMTP": false,
// 		"IMAP": false,
// 		"POP3": false,
// 	}
// 	//var protocols []ProtocolInfo
// 	for _, protocolElem := range emailProviderElem.SelectElements("incomingServer") {
// 		//protocol := ProtocolInfo{}
// 		protocolType := ""
// 		port := ""
// 		// 检查每个子元素是否存在再获取其内容
// 		if typeELem := protocolElem.SelectAttr("type"); typeELem != nil {
// 			protocolType = typeELem.Value //? type属性 -> <Type>
// 		}
// 		// if serverElem := protocolElem.SelectElement("hostname"); serverElem != nil {
// 		// 	protocol.Server = serverElem.Text() //<hostname> -> <Server>
// 		// }
// 		if portElem := protocolElem.SelectElement("port"); portElem != nil {
// 			port = portElem.Text()
// 		}

// 		// 分类端口
// 		switch protocolType {
// 		case "SMTP":
// 			if port == "465" {
// 				securePorts["SMTP"] = true
// 			} else if port == "25" || port == "587" {
// 				insecurePorts["SMTP"] = true
// 			} else {
// 				nonStandardPorts["SMTP"] = true
// 			}
// 		case "IMAP":
// 			if port == "993" {
// 				securePorts["IMAP"] = true
// 			} else if port == "143" {
// 				insecurePorts["IMAP"] = true
// 			} else {
// 				nonStandardPorts["IMAP"] = true
// 			}
// 		case "POP3":
// 			if port == "995" {
// 				securePorts["POP3"] = true
// 			} else if port == "110" {
// 				insecurePorts["POP3"] = true
// 			} else {
// 				nonStandardPorts["POP3"] = true
// 			}
// 		}
// 	}

// 	for _, protocolElem := range emailProviderElem.SelectElements("outgoingServer") {
// 		//protocol := ProtocolInfo{}
// 		protocolType := ""
// 		port := ""
// 		// 检查每个子元素是否存在再获取其内容
// 		if typeELem := protocolElem.SelectAttr("type"); typeELem != nil {
// 			protocolType = typeELem.Value //? type属性 -> <Type>
// 		}
// 		// if serverElem := protocolElem.SelectElement("hostname"); serverElem != nil {
// 		// 	protocol.Server = serverElem.Text() //<hostname> -> <Server>
// 		// }
// 		if portElem := protocolElem.SelectElement("port"); portElem != nil {
// 			port = portElem.Text()
// 		}

// 		// 分类端口
// 		switch protocolType {
// 		case "SMTP":
// 			if port == "465" {
// 				securePorts["SMTP"] = true
// 			} else if port == "25" || port == "587" {
// 				insecurePorts["SMTP"] = true
// 			} else {
// 				nonStandardPorts["SMTP"] = true
// 			}
// 		case "IMAP":
// 			if port == "993" {
// 				securePorts["IMAP"] = true
// 			} else if port == "143" {
// 				insecurePorts["IMAP"] = true
// 			} else {
// 				nonStandardPorts["IMAP"] = true
// 			}
// 		case "POP3":
// 			if port == "995" {
// 				securePorts["POP3"] = true
// 			} else if port == "110" {
// 				insecurePorts["POP3"] = true
// 			} else {
// 				nonStandardPorts["POP3"] = true
// 			}
// 		}
// 	}

// 	// 计算加密端口评分
// 	secureCount := 0
// 	insecureCount := 0
// 	nonStandardCount := 0

// 	for _, v := range securePorts {
// 		if v {
// 			secureCount++
// 		}
// 	}
// 	for _, v := range insecurePorts {
// 		if v {
// 			insecureCount++
// 		}
// 	}
// 	for _, v := range nonStandardPorts {
// 		if v {
// 			nonStandardCount++
// 		}
// 	}

// 	// 评分逻辑
// 	secureOnly := insecureCount == 0
// 	secureAndInsecure := secureCount > 0 && insecureCount > 0
// 	onlyInsecure := secureCount == 0
// 	hasNonStandard := nonStandardCount > 0

// 	var encryptionScore int
// 	if secureOnly {
// 		encryptionScore = 100
// 	} else if secureAndInsecure {
// 		encryptionScore = 60
// 	} else if onlyInsecure {
// 		encryptionScore = 10
// 	} else {
// 		encryptionScore = 0
// 	}
// 	var standardScore int
// 	if hasNonStandard {
// 		if len(nonStandardPorts) == 1 {
// 			standardScore = 80
// 		} else if len(nonStandardPorts) == 2 {
// 			standardScore = 60
// 		} else {
// 			standardScore = 50
// 		}
// 	} else {
// 		standardScore = 100
// 	}
// 	scores["encrypted_ports"] = encryptionScore
// 	scores["standard_ports"] = standardScore
// 	return scores
// }

// func calculateConnectScores_Autoconfig(config string) map[string]interface{} {
// 	scores := make(map[string]interface{})
// 	//这里为了方便又将config解析了一遍，后面应该和之前的端口评分合并
// 	doc := etree.NewDocument()
// 	if err := doc.ReadFromString(config); err != nil {
// 		scores["error"] = 0
// 		return scores
// 	}
// 	root := doc.SelectElement("clientConfig")
// 	if root == nil {
// 		scores["error"] = 0
// 		return scores
// 	}
// 	emailProviderElem := root.SelectElement("emailProvider")
// 	if emailProviderElem == nil {
// 		scores["error"] = 0
// 		return scores
// 	}

// 	// 遍历每个 Protocol 进行连接测试(三种模式都会尝试)
// 	successTLS := 0
// 	successPlain := 0
// 	successStartTLS := 0
// 	totalProtocols := 0

// 	for _, protocolElem := range emailProviderElem.SelectElements("incomingServer") {
// 		//遍历每个protocol来获取连接测试需要的协议类型、端口、主机名
// 		protocolType := ""
// 		port := ""
// 		hostname := ""
// 		if typeElem := protocolElem.SelectAttr("type"); typeElem != nil {
// 			protocolType = strings.ToLower(typeElem.Value)
// 			fmt.Println("ProtocolType:", protocolType)
// 		}
// 		if portElem := protocolElem.SelectElement("port"); portElem != nil {
// 			port = portElem.Text()
// 			fmt.Println("Port:", port)
// 		}
// 		if serverElem := protocolElem.SelectElement("hostname"); serverElem != nil {
// 			hostname = serverElem.Text()
// 			fmt.Println("Hostname:", hostname)
// 		}
// 		// 确保数据完整
// 		if protocolType == "" || port == "" || hostname == "" {
// 			continue
// 		}

// 		totalProtocols++
// 		canConnectPlain, _ := RunZGrab2(protocolType, hostname, port, "plain")
// 		canConnectStartTLS, _ := RunZGrab2(protocolType, hostname, port, "starttls")
// 		canConnectTLS, _ := RunZGrab2(protocolType, hostname, port, "tls")

// 		// 统计连接成功情况
// 		if canConnectTLS {
// 			successTLS++
// 		}
// 		if canConnectStartTLS {
// 			successStartTLS++
// 		}
// 		if canConnectPlain {
// 			successPlain++
// 		}

// 	}

// 	for _, protocolElem := range emailProviderElem.SelectElements("outgoingServer") {
// 		//遍历每个protocol来获取连接测试需要的协议类型、端口、主机名
// 		protocolType := ""
// 		port := ""
// 		hostname := ""
// 		if typeElem := protocolElem.SelectAttr("type"); typeElem != nil {
// 			protocolType = strings.ToLower(typeElem.Value)
// 			fmt.Println("ProtocolType:", protocolType)
// 		}
// 		if portElem := protocolElem.SelectElement("port"); portElem != nil {
// 			port = portElem.Text()
// 			fmt.Println("Port:", port)
// 		}
// 		if serverElem := protocolElem.SelectElement("hostname"); serverElem != nil {
// 			hostname = serverElem.Text()
// 			fmt.Println("Hostname:", hostname)
// 		}
// 		// 确保数据完整
// 		if protocolType == "" || port == "" || hostname == "" {
// 			continue
// 		}

// 		totalProtocols++
// 		canConnectPlain, _ := RunZGrab2(protocolType, hostname, port, "plain")
// 		canConnectStartTLS, _ := RunZGrab2(protocolType, hostname, port, "starttls")
// 		canConnectTLS, _ := RunZGrab2(protocolType, hostname, port, "tls")

// 		// 统计连接成功情况
// 		if canConnectTLS {
// 			successTLS++
// 		}
// 		if canConnectStartTLS {
// 			successStartTLS++
// 		}
// 		if canConnectPlain {
// 			successPlain++
// 		}

// 	}
// 	if totalProtocols == 0 {
// 		scores["Connection_Grade"] = "T"
// 		scores["Overall_Connection_Score"] = 0
// 		return scores
// 	}
// 	// 计算评分
// 	tlsScore := (successTLS * 100) / totalProtocols // 100 分制
// 	starttlsScore := (successStartTLS * 100) / totalProtocols
// 	plainScore := (successPlain * 100) / totalProtocols
// 	overall := (tlsScore*3 + starttlsScore*2 + plainScore*1) / 6

// 	//等级判断逻辑
// 	grade := "F"
// 	switch {
// 	case (tlsScore == 100 || starttlsScore == 100) && plainScore == 0:
// 		grade = "A+"
// 	case tlsScore >= 80 || starttlsScore >= 80:
// 		grade = "A"
// 	case tlsScore >= 50 || starttlsScore >= 50:
// 		grade = "B"
// 	case plainScore >= 50:
// 		grade = "C"
// 	case tlsScore == 0 && starttlsScore == 0 && plainScore == 0:
// 		grade = "F"
// 	}
// 	scores["TLS_Connections"] = tlsScore
// 	scores["Plaintext_Connections"] = plainScore
// 	scores["STARTTLS_Connections"] = starttlsScore
// 	scores["Overall_Connection_Score"] = overall
// 	scores["Connection_Grade"] = grade
// 	return scores
// }

// func scoreConfig_SRV(result SRVResult) (map[string]int, map[string]interface{}) {
// 	scores := make(map[string]int)
// 	//score_connect_Detail := make(map[string]interface{})
// 	// 计算端口评分
// 	portScores := calculatePortScores_SRV(result)
// 	scores["encrypted_ports"] = portScores["encrypted_ports"]
// 	scores["standard_ports"] = portScores["standard_ports"]

// 	//计算DNS记录的DNSSEC评分
// 	dnssecScores := calculateDNSSECScores_SRV(result)
// 	scores["dnssec_score"] = dnssecScores

// 	//计算实际连接测试评分
// 	connectScores := calculateConnectScores_SRV(result)

// 	// 计算最终评分（例如加权平均）
// 	scores["connect_score"] = connectScores["Overall_Connection_Score"].(int)
// 	scores["overall"] = (scores["encrypted_ports"] + scores["standard_ports"] + scores["cert_score"] + scores["connect_score"]) / 4

// 	return scores, connectScores
// }

// func calculatePortScores_SRV(result SRVResult) map[string]int {
// 	scores := make(map[string]int)

// 	securePorts := map[string]bool{}
// 	insecurePorts := map[string]bool{}
// 	nonStandardPorts := map[string]bool{}

// 	standardEncrypted := map[uint16]bool{993: true, 995: true, 465: true}
// 	standardInsecure := map[uint16]bool{143: true, 110: true, 25: true, 587: true}

// 	allRecords := append(result.RecvRecords, result.SendRecords...)
// 	for _, record := range allRecords {
// 		port := record.Port
// 		if standardEncrypted[port] {
// 			securePorts[record.Service] = true
// 		} else if standardInsecure[port] {
// 			insecurePorts[record.Service] = true
// 		} else {
// 			nonStandardPorts[record.Service] = true
// 		}
// 	}

// 	secureCount := len(securePorts)
// 	insecureCount := len(insecurePorts)
// 	nonStandardCount := len(nonStandardPorts)

// 	var encryptionScore int
// 	if secureCount > 0 && insecureCount == 0 {
// 		encryptionScore = 100
// 	} else if secureCount > 0 && insecureCount > 0 {
// 		encryptionScore = 60
// 	} else if secureCount == 0 && insecureCount > 0 {
// 		encryptionScore = 10
// 	} else {
// 		encryptionScore = 0
// 	}

// 	var standardScore int
// 	if nonStandardCount == 0 {
// 		standardScore = 100
// 	} else if nonStandardCount == 1 {
// 		standardScore = 80
// 	} else if nonStandardCount == 2 {
// 		standardScore = 60
// 	} else {
// 		standardScore = 50
// 	}

// 	scores["encrypted_ports"] = encryptionScore
// 	scores["standard_ports"] = standardScore
// 	return scores
// }

// func calculateDNSSECScores_SRV(result SRVResult) int {
// 	if result.DNSRecord == nil {
// 		return 0
// 	}
// 	trueCount := 0
// 	total := 0
// 	adBits := []*bool{
// 		result.DNSRecord.ADbit_imap,
// 		result.DNSRecord.ADbit_imaps,
// 		result.DNSRecord.ADbit_pop3,
// 		result.DNSRecord.ADbit_pop3s,
// 		result.DNSRecord.ADbit_smtp,
// 		result.DNSRecord.ADbit_smtps,
// 	}
// 	for _, bit := range adBits {
// 		if bit != nil {
// 			total++
// 			if *bit {
// 				trueCount++
// 			}
// 		}
// 	}
// 	if total == 0 {
// 		return 0
// 	}
// 	return int(float64(trueCount) / float64(total) * 100)
// }

// func calculateConnectScores_SRV(result SRVResult) map[string]interface{} {
// 	scores := make(map[string]interface{})

// 	successTLS := 0
// 	successPlain := 0
// 	successStartTLS := 0
// 	totalProtocols := 0

// 	testRecords := append(result.RecvRecords, result.SendRecords...)
// 	for _, record := range testRecords {
// 		protocol := detectProtocolFromService(record.Service)
// 		if protocol == "" || record.Port == 0 || record.Target == "" {
// 			continue
// 		}

// 		portStr := fmt.Sprintf("%d", record.Port)
// 		hostname := record.Target

// 		totalProtocols++

// 		canConnectPlain, _ := RunZGrab2(protocol, hostname, portStr, "plain")
// 		canConnectStartTLS, _ := RunZGrab2(protocol, hostname, portStr, "starttls")
// 		canConnectTLS, _ := RunZGrab2(protocol, hostname, portStr, "tls")

// 		if canConnectTLS {
// 			successTLS++
// 		}
// 		if canConnectStartTLS {
// 			successStartTLS++
// 		}
// 		if canConnectPlain {
// 			successPlain++
// 		}
// 	}

// 	if totalProtocols == 0 {
// 		scores["Connection_Grade"] = "T"
// 		scores["Overall_Connection_Score"] = 0
// 		return scores
// 	}

// 	tlsScore := (successTLS * 100) / totalProtocols
// 	starttlsScore := (successStartTLS * 100) / totalProtocols
// 	plainScore := (successPlain * 100) / totalProtocols
// 	overall := (tlsScore*3 + starttlsScore*2 + plainScore*1) / 6

// 	grade := "F"
// 	switch {
// 	case (tlsScore == 100 || starttlsScore == 100) && plainScore == 0:
// 		grade = "A+"
// 	case tlsScore >= 80 || starttlsScore >= 80:
// 		grade = "A"
// 	case tlsScore >= 50 || starttlsScore >= 50:
// 		grade = "B"
// 	case plainScore >= 50:
// 		grade = "C"
// 	case tlsScore == 0 && starttlsScore == 0 && plainScore == 0:
// 		grade = "F"
// 	}

// 	scores["TLS_Connections"] = tlsScore
// 	scores["Plaintext_Connections"] = plainScore
// 	scores["STARTTLS_Connections"] = starttlsScore
// 	scores["Overall_Connection_Score"] = overall
// 	scores["Connection_Grade"] = grade
// 	return scores
// }
// func detectProtocolFromService(service string) string {
// 	// 将常见 SRV 服务名转换为 zgrab2 的协议名
// 	service = strings.ToLower(service)
// 	switch {
// 	case strings.Contains(service, "imap"):
// 		return "imap"
// 	case strings.Contains(service, "pop3"):
// 		return "pop3"
// 	case strings.Contains(service, "submission"):
// 		return "smtp"
// 	}
// 	return ""
// }

package main

import (
	"bufio"
	"bytes"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/beevik/etree"
	"github.com/gorilla/websocket"
	"github.com/miekg/dns"
	"github.com/rs/cors"
	"github.com/zakjan/cert-chain-resolver/certUtil"
	"golang.org/x/net/publicsuffix"
)

var (
	//msg       = new(dns.Msg)
	dnsServer = "8.8.8.8:53"
	//client    = new(dns.Client)
)

type AuthInfo struct {
	EightBitMIME        string `json:"8bitmime"`
	Pipelining          string `json:"pipelining"`
	Size                string `json:"size"`
	StartTLS            string `json:"starttls"`
	Auth                string `json:"auth"`
	DSN                 string `json:"dsn"`
	EnhancedStatusCodes string `json:"enhancedstatuscodes"`
}
type TLSInfo struct {
	Error   []string      `json:"error"`
	Version string        `json:"version"`
	Cipher  []interface{} `json:"cipher"`
	TLSCA   string        `json:"tls ca"`
	//Auth    AuthInfo      `json:"auth"`
}

type ConnectInfo struct {
	Success bool     `json:"success"`
	Info    *TLSInfo `json:"info"` // 注意：需要是指针，才能兼容 null
	Error   string   `json:"error,omitempty"`
}

type ProtocolInfo struct {
	Type           string `json:"Type"`
	Server         string `json:"Server"`
	Port           string `json:"Port"`
	DomainRequired string `json:"DomainRequired,omitempty"`
	SPA            string `json:"SPA,omitempty"`
	SSL            string `json:"SSL,omitempty"` //
	AuthRequired   string `json:"AuthRequired,omitempty"`
	Encryption     string `json:"Encryption,omitempty"`
	UsePOPAuth     string `json:"UsePOPAuth,omitempty"`
	SMTPLast       string `json:"SMTPLast,omitempty"`
	TTL            string `json:"TTL,omitempty"`
	SingleCheck    string `json:"SingleCheck"`        //          // Status 用于标记某个Method(Autodiscover/Autoconfig/SRV)的单个Protocol检查结果
	Priority       string `json:"Priority,omitempty"` //SRV
	Weight         string `json:"Weight,omitempty"`
}
type DomainResult struct {
	Domain_id     int                  `json:"id"`
	Domain        string               `json:"domain"`
	CNAME         []string             `json:"cname,omitempty"`
	Autodiscover  []AutodiscoverResult `json:"autodiscover"`
	Autoconfig    []AutoconfigResult   `json:"autoconfig"`
	SRV           SRVResult            `json:"srv"`
	Timestamp     string               `json:"timestamp"`
	ErrorMessages []string             `json:"errors"`
}

type AutoconfigResponse struct {
	XMLName xml.Name `xml:"clientConfig"`
}

type AutodiscoverResponse struct {
	XMLName  xml.Name `xml:"http://schemas.microsoft.com/exchange/autodiscover/responseschema/2006 Autodiscover"`
	Response Response `xml:"http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a Response"` //3.13原 是规范的写法，但是有的配置中没有命名空间，导致解析不到Response直接算作成功获取配置信息了
}

type Response struct {
	User    User    `xml:"User"`
	Account Account `xml:"Account"`
	Error   *Error  `xml:"Error,omitempty"`
}

type User struct {
	AutoDiscoverSMTPAddress string `xml:"AutoDiscoverSMTPAddress"`
	DisplayName             string `xml:"DisplayName"`
	LegacyDN                string `xml:"LegacyDN"`
	DeploymentId            string `xml:"DeploymentId"`
}

type Account struct {
	AccountType     string   `xml:"AccountType"`
	Action          string   `xml:"Action"`
	MicrosoftOnline string   `xml:"MicrosoftOnline"`
	ConsumerMailbox string   `xml:"ConsumerMailbox"`
	Protocol        Protocol `xml:"Protocol"`
	RedirectAddr    string   `xml:"RedirectAddr"`
	RedirectUrl     string   `xml:"RedirectUrl"`
}

type Protocol struct{}

type Error struct {
	Time      string `xml:"Time,attr"`
	Id        string `xml:"Id,attr"`
	DebugData string `xml:"DebugData"`
	ErrorCode int    `xml:"ErrorCode"`
	Message   string `xml:"Message"`
}

type CertInfo struct {
	IsTrusted       bool
	VerifyError     string
	IsHostnameMatch bool
	IsInOrder       string
	IsExpired       bool
	IsSelfSigned    bool
	SignatureAlg    string
	AlgWarning      string
	TLSVersion      uint16
	Subject         string
	Issuer          string
	RawCert         []byte
	RawCerts        []string //8.15
}

// AutodiscoverResult 保存每次Autodiscover查询的结果
type AutodiscoverResult struct {
	Domain            string                   `json:"domain"`
	AutodiscoverCNAME []string                 `json:"autodiscovercname,omitempty"`
	Method            string                   `json:"method"` // 查询方法，如 POST, GET, SRV
	Index             int                      `json:"index"`
	URI               string                   `json:"uri"`       // 查询的 URI
	Redirects         []map[string]interface{} `json:"redirects"` // 重定向链
	Config            string                   `json:"config"`    // 配置信息
	CertInfo          *CertInfo                `json:"cert_info"`
	Error             string                   `json:"error"` // 错误信息（如果有）
}

// AutoconfigResult 保存每次Autoconfig查询的结果
type AutoconfigResult struct {
	Domain    string                   `json:"domain"`
	Method    string                   `json:"method"`
	Index     int                      `json:"index"`
	URI       string                   `json:"uri"`
	Redirects []map[string]interface{} `json:"redirects"`
	Config    string                   `json:"config"`
	CertInfo  *CertInfo                `json:"cert_info"`
	Error     string                   `json:"error"`
}

type SRVRecord struct {
	Service  string
	Priority uint16
	Weight   uint16
	Port     uint16
	Target   string
}

type DNSRecord struct {
	Domain      string `json:"domain"`
	SOA         string `json:"SOA,omitempty"`
	NS          string `json:"NS,omitempty"`
	ADbit_imap  *bool  `json:"ADbit_imap,omitempty"`
	ADbit_imaps *bool  `json:"ADbit_imaps,omitempty"`
	ADbit_pop3  *bool  `json:"ADbit_pop3,omitempty"`
	ADbit_pop3s *bool  `json:"ADbit_pop3s,omitempty"`
	ADbit_smtp  *bool  `json:"ADbit_smtp,omitempty"`
	ADbit_smtps *bool  `json:"ADbit_smtps,omitempty"`
}

type SRVResult struct {
	Domain      string      `json:"domain"`
	RecvRecords []SRVRecord `json:"recv_records,omitempty"` // 收件服务 (IMAP/POP3)
	SendRecords []SRVRecord `json:"send_records,omitempty"` // 发件服务 (SMTP)
	DNSRecord   *DNSRecord  `json:"dns_record,omitempty"`
}

// 8.10
type GuessResult struct {
	Host     string   `json:"host"`
	Port     int      `json:"port"`
	Protocol string   `json:"protocol"`
	IPs      []string `json:"ips"`
	Reach    bool     `json:"reach"`
}

// 尝试在界面展示Recently Seen 5.19
type ScanHistory struct {
	Domain    string    `json:"domain"`
	Timestamp time.Time `json:"timestamp"`
	Score     int       `json:"score"` // 新增：总分
	Grade     string    `json:"grade"` // 新增：等级，如 A/B/C/F
}

// 尝试保留原配置中的数据结构以供推荐时使用
type PortUsageDetail struct {
	Protocol string `json:"protocol"` // SMTP / IMAP / POP3
	Port     string `json:"port"`
	Status   string `json:"status"` // "secure" / "insecure" / "nonstandard"
	Host     string `json:"host"`   //7.27
	SSL      string `json:"ssl"`    //7.27
}

//8.12
// type ConnectInfo struct {
// 	Success bool   `json:"success"`
// 	Version string `json:"version,omitempty"`
// 	Cipher  string `json:"cipher,omitempty"`
// 	TLSCA   string `json:"tls_ca,omitempty"`
// 	Error   string `json:"error,omitempty"`
// }

type ConnectDetail struct {
	Type     string      `json:"type"` // imap / smtp / pop3
	Host     string      `json:"host"`
	Port     string      `json:"port"`
	Plain    ConnectInfo `json:"plain"`
	StartTLS ConnectInfo `json:"starttls"`
	TLS      ConnectInfo `json:"tls"`
}

var recentScans []ScanHistory

const maxRecent = 20

var semaphore = make(chan struct{}, 10) // 控制并发数8.18

// 8.15TODO
var tempDataStore = make(map[string]interface{})

// 8.20
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// 8.23
var progressClients = make(map[*websocket.Conn]bool)
var progressBroadcast = make(chan ProgressUpdate)

type ProgressUpdate struct {
	Type     string `json:"type"`     // 固定 "progress"
	Progress int    `json:"progress"` // 0 ~ 100
	Stage    string `json:"stage"`    // autodiscover / autoconfig / srv / guess
	Message  string `json:"message"`  // 说明文字
}

// ✅ 启动广播协程
func startProgressBroadcaster() {
	go func() {
		for update := range progressBroadcast {
			for conn := range progressClients {
				payload := map[string]interface{}{
					"type":     "progress",
					"progress": update.Progress,
					"stage":    update.Stage,
					"message":  update.Message,
				}
				if err := conn.WriteJSON(payload); err != nil {
					log.Println("WebSocket write error:", err)
					conn.Close()
					delete(progressClients, conn)
				}
			}
		}
	}()
}

// ✅ WebSocket Handler
func WSCheckAllProgressHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to websocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	progressClients[conn] = true
	log.Println("客户端已连接进度 WebSocket")
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket 关闭:", err)
			delete(progressClients, conn)
			break
		}
	}
}

// 接收完整结构
func storeTempDataHandler(w http.ResponseWriter, r *http.Request) {
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
	tempDataStore[id] = payload

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

// 获取完整结构
func getTempDataHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}

	data, ok := tempDataStore[id]
	if !ok {
		http.Error(w, "Data not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func AddRecentScanWithScore(domain string, score int, grade string) {
	filtered := []ScanHistory{}
	for _, item := range recentScans {
		if item.Domain != domain {
			filtered = append(filtered, item)
		}
	}
	recentScans = append([]ScanHistory{{
		Domain:    domain,
		Timestamp: time.Now(),
		Score:     score,
		Grade:     grade,
	}}, filtered...)

	if len(recentScans) > maxRecent {
		recentScans = recentScans[:maxRecent]
	}
}

func GetRecentScans() []ScanHistory {
	return recentScans
}

func handleRecentScans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if recentScans == nil {
		json.NewEncoder(w).Encode([]ScanHistory{})
	} else {
		json.NewEncoder(w).Encode(recentScans)
	}
}

// 处理 Autodiscover 查询请求 4.22
func autodiscoverHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email parameter is required", http.StatusBadRequest)
		return
	}
	// TODO: 这里调用 Autodiscover 查询逻辑
	//首先由用户输入的邮件用户名得到domain
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}
	domain := parts[1]
	// atIndex := strings.LastIndex(email, "@")
	// if atIndex == -1 {
	// 	http.Error(w, "Invalid email format", http.StatusBadRequest)
	// 	return
	// }
	// domain := email[atIndex+1:] //这样更能保证即使用户名里出现了@也可以解析虽然实际上不行

	//fmt.Fprintf(w, "Processing Autodiscover for: %s", email)
	results := queryAutodiscover(domain, email)
	for _, result := range results {
		if result.Config != "" && !strings.HasPrefix(result.Config, "Errorcode") && !strings.HasPrefix(result.Config, "Non-valid") && !strings.HasPrefix(result.Config, "Bad response") { //AUtodiscover机制中找到一个有效的配置就停止了
			//fmt.Fprint(w, result.Config)
			// 解析 config 并评分
			score, connectScores, _, _ := scoreConfig(result.Config, *result.CertInfo)
			securityDefense := evaluateSecurityDefense(result.URI, *result.CertInfo, result.Config, score, connectScores)
			// 构造返回 JSON
			response := map[string]interface{}{
				"config": result.Config, // 这里也可以选择不返回原始 XML，避免前端解析麻烦
				"score":  score,
				"score_detail": map[string]interface{}{
					"connection": connectScores,
					"defense":    securityDefense,
				},
				"cert_info": result.CertInfo,
			}
			// 添加到最近记录
			AddRecentScanWithScore(domain, score["overall"], connectScores["Connection_Grade"].(string)) //5.19
			// 返回 JSON
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}
	// 如果没有有效的结果，返回错误信息
	http.Error(w, "No valid Autodiscover configuration found", http.StatusNotFound)
	//fmt.Fprintf(w, "Processing Autodiscover for: %s", email)
}

// 处理 Autoconfig 查询请求
func autoconfigHandler(w http.ResponseWriter, r *http.Request) {
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

	results := queryAutoconfig(domain, email)
	for _, result := range results {
		if result.Config != "" {
			// 如果评分逻辑和 Autodiscover 不一样,可以另写一个 scoreAutoconfig 函数
			score, connectScores, _, _ := scoreConfig_Autoconfig(result.Config, *result.CertInfo)

			response := map[string]interface{}{
				"config": result.Config,
				"score":  score,
				"score_detail": map[string]interface{}{
					"connection": connectScores,
				},
				"cert_info": result.CertInfo,
			}
			// 添加到最近记录
			AddRecentScanWithScore(domain, score["overall"], connectScores["Connection_Grade"].(string)) //5.19
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}
	http.Error(w, "No valid Autoconfig configuration found", http.StatusNotFound)
}

func srvHandler(w http.ResponseWriter, r *http.Request) {
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

	result := querySRV(domain)
	if len(result.RecvRecords) > 0 || len(result.SendRecords) > 0 {
		score, connectScores, _, _ := scoreConfig_SRV(result)

		response := map[string]interface{}{
			"score":        score,
			"score_detail": map[string]interface{}{"connection": connectScores},
			"srv_records": map[string]interface{}{
				"recv": result.RecvRecords,
				"send": result.SendRecords,
			},
			"dns_record": result.DNSRecord,
		}
		// 添加到最近记录
		AddRecentScanWithScore(domain, score["overall"], connectScores["Connection_Grade"].(string)) //5.19
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// // 返回空记录的结构
	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(map[string]interface{}{
	// 	"message": "No SRV records found",
	// })
	http.Error(w, "No valid SRV configuration found", http.StatusNotFound)
}

func checkAllHandler(w http.ResponseWriter, r *http.Request) { //5.19新增以使得用户无需手动选择机制，三种都查询一遍
	progressBroadcast <- ProgressUpdate{Progress: 0, Stage: "start", Message: "开始检测"}
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
	progressBroadcast <- ProgressUpdate{Progress: 10, Stage: "autodiscover", Message: "开始 Autodiscover 检测"}
	adResults := queryAutodiscover(domain, email)
	var validResults []map[string]interface{} //7.22
	for _, result := range adResults {
		if result.Config != "" && !strings.HasPrefix(result.Config, "Errorcode") &&
			!strings.HasPrefix(result.Config, "Non-valid") &&
			!strings.HasPrefix(result.Config, "Bad response") {
			score, connectScores, ConnectDetails, PortsUsage := scoreConfig(result.Config, *result.CertInfo)
			securityDefense := evaluateSecurityDefense(result.URI, *result.CertInfo, result.Config, score, connectScores)
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
			})
			// autodiscoverResp = map[string]interface{}{
			// 	"config": result.Config,
			// 	"score":  score,
			// 	"score_detail": map[string]interface{}{
			// 		"connection":  connectScores,
			// 		"defense":     securityDefense,
			// 		"ports_usage": PortsUsage, //
			// 	},
			// 	"cert_info": result.CertInfo,
			// }

			// bestScore = score["overall"]
			// bestGrade = connectScores["Connection_Grade"].(string)

			//break
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
	progressBroadcast <- ProgressUpdate{Progress: 30, Stage: "autodiscover", Message: "Autodiscover 检测完成"}
	// AUTOCONFIG 查询
	//var autoconfigResp map[string]interface{}
	progressBroadcast <- ProgressUpdate{Progress: 40, Stage: "autoconfig", Message: "开始 Autoconfig 检测"}
	acResults := queryAutoconfig(domain, email)
	var validacResults []map[string]interface{}
	for _, result := range acResults {
		if result.Config != "" {
			score, connectScores, ConnectDetails, PortsUsage := scoreConfig_Autoconfig(result.Config, *result.CertInfo)
			securityDefense := evaluateSecurityDefense(result.URI, *result.CertInfo, result.Config, score, connectScores)
			// autoconfigResp = map[string]interface{}{
			// 	"config": result.Config,
			// 	"score":  score,
			// 	"score_detail": map[string]interface{}{
			// 		"connection":  connectScores,
			// 		"defense":     securityDefense,
			// 		"ports_usage": PortsUsage,
			// 	},
			// 	"cert_info": result.CertInfo,
			// }
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
			})

			// if score["overall"] > bestScore {
			// 	bestScore = score["overall"]
			// 	bestGrade = connectScores["Connection_Grade"].(string)
			// }
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
	progressBroadcast <- ProgressUpdate{Progress: 60, Stage: "autoconfig", Message: "Autoconfig 检测完成"}
	// SRV 查询
	progressBroadcast <- ProgressUpdate{Progress: 70, Stage: "srv", Message: "开始 SRV 记录检测"}
	srvResult := querySRV(domain)
	var srvResp map[string]interface{}
	// var srvScore map[string]int //
	// var srvConnectScores map[string]interface{} //

	if len(srvResult.RecvRecords) > 0 || len(srvResult.SendRecords) > 0 {
		srvScore, srvConnectScores, ConnectDetails, srvPortsUsage := scoreConfig_SRV(srvResult)
		securityDefense := evaluateSecurityDefense_SRV(srvResult.DNSRecord, srvScore, srvConnectScores)
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
	progressBroadcast <- ProgressUpdate{Progress: 85, Stage: "srv", Message: "SRV 检测完成"}
	//8.10本地添加GUESS部分
	progressBroadcast <- ProgressUpdate{Progress: 90, Stage: "guess", Message: "尝试猜测邮件服务器"}
	var guessResp map[string]interface{}
	guessed := guessMailServer(domain, 2*time.Second, 20)
	if len(guessed) > 0 {
		fmt.Print(guessed)
		connectScores, ConnectDetails, PortsUsage := scoreConfig_Guess(guessed)
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
	progressBroadcast <- ProgressUpdate{Progress: 98, Stage: "guess", Message: "猜测完成"}

	// 添加到最近记录
	AddRecentScanWithScore(domain, bestScore, bestGrade) //5.19
	progressBroadcast <- ProgressUpdate{Progress: 100, Stage: "done", Message: "检测完成"}
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

func main() {
	// http.HandleFunc("/autodiscover", autodiscoverHandler)
	// http.HandleFunc("/autoconfig", autoconfigHandler)
	// http.HandleFunc("/srv", srvHandler)
	startProgressBroadcaster() // ✅ 启动进度推送协程

	http.HandleFunc("/checkAll", checkAllHandler)     // ✅ 新增
	http.HandleFunc("/api/recent", handleRecentScans) // ✅ 新增
	// 新增临时数据接口
	http.HandleFunc("/store-temp-data", storeTempDataHandler)
	http.HandleFunc("/get-temp-data", getTempDataHandler)

	http.HandleFunc("/api/uploadCsvAndExportJsonl", UploadCsvAndExportJsonlHandler)

	http.HandleFunc("/ws/testconnect", WSConnectHandler)
	http.HandleFunc("/ws/checkall-progress", WSCheckAllProgressHandler)

	// 静态文件托管
	http.Handle("/downloads/", http.StripPrefix("/downloads/", http.FileServer(http.Dir("downloads"))))

	// 启用 CORS
	corsHandler := cors.Default().Handler(http.DefaultServeMux)

	log.Println("Server is running on :8081")
	log.Fatal(http.ListenAndServe("0.0.0.0:8081", corsHandler))
}

func WSConnectHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
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
		result = &ConnectInfo{
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

// func LiveTestConnection(host string, port int, protocol string, conn *websocket.Conn) {
// 	// send := func(msg string) {
// 	// 	conn.WriteMessage(websocket.TextMessage, []byte(msg))
// 	// }

// 	SendLog(conn, fmt.Sprintf("🔍 开始连接测试：%s:%d [%s]", host, port, protocol))
// 	time.Sleep(300 * time.Millisecond)

// 	SendLog(conn, "⚙️ 执行连接测试脚本...")
// 	time.Sleep(300 * time.Millisecond)

// 	log.Printf("Start testing host=%s, port=%d, protocol=%s", host, port, protocol)

// 	_, result, err := RunZGrab2WithProgress(protocol, host, fmt.Sprintf("%d", port), "ssl", conn)
// 	if err != nil {
// 		SendLog(conn, fmt.Sprintf("❌ 测试失败：%v", err))
// 		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "测试失败"))
// 		return
// 	}

// 	SendLog(conn, "✅ 测试完成，准备返回结构化结果")

// 	payload := map[string]interface{}{
// 		"type":   "result",
// 		"result": result,
// 	}
// 	finalJSON, _ := json.Marshal(payload)
// 	conn.WriteMessage(websocket.TextMessage, finalJSON)

// 	// finalJSON, _ := json.Marshal(result)
// 	// conn.WriteMessage(websocket.TextMessage, finalJSON)
// }

func RunZGrab2WithProgress(protocol, hostname, port, mode string, conn *websocket.Conn) (bool, *ConnectInfo, error) {
	SendLog(conn, "📦 正在执行 TLS 检测脚本...")

	pythonPath := "python" // 视你的系统为 python 或 python3
	scriptPath := "./tlscheck/test_tls.py"

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
	var result ConnectInfo
	if err := json.Unmarshal(outputBuf.Bytes(), &result); err != nil {
		return false, nil, fmt.Errorf("结果解析失败（非 JSON）：%v", err)
	}

	if !result.Success {
		return false, nil, fmt.Errorf("TLS 测试失败：%s", result.Error)
	}

	return true, &result, nil
}

// 8.18
// 上传CSV并返回处理结果文件下载链接
func UploadCsvAndExportJsonlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 限制大小为10MB
	if err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to retrieve file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 保存上传的文件到临时目录
	os.MkdirAll("tmp", os.ModePerm)
	tmpFilePath := filepath.Join("tmp", handler.Filename)
	tmpFile, err := os.Create(tmpFilePath)
	if err != nil {
		http.Error(w, "Failed to create temp file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer tmpFile.Close()
	io.Copy(tmpFile, file)

	// 重新打开读取（确保读取内容）
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

		// ✅ 移除 BOM（只在第一行第一列出现）
		if first {
			domain = strings.TrimPrefix(domain, "\uFEFF")
			first = false
		}

		domains = append(domains, domain)
	}

	// 创建输出目录
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
			//email := "info@" + domain
			result := processDomain(domain)
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

	// 返回结果文件路径
	downloadURL := "/downloads/" + filepath.Base(outputFile)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"download_url": downloadURL,
	})
}

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
		semaphore <- struct{}{}
		go func(domain string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			result := processDomain(domain) // 使用你已有的函数
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

// 处理单个域名
func processDomain(domain string) DomainResult {
	domainResult := DomainResult{
		Domain:        domain,
		Timestamp:     time.Now().Format(time.RFC3339),
		ErrorMessages: []string{},
	}
	//处理每个域名的一开始就查询CNAME字段
	email := "info@" + domain
	cnameRecords, err := lookupCNAME(domain)
	if err != nil {
		domainResult.ErrorMessages = append(domainResult.ErrorMessages, fmt.Sprintf("CNAME lookup error: %v", err))
	}
	domainResult.CNAME = cnameRecords
	// Autodiscover 查询
	autodiscoverResults := queryAutodiscover(domain, email)
	domainResult.Autodiscover = autodiscoverResults
	//domainResult.ErrorMessages = append(domainResult.ErrorMessages, errors...)
	// Autoconfig 查询
	autoconfigResults := queryAutoconfig(domain, email)
	domainResult.Autoconfig = autoconfigResults
	// if err := queryAutoconfig(domain, &result); err != nil {
	// 	result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("Autoconfig error: %v", err))
	// }
	// SRV 查询
	srvconfigResults := querySRV(domain)
	domainResult.SRV = srvconfigResults
	// if err := querySRV(domain, &result); err != nil {
	// 	result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("SRV error: %v", err))
	// }
	return domainResult
}

// 查询CNAME部分
func lookupCNAME(domain string) ([]string, error) {
	resolverAddr := "8.8.8.8:53" // Google Public DNS
	timeout := 5 * time.Second   // Timeout for DNS query

	client := &dns.Client{
		Net:     "udp",
		Timeout: timeout,
	}

	var lastErr error
	for i := 0; i < 3; i++ {
		m := dns.Msg{}
		m.SetQuestion(dns.Fqdn(domain), dns.TypeA) // 查询 A 记录
		r, _, err := client.Exchange(&m, resolverAddr)
		if err != nil {
			lastErr = err
			time.Sleep(1 * time.Second * time.Duration(i+1))
			continue
		}

		var dst []string
		for _, ans := range r.Answer {
			if record, ok := ans.(*dns.CNAME); ok {
				dst = append(dst, record.Target)
			}
		}

		if len(dst) > 0 {
			return dst, nil // 如果找到结果，立即返回
		}

		lastErr = nil
		break
	}

	return nil, lastErr
}

func queryAutodiscover(domain string, email string) []AutodiscoverResult {
	var results []AutodiscoverResult
	// //查询autodiscover.example.com的cname记录
	// autodiscover_prefixadd := "autodiscover." + domain
	// autodiscover_cnameRecords, _ := lookupCNAME(autodiscover_prefixadd)
	// method1:直接通过text manipulation，直接发出post请求
	uris := []string{
		fmt.Sprintf("http://%s/autodiscover/autodiscover.xml", domain),
		fmt.Sprintf("https://autodiscover.%s/autodiscover/autodiscover.xml", domain),
		fmt.Sprintf("http://autodiscover.%s/autodiscover/autodiscover.xml", domain),
		fmt.Sprintf("https://%s/autodiscover/autodiscover.xml", domain),
	}
	for i, uri := range uris {
		index := i + 1
		flag1, flag2, flag3, redirects, config, certinfo, err := getAutodiscoverConfig(domain, uri, email, "post", index, 0, 0, 0) //getAutodiscoverConfig照常
		fmt.Printf("flag1: %d\n", flag1)
		fmt.Printf("flag2: %d\n", flag2)
		fmt.Printf("flag3: %d\n", flag3)

		result := AutodiscoverResult{
			Domain:    domain,
			Method:    "POST",
			Index:     index,
			URI:       uri,
			Redirects: redirects,
			Config:    config,
			CertInfo:  certinfo,
		}
		if err != nil {
			result.Error = err.Error()
		}
		results = append(results, result)
	}

	//method2:通过dns找到server,再post请求
	service := "_autodiscover._tcp." + domain
	uriDNS, _, err := lookupSRVWithAD_autodiscover(domain) //
	if err != nil {
		result_srv := AutodiscoverResult{
			Domain: domain,
			Method: "srv-post",
			Index:  0,
			Error:  fmt.Sprintf("Failed to lookup SRV records for %s: %v", service, err),
		}
		results = append(results, result_srv)
	} else {
		//record_ADbit_SRV_autodiscover("autodiscover_record_ad_srv.txt", domain, adBit)
		_, _, _, redirects, config, certinfo, err1 := getAutodiscoverConfig(domain, uriDNS, email, "srv-post", 0, 0, 0, 0)
		result_srv := AutodiscoverResult{
			Domain:    domain,
			Method:    "srv-post",
			Index:     0,
			Redirects: redirects,
			Config:    config,
			CertInfo:  certinfo,
			//AutodiscoverCNAME: autodiscover_cnameRecords,
		}
		if err1 != nil {
			result_srv.Error = err1.Error()
		}
		results = append(results, result_srv)
	}

	//method3：先GET找到server，再post请求
	getURI := fmt.Sprintf("http://autodiscover.%s/autodiscover/autodiscover.xml", domain) //是通过这个getURI得到server的uri，然后再进行post请求10.26
	redirects, config, certinfo, err := GET_AutodiscoverConfig(domain, getURI, email)     //一开始的get请求返回的不是重定向的没有管
	result_GET := AutodiscoverResult{
		Domain:    domain,
		Method:    "get-post",
		Index:     0,
		URI:       getURI,
		Redirects: redirects,
		Config:    config,
		CertInfo:  certinfo,
		//AutodiscoverCNAME: autodiscover_cnameRecords,
	}
	if err != nil {
		result_GET.Error = err.Error()
	} //TODO:len(redirect)>0?
	results = append(results, result_GET)

	//method4:增加几条直接GET请求的路径
	direct_getURIs := []string{
		fmt.Sprintf("http://%s/autodiscover/autodiscover.xml", domain),               //uri1
		fmt.Sprintf("https://autodiscover.%s/autodiscover/autodiscover.xml", domain), //2
		fmt.Sprintf("http://autodiscover.%s/autodiscover/autodiscover.xml", domain),  //3
		fmt.Sprintf("https://%s/autodiscover/autodiscover.xml", domain),              //4
	}
	for i, direct_getURI := range direct_getURIs {
		index := i + 1
		_, _, _, redirects, config, certinfo, err := direct_GET_AutodiscoverConfig(domain, direct_getURI, email, "get", index, 0, 0, 0)
		result := AutodiscoverResult{
			Domain:    domain,
			Method:    "direct_get",
			Index:     index,
			URI:       direct_getURI,
			Redirects: redirects,
			Config:    config,
			CertInfo:  certinfo,
			//AutodiscoverCNAME: autodiscover_cnameRecords,
		}
		if err != nil {
			result.Error = err.Error()
		}
		results = append(results, result)
	}

	// //method5 猜测方法
	// guessed := guessMailServer(domain, 2*time.Second, 20)
	// for i, item := range guessed {
	// 	result := AutodiscoverResult{
	// 		Domain: domain,
	// 		Method: "guess",
	// 		Index:  i + 1,
	// 		URI:    item, // format: "host:port"
	// 	}
	// 	results = append(results, result)
	// }

	return results
}

func guessMailServer(domain string, timeout time.Duration, maxConcurrency int) []string {
	prefixMap := map[string][]string{
		"SMTP": {"smtp.", "smtps.", "mail.", "submission.", "mx."},
		"IMAP": {"imap.", "imap4.", "imaps.", "mail.", "mx."},
		"POP":  {"pop.", "pop3.", "pop3s.", "mail.", "mx."},
	}

	portMap := map[string][]int{
		"SMTP": {465, 587},
		"IMAP": {143, 993},
		"POP":  {110, 995},
	}

	var results []string
	var wg sync.WaitGroup
	var mu sync.Mutex
	semaphore := make(chan struct{}, maxConcurrency) // 控制最大并发数

	for proto, prefixes := range prefixMap {
		for _, port := range portMap[proto] {
			for _, prefix := range prefixes {
				wg.Add(1)
				go func(host string, port int) {
					defer wg.Done()
					semaphore <- struct{}{}        // 获取令牌
					defer func() { <-semaphore }() // 释放令牌

					conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
					if err == nil {
						conn.Close()
						mu.Lock()
						results = append(results, fmt.Sprintf("%s:%d", host, port))
						mu.Unlock()
					}
				}(prefix+domain, port)
			}
		}
	}

	wg.Wait()
	return results
}

func getAutodiscoverConfig(origin_domain string, uri string, email_add string, method string, index int, flag1 int, flag2 int, flag3 int) (int, int, int, []map[string]interface{}, string, *CertInfo, error) {
	xmlRequest := fmt.Sprintf(`
		<Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/outlook/requestschema/2006">
			<Request>
				<EMailAddress>%s</EMailAddress>
				<AcceptableResponseSchema>http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a</AcceptableResponseSchema>
			</Request>
		</Autodiscover>`, email_add)

	req, err := http.NewRequest("POST", uri, bytes.NewBufferString(xmlRequest))
	if err != nil {
		fmt.Printf("Error creating request for %s: %v\n", uri, err)
		return flag1, flag2, flag3, []map[string]interface{}{}, "", nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "text/xml")
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS10,
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 禁止重定向
		},
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request to %s: %v\n", uri, err)
		return flag1, flag2, flag3, []map[string]interface{}{}, "", nil, fmt.Errorf("failed to send request: %v", err)
	}

	redirects := getRedirects(resp) // 获取当前重定向链
	defer resp.Body.Close()         //
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		// 处理重定向
		flag1 = flag1 + 1
		fmt.Printf("flag1now:%d\n", flag1)
		location := resp.Header.Get("Location")
		fmt.Printf("Redirect to: %s\n", location)
		if location == "" {
			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("missing Location header in redirect")
		} else if flag1 > 10 { //12.27限制重定向次数
			//saveXMLToFile_autodiscover("./location.xml", origin_domain, email_add)
			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("too many redirect times")
		}

		newURI, err := url.Parse(location)
		if err != nil {
			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("failed to parse redirect URL: %s", location)
		}

		// 递归调用并合并重定向链
		newflag1, newflag2, newflag3, nextRedirects, result, certinfo, err := getAutodiscoverConfig(origin_domain, newURI.String(), email_add, method, index, flag1, flag2, flag3)
		//return append(redirects, nextRedirects...), result, err //12.27原
		return newflag1, newflag2, newflag3, append(redirects, nextRedirects...), result, certinfo, err
	} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// 处理成功响应
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("failed to read response body: %v", err)
		}

		var autodiscoverResp AutodiscoverResponse
		err = xml.Unmarshal(body, &autodiscoverResp)
		//这里先记录下unmarshal就不成功的xml
		if err != nil {
			// if (strings.HasPrefix(strings.TrimSpace(string(body)), `<?xml version="1.0"`) || strings.HasPrefix(strings.TrimSpace(string(body)), `<Autodiscover`)) && !strings.Contains(strings.TrimSpace(string(body)), `<html`) && !strings.Contains(strings.TrimSpace(string(body)), `<item`) && !strings.Contains(strings.TrimSpace(string(body)), `lastmod`) && !strings.Contains(strings.TrimSpace(string(body)), `lt`) {
			// 	//if !strings.Contains(strings.TrimSpace(string(body)), `<html`) && !strings.Contains(strings.TrimSpace(string(body)), `<item`) && !strings.Contains(strings.TrimSpace(string(body)), `lastmod`) && !strings.Contains(strings.TrimSpace(string(body)), `lt`) {
			// 	//saveno_XMLToFile("no_autodiscover_config.xml", string(body), email_add)
			// } //记录错误格式的xml
			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("failed to unmarshal XML: %v", err)
		}

		// 处理 redirectAddr 和 redirectUrl
		if autodiscoverResp.Response.Account.Action == "redirectAddr" {
			flag2 = flag2 + 1
			newEmail := autodiscoverResp.Response.Account.RedirectAddr
			//record_filename := filepath.Join("./autodiscover/records", "ReAddr.xml")
			//saveXMLToFile_with_ReAdrr_autodiscover(record_filename, string(body), email_add)
			if newEmail != "" && flag2 <= 10 {
				newflag1, newflag2, newflag3, nextRedirects, result, certinfo, err := getAutodiscoverConfig(origin_domain, uri, newEmail, method, index, flag1, flag2, flag3)
				return newflag1, newflag2, newflag3, append(redirects, nextRedirects...), result, certinfo, err
			} else if newEmail != "" { //12.27
				//saveXMLToFile_autodiscover("./flag2.xml", origin_domain, email_add)
				return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("too many RedirectAddr")
			} else {
				return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("nil ReAddr")
			}
		} else if autodiscoverResp.Response.Account.Action == "redirectUrl" {
			flag3 = flag3 + 1
			newUri := autodiscoverResp.Response.Account.RedirectUrl
			//record_filename := filepath.Join("./autodiscover/records", "Reurl.xml")
			//saveXMLToFile_with_Reuri_autodiscover(record_filename, string(body), email_add)
			if newUri != "" && flag3 <= 10 {
				newflag1, newflag2, newflag3, nextRedirects, result, certinfo, err := getAutodiscoverConfig(origin_domain, newUri, email_add, method, index, flag1, flag2, flag3)
				return newflag1, newflag2, newflag3, append(redirects, nextRedirects...), result, certinfo, err
			} else if newUri != "" {
				//saveXMLToFile_autodiscover("./flag3.xml", origin_domain, email_add)
				return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("too many RedirectUrl")
			} else {
				return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("nil Reuri")
			}
		} else if autodiscoverResp.Response.Account.Action == "settings" { //这才是我们需要的
			// 记录并返回成功配置(3.13修改，因为会将Response命名空间不合规的也解析到这里)
			// outputfile := fmt.Sprintf("./autodiscover/autodiscover_%s_%d_config.xml", method, index)
			// saveXMLToFile_autodiscover(outputfile, string(body), email_add)

			//只在可以直接返回xml配置的时候记录证书信息
			var certInfo CertInfo
			// 提取证书信息
			if resp.TLS != nil {
				//var encodedData []byte //8.15
				var encodedCerts []string //8.15
				goChain := resp.TLS.PeerCertificates
				endCert := goChain[0]

				// 证书验证
				dnsName := resp.Request.URL.Hostname()
				var VerifyError error
				certInfo.IsTrusted, VerifyError = verifyCertificate(goChain, dnsName)
				if VerifyError != nil {
					certInfo.VerifyError = VerifyError.Error()
				} else {
					certInfo.VerifyError = ""
				}

				certInfo.IsExpired = endCert.NotAfter.Before(time.Now())
				certInfo.IsHostnameMatch = verifyHostname(endCert, dnsName)
				certInfo.IsSelfSigned = IsSelfSigned(endCert)
				certInfo.IsInOrder = isChainInOrder(goChain)
				certInfo.TLSVersion = resp.TLS.Version

				// 提取证书的其他信息
				certInfo.Subject = endCert.Subject.CommonName
				certInfo.Issuer = endCert.Issuer.String()
				certInfo.SignatureAlg = endCert.SignatureAlgorithm.String()
				certInfo.AlgWarning = algWarnings(endCert)

				// 将证书编码为 base64 格式
				for _, cert := range goChain {
					encoded := base64.StdEncoding.EncodeToString(cert.Raw)
					//encodedData = append(encodedData, []byte(encoded)...)//8.16
					encodedCerts = append(encodedCerts, encoded)
				}
				//certInfo.RawCert = encodedData //8.15
				certInfo.RawCerts = encodedCerts
			}
			return flag1, flag2, flag3, redirects, string(body), &certInfo, nil
		} else if autodiscoverResp.Response.Error != nil {
			//fmt.Printf("Error: %s\n", string(body))
			// 处理错误响应
			errorConfig := fmt.Sprintf("Errorcode:%d-%s\n", autodiscoverResp.Response.Error.ErrorCode, autodiscoverResp.Response.Error.Message)
			//outputfile := fmt.Sprintf("./autodiscover/autodiscover_%s_%d_Errorconfig.txt", method, index)
			//saveXMLToFile_autodiscover(outputfile, errorConfig, email_add)
			return flag1, flag2, flag3, redirects, errorConfig, nil, nil
		} else {
			//fmt.Printf("Response element not valid:%s\n", string(body))
			//处理Response可能本身就不正确的响应,同时也会存储不合规的xml(unmarshal的时候合规但Response不合规)
			alsoErrorConfig := fmt.Sprintf("Non-valid Response element for %s\n:", email_add)
			//outputfile := fmt.Sprintf("./autodiscover/autodiscover_%s_%d_AlsoErrorConfig.xml", method, index)
			//saveXMLToFile_autodiscover(outputfile, string(body), email_add)
			return flag1, flag2, flag3, redirects, alsoErrorConfig, nil, nil
		}
	} else {
		// 处理非成功响应
		//outputfile := fmt.Sprintf("./autodiscover/autodiscover_%s_%d_badresponse.txt", method, index)
		badResponse := fmt.Sprintf("Bad response for %s: %d\n", email_add, resp.StatusCode)
		//saveXMLToFile_autodiscover(outputfile, badResponse, email_add)
		return flag1, flag2, flag3, redirects, badResponse, nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
func GET_AutodiscoverConfig(origin_domain string, uri string, email_add string) ([]map[string]interface{}, string, *CertInfo, error) { //使用先get后post方法
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS10,
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 禁止重定向
		},
		Timeout: 15 * time.Second,
	}
	resp, err := client.Get(uri)
	if err != nil {
		return []map[string]interface{}{}, "", nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	redirects := getRedirects(resp) // 获取当前重定向链

	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently { //仅通过get请求获取重定向地址
		location := resp.Header.Get("Location")
		fmt.Printf("Redirect to: %s\n", location)
		if location == "" {
			return nil, "", nil, fmt.Errorf("missing Location header in redirect")
		}
		newURI, err := url.Parse(location)
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to parse redirect URL: %s", location)
		}

		// 递归调用并合并重定向链
		_, _, _, nextRedirects, result, certinfo, err := getAutodiscoverConfig(origin_domain, newURI.String(), email_add, "get_post", 0, 0, 0, 0)
		return append(redirects, nextRedirects...), result, certinfo, err
	} else {
		return nil, "", nil, fmt.Errorf("not find Redirect Statuscode")
	}
}

func direct_GET_AutodiscoverConfig(origin_domain string, uri string, email_add string, method string, index int, flag1 int, flag2 int, flag3 int) (int, int, int, []map[string]interface{}, string, *CertInfo, error) { //一路get请求
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS10,
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 禁止重定向
		},
		Timeout: 15 * time.Second, // 设置请求超时时间
	}
	resp, err := client.Get(uri)
	if err != nil {
		return flag1, flag2, flag3, []map[string]interface{}{}, "", nil, fmt.Errorf("failed to send request: %v", err)
	}

	redirects := getRedirects(resp)
	defer resp.Body.Close() //

	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		flag1 = flag1 + 1
		location := resp.Header.Get("Location")
		fmt.Printf("Redirect to: %s\n", location)
		if location == "" {
			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("missing Location header in redirect")
		} else if flag1 > 10 {
			//saveXMLToFile_autodiscover("./location2.xml", origin_domain, email_add)
			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("too many redirect times")
		}

		newURI, err := url.Parse(location)
		if err != nil {
			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("failed to parse redirect URL: %s", location)
		}

		// 递归调用并合并重定向链
		newflag1, newflag2, newflag3, nextRedirects, result, certinfo, err := direct_GET_AutodiscoverConfig(origin_domain, newURI.String(), email_add, method, index, flag1, flag2, flag3)
		return newflag1, newflag2, newflag3, append(redirects, nextRedirects...), result, certinfo, err
	} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("failed to read response body: %v", err)
		}
		var autodiscoverResp AutodiscoverResponse
		err = xml.Unmarshal(body, &autodiscoverResp)
		if err != nil {
			// if (strings.HasPrefix(strings.TrimSpace(string(body)), `<?xml version="1.0"`) || strings.HasPrefix(strings.TrimSpace(string(body)), `<Autodiscover`)) && !strings.Contains(strings.TrimSpace(string(body)), `<html`) && !strings.Contains(strings.TrimSpace(string(body)), `<item`) && !strings.Contains(strings.TrimSpace(string(body)), `lastmod`) && !strings.Contains(strings.TrimSpace(string(body)), `lt`) {
			// 	//if !strings.Contains(strings.TrimSpace(string(body)), `<html`) && !strings.Contains(strings.TrimSpace(string(body)), `<item`) && !strings.Contains(strings.TrimSpace(string(body)), `lastmod`) && !strings.Contains(strings.TrimSpace(string(body)), `lt`) {
			// 	saveno_XMLToFile("no_autodiscover_config_directget.xml", string(body), email_add)
			// } //记录错误格式的xml
			return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("failed to unmarshal XML: %v", err)
		}
		if autodiscoverResp.Response.Account.Action == "redirectAddr" {
			flag2 = flag2 + 1
			newEmail := autodiscoverResp.Response.Account.RedirectAddr
			//outputfile := fmt.Sprintf("./autodiscover/autodiscover_%s_%d_redirectAddr_config.xml", method, index)
			//saveXMLToFile_autodiscover(outputfile, string(body), email_add)
			if newEmail != "" {
				return flag1, flag2, flag3, redirects, string(body), nil, nil //TODO, 这里直接返回带redirect_email了
			} else {
				return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("nil ReAddr")
			}
		} else if autodiscoverResp.Response.Account.Action == "redirectUrl" {
			flag3 = flag3 + 1
			newUri := autodiscoverResp.Response.Account.RedirectUrl
			//record_filename := filepath.Join("./autodiscover/records", "Reurl_dirGET.xml")
			//saveXMLToFile_with_Reuri_autodiscover(record_filename, string(body), email_add) //记录redirecturi,是否会出现继续reUri?
			if newUri != "" && flag3 <= 10 {
				newflag1, newflag2, newflag3, nextRedirects, result, certinfo, err := direct_GET_AutodiscoverConfig(origin_domain, newUri, email_add, method, index, flag1, flag2, flag3)
				return newflag1, newflag2, newflag3, append(redirects, nextRedirects...), result, certinfo, err
			} else if newUri != "" {
				//saveXMLToFile_autodiscover("./flag32.xml", origin_domain, email_add)
				return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("too many RedirectUrl")
			} else {
				return flag1, flag2, flag3, redirects, "", nil, fmt.Errorf("nil Reurl")
			}
		} else if autodiscoverResp.Response.Account.Action == "settings" {
			//outputfile := fmt.Sprintf("./autodiscover/autodiscover_%s_%d_config.xml", method, index)
			//saveXMLToFile_autodiscover(outputfile, string(body), email_add)
			//只在可以直接返回xml配置的时候记录证书信息
			var certInfo CertInfo
			// 提取证书信息
			if resp.TLS != nil {
				//var encodedData []byte
				var encodedCerts []string
				goChain := resp.TLS.PeerCertificates
				endCert := goChain[0]

				// 证书验证
				dnsName := resp.Request.URL.Hostname()

				var VerifyError error
				certInfo.IsTrusted, VerifyError = verifyCertificate(goChain, dnsName)
				if VerifyError != nil {
					certInfo.VerifyError = VerifyError.Error()
				} else {
					certInfo.VerifyError = ""
				}
				certInfo.IsExpired = endCert.NotAfter.Before(time.Now())
				certInfo.IsHostnameMatch = verifyHostname(endCert, dnsName)
				certInfo.IsSelfSigned = IsSelfSigned(endCert)
				certInfo.IsInOrder = isChainInOrder(goChain)
				certInfo.TLSVersion = resp.TLS.Version

				// 提取证书的其他信息
				certInfo.Subject = endCert.Subject.CommonName
				certInfo.Issuer = endCert.Issuer.String()
				certInfo.SignatureAlg = endCert.SignatureAlgorithm.String()
				certInfo.AlgWarning = algWarnings(endCert)

				// 将证书编码为 base64 格式
				for _, cert := range goChain {
					encoded := base64.StdEncoding.EncodeToString(cert.Raw)
					//encodedData = append(encodedData, []byte(encoded)...)
					encodedCerts = append(encodedCerts, encoded)
				}
				//certInfo.RawCert = encodedData
				certInfo.RawCerts = encodedCerts
			}
			return flag1, flag2, flag3, redirects, string(body), &certInfo, nil
		} else if autodiscoverResp.Response.Error != nil {
			//fmt.Printf("Error: %s\n", string(body))
			// 处理错误响应
			errorConfig := fmt.Sprintf("Errorcode:%d-%s\n", autodiscoverResp.Response.Error.ErrorCode, autodiscoverResp.Response.Error.Message)
			//outputfile := fmt.Sprintf("./autodiscover/autodiscover_%s_%d_Errorconfig.txt", method, index)
			//saveXMLToFile_autodiscover(outputfile, errorConfig, email_add)
			return flag1, flag2, flag3, redirects, errorConfig, nil, nil
		} else {
			//fmt.Printf("Response element not valid:%s\n", string(body))
			//处理Response可能本身就不正确的响应,同时也会存储不合规的xml(unmarshal的时候合规但Response不合规)
			alsoErrorConfig := fmt.Sprintf("Non-valid Response element for %s\n:", email_add)
			//outputfile := fmt.Sprintf("./autodiscover/autodiscover_%s_%d_AlsoErrorConfig.xml", method, index)
			//saveXMLToFile_autodiscover(outputfile, string(body), email_add)
			return flag1, flag2, flag3, redirects, alsoErrorConfig, nil, nil
		}
	} else {
		//outputfile := fmt.Sprintf("./autodiscover/autodiscover_%s_%d_badresponse.txt", method, index)
		bad_response := fmt.Sprintf("Bad response for %s:%d\n", email_add, resp.StatusCode)
		//saveXMLToFile_autodiscover(outputfile, bad_response, email_add)
		return flag1, flag2, flag3, redirects, bad_response, nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode) //同时也想记录请求发送失败时的状态码
	}
}

// 查询Autoconfig部分
func queryAutoconfig(domain string, email string) []AutoconfigResult {
	var results []AutoconfigResult
	//method1 直接通过url发送get请求得到config
	urls := []string{
		fmt.Sprintf("https://autoconfig.%s/mail/config-v1.1.xml?emailaddress=%s", domain, email),             //uri1
		fmt.Sprintf("https://%s/.well-known/autoconfig/mail/config-v1.1.xml?emailaddress=%s", domain, email), //uri2
		fmt.Sprintf("http://autoconfig.%s/mail/config-v1.1.xml?emailaddress=%s", domain, email),              //uri3
		fmt.Sprintf("http://%s/.well-known/autoconfig/mail/config-v1.1.xml?emailaddress=%s", domain, email),  //uri4
	}
	for i, url := range urls {
		index := i + 1
		config, redirects, certinfo, err := Get_autoconfig_config(domain, url, "directurl", index)

		result := AutoconfigResult{
			Domain:    domain,
			Method:    "directurl",
			Index:     index,
			URI:       url,
			Redirects: redirects,
			Config:    config,
			CertInfo:  certinfo,
		}
		if err != nil {
			result.Error = err.Error()
		}
		results = append(results, result)
	}

	//method2 ISPDB
	ISPurl := fmt.Sprintf("https://autoconfig.thunderbird.net/v1.1/%s", domain)
	config, redirects, certinfo, err := Get_autoconfig_config(domain, ISPurl, "ISPDB", 0)
	result_ISPDB := AutoconfigResult{
		Domain:    domain,
		Method:    "ISPDB",
		Index:     0,
		URI:       ISPurl,
		Redirects: redirects,
		Config:    config,
		CertInfo:  certinfo,
	}
	if err != nil {
		result_ISPDB.Error = err.Error()
	}
	results = append(results, result_ISPDB)

	//method3 MX查询
	mxHost, err := ResolveMXRecord(domain)
	if err != nil {
		result_MX := AutoconfigResult{
			Domain: domain,
			Method: "MX",
			Index:  0,
			Error:  fmt.Sprintf("Resolve MX Record error for %s: %v", domain, err),
		}
		results = append(results, result_MX)
	} else {
		mxFullDomain, mxMainDomain, err := extractDomains(mxHost)
		if err != nil {
			result_MX := AutoconfigResult{
				Domain: domain,
				Method: "MX",
				Index:  0,
				Error:  fmt.Sprintf("extract domain from mxHost error for %s: %v", domain, err),
			}
			results = append(results, result_MX)
		} else {
			if mxFullDomain == mxMainDomain {
				urls := []string{
					fmt.Sprintf("https://autoconfig.%s/mail/config-v1.1.xml?emailaddress=%s", mxFullDomain, email), //1
					fmt.Sprintf("https://autoconfig.thunderbird.net/v1.1/%s", mxFullDomain),                        //3
				}
				for i, url := range urls {
					config, redirects, certinfo, err := Get_autoconfig_config(domain, url, "MX_samedomain", i*2+1)
					result := AutoconfigResult{
						Domain:    domain,
						Method:    "MX_samedomain",
						Index:     i*2 + 1,
						URI:       url,
						Redirects: redirects,
						Config:    config,
						CertInfo:  certinfo,
					}
					if err != nil {
						result.Error = err.Error()
					}
					results = append(results, result)
				}
			} else {
				urls := []string{
					fmt.Sprintf("https://autoconfig.%s/mail/config-v1.1.xml?emailaddress=%s", mxFullDomain, email), //1
					fmt.Sprintf("https://autoconfig.%s/mail/config-v1.1.xml?emailaddress=%s", mxMainDomain, email), //2
					fmt.Sprintf("https://autoconfig.thunderbird.net/v1.1/%s", mxFullDomain),                        //3
					fmt.Sprintf("https://autoconfig.thunderbird.net/v1.1/%s", mxMainDomain),                        //4
				}
				for i, url := range urls {
					config, redirects, certinfo, err := Get_autoconfig_config(domain, url, "MX", i+1)
					result := AutoconfigResult{
						Domain:    domain,
						Method:    "MX",
						Index:     i + 1,
						URI:       url,
						Redirects: redirects,
						Config:    config,
						CertInfo:  certinfo,
					}
					if err != nil {
						result.Error = err.Error()
					}
					results = append(results, result)
				}
			}
		}

	}
	return results

}

func Get_autoconfig_config(domain string, url string, method string, index int) (string, []map[string]interface{}, *CertInfo, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS10,
			},
		},
		Timeout: 15 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", []map[string]interface{}{}, nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", []map[string]interface{}{}, nil, err
	}
	// 获取重定向历史记录
	redirects := getRedirects(resp)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", redirects, nil, fmt.Errorf("failed to read response body: %v", err)
	}
	var autoconfigResp AutoconfigResponse
	err = xml.Unmarshal(body, &autoconfigResp)
	if err != nil {
		// if (strings.HasPrefix(strings.TrimSpace(string(body)), `<?xml version="1.0"`) || strings.HasPrefix(strings.TrimSpace(string(body)), `<clientConfig`)) && !strings.Contains(strings.TrimSpace(string(body)), `<html`) && !strings.Contains(strings.TrimSpace(string(body)), `<item`) && !strings.Contains(strings.TrimSpace(string(body)), `lastmod`) && !strings.Contains(strings.TrimSpace(string(body)), `lt`) {
		// 	saveno_XMLToFile("no_autoconfig_config.xml", string(body), domain)
		// }
		return "", redirects, nil, fmt.Errorf("failed to unmarshal XML: %v", err)
	} else {
		var certInfo CertInfo
		// 提取证书信息
		if resp.TLS != nil {
			//var encodedData []byte
			var encodedCerts []string
			goChain := resp.TLS.PeerCertificates
			endCert := goChain[0]

			// 证书验证
			dnsName := resp.Request.URL.Hostname()
			var VerifyError error
			certInfo.IsTrusted, VerifyError = verifyCertificate(goChain, dnsName)
			if VerifyError != nil {
				certInfo.VerifyError = VerifyError.Error()
			} else {
				certInfo.VerifyError = ""
			}
			certInfo.IsExpired = endCert.NotAfter.Before(time.Now())
			certInfo.IsHostnameMatch = verifyHostname(endCert, dnsName)
			certInfo.IsSelfSigned = IsSelfSigned(endCert)
			certInfo.IsInOrder = isChainInOrder(goChain)
			certInfo.TLSVersion = resp.TLS.Version

			// 提取证书的其他信息
			certInfo.Subject = endCert.Subject.CommonName
			certInfo.Issuer = endCert.Issuer.String()
			certInfo.SignatureAlg = endCert.SignatureAlgorithm.String()
			certInfo.AlgWarning = algWarnings(endCert)

			// 将证书编码为 base64 格式
			for _, cert := range goChain {
				encoded := base64.StdEncoding.EncodeToString(cert.Raw)
				//encodedData = append(encodedData, []byte(encoded)...)
				encodedCerts = append(encodedCerts, encoded)
			}
			//certInfo.RawCert = encodedData
			certInfo.RawCerts = encodedCerts

		}

		config := string(body)
		// outputfile := fmt.Sprintf("./autoconfig/autoconfig_%s_%d.xml", method, index) //12.18 用Index加以区分
		// err = saveXMLToFile_autoconfig(outputfile, config, domain)
		// if err != nil {
		// 	return "", redirects, &certInfo, err
		// }
		return config, redirects, &certInfo, nil
	}
}

// 获取MX记录
func ResolveMXRecord(domain string) (string, error) {
	//创建DNS客户端并设置超时时间
	client := &dns.Client{
		Timeout: 15 * time.Second, // 设置超时时间
	}

	// 创建DNS消息
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeMX)
	//发送DNS查询
	response, _, err := client.Exchange(msg, dnsServer)
	if err != nil {
		fmt.Printf("Failed to query DNS for %s: %v\n", domain, err)
		return "", err
	}

	//处理响应
	if response.Rcode != dns.RcodeSuccess {
		fmt.Printf("DNS query failed with Rcode %d\n", response.Rcode)
		return "", fmt.Errorf("DNS query failed with Rcode %d", response.Rcode)
	}

	var mxRecords []*dns.MX
	for _, ans := range response.Answer {
		if mxRecord, ok := ans.(*dns.MX); ok {
			fmt.Printf("MX record for %s: %s, the priority is %d\n", domain, mxRecord.Mx, mxRecord.Preference)
			mxRecords = append(mxRecords, mxRecord)
		}
	}
	if len(mxRecords) == 0 {
		return "", fmt.Errorf("no MX Record")
	}

	// 根据Preference字段排序，Preference值越小优先级越高
	sort.Slice(mxRecords, func(i, j int) bool {
		return mxRecords[i].Preference < mxRecords[j].Preference
	})
	highestMX := mxRecords[0]
	return highestMX.Mx, nil

}

// 提取%MXFULLDOMAIN%和%MXMAINDOMAIN%
func extractDomains(mxHost string) (string, string, error) {
	mxHost = strings.TrimSuffix(mxHost, ".")

	// 获取%MXFULLDOMAIN%
	parts := strings.Split(mxHost, ".")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid MX Host name: %s", mxHost)
	}
	mxFullDomain := strings.Join(parts[1:], ".")
	fmt.Println("fulldomain:", mxFullDomain)

	// 获取%MXMAINDOMAIN%（提取第二级域名）
	mxMainDomain, err := publicsuffix.EffectiveTLDPlusOne(mxHost)
	if err != nil {
		return "", "", fmt.Errorf("cannot extract maindomain: %v", err)
	}
	fmt.Println("maindomain:", mxMainDomain)

	return mxFullDomain, mxMainDomain, nil
}

func querySRV(domain string) SRVResult {
	var dnsrecord DNSRecord
	dnsManager, isSOA, err := queryDNSManager(domain)
	if err != nil {
		fmt.Printf("Failed to query DNS manager for %s: %v\n", domain, err)
	} else {
		if isSOA {
			dnsrecord = DNSRecord{
				Domain: domain,
				SOA:    dnsManager,
			}
		} else {
			dnsrecord = DNSRecord{
				Domain: domain,
				NS:     dnsManager,
			}
		}
	}

	// 定义要查询的服务标签
	recvServices := []string{
		"_imap._tcp." + domain,
		"_imaps._tcp." + domain,
		"_pop3._tcp." + domain,
		"_pop3s._tcp." + domain,
	}
	sendServices := []string{
		"_submission._tcp." + domain,
		"_submissions._tcp." + domain,
	}

	var recvRecords, sendRecords []SRVRecord

	// 查询(IMAP/POP3)
	for _, service := range recvServices {
		records, adBit, err := lookupSRVWithAD_srv(service)
		//record_ADbit_SRV(service, "SRV_record_ad_srv.txt", domain, adBit)

		if err != nil || len(records) == 0 {
			fmt.Printf("Failed to query SRV for %s or no records found: %v\n", service, err)
			continue
		}

		// 更新 DNSRecord 的 AD 位
		if strings.HasPrefix(service, "_imaps") {
			dnsrecord.ADbit_imaps = &adBit
		} else if strings.HasPrefix(service, "_imap") {
			dnsrecord.ADbit_imap = &adBit
		} else if strings.HasPrefix(service, "_pop3s") {
			dnsrecord.ADbit_pop3s = &adBit
		} else if strings.HasPrefix(service, "_pop3") {
			dnsrecord.ADbit_pop3 = &adBit
		}

		// 添加 SRV 记录
		for _, record := range records {
			if record.Target == "." {
				continue
			}
			recvRecords = append(recvRecords, SRVRecord{
				Service:  service,
				Priority: record.Priority,
				Weight:   record.Weight,
				Port:     record.Port,
				Target:   record.Target,
			})
		}
	}

	// 查询 (SMTP)
	for _, service := range sendServices {
		records, adBit, err := lookupSRVWithAD_srv(service)
		//record_ADbit_SRV(service, "SRV_record_ad_srv.txt", domain, adBit)

		if err != nil || len(records) == 0 {
			fmt.Printf("Failed to query SRV for %s or no records found: %v\n", service, err)
			continue
		}

		// 更新 DNSRecord 的 AD 位
		if strings.HasPrefix(service, "_submissions") {
			dnsrecord.ADbit_smtps = &adBit
		} else if strings.HasPrefix(service, "_submission") {
			dnsrecord.ADbit_smtp = &adBit
		}

		// 添加 SRV 记录
		for _, record := range records {
			if record.Target == "." {
				continue
			}
			sendRecords = append(sendRecords, SRVRecord{
				Service:  service,
				Priority: record.Priority,
				Weight:   record.Weight,
				Port:     record.Port,
				Target:   record.Target,
			})
		}
	}

	// 对收件服务和发件服务进行排序
	sort.Slice(recvRecords, func(i, j int) bool {
		if recvRecords[i].Priority == recvRecords[j].Priority {
			return recvRecords[i].Weight > recvRecords[j].Weight
		}
		return recvRecords[i].Priority < recvRecords[j].Priority
	})

	sort.Slice(sendRecords, func(i, j int) bool {
		if sendRecords[i].Priority == sendRecords[j].Priority {
			return sendRecords[i].Weight > sendRecords[j].Weight
		}
		return sendRecords[i].Priority < sendRecords[j].Priority
	})

	// 返回组合后的结果
	return SRVResult{
		Domain:      domain,
		DNSRecord:   &dnsrecord,
		RecvRecords: recvRecords,
		SendRecords: sendRecords,
	}
}

func queryDNSManager(domain string) (string, bool, error) {
	resolverAddr := "8.8.8.8:53" // Google Public DNS
	timeout := 15 * time.Second  // DNS 查询超时时间

	client := &dns.Client{
		Net:     "udp",
		Timeout: timeout,
	}

	// 查询 SOA 记录
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeSOA)
	response, _, err := client.Exchange(msg, resolverAddr)
	if err != nil {
		return "", false, fmt.Errorf("SOA query failed: %v", err)
	}

	// 提取 SOA 记录的管理者信息
	for _, ans := range response.Answer {
		if soa, ok := ans.(*dns.SOA); ok {
			return soa.Ns, true, nil // SOA 记录中的权威 DNS 服务器名称
		}
	}

	// 若 SOA 查询无结果，尝试查询 NS 记录
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
	response, _, err = client.Exchange(msg, resolverAddr)
	if err != nil {
		return "", false, fmt.Errorf("NS query failed: %v", err)
	}

	var nsRecords []string
	for _, ans := range response.Answer {
		if ns, ok := ans.(*dns.NS); ok {
			nsRecords = append(nsRecords, ns.Ns)
		}
	}

	if len(nsRecords) > 0 {
		return strings.Join(nsRecords, ", "), false, nil // 返回 NS 记录列表
	}

	return "", false, fmt.Errorf("no SOA or NS records found for domain: %s", domain)
}

func lookupSRVWithAD_srv(service string) ([]*dns.SRV, bool, error) {
	// DNS Resolver configuration
	resolverAddr := "8.8.8.8:53" // Google Public DNS
	timeout := 15 * time.Second  // Timeout for DNS query

	// Create a DNS client
	client := &dns.Client{
		Net:     "udp", //
		Timeout: timeout,
	}
	// Create the SRV query
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(service), dns.TypeSRV)
	msg.RecursionDesired = true // Enable recursion
	msg.SetEdns0(4096, true)    // true 表示启用 DO 位，支持 DNSSEC

	// Perform the DNS query
	response, _, err := client.Exchange(msg, resolverAddr)
	if err != nil {
		return nil, false, fmt.Errorf("DNS query failed: %v", err)
	}

	// Check the AD bit in the DNS response flags
	adBit := response.AuthenticatedData
	// 解析 SRV 记录
	var srvRecords []*dns.SRV
	for _, ans := range response.Answer {
		if srv, ok := ans.(*dns.SRV); ok {
			srvRecords = append(srvRecords, srv)
		}
	}
	fmt.Printf("service:%s, adBit:%v\n", service, adBit)
	return srvRecords, adBit, nil
}

func getRedirects(resp *http.Response) (history []map[string]interface{}) {
	for resp != nil {
		req := resp.Request
		status := resp.StatusCode
		entry := map[string]interface{}{
			"URL":    req.URL.String(),
			"Status": status,
		}
		history = append(history, entry)
		resp = resp.Request.Response
	}
	if len(history) >= 1 {
		for l, r := 0, len(history)-1; l < r; l, r = l+1, r-1 {
			history[l], history[r] = history[r], history[l]
		}
	}
	return history
}

func lookupSRVWithAD_autodiscover(domain string) (string, bool, error) {
	// DNS Resolver configuration
	resolverAddr := "8.8.8.8:53" // Google Public DNS
	timeout := 5 * time.Second   // Timeout for DNS query

	// Create a DNS client
	client := &dns.Client{
		Net:     "udp", //
		Timeout: timeout,
	}

	// Create the SRV query
	service := "_autodiscover._tcp." + domain
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(service), dns.TypeSRV)
	msg.RecursionDesired = true // Enable recursion
	msg.SetEdns0(4096, true)    // true 表示启用 DO 位，支持 DNSSEC

	// Perform the DNS query
	response, _, err := client.Exchange(msg, resolverAddr)
	if err != nil {
		return "", false, fmt.Errorf("DNS query failed: %v", err)
	}

	// Check the AD bit in the DNS response flags
	adBit := response.AuthenticatedData

	var srvRecords []*dns.SRV
	for _, ans := range response.Answer {
		if srv, ok := ans.(*dns.SRV); ok {
			srvRecords = append(srvRecords, srv)
		}
	}
	var uriDNS string
	if len(srvRecords) > 0 {
		sort.Slice(srvRecords, func(i, j int) bool {
			if srvRecords[i].Priority == srvRecords[j].Priority {
				return srvRecords[i].Weight > srvRecords[j].Weight
			}
			return srvRecords[i].Priority < srvRecords[j].Priority
		})

		hostname := srvRecords[0].Target
		port := srvRecords[0].Port
		if hostname != "." {
			if port == 443 {
				uriDNS = fmt.Sprintf("https://%s/autodiscover/autodiscover.xml", hostname)
			} else if port == 80 {
				uriDNS = fmt.Sprintf("http://%s/autodiscover/autodiscover.xml", hostname)
			} else {
				uriDNS = fmt.Sprintf("https://%s:%d/autodiscover/autodiscover.xml", hostname, port)
			}
		} else {
			return "", adBit, fmt.Errorf("hostname == '.'")
		}
	} else {
		return "", adBit, fmt.Errorf("no srvRecord found")
	}

	return uriDNS, adBit, nil
}

func verifyCertificate(chain []*x509.Certificate, domain string) (bool, error) {
	if len(chain) == 1 {
		temp_chain, err := certUtil.FetchCertificateChain(chain[0])
		if err != nil {
			//log.Println("failed to fetch certificate chain")
			return false, fmt.Errorf("failed to fetch certificate chain:%v", err)
		}
		chain = temp_chain
	}

	intermediates := x509.NewCertPool()
	for i := 1; i < len(chain); i++ {
		intermediates.AddCert(chain[i])
	}

	certPool := x509.NewCertPool()
	pemFile := "IncludedRootsPEM313.txt" //修改获取roots的途径
	pem, err := os.ReadFile(pemFile)
	if err != nil {
		//log.Println("failed to read root certificate")
		return false, fmt.Errorf("failed to read root certificate:%v", err)
	}
	ok := certPool.AppendCertsFromPEM(pem)
	if !ok {
		//log.Println("failed to import root certificate")
		return false, fmt.Errorf("failed to import root certificate:%v", err)
	}

	opts := x509.VerifyOptions{
		Roots:         certPool,
		Intermediates: intermediates,
		DNSName:       domain,
	}

	if _, err := chain[0].Verify(opts); err != nil {
		//fmt.Println(err)
		return false, fmt.Errorf("certificate verify failed: %v", err)
	}

	return true, nil
}

func verifyHostname(cert *x509.Certificate, domain string) bool {
	return cert.VerifyHostname(domain) == nil
}

// Ref to: https://github.com/izolight/certigo/blob/v1.10.0/lib/encoder.go#L445
func IsSelfSigned(cert *x509.Certificate) bool {
	if bytes.Equal(cert.RawIssuer, cert.RawSubject) {
		return true
	} //12.25
	return cert.CheckSignatureFrom(cert) == nil
}

// Ref to: https://github.com/google/certificate-transparency-go/blob/master/ctutil/sctcheck/sctcheck.go
func isChainInOrder(chain []*x509.Certificate) string {
	// var issuer *x509.Certificate
	leaf := chain[0]
	for i := 1; i < len(chain); i++ {
		c := chain[i]
		if bytes.Equal(c.RawSubject, leaf.RawIssuer) && c.CheckSignature(leaf.SignatureAlgorithm, leaf.RawTBSCertificate, leaf.Signature) == nil {
			// issuer = c
			if i > 1 {
				return "not"
			}
			break
		}
	}
	if len(chain) < 1 {
		return "single"
	}
	return "yes"
}

var algoName = [...]string{
	x509.MD2WithRSA:      "MD2-RSA",
	x509.MD5WithRSA:      "MD5-RSA",
	x509.SHA1WithRSA:     "SHA1-RSA",
	x509.SHA256WithRSA:   "SHA256-RSA",
	x509.SHA384WithRSA:   "SHA384-RSA",
	x509.SHA512WithRSA:   "SHA512-RSA",
	x509.DSAWithSHA1:     "DSA-SHA1",
	x509.DSAWithSHA256:   "DSA-SHA256",
	x509.ECDSAWithSHA1:   "ECDSA-SHA1",
	x509.ECDSAWithSHA256: "ECDSA-SHA256",
	x509.ECDSAWithSHA384: "ECDSA-SHA384",
	x509.ECDSAWithSHA512: "ECDSA-SHA512",
}

var badSignatureAlgorithms = [...]x509.SignatureAlgorithm{
	x509.MD2WithRSA,
	x509.MD5WithRSA,
	x509.SHA1WithRSA,
	x509.DSAWithSHA1,
	x509.ECDSAWithSHA1,
}

func algWarnings(cert *x509.Certificate) (warning string) {
	alg, size := decodeKey(cert.PublicKey)
	if (alg == "RSA" || alg == "DSA") && size < 2048 {
		// warnings = append(warnings, fmt.Sprintf("Size of %s key should be at least 2048 bits", alg))
		warning = fmt.Sprintf("Size of %s key should be at least 2048 bits", alg)
	}
	if alg == "ECDSA" && size < 224 {
		warning = fmt.Sprintf("Size of %s key should be at least 224 bits", alg)
	}

	for _, alg := range badSignatureAlgorithms {
		if cert.SignatureAlgorithm == alg {
			warning = fmt.Sprintf("Signed with %s, which is an outdated signature algorithm", algString(alg))
		}
	}

	if alg == "RSA" {
		key := cert.PublicKey.(*rsa.PublicKey)
		if key.E < 3 {
			warning = "Public key exponent in RSA key is less than 3"
		}
		if key.N.Sign() != 1 {
			warning = "Public key modulus in RSA key appears to be zero/negative"
		}
	}

	return warning
}

// decodeKey returns the algorithm and key size for a public key.
func decodeKey(publicKey interface{}) (string, int) {
	switch publicKey.(type) {
	case *dsa.PublicKey:
		return "DSA", publicKey.(*dsa.PublicKey).P.BitLen()
	case *ecdsa.PublicKey:
		return "ECDSA", publicKey.(*ecdsa.PublicKey).Curve.Params().BitSize
	case *rsa.PublicKey:
		return "RSA", publicKey.(*rsa.PublicKey).N.BitLen()
	default:
		return "", 0
	}
}

func algString(algo x509.SignatureAlgorithm) string {
	if 0 < algo && int(algo) < len(algoName) {
		return algoName[algo]
	}
	return strconv.Itoa(int(algo))
}

func calculatePortScores(config string) (map[string]int, []PortUsageDetail) { //增加了一个返回参数[]PortUsageDetail
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

	var portsUsage []PortUsageDetail
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
			portsUsage = append(portsUsage, PortUsageDetail{
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

func calculateCertScores(cert CertInfo) map[string]int {
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

// 生成 CSV 文件，zgrab2 从该文件读取输入
func writeCSV(hostname string) (string, error) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "zgrab2_*.csv") // 生成唯一的临时文件
	if err != nil {
		return "", fmt.Errorf("无法创建临时文件: %v", err)
	}
	defer tmpFile.Close() // 关闭文件
	fmt.Printf("CSV 文件已创建: %s\n", tmpFile.Name())
	// 写入 CSV 内容
	writer := csv.NewWriter(tmpFile)
	defer writer.Flush() // 确保数据写入磁盘

	// // zgrab2 需要的 CSV 结构，通常包含 "ip" 或 "domain" 列
	// err = writer.Write([]string{"domain"}) // 设置 CSV 头部
	// if err != nil {
	// 	return "", fmt.Errorf("写入 CSV 头部失败: %v", err)
	// }
	//会额外读取domain为domain的，所以删去

	err = writer.Write([]string{hostname}) // 写入数据
	if err != nil {
		return "", fmt.Errorf("写入 CSV 数据失败: %v", err)
	}

	// 返回文件路径
	return tmpFile.Name(), nil
}

// func RunZGrab2(protocoltype, hostname, port string, tlsMode string) (bool, error) {
// 	// 生成临时 CSV 文件
// 	csvfile, err := writeCSV(hostname)
// 	if err != nil {
// 		fmt.Println("CSV 文件未创建")
// 		return false, fmt.Errorf("fail to create csv file: %v", err)
// 	}
// 	defer os.Remove(csvfile) // 开发阶段保留也可以先注释掉这行

// 	// 构造命令参数
// 	args := []string{protocoltype, "--port", port, "-f", csvfile}
// 	if tlsMode == "starttls" {
// 		args = append(args, "--starttls")
// 	} else if tlsMode == "tls" {
// 		// 注意协议后缀变化
// 		tlsFlag := fmt.Sprintf("--%ss", protocoltype)
// 		args = append(args, tlsFlag)
// 	}

// 	// 执行 zgrab2 命令
// 	zgrabPath := "./zgrab2"
// 	cmd := exec.Command(zgrabPath, args...)

// 	var out bytes.Buffer
// 	var stderr bytes.Buffer
// 	cmd.Stdout = &out
// 	cmd.Stderr = &stderr

// 	fmt.Println("Running command:", cmd.String())

// 	err = cmd.Run()
// 	if err != nil {
// 		fmt.Println("Error running command:", err)
// 		fmt.Println("Stderr:", stderr.String())
// 		fmt.Println("Stdout:", out.String())
// 		return false, fmt.Errorf("zgrab2 执行失败: %v\nStderr: %s", err, stderr.String())
// 	}

// 	fmt.Println("Raw JSON Output:", out.String())

// 	var result map[string]interface{}
// 	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
// 		return false, fmt.Errorf("解析 JSON 失败: %v", err)
// 	}

//		return checkConnectSuccess(result, protocoltype)
//	}//4.22

func RunZGrab2(protocol, hostname, port, mode string) (bool, error) { //4.22python
	fmt.Print("Running test\n")

	// pythonPath := "python3" //linux
	pythonPath := "python" //windows
	scriptPath := "./tlscheck/test_tls.py"

	cmd := exec.Command(pythonPath, scriptPath,
		"--protocol", protocol,
		"--host", hostname,
		"--port", port,
		"--mode", mode,
	)

	output, err := cmd.CombinedOutput()
	fmt.Println("Raw output:\n", string(output))

	if err != nil {
		return false, fmt.Errorf("execution error: %v, output: %s", err, string(output))
	}

	// 定义结构体用于解析
	var result ConnectInfo
	if err := json.Unmarshal(output, &result); err != nil {
		return false, fmt.Errorf("invalid JSON output: %v", err)
	}

	if !result.Success {
		return false, fmt.Errorf("TLS test failed: %s", result.Error)
	}

	// 打印信息
	fmt.Printf("TLS Version: %s\n", result.Info.Version)
	fmt.Printf("Cipher: %v\n", result.Info.Cipher)
	fmt.Printf("TLS CA: %.40s...\n", result.Info.TLSCA) // 只显示前40字符
	//fmt.Printf("Auth: %v\n", result.Info.Auth)//因为smtp/imap的该字段结构不一样，所以没有打印

	return true, nil
}

// 8.12
func RunZGrab2WithResult(protocol, hostname, port, mode string) (bool, *ConnectInfo, error) {
	fmt.Print("Running test\n")

	// pythonPath := "python3" //linux
	pythonPath := "python" //windows
	scriptPath := "./tlscheck/test_tls.py"

	cmd := exec.Command(pythonPath, scriptPath,
		"--protocol", protocol,
		"--host", hostname,
		"--port", port,
		"--mode", mode,
	)

	output, err := cmd.CombinedOutput()
	fmt.Println("Raw output:\n", string(output))

	if err != nil {
		return false, nil, fmt.Errorf("execution error: %v, output: %s", err, string(output))
	}
	var result ConnectInfo
	if err := json.Unmarshal(output, &result); err != nil {
		return false, nil, fmt.Errorf("invalid JSON output: %v", err)
	}

	if !result.Success {
		return false, nil, fmt.Errorf("TLS test failed: %s", result.Error)
	}

	return true, &result, nil
}

func checkConnectSuccess(result map[string]interface{}, protoType string) (bool, error) {
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("data 字段缺失或类型错误")
	}

	protoData, ok := data[protoType].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("%s 字段缺失或类型错误", protoType)
	}

	status, ok := protoData["status"].(string)
	if !ok {
		return false, fmt.Errorf("status 字段缺失或类型错误")
	}

	// 若连接成功
	if status == "success" {
		return true, nil
	}

	// 如果包含 error 字段，提取错误信息
	if errStr, exists := protoData["error"].(string); exists {
		// 检查是否为 no such host 错误
		if strings.Contains(errStr, "no such host") {
			return false, fmt.Errorf("fail to connect: no such host")
		}
		return false, fmt.Errorf("fail to connect: %s", errStr)
	}

	return false, fmt.Errorf("fail to connect, status: %s", status)
}

func isNoSuchHostError(err error) bool { //4.22Go->python
	// return err != nil && strings.Contains(err.Error(), "no such host")
	return err != nil && strings.Contains(err.Error(), "Name or service not known")
}

func calculateConnectScores(config string) (map[string]interface{}, []ConnectDetail) {
	var allConnectionDetails []ConnectDetail
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
		canConnectPlain, plainInfo, errPlain := RunZGrab2WithResult(protocolType, hostname, port, "plain")
		canConnectStartTLS, starttlsInfo, errStartTLS := RunZGrab2WithResult(protocolType, hostname, port, "starttls")
		canConnectTLS, tlsInfo, errTLS := RunZGrab2WithResult(protocolType, hostname, port, "ssl")
		// 将结果附加
		if plainInfo == nil {
			plainInfo = &ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if starttlsInfo == nil {
			starttlsInfo = &ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if tlsInfo == nil {
			tlsInfo = &ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}

		detail := ConnectDetail{
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

		if isNoSuchHostError(errPlain) || isNoSuchHostError(errStartTLS) || isNoSuchHostError(errTLS) {
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
	overall := (tlsScore*3 + starttlsScore*2 + plainScore*1) / 6

	grade := "F"
	switch {
	case (tlsScore == 100 || starttlsScore == 100) && plainScore == 0:
		grade = "A+"
	case tlsScore >= 80 || starttlsScore >= 80:
		grade = "A"
	case tlsScore >= 50 || starttlsScore >= 50:
		grade = "B"
	case plainScore >= 50:
		grade = "C"
	case tlsScore == 0 && starttlsScore == 0 && plainScore == 0:
		grade = "F"
	}

	scores["TLS_Connections"] = tlsScore
	scores["Plaintext_Connections"] = plainScore
	scores["STARTTLS_Connections"] = starttlsScore
	scores["Overall_Connection_Score"] = overall
	scores["Connection_Grade"] = grade
	scores["timestamp"] = time.Now().Format(time.RFC3339)
	scores["warnings"] = warnings

	return scores, allConnectionDetails
}

func scoreConfig(config string, certInfo CertInfo) (map[string]int, map[string]interface{}, []ConnectDetail, []PortUsageDetail) {
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

func scoreConfig_Autoconfig(config string, certInfo CertInfo) (map[string]int, map[string]interface{}, []ConnectDetail, []PortUsageDetail) {
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

func calculatePortScores_Autoconfig(config string) (map[string]int, []PortUsageDetail) {
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
	var portsUsage []PortUsageDetail
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
			portsUsage = append(portsUsage, PortUsageDetail{
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
			portsUsage = append(portsUsage, PortUsageDetail{
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

func calculateConnectScores_Autoconfig(config string) (map[string]interface{}, []ConnectDetail) {
	var allConnectionDetails []ConnectDetail
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
		canConnectPlain, plainInfo, errPlain := RunZGrab2WithResult(protocolType, hostname, port, "plain")
		canConnectStartTLS, starttlsInfo, errStartTLS := RunZGrab2WithResult(protocolType, hostname, port, "starttls")
		canConnectTLS, tlsInfo, errTLS := RunZGrab2WithResult(protocolType, hostname, port, "ssl")
		// 将结果附加
		if plainInfo == nil {
			plainInfo = &ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if starttlsInfo == nil {
			starttlsInfo = &ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if tlsInfo == nil {
			tlsInfo = &ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}

		detail := ConnectDetail{
			Type:     protocolType,
			Host:     hostname,
			Port:     port,
			Plain:    *plainInfo,
			StartTLS: *starttlsInfo,
			TLS:      *tlsInfo,
		}
		allConnectionDetails = append(allConnectionDetails, detail)

		if isNoSuchHostError(errPlain) || isNoSuchHostError(errStartTLS) || isNoSuchHostError(errTLS) {
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
		canConnectPlain, plainInfo, errPlain := RunZGrab2WithResult(protocolType, hostname, port, "plain")
		canConnectStartTLS, starttlsInfo, errStartTLS := RunZGrab2WithResult(protocolType, hostname, port, "starttls")
		canConnectTLS, tlsInfo, errTLS := RunZGrab2WithResult(protocolType, hostname, port, "ssl")
		// 将结果附加
		if plainInfo == nil {
			plainInfo = &ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if starttlsInfo == nil {
			starttlsInfo = &ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if tlsInfo == nil {
			tlsInfo = &ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}

		detail := ConnectDetail{
			Type:     protocolType,
			Host:     hostname,
			Port:     port,
			Plain:    *plainInfo,
			StartTLS: *starttlsInfo,
			TLS:      *tlsInfo,
		}
		allConnectionDetails = append(allConnectionDetails, detail)
		if isNoSuchHostError(errPlain) || isNoSuchHostError(errStartTLS) || isNoSuchHostError(errTLS) {
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
	overall := (tlsScore*3 + starttlsScore*2 + plainScore*1) / 6

	//等级判断逻辑
	grade := "F"
	switch {
	case (tlsScore == 100 || starttlsScore == 100) && plainScore == 0:
		grade = "A+"
	case tlsScore >= 80 || starttlsScore >= 80:
		grade = "A"
	case tlsScore >= 50 || starttlsScore >= 50:
		grade = "B"
	case plainScore >= 50:
		grade = "C"
	case tlsScore == 0 && starttlsScore == 0 && plainScore == 0:
		grade = "F"
	}
	scores["TLS_Connections"] = tlsScore
	scores["Plaintext_Connections"] = plainScore
	scores["STARTTLS_Connections"] = starttlsScore
	scores["Overall_Connection_Score"] = overall
	scores["Connection_Grade"] = grade
	scores["timestamp"] = time.Now().Format(time.RFC3339)
	scores["warnings"] = warnings
	return scores, allConnectionDetails
}

func scoreConfig_SRV(result SRVResult) (map[string]int, map[string]interface{}, []ConnectDetail, []PortUsageDetail) {
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

func calculatePortScores_SRV(result SRVResult) (map[string]int, []PortUsageDetail) {
	scores := make(map[string]int)

	securePorts := map[string]bool{}
	insecurePorts := map[string]bool{}
	nonStandardPorts := map[string]bool{}

	standardEncrypted := map[uint16]bool{993: true, 995: true, 465: true}
	standardInsecure := map[uint16]bool{143: true, 110: true, 25: true, 587: true}
	var portsUsage []PortUsageDetail
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
		portsUsage = append(portsUsage, PortUsageDetail{
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

func Identify_Port_Status(record SRVRecord) string {
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

func calculateDNSSECScores_SRV(result SRVResult) int {
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

func calculateConnectScores_SRV(result SRVResult) (map[string]interface{}, []ConnectDetail) {
	var allConnectionDetails []ConnectDetail
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
		canConnectPlain, plainInfo, errPlain := RunZGrab2WithResult(protocol, hostname, portStr, "plain")
		canConnectStartTLS, starttlsInfo, errStartTLS := RunZGrab2WithResult(protocol, hostname, portStr, "starttls")
		canConnectTLS, tlsInfo, errTLS := RunZGrab2WithResult(protocol, hostname, portStr, "ssl")
		// 将结果附加
		if plainInfo == nil {
			plainInfo = &ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if starttlsInfo == nil {
			starttlsInfo = &ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if tlsInfo == nil {
			tlsInfo = &ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}

		detail := ConnectDetail{
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

		if isNoSuchHostError(errPlain) || isNoSuchHostError(errStartTLS) || isNoSuchHostError(errTLS) {
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
	overall := (tlsScore*3 + starttlsScore*2 + plainScore*1) / 6

	grade := "F"
	switch {
	case (tlsScore == 100 || starttlsScore == 100) && plainScore == 0:
		grade = "A+"
	case tlsScore >= 80 || starttlsScore >= 80:
		grade = "A"
	case tlsScore >= 50 || starttlsScore >= 50:
		grade = "B"
	case plainScore >= 50:
		grade = "C"
	case tlsScore == 0 && starttlsScore == 0 && plainScore == 0:
		grade = "F"
	}

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

// 添加防御能力评估模块 6.17
// func evaluateSecurityDefense(uri string, certInfo CertInfo, config string,  dnsRecord *DNSRecord, score map[string]int, connectScores map[string]interface{}) map[string]int {
func evaluateSecurityDefense(uri string, certInfo CertInfo, config string, score map[string]int, connectScores map[string]interface{}) map[string]int {
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

func evaluateTamperingDefense(uri string, certInfo CertInfo) int {
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

func evaluateCertificateForgery(certInfo CertInfo) int {
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

func evaluateDNSHijack(dnsRecord *DNSRecord) int {
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

func evaluateSecurityDefense_SRV(dnsrecord *DNSRecord, score map[string]int, connectScores map[string]interface{}) map[string]int {
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

func scoreConfig_Guess(guessed []string) (map[string]interface{}, []ConnectDetail, []PortUsageDetail) {

	connectScores, ConnectDetails, portsUsage := calculateConnectScores_Guess(guessed)
	return connectScores, ConnectDetails, portsUsage

}

func calculateConnectScores_Guess(guessed []string) (map[string]interface{}, []ConnectDetail, []PortUsageDetail) {
	var allConnectionDetails []ConnectDetail
	scores := make(map[string]interface{})
	var portsUsage []PortUsageDetail
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
		portsUsage = append(portsUsage, PortUsageDetail{
			Protocol: protocolType,
			Port:     portStr,
			Status:   status,
			Host:     hostname,
		})
		totalProtocols++
		// canConnectPlain, errPlain := RunZGrab2(protocolType, hostname, portStr, "plain") //
		// canConnectStartTLS, errStartTLS := RunZGrab2(protocolType, hostname, portStr, "starttls")
		// canConnectTLS, errTLS := RunZGrab2(protocolType, hostname, portStr, "ssl") //
		canConnectPlain, plainInfo, errPlain := RunZGrab2WithResult(protocolType, hostname, portStr, "plain")
		canConnectStartTLS, starttlsInfo, errStartTLS := RunZGrab2WithResult(protocolType, hostname, portStr, "starttls")
		canConnectTLS, tlsInfo, errTLS := RunZGrab2WithResult(protocolType, hostname, portStr, "ssl")
		// 将结果附加
		if plainInfo == nil {
			plainInfo = &ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if starttlsInfo == nil {
			starttlsInfo = &ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}
		if tlsInfo == nil {
			tlsInfo = &ConnectInfo{
				Success: false,
				Error:   "nil result from zgrab2",
			}
		}

		detail := ConnectDetail{
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
		if isNoSuchHostError(errPlain) || isNoSuchHostError(errStartTLS) || isNoSuchHostError(errTLS) {
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
	overall := (tlsScore*3 + starttlsScore*2 + plainScore*1) / 6

	//等级判断逻辑
	grade := "F"
	switch {
	case (tlsScore == 100 || starttlsScore == 100) && plainScore == 0:
		grade = "A+"
	case tlsScore >= 80 || starttlsScore >= 80:
		grade = "A"
	case tlsScore >= 50 || starttlsScore >= 50:
		grade = "B"
	case plainScore >= 50:
		grade = "C"
	case tlsScore == 0 && starttlsScore == 0 && plainScore == 0:
		grade = "F"
	}
	scores["TLS_Connections"] = tlsScore
	scores["Plaintext_Connections"] = plainScore
	scores["STARTTLS_Connections"] = starttlsScore
	scores["Overall_Connection_Score"] = overall
	scores["Connection_Grade"] = grade
	scores["timestamp"] = time.Now().Format(time.RFC3339)
	scores["warnings"] = warnings
	return scores, allConnectionDetails, portsUsage

}
