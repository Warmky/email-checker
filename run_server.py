import os
import subprocess
import time
import signal

# 端口清理函数
def kill_port(port):
    try:
        subprocess.run(f"fuser -k {port}/tcp", shell=True, check=False)
    except Exception:
        pass

# 清理占用的端口
print("🔧 清理旧进程...")
for p in [5000, 5002, 8081, 3000]:
    kill_port(p)

# 1. 激活 conda 并启动 tls_analyzer.py
print("📡 启动 Python TLS Analyzer...")
tls_cmd = "source /opt/anaconda3/etc/profile.d/conda.sh && conda activate apkpure && python backend/tls_analyzer.py"
tls_proc = subprocess.Popen(["bash", "-c", tls_cmd], stdout=open("logs/tls.log", "w"), stderr=subprocess.STDOUT)

# 2. 启动 Go 后端
print("启动 Go 后端...")
go_proc = subprocess.Popen(
    ["go", "run", "main.go"],
    cwd="backend/cmd/server",
    stdout=open("logs/go.log", "w"),
    stderr=subprocess.STDOUT
)

# 3. 启动 React 前端
print("🌐 启动 React 前端...")
react_proc = subprocess.Popen(
    ["npm", "start"],
    cwd="frontend",
    stdout=open("logs/frontend.log", "w"),
    stderr=subprocess.STDOUT
)


print("✅ 所有服务已启动！")
print("👉 打开浏览器访问: http://localhost:3000")

try:
    # 持续保持脚本运行，否则子进程会退出
    while True:
        time.sleep(1)
except KeyboardInterrupt:
    print("\n⏹ 检测到 Ctrl+C，正在关闭所有进程...")

    for proc in [tls_proc, go_proc, react_proc]:
        try:
            os.killpg(os.getpgid(proc.pid), signal.SIGTERM)
        except Exception:
            pass

    print("✅ 所有进程已关闭")
