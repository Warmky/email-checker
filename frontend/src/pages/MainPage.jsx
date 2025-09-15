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


    const mechanisms = ["autodiscover", "autoconfig", "srv", "guess", "compare"];//9.10_2 新增加比较机制供管理者一眼看出不同机制得到的配置信息有何不同
    // 默认选中第一个有结果的机制（不含 compare）9.10_2
    const firstAvailable = mechanisms.find(m => m !== "compare" && results[m]) || mechanisms[0];
    const [currentMech, setCurrentMech] = useState(firstAvailable);
    const [lastSubmittedEmail, setLastSubmittedEmail] = useState("");

    //9.15 改造输入框
    const [suggestions, setSuggestions] = useState([]);

    const presetDomains = [
    "qq.com",
    "163.com",
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
    
    // 点击检测按钮
    const handleClick = () => {
        const targetEmail = email.trim();
        if (!targetEmail) return; // 空输入不处理
    
        handleSearch(targetEmail); // 调你原来的检测函数
        setLastSubmittedEmail(targetEmail); // 保存用户名用于展示
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
        borderRadius: "8px",
        overflow: "hidden",
        boxShadow: "0 4px 8px rgba(0,0,0,0.05)"
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

    const CollapsibleModule = ({ label, score, children }) => {
        const [open, setOpen] = useState(false);
        return (
            <div
                style={{
                    border: "1px solid #fff",  // 白色边框
                    borderRadius: "8px",
                    padding: "10px",
                    marginBottom: "10px",
                    backgroundColor: "#fff", // 可以是白底，如果页面背景深色，可以微调
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

    // 当前机制内容渲染函数7.28
    const renderMechanismContent = (mech) => {
        const result = results[mech];
        //9.11
        preprocessResults(results);

        if (mech === "compare") {
            const comparisonMap = comparePortsUsage(results); //这里比较的是不同机制间
    
            // 一致性评分
            let consistencyScore = 100;
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
            // 考虑机制内部不一致
            Object.entries(results).forEach(([mech, res]) => {
                if (res?.hasInternalDiff) {
                    consistencyScore = 30; // 🚨 内部不一致，优先判定为不一致
                }
            });

            
            //9.11_2
            // 证书与 DNSSEC 分数
            const certDnsScore = calculateCertDnsScore(results);
            // 综合评分
            const overallConfigScore = calculateOverallConfigScore(consistencyScore, certDnsScore);

            // // 配置获取过程安全性评分（取平均）
            // const mechScores = ["autodiscover", "autoconfig", "srv"]//这里应该是scores["cert_score"]
            //     .map(m => results[m]?.score?.overall || 0)
            //     .filter(s => s > 0);
            // const overallConfigScore = mechScores.length
            //     ? Math.round(mechScores.reduce((a, b) => a + b, 0) / mechScores.length)
            //     : 0;
    
            // 连接安全性（取最低/平均？）
            const connectScores = ["autodiscover", "autoconfig", "srv"]
                .map(m => results[m]?.score_detail?.connection?.Overall_Connection_Score || 0)
                .filter(s => s > 0);
            // const unifiedConnectScore = connectScores.length
            //     ? Math.min(...connectScores)
            //     : 0;
            const unifiedConnectScore = connectScores.length
                ? connectScores.reduce((a, b) => a + b, 0) / connectScores.length
                : 0;
    
            // 大评级框
            const gradeBox = (score) => (
                <div style={{
                    width: "100px",
                    height: "100px",
                    borderRadius: "10px",
                    border: "2px solid #333",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    fontSize: "28px",
                    fontWeight: "bold",
                    background: score >= 90? "#2ecc71" : score >= 50? "#f1c40f" : score >= 30? "#ff9800"  : "#e74c3c",    
                    color: "#fff",
                    marginRight: "20px"
                }}>
                    {getGrade(score)}
                </div>
            );


    
            return (
                <div style={{ marginTop: "2rem" }}>
                    <h3 style={{ marginBottom: "15px" }}>📊 Compare 总览</h3>
    
                {/* 上方主题分界线 */}
                <div style={{
                    borderTop: "2px solid #333",
                    paddingTop: "10px",
                    marginBottom: "20px",
                    display: "flex",
                    alignItems: "center"
                }}>
                    <span style={{ fontSize: "32px", marginRight: "10px" }}>🛡️</span>
                    <h3 style={{ margin: 0, color: "#333" }}>配置获取过程安全性</h3>
                </div>

                {/* 主体内容 */}
                <div style={{ display: "flex", alignItems: "flex-start", marginBottom: "20px" }}>
                    {gradeBox(overallConfigScore)} {/* 左边大评级框，增大尺寸 */}

                    {/* 右边两个模块 */}
                    <div style={{ flex: 1 }}>
                        {/* 上模块：配置信息差异性 9.11_2*/}
                        <CollapsibleModule label="配置信息差异性" score={consistencyScore}>
                        <ul style={{ margin: 0, paddingLeft: "18px", color: "#333" }}>
                            {/* 内部差异（逐个机制列出） */}
                            {Object.entries(results).map(([mech, res]) => {
                            if (res?.hasInternalDiff) {
                                return (
                                <li key={mech}>
                                    机制 <b>{mech}</b> 内部不同路径存在配置差异
                                </li>
                                );
                            }
                            return null;
                            })}

                            {/* 跨机制差异（只要 consistencyScore <= 50 就显示） */}
                            {consistencyScore <= 50 && (
                            <li>不同机制之间存在配置差异</li>
                            )}

                            {/* 完全一致（只有 100 分时显示） */}
                            {consistencyScore === 100 && (
                            <li>所有机制配置完全一致</li>
                            )}
                        </ul>
                        </CollapsibleModule>

                        {/* 下模块：配置获取过程安全性 */}
                        {/* 9.11_2 */}
                        <CollapsibleModule label="证书与 DNS 验证" score={overallConfigScore}>
                            {["autodiscover", "autoconfig"].map(m => {
                                const certInfo = results[m]?.cert_info;
                                if (!certInfo) return null;
                                const { issues } = getCertGrade(certInfo);
                                return (
                                    <div key={m} style={{ marginBottom: "10px" }}>
                                        <h4>{m.toUpperCase()} 机制配置服务器证书检测</h4>
                                        <ul style={{ margin: 0, paddingLeft: "18px", color: "#333" }}>
                                        {issues.length > 0 ? issues.map((i, idx) => <li key={idx}>{i}</li>) : <li>配置服务器返回的证书链完整，验证通过，连接信息正常。</li>}
                                        </ul>
                                    </div>
                                    );
                            })}

                            {results.srv && (
                                <div style={{ marginTop: "10px" }}>
                                    <h4>SRV 机制：DNSSEC 结果</h4>
                                    <ul style={{ margin: 0, paddingLeft: "18px", color: "#333" }}>
                                    {getDNSSummary(results.srv).details.map((d, idx) => (
                                        <li key={idx}>{d}</li>
                                    ))}
                                    </ul>
                                </div>
                            )}

                        </CollapsibleModule>
                    </div>
                </div>


                {/* 表格详情：直接显示，不折叠 */}
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

                

                {/* 连接安全性评级模块 */}
                <div style={{ marginTop: "20px" }}>
                    {/* 上方主题分界线 */}
                    <div style={{
                        borderTop: "2px solid #333",
                        paddingTop: "10px",
                        marginBottom: "10px",
                        display: "flex",
                        alignItems: "center"
                    }}>
                        <span style={{ fontSize: "32px", marginRight: "10px" }}>🔒</span>
                        <h3 style={{ margin: 0, color: "#333" }}>连接安全性</h3>
                    </div>

                    <div style={{ display: "flex", alignItems: "flex-start" }}>
                        {gradeBox(unifiedConnectScore)} {/* 左边大评级框，增大尺寸 */}

                        <div style={{ flex: 1 }}>
                            <CollapsibleModule label="连接提示" score={unifiedConnectScore}>
                                <ul style={{ margin: 0, paddingLeft: "18px", color: "#333" }}>
                                    {["autodiscover", "autoconfig", "srv"].map(m => {
                                        const connection = results[m]?.score_detail?.connection;
                                        if (!connection) return null;
                                        const warnings = connection.warnings || [];
                                        if (warnings.length > 0) {
                                            return warnings.map((w, idx) => <li key={m + idx}>{m} 机制: {w}</li>);
                                        }
                                        if (connection.Overall_Connection_Score < 50) {
                                            return <li key={m}>{m} 机制连接存在风险</li>;
                                        }  
                                        // 9.11_2
                                        return null;
                                    })}
                                    {unifiedConnectScore === 100 && <li>所有机制连接安全</li>}
                                </ul>
                            </CollapsibleModule>
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
                            <h3 style={{ margin: 0, color: "#333" }}>配置信息概况</h3>
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
                        </div>

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
                                }}
                                onClick={() => toggleCertChain(mech)}
                                >
                                <span style={{ fontSize: "32px", marginRight: "10px" }}>🔗</span>
                                <h3 style={{ margin: 0, color: "#333" }}>
                                    配置服务器证书链 {showCertChainMap[mech] ? "▲" : "▼"}
                                </h3>
                                </div>
                            
                                {showCertChainMap[mech] && (
                                <>
                                    <div style={{ marginBottom: "10px" }}>
                                    {certInfo.RawCerts.map((_, idx) => (
                                        <button
                                        key={idx}
                                        onClick={() => setActiveCertIdx(mech, idx)}
                                        style={{
                                            marginRight: "8px",
                                            padding: "4px 10px",
                                            backgroundColor:
                                            activeCertIdxMap[mech] === idx ? "#5b73a9" : "#ddd",
                                            color: activeCertIdxMap[mech] === idx ? "#fff" : "#000",
                                            border: "none",
                                            borderRadius: "6px",
                                            cursor: "pointer",
                                            fontWeight: "bold",
                                        }}
                                        >
                                        第{idx + 1}证书
                                        </button>
                                    ))}
                                    </div>
                                    <PeculiarCertificateViewer
                                    certificate={certInfo.RawCerts[activeCertIdxMap[mech] || 0]}
                                    />
                                </>
                                )}
                            </div> 
                        )}
                        
                        <div style={{ marginTop: "2rem" }}>
                            <div
                                style={{
                                borderTop: "2px solid #333",
                                paddingTop: "10px",
                                marginBottom: "20px",
                                display: "flex",
                                alignItems: "center",
                                cursor: "pointer",
                                }}
                                onClick={() => toggleAnalysis(mech)}
                            >
                                <span style={{ fontSize: "32px", marginRight: "10px" }}>📊</span>
                                <h3 style={{ margin: 0, color: "#3e5c79" }}>
                                {showAnalysis[mech] ? "收起评分与建议 ▲" : "展开评分与建议 ▼"}
                                </h3>
                            </div>
                            {showAnalysis[mech] && (
                                <>
                                    <div style={{ display: "flex", marginBottom: "1rem" }}>
                                        {["score", "recommend", "radar"].map(tab => (
                                            <button
                                                key={tab}
                                                onClick={() => changeTab(mech, tab)}
                                                style={{
                                                    padding: "8px 16px",
                                                    marginRight: "8px",
                                                    backgroundColor: (activeTab[mech] === tab ? "#2980b9" : "#7f8c8d"),
                                                    color: "#fff",
                                                    border: "none",
                                                    borderRadius: "4px"
                                                }}>
                                                {/* {tab.toUpperCase()} */}
                                                {tabLabelMap[tab]}
                                            </button>
                                        ))}
                                    </div>

                                    {activeTab[mech] === "score" && (
                                        <>
                                            {renderScoreBar("加密端口评分", score.encrypted_ports || 0)}
                                            {renderScoreBar("标准端口评分", score.standard_ports || 0)}
                                            {renderScoreBar(
                                                mech === "srv" ? "DNSSEC评分" : "证书评分",
                                                mech === "srv" ? score.dnssec_score || 0 : score.cert_score || 0
                                            )}
                                            {renderScoreBar("实际连接评分", score.connect_score || 0)}
                                            {renderConnectionDetail(detail)}
                                        </>
                                    )}

                                    {activeTab[mech] === "recommend" && (
                                        <div style={{ backgroundColor: "#7ab0ceff", padding: "15px", borderRadius: "6px" }}>
                                            {(mech === "autodiscover"|| mech === "autoconfig") && portsUsage && (() => {
                                                const rec = getAutodiscoverRecommendations(portsUsage, score);
                                                return (
                                                    <>
                                                        <h4>🔧 端口使用建议</h4>
                                                        <ul>{rec.tips.map((tip, i) => <li key={i}>{tip.text} <b>{tip.impact}</b></li>)}</ul>
                                                        <p><b>预估改进后评分:</b> {rec.improvedScore}</p>
                                                    </>
                                                );
                                            })()}
                                            {mech === "srv" && portsUsage && (() => {
                                                const rec = getSRVRecommendations(portsUsage, score);
                                                return (
                                                    <>
                                                        <h4>🔧 端口使用建议</h4>
                                                        <ul>{rec.tips.map((tip, i) => <li key={i}>{tip.text} <b>{tip.impact}</b></li>)}</ul>
                                                        <p><b>预估改进后评分:</b> {rec.improvedScore}</p>
                                                    </>
                                                );
                                            })()}
                                            {(mech === "autodiscover" || mech === "autoconfig") && certInfo && (() => {
                                                const rec = getCertRecommendations(certInfo, score);
                                                return (
                                                    <>
                                                        <h4>📜 证书配置建议</h4>
                                                        <ul>{rec.tips.map((tip, i) => <li key={i}>{tip.text} <b>{tip.impact}</b></li>)}</ul>
                                                        <p><b>预估改进后评分:</b> {rec.improvedScore}</p>
                                                    </>
                                                );
                                            })()}
                                        </div>
                                    )}

                                    {activeTab[mech] === "radar" && defense && (
                                        <DefenseRadarChart data={defense} />
                                    )}
                                </>
                            )}
                        </div>

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
                                        {/* 9.11 */}
                                        <th style={thStyle}>途径</th>
                                        {/* <th style={thStyle}>序号</th> */}
                                        <th style={thStyle}>请求URI</th>
                                        <th style={thStyle}>是否得到配置</th>
                                        <th style={thStyle}>加密评分</th>
                                        <th style={thStyle}>标准评分</th>
                                        <th style={thStyle}>综合评分</th>
                                        <th style={thStyle}>查看详情</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {result.all.map((item, idx) => (
                                        <tr key={idx}>
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
                                            <td style={tdStyle}>{item.score?.encrypted_ports ?? "-"}</td>
                                            <td style={tdStyle}>{item.score?.standard_ports ?? "-"}</td>
                                            <td style={tdStyle}>{item.score?.overall ?? "-"}</td>
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


                        {/* 8.29 */}
                        {(() => {
                            const collectPortsUsage = (allResults) => {
                                return allResults.map(r => ({
                                    uri: r.uri || "",
                                    ports: Array.isArray(r?.score_detail?.ports_usage) ? r.score_detail.ports_usage : []
                                }));
                            };

                            const allPorts = collectPortsUsage(result.all);
                            if (allPorts.length === 0) return null;

                            const keys = allPorts.map(item => ({
                                uri: item.uri,
                                ports: item.ports
                            }));


                            // 按 protocol 分组比较差异
                            const protocolGroups = {};
                            allPorts.forEach(item => {
                                item.ports.forEach(p => {
                                    if (!protocolGroups[p.protocol]) protocolGroups[p.protocol] = [];
                                    protocolGroups[p.protocol].push({
                                        uri: item.uri,
                                        host: p.host,
                                        port: p.port,
                                        ssl: p.ssl || "未知 SSL"
                                    });
                                });
                            });

                            // 判断哪些 protocol 有差异
                            const diffMap = {}; // { "IMAP": true/false, "POP3": true/false ... }
                            for (const proto in protocolGroups) {
                                const values = protocolGroups[proto].map(v => `${v.host}:${v.port} (${v.ssl})`);
                                if (new Set(values).size > 1) {
                                    diffMap[proto] = true; // 同协议但 host/port/ssl 不一致
                                } else {
                                    diffMap[proto] = false;
                                }
                            }

                            // 如果某条路径有某个 protocol 而其他路径没有 → 也算差异
                            const allProtocols = Object.keys(protocolGroups);
                            allPorts.forEach(item => {
                                allProtocols.forEach(proto => {
                                    const hasProto = item.ports.some(p => p.protocol === proto);
                                    if (!hasProto) diffMap[proto] = true;
                                });
                            });

                            const hasDiff = Object.values(diffMap).some(v => v);
                            result.hasInternalDiff = hasDiff; // ✅ 标记机制内部的差异性9.11
                            if (!hasDiff) return null;

                            return (
                                <div style={{ marginTop: "10px", color: "#e74c3c", fontWeight: "bold" }}>
                                    ⚠️ 检测到该机制下不同路径得到的配置信息中的关键字段不一致：
                                    <div style={{ marginTop: "10px", color: "#333", fontWeight: "normal" }}>
                                        {keys.map((item, idx) => (
                                            <div
                                                key={idx}
                                                style={{
                                                    border: "1px solid #ddd",
                                                    borderRadius: "8px",
                                                    padding: "10px",
                                                    marginBottom: "10px",
                                                    backgroundColor: "#fff"
                                                }}
                                            >
                                                <div style={{ fontSize: "14px", marginBottom: "6px", fontWeight: "bold" }}>
                                                    {item.uri || `路径 ${idx + 1}`}
                                                </div>
                                                <ul style={{ margin: 0, paddingLeft: "18px" }}>
                                                    {item.ports.map((p, i) => {
                                                        const isDiff = diffMap[p.protocol] === true;
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
                                                            </li>
                                                        );
                                                    })}
                                                </ul>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            );
                        })()}
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
                            <h3 style={{ margin: 0, color: "#333" }}>配置信息概况</h3>
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
                                <tr key={`recv-${idx}`}>
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
                                <tr key={`send-${idx}`}>
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

                    <p style={{ color: "#555", marginBottom: "1rem" }}>
                    （以下是基于常见邮件服务前缀和端口的初步探测结果，表示这些服务器端口可以建立 TCP 连接。）
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

                {/* 折叠主观分析 */}
                {mech === "srv" && (
                    <>
                        {/* 分界线标题 */}
                        <div style={{
                            borderTop: "2px solid #333",
                            paddingTop: "10px",
                            marginTop: "30px",
                            marginBottom: "20px",
                            display: "flex",
                            alignItems: "center",
                            cursor: "pointer",
                            userSelect: "none"
                        }}
                            onClick={() => toggleAnalysis(mech)}
                        >
                            <span style={{ fontSize: "28px", marginRight: "10px" }}>📊</span>
                            <h3 style={{ margin: 0, color: "#333" }}>
                                {showAnalysis[mech] ? "收起评分与建议" : "展开评分与建议"}
                            </h3>
                        </div>

                        {showAnalysis[mech] && (
                            <>
                                {/* Tab 切换按钮 */}
                                <div style={{ display: "flex", marginBottom: "1rem" }}>
                                    {["score", "recommend", "radar"].map(tab => (
                                        <button
                                            key={tab}
                                            onClick={() => changeTab(mech, tab)}
                                            style={{
                                                padding: "8px 16px",
                                                marginRight: "8px",
                                                backgroundColor: (activeTab[mech] === tab ? "#2980b9" : "#7f8c8d"),
                                                color: "#fff",
                                                border: "none",
                                                borderRadius: "4px",
                                                cursor: "pointer"
                                            }}>
                                            {tabLabelMap[tab]}
                                        </button>
                                    ))}
                                </div>

                                {/* Score Tab */}
                                {activeTab[mech] === "score" && (
                                    <>
                                        {renderScoreBar("加密端口评分", score.encrypted_ports || 0)}
                                        {renderScoreBar("标准端口评分", score.standard_ports || 0)}
                                        {renderScoreBar("DNSSEC评分", score.dnssec_score || 0)}
                                        {renderScoreBar("实际连接评分", score.connect_score || 0)}
                                        {renderConnectionDetail(detail)}
                                    </>
                                )}

                                {/* Recommend Tab */}
                                {activeTab[mech] === "recommend" && (
                                    <div style={{ backgroundColor: "#7ab0ceff", padding: "15px", borderRadius: "6px" }}>
                                        {portsUsage && (() => {
                                            const rec = getSRVRecommendations(portsUsage, score);
                                            return (
                                                <>
                                                    <h4>🔧 端口使用建议</h4>
                                                    <ul>
                                                        {rec.tips.map((tip, i) =>
                                                            <li key={i}>{tip.text} <b>{tip.impact}</b></li>
                                                        )}
                                                    </ul>
                                                    <p><b>预估改进后评分:</b> {rec.improvedScore}</p>
                                                </>
                                            );
                                        })()}
                                    </div>
                                )}

                                {/* Radar Tab */}
                                {activeTab[mech] === "radar" && defense && (
                                    <DefenseRadarChart data={defense} />
                                )}
                            </>
                        )}
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
            <h1 style={{ fontSize: "2.5rem", marginBottom: "1rem", color: "#29323eff" }}>
            邮件自动化配置检测
            </h1>

            <div style={{ maxWidth: "600px", width: "100%", position: "relative" }}>
                <input
                    type="text"
                    value={email}
                    onChange={handleChange}
                    placeholder="请输入您的邮件地址：如 user@example.com"
                    style={{
                    padding: "1rem",
                    width: "400px",
                    fontSize: "1.2rem",
                    borderRadius: "8px",
                    border: "1px solid #ccc",
                    outline: "none",
                    color: "#000",
                    }}
                />
                <button
                    onClick={handleClick}
                    style={{
                    marginLeft: "1rem",
                    padding: "1rem",
                    fontSize: "1.2rem",
                    borderRadius: "8px",
                    backgroundColor: "#3c71cdff",
                    color: "white",
                    border: "none",
                    cursor: "pointer",
                    fontWeight: "bold",
                    transition: "background 0.3s",
                    }}
                    onMouseOver={(e) => (e.target.style.backgroundColor = "#2e4053")}
                    onMouseOut={(e) => (e.target.style.backgroundColor = "#3a506b")}
                >
                    开始检测
                </button>

                {/* 下拉框 */}
                {suggestions.length > 0 && (
                    <div
                        style={{
                            position: "absolute",
                            top: "100%",
                            left: 0,
                            width: "430px",
                            background: "#fff",
                            border: "1px solid #ccc",
                            maxHeight: "150px",        // 限制高度
                            overflowY: "auto",         // 超出时滚动
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
                                borderBottom: idx !== suggestions.length - 1 ? "1px solid #eee" : "none"
                            }}
                            onMouseOver={(e) => (e.currentTarget.style.background = "#f5f5f5")}
                            onMouseOut={(e) => (e.currentTarget.style.background = "white")}
                        >
                            {s}
                        </div>
                    ))}
                </div>
            )}
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
                    <div style={{ width: "100%", maxWidth: "900px", backgroundColor: "#f5f8fa", padding: "2rem", borderRadius: "12px", boxShadow: "0 2px 10px rgba(0,0,0,0.04)", border: "1px solid #eee", marginTop: "1rem" }}>
    
                        {/* 机制 Tab */}
                        <div style={{ display: "flex", gap: "1rem", marginBottom: "1rem", flexWrap: "wrap" }}>
                            {mechanisms.map((mech) => (
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

            <CSVUploadForm />

            <h2 style={{ marginTop: "3rem", color: "#29394dff" }}>历史查询</h2>
            {recentlySeen.length > 0 ? (
                <ul>
                    {recentlySeen.map((item, index) => (
                        <li key={index} style={{ color: "#444" }}>
                            <strong>{item.domain}</strong> - Score: {item.score}, Grade: {item.grade}, Time:{" "}
                            {new Date(item.timestamp).toLocaleString()}
                        </li>
                    ))}
                </ul>
            ) : (
                <p style={{ color: "#888" }}>暂无记录</p>
            )}
        </div>
    );

}



function CSVUploadForm() {
    const [downloadUrl, setDownloadUrl] = useState(null);
    const [isUploading, setIsUploading] = useState(false);

    const handleUpload = async (e) => {
        const file = e.target.files[0];
        if (!file) return;

        setIsUploading(true);
        setDownloadUrl(null);

        const formData = new FormData();
        formData.append("file", file);

        try {
            const res = await fetch(`/api/uploadCsvAndExportJsonl`, {
            method: "POST",
            body: formData,
            mode: "cors",               // 显式允许跨域9.6
            credentials: "omit",        // 如果不需要带 cookie
            });


            if (!res.ok) {
                throw new Error("Upload failed");
            }

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
            if (!res.ok) {
                throw new Error("下载失败");
            }

            const blob = await res.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.href = url;
            a.download = "result.jsonl"; // 可以改成动态文件名
            document.body.appendChild(a);
            a.click();
            a.remove();
            window.URL.revokeObjectURL(url);
        } catch (err) {
            alert("下载失败：" + err.message);
        }
    };

    return (
        <div style={{ marginBottom: "30px", padding: "20px", textAlign: "center" }}>
            <h3 style={{ fontSize: "1.5rem", marginBottom: "1rem", color: "#29394dff" }}>📄 批量域名检测</h3>
            
            <label 
                style={{ 
                    display: "inline-block",
                    padding: "10px 20px",
                    backgroundColor: "#5daed7ff",
                    color: "white",
                    borderRadius: "8px",
                    cursor: "pointer",
                    fontWeight: "bold",
                    transition: "background 0.3s"
                }}
                onMouseOver={(e) => (e.target.style.backgroundColor = "#3c71cdff")}
                onMouseOut={(e) => (e.target.style.backgroundColor = "#6d92cbff")}
            >
                选择 CSV 文件
                <input 
                    type="file" 
                    accept=".csv" 
                    onChange={handleUpload} 
                    style={{ display: "none" }}
                />
            </label>

            {isUploading && <p style={{ marginTop: "1rem", color: "#888" }}>⏳ 处理中，请稍等...</p>}

            {downloadUrl && (
                <p style={{ marginTop: "1rem" }}>
                    ✅ 查询完成，
                    <button 
                        onClick={handleDownload}
                        style={{
                            marginLeft: "10px",
                            padding: "8px 16px",
                            backgroundColor: "#3a506b",
                            color: "white",
                            border: "none",
                            borderRadius: "6px",
                            cursor: "pointer",
                            fontWeight: "bold",
                            transition: "background 0.3s"
                        }}
                        onMouseOver={(e) => (e.target.style.backgroundColor = "#2e4053")}
                        onMouseOut={(e) => (e.target.style.backgroundColor = "#3a506b")}
                    >
                        点击下载结果
                    </button>
                </p>
            )}
        </div>
    );


}



export default MainPage;
