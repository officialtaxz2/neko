/* Paste into the affected browser's console. ES5 syntax for older TV engines.
 * Read-only, ten-minute bounded RAM trace. No URLs, SDP, tokens or chat content.
 * Optional marks: nekoMediaTraceMark('first_picture'|'audio'|'black'|'pause').
 * Stop and print: nekoMediaTraceStop(). Refresh loses unsaved evidence. */
(function () {
  'use strict';
  if (typeof window.nekoMediaTraceStop === 'function') window.nekoMediaTraceStop();
  var start = Date.now(), stopped = false, busy = false, tick, deadline;
  var samples = [], events = [], stats = [], dropped = { samples:0, events:0, stats:0 };
  var video, client, peer, removeVideo = function () {}, removeClient = function () {}, removePeer = function () {};
  var allowed = /^(system\/(init|disconnect)|signal\/(provide|restart|offer|answer|close)|member\/(connected|disconnected)|control\/(give|release|request|requesting|locked)|chat\/(message|emote)|screen\/(resolution|set))$/;
  function number(value) { return typeof value === 'number' && isFinite(value) ? value : null; }
  function elapsed() { return Date.now() - start; }
  function keep(list, value, max, kind) {
    if (list.length >= max) { list.shift(); dropped[kind]++; }
    list.push(value);
  }
  function event(kind) { if (!stopped) keep(events, { t_ms:elapsed(), kind:kind }, 600, 'events'); }
  function media() {
    if (!video) return null;
    var frames = null, lost = null;
    try {
      if (typeof video.getVideoPlaybackQuality === 'function') {
        var quality = video.getVideoPlaybackQuality();
        frames = number(quality.totalVideoFrames); lost = number(quality.droppedVideoFrames);
      } else frames = number(video.webkitDecodedFrameCount);
    } catch (_) {}
    return { time_s:number(video.currentTime), frames:frames, dropped:lost,
      ready:number(video.readyState), paused:video.paused, seeking:video.seeking,
      width:number(video.videoWidth), height:number(video.videoHeight), muted:video.muted,
      error:video.error && [1, 2, 3, 4].indexOf(video.error.code) !== -1 ? video.error.code : null };
  }
  function bind() {
    var node = document.querySelector('.video');
    var vm = node && node.__vue__;
    var nextVideo = node && node.querySelector('video');
    if (nextVideo !== video) {
      removeVideo(); video = nextVideo;
      var bound = video, listeners = [];
      if (bound) ['playing', 'pause', 'waiting', 'stalled', 'canplay', 'loadedmetadata', 'seeking', 'seeked', 'ended', 'error', 'emptied'].forEach(function (kind) {
        var callback = function () { event('video/' + kind); };
        bound.addEventListener(kind, callback);
        listeners.push(function () { bound.removeEventListener(kind, callback); });
      });
      removeVideo = function () { listeners.forEach(function (remove) { remove(); }); };
      event('video/replaced');
    }
    var nextClient = vm && vm.$client;
    if (nextClient !== client) {
      removeClient(); client = nextClient;
      if (client && typeof client.on === 'function' && typeof client.off === 'function') {
        var boundClient = client;
        var debug = function (message) {
          if (typeof message !== 'string') return;
          var match = /^received websocket event ([a-z/_-]+)(?: |$)/.exec(message);
          if (match && allowed.test(match[1])) event('event/' + match[1]);
        };
        boundClient.on('debug', debug);
        removeClient = function () { boundClient.off('debug', debug); };
      } else removeClient = function () {};
    }
    var nextPeer = client && client._peer;
    if (nextPeer !== peer) {
      removePeer(); peer = nextPeer;
      var boundPeer = peer, peerListeners = [];
      if (boundPeer && typeof boundPeer.addEventListener === 'function') {
        ['iceconnectionstatechange', 'connectionstatechange', 'signalingstatechange', 'track'].forEach(function (kind) {
          var callback = function () { event('peer/' + kind); };
          boundPeer.addEventListener(kind, callback);
          peerListeners.push(function () { boundPeer.removeEventListener(kind, callback); });
        });
      }
      removePeer = function () { peerListeners.forEach(function (remove) { remove(); }); };
      event('peer/replaced');
    }
    return vm;
  }
  function peerStats() {
    if (busy || !peer || typeof peer.getStats !== 'function') return;
    var bound = peer, promise;
    try { promise = bound.getStats(); } catch (_) { return; }
    if (!promise || typeof promise.then !== 'function') return;
    busy = true;
    promise.then(function (report) {
      busy = false;
      if (stopped || peer !== bound || typeof report.forEach !== 'function') return;
      var rows = [];
      report.forEach(function (entry) {
        if (rows.length >= 8) return;
        var kind = entry.kind || entry.mediaType;
        if (entry.type === 'inbound-rtp' && (kind === 'audio' || kind === 'video')) {
          rows.push({ kind:kind, bytes:number(entry.bytesReceived), packets:number(entry.packetsReceived),
            lost:number(entry.packetsLost), decoded:number(entry.framesDecoded), dropped:number(entry.framesDropped),
            fps:number(entry.framesPerSecond), jitter_s:number(entry.jitter), freezes:number(entry.freezeCount),
            freeze_s:number(entry.totalFreezesDuration), buffer_s:number(entry.jitterBufferDelay), buffer_count:number(entry.jitterBufferEmittedCount) });
        } else if (entry.type === 'candidate-pair' && entry.state === 'succeeded' && entry.nominated) {
          rows.push({ kind:'selected_pair', rtt_s:number(entry.currentRoundTripTime), incoming_bps:number(entry.availableIncomingBitrate) });
        }
      });
      keep(stats, { t_ms:elapsed(), rows:rows }, 300, 'stats');
    }, function () { busy = false; });
  }
  function sample() {
    var vm = bind();
    var selected = vm && vm.$accessor;
    var backend = selected && selected.hls && selected.hls.selected;
    if (backend !== 'hls' && backend !== 'll-hls') backend = selected && selected.media && selected.media.selected ? 'webcodecs-ws' : 'webrtc';
    keep(samples, { t_ms:elapsed(), backend:backend, visible:document.visibilityState === 'visible',
      socket:client && client._ws ? number(client._ws.readyState) : null,
      ice:peer ? peer.iceConnectionState : null, connection:peer ? peer.connectionState : null,
      signaling:peer ? peer.signalingState : null, media:media() }, 301, 'samples');
    peerStats();
  }
  function finish(reason) {
    if (stopped) return;
    sample(); stopped = true;
    removeVideo(); removeClient(); removePeer();
    window.clearInterval(tick); window.clearTimeout(deadline);
    delete window.nekoMediaTraceStop; delete window.nekoMediaTraceMark;
    console.log(JSON.stringify({ diagnostic:'media-browser-trace-v1', reason:reason, duration_ms:elapsed(),
      user_agent:String(navigator.userAgent || '').slice(0, 240),
      support:{ webRTC:typeof window.RTCPeerConnection !== 'undefined', videoDecoder:typeof window.VideoDecoder !== 'undefined',
        audioDecoder:typeof window.AudioDecoder !== 'undefined', audioWorklet:typeof window.AudioWorkletNode !== 'undefined',
        native_hls:video && typeof video.canPlayType === 'function' ? video.canPlayType('application/vnd.apple.mpegurl') : '' },
      samples:samples, events:events, stats:stats, discarded:dropped,
      limits:'Private browser state only; optional Vue/debug/getStats APIs may be unavailable. Frame/time counters are proxies, not proof of a visible picture or audible sound. Marks are operator observations. No URLs, SDP, addresses, payloads, credentials, network/player changes or causal attribution.' }));
  }
  window.nekoMediaTraceMark = function (kind) {
    if (['first_picture', 'audio', 'black', 'pause', 'recovered'].indexOf(kind) !== -1) event('operator/' + kind);
  };
  window.nekoMediaTraceStop = function () { finish('manual'); };
  sample(); tick = window.setInterval(sample, 2000);
  deadline = window.setTimeout(function () { finish('ten_minute_limit'); }, 600000);
  console.log('Media trace armed for ten minutes; no automatic Retry. Mark observations or stop with nekoMediaTraceStop(). Save JSON before reloading.');
})();
