async function renderPointDetail(thesisId, pointId) {
  const point = await api('/api/points/' + pointId).catch(() => null);
  if (!point) { app.innerHTML = '<div class="card">Пункт не найден</div>'; return; }
  const results = await api('/api/points/' + pointId + '/results').catch(() => []);
  const isStudent = getUser()?.role === 'student';
  const isTeacher = getUser()?.role === 'teacher';
  const isAdmin = getUser()?.role === 'admin';

  app.innerHTML = `
    <h1 class="page-title">${escapeHtml(point.title)}</h1>
    <div class="card">
      <p>${escapeHtml(point.description || 'Без описания')}</p>
      <p>Дедлайн: ${formatDate(point.deadline)} ${isOverdue(point.deadline) && point.status !== 'done' ? '<span class="overdue">(просрочено)</span>' : ''}</p>
      <p>Статус: <span class="status status-${point.status}">${statusLabel(point.status)}</span></p>
    </div>

    ${isStudent ? `
    <div class="card">
      <h2 class="section-title">Отправить результат</h2>
      <form id="resultForm" enctype="multipart/form-data">
        <div class="form-group"><label>Описание</label><textarea id="rContent" rows="3"></textarea></div>
        <div class="form-group"><label>Файл (макс. 10 МБ)</label><input type="file" id="rFile"></div>
        <button type="submit" class="btn btn-primary">Отправить</button>
      </form>
    </div>
    ` : ''}

    <div class="card">
      <h2 class="section-title">Результаты</h2>
      ${results.map(r => `
        <div class="point-card">
          <p>${escapeHtml(r.content || 'Без описания')}</p>
          ${r.file_url ? `<p><a href="${r.file_url}" target="_blank">📎 ${escapeHtml(r.file_name || 'Файл')}</a></p>` : ''}
          <p style="font-size:0.8rem;color:var(--gray)">Отправлено: ${new Date(r.submitted_at).toLocaleString('ru-RU')}</p>
          <p>Статус проверки: <span class="status status-${r.review_status}">${statusLabel(r.review_status)}</span></p>
          ${r.review ? `<p><strong>Комментарий:</strong> ${escapeHtml(r.review)}</p>` : ''}
          ${(isTeacher || isAdmin) && r.review_status === 'pending' ? `
            <div style="margin-top:0.5rem">
              <button class="btn btn-success btn-sm" onclick="showReviewModal(${r.id},'approved')">Одобрить</button>
              <button class="btn btn-danger btn-sm" onclick="showReviewModal(${r.id},'rejected')">Отклонить</button>
            </div>
          ` : ''}
        </div>
      `).join('')}
      ${results.length === 0 ? '<p class="empty">Результатов пока нет</p>' : ''}
    </div>

    <div id="reviewModal"></div>
  `;

  if (isStudent) {
    document.getElementById('resultForm').addEventListener('submit', async (e) => {
      e.preventDefault();
      const formData = new FormData();
      formData.append('content', document.getElementById('rContent').value);
      const file = document.getElementById('rFile').files[0];
      if (file) formData.append('file', file);
      const res = await fetch(API_BASE + '/api/points/' + pointId + '/results', {
        method: 'POST',
        headers: { 'Authorization': 'Bearer ' + getToken() },
        body: formData
      });
      if (!res.ok) { alert('Ошибка отправки'); return; }
      renderPointDetail(thesisId, pointId);
    });
  }

  window.showReviewModal = (resultId, status) => {
    document.getElementById('reviewModal').innerHTML = `
      <div class="modal-overlay" onclick="if(event.target===this)closeReviewModal()">
        <div class="modal-content">
          <div class="modal-header"><h3>${status === 'approved' ? 'Одобрить' : 'Отклонить'} результат</h3><button class="modal-close" onclick="closeReviewModal()">&times;</button></div>
          <form id="reviewForm">
            <div class="form-group"><label>Комментарий</label><textarea id="revText" rows="3"></textarea></div>
            <button type="submit" class="btn btn-primary">Сохранить</button>
          </form>
        </div>
      </div>`;
    document.getElementById('reviewForm').addEventListener('submit', async (e) => {
      e.preventDefault();
      await api('/api/results/' + resultId + '/review', {
        method: 'PUT',
        body: JSON.stringify({ review: document.getElementById('revText').value, review_status: status })
      });
      closeReviewModal(); renderPointDetail(thesisId, pointId);
    });
  };
  window.closeReviewModal = () => { document.getElementById('reviewModal').innerHTML = ''; };
}
