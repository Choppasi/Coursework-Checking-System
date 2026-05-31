function renderLogin() {
  hideNav();
  app.innerHTML = `
    <div class="login-container">
      <div class="login-box">
        <h1>📚 Курсовые работы</h1>
        <div class="tabs">
          <button class="tab active" onclick="switchTab('login')">Вход</button>
          <button class="tab" onclick="switchTab('register')">Регистрация</button>
        </div>
        <form id="loginForm">
          <div class="form-group"><label>Email</label><input type="email" id="loginEmail" required></div>
          <div class="form-group"><label>Пароль</label><input type="password" id="loginPassword" required></div>
          <button type="submit" class="btn btn-primary" style="width:100%">Войти</button>
          <p id="loginError" style="color:var(--danger);margin-top:0.5rem;font-size:0.85rem;"></p>
        </form>
      </div>
    </div>`;
  document.getElementById('loginForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const email = document.getElementById('loginEmail').value;
    const password = document.getElementById('loginPassword').value;
    try {
      const data = await api('/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password })
      });
      setToken(data.token);
      setUser(data.user);
      location.hash = '#dashboard';
    } catch (err) {
      document.getElementById('loginError').textContent = err.message;
    }
  });
}

function renderRegister() {
  hideNav();
  app.innerHTML = `
    <div class="login-container">
      <div class="login-box">
        <h1>📚 Курсовые работы</h1>
        <div class="tabs">
          <button class="tab" onclick="switchTab('login')">Вход</button>
          <button class="tab active" onclick="switchTab('register')">Регистрация</button>
        </div>
        <form id="regForm">
          <div class="form-group"><label>Email</label><input type="email" id="regEmail" required></div>
          <div class="form-group"><label>Пароль</label><input type="password" id="regPassword" required minlength="4"></div>
          <div class="form-group"><label>ФИО</label><input type="text" id="regFullName" required></div>
          <div class="form-group"><label>Роль</label>
            <select id="regRole">
              <option value="student">Студент</option>
              <option value="teacher">Преподаватель</option>
              <option value="admin">Администратор</option>
            </select>
          </div>
          <button type="submit" class="btn btn-primary" style="width:100%">Зарегистрироваться</button>
          <p id="regError" style="color:var(--danger);margin-top:0.5rem;font-size:0.85rem;"></p>
        </form>
      </div>
    </div>`;
  document.getElementById('regForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const body = {
      email: document.getElementById('regEmail').value,
      password: document.getElementById('regPassword').value,
      full_name: document.getElementById('regFullName').value,
      role: document.getElementById('regRole').value,
    };
    try {
      await api('/api/auth/register', { method: 'POST', body: JSON.stringify(body) });
      location.hash = '#login';
    } catch (err) {
      document.getElementById('regError').textContent = err.message;
    }
  });
}

function switchTab(tab) {
  if (tab === 'login') location.hash = '#login';
  else location.hash = '#register';
}
