// // import React, { useState } from "react";
// // import axios from "axios";
// // import CertificateChain from "./certchain"

// // export default function TlsAnalyzerPanel({ host, port }) {
// //     const [loading, setLoading] = useState(false);
// //     const [analysis, setAnalysis] = useState(null);
// //     const [showDebug, setShowDebug] = useState(false);

// //     const handleDeepAnalyze = async () => {
// //         setLoading(true);
// //         setAnalysis(null);
// //         try {
// //             // const res = await axios.post(
// //             //     "http://127.0.0.1:5002/api/tls/deep-analyze",
// //             //     // "/api/tls/deep-analyze",
// //             //     { host, port },
// //             //     { headers: { "Content-Type": "application/json" } }
// //             // );
// //             const res = await fetch("/tls-api/deep-analyze", {
// //                 method: "POST",
// //                 headers: { "Content-Type": "application/json" },
// //                 body: JSON.stringify({ host, port }),
// //             });
// //             const data = await res.json();
// //             // console.log("=== TLS Analysis Result ===", res.data);  //可用 <-- 打印整个返回对象
// //             console.log("=== TLS Analysis Result ===", data); 
// //             setAnalysis(data);
// //         } catch (error) {
// //             if (error.response) {
// //                 setAnalysis({
// //                     success: false,
// //                     error: error.response.data?.error || "服务端错误",
// //                 });
// //             } else {
// //                 setAnalysis({ success: false, error: error.message || "网络错误" });
// //             }
// //         } finally {
// //             setLoading(false);
// //         }
// //     };

// //     // 自动触发分析9.6
// //     useEffect(() => {
// //         if (host && port) {
// //             handleDeepAnalyze();
// //         }
// //     }, [host, port]);

// //     const renderFindings = () => {
// //         if (!analysis?.findings?.length) return null;

// //         // 汇总 TLS 支持的版本
// //         const supported = analysis.findings
// //             .filter(f => f.status === "COMPLETED" && (f.accepted_cipher_suites?.length || 0) > 0)
// //             .map(f => f.protocol.replace("_CIPHER_SUITES", "").replaceAll("_", " ").replace("TLS ", "TLS "))
// //             .join(", ");

// //         return (
// //             <>
// //                 <p><strong>目标：</strong>{analysis.host}:{analysis.port}</p>
// //                 <p><strong>已检测到支持版本：</strong>{supported || "未检测到"}</p>

// //                 {analysis.findings.map((f, idx) => {
// //                     let content = null;
// //                     const proto = f.protocol.toUpperCase();

// //                     // TLS 密码套件
// //                     if (f.accepted_cipher_suites?.length) {
// //                         content = (
// //                             <ul style={{ margin: 0, paddingLeft: 18 }}>
// //                                 {f.accepted_cipher_suites.map((s, i) => (
// //                                     <li key={i} style={{ lineHeight: 1.6 }}>
// //                                         {s.name}{typeof s.key_size === "number" ? ` (key_size=${s.key_size})` : ""}
// //                                     </li>
// //                                 ))}
// //                             </ul>
// //                         );
// //                     }
// //                     // Heartbleed
// //                     else if (proto === "HEARTBLEED" && f.is_vulnerable_to_heartbleed !== undefined) {
// //                         content = (
// //                             <div style={{ color: f.is_vulnerable_to_heartbleed ? "tomato" : "#355f56ff" }}>
// //                                 {f.is_vulnerable_to_heartbleed ? "❌ 存在 Heartbleed 漏洞 → 服务器可能被攻击者读取敏感信息" 
// //                                                             : "✅ 未检测到 Heartbleed 漏洞 → 服务器安全"}
// //                             </div>
// //                         );
// //                     }

// //                     // ROBOT
// //                     else if (proto === "ROBOT" && f.robot_result) {
// //                         const safe = f.robot_result.includes("NOT_VULNERABLE");
// //                         content = (
// //                             <div style={{ color: safe ? "#355f56ff" : "tomato" }}>
// //                                 {safe ? "✅ 未检测到 ROBOT 漏洞 → RSA 加密未受已知攻击威胁" 
// //                                     : `❌ ROBOT 状态: ${f.robot_result} → RSA 加密存在安全风险`}
// //                             </div>
// //                         );
// //                     }

// //                     // TLS 压缩
// //                     else if (proto === "TLS_COMPRESSION" && f.supports_compression !== undefined) {
// //                         const safe = !f.supports_compression;
// //                         content = (
// //                             <div style={{ color: safe ? "#355f56ff" : "tomato" }}>
// //                                 {safe ? "✅ 不支持 TLS 压缩 → 防止 CRIME 攻击" 
// //                                     : "⚠️ 支持 TLS 压缩 → 可能易受 CRIME 攻击"}
// //                             </div>
// //                         );
// //                     }

// //                     // 会话重协商
// //                     else if (proto === "SESSION_RENEGOTIATION" && f.is_vulnerable_to_client_renegotiation_dos !== undefined) {
// //                         const safe = !f.is_vulnerable_to_client_renegotiation_dos;
// //                         content = (
// //                             <div style={{ color: safe ? "#355f56ff" : "tomato" }}>
// //                                 {safe 
// //                                     ? `✅ 安全的会话重协商支持: ${f.supports_secure_renegotiation ? "是" : "否"} → 会话安全性良好`
// //                                     : "⚠️ 客户端发起的会话重协商存在风险 → 可能被 DoS 攻击"}
// //                             </div>
// //                         );
// //                     }

// //                     else if (proto === "CERTIFICATE_INFO" && f.certificate_deployments?.length) {
// //                         const chain = f.certificate_deployments[0].received_certificate_chain; // 取第一组部署
// //                         content = (
// //                             <CertificateChain chain={chain} />
// //                         );
// //                     }




// //                     return (
// //                         <div key={idx} style={{ marginTop: 0.75, padding: 0.75, background: "#679ba3ff", border: "1px solid #444", borderRadius: 6 }}>
// //                             <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline" }}>
// //                                 <h4 style={{ margin: 0 }}>{f.protocol}</h4>
// //                                 <span style={{ fontSize: 12, color: f.status === "COMPLETED" ? "#0ba450ff" : f.status === "ERROR" ? "tomato" : "#ccc" }}>
// //                                     {f.status}
// //                                 </span>
// //                             </div>
// //                             <div style={{ marginTop: 6 }}>
// //                                 {content || <div style={{ color: "#ccc" }}>无可用信息</div>}
// //                                 {f.error_reason && <div style={{ color: "tomato" }}>{f.error_reason}</div>}
// //                             </div>
// //                         </div>
// //                     );
// //                 })}
// //             </>
// //         );
// //     };


// //     return (
// //         <div style={{ marginTop: "1rem", padding: "1rem", backgroundColor: "#81a7adff", color: "#fff", borderRadius: 8 }}>
// //             <div style={{ display: "flex", gap: 8 }}>
// //                 <button
// //                     onClick={handleDeepAnalyze}
// //                     disabled={loading}
// //                     style={{
// //                         padding: "8px 16px",
// //                         backgroundColor: "#5bc889",
// //                         color: "#fff",
// //                         border: "none",
// //                         borderRadius: "6px",
// //                         cursor: loading ? "not-allowed" : "pointer",
// //                         fontWeight: 600
// //                     }}
// //                 >
// //                     {loading ? "分析中..." : "深度分析 (SSLyze 6.x)"}
// //                 </button>

// //                 {analysis?.debug_info && (
// //                     <button
// //                         onClick={() => setShowDebug(s => !s)}
// //                         style={{
// //                             padding: "8px 12px",
// //                             backgroundColor: "#444",
// //                             color: "#fff",
// //                             border: "none",
// //                             borderRadius: "6px",
// //                             cursor: "pointer"
// //                         }}
// //                     >
// //                         {showDebug ? "隐藏调试" : "查看调试"}
// //                     </button>
// //                 )}
// //             </div>

// //             {analysis && (
// //                 <div style={{ marginTop: "1rem" }}>
// //                     {analysis.success ? renderFindings() : (
// //                         <p style={{ color: "tomato" }}>
// //                             分析失败：{analysis.error || "未知错误"}
// //                         </p>
// //                     )}

// //                     {showDebug && analysis?.debug_info && (
// //                         <div style={{ marginTop: "1rem", padding: "0.75rem", background: "#111", border: "1px dashed #555", borderRadius: 6, maxHeight: 200, overflowY: "auto", fontFamily: "monospace", fontSize: 13 }}>
// //                             {analysis.debug_info.map((line, i) => (<div key={i}>{line}</div>))}
// //                         </div>
// //                     )}
// //                 </div>
// //             )}
// //         </div>
// //     );
// // }


// import React, { useState, useEffect } from "react";
// import CertificateChain from "./certchain";

// export default function TlsAnalyzerPanel({ host, port, cipherSuites = [] }) {
//     const [loading, setLoading] = useState(false);
//     const [analysis, setAnalysis] = useState(null);
//     const [showDebug, setShowDebug] = useState(false);

//     const handleDeepAnalyze = async () => {
//         if (!host || !port) return;
//         setLoading(true);
//         setAnalysis(null);
//         try {
//             const res = await fetch("/tls-api/deep-analyze", {
//                 method: "POST",
//                 headers: { "Content-Type": "application/json" },
//                 body: JSON.stringify({ host, port }),
//             });
//             const data = await res.json();
//             console.log("=== TLS Analysis Result ===", data);
//             setAnalysis(data);
//         } catch (error) {
//             if (error.response) {
//                 setAnalysis({
//                     success: false,
//                     error: error.response.data?.error || "服务端错误",
//                 });
//             } else {
//                 setAnalysis({ success: false, error: error.message || "网络错误" });
//             }
//         } finally {
//             setLoading(false);
//         }
//     };

//     // 自动触发分析
//     useEffect(() => {
//         if (host && port) {
//             handleDeepAnalyze();
//         }
//     }, [host, port]);

//     const renderFindings = () => {
//         if (!analysis?.findings?.length) return null;

//         const supported = analysis.findings
//             .filter(f => f.status === "COMPLETED" && (f.accepted_cipher_suites?.length || 0) > 0)
//             .map(f => f.protocol.replace("_CIPHER_SUITES", "").replaceAll("_", " ").replace("TLS ", "TLS "))
//             .join(", ");

//         return (
//             <>
//                 <p><strong>目标：</strong>{analysis.host}:{analysis.port}</p>
//                 <p><strong>已检测到支持版本：</strong>{supported || "未检测到"}</p>

//                 {analysis.findings.map((f, idx) => {
//                     let content = null;
//                     const proto = f.protocol.toUpperCase();

//                     if (f.accepted_cipher_suites?.length) {
//                         content = (
//                             <ul style={{ margin: 0, paddingLeft: 18 }}>
//                                 {f.accepted_cipher_suites.map((s, i) => (
//                                     <li key={i} style={{ lineHeight: 1.6 }}>
//                                         {/* {s.name}{typeof s.key_size === "number" ? ` (key_size=${s.key_size})` : ""} */}
//                                         {s.name}{s.key_size && s.key_size !== "unknown"? ` (key_size=${s.key_size})`: s.key_size === "unknown"? " (key_size=unknown)": ""}

//                                     </li>
//                                 ))}
//                             </ul>
//                         );
//                     } else if (proto === "HEARTBLEED" && f.is_vulnerable_to_heartbleed !== undefined) {
//                         content = (
//                             <div style={{ color: f.is_vulnerable_to_heartbleed ? "tomato" : "#355f56ff" }}>
//                                 {f.is_vulnerable_to_heartbleed ? "❌ 存在 Heartbleed 漏洞" : "✅ 未检测到 Heartbleed 漏洞"}
//                             </div>
//                         );
//                     } else if (proto === "ROBOT" && f.robot_result) {
//                         const safe = f.robot_result.includes("NOT_VULNERABLE");
//                         content = (
//                             <div style={{ color: safe ? "#355f56ff" : "tomato" }}>
//                                 {safe ? "✅ 未检测到 ROBOT 漏洞" : `❌ ROBOT 状态: ${f.robot_result}`}
//                             </div>
//                         );
//                     } else if (proto === "TLS_COMPRESSION" && f.supports_compression !== undefined) {
//                         const safe = !f.supports_compression;
//                         content = (
//                             <div style={{ color: safe ? "#355f56ff" : "tomato" }}>
//                                 {safe ? "✅ 不支持 TLS 压缩" : "⚠️ 支持 TLS 压缩"}
//                             </div>
//                         );
//                     } else if (proto === "SESSION_RENEGOTIATION" && f.is_vulnerable_to_client_renegotiation_dos !== undefined) {
//                         const safe = !f.is_vulnerable_to_client_renegotiation_dos;
//                         content = (
//                             <div style={{ color: safe ? "#355f56ff" : "tomato" }}>
//                                 {safe 
//                                     ? `✅ 安全的会话重协商支持: ${f.supports_secure_renegotiation ? "是" : "否"}`
//                                     : "⚠️ 客户端发起的会话重协商存在风险"}
//                             </div>
//                         );
//                     } else if (proto === "CERTIFICATE_INFO" && f.certificate_deployments?.length) {
//                         const chain = f.certificate_deployments[0].received_certificate_chain;
//                         content = <CertificateChain chain={chain} />;
//                     }

//                     return (
//                         <div key={idx} style={{ marginTop: 0.75, padding: 0.75, background: "#679ba3ff", border: "1px solid #444", borderRadius: 6 }}>
//                             <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline" }}>
//                                 <h4 style={{ margin: 0 }}>{f.protocol}</h4>
//                                 <span style={{ fontSize: 12, color: f.status === "COMPLETED" ? "#0ba450ff" : f.status === "ERROR" ? "tomato" : "#ccc" }}>
//                                     {f.status}
//                                 </span>
//                             </div>
//                             <div style={{ marginTop: 6 }}>
//                                 {content || <div style={{ color: "#ccc" }}>无可用信息</div>}
//                                 {f.error_reason && <div style={{ color: "tomato" }}>{f.error_reason}</div>}
//                             </div>
//                         </div>
//                     );
//                 })}
//             </>
//         );
//     };

//     return (
//         <div style={{ marginTop: "1rem", padding: "1rem", backgroundColor: "#81a7adff", color: "#fff", borderRadius: 8 }}>
//             <div style={{ display: "flex", gap: 8 }}>
//                 <button
//                     onClick={handleDeepAnalyze}
//                     disabled={loading}
//                     style={{
//                         padding: "8px 16px",
//                         backgroundColor: "#5bc889",
//                         color: "#fff",
//                         border: "none",
//                         borderRadius: "6px",
//                         cursor: loading ? "not-allowed" : "pointer",
//                         fontWeight: 600
//                     }}
//                 >
//                     {loading ? "分析中..." : "深度分析 (SSLyze 6.x)"}
//                 </button>

//                 {analysis?.debug_info && (
//                     <button
//                         onClick={() => setShowDebug(s => !s)}
//                         style={{
//                             padding: "8px 12px",
//                             backgroundColor: "#444",
//                             color: "#fff",
//                             border: "none",
//                             borderRadius: "6px",
//                             cursor: "pointer"
//                         }}
//                     >
//                         {showDebug ? "隐藏调试" : "查看调试"}
//                     </button>
//                 )}
//             </div>

//             {analysis && (
//                 <div style={{ marginTop: "1rem" }}>
//                     {analysis.success ? renderFindings() : (
//                         <p style={{ color: "tomato" }}>
//                             分析失败：{analysis.error || "未知错误"}
//                         </p>
//                     )}

//                     {showDebug && analysis?.debug_info && (
//                         <div style={{ marginTop: "1rem", padding: "0.75rem", background: "#111", border: "1px dashed #555", borderRadius: 6, maxHeight: 200, overflowY: "auto", fontFamily: "monospace", fontSize: 13 }}>
//                             {analysis.debug_info.map((line, i) => (<div key={i}>{line}</div>))}
//                         </div>
//                     )}
//                 </div>
//             )}
//         </div>
//     );
// }
import React, { useState, useEffect } from "react";
import CertificateChain from "./certchain";

export default function TlsAnalyzerPanel({ host, port }) {
    const [loading, setLoading] = useState(false);
    const [analysis, setAnalysis] = useState(null);
    const [showDebug, setShowDebug] = useState(false);

    const handleDeepAnalyze = async () => {
        if (!host || !port) return;
        setLoading(true);
        setAnalysis(null);
        try {
            const res = await fetch("/tls-api/deep-analyze", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ host, port }),
            });
            const data = await res.json();
            console.log("=== TLS Analysis Result ===", data);
            setAnalysis(data);
        } catch (error) {
            setAnalysis({ success: false, error: error.message || "网络错误" });
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (host && port) handleDeepAnalyze();
    }, [host, port]);

    const renderFindings = () => {
        if (!analysis?.findings?.length) return null;

        const supported = analysis.findings
            .filter(f => f.status === "COMPLETED" && (f.accepted_cipher_suites?.length || 0) > 0)
            .map(f => f.protocol.replace("_CIPHER_SUITES", "").replaceAll("_", " "))
            .join(", ");

        return (
            <>
                <p><strong>目标：</strong>{analysis.host}:{analysis.port}</p>
                <p><strong>已检测到支持版本：</strong>{supported || "未检测到"}</p>

                {analysis.findings.map((f, idx) => {
                    let content = null;
                    const proto = f.protocol?.toUpperCase();

                    // TLS 密码套件
                    if (f.accepted_cipher_suites?.length) {
                        content = (
                            <ul style={{ margin: 0, paddingLeft: 18 }}>
                                {f.accepted_cipher_suites.map((s, i) => (
                                    <li key={i} style={{ lineHeight: 1.6 }}>
                                        {s.name}
                                        {s.key_size
                                            ? s.key_size === "unknown"
                                                ? " (key_size=unknown)"
                                                : ` (key_size=${s.key_size})`
                                            : ""}
                                    </li>
                                ))}
                            </ul>
                        );
                    }

                    // Heartbleed
                    if (proto === "HEARTBLEED" && f.is_vulnerable_to_heartbleed !== undefined) {
                        content = (
                            <div style={{ color: f.is_vulnerable_to_heartbleed ? "tomato" : "#355f56ff" }}>
                                {f.is_vulnerable_to_heartbleed
                                    ? "❌ 存在 Heartbleed 漏洞"
                                    : "✅ 未检测到 Heartbleed 漏洞"}
                            </div>
                        );
                    }

                    // ROBOT
                    if (proto === "ROBOT" && f.robot_result) {
                        const safe = f.robot_result.includes("NOT_VULNERABLE");
                        content = (
                            <div style={{ color: safe ? "#355f56ff" : "tomato" }}>
                                {safe ? "✅ 未检测到 ROBOT 漏洞" : `❌ ROBOT 状态: ${f.robot_result}`}
                            </div>
                        );
                    }

                    // TLS 压缩
                    if (proto === "TLS_COMPRESSION" && f.supports_compression !== undefined) {
                        const safe = !f.supports_compression;
                        content = (
                            <div style={{ color: safe ? "#355f56ff" : "tomato" }}>
                                {safe ? "✅ 不支持 TLS 压缩" : "⚠️ 支持 TLS 压缩"}
                            </div>
                        );
                    }

                    // 会话重协商
                    if (proto === "SESSION_RENEGOTIATION" && f.is_vulnerable_to_client_renegotiation_dos !== undefined) {
                        const safe = !f.is_vulnerable_to_client_renegotiation_dos;
                        content = (
                            <div style={{ color: safe ? "#355f56ff" : "tomato" }}>
                                {safe
                                    ? `✅ 安全的会话重协商支持: ${f.supports_secure_renegotiation ? "是" : "否"}`
                                    : "⚠️ 客户端发起的会话重协商存在风险"}
                            </div>
                        );
                    }

                    // 证书信息
                    if (proto === "CERTIFICATE_INFO" && f.certificate_deployments?.length) {
                        const chain = f.certificate_deployments[0].received_certificate_chain;
                        content = <CertificateChain chain={chain} />;
                    }

                    return (
                        <div key={idx} style={{ marginTop: 0.75, padding: 0.75, background: "#bcdcec", border: "1px solid #444", borderRadius: 6 }}>
                            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline" }}>
                                <h4 style={{ margin: 0 }}>{f.protocol}</h4>
                                <span style={{ fontSize: 12, color: f.status === "COMPLETED" ? "#0ba450ff" : f.status === "ERROR" ? "tomato" : "#ccc" }}>
                                    {f.status}
                                </span>
                            </div>
                            <div style={{ marginTop: 6 }}>
                                {content || <div style={{ color: "#ccc" }}>无可用信息</div>}
                                {f.error_reason && <div style={{ color: "tomato" }}>{f.error_reason}</div>}
                            </div>
                        </div>
                    );
                })}
            </>
        );
    };

    return (
        <div style={{ marginTop: "1rem", padding: "1rem", backgroundColor: "#b9d7e8", color: "#fff", borderRadius: 8 }}>
            {/* 🔹 新增提示文字 9.17*/}
            <div style={{ marginBottom: "6px", fontSize: "0.85rem", color: "#fff" }}>
                点击下方按钮，对该邮件服务器进行全面的安全分析
            </div>
            <div style={{ display: "flex", gap: 8 }}>
                <button
                    onClick={handleDeepAnalyze}
                    disabled={loading}
                    style={{
                        padding: "8px 16px",
                        backgroundColor: "#5bc889",
                        color: "#fff",
                        border: "none",
                        borderRadius: "6px",
                        cursor: loading ? "not-allowed" : "pointer",
                        fontWeight: 600
                    }}
                >
                    {loading ? "分析中..." : "深度分析 (SSLyze 6.x)"}
                </button>

                {analysis?.debug_info && (
                    <button
                        onClick={() => setShowDebug(s => !s)}
                        style={{
                            padding: "8px 12px",
                            backgroundColor: "#444",
                            color: "#fff",
                            border: "none",
                            borderRadius: "6px",
                            cursor: "pointer"
                        }}
                    >
                        {showDebug ? "隐藏调试" : "查看调试"}
                    </button>
                )}
            </div>

            {analysis && (
                <div style={{ marginTop: "1rem" }}>
                    {analysis.success ? renderFindings() : (
                        <p style={{ color: "tomato" }}>
                            分析失败：{analysis.error || "未知错误"}
                        </p>
                    )}

                    {showDebug && analysis?.debug_info && (
                        <div style={{ marginTop: "1rem", padding: "0.75rem", background: "#111", border: "1px dashed #555", borderRadius: 6, maxHeight: 200, overflowY: "auto", fontFamily: "monospace", fontSize: 13 }}>
                            {analysis.debug_info.map((line, i) => (<div key={i}>{line}</div>))}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}
