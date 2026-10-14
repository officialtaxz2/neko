export const capabilities = {
  version: 1, backend: 'hls', modes: ['hls', 'll-hls'], container: 'fmp4', video_codec: 'h264-high-3.1',
  audio_codec: 'aac-lc', audio_rate: 48000,
  variants: [['high',1280,720,25], ['medium',854,480,20], ['low',640,360,15]].map(([id,width,height,frame_rate]) =>
    ({ id, source_id: 'high', width, height, frame_rate, video_codec: 'avc1.64001f', container: 'video/mp4', bandwidth: 4000000 })),
  limits: { idle_expires_in_ms: 30000, maximum_playlist_bytes: 65536, max_requests_per_lease: 4, max_blocking_per_lease: 2 },
}
export const offer = { version:1, backend:'hls', mode:'hls', path:'/api/media/hls/session', ticket:'a'.repeat(32), expires_in_ms:10000 }
