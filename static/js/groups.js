async function renderGroups() {
  const groups = await api('/api/groups').catch(() => []);
  const teachers = await api('/api/users/teachers').catch(() => []);
  const user = getUser();
  const isAdmin = user?.role === 'admin';
  const isTeacher = user?.role === 'teacher';
  const canCreate = isAdmin || isTeacher;

  app.innerHTML = `
    <h1 class="page-title">Группы</h1>
    ${canCreate ? `<button class="btn btn-primary" onclick="showGroupModal()">+ Создать группу</button>` : ''}
    <div class="table-wrap" style="margin-top:1rem">
      <table>
        <tr><th>Название</th><th>Курс</th><th>Год</th><th>Преподаватель</th>${isAdmin ? '<th></th>' : ''}</tr>
        ${groups.map(g => `
          <tr>
            <td><a href="#group/${g.id}">${escapeHtml(g.name)}</a></td>
            <td>${g.course}</td>
            <td>${g.year}</td>
            <td>${escapeHtml(g.teacher_name || '—')}</td>
            ${isAdmin ? `<td><button class="btn btn-danger btn-sm" onclick="deleteGroup(${g.id})">Удалить</button></td>` : ''}
          </tr>
        `).join('')}
      </table>
    </div>
    ${groups.length === 0 ? '<p class="empty">Групп пока нет</p>' : ''}

    <div id="groupModal"></div>
  `;

  window.showGroupModal = () => {
    const currentTeacherId = user?.id;
    document.getElementById('groupModal').innerHTML = `
      <div class="modal-overlay" onclick="if(event.target===this)closeGroupModal()">
        <div class="modal-content">
          <div class="modal-header"><h3>Новая группа</h3><button class="modal-close" onclick="closeGroupModal()">&times;</button></div>
          <form id="groupForm">
            <div class="form-group"><label>Название</label><input type="text" id="gName" required></div>
            ${isAdmin ? `
            <div class="form-group"><label>Преподаватель</label>
              <select id="gTeacher">${teachers.map(t => `<option value="${t.id}">${escapeHtml(t.full_name)}</option>`).join('')}</select>
            </div>` : `
            <input type="hidden" id="gTeacher" value="${currentTeacherId}">
            `}
            <div class="grid-2">
              <div class="form-group"><label>Курс</label><input type="number" id="gCourse" value="1" min="1" max="6"></div>
              <div class="form-group"><label>Год поступления</label><input type="number" id="gYear" value="2024"></div>
            </div>
            <button type="submit" class="btn btn-primary">Создать</button>
          </form>
        </div>
      </div>`;
    document.getElementById('groupForm').addEventListener('submit', async (e) => {
      e.preventDefault();
      await api('/api/groups', {
        method: 'POST',
        body: JSON.stringify({
          name: document.getElementById('gName').value,
          teacher_id: parseInt(document.getElementById('gTeacher').value),
          course: parseInt(document.getElementById('gCourse').value),
          year: parseInt(document.getElementById('gYear').value),
        })
      });
      closeGroupModal(); renderGroups();
    });
  };
  window.closeGroupModal = () => { document.getElementById('groupModal').innerHTML = ''; };
  window.deleteGroup = async (id) => {
    if (!confirm('Удалить группу?')) return;
    await api('/api/groups/' + id, { method: 'DELETE' });
    renderGroups();
  };
}

async function renderGroupDetail(id) {
  const data = await api('/api/groups/' + id).catch(() => null);
  if (!data) { app.innerHTML = '<div class="card">Группа не найдена</div>'; return; }
  const group = data.group;
  const members = data.members || [];
  const allStudents = await api('/api/users/students').catch(() => []);
  const isAdmin = getUser()?.role === 'admin';
  const isTeacher = getUser()?.role === 'teacher';
  const isStudent = getUser()?.role === 'student';
  const currentUser = getUser();
  const isMember = members.some(m => m.id === currentUser?.id);

  app.innerHTML = `
    <h1 class="page-title">${escapeHtml(group.name)}</h1>
    <div class="card">
      <p>Курс: ${group.course} | Год поступления: ${group.year}</p>
      <p>Преподаватель: ${escapeHtml(group.teacher_name || '—')}</p>
      ${isStudent && !isMember ? `<button class="btn btn-success" onclick="joinGroup(${group.id})">Записаться в группу</button>` : ''}
      ${isMember ? `<span class="status status-completed">Вы состоите в этой группе</span>` : ''}
    </div>

    <div class="card">
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:0.5rem">
        <h2 class="section-title">Студенты</h2>
        ${(isAdmin || isTeacher) ? `<button class="btn btn-primary btn-sm" onclick="showAddMember()">+ Добавить</button>` : ''}
      </div>
      <div class="table-wrap">
        <table>
          <tr><th>ФИО</th><th>Email</th>${(isAdmin || isTeacher) ? '<th></th>' : ''}</tr>
          ${members.map(m => `
            <tr>
              <td>${escapeHtml(m.full_name)}${m.id === currentUser?.id ? ' (вы)' : ''}</td>
              <td>${escapeHtml(m.email)}</td>
              ${(isAdmin || isTeacher) ? `<td><button class="btn btn-danger btn-sm" onclick="removeMember(${group.id},${m.id})">Исключить</button></td>` : ''}
            </tr>
          `).join('')}
        </table>
      </div>
      ${members.length === 0 ? '<p class="empty">В группе пока нет студентов</p>' : ''}
    </div>

    <div id="memberModal"></div>
  `;

  window.joinGroup = async (groupId) => {
    if (!confirm('Записаться в эту группу?')) return;
    try {
      await api('/api/groups/' + groupId + '/join', { method: 'POST' });
      alert('Вы успешно записались в группу!');
      renderGroupDetail(id);
    } catch (err) {
      const res = await err.json().catch(() => ({}));
      alert(res.error || 'Ошибка при записи в группу');
    }
  };

  window.showAddMember = () => {
    const available = allStudents.filter(s => !members.find(m => m.id === s.id));
    document.getElementById('memberModal').innerHTML = `
      <div class="modal-overlay" onclick="if(event.target===this)closeMemberModal()">
        <div class="modal-content">
          <div class="modal-header"><h3>Добавить студента</h3><button class="modal-close" onclick="closeMemberModal()">&times;</button></div>
          <form id="memberForm">
            <div class="form-group"><label>Студент</label>
              <select id="mStudent">${available.map(s => `<option value="${s.id}">${escapeHtml(s.full_name)} (${s.email})</option>`).join('')}</select>
            </div>
            <button type="submit" class="btn btn-primary">Добавить</button>
          </form>
        </div>
      </div>`;
    document.getElementById('memberForm').addEventListener('submit', async (e) => {
      e.preventDefault();
      await api('/api/groups/' + id + '/members', {
        method: 'POST',
        body: JSON.stringify({ student_id: parseInt(document.getElementById('mStudent').value) })
      });
      closeMemberModal(); renderGroupDetail(id);
    });
  };
  window.closeMemberModal = () => { document.getElementById('memberModal').innerHTML = ''; };
  window.removeMember = async (gid, sid) => {
    if (!confirm('Исключить студента?')) return;
    await api('/api/groups/' + gid + '/members/' + sid, { method: 'DELETE' });
    renderGroupDetail(id);
  };
}
