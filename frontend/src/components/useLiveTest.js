import { useEffect, useRef, useState } from "react";

export default function useLiveTest({ host, port, protocol, onFinish }) {
    const [messages, setMessages] = useState([]);
    const wsRef = useRef(null);

    useEffect(() => {
        if (!host || !port || !protocol) return;

        // ✅ 注意这里用反引号和正确的后端路径
        const wsUrl = `ws://localhost:8081/ws/testconnect?host=${host}&port=${port}&protocol=${protocol}`;
        console.log("🔌 Connecting WebSocket:", wsUrl);

        const ws = new WebSocket(wsUrl);
        wsRef.current = ws;

        ws.onopen = () => {
        setMessages((prev) => [...prev, "🔗 WebSocket 已连接"]);
        };

        ws.onmessage = (e) => {
        try {
            const data = JSON.parse(e.data);

            // 如果返回的是结构化结果（带 Host 字段），说明测试完成
            if (data.Host) {
            onFinish && onFinish(data);
            setMessages((prev) => [...prev, "✅ 测试结果已返回"]);
            } else if (data.type === "log") {
            // 后端 SendLog 发过来的消息
            setMessages((prev) => [...prev, data.content]);
            } else {
            // 兜底：当普通字符串
            setMessages((prev) => [...prev, e.data]);
            }
        } catch {
            setMessages((prev) => [...prev, e.data]);
        }
        };

        ws.onerror = (e) => {
        console.error("❌ WebSocket error", e);
        setMessages((prev) => [...prev, "❌ WebSocket error"]);
        };

        ws.onclose = () => {
        setMessages((prev) => [...prev, "🔚 测试已结束"]);
        };

        return () => {
        ws.close();
        };
    }, [host, port, protocol, onFinish]);

    return messages;
}
