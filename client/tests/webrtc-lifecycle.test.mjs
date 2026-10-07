import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { runInNewContext } from 'node:vm'
import test from 'node:test'
import ts from 'typescript'
import { shouldCreateClientOffer } from '../src/neko/negotiation.js'
import { isViewOnlyToken } from '../src/neko/share.js'

const require = createRequire(import.meta.url)
const compile = source => ts.transpileModule(source, { compilerOptions: {
  target:ts.ScriptTarget.ES2022, module:ts.ModuleKind.CommonJS, esModuleInterop:true,
} }).outputText
const events = {}
runInNewContext(compile(await readFile(new URL('../src/neko/events.ts', import.meta.url), 'utf8')), { exports:events })
const compiled = compile(await readFile(new URL('../src/neko/base.ts', import.meta.url), 'utf8'))

function harness() {
  let now = 0, key = 0
  const timers = new Map(), peers = [], sockets = [], exports = {}, seen = []
  const clock = { setTimeout(callback, delay) { timers.set(++key, { at:now + delay, callback }); return key },
    clearTimeout(id) { timers.delete(id) }, setInterval: () => ++key, clearInterval() {} }
  class Socket {
    static OPEN = 1
    constructor() { this.readyState = 0; sockets.push(this) }
    close() { this.readyState = 3 }
    send() {}
  }
  class Peer {
    constructor(config) { this.config = config; this.connectionState = 'new'; this.iceConnectionState = 'new'; peers.push(this) }
    addTransceiver() {}
    createDataChannel() { return { close() {} } }
    close() { this.connectionState = 'closed' }
    getSenders() { return [] }
    ice(state) { this.iceConnectionState = state; this.oniceconnectionstatechange() }
  }
  runInNewContext(compiled, { exports, window:clock, ...clock, URL,
    WebSocket:Socket, RTCPeerConnection:Peer, console,
    require(path) {
      if (path === 'eventemitter3') return require(path)
      if (path === './events') return events
      if (path === './data') return { OPCODE:{} }
      if (path === './negotiation') return { shouldCreateClientOffer }
      if (path === './share') return { isViewOnlyToken }
      throw new Error(`Unexpected base module: ${path}`)
    },
  })
  class Client extends exports.BaseClient {
    CONNECTING() { seen.push('connecting') }
    CONNECTED() { seen.push('connected') }
    RECONNECTING() { seen.push('reconnecting') }
    DISCONNECTED(reason) { seen.push(reason?.message || 'disconnected') }
    TRACK() {}
    DATA() {}
  }
  const client = new Client()
  client.on('error', () => {})
  client.connect('wss://neko.example/ws', 'private', 'viewer')
  const advance = milliseconds => {
    const end = now + milliseconds
    for (;;) {
      const next = [...timers].filter(([, timer]) => timer.at <= end).sort((a, b) => a[1].at - b[1].at)[0]
      if (!next) break
      now = next[1].at; timers.delete(next[0]); next[1].callback()
    }
    now = end
  }
  return { client, peers, sockets, seen, timers, advance }
}

test('checking stays bounded and duplicate socket/peer setup is refused', async () => {
  const h = harness()
  h.client.connect('wss://neko.example/ws', 'private', 'viewer')
  h.sockets[0].readyState = 1
  await h.client.createPeer(false, [])
  await h.client.createPeer(false, [])
  assert.equal(h.sockets.length, 1)
  assert.equal(h.peers.length, 1)
  h.peers[0].ice('checking')
  assert.equal(h.client.connected, false)
  h.advance(15000)
  assert.ok(h.seen.includes('connection timeout'))
  assert.equal(h.peers[0].connectionState, 'closed')
})

test('transient ICE rechecking keeps the original deadline but a recovered peer survives', async () => {
  const h = harness(); h.sockets[0].readyState = 1
  await h.client.createPeer(false, [])
  h.peers[0].ice('connected'); h.peers[0].ice('disconnected')
  h.advance(4000); h.peers[0].ice('checking'); h.advance(4000)
  assert.equal(h.peers[0].connectionState, 'closed')
  assert.equal(h.seen.filter(value => value.startsWith('peer ICE recovery timeout')).length, 1)
  const recovered = harness(); recovered.sockets[0].readyState = 1
  await recovered.client.createPeer(false, [])
  recovered.peers[0].ice('connected'); recovered.peers[0].ice('disconnected')
  recovered.advance(7000); recovered.peers[0].ice('completed'); recovered.advance(10000)
  assert.equal(recovered.client.connected, true)
  recovered.client.disconnect()
})

test('explicit TURN policy survives and stale timers cannot destroy a new socket', async () => {
  const h = harness(); h.sockets[0].readyState = 1
  const turn = { urls:['turns:relay.example:443'], username:'viewer', credential:'private' }
  await h.client.createPeer(false, [turn])
  assert.equal(h.peers[0].config.iceServers.length, 1)
  assert.equal(h.peers[0].config.iceServers[0], turn)
  const oldTimeout = [...h.timers.values()][0].callback
  h.peers[0].ice('connected'); h.peers[0].ice('disconnected')
  const oldRecovery = [...h.timers.values()][0].callback
  h.client.disconnect(); h.client.connect('wss://neko.example/ws', 'private', 'viewer')
  h.sockets[1].readyState = 1
  await h.client.createPeer(false, [])
  h.peers[1].ice('connected'); h.peers[1].ice('disconnected')
  const activeRecovery = h.client._iceRecoveryTimeout
  oldTimeout(); oldRecovery()
  assert.equal(h.sockets[1].readyState, 1)
  assert.equal(h.client._iceRecoveryTimeout, activeRecovery)
  assert.equal(h.seen.filter(value => value.includes('timeout')).length, 0)
  h.client.disconnect()
})
