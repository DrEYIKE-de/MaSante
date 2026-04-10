// api.js — HTTP client for the MaSante REST API.
// All calls return {ok, data, error} to simplify error handling in pages.

const BASE = '/api/v1';

async function request(method, path, body) {
  try {
    const opts = {
      method,
      headers: {},
      credentials: 'same-origin',
    };
    if (body !== undefined) {
      opts.headers['Content-Type'] = 'application/json';
      opts.body = JSON.stringify(body);
    }
    const res = await fetch(BASE + path, opts);
    const data = await res.json().catch(() => null);

    if (!res.ok) {
      const msg = (data && data.error) || httpMessage(res.status);
      return { ok: false, status: res.status, error: msg, data: null };
    }
    return { ok: true, status: res.status, data, error: null };
  } catch (e) {
    return { ok: false, status: 0, error: 'Connexion au serveur impossible', data: null };
  }
}

function httpMessage(status) {
  const messages = {
    400: 'Requete invalide',
    401: 'Non authentifie',
    403: 'Acces refuse',
    404: 'Ressource introuvable',
    409: 'Conflit',
    429: 'Trop de tentatives',
    500: 'Erreur serveur',
  };
  return messages[status] || 'Erreur inconnue';
}

// Convenience methods.
export const api = {
  get: (path) => request('GET', path),
  post: (path, body) => request('POST', path, body),
  put: (path, body) => request('PUT', path, body),
  del: (path) => request('DELETE', path),
};

// Setup.
export const setup = {
  status: () => api.get('/setup/status'),
  center: (data) => api.post('/setup/center', data),
  admin: (data) => api.post('/setup/admin', data),
  schedule: (data) => api.post('/setup/schedule', data),
  sms: (data) => api.post('/setup/sms', data),
  complete: () => api.post('/setup/complete', {}),
};

// Auth.
export const auth = {
  login: (username, password) => api.post('/auth/login', { username, password }),
  logout: () => api.post('/auth/logout', {}),
  me: () => api.get('/auth/me'),
};

// Dashboard.
export const dashboard = {
  stats: () => api.get('/dashboard/stats'),
  today: () => api.get('/dashboard/today'),
  overdue: () => api.get('/dashboard/overdue'),
};

// Patients.
export const patients = {
  list: (params = '') => api.get('/patients' + (params ? '?' + params : '')),
  search: (q) => api.get('/patients/search?q=' + encodeURIComponent(q)),
  get: (id) => api.get('/patients/' + id),
  create: (data) => api.post('/patients', data),
  update: (id, data) => api.put('/patients/' + id, data),
  exit: (id, data) => api.put('/patients/' + id + '/exit', data),
};

// Appointments.
export const appointments = {
  create: (data) => api.post('/appointments', data),
  get: (id) => api.get('/appointments/' + id),
  complete: (id, data) => api.put('/appointments/' + id + '/complete', data),
  missed: (id, data) => api.put('/appointments/' + id + '/missed', data),
  reschedule: (id, data) => api.put('/appointments/' + id + '/reschedule', data),
  cancel: (id) => api.del('/appointments/' + id),
  slots: (date) => api.get('/appointments/slots?date=' + date),
};

// Calendar.
export const calendar = {
  week: (date) => api.get('/calendar/week?date=' + date),
};

// Reminders.
export const reminders = {
  list: () => api.get('/reminders'),
  stats: () => api.get('/reminders/stats'),
  templates: () => api.get('/reminders/templates'),
  updateTemplate: (id, data) => api.put('/reminders/templates/' + id, data),
  test: (to, message) => api.post('/reminders/test', { to, message }),
  sendAll: () => api.post('/reminders/send-all', {}),
};

// Users.
export const users = {
  list: () => api.get('/users'),
  create: (data) => api.post('/users', data),
  update: (id, data) => api.put('/users/' + id, data),
  disable: (id) => api.del('/users/' + id),
  resetPassword: (id, password) => api.put('/users/' + id + '/reset-password', { password }),
};

// Profile.
export const profile = {
  get: () => api.get('/profile'),
  update: (data) => api.put('/profile', data),
  changePassword: (current_password, new_password) => api.put('/profile/password', { current_password, new_password }),
  activity: () => api.get('/profile/activity'),
};

// Exports.
export const exports = {
  patientsExcel: (status) => '/api/v1/export/patients/excel' + (status ? '?status=' + status : ''),
  patientsPdf: (status) => '/api/v1/export/patients/pdf' + (status ? '?status=' + status : ''),
  monthlyExcel: (month) => '/api/v1/export/monthly/excel?month=' + month,
  monthlyPdf: (month) => '/api/v1/export/monthly/pdf?month=' + month,
};
