// timesheet.ts — интерфейс просмотра табеля по переработкам
let reports = [];
let users = [];
let user = null;
let token = '';
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
function getDateRange(startDate, endDate) {
    const dates = [];
    const current = new Date(startDate);
    const end = new Date(endDate);
    while (current <= end) {
        dates.push(current.toISOString().split('T')[0]);
        current.setDate(current.getDate() + 1);
    }
    return dates;
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
async function loadUsers() {
    try {
        const res = await fetch('/api/v1/users', {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        users = (await res.json()).filter((u) => u.is_active);
    }
    catch (err) {
        console.warn('Пользователи не загружены');
        users = [];
    }
}
async function loadReports() {
    const container = document.getElementById('timesheetContainer');
    try {
        const dateFrom = document.getElementById('dateFrom')?.value || '';
        const dateTo = document.getElementById('dateTo')?.value || '';
        if (!dateFrom || !dateTo) {
            if (container) {
                container.innerHTML = '<p class="timesheet-empty">Выберите период для отображения табеля.</p>';
            }
            return;
        }
        if (container) {
            container.innerHTML = '<p class="timesheet-empty">Загрузка табеля...</p>';
        }
        const params = new URLSearchParams({
            date_from: dateFrom,
            date_to: dateTo,
        });
        if (user) {
            params.set('current_user_id', user.id.toString());
            params.set('current_user_role', user.role);
        }
        const res = await fetch(`/api/v1/reports?${params.toString()}`, {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        if (res.status === 401)
            return logout();
        if (!res.ok) {
            const error = await res.json().catch(() => ({ error: 'Не удалось загрузить данные для табеля' }));
            throw new Error(error.error || 'Не удалось загрузить данные для табеля');
        }
        reports = await res.json();
        if (!Array.isArray(reports)) {
            reports = [];
        }
        renderTimesheet();
    }
    catch (err) {
        console.error('Ошибка загрузки отчётов:', err);
        showError(err instanceof Error ? err.message : 'Не удалось загрузить данные для табеля');
    }
}
// === Рендеринг интерфейса ===
function renderPage() {
    const today = new Date().toISOString().split('T')[0];
    const weekAgo = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    document.getElementById('content').innerHTML = `
        <link rel="stylesheet" href="/static/css/app.css">
        <link rel="stylesheet" href="/static/css/timesheet.css?v=20260710-1">
        <div class="timesheet-container">
            <h2>Табель переработок</h2>
            <div class="filters">
                <div class="form-group">
                    <label>Период с</label>
                    <input type="date" id="dateFrom" value="${weekAgo}" />
                </div>
                <div class="form-group">
                    <label>по</label>
                    <input type="date" id="dateTo" value="${today}" />
                </div>
                <button id="loadTimesheetBtn" class="btn">Загрузить табель</button>
            </div>
            <div id="timesheetContainer"></div>
        </div>
    `;
    setupEventListeners();
}
function setupEventListeners() {
    document.getElementById('loadTimesheetBtn')?.addEventListener('click', () => exportTimesheet());
    document.getElementById('dateFrom')?.addEventListener('change', () => {
        const from = document.getElementById('dateFrom').value;
        const to = document.getElementById('dateTo').value;
        if (from && to && new Date(from) > new Date(to)) {
            document.getElementById('dateTo').value = from;
        }
        loadReports();
    });
    document.getElementById('dateTo')?.addEventListener('change', () => {
        const from = document.getElementById('dateFrom').value;
        const to = document.getElementById('dateTo').value;
        if (from && to && new Date(to) < new Date(from)) {
            document.getElementById('dateFrom').value = to;
        }
        loadReports();
    });
}
function exportTimesheet() {
    const button = document.getElementById('loadTimesheetBtn');
    const dateFrom = document.getElementById('dateFrom')?.value || '';
    const dateTo = document.getElementById('dateTo')?.value || '';
    if (!dateFrom || !dateTo) {
        showError('Выберите период для экспорта табеля');
        return;
    }
    if (new Date(dateFrom) > new Date(dateTo)) {
        showError('Дата начала не может быть позже даты окончания');
        return;
    }
    const params = new URLSearchParams({
        date_from: dateFrom,
        date_to: dateTo,
    });
    window.downloadExcelExport?.(`/api/v1/export/timesheet?${params.toString()}`, 'timesheet.xlsx', button);
}
function renderTimesheet() {
    const container = document.getElementById('timesheetContainer');
    if (!container)
        return;
    const dateFrom = document.getElementById('dateFrom')?.value;
    const dateTo = document.getElementById('dateTo')?.value;
    if (!dateFrom || !dateTo) {
        container.innerHTML = '<p>Выберите период для отображения табеля.</p>';
        return;
    }
    const dates = getDateRange(dateFrom, dateTo);
    if (dates.length === 0) {
        container.innerHTML = '<p class="timesheet-empty">Выберите корректный период для отображения табеля.</p>';
        return;
    }
    if (users.length === 0) {
        container.innerHTML = '<p class="timesheet-empty">Нет активных пользователей для отображения табеля.</p>';
        return;
    }
    const dateMap = {};
    // Инициализация
    dates.forEach(date => {
        dateMap[date] = {};
    });
    // Заполнение данными
    reports.forEach(r => {
        const date = r.created_at.split('T')[0];
        if (dateMap[date] !== undefined && r.overwork > 0) {
            dateMap[date][r.user_id] = (dateMap[date][r.user_id] || 0) + r.overwork;
        }
    });
    // Генерация таблицы
    let tableHtml = `
        <table class="timesheet-table">
            <thead>
                <tr>
                    <th>Сотрудник</th>
                    ${dates.map(date => `<th>${new Date(date).toLocaleDateString()}</th>`).join('')}
                </tr>
            </thead>
            <tbody>
    `;
    users.forEach(user => {
        tableHtml += `<tr><td>${escapeHtml(user.name)}</td>`;
        dates.forEach(date => {
            const overwork = dateMap[date][user.id] || 0;
            tableHtml += `<td class="${overwork > 0 ? 'overwork-cell' : ''}">${overwork || ''}</td>`;
        });
        tableHtml += `</tr>`;
    });
    tableHtml += `</tbody></table>`;
    const totalOverwork = reports.reduce((sum, report) => sum + Math.max(0, Number(report.overwork) || 0), 0);
    if (reports.length === 0 || totalOverwork === 0) {
        tableHtml = `
            <p class="timesheet-empty">За выбранный период переработок не найдено.</p>
            ${tableHtml}
        `;
    }
    container.innerHTML = tableHtml;
}
function showError(msg) {
    const container = document.getElementById('timesheetContainer');
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
    await loadUsers();
    renderPage();
    await loadReports();
});
// === Выход ===
document.getElementById('logoutBtn')?.addEventListener('click', logout);
export {};
