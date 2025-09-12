export const checkInternalDiff = (res) => {
  if (!res?.all) return false;

  // 收集所有路径的端口配置
  const allPorts = res.all.map(r => ({
    uri: r.uri || "",
    ports: Array.isArray(r?.score_detail?.ports_usage)
      ? r.score_detail.ports_usage
      : []
  }));

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

  // 判断同一协议下是否有差异
  const diffMap = {};
  for (const proto in protocolGroups) {
    const values = protocolGroups[proto].map(
      v => `${v.host}:${v.port} (${v.ssl})`
    );
    diffMap[proto] = new Set(values).size > 1;
  }

  // 如果某条路径缺失协议 → 算差异
  const allProtocols = Object.keys(protocolGroups);
  allPorts.forEach(item => {
    allProtocols.forEach(proto => {
      const hasProto = item.ports.some(p => p.protocol === proto);
      if (!hasProto) diffMap[proto] = true;
    });
  });

  return Object.values(diffMap).some(v => v);
};
//9.11