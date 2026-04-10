import { describe, it, expect, vi, beforeEach } from 'vitest'

// Mock fetch globally.
const mockFetch = vi.fn()
global.fetch = mockFetch

// Import after mocking fetch.
const { setup, auth, patients, appointments, users, profile, reminders } = await import('../api')

function mockResponse(data, status = 200) {
  return { ok: status >= 200 && status < 300, status, json: () => Promise.resolve(data) }
}

beforeEach(() => { mockFetch.mockReset() })

describe('API client', () => {
  describe('setup', () => {
    it('status calls GET /setup/status', async () => {
      mockFetch.mockResolvedValue(mockResponse({ setup_complete: false, current_step: 0 }))
      const res = await setup.status()
      expect(res.ok).toBe(true)
      expect(res.data.setup_complete).toBe(false)
      expect(mockFetch).toHaveBeenCalledWith('/api/v1/setup/status', expect.objectContaining({ method: 'GET' }))
    })

    it('center calls POST /setup/center', async () => {
      mockFetch.mockResolvedValue(mockResponse({ status: 'ok' }))
      const data = { name: 'Test', type: 'centre_sante', country: 'Cameroun', city: 'Douala' }
      const res = await setup.center(data)
      expect(res.ok).toBe(true)
      expect(mockFetch).toHaveBeenCalledWith('/api/v1/setup/center', expect.objectContaining({ method: 'POST' }))
    })

    it('admin sends correct JSON body', async () => {
      mockFetch.mockResolvedValue(mockResponse({ status: 'ok' }))
      await setup.admin({ full_name: 'Dr Test', username: 'admin', password: 'password1', email: 'a@b.cm', title: 'Dr' })
      const body = JSON.parse(mockFetch.mock.calls[0][1].body)
      expect(body.full_name).toBe('Dr Test')
      expect(body.username).toBe('admin')
      expect(body.password).toBe('password1')
    })
  })

  describe('auth', () => {
    it('login sends credentials', async () => {
      mockFetch.mockResolvedValue(mockResponse({ id: 1, username: 'admin', role: 'admin' }))
      const res = await auth.login('admin', 'password1')
      expect(res.ok).toBe(true)
      expect(res.data.username).toBe('admin')
      const body = JSON.parse(mockFetch.mock.calls[0][1].body)
      expect(body.username).toBe('admin')
      expect(body.password).toBe('password1')
    })

    it('login error returns error message', async () => {
      mockFetch.mockResolvedValue(mockResponse({ error: 'mot de passe incorrect' }, 401))
      const res = await auth.login('admin', 'wrong')
      expect(res.ok).toBe(false)
      expect(res.error).toBe('mot de passe incorrect')
    })

    it('me calls GET /auth/me', async () => {
      mockFetch.mockResolvedValue(mockResponse({ id: 1, username: 'admin' }))
      await auth.me()
      expect(mockFetch).toHaveBeenCalledWith('/api/v1/auth/me', expect.objectContaining({ method: 'GET' }))
    })
  })

  describe('patients', () => {
    it('create sends patient data', async () => {
      mockFetch.mockResolvedValue(mockResponse({ ID: 1, Code: 'MS-2026-00001' }, 201))
      const res = await patients.create({ last_name: 'Test', first_name: 'Patient', sex: 'M' })
      expect(res.ok).toBe(true)
      expect(res.data.Code).toBe('MS-2026-00001')
    })

    it('search encodes query', async () => {
      mockFetch.mockResolvedValue(mockResponse([]))
      await patients.search('test query')
      expect(mockFetch).toHaveBeenCalledWith('/api/v1/patients/search?q=test%20query', expect.anything())
    })

    it('exit sends reason', async () => {
      mockFetch.mockResolvedValue(mockResponse({ status: 'ok' }))
      await patients.exit(1, { reason: 'deces', date: '2026-04-10', notes: 'test' })
      const body = JSON.parse(mockFetch.mock.calls[0][1].body)
      expect(body.reason).toBe('deces')
    })
  })

  describe('appointments', () => {
    it('create sends appointment', async () => {
      mockFetch.mockResolvedValue(mockResponse({ ID: 1 }, 201))
      await appointments.create({ patient_id: 1, date: '2026-04-15', time: '10:00', type: 'consultation' })
      const body = JSON.parse(mockFetch.mock.calls[0][1].body)
      expect(body.patient_id).toBe(1)
      expect(body.time).toBe('10:00')
    })

    it('slots encodes date', async () => {
      mockFetch.mockResolvedValue(mockResponse([]))
      await appointments.slots('2026-04-15')
      expect(mockFetch).toHaveBeenCalledWith('/api/v1/appointments/slots?date=2026-04-15', expect.anything())
    })
  })

  describe('error handling', () => {
    it('network error returns connection message', async () => {
      mockFetch.mockRejectedValue(new Error('network'))
      const res = await auth.me()
      expect(res.ok).toBe(false)
      expect(res.error).toBe('Connexion impossible')
    })

    it('500 returns server error', async () => {
      mockFetch.mockResolvedValue(mockResponse({ error: 'erreur interne du serveur' }, 500))
      const res = await patients.list()
      expect(res.ok).toBe(false)
      expect(res.error).toBe('erreur interne du serveur')
    })
  })
})
