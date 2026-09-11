import assert from 'node:assert/strict'
import test from 'node:test'

import { isViewOnlyToken, viewOnlyTokenFromHash } from '../src/neko/share.js'

const token = 'Ab3dEf7h_Jk9-mN2'

test('view-only token uses exactly 16 URL-safe characters', () => {
  assert.equal(isViewOnlyToken(token), true)
  assert.equal(isViewOnlyToken('Ab3dEf7h_Jk9-mN'), false)
  assert.equal(isViewOnlyToken('Ab3dEf7h+Jk9/mN2'), false)
  assert.equal(isViewOnlyToken('0123456789abcdef'.repeat(4)), false)
})

test('view-only link uses the compact root fragment route', () => {
  assert.equal(viewOnlyTokenFromHash(`#/${token}`), token)
  assert.equal(viewOnlyTokenFromHash(`#/watch/${token}`), undefined)
  assert.equal(viewOnlyTokenFromHash(`#/${token}/extra`), undefined)
  assert.equal(viewOnlyTokenFromHash(''), undefined)
})
