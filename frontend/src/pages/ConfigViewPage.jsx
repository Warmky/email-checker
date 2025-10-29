import { useEffect, useState } from "react";
import { useLocation } from "react-router-dom";
import { PeculiarCertificateViewer } from '@peculiar/certificates-viewer-react';
import TlsAnalyzerPanel from "../components/TlsAnalyzerPanel";


function ConfigViewPage() {
  const location = useLocation();
  const queryParams = new URLSearchParams(location.search);
  const tempId = queryParams.get("id");

  const [uri, setUri] = useState("");
  const [configContent, setConfigContent] = useState("⚠️ 无法加载配置");
  const [connectDetails, setConnectDetails] = useState([]);
  const [rawCerts, setRawCerts] = useState([]);
  const [showCertChain, setShowCertChain] = useState(false);
  const [activeCertIdx, setActiveCertIdx] = useState(0);

  // 👇 新增 state
  const [liveLogs, setLiveLogs] = useState([]);
  const [liveResult, setLiveResult] = useState(null);
  const [testingHost, setTestingHost] = useState(null);

  const [mech, setMech] = useState("");
  const [portsUsage, setPortsUsage] = useState([]); // ✅ 新增 state

  const [showTlsCert, setShowTlsCert] = useState(false);
  const [selectedMode, setSelectedMode] = useState("ssl");
  const [rowModes, setRowModes] = useState({});
  
  const [currentHostForAnalysis, setCurrentHostForAnalysis] = useState(null);
  const [currentPortForAnalysis, setCurrentPortForAnalysis] = useState(null);
  const [showTlsAnalyzer, setShowTlsAnalyzer] = useState(false);


  useEffect(() => {
    if (!tempId) return;

    fetch(`/get-temp-data?id=${tempId}`)
      .then(res => res.json())
      .then(data => {
        console.log("✅ 拉取详情数据:", data);
        setUri(data.uri || "");
        setConfigContent(data.config || "⚠️ 无法获取配置内容");
        setConnectDetails(data.details || []);
        setRawCerts(data.rawCerts || []);
        setMech(data.mech || ""); // ✅ 保存 mech
        setPortsUsage(data.portsUsage || []); // ✅ 保存 portsUsage
      })
      .catch(err => {
        console.error("❌ Failed to fetch temp data:", err);
        setConfigContent("⚠️ 加载数据失败");
      });
  }, [tempId]);

  const tdStyle = { padding: "6px 8px", borderBottom: "1px solid #333" };

  // 👇 复用旧逻辑：点击 Retest 走 WebSocket
  const handleRetest = (item, mode) => {
    const { host, port, type } = item;
  // setTestingHost(item.host);
    setLiveLogs([]);
    setLiveResult(null);
    setTestingHost(`${type}://${host}:${port} [${type}, ${mode}]`); 

    // const ws = new WebSocket(
    //   `ws://localhost:8081/ws/testconnect?host=${item.host}&port=${item.port}&protocol=${item.type}&mode=${mode}`
    // );

    const wsProtocol = window.location.protocol === "https:" ? "wss" : "ws";
    const ws = new WebSocket(
      `${wsProtocol}://${window.location.host}/ws/testconnect?host=${item.host}&port=${item.port}&protocol=${item.type}&mode=${mode}`
    );


    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === "log") {
        setLiveLogs((prev) => [...prev, data.content]);
      } else if (data.type === "result") {
        setLiveResult(data.result);
      }
    };
  };


  const renderConnectionInfo = (info) => {
    if (!info || !info.success) return <span style={{ color: "red" }}>❌</span>;
    return (
      <div style={{ fontSize: "0.9em" }}>
        ✅ <br />
        TLS: {info.info?.version || "?"} <br />
        Cipher: {info.info?.cipher?.join(", ") || "N/A"}
      </div>
    );
  };

  const [retestPorts, setRetestPorts] = useState({}); // Retest 对应端口

  const standardPorts = {
    IMAP: { plain: 143, starttls: 143, ssl: 993 },
    POP3: { plain: 110, starttls: 110, ssl: 995 },
    SMTP: { plain: 25, starttls: 587, ssl: 465 },
  };

  // const renderConnectDetailTable = () => {
  //   if (!Array.isArray(connectDetails) || connectDetails.length === 0) return null;

  //   // 切换模式时更新 Retest 模式和端口
  //   const handleModeChange = (idx, mode) => {
  //     setRowModes(prev => ({ ...prev, [idx]: mode }));
  //     const mapping = standardPorts[connectDetails[idx].type.toUpperCase()] || {};
  //     const recommendedPort = mapping[mode] || connectDetails[idx].port;
  //     setRetestPorts(prev => ({ ...prev, [idx]: recommendedPort }));
  //   };

  //   return (
  //     <div style={{ marginTop: "2rem" }}>
  //       <h3 style={{ color: "#4da6ff", marginBottom: "1rem" }}>🔌 实际连接测试结果</h3>
  //       <table
  //         style={{
  //           width: "100%",
  //           borderCollapse: "collapse",
  //           color: "#eee",
  //           fontFamily: '"Segoe UI", "Roboto", "Arial", sans-serif',
  //           fontSize: "15px",
  //           textAlign: "center",
  //           border: "1px solid #444",
  //           borderRadius: "8px",
  //           overflow: "hidden"
  //         }}
  //       >
  //         <thead>
  //           <tr style={{ backgroundColor: "#6395c6ff", color: "#fff" }}>
  //             <th style={{ padding: "10px" }}>协议</th>
  //             <th style={{ padding: "10px" }}>主机</th>
  //             <th style={{ padding: "10px" }}>端口</th>
  //             <th style={{ padding: "10px" }}>明文</th>
  //             <th style={{ padding: "10px" }}>STARTTLS</th>
  //             <th style={{ padding: "10px" }}>TLS</th>
  //             <th style={{ padding: "10px" }}>重新测试</th>
  //           </tr>
  //         </thead>
  //         <tbody>
  //           {connectDetails.map((item, idx) => {
  //             const defaultMode = Object.entries(standardPorts[item.type.toUpperCase()] || {})
  //                                   .find(([m, p]) => Number(p) === Number(item.port))?.[0] || "plain";
  //             const currentMode = rowModes[idx] ?? defaultMode;
  //             const currentRetestPort = retestPorts[idx] ?? (standardPorts[item.type.toUpperCase()]?.[currentMode] || item.port);

  //             return (
  //               <tr key={idx} style={{ backgroundColor: idx % 2 === 0 ? "#7da6cfff" : "#5a8fc1ff" }}>
  //                 <td style={{ padding: "8px" }}>{item.type}</td>
  //                 <td style={{ padding: "8px" }}>{item.host}</td>
  //                 <td style={{ padding: "8px" }}>{item.port}</td> {/* 原始端口 */}
  //                 <td style={{ padding: "8px" }}>{renderConnectionInfo(item.plain)}</td>
  //                 <td style={{ padding: "8px" }}>{renderConnectionInfo(item.starttls)}</td>
  //                 <td style={{ padding: "8px" }}>{renderConnectionInfo(item.tls)}</td>
  //                 <td style={{ padding: "8px" }}>
  //                   <select
  //                     style={{ 
  //                       marginRight: "8px",
  //                       fontFamily: 'Comic Sans MS, "Arial", "Roboto", "Courier New", sans-serif', // 字体
  //                       fontSize: "14px", // 大小
  //                       color: "#59a3b2ff",     // 文字颜色
  //                     }}
  //                     value={currentMode}
  //                     onChange={(e) => handleModeChange(idx, e.target.value)}
  //                   >
  //                     <option value="plain" style={{ fontFamily: '"Segoe UI", "Roboto", "Arial", sans-serif', fontSize: "14px" }}>plain</option>
  //                     <option value="starttls" style={{ fontFamily: '"Segoe UI", "Roboto", "Arial", sans-serif', fontSize: "14px" }}>starttls</option>
  //                     <option value="ssl" style={{ fontFamily: '"Segoe UI", "Roboto", "Arial", sans-serif', fontSize: "14px" }}>ssl</option>
  //                   </select>

  //                   <span style={{ marginRight: "8px" }}>Port: {currentRetestPort}</span>
  //                   <button
  //                     style={{
  //                       padding: "4px 10px",                 // 内边距
  //                       fontFamily: '"Arial", "Segoe UI", "Roboto", sans-serif', // 字体
  //                       fontSize: "14px",                    // 字号
  //                       color: "#fff",                        // 文字颜色
  //                       backgroundColor: "#70a4d8ff",           // 背景颜色
  //                       border: "none",                       // 去掉边框
  //                       borderRadius: "6px",                  // 圆角
  //                       cursor: "pointer",                    // 鼠标样式
  //                     }}
  //                     onClick={() => handleRetest({ ...item, port: currentRetestPort }, currentMode)}
  //                   >
  //                     Retest
  //                   </button>

  //                 </td>
  //               </tr>
  //             );
  //           })}
  //         </tbody>
  //       </table>

  //       {/* Retest 日志区域 */}
  //       {testingHost && (
  //         <div style={{ marginTop: "2rem", padding: "1rem", backgroundColor: "#dee9efff" }}>
  //           <h4>🔍 正在重新连接测试： {testingHost}</h4>
  //           <div
  //             style={{
  //               maxHeight: "200px",
  //               overflowY: "auto",
  //               background: "#d6dadcff",
  //               padding: "1rem",
  //               fontFamily: "monospace",
  //               fontSize: "14px",
  //             }}
  //           >
  //             {/* {liveLogs.map((line, i) => (<div key={i}>{line}</div>))} */}
  //             {liveLogs
  //               .filter(line => !line.trim().startsWith("📄 {"))
  //               .map((line, i) => (
  //                 <div key={i}>{line}</div>
  //               ))}

  //           </div>

  //           {liveResult && (
  //             <div
  //               style={{
  //                 marginTop: "1rem",
  //                 padding: "1rem",
  //                 backgroundColor: "#cbd7dfff",
  //                 border: "1px solid #666",
  //               }}
  //             >
  //               <h4>🔄 重新测试结果</h4>
  //               {liveResult.success ? (
  //                 <>
  //                   <p>✅ 测试成功</p>
  //                   <p><strong>TLS 版本：</strong>{liveResult.info?.version || "未知"}</p>
  //                   <p><strong>加密套件：</strong>{liveResult.info?.cipher?.join(", ") || "N/A"}</p>
  //                   {liveResult.info?.["tls ca"] && (
  //                     <>
  //                       <div style={{ marginTop: "1rem" }}>
  //                         <h4
  //                           onClick={() => setShowTlsCert(prev => !prev)}
  //                           style={{ cursor: "pointer", color: "#505861ff", userSelect: "none" }}
  //                         >
  //                           🔐 查看服务器证书信息 {showTlsCert ? "▲" : "▼"}
  //                         </h4>
  //                         {showTlsCert && (
  //                           <PeculiarCertificateViewer certificate={liveResult.info["tls ca"]} />
  //                         )}
  //                       </div>
  //                       <a
  //                         href={`data:text/plain;charset=utf-8,${encodeURIComponent(liveResult.info["tls ca"])}`}
  //                         download={`certificate_${testingHost}.crt`}
  //                         style={{
  //                           display: "inline-block",
  //                           marginTop: "1rem",
  //                           backgroundColor: "#6f99ccff",
  //                           color: "#fff",
  //                           padding: "8px 12px",
  //                           textDecoration: "none",
  //                           borderRadius: "4px",
  //                         }}
  //                       >
  //                         ⬇️ 下载服务器证书
  //                       </a>
  //                     </>
  //                   )}
  //                   {/* 深度分析按钮 */}
  //                       <div style={{ marginTop: "1rem" }}>
  //                         <button
  //                           style={{
  //                             backgroundColor: "#586c9bff",
  //                             color: "#fff",
  //                             padding: "6px 12px",
  //                             borderRadius: "4px",
  //                             border: "none",
  //                             cursor: "pointer",
  //                             fontFamily: "Arial, sans-serif",
  //                             fontWeight: "bold",
  //                           }}
  //                           onClick={() => {
  //                             let host, port;
  //                             const match = testingHost.match(/^[a-z]+:\/\/([^:\s]+):(\d+)/i);
  //                             if (match) {
  //                               host = match[1];
  //                               port = parseInt(match[2], 10);
  //                               console.log("host:", host, "port:", port);
  //                             } else {
  //                               console.error("Failed to parse host and port from testingHost:", testingHost);
  //                             }
  //                             setCurrentHostForAnalysis(host);
  //                             setCurrentPortForAnalysis(port);
  //                             setShowTlsAnalyzer(prev => !prev);
  //                           }}
  //                         >
  //                           {showTlsAnalyzer ? "隐藏深度分析" : "深度分析"}
  //                         </button>
  //                       </div>

  //                   {/* 深度分析面板 */}
  //                   {showTlsAnalyzer && currentHostForAnalysis && currentPortForAnalysis && (
  //                     <div style={{ marginTop: "1rem" }}>
  //                       <TlsAnalyzerPanel 
  //                         host={currentHostForAnalysis} 
  //                         port={currentPortForAnalysis} 
  //                         cipherSuites={liveResult.info?.cipher || []} // 可传密码套件
  //                       />
  //                     </div>
  //                   )}
  //                 </>
  //               ) : (
  //                 <>
  //                   <p style={{ color: "red" }}>❌ 测试失败</p>
  //                   <p>
  //                     <strong>错误信息：</strong>
  //                     {liveResult.error || liveResult.info?.error?.join(", ") || "未知错误"}
  //                   </p>
  //                 </>
  //               )}
  //             </div>
  //           )}
  //         </div>
  //       )}
  //     </div>
  //   );
  // };
  const renderConnectDetailTable = () => {
    if (!Array.isArray(connectDetails) || connectDetails.length === 0) return null;
  
    const handleModeChange = (idx, mode) => {
      setRowModes(prev => ({ ...prev, [idx]: mode }));
      const mapping = standardPorts[connectDetails[idx].type.toUpperCase()] || {};
      const recommendedPort = mapping[mode] || connectDetails[idx].port;
      setRetestPorts(prev => ({ ...prev, [idx]: recommendedPort }));
    };
  
    return (
      <div style={{ marginTop: "2rem" }}>
        {/* 标题行 */}
        <div
          style={{
            borderTop: "2px solid #333",
            paddingTop: "10px",
            marginBottom: "20px",
            display: "flex",
            alignItems: "center",
          }}
        >
          <span style={{ fontSize: "32px", marginRight: "10px" }}>🔍</span>
          <h3 style={{ margin: 0, color: "#333" }}>实际连接测试结果</h3>
        </div>
  
        {/* 表格 */}
        <div style={{
          background: "#fff",
          border: "1px solid #ddd",
          borderRadius: "12px",
          boxShadow: "0 2px 6px rgba(0,0,0,0.08)",
          overflowX: "auto"
        }}>
          <table
            style={{
              width: "100%",
              borderCollapse: "collapse",
              fontFamily: '"Segoe UI", "Roboto", "Arial", sans-serif',
              fontSize: "15px",
              textAlign: "center"
            }}
          >
            <thead>
              <tr style={{ backgroundColor: "#a8b9cb", color: "#fff" }}>
                <th style={{ padding: "10px" }}>协议</th>
                <th style={{ padding: "10px" }}>主机</th>
                <th style={{ padding: "10px" }}>端口</th>
                <th style={{ padding: "10px" }}>明文</th>
                <th style={{ padding: "10px" }}>STARTTLS</th>
                <th style={{ padding: "10px" }}>TLS</th>
                <th style={{ padding: "10px" }}>重新测试</th>
              </tr>
            </thead>
            <tbody>
              {connectDetails.map((item, idx) => {
                const defaultMode = Object.entries(standardPorts[item.type.toUpperCase()] || {})
                  .find(([m, p]) => Number(p) === Number(item.port))?.[0] || "plain";
                const currentMode = rowModes[idx] ?? defaultMode;
                const currentRetestPort = retestPorts[idx] ?? (standardPorts[item.type.toUpperCase()]?.[currentMode] || item.port);
  
                return (
                  <tr key={idx} style={{ backgroundColor: idx % 2 === 0 ? "#f9fbfd" : "#f1f6fb" }}>
                    <td style={{ padding: "8px" }}>{item.type}</td>
                    <td style={{ padding: "8px" }}>{item.host}</td>
                    <td style={{ padding: "8px" }}>{item.port}</td>
                    <td style={{ padding: "8px" }}>{renderConnectionInfo(item.plain)}</td>
                    <td style={{ padding: "8px" }}>{renderConnectionInfo(item.starttls)}</td>
                    <td style={{ padding: "8px" }}>{renderConnectionInfo(item.tls)}</td>
                    <td style={{ padding: "8px" }}>
                      <select
                        style={{
                          marginRight: "8px",
                          fontSize: "14px",
                          color: "#1a73e8",
                          border: "1px solid #ccc",
                          borderRadius: "6px",
                          padding: "2px 6px"
                        }}
                        value={currentMode}
                        onChange={(e) => handleModeChange(idx, e.target.value)}
                      >
                        <option value="plain">plain</option>
                        <option value="starttls">starttls</option>
                        <option value="ssl">ssl</option>
                      </select>
  
                      <span style={{ marginRight: "8px" }}>Port: {currentRetestPort}</span>
                      <button
                        style={{
                          padding: "4px 10px",
                          fontSize: "14px",
                          color: "#fff",
                          backgroundColor: "#9acada",
                          border: "none",
                          borderRadius: "6px",
                          cursor: "pointer"
                        }}
                        onClick={() => handleRetest({ ...item, port: currentRetestPort }, currentMode)}
                      >
                        Retest
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
  
        {/* Retest 日志区域 */}
        {testingHost && (
          <div style={{ marginTop: "2rem", padding: "1rem", backgroundColor: "#f8f9fa", borderRadius: "8px", border: "1px solid #ddd" }}>
            <h4>🔄 正在重新连接测试： {testingHost}</h4>
            <div
              style={{
                maxHeight: "200px",
                overflowY: "auto",
                background: "#eef3f8",
                padding: "1rem",
                fontFamily: "monospace",
                fontSize: "14px",
                borderRadius: "6px"
              }}
            >
              {liveLogs.filter(line => !line.trim().startsWith("📄 {")).map((line, i) => (
                <div key={i}>{line}</div>
              ))}
            </div>
  
            {liveResult && (
              <div style={{ marginTop: "1rem", padding: "1rem", backgroundColor: "#f1f6fb", border: "1px solid #ccc", borderRadius: "6px" }}>
                <h4>📊 重新测试结果</h4>
                {liveResult.success ? (
                  <>
                    <p>✅ 测试成功</p>
                    <p><strong>TLS 版本：</strong>{liveResult.info?.version || "未知"}</p>
                    <p><strong>加密套件：</strong>{liveResult.info?.cipher?.join(", ") || "N/A"}</p>
                    {liveResult.info?.["tls ca"] && (
                      <>
                        <h4
                          onClick={() => setShowTlsCert(prev => !prev)}
                          style={{ cursor: "pointer", color: "#222", userSelect: "none" }}
                        >
                          🔐 查看服务器证书 {showTlsCert ? "▲" : "▼"}
                        </h4>
                        {showTlsCert && <PeculiarCertificateViewer certificate={liveResult.info["tls ca"]} />}
                        <a
                          href={`data:text/plain;charset=utf-8,${encodeURIComponent(liveResult.info["tls ca"])}`}
                          download={`certificate_${testingHost}.crt`}
                          style={{
                            display: "inline-block",
                            marginTop: "1rem",
                            backgroundColor: "#b0bad1",
                            color: "#fff",
                            padding: "8px 12px",
                            textDecoration: "none",
                            borderRadius: "6px"
                          }}
                        >
                          ⬇️ 下载服务器证书
                        </a>
                      </>
                    )}

                   {/* 深度分析按钮 */}
                    <div style={{ marginTop: "1rem" }}>
                      <button
                        style={{
                          backgroundColor: "#586c9bff",
                          color: "#fff",
                          padding: "6px 12px",
                          borderRadius: "4px",
                          border: "none",
                          cursor: "pointer",
                          fontFamily: "Arial, sans-serif",
                          fontWeight: "bold",
                        }}
                        onClick={() => {
                          let host, port;
                          const match = testingHost.match(/^[a-z]+:\/\/([^:\s]+):(\d+)/i);
                          if (match) {
                            host = match[1];
                            port = parseInt(match[2], 10);
                            console.log("host:", host, "port:", port);
                            } else {
                            console.error("Failed to parse host and port from testingHost:", testingHost);
                          }
                          setCurrentHostForAnalysis(host);
                          setCurrentPortForAnalysis(port);
                          setShowTlsAnalyzer(prev => !prev);
                        }}
                      >
                        {showTlsAnalyzer ? "隐藏深度分析" : "深度分析（SSLyze 6.x）"}
                      </button>
                    </div>

                      {/* 深度分析面板 */}
                      {showTlsAnalyzer && currentHostForAnalysis && currentPortForAnalysis && (
                        <div style={{ marginTop: "1rem" }}>
                          <TlsAnalyzerPanel 
                            host={currentHostForAnalysis} 
                            port={currentPortForAnalysis} 
                            cipherSuites={liveResult.info?.cipher || []} // 可传密码套件
                          />
                        </div>
                      )}
                    
                  </>
                ) : (
                  <p style={{ color: "red" }}>❌ 测试失败：{liveResult.error || liveResult.info?.error?.join(", ") || "未知错误"}</p>
                )}
              </div>
            )}
          </div>
        )}
      </div>
    );
  };

  // const renderCertChain = () => {
  //   if (!Array.isArray(rawCerts) || rawCerts.length === 0) return null;
  //   return (
  //     <div style={{ marginTop: "2rem" }}>
  //       <h3
  //         onClick={() => setShowCertChain((prev) => !prev)}
  //         style={{ cursor: "pointer", color: "#52a1ebff", userSelect: "none" }}
  //       >
  //         🔐 配置服务器证书链 {showCertChain ? "▲" : "▼"}
  //       </h3>
  //       {showCertChain && (
  //         <>
  //           <div style={{ marginBottom: "10px" }}>
  //             {rawCerts.map((_, idx) => (
  //               <button
  //                 key={idx}
  //                 onClick={() => setActiveCertIdx(idx)}
  //                 style={{
  //                   marginRight: "8px",
  //                   padding: "4px 10px",
  //                   backgroundColor: activeCertIdx === idx ? "#007bff" : "#ddd",
  //                   color: activeCertIdx === idx ? "#fff" : "#974646ff",
  //                   border: "none",
  //                   borderRadius: "4px",
  //                   cursor: "pointer",
  //                 }}
  //               >
  //                 #{idx + 1}
  //               </button>
  //             ))}
  //           </div>
  //           <PeculiarCertificateViewer certificate={rawCerts[activeCertIdx]} />
  //         </>
  //       )}
  //     </div>
  //   );
  // };
  const renderCertChain = () => {
    if (!Array.isArray(rawCerts) || rawCerts.length === 0) return null;
    return (
      <div style={{ marginTop: "2rem" }}>
        <div
          style={{
            borderTop: "2px solid #333",
            paddingTop: "10px",
            marginBottom: "20px",
            display: "flex",
            alignItems: "center",
            cursor: "pointer",
            color: "#333"
          }}
          onClick={() => setShowCertChain(prev => !prev)}
        >
          <span style={{ fontSize: "32px", marginRight: "10px" }}>🔐</span>
          <h3 style={{ margin: 0 }}>配置服务器证书链 {showCertChain ? "▲" : "▼"}</h3>
        </div>
  
        {showCertChain && (
          <div style={{
            background: "#fff",
            border: "1px solid #ddd",
            borderRadius: "12px",
            padding: "1rem",
            boxShadow: "0 2px 6px rgba(0,0,0,0.08)"
          }}>
            <div style={{ marginBottom: "10px" }}>
              {rawCerts.map((_, idx) => (
                <button
                  key={idx}
                  onClick={() => setActiveCertIdx(idx)}
                  style={{
                    marginRight: "8px",
                    padding: "6px 12px",
                    backgroundColor: activeCertIdx === idx ? "#87a9db" : "#e0e0e0",
                    color: activeCertIdx === idx ? "#fff" : "#333",
                    border: "none",
                    borderRadius: "6px",
                    cursor: "pointer"
                  }}
                >
                  证书 #{idx + 1}
                </button>
              ))}
            </div>
            <PeculiarCertificateViewer certificate={rawCerts[activeCertIdx]} />
          </div>
        )}
      </div>
    );
  };

  // return (
  //   <div style={{ backgroundColor: "#ffffff", minHeight: "100vh", padding: "2rem", color: "#222" }}>
  //     {/* 只有 autodiscover 或 autoconfig 显示配置块 */}
  //     {(mech === "autodiscover" || mech === "autoconfig") && (
  //       <>
  //         <h2 style={{ color: "#4da6ff", marginBottom: "1rem" }}>📄 配置文件内容</h2>
  //         <p>
  //           <strong style={{ color: "#b8c4cbff" }}>请求的 URI：</strong> <span style={{ color: "#9ad1ff" }}>{uri}</span>
  //         </p>

  //         <pre
  //           style={{
  //             background: "#f9fbfd",
  //             color: "#2d2d2d",
  //             padding: "20px",
  //             borderRadius: "8px",
  //             whiteSpace: "pre-wrap",
  //             maxHeight: "80vh",
  //             overflowY: "auto",
  //             border: "1px solid #ccc",
  //             fontFamily: `"Fira Code", "Source Code Pro", Menlo, Consolas, monospace`,
  //             fontSize: "0.95rem",
  //           }}
  //         >
  //           {configContent}
  //         </pre>

  //         {configContent && configContent !== "⚠️ 无法获取配置内容" && (
  //           <a
  //             href={`data:text/xml;charset=utf-8,${encodeURIComponent(configContent)}`}
  //             download={`config_from_${encodeURIComponent(uri || "unknown")}.xml`}
  //             style={{
  //               display: "inline-block",
  //               marginTop: "1rem",
  //               backgroundColor: "#1a73e8",
  //               color: "#fff",
  //               padding: "10px 15px",
  //               textDecoration: "none",
  //               borderRadius: "6px",
  //               fontWeight: "bold",
  //               transition: "background 0.3s",
  //             }}
  //             onMouseOver={(e) => (e.target.style.backgroundColor = "#155ab6")}
  //             onMouseOut={(e) => (e.target.style.backgroundColor = "#1a73e8")}
  //           >
  //             ⬇️  下 载 配 置 文 件
  //           </a>
  //         )}
  //       </>
  //     )}

  //     {/* ✅ 配置信息卡片展示 */}
  //     {Array.isArray(portsUsage) && portsUsage.length > 0 && (
  //       <div style={{ marginTop: "2rem" }}>
  //         <h3 style={{ marginBottom: "1rem", color: "#1a73e8" }}>🔌 配置信息概况</h3>
  //         <div style={{ display: "flex", flexWrap: "wrap", gap: "1rem" }}>
  //           {portsUsage.map((item, idx) => (
  //             <div
  //               key={idx}
  //               style={{
  //                 backgroundColor: "#f0f7ff",
  //                 color: "#222",
  //                 padding: "1rem",
  //                 borderRadius: "12px",
  //                 boxShadow: "0 2px 6px rgba(0, 0, 0, 0.08)",
  //                 border: "1px solid #d0e3f0",
  //                 minWidth: "220px",
  //                 flex: "1",
  //                 maxWidth: "280px",
  //               }}
  //             >
  //               <table style={{ width: "100%", borderCollapse: "collapse" }}>
  //                 <tbody>
  //                   <tr>
  //                     <td style={tdStyle}><strong>协议</strong></td>
  //                     <td style={tdStyle}>{item.protocol}</td>
  //                   </tr>
  //                   <tr>
  //                     <td style={tdStyle}><strong>端口</strong></td>
  //                     <td style={tdStyle}>{item.port}</td>
  //                   </tr>
  //                   <tr>
  //                     <td style={tdStyle}><strong>主机名</strong></td>
  //                     <td style={tdStyle}>{item.host}</td>
  //                   </tr>
  //                   <tr>
  //                     <td style={tdStyle}><strong>SSL类型</strong></td>
  //                     <td style={tdStyle}>{item.ssl}</td>
  //                   </tr>
  //                   <tr>
  //                     <td style={tdStyle}><strong>用户名</strong></td>
  //                     <td style={tdStyle}>你的邮件地址</td>
  //                   </tr>
  //                   <tr>
  //                     <td style={tdStyle}><strong>密码</strong></td>
  //                     <td style={tdStyle}>你的邮箱密码</td>
  //                   </tr>
  //                 </tbody>
  //               </table>
  //             </div>
  //           ))}
  //         </div>
  //       </div>
  //     )}

  //     {renderConnectDetailTable()}
  //     {renderCertChain()}
  //   </div>
  // );

  return (
    <div
      style={{
        minHeight: "100vh",
        padding: "3rem 1rem",
        display: "flex",
        justifyContent: "center",
        alignItems: "flex-start",
        backgroundColor: "transparent", // ✅ 保证显示背景图
      }}
    >
      <div
        style={{
          width: "100%",
          maxWidth: "1000px",
          backgroundColor: "rgba(255, 255, 255, 0.55)", // ✅ 半透明磨砂白
          backdropFilter: "blur(12px)",
          borderRadius: "16px",
          padding: "2rem",
          boxShadow: "0 4px 20px rgba(0,0,0,0.1)",
          border: "1px solid rgba(255, 255, 255, 0.4)",
          color: "#222",
        }}
      >
        {/* === 配置文件内容 === */}
        {(mech === "autodiscover" || mech === "autoconfig") && (
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
              <h3 style={{ margin: 0, color: "#333" }}>配置文件内容</h3>
            </div>
  
            <p>
              <strong style={{ color: "#698fd1" }}>请求的 URI：</strong>{" "}
              <span style={{ color: "#698fd1" }}>{uri}</span>
            </p>
  
            <pre
              style={{
                background: "rgba(249, 251, 253, 0.85)",
                color: "#2d2d2d",
                padding: "20px",
                borderRadius: "8px",
                whiteSpace: "pre-wrap",
                maxHeight: "80vh",
                overflowY: "auto",
                border: "1px solid rgba(204, 204, 204, 0.6)",
                fontFamily: `"Fira Code", "Source Code Pro", Menlo, Consolas, monospace`,
                fontSize: "0.95rem",
              }}
            >
              {configContent}
            </pre>
  
            {configContent && configContent !== "⚠️ 无法获取配置内容" && (
              <a
                href={`data:text/xml;charset=utf-8,${encodeURIComponent(configContent)}`}
                download={`config_from_${encodeURIComponent(uri || "unknown")}.xml`}
                style={{
                  display: "inline-block",
                  marginTop: "1rem",
                  backgroundColor: "#1a73e8",
                  color: "#fff",
                  padding: "8px 14px",
                  textDecoration: "none",
                  borderRadius: "6px",
                  fontWeight: 500,
                  fontSize: "0.95rem",
                  transition: "background 0.3s, transform 0.2s",
                  boxShadow: "0 1px 3px rgba(0,0,0,0.2)",
                }}
                onMouseOver={(e) => {
                  e.target.style.backgroundColor = "#155ab6";
                  e.target.style.transform = "translateY(-1px)";
                }}
                onMouseOut={(e) => {
                  e.target.style.backgroundColor = "#1a73e8";
                  e.target.style.transform = "none";
                }}
              >
                ⬇️ 下载配置文件
              </a>
            )}
          </div>
        )}
  
        {/* === 配置信息概况 === */}
        {Array.isArray(portsUsage) && portsUsage.length > 0 && (
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
              <span style={{ fontSize: "32px", marginRight: "10px" }}>🔌</span>
              <h3 style={{ margin: 0, color: "#333" }}>配置信息概况</h3>
            </div>
  
            <div style={{ display: "flex", flexWrap: "wrap", gap: "1rem" }}>
              {portsUsage.map((item, idx) => (
                <div
                  key={idx}
                  style={{
                    backgroundColor: "rgba(255, 255, 255, 0.8)",
                    color: "#222",
                    padding: "1rem",
                    borderRadius: "12px",
                    boxShadow: "0 2px 6px rgba(0, 0, 0, 0.08)",
                    border: "1px solid rgba(220, 220, 220, 0.8)",
                    minWidth: "220px",
                    flex: "1",
                    maxWidth: "280px",
                  }}
                >
                  <table style={{ width: "100%", borderCollapse: "collapse" }}>
                    <tbody>
                      <tr><td style={tdStyle}><strong>协议</strong></td><td style={tdStyle}>{item.protocol}</td></tr>
                      <tr><td style={tdStyle}><strong>端口</strong></td><td style={tdStyle}>{item.port}</td></tr>
                      <tr><td style={tdStyle}><strong>主机名</strong></td><td style={tdStyle}>{item.host}</td></tr>
                      <tr><td style={tdStyle}><strong>SSL类型</strong></td><td style={tdStyle}>{item.ssl}</td></tr>
                      <tr><td style={tdStyle}><strong>用户名</strong></td><td style={tdStyle}>你的邮件地址</td></tr>
                      <tr><td style={tdStyle}><strong>密码</strong></td><td style={tdStyle}>你的邮箱密码</td></tr>
                    </tbody>
                  </table>
                </div>
              ))}
            </div>
          </div>
        )}
  
        {/* === 连接详情 === */}
        <div style={{ marginTop: "2rem" }}>{renderConnectDetailTable()}</div>
  
        {/* === 证书链 === */}
        <div style={{ marginTop: "2rem" }}>{renderCertChain()}</div>
      </div>
    </div>
  );
  
}

export default ConfigViewPage;
