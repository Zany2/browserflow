export function getClientIp(row) {
  return row?.client_ip || row?.ip || row?.remote_ip || row?.last_ip || row?.source_ip || ''
}

export function getClientNodeId(row) {
  return row?.node_id || row?.nodeId || ''
}

export function getClientNodeName(row) {
  return row?.node_name || row?.nodeName || ''
}

export function buildClientNodeIdentity(clientIp, nodeId) {
  clientIp = String(clientIp || '').trim()
  nodeId = String(nodeId || '').trim()
  if (!clientIp) return nodeId
  if (!nodeId || nodeId === clientIp) return clientIp
  return `${clientIp}|${nodeId}`
}

export function buildClientNodeOptions(clients, options = {}) {
  const seen = new Set()
  const ipOptions = []
  const nodeOptions = []

  for (const client of clients || []) {
    const clientIp = getClientIp(client)
    const nodeId = getClientNodeId(client)
    const nodeName = getClientNodeName(client)
    const key = buildClientNodeIdentity(clientIp, nodeId)
    if (!key || seen.has(key)) continue

    seen.add(key)
    if (clientIp && !ipOptions.includes(clientIp)) ipOptions.push(clientIp)

    nodeOptions.push({
      key,
      client_ip: clientIp,
      source_ip: clientIp,
      node_id: nodeId,
      source_node_id: nodeId,
      node_name: nodeName,
      label: options.label
        ? options.label({ client, clientIp, nodeId, nodeName })
        : [clientIp, nodeName || nodeId].filter(Boolean).join(' / '),
      raw: client,
    })
  }

  return { ipOptions, nodeOptions }
}
