const API_URL = '/todos';
let currentFilter = 'all';
let todos = [];

// Загрузка задач при старте
document.addEventListener('DOMContentLoaded', loadTodos);

// Обработчик формы добавления
document.getElementById('todoForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const title = document.getElementById('title').value.trim();
    const description = document.getElementById('description').value.trim();
    
    if (!title) return;
    
    try {
        const response = await fetch(API_URL, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ title, description })
        });
        
        if (response.ok) {
            document.getElementById('title').value = '';
            document.getElementById('description').value = '';
            loadTodos();
        }
    } catch (error) {
        console.error('Error creating todo:', error);
    }
});

// Обработчики фильтров
document.querySelectorAll('.filter-btn').forEach(btn => {
    btn.addEventListener('click', () => {
        document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        currentFilter = btn.dataset.filter;
        renderTodos();
    });
});

// Модальное окно
const modal = document.getElementById('editModal');
const closeBtn = document.querySelector('.close');

closeBtn.addEventListener('click', () => modal.classList.remove('show'));
modal.addEventListener('click', (e) => {
    if (e.target === modal) modal.classList.remove('show');
});

// Обработчик формы редактирования
document.getElementById('editForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const id = parseInt(document.getElementById('editId').value);
    const title = document.getElementById('editTitle').value.trim();
    const description = document.getElementById('editDescription').value.trim();
    const completed = document.getElementById('editCompleted').checked;
    
    try {
        const response = await fetch(`${API_URL}/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ title, description, completed })
        });
        
        if (response.ok) {
            modal.classList.remove('show');
            loadTodos();
        }
    } catch (error) {
        console.error('Error updating todo:', error);
    }
});

// Загрузка задач с сервера
async function loadTodos() {
    const todoList = document.getElementById('todoList');
    todoList.innerHTML = '<div class="loading">Загрузка...</div>';
    
    try {
        const response = await fetch(API_URL);
        todos = await response.json();
        renderTodos();
    } catch (error) {
        console.error('Error loading todos:', error);
        todoList.innerHTML = '<div class="empty-state"><span>⚠️</span>Ошибка загрузки</div>';
    }
}

// Отображение задач
function renderTodos() {
    const todoList = document.getElementById('todoList');
    
    let filteredTodos = todos;
    if (currentFilter === 'pending') {
        filteredTodos = todos.filter(t => !t.completed);
    } else if (currentFilter === 'completed') {
        filteredTodos = todos.filter(t => t.completed);
    }
    
    if (filteredTodos.length === 0) {
        todoList.innerHTML = `
            <div class="empty-state">
                <span>📝</span>
                ${currentFilter === 'all' ? 'Нет задач. Добавьте первую!' : 'Нет задач в этой категории'}
            </div>
        `;
        updateStats(0, 0, 0);
        return;
    }
    
    todoList.innerHTML = filteredTodos.map(todo => `
        <div class="todo-item ${todo.completed ? 'completed' : ''}">
            <input 
                type="checkbox" 
                class="todo-checkbox" 
                ${todo.completed ? 'checked' : ''}
                onchange="toggleComplete(${todo.id}, this.checked)"
            >
            <div class="todo-content">
                <div class="todo-title">${escapeHtml(todo.title)}</div>
                ${todo.description ? `<div class="todo-description">${escapeHtml(todo.description)}</div>` : ''}
            </div>
            <div class="todo-actions">
                <button class="edit-btn" onclick="openEditModal(${todo.id})">✏️</button>
                <button class="delete-btn" onclick="deleteTodo(${todo.id})">🗑️</button>
            </div>
        </div>
    `).join('');
    
    // Обновление статистики
    const total = todos.length;
    const completed = todos.filter(t => t.completed).length;
    const pending = total - completed;
    updateStats(total, completed, pending);
}

// Обновление статистики
function updateStats(total, completed, pending) {
    document.getElementById('totalCount').textContent = total;
    document.getElementById('completedCount').textContent = completed;
    document.getElementById('pendingCount').textContent = pending;
}

// Переключение статуса задачи
async function toggleComplete(id, completed) {
    const todo = todos.find(t => t.id === id);
    if (!todo) return;
    
    try {
        await fetch(`${API_URL}/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ 
                title: todo.title, 
                description: todo.description, 
                completed 
            })
        });
        loadTodos();
    } catch (error) {
        console.error('Error toggling todo:', error);
    }
}

// Открытие модального окна редактирования
function openEditModal(id) {
    const todo = todos.find(t => t.id === id);
    if (!todo) return;
    
    document.getElementById('editId').value = todo.id;
    document.getElementById('editTitle').value = todo.title;
    document.getElementById('editDescription').value = todo.description || '';
    document.getElementById('editCompleted').checked = todo.completed;
    
    modal.classList.add('show');
}

// Удаление задачи
async function deleteTodo(id) {
    if (!confirm('Вы уверены, что хотите удалить эту задачу?')) return;
    
    try {
        await fetch(`${API_URL}/${id}`, { method: 'DELETE' });
        loadTodos();
    } catch (error) {
        console.error('Error deleting todo:', error);
    }
}

// Экранирование HTML
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
