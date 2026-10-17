// Bundles entry.js into OUTDIR (argv[2]) for the hlsbrowser Go test.
import { build } from 'esbuild'
import { copyFileSync, writeFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))
const out = process.argv[2]
const require = createRequire(import.meta.url)
await build({
  entryPoints: [join(here, 'entry.js')], bundle: true, format: 'esm', outfile: join(out, 'app.js'), logLevel: 'warning',
  // Vite's ?url import of the hls.js worker; the test server serves it at a fixed path.
  plugins: [{ name: 'url', setup(b) {
    b.onResolve({ filter: /\?url$/ }, (args) => ({ path: args.path, namespace: 'url' }))
    b.onLoad({ filter: /.*/, namespace: 'url' }, () => ({ contents: 'export default "/hls.worker.js"', loader: 'js' }))
  } }],
})
copyFileSync(require.resolve('hls.js/dist/hls.worker.js'), join(out, 'hls.worker.js'))
writeFileSync(join(out, 'index.html'), '<!doctype html><video playsinline muted autoplay width="640" height="360"></video><script type="module" src="/app.js"></script>\n')
