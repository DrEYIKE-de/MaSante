const B = '/api/v1'

async function request(method, path, body) {
  try {
    const opts = { method, headers: {}, credentials: 'same-origin' }
    if (body !== undefined) {
      opts.headers['Content-Type'] = 'application/json'
      opts.body = JSON.stringify(body)
    }
    const res = await fetch(B + path, opts)
    const data = await res.json().catch(() => null)

    if (res.status === 401 && !path.includes('/auth/')) {
      window.location.hash = '#/login'
      return { ok: false, status: 401, error: 'Session expiree', data: null }
    }
    if (!res.ok) {
      return { ok: false, status: res.status, error: (data && data.error) || 'Erreur', data: null }
    }
    return { ok: true, status: res.status, data, error: null }
  } catch (e) {
    return { ok: false, status: 0, error: 'Connexion impossible', data: null }
  }
}

const api = {
  get: (p) => request('GET', p),
  post: (p, b) => request('POST', p, b),
  put: (p, b) => request('PUT', p, b),
  del: (p) => request('DELETE', p),
}

export const setup = {
  status: () => api.get('/setup/status'),
  center: (d) => api.post('/setup/center', d),
  admin: (d) => api.post('/setup/admin', d),
  schedule: (d) => api.post('/setup/schedule', d),
  sms: (d) => api.post('/setup/sms', d),
  complete: () => api.post('/setup/complete', {}),
}

export const auth = {
  login: (u, p) => api.post('/auth/login', { username: u, password: p }),
  logout: () => api.post('/auth/logout', {}),
  me: () => api.get('/auth/me'),
}

export const dashboard = {
  stats: () => api.get('/dashboard/stats'),
  today: () => api.get('/dashboard/today'),
  overdue: () => api.get('/dashboard/overdue'),
}

export const patients = {
  list: (params) => api.get('/patients' + (params ? '?' + params : '')),
  search: (q) => api.get('/patients/search?q=' + encodeURIComponent(q)),
  get: (id) => api.get('/patients/' + id),
  create: (d) => api.post('/patients', d),
  update: (id, d) => api.put('/patients/' + id, d),
  exit: (id, d) => api.put('/patients/' + id + '/exit', d),
}

export const appointments = {
  create: (d) => api.post('/appointments', d),
  get: (id) => api.get('/appointments/' + id),
  complete: (id, d) => api.put('/appointments/' + id + '/complete', d),
  missed: (id, d) => api.put('/appointments/' + id + '/missed', d),
  reschedule: (id, d) => api.put('/appointments/' + id + '/reschedule', d),
  cancel: (id) => api.del('/appointments/' + id),
  slots: (date) => api.get('/appointments/slots?date=' + date),
}

export const calendar = {
  week: (date) => api.get('/calendar/week?date=' + date),
}

export const reminders = {
  list: () => api.get('/reminders'),
  stats: () => api.get('/reminders/stats'),
  templates: () => api.get('/reminders/templates'),
  updateTemplate: (id, d) => api.put('/reminders/templates/' + id, d),
  test: (to, msg) => api.post('/reminders/test', { to, message: msg }),
  sendAll: () => api.post('/reminders/send-all', {}),
}

export const users = {
  list: () => api.get('/users'),
  create: (d) => api.post('/users', d),
  update: (id, d) => api.put('/users/' + id, d),
  disable: (id) => api.del('/users/' + id),
  resetPassword: (id, p) => api.put('/users/' + id + '/reset-password', { password: p }),
}

export const profile = {
  get: () => api.get('/profile'),
  update: (d) => api.put('/profile', d),
  changePassword: (c, n) => api.put('/profile/password', { current_password: c, new_password: n }),
  activity: () => api.get('/profile/activity'),
}
