const API_BASE = '';
let currentUser = null;

function getToken() { return localStorage.getItem('token'); }
function setToken(t) { localStorage.setItem('token', t); }
function removeToken() { localStorage.removeItem('token'); }
function setUser(u) { currentUser = u; localStorage.setItem('user', JSON.stringify(u)); }
function getUser() { const u = localStorage.getItem('user'); return u ? JSON.parse(u) : null; }
function removeUser() { currentUser = null; localStorage.removeItem('user'); }

async function api(path, opts = {}) {
  const url = API_BASE + path;
  const headers = { 'Content-Type': 'application/json', ...opts.headers };
  const token = getToken();
  if (token) headers['Authorization'] = 'Bearer ' + token;
  const res = await fetch(url, { ...opts, headers });
  if (res.status === 401) {
    removeToken(); removeUser(); location.hash = '#login'; return null;
  }
  if (res.status === 204) return null;
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || 'Ошибка ' + res.status);
  return data;
}

function formatDate(d) {
  if (!d) return '—';
  const date = new Date(d);
  return date.toLocaleDateString('ru-RU');
}

function isOverdue(dateStr) {
  if (!dateStr) return false;
  return new Date(dateStr) < new Date(new Date().setHours(0,0,0,0));
}

const app = document.getElementById('app');
const nav = document.getElementById('mainNav');

function showNav() {
  nav.classList.remove('hidden');
  const role = currentUser?.role;
  document.querySelectorAll('.admin-only').forEach(el => el.classList.toggle('hidden', role !== 'admin'));
  document.querySelectorAll('.teacher-only').forEach(el => el.classList.toggle('hidden', role !== 'teacher' && role !== 'admin'));
  updateNotifBadge();
}
function hideNav() { nav.classList.add('hidden'); }

function route() {
  const hash = location.hash.replace('#', '') || 'dashboard';
  const parts = hash.split('/');
  const page = parts[0];
  const id = parts[1];
  const subPage = parts[2];
  const subId = parts[3];

  if (!getToken() && page !== 'login' && page !== 'register') {
    location.hash = '#login';
    return;
  }
  if (getToken() && (page === 'login' || page === 'register')) {
    location.hash = '#dashboard';
    return;
  }

  currentUser = getUser();
  if (getToken()) showNav(); else hideNav();

  app.innerHTML = '<div class="loading">Загрузка...</div>';
  try {
    if (page === 'thesis' && id && subPage === 'point' && subId) {
      renderPointDetail(parseInt(id), parseInt(subId));
    } else if (page === 'login') {
      renderLogin();
    } else if (page === 'register') {
      renderRegister();
    } else if (page === 'dashboard') {
      renderDashboard();
    } else if (page === 'theses') {
      renderTheses();
    } else if (page === 'thesis') {
      renderThesisDetail(parseInt(id));
    } else if (page === 'groups') {
      renderGroups();
    } else if (page === 'group') {
      renderGroupDetail(parseInt(id));
    } else if (page === 'users') {
      renderUsers();
    } else if (page === 'notifications') {
      renderNotifications();
    } else if (page === 'profile') {
      renderProfile();
    } else {
      renderDashboard();
    }
  } catch (e) {
    app.innerHTML = '<div class="card">Ошибка загрузки страницы</div>';
    console.error(e);
  }
}

window.addEventListener('hashchange', route);
document.addEventListener('DOMContentLoaded', () => {
  document.getElementById('logoutBtn').addEventListener('click', (e) => {
    e.preventDefault();
    removeToken(); removeUser(); location.hash = '#login';
  });
  route();
});

function updateNotifBadge() {
  const badge = document.getElementById('notifBadge');
  if (!badge || !currentUser) return;
  api('/api/notifications/unread').then(list => {
    const count = list?.length || 0;
    badge.textContent = count;
    badge.classList.toggle('hidden', count === 0);
  }).catch(() => {});
}


