// export const checkInternalDiff = (res) => {
//   if (!res?.all) return false;

//   // 收集所有路径的端口配置
//   const allPorts = res.all.map(r => ({
//     uri: r.uri || "",
//     ports: Array.isArray(r?.score_detail?.ports_usage)
//       ? r.score_detail.ports_usage
//       : []
//   }));

//   const protocolGroups = {};
//   allPorts.forEach(item => {
//     item.ports.forEach(p => {
//       if (!protocolGroups[p.protocol]) protocolGroups[p.protocol] = [];
//       protocolGroups[p.protocol].push({
//         uri: item.uri,
//         host: p.host,
//         port: p.port,
//         ssl: p.ssl || "未知 SSL"
//       });
//     });
//   });

//   // 判断同一协议下是否有差异
//   const diffMap = {};
//   for (const proto in protocolGroups) {
//     const values = protocolGroups[proto].map(
//       v => `${v.host}:${v.port} (${v.ssl})`
//     );
//     diffMap[proto] = new Set(values).size > 1;
//   }

//   // 如果某条路径缺失协议 → 算差异
//   const allProtocols = Object.keys(protocolGroups);
//   allPorts.forEach(item => {
//     allProtocols.forEach(proto => {
//       const hasProto = item.ports.some(p => p.protocol === proto);
//       if (!hasProto) diffMap[proto] = true;
//     });
//   });

//   return Object.values(diffMap).some(v => v);
// };
// //9.11


export const checkInternalDiff = (res) => {
  if (!res?.all) return false;

  // 收集所有路径的端口配置
  const allPorts = res.all.map(r => ({
    uri: r.uri || "",
    ports: Array.isArray(r?.score_detail?.ports_usage)
      ? r.score_detail.ports_usage
      : []
  }));

  // 所有协议类型
  const protocolGroups = {};
  allPorts.forEach(item => {
    item.ports.forEach(p => {
      if (!protocolGroups[p.protocol]) protocolGroups[p.protocol] = [];
    });
  });

  // 为每个路径构建“协议配置集合”
  allPorts.forEach((item, idx) => {
    for (const proto in protocolGroups) {
      const portsOfProto = item.ports
        .filter(p => p.protocol === proto)
        .map(p => `${p.host}:${p.port} (${p.ssl})`);
      const unique = [...new Set(portsOfProto)];
      protocolGroups[proto][idx] = unique;
    }
  });

  // 比较每个协议在不同路径间的配置集合是否一致
  const diffMap = {};
  for (const proto in protocolGroups) {
    const sets = protocolGroups[proto].map(arr => (arr || []).sort().join(";"));
    diffMap[proto] = new Set(sets).size > 1;
  }

  // 如果某条路径缺少协议，也算差异
  const allProtocols = Object.keys(protocolGroups);
  allPorts.forEach(item => {
    allProtocols.forEach(proto => {
      const hasProto = item.ports.some(p => p.protocol === proto);
      if (!hasProto) diffMap[proto] = true;
    });
  });

  // 返回是否有任何差异
  return Object.values(diffMap).some(v => v);
};
