export const VIEW_ONLY_TOKEN_PATTERN = /^[A-Za-z0-9_-]{16}$/

export function isViewOnlyToken(token) {
  return typeof token === 'string' && VIEW_ONLY_TOKEN_PATTERN.test(token)
}

export function viewOnlyTokenFromHash(hash) {
  if (typeof hash !== 'string') return undefined

  const match = hash.match(/^#\/([A-Za-z0-9_-]{16})$/)
  return match ? match[1] : undefined
}
