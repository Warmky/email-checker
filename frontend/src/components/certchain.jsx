import { useState } from "react";
import { PeculiarCertificateViewer } from "@peculiar/certificates-viewer-react";

export default function CertificateChain({ chain }) {
    return (
        <div style={{ padding: "16px", backgroundColor: "#7dc7adff", borderRadius: "8px" }}>
            <h3 style={{ fontSize: "1.125rem", fontWeight: "bold", marginBottom: "12px" }}>
                证书链
            </h3>
            {chain.map((cert, index) => (
                <CertificateCard
                    key={index}
                    cert={cert}
                    level={index}
                    isLeaf={index === 0}
                    isRoot={index === chain.length - 1}
                />
            ))}
        </div>
    );
}

function CertificateCard({ cert, level, isLeaf, isRoot }) {
    const [expanded, setExpanded] = useState(isLeaf); // Leaf 默认展开
    const paddingLeft = `${level * 20}px`;

    const getValidityStatus = () => {
        const now = new Date();
        const end = new Date(cert.not_after);
        const daysLeft = Math.floor((end - now) / (1000 * 60 * 60 * 24));

        if (daysLeft < 0) return { type: "expired", label: `已过期`, icon: "❌" };
        if (daysLeft < 30) return { type: "warning", label: `即将过期 (${daysLeft} 天)`, icon: "⚠️" };
        return { type: "ok", label: `有效 (${daysLeft} 天剩余)`, icon: "✅" };
    };

    const status = getValidityStatus();

    const getStatusStyle = () => {
        if (status.type === "expired") return { backgroundColor: "#fee2e2", color: "#b91c1c" };
        if (status.type === "warning") return { backgroundColor: "#ffedd5", color: "#c05621" };
        return { backgroundColor: "#d1fae5", color: "#2f855a" };
    };

    return (
        <div
            style={{
                marginLeft: paddingLeft,
                border: "1px solid #ddd",
                borderRadius: "6px",
                padding: "12px",
                marginBottom: "8px",
                backgroundColor: "#a4dfc2ff",
                boxShadow: "0 1px 3px rgba(0,0,0,0.1)",
            }}
        >
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
                <div>
                    <div style={{ fontWeight: "600" }}>
                        {isLeaf ? "Leaf Certificate" : isRoot ? "Root CA" : "Intermediate CA"}
                    </div>
                    <div>CN: {cert.subject}</div>
                    <div>Issuer: {cert.issuer}</div>
                    <div style={{ marginTop: "4px", fontSize: "0.875rem", color: "#4a5568" }}>
                        有效期: {cert.not_before} → {cert.not_after}
                    </div>
                </div>

                <div style={{ display: "flex", flexDirection: "column", alignItems: "flex-end" }}>
                    <span
                        style={{
                            display: "inline-block",
                            padding: "4px 8px",
                            borderRadius: "4px",
                            fontSize: "0.75rem",
                            fontWeight: "600",
                            marginBottom: "4px",
                            ...getStatusStyle(),
                        }}
                    >
                        {status.icon} {status.label}
                    </span>

                    <button
                        style={{
                            fontSize: "0.875rem",
                            color: "#3182ce",
                            textDecoration: "underline",
                            background: "none",
                            border: "none",
                            cursor: "pointer",
                        }}
                        onClick={() => setExpanded(!expanded)}
                    >
                        {expanded ? "收起" : "展开"}
                    </button>
                </div>
            </div>

            {expanded && (
                <div style={{ marginTop: "8px" }}>
                    {/* 使用 PeculiarCertificateViewer 展示详细证书 */}
                    {cert.pem ? (
                        <PeculiarCertificateViewer certificate={cert.pem} />
                    ) : (
                        <div>
                            SHA1: {cert.sha1_fingerprint}<br/>
                            SHA256: {cert.sha256_fingerprint}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}

