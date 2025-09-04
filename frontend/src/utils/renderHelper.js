//7.22
import { motion } from 'framer-motion';

import {
    Radar,
    RadarChart,
    PolarGrid,
    PolarAngleAxis,
    PolarRadiusAxis,
    ResponsiveContainer,
    Tooltip
} from 'recharts';

const getBarColor = (score) => {
    if (score >= 90) return "#2ecc71";
    if (score >= 75) return "#27ae60";
    if (score >= 60) return "#f1c40f";
    if (score >= 40) return "#e67e22";
    return "#e74c3c";
};

const getGradeColor = (grade) => {
    switch (grade) {
        case "A+": return "#2ecc71";
        case "A": return "#27ae60";
        case "B": return "#f1c40f";
        case "C": return "#e67e22";
        case "D": return "#e74c3c";
        case "F": return "#c0392b";
        default: return "#7f8c8d";
    }
};

const radarLabelMap = {
    sniffing_defense: "监听防御能力",
    tampering_defense: "配置篡改防御能力",
    domain_takeover_defense: "域名接管防御能力",
    fake_cert_defense: "伪造证书防御能力",
    dns_hijack_defense: "DNS 劫持防御能力"
};


export const renderScoreBar = (label, score) => (
        <div style={{ marginBottom: "10px" }}>
            <strong>{label}:</strong>
            <div style={{
                height: "18px",
                backgroundColor: "#eee",
                borderRadius: "10px",
                overflow: "hidden",
                marginTop: "4px"
            }}>
                <div style={{
                    height: "100%",
                    width: `${score}%`,
                    backgroundColor: getBarColor(score),
                    transition: "width 0.5s ease"
                }} />
            </div>
            <span style={{ fontSize: "14px", fontWeight: "bold" }}>{score}%</span>
        </div>
    );

export const renderConnectionDetail = (detail) => (
    <div style={{ marginTop: "20px" }}>
        {/* <h4>Connection Security</h4> */}
        <h4>实际连接安全性检测评分</h4>
        <div style={{ display: "flex", alignItems: "center", marginBottom: "10px" }}>
        <div style={{
                fontSize: "36px",
                fontWeight: "bold",
                color: getGradeColor(detail.Connection_Grade),
                border: "3px solid",
                borderColor: getGradeColor(detail.Connection_Grade),
                borderRadius: "8px",
                padding: "10px 16px",
                marginRight: "20px"
            }}>
                {detail.Connection_Grade}
            </div>
            <div style={{ flexGrow: 1 }}>
                <ScoreBar label="TLS" value={detail.TLS_Connections} color="#4CAF50" />
                <ScoreBar label="STARTTLS" value={detail.STARTTLS_Connections} color="#FFC107" />
                <ScoreBar label="Plaintext" value={detail.Plaintext_Connections} color="#F44336" />
            </div>
        </div>
        {detail.warnings?.length > 0 && (
            <ul style={{ color: "#e74c3c" }}>
                {detail.warnings.map((w, i) => <li key={i}>{w}</li>)}
            </ul>
        )}
    </div>
);

export function ScoreBar({ label, value, color }) {
    return (
        <div style={{ marginBottom: "10px" }}>
            <span style={{ display: "inline-block", width: "100px" }}>{label}</span>
            <div style={{
                display: "inline-block",
                width: "60%",
                height: "20px",
                backgroundColor: "#eee",
                borderRadius: "4px",
                verticalAlign: "middle"
            }}>
                <div style={{
                    width: `${value}%`,
                    height: "100%",
                    backgroundColor: color,
                    borderRadius: "4px"
                }}></div>
            </div>
            <span style={{ marginLeft: "10px" }}>{value}%</span>
        </div>
    );
}

export function DefenseRadarChart({ data }) {
    if (!data || typeof data !== 'object') return null;

    const chartData = Object.entries(data).map(([key, value]) => ({
        dimension: radarLabelMap[key]||key.replace(/_/g, ' ').replace('defense', '').trim(),
        score: typeof value === 'number' ? value : 0
    }));

    return (
        <motion.div
            style={{
                marginTop: "30px",
                height: "360px",
                background: "radial-gradient(circle at center, #87b6c4ff 0%, #479aceff 100%)",
                borderRadius: "16px",
                padding: "20px",
                color: "#fff",
                boxShadow: "0 0 20px rgba(136, 132, 216, 0.3)"
            }}
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.6, ease: "easeOut" }}
        >
            <h4 style={{ marginBottom: "12px", fontSize: "1.2rem", fontWeight: "600", textAlign: "center" }}>
                🛡️ 防御能力雷达图
            </h4>

            <ResponsiveContainer width="100%" height="100%">
                <RadarChart cx="50%" cy="50%" outerRadius="75%" data={chartData}>
                    <PolarGrid
                        stroke="#444"
                        strokeDasharray="4 4"
                    />
                    <PolarAngleAxis
                        dataKey="dimension"
                        stroke="#345c86ff"
                        tick={{ fontSize: 12 }}
                    />
                    <PolarRadiusAxis
                        angle={30}
                        domain={[0, 100]}
                        tick={{ fontSize: 11, fill: "#888" }}
                        axisLine={false}
                        tickLine={false}
                    />
                    <Tooltip
                        contentStyle={{
                            backgroundColor: "#222",
                            border: "none",
                            borderRadius: "8px",
                            color: "#fff",
                            fontSize: "0.9rem"
                        }}
                        formatter={(value) => [`${value}/100`, "Score"]}
                    />
                    <Radar
                        name="Defense Score"
                        dataKey="score"
                        stroke="#8884d8"
                        fill="url(#radarGradient)"
                        fillOpacity={0.7}
                        dot={{ r: 3, fill: "#ffffff" }}
                        isAnimationActive={true}
                    />
                    <defs>
                        <linearGradient id="radarGradient" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="0%" stopColor="#8884d8" stopOpacity={0.8} />
                            <stop offset="100%" stopColor="#8884d8" stopOpacity={0.2} />
                        </linearGradient>
                    </defs>
                </RadarChart>
            </ResponsiveContainer>
        </motion.div>
    );
}

export const getAutodiscoverRecommendations = (portUsage, originalScores) => {
    if (!Array.isArray(portUsage)) return [];

    const recommendations = [];
    let improvedEncryptedPorts = originalScores.encrypted_ports;
    let improvedStandardPorts = originalScores.standard_ports;

    const usedProtocols = new Set();

    for (const item of portUsage) {
        usedProtocols.add(item.protocol);

        if (item.status === "insecure") {
            recommendations.push({
                text: `⚠️ 建议将 ${item.protocol} 从不安全端口 ${item.port} 更换为安全端口（如 ${getSecurePort(item.protocol)}），可提升加密评分`,
                impact: "+40"
            });
            improvedEncryptedPorts = Math.min(improvedEncryptedPorts + 40, 100);
        } else if (item.status === "nonstandard") {
            recommendations.push({
                text: `⚠️ ${item.protocol} 使用了非标准端口 ${item.port}，建议更换为标准端口（如 ${getStandardPort(item.protocol)}）`,
                impact: "+20"
            });
            improvedStandardPorts = Math.min(improvedStandardPorts + 20, 100);
        }
    }

    const simulatedOverall = Math.round(
        (improvedEncryptedPorts + improvedStandardPorts + originalScores.cert_score + originalScores.connect_score) / 4
    );

    if (recommendations.length === 0) {
        recommendations.push({ text: "✅ 配置使用了标准且安全的端口。无需修改。", impact: "" });
    }

    return {
        tips: recommendations,
        improvedScore: simulatedOverall,
        improvedEncryptedPorts,
        improvedStandardPorts,
    };
};

export const getSRVRecommendations = (portUsage, originalScores) => {
    if (!Array.isArray(portUsage)) return [];

    const recommendations = [];
    let improvedEncryptedPorts = originalScores.encrypted_ports;
    let improvedStandardPorts = originalScores.standard_ports;

    const usedProtocols = new Set();

    for (const item of portUsage) {
        usedProtocols.add(item.Protocol);

        if (item.status === "insecure") {
            recommendations.push({
                text: `⚠️ 建议将 ${item.protocol} 从不安全端口 ${item.port} 更换为安全端口（如 ${getSecurePort(item.protocol)}），可提升加密评分`,
                impact: "+40"
            });
            improvedEncryptedPorts = Math.min(improvedEncryptedPorts + 40, 100);
        } else if (item.status === "nonstandard") {
            recommendations.push({
                text: `⚠️ ${item.protocol} 使用了非标准端口 ${item.port}，建议更换为标准端口（如 ${getStandardPort(item.protocol)}）`,
                impact: "+20"
            });
            improvedStandardPorts = Math.min(improvedStandardPorts + 20, 100);
        }
    }

    const simulatedOverall = Math.round(
        (improvedEncryptedPorts + improvedStandardPorts + originalScores.dnssec_score + originalScores.connect_score) / 4
    );

    if (recommendations.length === 0) {
        recommendations.push({ text: "✅ 配置使用了标准且安全的端口。无需修改。", impact: "" });
    }
    console.log("SRV portsUsage sample:", portUsage);
    console.log("Recommendations generated:", recommendations);

    return {
        tips: recommendations,
        improvedScore: simulatedOverall,
        improvedEncryptedPorts,
        improvedStandardPorts,
    };
};

export const getSecurePort = (protocol) => {
    switch (protocol) {
        case "SMTP": return "465";
        case "IMAP": return "993";
        case "POP3": return "995";
        default: return "安全端口";
    }
};

export const getStandardPort = (protocol) => {
    switch (protocol) {
        case "SMTP": return "25 / 587";
        case "IMAP": return "143";
        case "POP3": return "110";
        default: return "标准端口";
    }
};

export const getCertRecommendations = (certInfo, originalScores) => {
    if (!certInfo || typeof certInfo !== 'object') return [];

    const recommendations = [];
    let certScore = originalScores.cert_score;

    if (!certInfo.IsTrusted) {
        recommendations.push({
            text: "⚠️ 当前证书不是由受信任 CA 签发，建议更换为权威机构签发的证书。",
            impact: "+25"
        });
        certScore = Math.min(certScore + 25, 100);
    }

    if (!certInfo.IsHostnameMatch) {
        recommendations.push({
            text: "⚠️ 证书域名与服务器地址不匹配，建议使用与实际主机名一致的证书。",
            impact: "+20"
        });
        certScore = Math.min(certScore + 20, 100);
    }

    if (certInfo.IsExpired) {
        recommendations.push({
            text: "⚠️ 当前证书已过期，建议尽快更换为新的有效证书。",
            impact: "+30"
        });
        certScore = Math.min(certScore + 30, 100);
    }

    if (certInfo.IsSelfSigned) {
        recommendations.push({
            text: "⚠️ 当前为自签名证书，缺乏第三方验证，建议使用由可信 CA 签发的证书。",
            impact: "+25"
        });
        certScore = Math.min(certScore + 25, 100);
    }

    if (certInfo.SignatureAlg && certInfo.SignatureAlg.toLowerCase().includes("sha1")) {
        recommendations.push({
            text: "⚠️ 证书使用了弱签名算法 SHA1，建议使用 SHA256 或更强的算法。",
            impact: "+15"
        });
        certScore = Math.min(certScore + 15, 100);
    }

    if (certInfo.TLSVersion < 0x0303) { // TLS 1.2 以下
        recommendations.push({
            text: "⚠️ 服务器使用了过时的 TLS 协议版本，建议升级到 TLS 1.2 或更高。",
            impact: "+10"
        });
        certScore = Math.min(certScore + 10, 100);
    }

    if (recommendations.length === 0) {
        recommendations.push({
            text: "✅ 当前证书配置良好，未发现需要优化的问题。",
            impact: ""
        });
    }

    // 模拟总分改进后（其他分数保持不变）
    const simulatedOverall = Math.round(
        (originalScores.encrypted_ports + originalScores.standard_ports + certScore + originalScores.connect_score) / 4
    );

    return {
        tips: recommendations,
        improvedScore: simulatedOverall,
        improvedCertScore: certScore
    };
};


