async function renderNotifications() {
  const list = await api('/api/notifications').catch(() => []);

  app.innerHTML = `
    <h1 class="page-title">Уведомления</h1>
    <button class="btn btn-outline btn-sm" onclick="markAllRead()">Отметить все прочитанными</button>
    <div style="margin-top:1rem">
      ${list.map(n => `
        <div class="card" style="opacity:${n.is_read ? 0.7 : 1}">
          <div style="display:flex;justify-content:space-between">
            <strong>${escapeHtml(n.title)}</strong>
            <span style="font-size:0.8rem;color:var(--gray)">${new Date(n.created_at).toLocaleString('ru-RU')}</span>
          </div>
          <p>${escapeHtml(n.message)}</p>
          ${!n.is_read ? `<button class="btn btn-outline btn-sm" onclick="markRead(${n.id})">Отметить прочитанным</button>` : ''}
        </div>
      `).join('')}
    </div>
    ${list.length === 0 ? '<p class="empty">Уведомлений нет</p>' : ''}
  `;

  window.markRead = async (id) => {
    await api('/api/notifications/' + id + '/read', { method: 'PUT' });
    renderNotifications(); updateNotifBadge();
  };
  window.markAllRead = async () => {
    await api('/api/notifications/read-all', { method: 'PUT' });
    renderNotifications(); updateNotifBadge();
  };
}
