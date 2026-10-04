import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { runInNewContext } from 'node:vm'
import test from 'node:test'
import ts from 'typescript'

// Target-server only. Exercise the real Markdown parser and Vue component;
// no browser, room connection or project code is executed in Codex.
const require = createRequire(import.meta.url)
const source = await readFile(new URL('../src/components/markdown.ts', import.meta.url), 'utf8')
const compiled = ts.transpileModule(source, {
  compilerOptions: {
    target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS,
    experimentalDecorators: true, esModuleInterop: true, useDefineForClassFields: false,
  },
}).outputText
const exports = {}
runInNewContext(compiled, { exports, require })
const Markdown = exports.default

function render(source, openInApp = false) {
  const component = new Markdown({ propsData: { source, openInApp } })
  try {
    const vnode = component.$options.render.call(component, component.$createElement)
    // A component with a generated `template` would reach Vue's compiler.
    assert.equal(vnode.tag, 'div')
    assert.equal(vnode.componentOptions, undefined)
    return vnode.data.domProps.innerHTML
  } finally {
    component.$destroy()
  }
}

test('room messages render as data without compiling Vue expressions', () => {
  const expression = '{{constructor.constructor("return 17")()}}'
  const html = render(expression + ' <img src=x onerror=alert(1)>')
  assert.match(html, /\{\{constructor\.constructor/)
  assert.match(html, /&lt;img/)
  assert.doesNotMatch(html, /<img\b/)
  assert.match(render('`{{7*7}}`'), /<code>\{\{7\*7\}\}<\/code>/)
})

test('emoji and URL attributes cannot inject HTML or executable directives', () => {
  const emoji = render(':x"onmouseover="alert(1):')
  assert.match(emoji, /data-emoji="x&quot;onmouseover=&quot;alert\(1\)"/)
  assert.match(emoji, /title=":x&quot;onmouseover=&quot;alert\(1\):"/)
  assert.doesNotMatch(emoji, /v-tooltip| onmouseover="/)

  const url = render('https://example.com/"onmouseover="alert(1)/end', true)
  assert.match(url, /href="https:\/\/example\.com\/&quot;onmouseover=&quot;alert\(1\)\/end"/)
  assert.match(url, /data-href="https:\/\/example\.com\/&quot;onmouseover=&quot;alert\(1\)\/end"/)
  assert.doesNotMatch(url, / onmouseover="/)
  assert.doesNotMatch(render('[unsafe](javascript:alert(1))'), /href="javascript:/i)
})

test('normal formatting, emoji sprites, spoilers and open-in-app links remain available', () => {
  const html = render('**bold** :smile: ||secret|| https://example.com/', true)
  assert.match(html, /<strong>bold<\/strong>/)
  assert.match(html, /class="emoji" data-emoji="smile" title=":smile:"/)
  assert.match(html, /class="spoiler"/)
  assert.match(html, /open-in-app/)
  assert.match(html, /data-href="https:\/\/example\.com\/"/)
})
