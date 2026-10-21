import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { runInNewContext } from 'node:vm'
import test from 'node:test'
import ts from 'typescript'

// Stored values are read during store and component setup. A damaged value
// must fall back there, not throw (it blanked the page or broke the player).
const source = await readFile(new URL('../src/utils/localstorage.ts', import.meta.url), 'utf8')
const compiled = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS } }).outputText

function storage(values) {
  const exports = {}
  const items = new Map(Object.entries(values))
  runInNewContext(compiled, { exports, JSON, Number, Array, localStorage: {
    getItem: (key) => items.get(key) ?? null,
    setItem: (key, value) => items.set(key, String(value)),
  } })
  return exports
}

test('numbers fall back when the stored value is not a number', () => {
  const { get } = storage({ volume: 'abc', scroll: '', rate: '55' })
  assert.equal(get('volume', 100), 100)
  assert.equal(get('scroll', 10), 10)
  assert.equal(get('rate', 30), 55)
  assert.equal(get('missing', 7), 7)
})

test('booleans and strings keep their stored values', () => {
  const { get } = storage({ muted: '1', autoplay: '0', layout: 'de' })
  assert.equal(get('muted', false), true)
  assert.equal(get('autoplay', true), false)
  assert.equal(get('layout', 'us'), 'de')
})

test('recent lists survive damaged JSON', () => {
  const stored = storage({ broken: '[broken', object: '{"a":1}', mixed: '["😀",1,null,"🐱"]' })
  const { set } = stored
  const getList = (key) => [...stored.getList(key)] // arrays from the script's own realm
  assert.deepEqual(getList('broken'), [])
  assert.deepEqual(getList('object'), [])
  assert.deepEqual(getList('mixed'), ['😀', '🐱'])
  assert.deepEqual(getList('missing'), [])
  set('fresh', JSON.stringify(['a', 'b']))
  assert.deepEqual(getList('fresh'), ['a', 'b'])
})
