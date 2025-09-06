from flask import Flask, request, jsonify
from flask_cors import CORS
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.x509.oid import NameOID
import logging
from sslyze import (
    ServerNetworkLocation,
    ServerNetworkConfiguration,
    ProtocolWithOpportunisticTlsEnum,
    Scanner,
    ServerScanRequest,
    ScanCommand,
    ScanCommandAttemptStatusEnum,
)

app = Flask(__name__)
CORS(app)

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")

def run_tls_scan(host: str, port: int):
    try:
        logging.info(f"=== Starting TLS scan for {host}:{port} ===")

        server_location = ServerNetworkLocation(hostname=host, port=port)
        opportunistic = ProtocolWithOpportunisticTlsEnum.from_default_port(port)
        net_conf = ServerNetworkConfiguration(
            tls_server_name_indication=host,
            tls_opportunistic_encryption=opportunistic
        )

        # 加入常见漏洞检测
        scan_cmds = {
            ScanCommand.TLS_1_0_CIPHER_SUITES,
            ScanCommand.TLS_1_1_CIPHER_SUITES,
            ScanCommand.TLS_1_2_CIPHER_SUITES,
            ScanCommand.TLS_1_3_CIPHER_SUITES,
            ScanCommand.HEARTBLEED,
            ScanCommand.ROBOT,
            ScanCommand.TLS_COMPRESSION,
            ScanCommand.SESSION_RENEGOTIATION,
            ScanCommand.CERTIFICATE_INFO,
        }

        scanner = Scanner()
        scanner.queue_scans([ServerScanRequest(
            server_location=server_location,
            network_configuration=net_conf,
            scan_commands=scan_cmds
        )])

        findings = []
        debug_info = []

        for res in scanner.get_results():
            logging.info(f"Scan status: {res.scan_status}")
            logging.info(f"Connectivity: {res.connectivity_result}")

            attempts = getattr(res, "scan_result", None)
            if not attempts:
                debug_info.append("No scan_result present.")
                continue

            def handle_attempt(attempt, label: str):
                if attempt is None:
                    return
                status = attempt.status
                if status == ScanCommandAttemptStatusEnum.COMPLETED:
                    result = attempt.result
                    finding = {"protocol": label, "status": "COMPLETED"}

                    # ✅ Cipher Suites
                    if hasattr(result, "accepted_cipher_suites"):
                        suites = [{
                            "name": s.cipher_suite.name,
                            "key_size": getattr(getattr(s.cipher_suite, "encryption_algorithm", None), "key_size", "unknown")
                        } for s in result.accepted_cipher_suites]
                        finding["accepted_cipher_suites"] = suites

                    # ✅ Heartbleed
                    elif hasattr(result, "is_vulnerable_to_heartbleed"):
                        finding["is_vulnerable_to_heartbleed"] = result.is_vulnerable_to_heartbleed
                        finding["risk"] = "HIGH" if result.is_vulnerable_to_heartbleed else "LOW"

                    # ✅ ROBOT
                    elif hasattr(result, "robot_result"):
                        finding["robot_result"] = result.robot_result.name
                        finding["risk"] = "HIGH" if "VULNERABLE" in result.robot_result.name else "LOW"

                    # ✅ TLS Compression
                    elif hasattr(result, "supports_compression"):
                        finding["supports_compression"] = result.supports_compression
                        finding["risk"] = "MEDIUM" if result.supports_compression else "LOW"

                    # ✅ Session renegotiation
                    elif hasattr(result, "is_vulnerable_to_client_renegotiation_dos"):
                        finding["client_renegotiations_success_count"] = result.client_renegotiations_success_count
                        finding["is_vulnerable_to_client_renegotiation_dos"] = result.is_vulnerable_to_client_renegotiation_dos
                        finding["supports_secure_renegotiation"] = result.supports_secure_renegotiation
                        finding["risk"] = "MEDIUM" if result.is_vulnerable_to_client_renegotiation_dos else "LOW"

                    # ✅ Certificate info
                    elif hasattr(result, "certificate_deployments"):
                    #elif hasattr(result, "received_certificate_chain") and result.received_certificate_chain:

                        def serialize_cert(cert):
                            try:
                                sha1 = cert.fingerprint(hashes.SHA1()).hex()
                            except Exception:
                                sha1 = None
                            try:
                                sha256 = cert.fingerprint(hashes.SHA256()).hex()
                            except Exception:
                                sha256 = None

                            try:
                                pem = cert.public_bytes(serialization.Encoding.PEM).decode()
                            except Exception:
                                pem = None

                            return {
                                "subject": str(cert.subject),
                                "issuer": str(cert.issuer),
                                "not_before": cert.not_valid_before.isoformat(),
                                "not_after": cert.not_valid_after.isoformat(),
                                "sha1_fingerprint": sha1,
                                "sha256_fingerprint": sha256,
                                "pem": pem  # 新增字段
                            }

                        def serialize_deployment(deploy):
                            return {
                                "received_certificate_chain": [serialize_cert(c) for c in deploy.received_certificate_chain],
                                "leaf_certificate_has_must_staple_extension": deploy.leaf_certificate_has_must_staple_extension,
                                "leaf_certificate_is_ev": deploy.leaf_certificate_is_ev,
                                "leaf_certificate_signed_certificate_timestamps_count": deploy.leaf_certificate_signed_certificate_timestamps_count,
                                "received_chain_contains_anchor_certificate": deploy.received_chain_contains_anchor_certificate,
                                "received_chain_has_valid_order": deploy.received_chain_has_valid_order
                            }

                        finding["certificate_deployments"] = [serialize_deployment(d) for d in result.certificate_deployments]
                        # chain_info = [serialize_cert(cert) for cert in result.received_certificate_chain]
                        # finding["certificate_chain"] = chain_info


                    findings.append(finding)
                    debug_info.append(f"{label}: {finding}")

                elif status == ScanCommandAttemptStatusEnum.ERROR:
                    findings.append({
                        "protocol": label,
                        "status": "ERROR",
                        "error_reason": getattr(attempt, "error_reason", ""),
                        "error_trace": str(getattr(attempt, "error_trace", ""))[:500]
                    })
                else:
                    findings.append({"protocol": label, "status": str(status)})

            # 调用处理函数
            handle_attempt(attempts.tls_1_3_cipher_suites, "TLS_1_3_CIPHER_SUITES")
            handle_attempt(attempts.tls_1_2_cipher_suites, "TLS_1_2_CIPHER_SUITES")
            handle_attempt(attempts.tls_1_1_cipher_suites, "TLS_1_1_CIPHER_SUITES")
            handle_attempt(attempts.tls_1_0_cipher_suites, "TLS_1_0_CIPHER_SUITES")
            handle_attempt(attempts.heartbleed, "HEARTBLEED")
            handle_attempt(attempts.robot, "ROBOT")
            handle_attempt(attempts.tls_compression, "TLS_COMPRESSION")
            handle_attempt(attempts.session_renegotiation, "SESSION_RENEGOTIATION")
            handle_attempt(attempts.certificate_info, "CERTIFICATE_INFO")

        logging.info(f"=== Scan for {host}:{port} completed ===")

        if not findings:
            return {
                "success": False,
                "error": "没有扫描到任何结果（握手失败或命令未执行）",
                "host": host,
                "port": port,
                "debug_info": debug_info
            }

        return {
            "success": True,
            "host": host,
            "port": port,
            "status": "success",
            "findings": findings,
            "debug_info": debug_info
        }

    except Exception as e:
        logging.exception(f"Error during scan for {host}:{port}: {e}")
        return {"success": False, "error": str(e), "debug_info": repr(e)}

@app.route("/deep-analyze", methods=["POST"])
def deep_analyze():
    data = request.get_json()
    host = data.get("host")
    port = int(data.get("port", 443))
    if not host:
        return jsonify({"success": False, "error": "Missing host"}), 400
    return jsonify(run_tls_scan(host, port))

if __name__ == "__main__":
    app.run(host="127.0.0.1", port=5002, debug=True)
