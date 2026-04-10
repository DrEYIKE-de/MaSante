import { describe, it, expect } from 'vitest'
import { store } from '../store'

describe('store', () => {
  it('initializes with null user', () => {
    expect(store.user).toBeNull()
  })

  it('initializes with null setupDone', () => {
    expect(store.setupDone).toBeNull()
  })

  it('is reactive — user can be set', () => {
    store.user = { id: 1, username: 'admin', full_name: 'Dr Test', role: 'admin' }
    expect(store.user.username).toBe('admin')
    expect(store.user.role).toBe('admin')
    store.user = null // reset
  })

  it('setupStep starts at 0', () => {
    expect(store.setupStep).toBe(0)
  })
})
