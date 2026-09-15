import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

import {
  MEDIA_BACKEND_STORAGE_KEY,
  MEDIA_BACKEND_WEBRTC,
  MEDIA_BACKEND_WEBCODECS,
  mediaBackendNavigationURL,
  normalizeMediaBackendPreference,
  readMediaBackendPreference,
  resolveMediaBackendSelection,
  useCompactMediaStatus,
  writeMediaBackendPreference,
} from '../src/neko/media-selection.js'

function memoryStorage(initial) {
  const values = new Map()
  if (initial !== undefined) values.set(MEDIA_BACKEND_STORAGE_KEY, initial)

  return {
    getItem: (key) => values.get(key) ?? null,
    setItem: (key, value) => values.set(key, value),
    value: () => values.get(MEDIA_BACKEND_STORAGE_KEY),
  }
}

test('media backend persistence defaults absent and invalid values to WebRTC', () => {
  assert.equal(normalizeMediaBackendPreference(undefined), MEDIA_BACKEND_WEBRTC)
  assert.equal(normalizeMediaBackendPreference('hls'), MEDIA_BACKEND_WEBRTC)
  assert.equal(readMediaBackendPreference(memoryStorage()), MEDIA_BACKEND_WEBRTC)
  assert.equal(readMediaBackendPreference(memoryStorage('invalid')), MEDIA_BACKEND_WEBRTC)
  assert.equal(readMediaBackendPreference(memoryStorage(MEDIA_BACKEND_WEBCODECS)), MEDIA_BACKEND_WEBCODECS)

  const deniedStorage = {
    getItem: () => {
      throw new Error('denied')
    },
  }
  assert.equal(readMediaBackendPreference(deniedStorage), MEDIA_BACKEND_WEBRTC)
})

test('media backend choices persist only normalized explicit values', () => {
  const storage = memoryStorage()
  assert.equal(writeMediaBackendPreference(MEDIA_BACKEND_WEBCODECS, storage), true)
  assert.equal(storage.value(), MEDIA_BACKEND_WEBCODECS)
  assert.equal(writeMediaBackendPreference('invalid', storage), true)
  assert.equal(storage.value(), MEDIA_BACKEND_WEBRTC)

  assert.equal(
    writeMediaBackendPreference(MEDIA_BACKEND_WEBCODECS, {
      setItem: () => {
        throw new Error('denied')
      },
    }),
    false,
  )
})

test('exact URL selection overrides storage without mutating it', () => {
  const storage = memoryStorage(MEDIA_BACKEND_WEBRTC)
  assert.deepEqual(resolveMediaBackendSelection('', storage.value()), {
    backend: MEDIA_BACKEND_WEBRTC,
    overridden: false,
    invalidOverride: false,
  })
  assert.deepEqual(resolveMediaBackendSelection('', MEDIA_BACKEND_WEBCODECS), {
    backend: MEDIA_BACKEND_WEBCODECS,
    overridden: false,
    invalidOverride: false,
  })
  assert.deepEqual(resolveMediaBackendSelection('?media=webcodecs-ws', storage.value()), {
    backend: MEDIA_BACKEND_WEBCODECS,
    overridden: true,
    invalidOverride: false,
  })
  assert.equal(storage.value(), MEDIA_BACKEND_WEBRTC)

  for (const search of ['?media=webrtc', '?media=invalid', '?media=webcodecs-ws&media=webcodecs-ws']) {
    assert.deepEqual(resolveMediaBackendSelection(search, MEDIA_BACKEND_WEBCODECS), {
      backend: MEDIA_BACKEND_WEBRTC,
      overridden: true,
      invalidOverride: true,
    })
  }
})

test('selection navigation removes only the stateless diagnostic override', () => {
  const next = new URL(
    mediaBackendNavigationURL(
      'https://neko.example/room?media=webcodecs-ws&scroll=20&media=webcodecs-ws#/Ab3dEf7h_Jk9-mN2',
    ),
  )

  assert.equal(next.searchParams.has('media'), false)
  assert.equal(next.searchParams.get('scroll'), '20')
  assert.equal(next.hash, '#/Ab3dEf7h_Jk9-mN2')
})

test('healthy streaming is compact while negotiation, recovery and failure remain prominent', () => {
  assert.equal(useCompactMediaStatus('streaming'), true)
  for (const status of ['off', 'idle', 'negotiating', 'connecting', 'reconnecting', 'terminal']) {
    assert.equal(useCompactMediaStatus(status), false)
  }
})

test('sidebar changes persist and re-enter startup while the WebRTC action persists WebRTC', async () => {
  const settings = await readFile(new URL('../src/components/settings.vue', import.meta.url), 'utf8')
  const client = await readFile(new URL('../src/neko/index.ts', import.meta.url), 'utf8')
  const video = await readFile(new URL('../src/components/video.vue', import.meta.url), 'utf8')

  assert.match(settings, /<select v-model="media_backend">/)
  assert.match(
    settings,
    /set media_backend\(value: string\) \{\s*this\.\$accessor\.settings\.setMediaBackend\(value\)\s*this\.\$client\.changeMediaBackend\(value\)/,
  )
  assert.match(client, /this\.\$accessor\.settings\.setMediaBackend\(MEDIA_BACKEND_WEBRTC\)/)
  assert.match(client, /public clearMediaBackendOverride\(\) \{\s*if \(!this\.mediaBackendURLOverride\) return/)
  assert.match(client, /window\.location\.assign\(mediaBackendNavigationURL\(window\.location\.href\)\)/)
  assert.match(video, /\{ compact: webCodecsCompactStatus \}/)
  assert.match(video, /v-if="webCodecsTerminal" class="webcodecs-actions"/)
})
