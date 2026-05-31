async function renderUsers() {
  const users = await api('/api/users').catch(() => []);
  const me = getUser();
  const isAdmin = me?.role === 'admin';
  app.innerHTML = `
    <h1 class="page-title">Пользователи</h1>
    <div class="table-wrap">
      <table>
        <tr><th>ID</th><th>ФИО</th><th>Email</th><th>Роль</th>${isAdmin ? '<th></th>' : ''}</tr>
        ${users.map(u => `
          <tr>
            <td>${u.id}</td>
            <td>${escapeHtml(u.full_name)}</td>
            <td>${escapeHtml(u.email)}</td>
            <td><span class="status status-${u.role === 'admin' ? 'danger' : u.role === 'teacher' ? 'warning' : 'primary'}">${u.role}</span></td>
            ${isAdmin && u.id !== me.id ? `<td><button class="btn btn-danger btn-sm" onclick="deleteUser(${u.id})">Удалить</button></td>` : isAdmin ? '<td></td>' : ''}
          </tr>
        `).join('')}
      </table>
    </div>
    ${users.length === 0 ? '<p class="empty">Пользователей нет</p>' : ''}
  `;

  window.deleteUser = async (id) => {
    if (!confirm('Удалить пользователя?')) return;
    await api('/api/users/' + id, { method: 'DELETE' });
    renderUsers();
  };
}

async function renderProfile() {
  const user = await api('/api/auth/me').catch(() => getUser());
  app.innerHTML = `
    <h1 class="page-title">Профиль</h1>
    <div class="card">
      <p><strong>ФИО:</strong> ${escapeHtml(user.full_name)}</p>
      <p><strong>Email:</strong> ${escapeHtml(user.email)}</p>
      <p><strong>Роль:</strong> ${user.role}</p>
    </div>
    <div class="card">
      <h2 class="section-title">Сменить пароль</h2>
      <form id="passForm">
        <div class="form-group"><label>Старый пароль</label><input type="password" id="oldPass" required></div>
        <div class="form-group"><label>Новый пароль</label><input type="password" id="newPass" required minlength="4"></div>
        <button type="submit" class="btn btn-primary">Сменить</button>
        <p id="passError" style="color:var(--danger);margin-top:0.5rem;font-size:0.85rem;"></p>
      </form>
    </div>
  `;
  document.getElementById('passForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    try {
      await api('/api/auth/password', {
        method: 'PUT',
        body: JSON.stringify({ old_password: document.getElementById('oldPass').value, new_password: document.getElementById('newPass').value })
      });
      document.getElementById('passError').textContent = 'Пароль изменен';
      document.getElementById('passError').style.color = 'var(--success)';
    } catch (err) {
      document.getElementById('passError').textContent = err.message;
      document.getElementById('passError').style.color = 'var(--danger)';
    }
  });
}
