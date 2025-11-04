import React, { useState, useEffect } from "react";
import {
    ScoreBar,
    renderScoreBar,
    renderConnectionDetail,
    getAutodiscoverRecommendations,
    DefenseRadarChart,
    getSRVRecommendations,
    getCertRecommendations,
} from "../utils/renderHelper";
import { PeculiarCertificateViewer } from '@peculiar/certificates-viewer-react';
import LinearProgress from '@mui/material/LinearProgress';
import "../App.css";
import { checkInternalDiff } from "../components/checkInternalDiff";//9.11
import TlsAnalyzerPanel from "../components/TlsAnalyzerPanel";
import { AiOutlineFileText } from "react-icons/ai";

function MainPage() {
    const [email, setEmail] = useState("");
    const [results, setResults] = useState({});
    const [errors, setErrors] = useState("");
    const [loading, setLoading] = useState(false);
    const [recentlySeen, setRecentlySeen] = useState([]);
    const [showRawConfig, setShowRawConfig] = useState({});
    const [showRawCertsMap, setShowRawCertsMap] = useState({});
    const [activeCertIdxMap, setActiveCertIdxMap] = useState({});
    const [showCertChainMap, setShowCertChainMap] = useState({});
    const [showAnalysis, setShowAnalysis] = useState({});
    const [activeTab, setActiveTab] = useState({});
    const [progress, setProgress] = useState(0);
    const [liveLogs, setLiveLogs] = useState([]); 
    const [liveResult, setLiveResult] = useState(null); 
    const [testingHost, setTestingHost] = useState(null);

    const [stage, setStage] = useState("");
    const [progressMessage, setProgressMessage] = useState("");
    const [activeDomain, setActiveDomain] = useState(null);//10.19

    const mechanisms = ["autodiscover", "autoconfig", "srv", "guess", "overview"];//9.10_2 新增加比较机制供管理者一眼看出不同机制得到的配置信息有何不同
    // 默认选中第一个有结果的机制（不含 compare）9.10_2
    const firstAvailable = mechanisms.find(m => m !== "overview" && results[m]) || mechanisms[0];
    const [currentMech, setCurrentMech] = useState(firstAvailable);
    const [lastSubmittedEmail, setLastSubmittedEmail] = useState("");

    const [recommendedDomains, setRecommendedDomains] = useState([]);
    const [checkAllResult, setCheckAllResult] = useState(null); // 保存检测结果或缓存结果
    const [hasAnyResultState, setHasAnyResult] = useState(false);    // 控制结果区是否渲染

    const [isRecommendedClick, setIsRecommendedClick] = useState(false); //10.30

    // 响应式判断10.30
    const [windowWidth, setWindowWidth] = useState(window.innerWidth);
    useEffect(() => {
        const handleResize = () => setWindowWidth(window.innerWidth);
        window.addEventListener("resize", handleResize);
        return () => window.removeEventListener("resize", handleResize);
    }, []);
    const isMobile = windowWidth < 600;
    const isTablet = windowWidth >= 600 && windowWidth < 900;

    useEffect(() => {
    fetch("/api/recommended")
        .then(res => res.json())
        .then(data => setRecommendedDomains(data))
        .catch(err => console.error(err));
    }, []);
    //9.18_2
    const handleClickRecommended = (domain) => {
        // 直接使用缓存数据，如果有的话
        const cached = recommendedDomains.find(d => d.domain === domain);
        if (cached && cached.response) {
            setCheckAllResult(cached.response); // 直接填充缓存结果
            setHasAnyResult(true);
        } else {
          // 如果没有缓存，也可以去后端重新查询 /checkAll?email=xxx@domain
          const email = `user@${domain}`; // 可以默认用 user@domain
          fetch(`/checkAll?email=${encodeURIComponent(email)}`)
            .then(res => res.json())
            .then(data => {
              setCheckAllResult(data);
              setHasAnyResult(true);
            })
            .catch(err => console.error(err));
        }
      };

    //9.23
    // 预设推荐域名
    const defaultRecommended = [
        { domain: "qq.com" },
        { domain: "outlook.com" },
        { domain: "gmail.com" },
        { domain: "yandex.com" },
        { domain: "126.com" },
        {domain:"yahoo.com"},
        {domain:"terra.com.ar"},
        {domain:"zohu.com"}
    ];
    // const [recommended, setRecommended] = useState([]);
    const [recommended, setRecommended] = useState(defaultRecommended);

    // useEffect(() => {
    // fetch("/api/recommended")
    //     .then(res => res.json())
    //     .then(data => setRecommended(data))
    //     .catch(err => console.error("加载推荐域名失败:", err));
    // }, []);
    useEffect(() => {
        fetch("/api/recommended")
          .then(res => res.json())
          .then(data => {
            // 如果后端返回有数据，就覆盖预设
            if (data && data.length > 0) {
              setRecommended(data);
            }
          })
          .catch(err => {
            console.error("加载推荐域名失败:", err);
            // 如果失败，保持默认数组，不影响UI
          });
      }, []);



    //9.15 改造输入框
    const [suggestions, setSuggestions] = useState([]);

    const presetDomains = [
    "qq.com",
    "126.com",
    "gmail.com",
    "yandex.com",
    "outlook.com"
    ];

    const handleChange = (e) => {
        const value = e.target.value;
        setEmail(value);

        // 如果输入包含 @ 且不是结尾，就提示
        const atIndex = value.indexOf("@");
        if (atIndex !== -1 && atIndex === value.length - 1) {
            const username = value.slice(0, atIndex);
            const newSuggestions = presetDomains.map((d) => `${username}@${d}`);
            setSuggestions(newSuggestions);
        } else {
            setSuggestions([]);
        }
    };

    const handleSelect = (suggestion) => {
        setEmail(suggestion);
        setSuggestions([]);
    };
    
    // // 点击检测按钮9.23原
    // const handleClick = () => {
    //     const targetEmail = email.trim();
    //     if (!targetEmail) return; // 空输入不处理
    
    //     handleSearch(targetEmail); // 调你原来的检测函数
    //     setLastSubmittedEmail(targetEmail); // 保存用户名用于展示
    // };

    //9.23改后
    // const handleClick = (e, customEmail) => {
    //     const targetEmail = customEmail || email.trim();
    //     if (!targetEmail) return;
    //     handleSearch(targetEmail); // 调你原来的检测函数
    //     setLastSubmittedEmail(targetEmail); // 保存用户名用于展示
    // };
    //10.30
    const handleClick = (e, customEmail, isRecommended = false) => {
        const targetEmail = customEmail || email.trim();
        if (!targetEmail) return;
    
        handleSearch(targetEmail); // 调你原来的检测函数
        setLastSubmittedEmail(targetEmail); // 保存用户名用于展示
        setIsRecommendedClick(isRecommended); // 标记来源
    };
    



    const certLabelMap = {
        IsTrusted: "是否可信",
        VerifyError: "验证错误",
        IsHostnameMatch: "域名是否匹配",
        IsInOrder: "链顺序是否正确",
        IsExpired: "是否过期",
        IsSelfSigned: "是否自签名",
        SignatureAlg: "签名算法",
        AlgWarning: "算法警告",
        TLSVersion: "TLS版本",
        Subject: "证书主体",
        Issuer: "签发机构"
    };

    const dnsFieldMap = {
        domain: "域名",
        SOA: "SOA 记录",
        NS: "NS 记录",
        ADbit_imap: "IMAP ADBit",
        ADbit_imaps: "IMAPS ADBit",
        ADbit_pop3: "POP3 ADBit",
        ADbit_pop3s: "POP3S ADBit",
        ADbit_smtp: "SMTP ADBit",
        ADbit_smtps: "SMTPS ADBit"
    };
    
    const tabLabelMap = {
        score: "评分",
        recommend: "建议",
        radar: "防御雷达图"
    };

    const spinnerStyle = {
        border: "4px solid #f3f3f3",
        borderTop: "4px solid #799cc8ff",
        borderRadius: "50%",
        width: "50px",
        height: "50px",
        animation: "spin 1s linear infinite",
        margin: "0 auto"
    };




    useEffect(() => {
        fetchRecent();
    }, []);

    useEffect(() => {
        // 当 results 更新时，自动切换到第一个有结果机制
        const first = mechanisms.find(m => results[m]);
        if (first) setCurrentMech(first);
    }, [results]);

    const fetchRecent = () => {
        fetch(`/api/recent`)
            .then((res) => res.json())
            .then((data) => setRecentlySeen(data))
            .catch((err) => console.error("Failed to fetch recent scans:", err));
    };

    const handleSearch = async (targetEmail) => {  //9.10
        // if (!email) {
        //     setErrors("请输入邮箱地址");
        //     return;
        // }
        const finalEmail = targetEmail || email.trim();
        if (!finalEmail) {        
            setErrors("请输入邮箱地址");
            return;
        }

        setErrors("");
        setLoading(true);
        setProgress(0);
        setStage("开始检测");
        setProgressMessage("");

        // const ws = new WebSocket("ws://localhost:8081/ws/checkall-progress");
         // ✅ WebSocket 改成相对当前域名
        const wsProtocol = window.location.protocol === "https:" ? "wss" : "ws";
        const ws = new WebSocket(`${wsProtocol}://${window.location.host}/ws/checkall-progress`);

        ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            if (data.type === "progress") {
                setProgress(data.progress);
                setStage(data.stage);
                setProgressMessage(data.message);

                // ✅ 检测完成时关闭 WS
                if (data.progress === 100) {
                    ws.close();
                }
            }
        };

        ws.onerror = () => {
            console.error("WebSocket 连接失败");
        };

        ws.onclose = () => {
            console.log("进度 WebSocket 已关闭");
        };

        try {
            const response = await fetch(`/checkAll?email=${finalEmail}`); //9.10
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const result = await response.json();
            setResults(result);
        } catch (err) {
            console.error(err);
            setErrors("检测失败，请重试");
        } finally {
            setLoading(false);
            // ❌ 不要在这里关闭 WS，否则进度还没推完就断掉
        }
    };

    const toggleRaw = (mech) => {
        setShowRawConfig((prev) => ({ ...prev, [mech]: !prev[mech] }));
    };

    const toggleRawCerts = (mech) => {
        setShowRawCertsMap((prev) => ({ ...prev, [mech]: !prev[mech] }));
    };

    const toggleCertChain = (mech) => {
        setShowCertChainMap((prev) => ({ ...prev, [mech]: !prev[mech] }));
    };

    const setActiveCertIdx = (mech, idx) => {
        setActiveCertIdxMap((prev) => ({ ...prev, [mech]: idx }));
    };

    const toggleAnalysis = (mech) => {
        setShowAnalysis((prev) => ({ ...prev, [mech]: !prev[mech] }));
    };

    const changeTab = (mech, tabName) => {
        setActiveTab((prev) => ({ ...prev, [mech]: tabName }));
    };

    // 在组件里定义一个通用函数
    const handleViewDetailsClick = async (mechType, details) => {
        try {
            const res = await fetch("/store-temp-data", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ details }),
            });

            if (!res.ok) throw new Error("Failed to store data");

            const { id } = await res.json();
            const newTab = window.open(`/config-view?uri=${mechType}_results&id=${id}`, "_blank");
            if (!newTab) alert("⚠️ 请允许浏览器弹出窗口。");
        } catch (err) {
            console.error(`❌ Error storing ${mechType} detail:`, err);
            alert(`❌ 无法打开连接详情（${mechType.toUpperCase()}）页面。`);
        }
    };

    //10.13
    const prefixAppMap = [
        {
        app: "Thunderbird",
        prefixes: ["imap.", "mail.", "pop3.", "pop.", "smtp."],
        },
        {
        app: "Outlook",
        prefixes: ["imap.", "mail.", "pop.", "smtp."],
        },
        {
        app: "FairEmail",
        prefixes: ["mx.", "imaps.", "smtps."],
        },
        {
        app: "The bai!",
        prefixes: ["imap4."],
        },
    ];

    function detectMailAppsSmart(hostname) {
        const matches = prefixAppMap
        .filter(({ prefixes }) => prefixes.some(pre => hostname.startsWith(pre)))
        .map(({ app }) => app);
    
        if (matches.length === 0) return "未知来源";
    
        // 随机挑选两个或一个（真实匹配中随机取样）
        if (matches.length > 2) {
        const shuffled = matches.sort(() => 0.5 - Math.random());
        return shuffled.slice(0, 2).join(" / ");
        }
    
        return matches.join(" / ");
    }


    // 公用按钮样式
    const viewButtonStyle = {
        display: "inline-block",
        marginTop: "0.3rem",
        backgroundColor: "#81b5d5ff", 
        color: "white",
        padding: "8px 14px",
        textDecoration: "none",
        borderRadius: "6px",
        cursor: "pointer",
        fontSize: "14px",
        border: "none",
        transition: "background 0.3s ease"
    };

    const viewButtonHoverStyle = {
        padding: "12px 24px", // 增加内边距
        fontSize: "16px", // 字体更大
        fontWeight: "bold", // 字体加粗
        backgroundColor: "#2980b9", // 主色调（蓝色）
        color: "#fff", // 白色文字
        border: "none",
        borderRadius: "8px", // 圆角更大
        cursor: "pointer",
        margin: "10px 0",
        transition: "background-color 0.3s ease", // 平滑过渡
    };

    // hover 效果（用 className 或内联处理）
    const viewButtonHover = {
        backgroundColor: "#219150" // 稍深的绿色
    };

    const tabStyle = (mech) => ({
        cursor: "pointer",
        padding: "10px 20px",
        borderBottom: mech === currentMech ? "3px solid #3498db" : "3px solid transparent",
        color: mech === currentMech ? "#3498db" : "#666",
        fontWeight: mech === currentMech ? "bold" : "normal",
        userSelect: "none",
        marginRight: "10px",
        fontSize: "16px"
    });

    const thStyle = {
        padding: "12px",
        borderBottom: "2px solid #ceddebff",
        textAlign: "center",
        fontSize: "15px",
        backgroundColor: "#d3ebf1ff",
        color: "#333",
        fontWeight: 500,
        height: "40px",
        whiteSpace: "nowrap" // 表头文字不换行
    };

    const tdStyle = {
        padding: "10px",
        borderBottom: "1px solid #759dc2ff",
        textAlign: "center",
        fontSize: "14px",
        color: "#333",
        maxWidth: "200px",     // 限制宽度
        wordBreak: "break-word" // 自动换行
    };


    const tableStyle = {
        width: "95%",
        margin: "15px auto",
        borderCollapse: "collapse",
        borderRadius: "10px",
        overflow: "hidden",
        background: "rgba(255, 255, 255, 0.92)", // ✅ 表格内部白色填充（略透明）
        boxShadow: "0 3px 8px rgba(0,0,0,0.08)", // ✅ 柔和阴影
    };
    

    //9.10_2 完善compare具体内容
    const normalizeAuto = (mech, results) => {  //此处可参照后端函数calculatePortScores
        const ports = results[mech]?.score_detail?.ports_usage || [];
        return ports.map(item => {
            let ssl = item.ssl;
            // Autodiscover 里可能是 "on" / "off"
            if (ssl === "on"||ssl === "ssl"||ssl==="tls") ssl = "SSL";
            if (ssl === "off") ssl = "PLAIN";
            return {
                protocol: item.protocol?.toUpperCase() || "",
                port: item.port,
                host: item.host,
                ssl: ssl?.toUpperCase() || ""
            };
        });
    };

    const normalizeSrv = (results) => {
        const srv = results.srv?.srv_records || {};
        const all = [...(srv.recv || []), ...(srv.send || [])];
            return all.map(item => {
            const service = item.Service || "";
            // 例如 "_imaps._tcp.yandex.com"
            const protoMatch = service.match(/^_([a-z]+)/i);
            let protocol = "";
            let ssl = "";
            if (protoMatch) {
                const proto = protoMatch[1].toLowerCase();
                if (proto.startsWith("imap")) protocol = "IMAP";
                else if (proto.startsWith("pop")) protocol = "POP3";
                else if (proto.startsWith("submission")) protocol = "SMTP";
                // SSL 类型
                ssl = proto.endsWith("s") ? "SSL" : "STARTTLS"; //此处存疑TODO
            }
            return {
                protocol,
                port: item.Port,
                host: (item.Target || "").replace(/\.$/, ""), // 去掉末尾点
                ssl
            };
        });
    };
    
    const comparePortsUsage = (results) => {
        const mechList = ["autodiscover", "autoconfig", "srv"];

        const allNormalized = {
            autodiscover: normalizeAuto("autodiscover", results),
            autoconfig: normalizeAuto("autoconfig", results),
            srv: normalizeSrv(results)
        };

        const comparisonMap = {}; // key = protocol-port

        mechList.forEach(mech => {
            allNormalized[mech].forEach(item => {
                const key = `${item.protocol}-${item.port}`;
                if (!comparisonMap[key]) comparisonMap[key] = {};
                comparisonMap[key][mech] = item;
            });
        });

        return comparisonMap;
    };


      //9.11
    // 评级函数：根据分数给 A/B/C...
    const getGrade = (score) => {
        if (score >= 90) return "A";
        if (score >= 50) return "B";
        if (score >= 30) return "C";
        return "F";
    };

    // 简单 Accordion 组件
    const Accordion = ({ title, children }) => {
        const [open, setOpen] = React.useState(false);
        return (
            <div style={{ marginBottom: "10px" }}>
                <div
                    style={{
                        cursor: "pointer",
                        fontWeight: "bold",
                        padding: "6px 10px",
                        background: "#f7f7f7",
                        border: "1px solid #ddd",
                        borderRadius: "6px",
                    }}
                    onClick={() => setOpen(!open)}
                >
                    {title} {open ? "▲" : "▼"}
                </div>
                {open && (
                    <div
                        style={{
                            padding: "10px",
                            border: "1px solid #ddd",
                            borderTop: "none",
                            background: "#fafafa",
                            borderRadius: "0 0 6px 6px",
                        }}
                    >
                        {children}
                    </div>
                )}
            </div>
        );
    };

    const preprocessResults = (results) => {  //9.11
        Object.entries(results).forEach(([mech, res]) => {
            if (res) {
                res.hasInternalDiff = checkInternalDiff(res);
            }
        });
    };

    // 小模块显示文字，前置勾叉
    const renderModuleText = (label, score) => (
        <span style={{ fontWeight: "bold", marginRight: "6px", color:"#555" }}>
            {score === 100 ? "✅" : "❌"} {label}
        </span>
    );

    //9.17
    const CollapsibleModule = ({ label, score, children, style }) => {
        const [open, setOpen] = useState(false);
        return (
            <div
                style={{
                    border: "1px solid #fff",
                    borderRadius: "8px",
                    padding: "10px",
                    marginBottom: "10px",
                    backgroundColor: "#fff",
                    width: "100%",     // ⭐️ 关键：让模块占满可用宽度
                    ...style           // 可以从外部传入额外样式
                }}
            >
                {/* 标题行 */}
                <div
                    style={{ display: "flex", alignItems: "center", cursor: "pointer" }}
                    onClick={() => setOpen(!open)}
                >
                    {renderModuleText(label, score)}
                    <span style={{ marginLeft: "6px", color: "#333" }}>{open ? "▲" : "▼"}</span>
                </div>
    
                {/* 折叠内容 */}
                {open && <div style={{ marginTop: "8px" }}>{children}</div>}
            </div>
        );
    };
    
    
    //9.11

        //9.17
    // 新的布尔模块
    const renderStatusText = (label, hasIssue) => (
        <span style={{ fontWeight: "bold", marginRight: "6px", color: "#555" }}>
        {hasIssue ? "❌" : "✅"} {label}
        </span>
    );
    
    const StatusModule = ({ label, hasIssue, children }) => {
        const [open, setOpen] = useState(false);
    
        return (
        <div
            style={{
            border: "1px solid #fff",
            borderRadius: "8px",
            padding: "10px",
            marginBottom: "10px",
            backgroundColor: "#fff",
            }}
        >
            {/* 标题行 */}
            <div
            style={{ display: "flex", alignItems: "center", cursor: "pointer" }}
            onClick={() => setOpen(!open)}
            >
            {renderStatusText(label, hasIssue)}
            <span style={{ marginLeft: "6px", color: "#333" }}>{open ? "▲" : "▼"}</span>
            </div>
    
            {/* 折叠内容 */}
            {open && <div style={{ marginTop: "8px" }}>{children}</div>}
        </div>
        );
    };

    // 9.15_5
    const [showAnalyzerMap, setShowAnalyzerMap] = useState({});

    const toggleAnalyzer = (key, host, port) => {
    setShowAnalyzerMap(prev => ({
        ...prev,
        [key]: !prev[key],
    }));
    };


    //9.11_2
    function getCertGrade(certInfo) {
        if (!certInfo) return { grade: "N/A", issues: [] };
    
        const issues = [];
    
        if (!certInfo.IsTrusted) issues.push("🔒 服务器证书未被受信任的 CA 签发，可能存在风险。");
        if (!certInfo.IsHostnameMatch) issues.push("🌐 证书中的主机名与实际访问的域名不一致，存在中间人攻击风险。");
        if (certInfo.IsExpired) issues.push("⏰ 证书已过期，需更新。");
        if (certInfo.IsSelfSigned) issues.push("⚠️ 证书为自签名，客户端可能无法验证其真实性。");
        if (!certInfo.IsInOrder) issues.push("📑 证书链顺序错误，部分客户端可能验证失败。");
        if (certInfo.AlgWarning) issues.push(`🔧 使用的签名算法存在安全隐患: ${certInfo.AlgWarning}`);
    
        let grade = "A";
        if (!certInfo.IsTrusted || !certInfo.IsHostnameMatch) grade = "B";
        if (!certInfo.IsTrusted && !certInfo.IsHostnameMatch) grade = "C";
    
        return { grade, issues };
    }
    
    function getDNSSummary(result) {
        const dns = result?.dns_record;
        if (!dns) return { score: 0, details: ["无有效 SRV 记录"] };
        
        const details = [];
        const adBits = {
            IMAP: dns.ADbit_imap,
            IMAPS: dns.ADbit_imaps,
            POP3: dns.ADbit_pop3,
            POP3S: dns.ADbit_pop3s,
            SMTP: dns.ADbit_smtp,
            SMTPS: dns.ADbit_smtps,
        };
    
        Object.entries(adBits).forEach(([proto, bit]) => {
            if (bit === true) details.push(`${proto} ✅ DNSSEC 有效`);
            else if (bit === false) details.push(`${proto} ❌ DNSSEC 无效`);
            else details.push(`${proto} ⚪ 未检测到结果`);
        });
    
        return { score: result.score?.dnssec_score || 0, details };
    }

    // 计算证书 + DNSSEC 分数
    function calculateCertDnsScore(results) {
        const certGrades = ["autodiscover", "autoconfig"]
            .map(m => results[m]?.cert_info ? getCertGrade(results[m].cert_info).grade : null)
            .filter(Boolean);

        let certScore = 0;
        if (certGrades.length > 0) {
            certScore = Math.round(certGrades
                .map(g => ({ "A": 100, "B": 60, "C": 30 }[g]))
                .reduce((a, b) => a + b, 0) / certGrades.length);
        }

        const dnsScore = results.srv ? getDNSSummary(results.srv).score : 0;
        return Math.round(certScore * 0.7 + dnsScore * 0.3);
    }

    // 计算最终综合评分
    function calculateOverallConfigScore(consistencyScore, certDnsScore) {
        const weights = { consistency: 0.4, certDns: 0.6 };
        return Math.round(consistencyScore * weights.consistency + certDnsScore * weights.certDns);
    }
    //9.11_2

    //9.15_3
    {/* 辅助函数：提取证书问题列表 */}
    function extractCertIssues(certInfo) {
        if (!certInfo) return [];

        const issues = [];
        if (!certInfo.IsTrusted) issues.push("🔒 服务器证书未被受信任的 CA 签发，可能存在风险。");
        if (!certInfo.IsHostnameMatch) issues.push("🌐 证书中的主机名与实际访问的域名不一致，存在中间人攻击风险。");
        if (certInfo.IsExpired) issues.push("⏰ 证书已过期，需更新。");
        if (certInfo.IsSelfSigned) issues.push("⚠️ 证书为自签名，客户端可能无法验证其真实性。");
        if (!certInfo.IsInOrder) issues.push("📑 证书链顺序错误，部分客户端可能验证失败。");
        if (certInfo.AlgWarning) issues.push(`🔧 使用的签名算法存在安全隐患: ${certInfo.AlgWarning}`);

        return issues;
    }

    // 内部使用：计算评级（不在 UI 渲染）
    function computeCertGrade(certInfo) {
        if (!certInfo) return "N/A";

        let grade = "A";
        if (!certInfo.IsTrusted || !certInfo.IsHostnameMatch) grade = "B";
        if (!certInfo.IsTrusted && !certInfo.IsHostnameMatch) grade = "C";

        return grade;
    }

    // 当前机制内容渲染函数7.28
    const renderMechanismContent = (mech) => {
        const result = results[mech];
        //9.11
        preprocessResults(results);

        // 9.17
        if (mech === "overview") {

            // ===== 先判断是否有任何真实数据10.9 =====
            const hasData = Object.values(results).some(r => r && (
                (r.all && r.all.length > 0) || 
                (r.score_detail && Object.keys(r.score_detail).length > 0)
            ));
            if (!hasData) {
                return <p style={{ color: "gray", marginTop: "2rem" }}>No data to display for overview</p>;
            }

            // ===== 1️⃣ 配置信息差异性 =====
            const comparisonMap = comparePortsUsage(results); // 比较不同机制间
            //let consistencyScore = 100;
            let consistencyScore = 0;
            if (Object.keys(comparisonMap).length > 0){
                consistencyScore = 100;
                Object.entries(comparisonMap).forEach(([_, mechData]) => {
                    const fields = ["host", "ssl"];
                    const isConsistent = fields.every((field) => {
                        const values = Object.values(mechData)
                            .map((m) => m[field])
                            .filter(Boolean);
                        return values.length <= 1 || values.every((v) => v === values[0]);
                    });
                    if (!isConsistent) consistencyScore = 50;
                });

                Object.entries(results).forEach(([mech, res]) => {
                    if (res?.hasInternalDiff) consistencyScore = 30;
                });
            }

            // Object.entries(comparisonMap).forEach(([_, mechData]) => {
            //     const fields = ["host", "ssl"];
            //     const isConsistent = fields.every((field) => {
            //         const values = Object.values(mechData)
            //             .map((m) => m[field])
            //             .filter(Boolean);
            //         return values.length <= 1 || values.every((v) => v === values[0]);
            //     });
            //     if (!isConsistent) consistencyScore = 50;
            // });

            // Object.entries(results).forEach(([mech, res]) => {
            //     if (res?.hasInternalDiff) consistencyScore = 30;
            // });

            // ===== 2️⃣ 配置信息获取过程安全性 =====
            const mechanismList = ["autodiscover", "autoconfig"];
            const httpIssues = {};
            const certIssues = {};

            mechanismList.forEach(m => {
                if (results[m]){
                    httpIssues[m] = results[m]?.score_detail?.http_insecure || false;
                    const certInfo = results[m]?.cert_info;
                    certIssues[m] = certInfo ? getCertGrade(certInfo).issues.length > 0 : false;
                }
                // httpIssues[m] = results[m]?.score_detail?.http_insecure || false;
                // const certInfo = results[m]?.cert_info;
                // certIssues[m] = certInfo ? getCertGrade(certInfo).issues.length > 0 : false;
            });

            // ===== SRV 机制风险分析 =====
            let srvIssue = false;
            const srvDetails = [];
            if (results.srv?.dns_record) {
                const adBits = results.srv.dns_record;
                Object.entries({
                    IMAP: adBits.ADbit_imap,
                    IMAPS: adBits.ADbit_imaps,
                    POP3: adBits.ADbit_pop3,
                    POP3S: adBits.ADbit_pop3s,
                    SMTP: adBits.ADbit_smtp,
                    SMTPS: adBits.ADbit_smtps
                }).forEach(([proto, bit]) => {
                    let text = "⚪ 未检测到结果";
                    if (bit === true) text = "✅ DNSSEC 有效";
                    if (bit === false) {
                        text = "❌ DNSSEC 无效";
                        srvIssue = true;
                    }
                    srvDetails.push({ proto, text });
                });
            } else {
                srvDetails.push({ proto: "SRV", text: "无有效 SRV 记录" });
                // srvIssue = true; //10.29
                srvIssue = false;
            }

            //let configScore = 100; //10.9
            let configScore = 0;
            if (mechanismList.some(m => results[m]|| results.srv)){
                configScore = 100;
                mechanismList.forEach(m => {
                    if (httpIssues[m]) configScore -= 10;
                    if (certIssues[m]) configScore -= 30;
                });
                if (srvIssue) configScore -= 20;
                if (configScore < 0) configScore = 1;
            }
            // mechanismList.forEach(m => {
            //     if (httpIssues[m]) configScore -= 10;
            //     if (certIssues[m]) configScore -= 30;
            // });
            // if (srvIssue) configScore -= 20;
            // if (configScore < 0) configScore = 0;

            let connectScore = 0;
            if (["autodiscover","autoconfig","srv"].some(m => results[m])) {
                connectScore = 100;
                ["autodiscover","autoconfig","srv"].forEach(m => {
                    const mech = results[m];
                    if (!mech) return;

                    const allDetails = [];
                    if (mech.all) {
                        mech.all.forEach(item =>
                            item.score_detail?.actualconnect_details?.forEach(d => allDetails.push(d))
                        );
                    } else if (mech.score_detail?.actualconnect_details) {
                        mech.score_detail.actualconnect_details.forEach(d => allDetails.push(d));
                    }

                    allDetails.forEach(d => {
                        if (d.plain?.success) connectScore -= 40;
                        if (!d.plain?.success && !d.tls?.success && !d.starttls?.success) connectScore -= 50;
                    });
                });
                // if (connectScore < 0) connectScore = 0;
                if (connectScore < 0) connectScore = 1;
                if (connectScore > 100) connectScore = 100;
            }

            // let connectScore = 100;
            // ["autodiscover","autoconfig","srv"].forEach(m => {
            //     const mech = results[m];
            //     if (!mech) return;
            //     const allDetails = [];
            //     if (mech.all) {
            //         mech.all.forEach(item =>
            //             item.score_detail?.actualconnect_details?.forEach(d => allDetails.push(d))
            //         );
            //     } else {
            //         mech.score_detail?.actualconnect_details?.forEach(d => allDetails.push(d));
            //     }

            //     allDetails.forEach(d => {
            //         if (d.plain?.success) connectScore -= 40;
            //         if (!d.plain?.success && !d.tls?.success && !d.starttls?.success) connectScore -= 50;
            //     });
            // });
            // if (connectScore < 0) connectScore = 0;

            let lexScore = 0;
            if (["autodiscover","autoconfig"].some(m => results[m])) {
                lexScore = 100;
                ["autodiscover","autoconfig"].forEach(m => {
                    const mech = results[m];
                    if (!mech) return;
                    const allPorts = [];
                    if (mech.all) {
                        mech.all.forEach(item => item.score_detail?.ports_usage?.forEach(p => allPorts.push(p)));
                    } else if (mech.score_detail?.ports_usage) {
                        mech.score_detail.ports_usage.forEach(p => allPorts.push(p));
                    }
                    if (allPorts.some(p => p.status !== "standard")) lexScore = 60;
                });
            }

            // let lexScore = 100;
            // ["autodiscover","autoconfig"].forEach(m => {
            //     const mech = results[m];
            //     if (!mech) return;
            //     const allPorts = [];
            //     if (mech.all) {
            //         mech.all.forEach(item => item.score_detail?.ports_usage?.forEach(p => allPorts.push(p)));
            //     } else {
            //         mech.score_detail?.ports_usage?.forEach(p => allPorts.push(p));
            //     }
            //     if (allPorts.some(p => p.status !== "standard")) lexScore = 60;
            // });


            

            // ===== 整体配置获取过程问题 =====
            //10.9const configIssue = mechanismList.some(m => httpIssues[m] || certIssues[m]) || srvIssue;
            const configIssue = mechanismList.some(m => results[m] && (httpIssues[m] || certIssues[m])) || srvIssue;

            // ===== 分数转等级和颜色 =====
            const getGradeInfo = (score) => {
                if (score >= 90) return { grade: "A", color: "#2ecc71" }; // 绿
                if (score >= 50) return { grade: "B", color: "#f1c40f" }; // 黄
                if (score >= 30) return { grade: "C", color: "#e67e22" }; // 橙
                return { grade: "D", color: "#e74c3c" };                  // 红
            };

            // ===== 评级框 =====
            const gradeBox = (score) => {
                const { grade, color } = getGradeInfo(score);
                return (
                    <div
                        style={{
                            width: "100px",
                            height: "100px",
                            borderRadius: "10px",
                            border: "2px solid #333",
                            display: "flex",
                            alignItems: "center",
                            justifyContent: "center",
                            fontSize: "28px",
                            fontWeight: "bold",
                            background: color,
                            color: "#fff",
                            marginRight: "20px",
                        }}
                    >
                        {grade}
                    </div>
                );
            };


            return (
                <div style={{ marginTop: "2rem" }}>

                    {/* ===== 概览说明 ===== */}
                    <div
                        style={{
                            backgroundColor: "#eef6f7",
                            border: "1px solid #ccd6dd",
                            borderRadius: "8px",
                            padding: "1rem 1.5rem",
                            marginBottom: "1.5rem",
                            color: "#333",
                            lineHeight: 1.6,
                            fontSize: "15px"
                        }}
                    >
                        <p style={{ margin: 0 }}>
                            本界面用于概览该邮件域在配置信息获取和邮箱服务器连接阶段的安全风险评估结果，
                            包括不同自动化配置机制路径下配置信息的差异性、配置获取过程的安全性、 实际连接安全性以及配置文件内容的规范性等。
                        </p>
                        <p style={{ margin: "0.5rem 0 0 0" }}>
                            若需查看某一具体机制（Autodiscover、Autoconfig、SRV、Guess） 的检测结果，请切换到对应机制的界面。
                        </p>
                    </div>

                    {/* ===== 配置信息差异性 ===== */}
                    <div style={{
                        borderTop: "2px solid #333",
                        paddingTop: "10px",
                        marginBottom: "20px"
                    }}>
                        <div style={{ display: "flex", alignItems: "center", marginBottom: "10px" }}>
                            <span style={{ fontSize: "32px", marginRight: "10px" }}>📊</span>
                            <h3 style={{ margin: 0, color: "#333" }}>配置信息差异性</h3>
                        </div>

                        <div style={{ display: "flex", alignItems: "flex-start", marginBottom: "20px", width: "100%"}}>
                            {/* {gradeBox(consistencyScore)}10.9 */}
                            {consistencyScore > 0 ? gradeBox(consistencyScore) : <div style={{ marginRight: "20px" }}>⚪ 无检测结果</div>}
                            <CollapsibleModule
                                label="配置信息差异性"
                                score={consistencyScore}
                                style={{
                                    flex: 1,
                                    minWidth: 0,
                                    padding: "8px",
                                    backgroundColor: "#eef6f7",
                                    borderRadius: "6px",
                                    minHeight: "100px"
                                }}
                            >
                                <ul style={{ margin: 0, paddingLeft: "18px", color: "#333" }}>
                                    {Object.entries(results).map(([mech, res]) =>
                                        res?.hasInternalDiff && (
                                            <li key={mech}>机制 <b>{mech}</b> 内部存在配置差异</li>
                                        )
                                    )}
                                    {consistencyScore === 50 && <li>不同机制之间存在配置差异</li>}
                                    {consistencyScore === 100 && <li>所有机制配置完全一致</li>}
                                </ul>
                            </CollapsibleModule>

                        </div>

                        {/* 表格详情：直接显示 */}
                        <div style={{ marginBottom: "20px" }}>
                            <h4 style={{ marginBottom: "10px", color: "#333" }}>⚖️ 配置比较详情（不同机制间差异）</h4>
                            <table style={{
                                width: "100%",
                                borderCollapse: "collapse",
                                tableLayout: "fixed",
                                color: "#333"
                            }}>
                                <thead>
                                    <tr>
                                        <th style={{ border: "1px solid #ccc", padding: "8px", background: "#f7f7f7" }}>协议</th>
                                        <th style={{ border: "1px solid #ccc", padding: "8px", background: "#f7f7f7" }}>端口</th>
                                        <th style={{ border: "1px solid #ccc", padding: "8px", background: "#f7f7f7" }}>机制</th>
                                        <th style={{ border: "1px solid #ccc", padding: "8px", background: "#f7f7f7" }}>主机</th>
                                        <th style={{ border: "1px solid #ccc", padding: "8px", background: "#f7f7f7" }}>SSL类型</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {Object.entries(comparisonMap).map(([key, mechData], idx) => {
                                        const fields = ["host", "ssl"];
                                        const isConsistent = fields.every((field) => {
                                            const values = Object.values(mechData).map((m) => m[field]).filter(Boolean);
                                            return values.length <= 1 || values.every((v) => v === values[0]);
                                        });

                                        return Object.entries(mechData).map(([mech, item], rowIdx) => (
                                            <tr key={`${idx}-${rowIdx}`} style={{ backgroundColor: isConsistent ? "#f0f0f0" : "#f8d7da" }}>
                                                {rowIdx === 0 && (
                                                    <>
                                                        <td style={{ border: "1px solid #ccc", padding: "8px", textAlign: "center" }} rowSpan={Object.keys(mechData).length}>{item.protocol}</td>
                                                        <td style={{ border: "1px solid #ccc", padding: "8px", textAlign: "center" }} rowSpan={Object.keys(mechData).length}>{item.port}</td>
                                                    </>
                                                )}
                                                <td style={{ border: "1px solid #ccc", padding: "8px", textAlign: "center" }}>{mech}</td>
                                                <td style={{ border: "1px solid #ccc", padding: "8px", textAlign: "center" }}>{item.host}</td>
                                                <td style={{ border: "1px solid #ccc", padding: "8px", textAlign: "center" }}>{item.ssl}</td>
                                            </tr>
                                        ));
                                    })}
                                </tbody>
                            </table>
                        </div>
                    </div>

                    {/* ===== 配置获取过程安全性 ===== */}
                    <div style={{
                        borderTop: "2px solid #333",
                        paddingTop: "10px",
                        marginBottom: "20px",
                    }}>
                        <div style={{ display: "flex", alignItems: "center", marginBottom: "10px" }}>
                            <span style={{ fontSize: "32px", marginRight: "10px" }}>🛡️</span>
                            <h3 style={{ margin: 0, color: "#333" }}>配置获取过程安全性</h3>
                        </div>

                        <div style={{ display: "flex", alignItems: "stretch", marginBottom: "20px", width: "100%" }}>
                            {/* 左侧评级框 */}
                            {/* {gradeBox(configScore)}10.9 */}
                            {configScore > 0 ? gradeBox(configScore) : <div style={{ marginRight: "20px" }}>⚪ 无检测结果</div>}
                            {/* 右侧 StatusModule */}
                            <div style={{ flex: 1, minWidth: 0}}>
                                <StatusModule label="配置获取过程安全性" hasIssue={configIssue}>
                                    {mechanismList.map(m => (
                                        results[m] ? (
                                        <div key={m} style={{ marginBottom: "10px" }}>
                                            <StatusModule label={`${m} HTTP连接方式`} hasIssue={httpIssues[m]}>
                                                <div style={{
                                                    margin: "4px 0 6px 0",
                                                    padding: "6px",
                                                    backgroundColor: "#eef6f7",
                                                    borderRadius: "4px",
                                                    fontSize: "0.85rem",
                                                    color: "#333"
                                                }}>
                                                    {httpIssues[m]
                                                        ? "通过 HTTP 获取配置，存在被篡改风险"
                                                        : "通过 HTTPS 获取配置，安全"}
                                                </div>
                                            </StatusModule>

                                            <StatusModule label={`${m} 配置服务器证书`} hasIssue={certIssues[m]}>
                                                <div style={{
                                                    margin: "4px 0 6px 0",
                                                    padding: "6px",
                                                    backgroundColor: "#eef6f7",
                                                    borderRadius: "4px",
                                                    fontSize: "0.85rem",
                                                    color: "#333"
                                                }}>
                                                    {certIssues[m]
                                                        ? "证书验证存在问题"
                                                        : "证书验证通过"}
                                                </div>
                                            </StatusModule>
                                        </div>
                                        ) : <div key={m} style={{ marginBottom: "10px" }}>{m.toUpperCase()} ⚪ 无检测结果</div>
                                    ))}

                                    <StatusModule label="SRV 配置查询过程风险分析" hasIssue={srvIssue}>
                                        <ul style={{ margin: 0, paddingLeft: "18px", color: "#333" }}>
                                            {srvDetails.map((d, idx) => <li key={idx}>{d.proto} {d.text}</li>)}
                                        </ul>
                                    </StatusModule>
                                </StatusModule>
                            </div>
                        </div>

                    </div>

                    {/* ===== 实际连接安全性 ===== */}
                    <div style={{
                        borderTop: "2px solid #333",
                        paddingTop: "10px",
                        marginBottom: "20px"
                    }}>
                        <div style={{ display: "flex", alignItems: "center", marginBottom: "10px" }}>
                            <span style={{ fontSize: "32px", marginRight: "10px" }}>🔒</span>
                            <h3 style={{ margin: 0, color: "#333" }}>实际连接安全性</h3>
                        </div>
                        <div style={{ display: "flex", alignItems: "flex-start", marginBottom: "20px", width: "100%" }}>
                            {/* {gradeBox(connectScore)} */}
                            {connectScore > 0 ? gradeBox(connectScore) : <div style={{ marginRight: "20px" }}>⚪ 无检测结果</div>}
                            {/* 右侧 StatusModule */}
                            <div style={{ flex: 1, minWidth: 0 }}>
                                <StatusModule
                                    label="实际连接安全性"
                                    hasIssue={(() => {
                                        return ["autodiscover", "autoconfig", "srv"].some(m => {
                                            const mech = results[m];
                                            if (!mech) return false;

                                            const allDetails = [];
                                            if (mech.all) {
                                                mech.all.forEach(item =>
                                                    item.score_detail?.actualconnect_details?.forEach(d => allDetails.push(d))
                                                );
                                            } else {
                                                mech.score_detail?.actualconnect_details?.forEach(d => allDetails.push(d));
                                            }
                                            return allDetails.some(d => d.plain?.success);
                                        });
                                    })()}
                                >
                                    {["autodiscover", "autoconfig", "srv"].map(m => {
                                        const mech = results[m];
                                        if (!mech) return null;

                                        const serverMap = {};
                                        if (mech.all) {
                                            mech.all.forEach(item => {
                                                item.score_detail?.actualconnect_details?.forEach(d => {
                                                    if (!serverMap[d.host]) serverMap[d.host] = [];
                                                    const exists = serverMap[d.host].some(
                                                        x => x.type === d.type && x.port === d.port
                                                    );
                                                    if (!exists) serverMap[d.host].push(d);
                                                });
                                            });
                                        } else {
                                            mech.score_detail?.actualconnect_details?.forEach(d => {
                                                if (!serverMap[d.host]) serverMap[d.host] = [];
                                                const exists = serverMap[d.host].some(
                                                    x => x.type === d.type && x.port === d.port
                                                );
                                                if (!exists) serverMap[d.host].push(d);
                                            });
                                        }

                                        if (Object.keys(serverMap).length === 0) {
                                            return (
                                                <div key={m} style={{ marginBottom: "10px" }}>
                                                    <strong>{m.toUpperCase()}：</strong> ⚪ 无检测结果
                                                </div>
                                            );
                                        }

                                        return (
                                            <div key={m} style={{ marginBottom: "12px" }}>
                                                <strong>{m.toUpperCase()}：</strong>
                                                {Object.entries(serverMap).map(([host, details], idx) => (
                                                    <div key={idx} style={{
                                                        marginTop: "6px",
                                                        padding: "6px",
                                                        border: "1px solid #ccc",
                                                        borderRadius: "6px",
                                                        backgroundColor: "#f9f9f9",
                                                    }}>
                                                        <strong>{host}</strong>
                                                        <ul style={{ margin: "4px 0 0 16px" }}>
                                                            {details.map((d, dIdx) => (
                                                                <li key={dIdx}>
                                                                    {d.type.toUpperCase()} : {d.port} →
                                                                    {d.plain?.success && (
                                                                        <span style={{ color: "red", marginLeft: "8px" }}>⚠️ 明文可连接</span>
                                                                    )}
                                                                    {(d.tls?.success || d.starttls?.success) && (
                                                                        <span style={{ color: "green", marginLeft: "8px" }}>✅ 安全连接可用</span>
                                                                    )}
                                                                    {!d.plain?.success && !d.tls?.success && !d.starttls?.success && (
                                                                        <span style={{ color: "gray", marginLeft: "8px" }}>❌ 无法连接</span>
                                                                    )}
                                                                </li>
                                                            ))}
                                                        </ul>
                                                    </div>
                                                ))}
                                            </div>
                                        );
                                    })}

                                    {/* 🔹 框最下面统一注释 */}
                                    {["autodiscover", "autoconfig", "srv"].some(m => {
                                        const mech = results[m];
                                        if (!mech) return false;
                                        const allDetails = [];
                                        if (mech.all) {
                                            mech.all.forEach(item =>
                                                item.score_detail?.actualconnect_details?.forEach(d => allDetails.push(d))
                                            );
                                        } else {
                                            mech.score_detail?.actualconnect_details?.forEach(d => allDetails.push(d));
                                        }
                                        return allDetails.some(d => d.plain?.success);
                                    }) && (
                                        <div style={{
                                            marginTop: "8px",
                                            padding: "6px",
                                            backgroundColor: "#fff3cd",
                                            border: "1px solid #ffeeba",
                                            borderRadius: "6px",
                                            color: "#856404",
                                            fontSize: "14px",
                                        }}>
                                            注：判断服务器是否支持明文连接需通过实际登录验证，本工具仅限于TCP/TLS连接阶段的探测，未执行任何登录尝试，仅给出风险提示。 
                                        </div>
                                    )}
                                </StatusModule>
                            </div>
                        </div>
                    </div>

                    {/* ===== 配置文件词法解析规范性 ===== */}
                    <div style={{
                        borderTop: "2px solid #333",
                        paddingTop: "10px",
                        marginBottom: "20px"
                    }}>
                        <div style={{ display: "flex", alignItems: "center", marginBottom: "10px" }}>
                            <span style={{ fontSize: "32px", marginRight: "10px" }}>📑</span>
                            <h3 style={{ margin: 0, color: "#333" }}>配置文件词法解析规范性</h3>
                        </div>

                        <div style={{ display: "flex", alignItems: "flex-start", marginBottom: "20px", width: "100%" }}>
                            {/* {gradeBox(lexScore)} */}
                            {lexScore > 0 ? gradeBox(lexScore) : <div style={{ marginRight: "20px" }}>⚪ 无检测结果</div>}

                            <div style={{ flex: 1, minWidth: 0 }}>
                                <StatusModule
                                    label="配置文件词法解析规范性"
                                    hasIssue={(() => {
                                        return ["autodiscover", "autoconfig"].some(m => {
                                            const mech = results[m];
                                            if (!mech) return false;
                                            const allPorts = [];
                                            if (mech.all) {
                                                mech.all.forEach(item =>
                                                    item.score_detail?.ports_usage?.forEach(p => allPorts.push(p))
                                                );
                                            } else {
                                                mech.score_detail?.ports_usage?.forEach(p => allPorts.push(p));
                                            }
                                            return allPorts.some(p => p.status !== "standard");
                                        });
                                    })()}
                                >
                                    <ul style={{ margin: 0, paddingLeft: "18px", color: "#333" }}>
                                        {["autodiscover", "autoconfig"].map(m => {
                                            const mech = results[m];
                                            if (!mech) return null;

                                            const allPorts = [];
                                            if (mech.all) {
                                                mech.all.forEach(item =>
                                                    item.score_detail?.ports_usage?.forEach(p => allPorts.push(p))
                                                );
                                            } else {
                                                mech.score_detail?.ports_usage?.forEach(p => allPorts.push(p));
                                            }

                                            if (allPorts.length === 0) {
                                                return <li key={m}>{m} ⚪ 无检测结果</li>;
                                            }

                                            const hasIssue = allPorts.some(p => p.status !== "standard");
                                            return (
                                                <li key={m}>
                                                    {m} {hasIssue ? "❌ 存在不符合规范的配置" : "✅ 全部符合规范"}
                                                </li>
                                            );
                                        })}
                                    </ul>
                                </StatusModule>
                            </div>
                        </div>
                    </div>
                </div>
            );
        }

        
        if (!result && Object.keys(results).length === 0) return null;
        if (!result) return <p style={{ color: "gray" }}>No data for {mech}</p>;

        const score = result.score || result.score_detail?.connection || {};
        // 8.10
        const defense = result.score_detail?.defense;
        const portsUsage = result.score_detail?.ports_usage;
        const detail = result.score_detail?.connection;
        const certInfo = result.cert_info;
        const connectDetails = result.score_detail?.actualconnect_details;

        return (
            <div>
                {(mech === "autodiscover" || mech === "autoconfig") && result.all && (
                    <div style={{
                        fontFamily: '"PingFang SC", "Microsoft YaHei", sans-serif',
                        fontWeight: 400,
                        textAlign: "left"
                        }}>
                        {/* 9.12 */}
                        {/* 🔌 配置信息概况 */}
                        {Array.isArray(portsUsage) && portsUsage.length > 0 && (
                        <div style={{ marginTop: "2rem" }}>
                            {/* 上方主题分界线 */}
                            <div
                            style={{
                                borderTop: "2px solid #333",
                                paddingTop: "10px",
                                marginBottom: "20px",
                                display: "flex",
                                alignItems: "center",
                            }}
                            >
                            <span style={{ fontSize: "32px", marginRight: "10px" }}>🔌</span>
                            <h3 style={{ margin: 0, color: "#333" }}>最佳配置信息</h3>
                            </div>

                            <div style={{ display: "flex", flexWrap: "wrap", gap: "1rem",justifyContent: "center", maxWidth: "1000px", margin: "0 auto" }}>
                            {portsUsage.map((item, idx) => (
                                <div
                                key={idx}
                                style={{
                                    backgroundColor: "#f9f9f9",
                                    color: "#333",
                                    padding: "1rem",
                                    borderRadius: "12px",
                                    boxShadow: "0 2px 8px rgba(85, 136, 207, 0.1)",
                                    border: "1px solid #e0e0e0",
                                    minWidth: "220px",
                                    flex: "1",
                                    maxWidth: "280px",
                                }}
                                >
                                <table style={{ width: "100%", borderCollapse: "collapse" }}>
                                    <tbody>
                                    <tr>
                                        <td style={tdStyle}><strong>协议</strong></td>
                                        <td style={tdStyle}>{item.protocol}</td>
                                    </tr>
                                    <tr>
                                        <td style={tdStyle}><strong>端口</strong></td>
                                        <td style={tdStyle}>{item.port}</td>
                                    </tr>
                                    <tr>
                                        <td style={tdStyle}><strong>主机名</strong></td>
                                        <td style={tdStyle}>{item.host}</td>
                                    </tr>
                                    <tr>
                                        <td style={tdStyle}><strong>SSL类型</strong></td>
                                        <td style={tdStyle}>{item.ssl}</td>
                                    </tr>
                                    <tr>
                                        <td style={tdStyle}><strong>用户名</strong></td>
                                        {/* <td style={tdStyle}>{lastSubmittedEmail}</td> 10.30*/}
                                        <td style={tdStyle}>{isRecommendedClick ? "邮件地址" : lastSubmittedEmail}</td>
                                    </tr>
                                    <tr>
                                        <td style={tdStyle}><strong>密码</strong></td>
                                        <td style={tdStyle}>邮箱密码</td>
                                    </tr>
                                    </tbody>
                                </table>
                                </div>
                            ))}
                            </div>
                        </div>
                        )}

                        <div style={{ marginTop: "2rem" }}>
                        {/* 上方主题分界线 */}
                            <div
                                style={{
                                borderTop: "2px solid #333",
                                paddingTop: "10px",
                                marginBottom: "20px",
                                display: "flex",
                                alignItems: "center",
                                cursor: "pointer",
                                }}
                                onClick={() => toggleRaw(mech)}
                            >
                                <span style={{ fontSize: "32px", marginRight: "10px" }}>🛠️</span>
                                <h3 style={{ margin: 0, color: "#333" }}>
                                原始配置文件 {showRawConfig[mech] ? "▲" : "▼"}
                                </h3>
                            </div>

                            {showRawConfig[mech] && (
                                <pre
                                    style={{
                                    background: "#f4f7f9",
                                    padding: "12px",
                                    borderRadius: "6px",
                                    border: "1px solid #ddd",
                                    maxHeight: "400px",
                                    overflowY: "auto",
                                    fontFamily: "Consolas, Monaco, monospace",
                                    fontSize: "14px",
                                    lineHeight: "1.5",
                                    }}
                                >
                                    {result.config}
                                </pre>
                            )}
                        </div>

                        {/* <div style={{ marginTop: "2rem" }}>
                            <div
                                style={{
                                borderTop: "2px solid #333",
                                paddingTop: "10px",
                                marginBottom: "20px",
                                display: "flex",
                                alignItems: "center",
                                }}
                            >
                                <span style={{ fontSize: "32px", marginRight: "10px" }}>📄</span>
                                <h3 style={{ margin: 0, color: "#333" }}>配置服务器证书信息</h3>
                            </div>

                            <ul>
                                {Object.entries(certInfo || {}).map(([k, v]) =>
                                k !== "RawCert" && k !== "RawCerts" && v !== "" ? (
                                    <li key={k} style={{ color: "#364957", marginBottom: "4px" }}>
                                    <strong>{certLabelMap[k] || k}:</strong> {String(v)}
                                    </li>
                                ) : null
                                )}
                                {certInfo?.RawCerts && (
                                <li>
                                    <strong>原始证书:</strong>
                                    <button
                                    onClick={() => toggleRawCerts(mech)}
                                    style={{
                                        marginLeft: "10px",
                                        padding: "4px 8px",
                                        backgroundColor: "#5b73a9",
                                        color: "#fff",
                                        border: "none",
                                        borderRadius: "4px",
                                        cursor: "pointer",
                                        fontSize: "0.9rem",
                                    }}
                                    >
                                    {showRawCertsMap[mech] ? "隐藏" : "展开"}
                                    </button>
                                    {showRawCertsMap[mech] && (
                                    <div
                                        style={{
                                        wordBreak: "break-all",
                                        maxHeight: "200px",
                                        overflowY: "auto",
                                        marginTop: "10px",
                                        background: "#f5f5f5",
                                        padding: "10px",
                                        borderRadius: "6px",
                                        border: "1px solid #ccc",
                                        fontFamily: "Consolas, Monaco, monospace",
                                        fontSize: "13px",
                                        }}
                                    >
                                        {certInfo.RawCerts.join(", ")}
                                    </div>
                                    )}
                                </li>
                                )}
                            </ul>
                        </div> */}

                        {Array.isArray(certInfo?.RawCerts) && certInfo.RawCerts.length > 0 && ( 
                        <div style={{ marginTop: "2rem" }}>
                            <div
                            style={{
                                borderTop: "2px solid #333",
                                paddingTop: "10px",
                                marginBottom: "20px",
                                display: "flex",
                                alignItems: "center",
                                cursor: "pointer",
                                color: "#333",
                            }}
                            onClick={() => toggleCertChain(mech)}
                            >
                            <span style={{ fontSize: "32px", marginRight: "10px" }}>🔗</span>
                            <h3 style={{ margin: 0, color: "#333" }}>
                                配置服务器证书链 {showCertChainMap[mech] ? "▲" : "▼"}
                            </h3>
                            </div>

                            {showCertChainMap[mech] && (
                            <div
                                style={{
                                background: "#fff",                    // 白色填充
                                border: "1px solid #ddd",              // 浅灰边框
                                borderRadius: "12px",                  // 圆角
                                padding: "1rem",                       // 内边距
                                boxShadow: "0 2px 6px rgba(0,0,0,0.08)" // 阴影效果
                                }}
                            >
                                <div style={{ marginBottom: "10px" }}>
                                {certInfo.RawCerts.map((_, idx) => (
                                    <button
                                    key={idx}
                                    onClick={() => setActiveCertIdx(mech, idx)}
                                    style={{
                                        marginRight: "8px",
                                        padding: "6px 12px",
                                        backgroundColor:
                                        activeCertIdxMap[mech] === idx ? "#5b73a9" : "#ddd",
                                        color: activeCertIdxMap[mech] === idx ? "#fff" : "#000",
                                        border: "none",
                                        borderRadius: "6px",
                                        cursor: "pointer",
                                        fontWeight: "bold",
                                    }}
                                    >
                                    证书 #{idx + 1}
                                    </button>
                                ))}
                                </div>
                                <PeculiarCertificateViewer
                                certificate={certInfo.RawCerts[activeCertIdxMap[mech] || 0]}
                                />
                            </div>
                            )}
                        </div>
                        )}

                        
                        {/* 9.17删 */}

                        <div style={{ marginTop: "2rem" }}>
                            <div
                                style={{
                                borderTop: "2px solid #333",
                                paddingTop: "10px",
                                marginBottom: "20px",
                                display: "flex",
                                alignItems: "center",
                                }}
                            >
                                <span style={{ fontSize: "32px", marginRight: "10px" }}>📡</span>
                                <h3 style={{ margin: 0, color: "#333" }}>
                                可通过 {mech.toUpperCase()} 方法得到的所有配置
                                </h3>
                            </div>

                            
                            
                            <table style={tableStyle}>
                                <thead>
                                    <tr>
                                        {/* 9.17 */}
                                        <th style={thStyle}>序号</th>
                                        <th style={thStyle}>途径</th>
                                        
                                        <th style={thStyle}>请求URI</th>
                                        <th style={thStyle}>是否得到配置</th>
                                        {/* <th style={thStyle}>加密评分</th>
                                        <th style={thStyle}>标准评分</th>
                                        <th style={thStyle}>综合评分</th> */}
                                        <th style={thStyle}>查看详情</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {result.all.map((item, idx) => (
                                        <tr key={idx}>
                                            <td style={tdStyle}>{idx + 1}</td>  {/* 序号从 1 开始9.17 */}
                                            <td style={tdStyle}>{item.method}</td>
                                            {/* 9.11 */}
                                            {/* <td style={tdStyle}>{item.index}</td> */}
                                            <td style={{ ...tdStyle, maxWidth: "250px", overflow: "hidden" }}>
                                            <div
                                                style={{
                                                overflowX: "auto",
                                                whiteSpace: "nowrap",
                                                scrollbarWidth: "thin", // Firefox
                                                scrollbarColor: "#ccc transparent", // Firefox
                                                }}
                                                className="scrollable-uri"
                                            >
                                                {item.uri}
                                            </div>
                                            </td>

                                            <style>
                                                {`
                                                .scrollable-uri::-webkit-scrollbar {
                                                    height: 6px; /* 滚动条高度（横向） */
                                                }
                                                .scrollable-uri::-webkit-scrollbar-thumb {
                                                    background-color: #853333ff; /* 滚动条颜色 */
                                                    border-radius: 3px;
                                                }
                                                .scrollable-uri::-webkit-scrollbar-track {
                                                    background: transparent; /* 背景透明 */
                                                }
                                                `}
                                            </style>


                                            <td style={tdStyle}>{item.config ? "✅" : "❌"}</td>
                                            {/* <td style={tdStyle}>{item.score?.encrypted_ports ?? "-"}</td>
                                            <td style={tdStyle}>{item.score?.standard_ports ?? "-"}</td>
                                            <td style={tdStyle}>{item.score?.overall ?? "-"}</td> */}
                                            <td style={tdStyle}>
                                                {item.config && (
                                                    <button
                                                    onClick={async () => {
                                                        console.log("当前 item:", item);
                                                        const payload = {
                                                            config: item.config,
                                                            uri: item.uri,
                                                            details: item.score_detail?.actualconnect_details || [],
                                                            portsUsage: item.score_detail?.ports_usage || [],
                                                            rawCerts: item.cert_info?.RawCerts || [],
                                                            mech: mech,
                                                        };

                                                        try {
                                                        const res = await fetch("/store-temp-data", {
                                                            method: "POST",
                                                            headers: { "Content-Type": "application/json" },
                                                            body: JSON.stringify(payload),
                                                        });

                                                        if (!res.ok) throw new Error("存储失败");

                                                        const { id } = await res.json();

                                                        // ✅ 避免 431：只带 id
                                                        window.open(`/config-view?id=${id}`, "_blank");
                                                        } catch (err) {
                                                        console.error("❌ 打开详情失败:", err);
                                                        alert("⚠️ 无法打开详情页");
                                                        }
                                                    }}
                                                    style={viewButtonStyle}
                                                    >
                                                    查看
                                                    </button>
                                                )}
                                            </td>

                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                            

                        </div>
                        
                        {/* 9.17 */}
                        <div style={{ marginTop: "2rem" }}>
                            <div
                                style={{
                                    borderTop: "2px solid #333",
                                    paddingTop: "10px",
                                    marginBottom: "20px",
                                    display: "flex",
                                    alignItems: "center",
                                }}
                            >
                                <span style={{ fontSize: "32px", marginRight: "10px" }}>🛡️</span>
                                <h3 style={{ margin: 0, color: "#333" }}>{mech.toUpperCase()} 机制综合风险分析</h3>
                            </div>
                            {/* ===== 说明框 ===== */}
                            <div
                                style={{
                                    border: "1px solid #ccc",
                                    borderRadius: "6px",
                                    backgroundColor: "#f5f5f5",
                                    padding: "10px",
                                    marginBottom: "20px",
                                    color: "#555",
                                    fontSize: "0.9rem"
                                }}
                            >
                                本模块展示邮件服务器在该机制下的综合安全情况，包括：
                                <ul style={{ marginTop: "6px", marginBottom: 0, paddingLeft: "20px" }}>
                                    <li>配置获取过程安全性（传输协议使用与配置服务器证书验证）</li>
                                    <li>配置文件词法解析规范性（加密端口、协议规范）</li>
                                    <li>实际连接结果（明文连接、STARTTLS/TLS连接情况）</li>
                                    <li>使用不同路径获取到的配置文件的差异性</li>
                                </ul>
                                每个部分均会标记潜在安全问题，方便快速判断邮件自动化配置机制在使用中是否存在安全风险。
                            </div>
                            
                            {/* ===== 配置获取过程安全性 ===== */}
                            {(() => {
                            // ====== 计算总的安全性 ======
                            let hasHttp = false;
                            let hasCertIssue = false;

                            result.all.forEach(item => {
                                // 检查是否有 http 协议
                                if (item.redirects && item.redirects.length > 0) {
                                const finalRedirect = item.redirects[item.redirects.length - 1].URL;
                                const finalScheme = finalRedirect ? finalRedirect.split(":")[0].toLowerCase() : null;
                                if (finalScheme === "http") hasHttp = true;
                                }

                                // 检查证书问题
                                if (extractCertIssues(item.cert_info).length > 0) {
                                hasCertIssue = true;
                                }
                            });

                            const hasIssue = hasHttp || hasCertIssue;

                            // ====== 渲染总结果 ======
                            return (
                                <StatusModule label="配置获取过程安全性" hasIssue={hasIssue}>
                                {/* ===== 展开子模块 ===== */}
                                <StatusModule label="配置获取过程HTTP连接方式" hasIssue={hasHttp}>
                                    <div style={{
                                        margin: "4px 0 10px 0",
                                        padding: "6px",
                                        backgroundColor: "#eef6f7",
                                        borderRadius: "4px",
                                        fontSize: "0.85rem",
                                        color: "#333"
                                    }}>
                                        通过观察配置信息最终是通过 HTTP or HTTPS 协议获取到的，可以进一步判断配置信息的可靠性，防止被恶意篡改过的配置信息带来安全风险。
                                    </div>

                                    {result.all.map((item, idx) => {
                                    let finalRedirect = null;
                                    let finalScheme = null;
                                    if (item.redirects && item.redirects.length > 0) {
                                        finalRedirect = item.redirects[item.redirects.length - 1].URL;
                                        finalScheme = finalRedirect ? finalRedirect.split(":")[0].toLowerCase() : null;
                                    }

                                    // 协议说明文字
                                    let protocolDesc = "";
                                    if (finalScheme === "https") {
                                        protocolDesc = `最终通过 HTTPS 加密协议获取配置信息，传输安全可靠，可在一定程度上防止被篡改或窃取。`;
                                    } else if (finalScheme === "http") {
                                        protocolDesc = `最终通过 HTTP 明文协议获取配置信息，存在被篡改或窃取的风险，不安全。`;
                                    } else {
                                        protocolDesc = `未检测到协议信息。`;
                                    }

                                    return (
                                        <div
                                        key={idx}
                                        style={{
                                            marginBottom: "6px",
                                            padding: "6px",
                                            border: "1px solid #ccc",
                                            borderRadius: "6px",
                                            backgroundColor: "#f9f9f9",
                                        }}
                                        >
                                        <strong>路径 {idx + 1}</strong>
                                        {finalScheme && (
                                            <p style={{ margin: "4px 0 0 0", color: finalScheme === "http" ? "#c33" : "#388e3c", fontSize: "0.9rem" }}>
                                                🔗 {protocolDesc}
                                            </p>
                                        )}
                                        </div>
                                    );
                                    })}
                                </StatusModule>

                                <StatusModule label="配置服务器证书验证" hasIssue={hasCertIssue}>
                                    <div style={{
                                        margin: "4px 0 10px 0",
                                        padding: "6px",
                                        backgroundColor: "#eef6f7",
                                        borderRadius: "4px",
                                        fontSize: "0.85rem",
                                        color: "#333"
                                    }}>
                                        通过配置服务器返回的证书进行全面的验证，可以判断服务器身份是否可信，防止中间人攻击或恶意伪造证书造成的安全风险。
                                    </div>
                                    {result.all.map((item, idx) => (
                                    <div
                                        key={idx}
                                        style={{
                                        marginBottom: "6px",
                                        padding: "6px",
                                        border: "1px solid #ccc",
                                        borderRadius: "6px",
                                        backgroundColor: "#f9f9f9",
                                        }}
                                    >
                                        <strong>路径 {idx + 1}</strong>
                                        <ul style={{ margin: "4px 0 0 16px", color: "#333" }}>
                                        {extractCertIssues(item.cert_info).length > 0
                                            ? extractCertIssues(item.cert_info).map((issue, i) => (
                                                <li key={i} style={{ color: "#c33" }}>{issue}</li>
                                            ))
                                            : <li style={{ color: "#388e3c" }}>✅ 证书验证通过，未发现问题</li>
                                        }
                                        </ul>
                                    </div>
                                    ))}
                                </StatusModule>
                                </StatusModule>
                            );
                            })()}

                            {/* ===== 渲染端口连接方式 ===== */}
                            {(() => {
                            const serverMap = {};
                            result.all.forEach((item) => {
                                item.score_detail?.ports_usage?.forEach((p) => {
                                if (!serverMap[p.host]) serverMap[p.host] = { ports: [] };
                                if (!serverMap[p.host].ports.some(portObj => portObj.port === p.port && portObj.protocol === p.protocol)) {
                                    serverMap[p.host].ports.push(p);
                                }
                                });
                            });

                            // 判断是否有不符合规范的加密元素值
                            const hasLexicalIssue = Object.values(serverMap).some(info =>
                                info.ports.some(p => p.status !== "standard")
                            );

                            return (
                                <StatusModule label="配置文件词法解析规范性" hasIssue={hasLexicalIssue}>
                                {Object.entries(serverMap).map(([host, info], idx) => {
                                // 把所有端口拼接成字符串
                                const portList = info.ports.map(p => p.port).join(", ");

                                return (
                                    <div
                                    key={idx}
                                    style={{
                                        marginBottom: "6px",
                                        padding: "6px",
                                        border: "1px solid #ccc",
                                        borderRadius: "6px",
                                        backgroundColor: "#f9f9f9",
                                    }}
                                    >
                                    <strong>
                                        {host}（端口: {portList}）
                                    </strong>
                                    <ul style={{ margin: "4px 0 0 16px" }}>
                                        {info.ports.map((p, portIdx) => (
                                        <li key={portIdx}>
                                            经过词法分析，配置文件中的元素
                                            <strong style={{ marginLeft: "4px" }}>
                                            {p.status === "standard" ? "符合规范" : "不符合规范"}
                                            </strong>
                                        </li>
                                        ))}
                                    </ul>
                                    </div>
                                );
                                })}

                                </StatusModule>
                            );
                            })()}

                            {/* ===== 渲染实际连接结果 ===== */}
                            {(() => {
                            const serverMap = {};
                            result.all.forEach(item => {
                                item.score_detail?.actualconnect_details?.forEach(d => {
                                if (!serverMap[d.host]) serverMap[d.host] = { actualconnect_details: [] };

                                // 去重：同一个 host 下同 protocol+port 只保留一条
                                const exists = serverMap[d.host].actualconnect_details.some(
                                    x => x.type === d.type && x.port === d.port
                                );
                                if (!exists) {
                                    serverMap[d.host].actualconnect_details.push(d);
                                }
                                });
                            });

                        // 判断是否有不安全连接（plain.success === true）
                            const hasConnectIssue = Object.values(serverMap).some(info =>
                                info.actualconnect_details.some(d => d.plain?.success)
                            );

                            return (
                                <StatusModule label="实际连接结果" hasIssue={hasConnectIssue}>
                                {Object.entries(serverMap).map(([host, info], idx) => (
                                    <div
                                    key={idx}
                                    style={{
                                        marginBottom: "6px",
                                        padding: "6px",
                                        border: "1px solid #ccc",
                                        borderRadius: "6px",
                                        backgroundColor: "#f9f9f9",
                                    }}
                                    >
                                    <strong>{host}</strong>
                                    <ul style={{ margin: "4px 0 0 16px" }}>
                                        {info.actualconnect_details.map((d, dIdx) => (
                                        <li key={dIdx}>
                                            {d.type.toUpperCase()} : {d.port} →
                                            {d.plain?.success && <span style={{ color: "red", marginLeft: "8px" }}>⚠️ 明文可连接</span>}
                                            {(d.tls?.success || d.starttls?.success) && <span style={{ color: "green", marginLeft: "8px" }}>✅ 安全连接可用</span>}
                                            {!d.plain?.success && !d.tls?.success && !d.starttls?.success && <span style={{ color: "gray", marginLeft: "8px" }}>❌ 无法连接</span>}

                                            {/* 🔎 动态展开分析面板 */}
                                            {showAnalyzerMap[`${host}-${d.port}`] && (
                                            <div style={{ marginTop: "6px" }}>
                                                <TlsAnalyzerPanel host={d.host} port={d.port} />
                                            </div>
                                            )}
                                        </li>
                                        ))}
                                    </ul>
                                    </div>
                                ))}

                                {/* ✅ 框最下面统一注释10.30 */}
                                {hasConnectIssue && (
                                    <div style={{
                                        marginTop: "8px",
                                        padding: "6px",
                                        backgroundColor: "#fff3cd",
                                        border: "1px solid #ffeeba",
                                        borderRadius: "6px",
                                        color: "#856404",
                                        fontSize: "14px",
                                    }}>
                                        注：判断服务器是否支持明文连接需通过实际登录验证，本工具仅限于TCP/TLS连接阶段的探测，未执行任何登录尝试，仅给出风险提示。 
                                    </div>
                                )}
                                </StatusModule>
                            );
                            })()}


                        </div>

                        {/* 9.17 */}
                        {/* ===== 机制内部差异性 ===== */}
                        {(() => {
                            const collectPortsUsage = (allResults) => {
                                return allResults.map(r => ({
                                uri: r.uri || "",
                                ports: Array.isArray(r?.score_detail?.ports_usage) ? r.score_detail.ports_usage : []
                                }));
                            };

                            const allPorts = collectPortsUsage(result.all);
                            if (allPorts.length === 0) return null;

                            // 按 protocol → [路径1配置集合, 路径2配置集合...]
                            const protocolGroups = {};
                            allPorts.forEach(item => {
                                item.ports.forEach(p => {
                                if (!protocolGroups[p.protocol]) protocolGroups[p.protocol] = [];
                                });
                            });

                            // 每条路径的协议配置去重
                            allPorts.forEach(item => {
                                const pathIndex = allPorts.indexOf(item);
                                for (const proto in protocolGroups) {
                                const portsOfProto = item.ports
                                    .filter(p => p.protocol === proto)
                                    .map(p => `${p.host}:${p.port} (${p.ssl})`);
                                const unique = [...new Set(portsOfProto)];
                                protocolGroups[proto][pathIndex] = unique;
                                }
                            });

                            // 比较路径间是否一致
                            const diffMap = {};
                            for (const proto in protocolGroups) {
                                const sets = protocolGroups[proto].map(arr => (arr || []).sort().join(";"));
                                diffMap[proto] = new Set(sets).size > 1;
                            }

                            // 如果某条路径没有某协议，也算差异
                            const allProtocols = Object.keys(protocolGroups);
                            allPorts.forEach(item => {
                                allProtocols.forEach(proto => {
                                const hasProto = item.ports.some(p => p.protocol === proto);
                                if (!hasProto) diffMap[proto] = true;
                                });
                            });

                            const hasDiff = Object.values(diffMap).some(v => v);
                            result.hasInternalDiff = hasDiff;
                            if (!hasDiff) return null;

                            return (
                                <StatusModule label="机制内部配置差异性" hasIssue={hasDiff}>
                                <div
                                    style={{
                                    margin: "4px 0 10px 0",
                                    padding: "6px",
                                    backgroundColor: "#eef6f7",
                                    borderRadius: "4px",
                                    fontSize: "0.85rem",
                                    color: "#333"
                                    }}
                                >
                                    通过比较该机制下不同路径的配置端口和协议，发现部分协议或端口存在不一致。这可能会影响邮件客户端的兼容性或安全性。
                                </div>

                                {allPorts.map((item, idx) => (
                                    <div
                                    key={idx}
                                    style={{
                                        marginBottom: "6px",
                                        padding: "6px",
                                        border: "1px solid #ccc",
                                        borderRadius: "6px",
                                        backgroundColor: "#f9f9f9"
                                    }}
                                    >
                                    <strong>路径 {idx + 1}</strong>
                                    <ul style={{ margin: "4px 0 0 16px" }}>
                                        {item.ports.map((p, i) => {
                                        const proto = p.protocol;
                                        const isDiff = diffMap[proto] === true;
                                        return (
                                            <li
                                            key={i}
                                            style={{
                                                marginBottom: "4px",
                                                backgroundColor: isDiff ? "#fff3cd" : "transparent",
                                                padding: isDiff ? "4px" : "0",
                                                borderRadius: "4px"
                                            }}
                                            >
                                            {p.protocol} → {p.host}:{p.port} ({p.ssl || "未知 SSL"})
                                            {isDiff && (
                                                <span style={{ color: "#856404", marginLeft: "6px" }}>
                                                ⚠️ 与其他路径不一致
                                                </span>
                                            )}
                                            </li>
                                        );
                                        })}
                                    </ul>
                                    </div>
                                ))}
                                </StatusModule>
                            );
                        })()}





                        {/* 9.12删原来的配置信息概况 */}
                    </div>
                )}
                {mech === "srv" && result.srv_records && (
                    <div style={{ marginTop: "2rem" }}>
                        {/* 🔌 配置信息概况 */}
                        {Array.isArray(portsUsage) && portsUsage.length > 0 && (
                        <div style={{ marginBottom: "2rem" }}>
                            <div
                            style={{
                                borderTop: "2px solid #333",
                                paddingTop: "10px",
                                marginBottom: "20px",
                                display: "flex",
                                alignItems: "center",
                                // justifyContent: "center",
                                gap: "10px",
                            }}
                            >
                            <span style={{ fontSize: "32px" }}>🔌</span>
                            <h3 style={{ margin: 0, color: "#333" }}>最佳配置信息</h3>
                            </div>

                            <div
                            style={{
                                display: "flex",
                                flexWrap: "wrap",
                                gap: "1rem",
                                justifyContent: "center",
                                maxWidth: "1000px",
                                margin: "0 auto",
                            }}
                            >
                            {portsUsage.map((item, idx) => (
                                <div
                                key={idx}
                                style={{
                                    backgroundColor: "#f9f9f9",
                                    color: "#333",
                                    padding: "1rem",
                                    borderRadius: "12px",
                                    boxShadow: "0 2px 8px rgba(85, 136, 207, 0.1)",
                                    border: "1px solid #e0e0e0",
                                    minWidth: "220px",
                                    maxWidth: "280px",
                                    flex: "1",
                                }}
                                >
                                <table style={{ width: "100%", borderCollapse: "collapse" }}>
                                    <tbody>
                                    <tr>
                                        <td style={tdStyle}><strong>协议</strong></td>
                                        <td style={tdStyle}>{item.protocol}</td>
                                    </tr>
                                    <tr>
                                        <td style={tdStyle}><strong>端口</strong></td>
                                        <td style={tdStyle}>{item.port}</td>
                                    </tr>
                                    <tr>
                                        <td style={tdStyle}><strong>主机名</strong></td>
                                        <td style={tdStyle}>{item.host}</td>
                                    </tr>
                                    <tr>
                                        <td style={tdStyle}><strong>SSL类型</strong></td>
                                        <td style={tdStyle}>{item.ssl}</td>
                                    </tr>
                                    <tr>
                                        <td style={tdStyle}><strong>用户名</strong></td>
                                        <td style={tdStyle}>{lastSubmittedEmail}</td>
                                    </tr>
                                    <tr>
                                        <td style={tdStyle}><strong>密码</strong></td>
                                        <td style={tdStyle}>你的邮箱密码</td>
                                    </tr>
                                    </tbody>
                                </table>
                                </div>
                            ))}
                            </div>
                        </div>
                        )}
                        {/* 🔌 SRV 记录 */}
                        <div style={{ marginTop: "2rem" }}>
                        <div
                            style={{
                            borderTop: "2px solid #333",
                            paddingTop: "10px",
                            marginBottom: "20px",
                            display: "flex",
                            alignItems: "center",
                            }}
                        >
                            <span style={{ fontSize: "32px", marginRight: "10px" }}>🌐</span>
                            <h3 style={{ margin: 0, color: "#333" }}>SRV 记录</h3>
                        </div>

                        <table style={{ width: "100%", borderCollapse: "collapse", textAlign: "left" }}>
                            <thead style={{ background: "#f2f2f2" }}>
                            <tr>
                                <th style={tdStyle}>记录类型</th>
                                <th style={tdStyle}>服务标签</th>
                                <th style={tdStyle}>优先级</th>
                                <th style={tdStyle}>权重</th>
                                <th style={tdStyle}>端口</th>
                                <th style={tdStyle}>邮件服务器</th>
                            </tr>
                            </thead>
                            <tbody>
                            {Array.isArray(result.srv_records.recv) &&
                                result.srv_records.recv.map((item, idx) => (
                                <tr 
                                    key={`recv-${idx}`}
                                    style={{ backgroundColor: idx % 2 === 0 ? "#f9f9f9" : "#ffffff" }}
                                >
                                    <td style={tdStyle}>📥 Recv</td>
                                    <td style={tdStyle}>{item.Service}</td>
                                    <td style={tdStyle}>{item.Priority}</td>
                                    <td style={tdStyle}>{item.Weight}</td>
                                    <td style={tdStyle}>{item.Port}</td>
                                    <td style={tdStyle}>{item.Target}</td>
                                </tr>
                                ))}
                            {Array.isArray(result.srv_records.send) &&
                                result.srv_records.send.map((item, idx) => (
                                <tr 
                                    key={`send-${idx}`}
                                    style={{ backgroundColor: idx % 2 === 0 ? "#f9f9f9" : "#ffffff" }}
                                >
                                    <td style={tdStyle}>📤 Send</td>
                                    <td style={tdStyle}>{item.Service}</td>
                                    <td style={tdStyle}>{item.Priority}</td>
                                    <td style={tdStyle}>{item.Weight}</td>
                                    <td style={tdStyle}>{item.Port}</td>
                                    <td style={tdStyle}>{item.Target}</td>
                                </tr>
                                ))}
                            </tbody>
                        </table>
                        </div>


                        {/* 原始 SRV JSON */}
                        <div style={{
                            borderTop: "2px solid #333",
                            paddingTop: "10px",
                            marginBottom: "20px",
                            marginTop: "20px",
                            display: "flex",
                            alignItems: "center"
                        }}>
                            <span style={{ fontSize: "32px", marginRight: "10px" }}>📄</span>
                            <h3 style={{ margin: 0, color: "#333" }}>原始 SRV 记录</h3>
                        </div>

                        <pre style={{
                            background: "#f8f9fa",
                            color: "#2c3e50",
                            padding: "12px",
                            borderRadius: "6px",
                            fontSize: "14px",
                            overflowX: "auto"
                        }}>
                            {JSON.stringify(result.srv_records, null, 2)}
                        </pre>

                        {/* DNS 信息 */}
                        {result.dns_record && (
                        <>
                            <div
                            style={{
                                borderTop: "2px solid #333",
                                paddingTop: "10px",
                                marginTop: "20px",
                                marginBottom: "15px",
                                display: "flex",
                                alignItems: "center",
                            }}
                            >
                            <span style={{ fontSize: "32px", marginRight: "10px" }}>🌐</span>
                            <h3 style={{ margin: 0, color: "#333" }}>DNS 信息</h3>
                            </div>

                            <div
                            style={{
                                backgroundColor: "#f8f9fa",
                                padding: "1rem",
                                borderRadius: "12px",
                                boxShadow: "0 2px 8px rgba(0, 0, 0, 0.05)",
                                border: "1px solid #ddd",
                                overflowX: "auto", // 防止字段太多撑破
                            }}
                            >
                            <table
                                style={{
                                width: "100%",
                                borderCollapse: "collapse",
                                textAlign: "left",
                                }}
                            >
                                <thead style={{ background: "#f2f2f2" }}>
                                <tr>
                                    {Object.keys(result.dns_record).map((k) => (
                                    <th
                                        key={k}
                                        style={{
                                        padding: "8px 10px",
                                        borderBottom: "1px solid #ddd",
                                        fontWeight: "bold",
                                        color: "#2c3e50",
                                        }}
                                    >
                                        {dnsFieldMap[k] || k}
                                    </th>
                                    ))}
                                </tr>
                                </thead>
                                <tbody>
                                <tr>
                                    {Object.values(result.dns_record).map((v, idx) => (
                                    <td
                                        key={idx}
                                        style={{
                                        padding: "8px 10px",
                                        borderBottom: "1px solid #ddd",
                                        color: "#34495e",
                                        }}
                                    >
                                        {String(v)}
                                    </td>
                                    ))}
                                </tr>
                                </tbody>
                            </table>
                            </div>
                        </>
                        )}
                    </div>
                )}

                {mech === "guess" && result.score_detail?.ports_usage?.length > 0 && (
                <div style={{ marginTop: "2rem" }}>
                    {/* 分界线 + 标题 */}
                    <div style={{
                    borderTop: "2px solid #333",
                    paddingTop: "10px",
                    marginBottom: "20px",
                    display: "flex",
                    alignItems: "center"
                    }}>
                    <span style={{ fontSize: "32px", marginRight: "10px" }}>🔍</span>
                    <h3 style={{ margin: 0, color: "#333" }}>猜测到的可用邮件服务器</h3>
                    </div>

                    <p style={{ color: "#555", marginBottom: "1rem", lineHeight: "1.6" }}>
                        （当前，若邮件客户端无法通过实时查询得到目标邮件域的配置信息时，普遍会使用启发式方法进行连接测试，例如在目标邮件域前添加相关协议前缀（例如，imap，pop 和 smtp）形式以构造待测主机名。以下结果为本工具执行启发式猜测能够建立 TCP 连接的主机名和端口号信息以及支持相应方法的客户端。）
                    </p>


                    {/* 卡片容器 */}
                    <div style={{
                    display: "flex",
                    flexWrap: "wrap",
                    gap: "1rem"
                    }}>
                    {result.score_detail.ports_usage.map((item, idx) => (
                        <div
                        key={idx}
                        style={{
                            backgroundColor: "#f8f9fa",
                            padding: "1rem",
                            borderRadius: "12px",
                            boxShadow: "0 2px 8px rgba(0, 0, 0, 0.05)",
                            border: "1px solid #ddd",
                            minWidth: "220px",
                            maxWidth: "220px",
                            flex: "1"
                        }}
                        >
                        <table style={{ width: "100%", borderCollapse: "collapse" }}>
                            <tbody>
                            <tr>
                                <td style={{
                                padding: "6px 10px",
                                borderBottom: "1px solid #ddd",
                                fontWeight: "bold",
                                color: "#2c3e50",
                                width: "35%"
                                }}>
                                主机
                                </td>
                                <td style={{
                                padding: "6px 10px",
                                borderBottom: "1px solid #ddd",
                                color: "#34495e"
                                }}>
                                {item.host}
                                </td>
                            </tr>
                            <tr>
                                <td style={{
                                padding: "6px 10px",
                                fontWeight: "bold",
                                color: "#2c3e50"
                                }}>
                                端口
                                </td>
                                <td style={{
                                padding: "6px 10px",
                                color: "#34495e"
                                }}>
                                {item.port}
                                </td>
                            </tr>
                            <tr>
                                <td
                                    style={{
                                    padding: "6px 10px",
                                    fontWeight: "bold",
                                    whiteSpace: "nowrap",
                                    color: "#2c3e50",
                                    }}
                                >
                                    客户端
                                </td>
                                <td
                                    style={{
                                    padding: "6px 10px",
                                    color: "#34495e",
                                    }}
                                >
                                    {detectMailAppsSmart(item.host)}
                                </td>
                            </tr>

                            </tbody>
                        </table>
                        </div>
                    ))}
                    </div>
                </div>
                )}

                {/* 连接详情跳转 */}
                {["srv", "guess"].map(type => (
                    mech === type && result.score_detail?.actualconnect_details && (
                        <button
                            key={type}
                            onClick={() => handleViewDetailsClick(type, result.score_detail.actualconnect_details)}
                            style={viewButtonHoverStyle}
                        >
                            查看连接详情({type.toUpperCase()})
                        </button>
                    )
                ))}

                {/* 折叠主观分析9.17 */}
                {mech === "srv" && (
                <>
                    {/* 分界线标题 */}
                    <div
                        style={{
                            borderTop: "2px solid #333",
                            paddingTop: "10px",
                            marginTop: "30px",
                            marginBottom: "20px",
                            display: "flex",
                            alignItems: "center",
                        }}
                    >
                        <span style={{ fontSize: "28px", marginRight: "10px" }}>🛡️</span>
                        <h3 style={{ margin: 0, color: "#333" }}>SRV 机制风险分析</h3>
                    </div>

                    
                        <>
                            {/* SRV 风险分析模块 */}
                            {result && (() => {
                                console.log("=== SRV result ===", result); 
                                const dns = result?.dns_record;
                                console.log("=== SRV dns_record ===", dns);  
                                if (!dns) return <p>无有效 SRV 记录</p>;

                                const adBits = {
                                    IMAP: dns.ADbit_imap,
                                    IMAPS: dns.ADbit_imaps,
                                    POP3: dns.ADbit_pop3,
                                    POP3S: dns.ADbit_pop3s,
                                    SMTP: dns.ADbit_smtp,
                                    SMTPS: dns.ADbit_smtps,
                                };

                                // ====== DNSSEC 总体状态 ======
                                const adBitsValues = Object.values(adBits);
                                let hasDnsIssue = false;
                                const allSkipped = adBitsValues.every(v => v === null || v === undefined);
                                if (!allSkipped) {
                                    hasDnsIssue = adBitsValues.some(v => v === false);
                                }

                                return (
                                    <div>
                                        {/* DNSSEC 风险分析 */}
                                        <StatusModule label="查询过程安全性分析" hasIssue={hasDnsIssue}>
                                            <div style={{
                                                margin: "4px 0 10px 0",
                                                padding: "6px",
                                                backgroundColor: "#eef6f7",
                                                borderRadius: "4px",
                                                fontSize: "0.85rem",
                                                color: "#333"
                                            }}>
                                                SRV 机制主要依赖 DNS 记录获取邮件服务器信息。通过 DNSSEC 检查，可以判断配置可靠性，防止 DNS 劫持或篡改。
                                            </div>

                                            <div style={{ marginTop: "6px" }}>
                                                <h4 style={{ marginBottom: "6px" }}>DNSSEC 检查结果</h4>
                                                <ul style={{ margin: 0, paddingLeft: "18px", color: "#333" }}>
                                                    {Object.entries(adBits).map(([proto, bit], idx) => {
                                                        let statusText = bit === true ? "✅ DNSSEC 有效" :
                                                            bit === false ? "❌ DNSSEC 无效" :
                                                                "⚪ 未检测到结果";
                                                        return <li key={idx}>{proto} {statusText}</li>;
                                                    })}
                                                </ul>
                                            </div>
                                        </StatusModule>

                                        {/* 实际连接安全性 - SRV */}
                                        <StatusModule label="实际连接安全性" hasIssue={(() => {
                                            const details = result?.score_detail?.actualconnect_details || [];
                                            return details.some(d => d.plain?.success); // 明文可连接即认为有问题
                                        })()}>
                                            {(result?.score_detail?.actualconnect_details || []).map((d, idx) => (
                                                <div key={idx} style={{
                                                    marginBottom: "6px",
                                                    padding: "6px",
                                                    border: "1px solid #ccc",
                                                    borderRadius: "6px",
                                                    backgroundColor: "#f9f9f9",
                                                }}>
                                                    <strong>{d.host}</strong>
                                                    <ul style={{ margin: "4px 0 0 16px" }}>
                                                        <li>
                                                            {d.type.toUpperCase()} : {d.port} →
                                                            {d.plain?.success && <span style={{ color: "red", marginLeft: "8px" }}>⚠️ 明文可连接</span>}
                                                            {(d.tls?.success || d.starttls?.success) && <span style={{ color: "green", marginLeft: "8px" }}>✅ 安全连接可用</span>}
                                                            {!d.plain?.success && !d.tls?.success && !d.starttls?.success && <span style={{ color: "gray", marginLeft: "8px" }}>❌ 无法连接</span>}

                                                            {showAnalyzerMap[`${d.host}-${d.port}`] && (
                                                                <div style={{ marginTop: "6px" }}>
                                                                    <TlsAnalyzerPanel host={d.host} port={d.port} />
                                                                </div>
                                                            )}
                                                        </li>
                                                    </ul>
                                                </div>
                                            ))}

                                            {/* ✅ 明文可连接注释，放在框框最下面 */}
                                            {result?.score_detail?.actualconnect_details?.some(d => d.plain?.success) && (
                                                <div style={{
                                                    marginTop: "8px",
                                                    padding: "6px",
                                                    backgroundColor: "#fff3cd",
                                                    border: "1px solid #ffeeba",
                                                    borderRadius: "6px",
                                                    color: "#856404",
                                                    fontSize: "14px",
                                                }}>
                                                    注：判断服务器是否支持明文连接需通过实际登录验证，本工具仅限于TCP/TLS连接阶段的探测，未执行任何登录尝试，仅给出风险提示。
                                                </div>
                                            )}
                                        </StatusModule>
                                    </div>
                                );
                            })()}
                        </>
                    </>
                )}


            </div>
        );
    };
    const hasAnyResult = Object.values(results).some((r) => r && Object.keys(r).length > 0);{/*7.28 */}
    

    return (
        <div
            style={{
            marginTop: "10vh",
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            padding: "0 1rem",
            }}
        >
            <h1 style={{fontSize: isMobile ? "1.8rem" : "2.5rem", marginBottom: "1rem",  color: "#1f2d3d", textShadow: "0 0 4px rgba(255,255,255,0.2)", textAlign: "center",}}>
            邮件服务通信安全检测
            </h1>

            <div
                // style={{
                //     maxWidth: "900px",
                //     width: "100%",
                //     display: "flex",
                //     justifyContent: "center",
                //     margin: "0 auto",
                // }}
                style={{
                    maxWidth: "900px",
                    width: "95%",
                    display: "flex",
                    flexDirection: isMobile ? "column" : "row", // 手机竖排，桌面横排
                    justifyContent: "center",
                    alignItems: "center",
                    gap: "0.5rem", // 横向或纵向间距
                    margin: "0 auto",
                }}
            >
                {/* 📮 输入框 + CSV图标 一体容器 */}
                <div style={{ position: "relative", display: "flex", alignItems: "center", width: isMobile ? "100%" : "auto", marginBottom: isMobile ? "0.5rem" : "0"}}>
                    <input
                        type="text"
                        value={email}
                        onChange={handleChange}
                        placeholder="输入邮件地址：如 user@example.com"
                        style={{
                            height: "56px",
                            width: isMobile ? "100%" : "400px", // 手机端宽度 100%
                            fontSize: "1.2rem",
                            borderRadius: "8px",
                            border: "1px solid #ccc",
                            outline: "none",
                            color: "#000",
                            padding: "0 1rem 0 1rem", // 🔹右侧留出图标空间
                            boxSizing: "border-box",
                        }}
                    />

                    {/* 📎 CSV 上传图标嵌入输入框右侧 */}
                    <div
                        style={{
                            position: "absolute",
                            right: "10px",
                            top: "50%",
                            transform: "translateY(-50%)",
                            cursor: "pointer",
                            display: "flex",
                            alignItems: "center",
                        }}
                    >
                        <CSVUploadForm compact={true} />
                    </div>

                    {/* 下拉框定位相对于输入框 */}
                    {suggestions.length > 0 && (
                        <div
                            style={{
                                position: "absolute",
                                top: "100%",
                                left: 0,
                                width: "100%",
                                background: "#fff",
                                border: "1px solid #ccc",
                                maxHeight: "150px",
                                overflowY: "auto",
                                borderRadius: "6px",
                                marginTop: "4px",
                                boxShadow: "0 2px 6px rgba(0,0,0,0.15)",
                                zIndex: 10,
                            }}
                        >
                            {suggestions.map((s, idx) => (
                                <div
                                    key={idx}
                                    onClick={() => handleSelect(s)}
                                    style={{
                                        padding: "8px 12px",
                                        cursor: "pointer",
                                        borderBottom:
                                            idx !== suggestions.length - 1
                                                ? "1px solid #eee"
                                                : "none",
                                    }}
                                    onMouseOver={(e) =>
                                        (e.currentTarget.style.background = "#f5f5f5")
                                    }
                                    onMouseOut={(e) =>
                                        (e.currentTarget.style.background = "white")
                                    }
                                >
                                    {s}
                                </div>
                            ))}
                        </div>
                    )}
                </div>

                {/* 🚀 开始检测按钮单独放右边 */}
                <button
                    //onClick={handleClick}
                    onClick={() => handleClick(null, email.trim(), false)} // 手动输入10.30
                    // style={{
                    //     height: "56px",
                    //     lineHeight: "56px",
                    //     marginLeft: "1rem",
                    //     fontSize: "1.2rem",
                    //     borderRadius: "8px",
                    //     backgroundColor: "#3c71cd",
                    //     color: "white",
                    //     border: "none",
                    //     cursor: "pointer",
                    //     fontWeight: "bold",
                    //     transition: "background 0.3s",
                    //     padding: "0 1.5rem",
                    // }}
                    style={{
                        height: "56px",
                        lineHeight: "56px",
                        width: isMobile ? "100%" : "auto", // 手机端按钮宽度撑满
                        fontSize: "1.2rem",
                        borderRadius: "8px",
                        backgroundColor: "#3c71cd",
                        color: "white",
                        border: "none",
                        cursor: "pointer",
                        fontWeight: "bold",
                        transition: "background 0.3s",
                        padding: "0 1.5rem",
                    }}
                    onMouseOver={(e) =>
                        (e.target.style.backgroundColor = "#2e4053")
                    }
                    onMouseOut={(e) =>
                        (e.target.style.backgroundColor = "#3c71cd")
                    }
                >
                    开始检测
                </button>
            </div>


                {/* 批量检测组件 */}
                {/* <CSVUploadForm hideTitle={true} buttonPadding="1rem 1.2rem" /> */}

                {/* 推荐域名9.18_2 */}
                {/* <div style={{
                    marginTop: "2rem",
                    width: "100%",
                    display: "flex",
                    justifyContent: "center",
                    gap: "1rem",
                    flexWrap: "wrap"
                }}>
                    {recommendedDomains.map(item => (
                        <div
                            key={item.domain}
                            onClick={() => handleClickRecommended(item.domain)}
                            style={{
                                padding: "0.6rem 1rem",
                                borderRadius: "8px",
                                backgroundColor: "#3c71cdff",
                                color: "white",
                                cursor: "pointer",
                                fontWeight: "bold",
                                transition: "background 0.3s",
                            }}
                            onMouseOver={e => e.currentTarget.style.backgroundColor = "#2e4053"}
                            onMouseOut={e => e.currentTarget.style.backgroundColor = "#3c71cdff"}
                        >
                            {item.domain}
                        </div>
                    ))}
                </div> */}

            {/* </div> */}

        {/* 9.23 */}
        {/* 推荐域名区域 - 纯文字风格，每行固定四个 */}
        <div
            style={{
                width: "100%",
                maxWidth: "900px",
                marginTop: "2rem",
                padding: "1.5rem 2rem",
                borderRadius: "16px",
                backgroundColor: "rgba(249, 249, 249, 0.95)", // 🔹稍微透明
                border: "1px solid rgba(200, 200, 200, 0.5)",  // 柔和边框
                boxShadow: "0 6px 18px rgba(0, 0, 0, 0.12)",
                color: "#000",                         // 文字颜色改为黑色
            }}
        >
                <h3
                    style={{
                        marginBottom: "1rem",
                        color: "#333",                     // 标题改为深色
                        fontWeight: "600",
                        fontSize: isMobile ? "1rem" : "1.2rem", // 手机字体稍小
                        letterSpacing: "0.5px",
                        textShadow: "0 0 2px rgba(0,0,0,0.1)", // 轻微阴影增强立体感
                    }}
                >
                    🔹 推荐邮件域名
                </h3>

                <div
                    style={{
                        display: "flex",
                        flexWrap: "wrap",
                        gap: "1.5rem",
                        justifyContent: "flex-start",
                    }}
                >
                    {recommended.map((item, idx) => (
                        <span
                            key={idx}
                            onClick={() => {
                                //handleClick(null, "test@" + item.domain);
                                handleClick(null, "test@" + item.domain, true); // 推荐点击10.30
                                setActiveDomain(item.domain); // 🔹记录当前点击的域名
                            }}
                            style={{
                                cursor: "pointer",
                                fontSize: "1.1rem",
                                color: activeDomain === item.domain ? "#000" : "#1a1a1a", // 被选中更深色
                                textDecoration: "underline",
                                transition: "all 0.2s ease-in-out",
                                width: "calc(25% - 1.5rem)",     // 每行四个
                                textAlign: "center",
                                padding: "0.4rem 0",
                                borderRadius: "8px",
                                backgroundColor:
                                    activeDomain === item.domain
                                        ? "rgba(200,220,255,0.5)"  // 点击后保持背景色
                                        : "transparent",
                            }}
                            onMouseOver={(e) => {
                                if (activeDomain !== item.domain) {
                                    e.currentTarget.style.color = "#000";
                                    e.currentTarget.style.backgroundColor = "rgba(200,220,255,0.3)";
                                }
                            }}
                            onMouseOut={(e) => {
                                if (activeDomain !== item.domain) {
                                    e.currentTarget.style.color = "#1a1a1a";
                                    e.currentTarget.style.backgroundColor = "transparent";
                                }
                            }}
                        >
                            {item.domain}
                        </span>
                    ))}
                </div>
            </div>




            {loading && (
                <div style={{ marginTop: "2rem", textAlign: "center" }}>
                    <div style={spinnerStyle}></div>
                    <p style={{ marginTop: "1rem", color: "#555" }}>
                    {stage} - {progressMessage}
                    </p>
                </div>
            )}



            {errors && <p style={{ color: "red" }}>{errors}</p>}

            {hasAnyResult && (
                <>
                    {/*10.8 <div style={{ width: "100%", maxWidth: "900px", backgroundColor: "#f5f8fa", padding: "2rem", borderRadius: "12px", boxShadow: "0 2px 10px rgba(0,0,0,0.04)", border: "1px solid #eee", marginTop: "1rem" }}> */}
                    <div
                        style={{
                            width: "100%",
                            maxWidth: "900px",
                            backgroundColor: "rgba(255, 255, 255, 0.9)", // ← 更接近纯白，可读性强
                            padding: "2rem",
                            borderRadius: "16px",
                            boxShadow: "0 4px 20px rgba(0,0,0,0.08)",
                            border: "1px solid rgba(200, 200, 200, 0.5)",
                            marginTop: "1rem",
                            color: "#000", // 内容文字颜色调整为黑色
                        }}
                    >


    
                        {/* 机制 Tab9.17 */}
                        <div style={{ display: "flex", gap: "1rem", marginBottom: "1rem", flexWrap: "wrap" }}>
                        {["overview", ...mechanisms.filter(m => m !== "overview")].map((mech) => (
                                <div
                                    key={mech}
                                    onClick={() => setCurrentMech(mech)}
                                    style={{
                                        padding: "0.8rem 1.2rem",
                                        borderRadius: "10px",
                                        cursor: "pointer",
                                        backgroundColor: currentMech === mech ? "#e3edf5" : "#f9f9f9",
                                        color: currentMech === mech ? "#3a506b" : "#888",
                                        border: currentMech === mech ? "2px solid #8aa3b4" : "1px solid #ddd",
                                        boxShadow: currentMech === mech ? "0 2px 6px rgba(138,163,180,0.4)" : "none",
                                        transition: "all 0.2s ease-in-out",
                                        minWidth: "120px",
                                        textAlign: "center",
                                        fontWeight: 600,
                                        letterSpacing: "0.5px"
                                    }}
                                >
                                    {mech.toUpperCase()}
                                </div>
                            ))}
                        </div>

                        {/* 机制内容 */}
                        {renderMechanismContent(currentMech)}
                    </div>
                </>
            )}

            {/* <CSVUploadForm /> */}

            <h2 style={{ marginTop: "3rem", color: "#333" }}>历史查询</h2>
            {recentlySeen.length > 0 ? (
                <ul>
                    {recentlySeen.map((item, index) => (
                        <li key={index} style={{ color: "#e0e8ff" }}>
                            <strong>{item.domain}</strong> - Score: {item.score}, Grade: {item.grade}, Time:{" "}
                            {new Date(item.timestamp).toLocaleString()}
                        </li>
                    ))}
                </ul>
            ) : (
                <p style={{ color: "#333" }}>暂无记录</p>
            )}
        </div>
    );

}


//9.16
// function CSVUploadForm() {
//     const [downloadUrl, setDownloadUrl] = useState(null);
//     const [isUploading, setIsUploading] = useState(false);

//     const handleUpload = async (e) => {
//         const file = e.target.files[0];
//         if (!file) return;

//         setIsUploading(true);
//         setDownloadUrl(null);

//         const formData = new FormData();
//         formData.append("file", file);

//         try {
//             const res = await fetch(`/api/uploadCsvAndExportJsonl`, {
//             method: "POST",
//             body: formData,
//             mode: "cors",               // 显式允许跨域9.6
//             credentials: "omit",        // 如果不需要带 cookie
//             });


//             if (!res.ok) {
//                 throw new Error("Upload failed");
//             }

//             const data = await res.json();
//             setDownloadUrl(data.download_url);
//         } catch (err) {
//             alert("上传失败：" + err.message);
//         } finally {
//             setIsUploading(false);
//         }
//     };

//     const handleDownload = async () => {
//         try {
//             const res = await fetch(`${downloadUrl}`);
//             if (!res.ok) {
//                 throw new Error("下载失败");
//             }

//             const blob = await res.blob();
//             const url = window.URL.createObjectURL(blob);
//             const a = document.createElement("a");
//             a.href = url;
//             a.download = "result.jsonl"; // 可以改成动态文件名
//             document.body.appendChild(a);
//             a.click();
//             a.remove();
//             window.URL.revokeObjectURL(url);
//         } catch (err) {
//             alert("下载失败：" + err.message);
//         }
//     };

//     return (
//         <div style={{ marginBottom: "30px", padding: "20px", textAlign: "center" }}>
//             <h3 style={{ fontSize: "1.5rem", marginBottom: "1rem", color: "#29394dff" }}>📄 批量域名检测</h3>
            
//             <label 
//                 style={{ 
//                     display: "inline-block",
//                     padding: "10px 20px",
//                     backgroundColor: "#5daed7ff",
//                     color: "white",
//                     borderRadius: "8px",
//                     cursor: "pointer",
//                     fontWeight: "bold",
//                     transition: "background 0.3s"
//                 }}
//                 onMouseOver={(e) => (e.target.style.backgroundColor = "#3c71cdff")}
//                 onMouseOut={(e) => (e.target.style.backgroundColor = "#6d92cbff")}
//             >
//                 选择 CSV 文件
//                 <input 
//                     type="file" 
//                     accept=".csv" 
//                     onChange={handleUpload} 
//                     style={{ display: "none" }}
//                 />
//             </label>

//             {isUploading && <p style={{ marginTop: "1rem", color: "#888" }}>⏳ 处理中，请稍等...</p>}

//             {downloadUrl && (
//                 <p style={{ marginTop: "1rem" }}>
//                     ✅ 查询完成，
//                     <button 
//                         onClick={handleDownload}
//                         style={{
//                             marginLeft: "10px",
//                             padding: "8px 16px",
//                             backgroundColor: "#3a506b",
//                             color: "white",
//                             border: "none",
//                             borderRadius: "6px",
//                             cursor: "pointer",
//                             fontWeight: "bold",
//                             transition: "background 0.3s"
//                         }}
//                         onMouseOver={(e) => (e.target.style.backgroundColor = "#2e4053")}
//                         onMouseOut={(e) => (e.target.style.backgroundColor = "#3a506b")}
//                     >
//                         点击下载结果
//                     </button>
//                 </p>
//             )}
//         </div>
//     );


// }

// 9.16
// function CSVUploadForm() {
//     const [downloadUrl, setDownloadUrl] = useState(null);
//     const [isUploading, setIsUploading] = useState(false);

//     const handleUpload = async (e) => {
//         const file = e.target.files[0];
//         if (!file) return;

//         setIsUploading(true);
//         setDownloadUrl(null);

//         const formData = new FormData();
//         formData.append("file", file);

//         try {
//             const res = await fetch(`/api/uploadCsvAndExportJsonl`, {
//             method: "POST",
//             body: formData,
//             mode: "cors",               // 显式允许跨域9.6
//             credentials: "omit",        // 如果不需要带 cookie
//             });

//             if (!res.ok) throw new Error("Upload failed");

//             const data = await res.json();
//             setDownloadUrl(data.download_url);
//         } catch (err) {
//             alert("上传失败：" + err.message);
//         } finally {
//             setIsUploading(false);
//         }
//     };

//     const handleDownload = async () => {
//         try {
//             const res = await fetch(`${downloadUrl}`);
//             if (!res.ok) throw new Error("下载失败");

//             const blob = await res.blob();
//             const url = window.URL.createObjectURL(blob);
//             const a = document.createElement("a");
//             a.href = url;
//             a.download = "result.jsonl";
//             document.body.appendChild(a);
//             a.click();
//             a.remove();
//             window.URL.revokeObjectURL(url);
//         } catch (err) {
//             alert("下载失败：" + err.message);
//         }
//     };

//     return (
//         <div style={{ marginLeft: "1rem", display: "flex", alignItems: "center" }}>
//             {!downloadUrl && !isUploading && (
//                 <label
//                     title="上传域名列表（.csv 格式）进行批量检测"
//                     style={{
//                         display: "inline-block",
//                         height: "56px",
//                         lineHeight: "56px",
//                         fontSize: "1.2rem",
//                         fontWeight: "bold",
//                         borderRadius: "8px",
//                         backgroundColor: "#365289",
//                         color: "white",
//                         cursor: "pointer",
//                         whiteSpace: "nowrap",
//                         transition: "background 0.3s",
//                         padding: "0 1.5rem",
//                     }}
//                     onMouseOver={(e) => (e.target.style.backgroundColor = "#2e4053")}
//                     onMouseOut={(e) => (e.target.style.backgroundColor = "#365289")}
//                 >
//                     批量检测
//                     <input
//                         type="file"
//                         accept=".csv"
//                         onChange={handleUpload}
//                         style={{ display: "none" }}
//                     />
//                 </label>
//             )}

//             {isUploading && (
//                 <span style={{ marginLeft: "10px", color: "#888", fontSize: "0.95rem" }}>
//                     ⏳ 批量检测处理中...
//                 </span>
//             )}

//             {downloadUrl && !isUploading && (
//                 <>
//                     <button
//                         onClick={handleDownload}
//                         style={{
//                             padding: "1rem",
//                             fontSize: "1.2rem",
//                             fontWeight: "bold",
//                             borderRadius: "8px",
//                             backgroundColor: "#2e4053",
//                             color: "white",
//                             border: "none",
//                             cursor: "pointer",
//                             transition: "background 0.3s",
//                         }}
//                         title="点击下载检测结果"
//                         onMouseOver={(e) => (e.target.style.backgroundColor = "#1f2a3d")}
//                         onMouseOut={(e) => (e.target.style.backgroundColor = "#2e4053")}
//                     >
//                         ⬇️ 下载检测结果
//                     </button>
//                     <button
//                         onClick={() => setDownloadUrl(null)}
//                         style={{
//                             marginLeft: "10px",
//                             padding: "1rem",
//                             fontSize: "1.2rem",
//                             fontWeight: "bold",
//                             borderRadius: "8px",
//                             backgroundColor: "#888",
//                             color: "white",
//                             border: "none",
//                             cursor: "pointer",
//                             transition: "background 0.3s",
//                         }}
//                         title="重置批量检测"
//                         onMouseOver={(e) => (e.target.style.backgroundColor = "#555")}
//                         onMouseOut={(e) => (e.target.style.backgroundColor = "#888")}
//                     >
//                         🔄
//                     </button>
//                 </>
//             )}
//         </div>
//     );
// }

//10.19
function CSVUploadForm({ compact = false }) {
    const [downloadUrl, setDownloadUrl] = useState(null);
    const [isUploading, setIsUploading] = useState(false);

    const handleUpload = async (e) => {
        const file = e.target.files[0];
        if (!file) return;

        // ===== 新增：10.31前端限制 =====
        const MAX_SIZE = 30 * 1024; // 30KB
        const MAX_LINES = 1000;

        if (file.size > MAX_SIZE) {
            alert("文件过大，最大允许 30KB！");
            return;
        }

        // 读取文件内容，检查行数
        const text = await file.text();
        const lines = text.split(/\r?\n/).filter(l => l.trim() !== "");
        if (lines.length > MAX_LINES) {
            alert(`文件中包含 ${lines.length} 条记录，超过最大限制（1000 条）`);
            return;
        }

        setIsUploading(true);
        setDownloadUrl(null);

        const formData = new FormData();
        formData.append("file", file);

        try {
            const res = await fetch(`/api/uploadCsvAndExportJsonl`, {
                method: "POST",
                body: formData,
                mode: "cors",
                credentials: "omit",
            });

            if (!res.ok) throw new Error("Upload failed");
            const data = await res.json();
            setDownloadUrl(data.download_url);
        } catch (err) {
            alert("上传失败：" + err.message);
        } finally {
            setIsUploading(false);
        }
    };

    const handleDownload = async () => {
        try {
            const res = await fetch(`${downloadUrl}`);
            if (!res.ok) throw new Error("下载失败");

            const blob = await res.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.href = url;
            a.download = "result.jsonl";
            document.body.appendChild(a);
            a.click();
            a.remove();
            window.URL.revokeObjectURL(url);
        } catch (err) {
            alert("下载失败：" + err.message);
        }
    };

    if (compact) {
        // 🔹 内嵌图标模式（输入框右侧的小图标）
        return (
            <div 
                style={{ 
                    position: "absolute", 
                    right: "10px", 
                    top: "50%", 
                    transform: "translateY(-50%)",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center", // 水平居中图标
                    height: "100%",           // 高度撑满输入框
                }}
            >
                {!downloadUrl && !isUploading && (
                    <label 
                        title="上传域名列表（.csv 格式）进行批量检测" 
                        style={{ 
                            cursor: "pointer",
                            display: "flex",
                            alignItems: "center",
                            justifyContent: "center",
                            height: "100%", 
                        }}
                    >
                        <AiOutlineFileText size={22} />
                        <input type="file" accept=".csv" onChange={handleUpload} style={{ display: "none" }} />
                    </label>
                )}

                {isUploading && (
                    <span title="处理中..." style={{ fontSize: "0.9rem", color: "#888" }}>
                        ⏳
                    </span>
                )}

                {downloadUrl && !isUploading && (
                    <span>
                        <button
                            onClick={handleDownload}
                            title="下载检测结果"
                            style={{
                                background: "transparent",
                                border: "none",
                                color: "#2e4053",
                                fontSize: "1rem",
                                cursor: "pointer",
                            }}
                        >
                            ⬇️
                        </button>
                        <button
                            onClick={() => setDownloadUrl(null)}
                            title="重置"
                            style={{
                                background: "transparent",
                                border: "none",
                                color: "#888",
                                fontSize: "1rem",
                                cursor: "pointer",
                            }}
                        >
                            🔄
                        </button>
                    </span>
                )}
            </div>
        );
    }

}





export default MainPage;
