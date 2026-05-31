async function renderDashboard() {
  const user = getUser();
  if (!user) { location.hash = '#login'; return; }

  if (user.role === 'student') {
    await renderStudentDashboard(user);
  } else if (user.role === 'teacher') {
    await renderTeacherDashboard(user);
  } else {
    await renderAdminDashboard(user);
  }
}

async function renderStudentDashboard(user) {
  const theses = await api('/api/theses').catch(() => []);
  const active = theses[0];
  let points = [];
  let progress = 0;
  let stats = { total: 0, done: 0, inProgress: 0, overdue: 0 };

  if (active) {
    points = await api('/api/theses/' + active.id + '/points').catch(() => []);
    stats.total = points.length;
    stats.done = points.filter(p => p.status === 'done').length;
    stats.inProgress = points.filter(p => p.status === 'in_progress').length;
    stats.overdue = points.filter(p => isOverdue(p.deadline) && p.status !== 'done').length;
    progress = stats.total ? Math.round((stats.done / stats.total) * 100) : 0;
  }

  const notifs = await api('/api/notifications/unread').catch(() => []);

  app.innerHTML = `
    <h1 class="page-title">Добро пожаловать, ${user.full_name}</h1>
    <div class="stats-grid">
      <div class="stat-card"><div class="value">${stats.total}</div><div class="label">Всего пунктов</div></div>
      <div class="stat-card"><div class="value">${stats.done}</div><div class="label">Выполнено</div></div>
      <div class="stat-card"><div class="value">${stats.inProgress}</div><div class="label">В процессе</div></div>
      <div class="stat-card"><div class="value">${stats.overdue}</div><div class="label">Просрочено</div></div>
    </div>

    ${active ? `
    <div class="card">
      <h2 class="section-title">Активная курсовая: ${escapeHtml(active.title)}</h2>
      <div style="margin-bottom:0.5rem"><div class="progress-bar"><div class="progress-fill" style="width:${progress}%"></div></div></div>
      <p>Прогресс: ${progress}%</p>
      <p>Статус: <span class="status status-${active.status}">${statusLabel(active.status)}</span></p>
      <p>Дедлайн: ${formatDate(active.deadline)} ${isOverdue(active.deadline) ? '<span class="overdue">(просрочено)</span>' : ''}</p>
      <a href="#thesis/${active.id}" class="btn btn-primary btn-sm">Открыть курсовую</a>
    </div>
    ` : '<div class="card empty">У вас пока нет курсовой работы</div>'}

    <div class="card">
      <h2 class="section-title">Ближайшие дедлайны</h2>
      ${points.filter(p => p.status !== 'done').sort((a,b) => (a.deadline||'9999')>(b.deadline||'9999')?1:-1).slice(0,5).map(p => `
        <div class="point-card status-${p.status}">
          <strong>${escapeHtml(p.title)}</strong>
          <div>Дедлайн: ${formatDate(p.deadline)} ${isOverdue(p.deadline) ? '<span class="overdue">(просрочено)</span>' : ''}</div>
          <span class="status status-${p.status}">${statusLabel(p.status)}</span>
        </div>
      `).join('') || '<p class="empty">Нет данных</p>'}
    </div>

    ${notifs.length ? `
    <div class="card">
      <h2 class="section-title">Последние уведомления</h2>
      ${notifs.slice(0,3).map(n => `<p>• ${escapeHtml(n.title)}: ${escapeHtml(n.message)}</p>`).join('')}
      <a href="#notifications" class="btn btn-outline btn-sm">Все уведомления</a>
    </div>
    ` : ''}
  `;
}

async function renderTeacherDashboard(user) {
  const groups = await api('/api/groups').catch(() => []);
  const theses = await api('/api/theses').catch(() => []);
  const pendingReviews = [];
  for (const t of theses) {
    const points = await api('/api/theses/' + t.id + '/points').catch(() => []);
    for (const p of points) {
      const results = await api('/api/points/' + p.id + '/results').catch(() => []);
      pendingReviews.push(...results.filter(r => r.review_status === 'pending'));
    }
  }

  app.innerHTML = `
    <h1 class="page-title">Добро пожаловать, ${user.full_name}</h1>
    <div class="stats-grid">
      <div class="stat-card"><div class="value">${groups.length}</div><div class="label">Групп</div></div>
      <div class="stat-card"><div class="value">${theses.length}</div><div class="label">Курсовых</div></div>
      <div class="stat-card"><div class="value">${pendingReviews.length}</div><div class="label">На проверке</div></div>
    </div>

    <div class="card">
      <h2 class="section-title">Группы</h2>
      <div class="table-wrap">
        <table>
          <tr><th>Название</th><th>Курс</th><th>Год</th></tr>
          ${groups.map(g => `<tr><td><a href="#group/${g.id}">${escapeHtml(g.name)}</a></td><td>${g.course}</td><td>${g.year}</td></tr>`).join('')}
        </table>
      </div>
      <a href="#groups" class="btn btn-outline btn-sm" style="margin-top:0.5rem">Управление группами</a>
    </div>

    <div class="card">
      <h2 class="section-title">Ожидают проверки</h2>
      ${pendingReviews.length ? pendingReviews.slice(0,5).map(r => `
        <div class="point-card">
          <strong>Результат #${r.id}</strong>
          <p>${escapeHtml(r.content || 'Без описания')}</p>
          <a href="#thesis" class="btn btn-primary btn-sm">Проверить</a>
        </div>
      `).join('') : '<p class="empty">Нет работ на проверке</p>'}
    </div>
  `;
}

async function renderAdminDashboard(user) {
  const users = await api('/api/users').catch(() => []);
  const groups = await api('/api/groups').catch(() => []);
  const theses = await api('/api/theses').catch(() => []);

  app.innerHTML = `
    <h1 class="page-title">Администрирование</h1>
    <div class="stats-grid">
      <div class="stat-card"><div class="value">${users.length}</div><div class="label">Пользователей</div></div>
      <div class="stat-card"><div class="value">${groups.length}</div><div class="label">Групп</div></div>
      <div class="stat-card"><div class="value">${theses.length}</div><div class="label">Курсовых</div></div>
    </div>

    <div class="card">
      <h2 class="section-title">Быстрые действия</h2>
      <a href="#groups" class="btn btn-primary">Управление группами</a>
      <a href="#users" class="btn btn-outline">Пользователи</a>
      <a href="#theses" class="btn btn-outline">Курсовые работы</a>
    </div>
  `;
}

function statusLabel(s) {
  const map = {
    planning: 'Планирование', in_progress: 'В процессе', review: 'На проверке',
    completed: 'Завершена', failed: 'Не сдана',
    pending: 'Ожидает', done: 'Выполнено', rejected: 'Отклонено', approved: 'Одобрено'
  };
  return map[s] || s;
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}
