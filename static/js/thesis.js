async function renderTheses() {
  const list = await api('/api/theses').catch(() => []);
  const isAdmin = getUser()?.role === 'admin';
  const isTeacher = getUser()?.role === 'teacher';
  const students = (isAdmin || isTeacher) ? await api('/api/users/students').catch(() => []) : [];
  const groups = (isAdmin || isTeacher) ? await api('/api/groups').catch(() => []) : [];

  app.innerHTML = `
    <h1 class="page-title">Курсовые работы</h1>
    ${(isAdmin || isTeacher) ? `<button class="btn btn-primary" onclick="showThesisModal()">+ Создать курсовую</button>` : ''}
    <div class="table-wrap" style="margin-top:1rem">
      <table>
        <tr><th>Тема</th><th>Студент</th><th>Группа</th><th>Статус</th><th>Дедлайн</th></tr>
        ${list.map(t => `
          <tr>
            <td><a href="#thesis/${t.id}">${escapeHtml(t.title)}</a></td>
            <td>${escapeHtml(t.student_name || '—')}</td>
            <td>${t.group_id ? 'Группа #' + t.group_id : '—'}</td>
            <td><span class="status status-${t.status}">${statusLabel(t.status)}</span></td>
            <td>${formatDate(t.deadline)} ${isOverdue(t.deadline) && t.status !== 'completed' ? '<span class="overdue">!</span>' : ''}</td>
          </tr>
        `).join('')}
      </table>
    </div>
    ${list.length === 0 ? '<p class="empty">Курсовых работ пока нет</p>' : ''}
    <div id="thesisModal"></div>
  `;

  window.showThesisModal = () => {
    document.getElementById('thesisModal').innerHTML = `
      <div class="modal-overlay" onclick="if(event.target===this)closeThesisModal()">
        <div class="modal-content">
          <div class="modal-header"><h3>Новая курсовая</h3><button class="modal-close" onclick="closeThesisModal()">&times;</button></div>
          <form id="thesisForm">
            <div class="form-group">
              <label>Тип создания</label>
              <select id="thType" onchange="toggleThesisFields()">
                <option value="student">Для одного студента</option>
                <option value="group">Для всей группы</option>
              </select>
            </div>
            <div class="form-group" id="thStudentGroup">
              <label>Студент</label>
              <select id="thStudent">${students.map(s => `<option value="${s.id}">${escapeHtml(s.full_name)}</option>`).join('')}</select>
            </div>
            <div class="form-group" id="thGroupGroup" style="display:none">
              <label>Группа</label>
              <select id="thGroup">${groups.map(g => `<option value="${g.id}">${escapeHtml(g.name)}</option>`).join('')}</select>
            </div>
            <div class="form-group"><label>Тема</label><input type="text" id="thTitle" required></div>
            <div class="form-group"><label>Описание</label><textarea id="thDesc" rows="3"></textarea></div>
            <div class="grid-2">
              <div class="form-group"><label>Начало</label><input type="date" id="thStart"></div>
              <div class="form-group"><label>Дедлайн</label><input type="date" id="thDeadline"></div>
            </div>
            <button type="submit" class="btn btn-primary">Создать</button>
          </form>
        </div>
      </div>`;
    
    window.toggleThesisFields = () => {
      const type = document.getElementById('thType').value;
      document.getElementById('thStudentGroup').style.display = type === 'student' ? 'block' : 'none';
      document.getElementById('thGroupGroup').style.display = type === 'group' ? 'block' : 'none';
    };
    
    document.getElementById('thesisForm').addEventListener('submit', async (e) => {
      e.preventDefault();
      const type = document.getElementById('thType').value;
      const payload = {
        title: document.getElementById('thTitle').value,
        description: document.getElementById('thDesc').value,
        start_date: document.getElementById('thStart').value,
        deadline: document.getElementById('thDeadline').value,
      };
      
      if (type === 'student') {
        payload.student_id = parseInt(document.getElementById('thStudent').value);
      } else {
        payload.group_id = parseInt(document.getElementById('thGroup').value);
      }
      
      await api('/api/theses', {
        method: 'POST',
        body: JSON.stringify(payload)
      });
      closeThesisModal(); renderTheses();
    });
  };
  window.closeThesisModal = () => { document.getElementById('thesisModal').innerHTML = ''; };
}

async function renderThesisDetail(id) {
  const thesis = await api('/api/theses/' + id).catch(() => null);
  if (!thesis) { app.innerHTML = '<div class="card">Курсовая не найдена</div>'; return; }
  const points = await api('/api/theses/' + id + '/points').catch(() => []);
  const isAdmin = getUser()?.role === 'admin';
  const isTeacher = getUser()?.role === 'teacher';
  const isStudent = getUser()?.role === 'student';

  app.innerHTML = `
    <h1 class="page-title">${escapeHtml(thesis.title)}</h1>
    <div class="card">
      <p>${escapeHtml(thesis.description || 'Без описания')}</p>
      <p>Статус: <span class="status status-${thesis.status}">${statusLabel(thesis.status)}</span></p>
      <p>Студент: ${escapeHtml(thesis.student_name || '—')}</p>
      <p>Начало: ${formatDate(thesis.start_date)} | Дедлайн: ${formatDate(thesis.deadline)}</p>
      ${(isAdmin || isTeacher) ? `
        <div style="margin-top:0.5rem">
          <button class="btn btn-danger btn-sm" onclick="deleteThesis(${id})">Удалить курсовую</button>
        </div>
      ` : ''}
    </div>

    <div class="card">
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:0.5rem">
        <h2 class="section-title">Пункты плана</h2>
        ${(isAdmin || isTeacher) ? `<button class="btn btn-primary btn-sm" onclick="showPointModal(${id})">+ Добавить пункт</button>` : ''}
      </div>
      ${points.map((p, idx) => `
        <div class="point-card status-${p.status}">
          <div style="display:flex;justify-content:space-between;align-items:flex-start">
            <div>
              <strong>#${idx+1} ${escapeHtml(p.title)}</strong>
              <p style="color:var(--gray);font-size:0.85rem">${escapeHtml(p.description || '')}</p>
              <p style="font-size:0.85rem">Дедлайн: ${formatDate(p.deadline)} ${isOverdue(p.deadline) && p.status !== 'done' ? '<span class="overdue">(просрочено)</span>' : ''}</p>
            </div>
            <span class="status status-${p.status}">${statusLabel(p.status)}</span>
          </div>
          <div style="margin-top:0.5rem">
            <a href="#thesis/${id}/point/${p.id}" class="btn btn-primary btn-sm">Открыть</a>
            ${(isAdmin || isTeacher) ? `<button class="btn btn-danger btn-sm" onclick="deletePoint(${p.id},${id})">Удалить</button>` : ''}
          </div>
        </div>
      `).join('')}
      ${points.length === 0 ? '<p class="empty">Пунктов пока нет</p>' : ''}
    </div>
    <div id="pointModal"></div>
  `;

  window.showPointModal = () => {
    document.getElementById('pointModal').innerHTML = `
      <div class="modal-overlay" onclick="if(event.target===this)closePointModal()">
        <div class="modal-content">
          <div class="modal-header"><h3>Новый пункт</h3><button class="modal-close" onclick="closePointModal()">&times;</button></div>
          <form id="pointForm">
            <div class="form-group"><label>Название</label><input type="text" id="pTitle" required></div>
            <div class="form-group"><label>Описание</label><textarea id="pDesc" rows="3"></textarea></div>
            <div class="grid-2">
              <div class="form-group"><label>Порядок</label><input type="number" id="pOrder" value="${points.length+1}"></div>
              <div class="form-group"><label>Дедлайн</label><input type="date" id="pDeadline"></div>
            </div>
            <button type="submit" class="btn btn-primary">Добавить</button>
          </form>
        </div>
      </div>`;
    document.getElementById('pointForm').addEventListener('submit', async (e) => {
      e.preventDefault();
      await api('/api/theses/' + id + '/points', {
        method: 'POST',
        body: JSON.stringify({
          title: document.getElementById('pTitle').value,
          description: document.getElementById('pDesc').value,
          order: parseInt(document.getElementById('pOrder').value),
          deadline: document.getElementById('pDeadline').value,
        })
      });
      closePointModal(); renderThesisDetail(id);
    });
  };
  window.closePointModal = () => { document.getElementById('pointModal').innerHTML = ''; };
  window.deletePoint = async (pid, tid) => {
    if (!confirm('Удалить пункт?')) return;
    await api('/api/points/' + pid, { method: 'DELETE' });
    renderThesisDetail(tid);
  };
  window.deleteThesis = async (tid) => {
    if (!confirm('Удалить курсовую работу?')) return;
    await api('/api/theses/' + tid, { method: 'DELETE' });
    location.hash = '#theses';
  };
}
