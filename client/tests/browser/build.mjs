// Bundles entry.js into OUTDIR (argv[2]) for the hlsbrowser Go test
// (server/internal/mediahls/browser_integration_test.go), with hls.js's worker beside it.
import { build } from 'esbuild'
import { copyFileSync, writeFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))
const out = process.argv[2]
const require = createRequire(import.meta.url)
await build({
  entryPoints: [join(here, 'entry.js')],
  bundle: true,
  format: 'esm',
  outfile: join(out, 'app.js'),
  logLevel: 'warning',
})
copyFileSync(require.resolve('hls.js/dist/hls.worker.js'), join(out, 'hls.worker.js'))
writeFileSync(
  join(out, 'index.html'),
  '<!doctype html><video playsinline muted autoplay width="640" height="360"></video><script type="module" src="/app.js"></script>\n',
)
