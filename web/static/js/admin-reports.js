// admin-reports.ts — интерфейс просмотра отчётов для ответственных и админов
let reports = [];
let projects = [];
let users = [];
let user = null;
let token = '';
const content = document.getElementById('content');
if (!content) {
    throw new Error('Не найден элемент #content');
}
// === Вспомогательные функции ===
function logout() {
    localStorage.removeItem('token');
    window.location.href = '/login';
}
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
// === Загрузка данных ===
async function loadCurrentUser() {
    try {
        const res = await fetch('/api/v1/users', {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        const allUsers = await res.json();
        const email = decodeEmailFromToken(token);
        if (email) {
            user = allUsers.find(u => u.email === email) || allUsers[0];
        }
        else {
            user = allUsers[0];
        }
    }
    catch (err) {
        console.error('Не удалось определить текущего пользователя');
        user = null;
    }
}
function decodeEmailFromToken(token) {
    if (!token)
        return null;
    try {
        const payload = token.split('.')[1];
        if (!payload)
            return null;
        const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
        const parsed = JSON.parse(decoded);
        return parsed.email || null;
    }
    catch (e) {
        console.warn('Не удалось декодировать JWT', e);
        return null;
    }
}
async function loadProjects() {
    try {
        const res = await fetch('/api/v1/projects', {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        projects = await res.json();
    }
    catch (err) {
        console.warn('Проекты не загружены');
        projects = [];
    }
}
async function loadUsers() {
    try {
        const res = await fetch('/api/v1/users', {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        users = await res.json();
    }
    catch (err) {
        console.warn('Пользователи не загружены');
        users = [];
    }
}
async function loadReports() {
    try {
        const dateFrom = document.getElementById('dateFrom')?.value || '';
        const dateTo = document.getElementById('dateTo')?.value || '';
        const projectFilter = document.getElementById('projectFilter')?.value || '';
        let url = '/api/v1/reports';
        const params = new URLSearchParams();
        if (dateFrom)
            params.append('date_from', dateFrom);
        if (dateTo)
            params.append('date_to', dateTo);
        if (projectFilter)
            params.append('project_id', projectFilter);
        params.append('current_user_id', user.id.toString()); // Добавляем ID пользователя
        params.append('current_user_role', user.role); // Добавляем роль пользователя
        if (params.toString()) {
            url += '?' + params.toString();
        }
        const res = await fetch(url, {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        if (res.status === 401)
            return logout();
        reports = await res.json();
        renderReports();
    }
    catch (err) {
        console.error('Ошибка загрузки отчётов:', err);
        showError('Не удалось загрузить отчёты');
    }
}
// === Рендеринг интерфейса ===
function renderPage() {
    const isManager = user?.role === 'admin' || projects.some(p => p.responsible_id === user?.id);
    content.innerHTML = `
        <link rel="stylesheet" href="/static/css/app.css">
        <link rel="stylesheet" href="/static/css/admin-reports.css?v=20260710-1">
        <div class="admin-reports-container">
            <h2>Просмотр отчётов</h2>
            <div class="filters">
                <div class="form-group">
                    <label>Период с</label>
                    <input type="date" id="dateFrom" />
                </div>
                <div class="form-group">
                    <label>по</label>
                    <input type="date" id="dateTo" />
                </div>
                ${user?.role === 'admin' ? `
                    <div class="form-group">
                        <label>Проект</label>
                        <select id="projectFilter">
                            <option value="">Все проекты</option>
                            ${projects.map(p => `<option value="${p.id}">${escapeHtml(p.name)}</option>`).join('')}
                        </select>
                    </div>
                ` : ''}
                <button id="searchBtn" class="btn btn-primary">Поиск</button>
            </div>
            <div id="reportsContainer" class="reports-list">
                <p>Загрузка отчётов...</p>
            </div>
        </div>
    `;
    setupEventListeners();
}
function setupEventListeners() {
    document.getElementById('searchBtn')?.addEventListener('click', () => loadReports());
    document.getElementById('dateFrom')?.addEventListener('change', () => loadReports());
    document.getElementById('dateTo')?.addEventListener('change', () => loadReports());
    document.getElementById('projectFilter')?.addEventListener('change', () => loadReports());
}
// ... предыдущие импорты и интерфейсы без изменений ...
// === Вспомогательная функция для безопасного экранирования HTML и преобразования ссылок ===
function escapeHtmlAndLinkify(text) {
    if (!text)
        return '';
    // Сначала экранируем HTML-символы
    let escaped = text
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
    // Регулярное выражение для поиска URL (http/https/www)
    // Поддерживает: http://, https://, www. в начале URL
    const urlRegex = /(https?:\/\/[^\s<>"{}|\\^`\[\]]+|www\.[^\s<>"{}|\\^`\[\]]+)/g;
    // Функция для создания ссылки с target="_blank"
    const makeLink = (match) => {
        let url = match;
        if (match.startsWith('www.')) {
            url = 'http://' + match;
        }
        return `<a href="${url}" target="_blank" rel="noopener noreferrer" class="report-link">${match}</a>`;
    };
    return escaped.replace(urlRegex, makeLink);
}
// === Обновлённая функция renderReports с поддержкой ссылок ===
function renderReports() {
    const container = document.getElementById('reportsContainer');
    if (!container)
        return;
    if (reports.length === 0) {
        container.innerHTML = '<p>Нет отчётов, соответствующих фильтрам.</p>';
        return;
    }
    // Сортировка по дате (новые — сначала)
    const sortedReports = [...reports].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
    container.innerHTML = `
        <h3>Найдено отчётов: ${reports.length}</h3>
        ${sortedReports.map(r => {
        const date = new Date(r.created_at).toLocaleString();
        const project = projects.find(p => p.id === r.project_id);
        const reportUser = users.find(u => u.id === r.user_id);
        const formattedText = escapeHtmlAndLinkify(r.text);
        const canManageReport = user?.role === 'admin' || user?.id === r.user_id;
        const actions = canManageReport ? `
                <div class="report-actions">
                    <button class="btn btn-neutral btn-sm edit-report-btn" data-id="${r.id}">Редактировать</button>
                    <button class="btn btn-danger btn-sm delete-report-btn" data-id="${r.id}">Удалить</button>
                </div>
            ` : '';
        // Форма редактирования — изначально скрыта
        const editForm = canManageReport ? `
                <div class="edit-report-form" id="edit-form-${r.id}" style="display: none; margin-top: 10px; padding: 10px; background: #2d2d2d; border-radius: 6px;">
                    <div class="form-group">
                        <label>Проект</label>
                        <select class="edit-project-select">
                            <option value="">Не привязан к проекту</option>
                            ${projects.map(p => `<option value="${p.id}" ${p.id === r.project_id ? 'selected' : ''}>${escapeHtml(p.name)}</option>`).join('')}
                        </select>
                    </div>
                    <div class="form-group">
                        <label>Текст отчёта</label>
                        <textarea class="edit-report-text" rows="4" style="width: 100%;">${escapeHtml(r.text)}</textarea>
                    </div>
                    <div class="form-group">
                        <label>Переработки (часы)</label>
                        <input type="number" class="edit-overwork-input" value="${r.overwork}" min="0" step="1" style="width: 100px;" />
                    </div>
                    <div class="form-group">
                        <label>Дата создания</label>
                        <input type="datetime-local" class="edit-date-input" value="${formatDateTimeLocal(r.created_at)}" style="width: 100%;" />
                    </div>
                    <div>
                        <button class="btn btn-primary btn-sm save-report-btn" data-id="${r.id}">Сохранить</button>
                        <button class="btn btn-neutral btn-sm cancel-edit-btn" data-id="${r.id}" style="margin-left: 8px;">Отмена</button>
                    </div>
                </div>
            ` : '';
        return `
                <div class="report-item" data-id="${r.id}">
                    <div class="report-field">
                        <span class="report-label">Дата:</span>
                        <span class="report-value">${date}</span>
                    </div>
                    <div class="report-field">
                        <span class="report-label">Пользователь:</span>
                        <span class="report-value">${reportUser ? escapeHtml(reportUser.name) : `ID ${r.user_id}`}</span>
                    </div>
                    <div class="report-field">
                        <span class="report-label">Проект:</span>
                        <span class="report-value">${project ? escapeHtml(project.name) : `ID ${r.project_id}`}</span>
                    </div>
                    <div class="report-field">
                        <span class="report-label">Текст:</span>
                        <div class="report-value report-text">${formattedText}</div>
                    </div>
                    <div class="report-field">
                        <span class="report-label">Переработки:</span>
                        <span class="report-value">${r.overwork || 0} час(ов)</span>
                    </div>
                    ${actions}
                    ${editForm}
                </div>
            `;
    }).join('')}
    `;
    // Назначаем обработчики
    attachEditEventListeners();
    attachDeleteEventListeners();
}
// === Вспомогательная функция для форматирования даты в datetime-local ===
function formatDateTimeLocal(dateString) {
    const date = new Date(dateString);
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    return `${year}-${month}-${day}T${hours}:${minutes}`;
}
// === Функции редактирования ===
function attachEditEventListeners() {
    document.querySelectorAll('.edit-report-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const id = e.target.getAttribute('data-id');
            openEditForm(Number(id));
        });
    });
    document.querySelectorAll('.save-report-btn').forEach(btn => {
        btn.addEventListener('click', async (e) => {
            const id = e.target.getAttribute('data-id');
            await saveReportEdit(Number(id));
        });
    });
    document.querySelectorAll('.cancel-edit-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const id = e.target.getAttribute('data-id');
            cancelEdit(Number(id));
        });
    });
}
function openEditForm(id) {
    document.querySelectorAll('.edit-report-form').forEach(form => {
        form.style.display = 'none';
    });
    const form = document.getElementById(`edit-form-${id}`);
    if (form) {
        form.style.display = 'block';
    }
    const item = document.querySelector(`.report-item[data-id="${id}"]`);
    if (item) {
        const textSpan = item.querySelector('.report-text');
        if (textSpan)
            textSpan.innerHTML = '<em>Редактируется...</em>';
    }
}
async function saveReportEdit(id) {
    const form = document.getElementById(`edit-form-${id}`);
    const projectSelect = form.querySelector('.edit-project-select');
    const textArea = form.querySelector('.edit-report-text');
    const overworkInput = form.querySelector('.edit-overwork-input');
    const dateInput = form.querySelector('.edit-date-input');
    const project_id = projectSelect.value ? Number(projectSelect.value) : null;
    const text = textArea.value.trim();
    const overwork = overworkInput.value ? parseInt(overworkInput.value, 10) : 0;
    const created_at = dateInput.value ? new Date(dateInput.value).toISOString() : new Date().toISOString();
    if (!text) {
        alert('Текст отчёта не может быть пустым');
        return;
    }
    if (isNaN(overwork) || overwork < 0) {
        alert('Переработки должны быть неотрицательным числом');
        return;
    }
    if (!created_at) {
        alert('Некорректная дата создания');
        return;
    }
    const updatedReport = {
        project_id,
        text,
        overwork,
        created_at // Отправляем обновлённую дату
    };
    try {
        const res = await fetch(`/api/v1/reports/${id}`, {
            method: 'PUT',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(updatedReport)
        });
        if (res.ok) {
            const report = reports.find(r => r.id === id);
            if (report) {
                report.project_id = project_id;
                report.text = text;
                report.overwork = overwork;
                report.created_at = created_at;
            }
            renderReports();
            alert('Отчёт обновлён');
        }
        else {
            const err = await res.json().catch(() => ({ error: 'Ошибка при обновлении' }));
            alert(`Ошибка: ${err.error}`);
        }
    }
    catch (err) {
        alert('Ошибка сети при сохранении');
    }
}
function cancelEdit(id) {
    const form = document.getElementById(`edit-form-${id}`);
    if (form) {
        form.style.display = 'none';
    }
    renderReports();
}
// === Функции удаления ===
function attachDeleteEventListeners() {
    document.querySelectorAll('.delete-report-btn').forEach(btn => {
        btn.addEventListener('click', async (e) => {
            const reportId = e.target.getAttribute('data-id');
            if (!confirm('Вы уверены, что хотите удалить этот отчёт?')) {
                return;
            }
            try {
                const res = await fetch(`/api/v1/reports/${reportId}`, {
                    method: 'DELETE',
                    headers: {
                        'Authorization': `Bearer ${token}`
                    }
                });
                if (res.ok) {
                    reports = reports.filter(r => r.id !== Number(reportId));
                    renderReports();
                    alert('Отчёт успешно удалён');
                }
                else {
                    const err = await res.json().catch(() => ({ error: 'Ошибка удаления' }));
                    alert(`Ошибка: ${err.error}`);
                }
            }
            catch (err) {
                alert('Ошибка сети при удалении');
            }
        });
    });
}
// ... остальной код без изменений ...
function showError(msg) {
    const container = document.getElementById('reportsContainer');
    if (container) {
        container.innerHTML = `<div class="error">${escapeHtml(msg)}</div>`;
    }
}
// === Инициализация ===
document.addEventListener('DOMContentLoaded', async () => {
    token = localStorage.getItem('token');
    if (!token)
        return logout();
    await loadCurrentUser();
    await Promise.all([
        loadProjects(),
        loadUsers()
    ]);
    renderPage();
    await loadReports();
});
// === Выход ===
document.getElementById('logoutBtn')?.addEventListener('click', logout);
export {};
