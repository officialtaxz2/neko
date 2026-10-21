export function set<T extends string | number | boolean>(key: string, val: T) {
  switch (typeof val) {
    case 'number':
      localStorage.setItem(key, val.toString())
      break
    case 'string':
      localStorage.setItem(key, val)
      break
    case 'boolean':
      localStorage.setItem(key, val ? '1' : '0')
      break
  }
}

export function get<T extends string | number | boolean>(key: string, def: T): T {
  const store = localStorage.getItem(key)
  if (store) {
    switch (typeof def) {
      case 'number': {
        // A damaged value falls back instead of reaching media or input APIs.
        const value = parseInt(store)
        return (Number.isFinite(value) ? value : def) as T
      }
      case 'string':
        return store as T
      case 'boolean':
        return (store === '1') as T
      default:
        return def
    }
  }

  return def
}

// Recently used lists are stored as JSON. A damaged value starts an empty list
// instead of breaking the store or component that reads it on startup.
export function getList(key: string): string[] {
  try {
    const list = JSON.parse(get(key, '[]'))
    return Array.isArray(list) ? list.filter((item) => typeof item === 'string') : []
  } catch {
    return []
  }
}
