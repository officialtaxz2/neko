import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

import {
  MEDIA_BACKEND_STORAGE_KEY,
  MEDIA_BACKEND_WEBRTC,
  mediaBackendNavigationURL,
  normalizeMediaBackendPreference,
  readMediaBackendPreference,
  resolveMediaBackendSelection,
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
  assert.equal(normalizeMediaBackendPreference('hls-invalid'), MEDIA_BACKEND_WEBRTC)
  assert.equal(readMediaBackendPreference(memoryStorage()), MEDIA_BACKEND_WEBRTC)
  assert.equal(readMediaBackendPreference(memoryStorage('invalid')), MEDIA_BACKEND_WEBRTC)
  // The removed WebCodecs backend may still be stored by older clients.
  assert.equal(normalizeMediaBackendPreference('webcodecs-ws'), MEDIA_BACKEND_WEBRTC)
  assert.equal(readMediaBackendPreference(memoryStorage('webcodecs-ws')), MEDIA_BACKEND_WEBRTC)

  const deniedStorage = {
    getItem: () => {
      throw new Error('denied')
    },
  }
  assert.equal(readMediaBackendPreference(deniedStorage), MEDIA_BACKEND_WEBRTC)
})

test('media backend choices persist only normalized explicit values', () => {
  const storage = memoryStorage()
  assert.equal(writeMediaBackendPreference('hls', storage), true)
  assert.equal(storage.value(), 'hls')
  assert.equal(writeMediaBackendPreference('webcodecs-ws', storage), true)
  assert.equal(storage.value(), MEDIA_BACKEND_WEBRTC)
  assert.equal(writeMediaBackendPreference('invalid', storage), true)
  assert.equal(storage.value(), MEDIA_BACKEND_WEBRTC)

  assert.equal(
    writeMediaBackendPreference('hls', {
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
  assert.deepEqual(resolveMediaBackendSelection('', 'webcodecs-ws'), {
    backend: MEDIA_BACKEND_WEBRTC,
    overridden: false,
    invalidOverride: false,
  })
  assert.deepEqual(resolveMediaBackendSelection('?media=hls', storage.value()), {
    backend: 'hls',
    overridden: true,
    invalidOverride: false,
  })
  assert.equal(storage.value(), MEDIA_BACKEND_WEBRTC)

  for (const search of ['?media=webrtc', '?media=invalid', '?media=webcodecs-ws']) {
    assert.deepEqual(resolveMediaBackendSelection(search, 'hls'), {
      backend: MEDIA_BACKEND_WEBRTC,
      overridden: true,
      invalidOverride: true,
    })
  }
})

test('selection navigation removes only the stateless diagnostic override', () => {
  const next = new URL(
    mediaBackendNavigationURL(
      'https://neko.example/room?media=hls&scroll=20&media=hls#/Ab3dEf7h_Jk9-mN2',
    ),
  )

  assert.equal(next.searchParams.has('media'), false)
  assert.equal(next.searchParams.get('scroll'), '20')
  assert.equal(next.hash, '#/Ab3dEf7h_Jk9-mN2')
})

test('HLS modes remain exact opt-ins and preserve passive link state during selection', () => {
  for (const mode of ['hls', 'll-hls']) {
    const storage = memoryStorage()
    assert.equal(writeMediaBackendPreference(mode, storage), true)
    assert.equal(readMediaBackendPreference(storage), mode)
    assert.deepEqual(resolveMediaBackendSelection(`?media=${mode}`, 'webrtc'), {
      backend: mode, overridden: true, invalidOverride: false,
    })
    const next = new URL(mediaBackendNavigationURL(`https://neko.example/prefix/?media=${mode}&embed=1#/Ab3dEf7h_Jk9-mN2`))
    assert.equal(next.search, '?embed=1')
    assert.equal(next.hash, '#/Ab3dEf7h_Jk9-mN2')
  }
  for (const search of ['?media=HLS', '?media=hls&media=ll-hls', '?media=ll-hls-extra']) {
    assert.equal(resolveMediaBackendSelection(search, 'hls').backend, 'webrtc')
  }
})

test('sidebar changes persist and re-enter startup while the WebRTC action persists WebRTC', async () => {
  const settings = await readFile(new URL('../src/components/settings.vue', import.meta.url), 'utf8')
  const client = await readFile(new URL('../src/neko/index.ts', import.meta.url), 'utf8')
  const video = await readFile(new URL('../src/components/video.vue', import.meta.url), 'utf8')

  assert.match(settings, /<select v-model="media_backend">/)
  assert.match(
    settings,
    /set media_backend\(value: string\) \{\s*if \(!this\.\$client\.canSelectMediaBackend\(value\)\) return\s*this\.\$accessor\.settings\.setMediaBackend\(value\)\s*this\.\$client\.changeMediaBackend\(value\)/,
  )
  assert.match(client, /this\.\$accessor\.settings\.setMediaBackend\(MEDIA_BACKEND_WEBRTC\)/)
  assert.match(client, /public clearMediaBackendOverride\(\) \{\s*if \(!this\.mediaBackendURLOverride\) return/)
  assert.match(client, /window\.location\.assign\(mediaBackendNavigationURL\(window\.location\.href\)\)/)
  assert.match(video, /v-if="hlsStatus === 'terminal'" class="media-actions"/)
  assert.match(video, /@click\.stop\.prevent="useWebRTC"/)
  assert.doesNotMatch(settings, /webcodecs/i)
})
