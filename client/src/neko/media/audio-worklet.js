const AUDIO_UNDERFLOW_GRACE_SECONDS = 0.1

class NekoMediaAudioProcessor extends AudioWorkletProcessor {
  constructor() {
    super()
    this.queue = []
    this.bufferedFrames = 0
    this.active = false
    this.underflowReported = false
    this.underflowStartedAt = null
    this.port.onmessage = ({ data }) => {
      if (data?.type === 'reset' || data?.type === 'rebuffer') {
        if (data.type === 'rebuffer') {
          for (const chunk of this.queue) {
            this.port.postMessage({
              type: 'discarded',
              id: chunk.id,
              generation: chunk.generation,
              durationMS: chunk.durationMS,
            })
          }
        }
        this.queue = []
        this.bufferedFrames = 0
        this.active = false
        this.underflowReported = false
        this.underflowStartedAt = null
        return
      }
      if (data?.type !== 'chunk' || !Array.isArray(data.planes) || data.planes.length !== 2) return
      const frames = Number(data.numberOfFrames)
      if (data.planes.some((plane) => !(plane instanceof ArrayBuffer))) {
        this.port.postMessage({ type: 'overflow' })
        return
      }
      const planes = data.planes.map((plane) => new Float32Array(plane))
      if (
        !Number.isInteger(frames) ||
        frames <= 0 ||
        !Number.isFinite(data.startTime) ||
        !Number.isFinite(data.durationMS) ||
        planes.some((plane) => plane.length !== frames) ||
        this.bufferedFrames + frames > sampleRate * 0.2
      ) {
        this.port.postMessage({ type: 'overflow' })
        return
      }
      this.queue.push({
        id: data.id,
        generation: data.generation,
        durationMS: data.durationMS,
        startTime: data.startTime,
        planes,
        offset: 0,
      })
      this.bufferedFrames += frames
      this.active = true
      this.underflowReported = false
      this.underflowStartedAt = null
    }
  }

  process(_inputs, outputs) {
    const output = outputs[0]
    if (!output || output.length < 2) return true
    output[0].fill(0)
    output[1].fill(0)

    let outputOffset = 0
    while (outputOffset < output[0].length && this.queue.length > 0) {
      const chunk = this.queue[0]
      const frameTime = currentTime + outputOffset / sampleRate
      if (frameTime < chunk.startTime) {
        const silence = Math.min(output[0].length - outputOffset, Math.ceil((chunk.startTime - frameTime) * sampleRate))
        outputOffset += silence
        continue
      }

      const remaining = chunk.planes[0].length - chunk.offset
      const count = Math.min(remaining, output[0].length - outputOffset)
      output[0].set(chunk.planes[0].subarray(chunk.offset, chunk.offset + count), outputOffset)
      output[1].set(chunk.planes[1].subarray(chunk.offset, chunk.offset + count), outputOffset)
      chunk.offset += count
      outputOffset += count
      this.bufferedFrames = Math.max(0, this.bufferedFrames - count)

      if (chunk.offset >= chunk.planes[0].length) {
        this.queue.shift()
        this.port.postMessage({
          type: 'consumed',
          id: chunk.id,
          generation: chunk.generation,
          durationMS: chunk.durationMS,
        })
      }
    }

    if (this.active && this.queue.length === 0 && outputOffset < output[0].length && !this.underflowReported) {
      if (this.underflowStartedAt === null) {
        this.underflowStartedAt = currentTime
      } else if (currentTime - this.underflowStartedAt >= AUDIO_UNDERFLOW_GRACE_SECONDS) {
        this.underflowReported = true
        this.active = false
        this.underflowStartedAt = null
        this.port.postMessage({ type: 'underflow' })
      }
    } else {
      this.underflowStartedAt = null
    }
    return true
  }
}

registerProcessor('neko-media-audio', NekoMediaAudioProcessor)
