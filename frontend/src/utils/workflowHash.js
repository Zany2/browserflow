// createWorkflowContentHash builds stable hash for Automa workflow content 生成 Automa 工作流内容稳定哈希
export async function createWorkflowContentHash(workflow) {
  const coreWorkflowData = {
    name: normalizeHashText(workflow?.name),
    icon: normalizeHashText(workflow?.icon),
    table: workflow?.table ?? workflow?.dataColumns ?? [],
    drawflow: normalizeHashDrawflow(parseHashJsonValue(workflow?.drawflow)),
    settings: workflow?.settings ?? {},
    globalData: workflow?.globalData ?? '',
    description: normalizeHashText(workflow?.description),
  }
  return sha256Hex(stableStringify(coreWorkflowData))
}

function parseHashJsonValue(value) {
  if (value === undefined) return null
  if (typeof value !== 'string') return value
  const text = value.trim()
  if (!text) return value

  try {
    return JSON.parse(text)
  } catch {
    return value
  }
}

function normalizeHashText(value) {
  return String(value || '').trim()
}

function normalizeHashDrawflow(value) {
  if (!value || typeof value !== 'object') return value

  const drawflow = cloneHashValue(value)
  if (Array.isArray(drawflow.edges)) {
    drawflow.edges = drawflow.edges.map(normalizeHashEdge)
  }

  return drawflow
}

function normalizeHashEdge(edge) {
  if (!edge || typeof edge !== 'object') return edge

  const nextEdge = { ...edge }
  delete nextEdge.sourceNode
  delete nextEdge.targetNode
  return nextEdge
}

function cloneHashValue(value) {
  if (Array.isArray(value)) return value.map(cloneHashValue)
  if (value && typeof value === 'object') {
    return Object.keys(value).reduce((nextValue, key) => {
      nextValue[key] = cloneHashValue(value[key])
      return nextValue
    }, {})
  }

  return value
}

function stableStringify(value) {
  if (value === undefined) return 'null'
  if (Array.isArray(value)) {
    return `[${value.map((item) => stableStringify(item)).join(',')}]`
  }
  if (value && typeof value === 'object') {
    return `{${Object.keys(value)
      .sort()
      .map((key) => `${goJsonStringify(key)}:${stableStringify(value[key])}`)
      .join(',')}}`
  }
  return typeof value === 'string' ? goJsonStringify(value) : JSON.stringify(value)
}

function goJsonStringify(value) {
  return JSON.stringify(value)
    .replace(/</g, '\\u003c')
    .replace(/>/g, '\\u003e')
    .replace(/&/g, '\\u0026')
    .replace(/\u2028/g, '\\u2028')
    .replace(/\u2029/g, '\\u2029')
}

async function sha256Hex(value) {
  if (window.crypto?.subtle) {
    try {
      const bytes = new TextEncoder().encode(value)
      const hashBuffer = await window.crypto.subtle.digest('SHA-256', bytes)
      return Array.from(new Uint8Array(hashBuffer))
        .map((item) => item.toString(16).padStart(2, '0'))
        .join('')
    } catch {
      return sha256HexFallback(value)
    }
  }

  return sha256HexFallback(value)
}

function sha256HexFallback(value) {
  const bytes = new TextEncoder().encode(value)
  const words = bytesToSha256Words(bytes)
  const bitLength = bytes.length * 8
  words[bitLength >> 5] |= 0x80 << (24 - (bitLength % 32))
  words[(((bitLength + 64) >> 9) << 4) + 15] = bitLength

  const hash = [
    0x6a09e667,
    0xbb67ae85,
    0x3c6ef372,
    0xa54ff53a,
    0x510e527f,
    0x9b05688c,
    0x1f83d9ab,
    0x5be0cd19,
  ]
  const constants = [
    0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
    0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
    0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
    0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
    0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
    0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
    0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
    0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
  ]

  for (let i = 0; i < words.length; i += 16) {
    const chunk = words.slice(i, i + 16)
    const state = hash.slice()
    for (let j = 0; j < 64; j += 1) {
      if (j >= 16) {
        const s0 = rotateRight(chunk[j - 15], 7) ^ rotateRight(chunk[j - 15], 18) ^ (chunk[j - 15] >>> 3)
        const s1 = rotateRight(chunk[j - 2], 17) ^ rotateRight(chunk[j - 2], 19) ^ (chunk[j - 2] >>> 10)
        chunk[j] = add32(chunk[j - 16], s0, chunk[j - 7], s1)
      }
      const s1 = rotateRight(state[4], 6) ^ rotateRight(state[4], 11) ^ rotateRight(state[4], 25)
      const ch = (state[4] & state[5]) ^ (~state[4] & state[6])
      const temp1 = add32(state[7], s1, ch, constants[j], chunk[j])
      const s0 = rotateRight(state[0], 2) ^ rotateRight(state[0], 13) ^ rotateRight(state[0], 22)
      const maj = (state[0] & state[1]) ^ (state[0] & state[2]) ^ (state[1] & state[2])
      const temp2 = add32(s0, maj)

      state[7] = state[6]
      state[6] = state[5]
      state[5] = state[4]
      state[4] = add32(state[3], temp1)
      state[3] = state[2]
      state[2] = state[1]
      state[1] = state[0]
      state[0] = add32(temp1, temp2)
    }

    for (let j = 0; j < 8; j += 1) {
      hash[j] = add32(hash[j], state[j])
    }
  }

  return hash.map((item) => (item >>> 0).toString(16).padStart(8, '0')).join('')
}

function bytesToSha256Words(bytes) {
  const words = []
  bytes.forEach((byte, index) => {
    words[index >> 2] = (words[index >> 2] || 0) | (byte << (24 - (index % 4) * 8))
  })
  return words
}

function rotateRight(value, shift) {
  return (value >>> shift) | (value << (32 - shift))
}

function add32(...values) {
  return values.reduce((sum, value) => (sum + (value | 0)) | 0, 0)
}
