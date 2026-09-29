// reports.js — интерфейс создания и просмотра отчётов (обновлённая версия с редактированием)

export {};
declare global {
    interface Window {
        downloadExcelExport?: (endpoint: string, fallbackFilename: string, trigger?: HTMLButtonElement | null) => Promise<void>;
    }
}

interface Report {
    id: number;
    user_id: number;
    project_id: number | null;
    text: string;
    overwork: number;
    created_at: string;
}

interface Project {
    id: number;
    name: string;
    responsible_id: number;
}

interface User {
    id: number;
    name: string;
    email: string;
    role: string;
}

let reports: Report[] = [];
let projects: Project[] = [];
let user: User | null = null;
let token: string | null = '';

const content = document.getElementById('content');
if (!content) {
    throw new Error('Не найден элемент #content');
}

// === Вспомогательные функции ===
function logout(): void {
    localStorage.removeItem('token');
    window.location.href = '/login';
}

function escapeHtml(text: string): string {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// === Загрузка данных ===
async function loadCurrentUser(): Promise<void> {
    try {
        const res = await fetch('/api/v1/users', {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        const allUsers: User[] = await res.json();

        const email = decodeEmailFromToken(token);
        if (email) {
            user = allUsers.find(u => u.email === email) || allUsers[0];
        } else {
            user = allUsers[0];
        }
    } catch (err) {
        console.error('Не удалось определить текущего пользователя');
        user = null;
    }
}

function decodeEmailFromToken(token: string | null): string | null {
    if (!token) return null;
    try {
        const payload = token.split('.')[1];
        if (!payload) return null;
        const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
        const parsed = JSON.parse(decoded);
        return parsed.email || null;
    } catch (e) {
        console.warn('Не удалось декодировать JWT', e);
        return null;
    }
}

async function loadProjects(): Promise<void> {
    try {
        const res = await fetch('/api/v1/projects', {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        projects = await res.json();
    } catch (err) {
        console.warn('Проекты не загружены');
        projects = [];
    }
}

async function loadReports(): Promise<void> {
    try {
        const res = await fetch(`/api/v1/reports/by-user?user_id=${user?.id}`, {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        reports = await res.json();
        renderReports();
    } catch (err) {
        console.error('Ошибка загрузки отчётов:', err);
        showError('Не удалось загрузить отчёты');
    }
}

function renderPage(): void {
    content!.innerHTML = `
        <link rel="stylesheet" href="/static/css/app.css">
        <link rel="stylesheet" href="/static/css/reports.css?v=20260710-1">
        <div class="report-controls">
            <div class="report-header">
                <h2>Создать отчёт</h2>
                <button id="exportReportsBtn" class="btn btn-excel">Скачать Excel</button>
            </div>
            <form id="createReportForm">
                <div class="form-group">
                    <label>Проект</label>
                    <select id="projectSelect">
                        <option value="">Не привязан к проекту</option>
                        ${projects.map(p => `<option value="${p.id}">${escapeHtml(p.name)}</option>`).join('')}
                    </select>
                </div>
                <div class="form-group">
                    <label>Текст отчёта *</label>
                    <textarea id="reportText" rows="6" placeholder="Опишите проделанную работу..." required></textarea>
                </div>
                <div class="form-group">
                    <label>Переработки (часы)</label>
                    <input type="number" id="overworkInput" value="0" min="0" step="1" style="width: 100px;" />
                </div>
                <button type="submit" class="btn btn-primary">Отправить отчёт</button>
            </form>
        </div>
        <div id="reportsContainer" class="reports-list">
            <p>Загрузка отчётов...</p>
        </div>
    `;

    setupEventListeners();
}

// === Функция отправки отчёта с блокировкой ===
let isSubmitting = false; // Флаг, чтобы предотвратить повторные отправки

async function handleCreateReport(e: Event): Promise<void> {
    e.preventDefault();

    // Проверка флага
    if (isSubmitting) {
        console.log('Форма уже отправляется, ждите...');
        return;
    }

    const projectSelect = document.getElementById('projectSelect') as HTMLSelectElement;
    const reportText = document.getElementById('reportText') as HTMLTextAreaElement;
    const overworkInput = document.getElementById('overworkInput') as HTMLInputElement;
    const submitBtn = e.currentTarget as HTMLButtonElement;

    const project_id = projectSelect.value ? Number(projectSelect.value) : null;
    const text = reportText.value.trim();
    const overwork = overworkInput.value ? parseInt(overworkInput.value, 10) : 0;

    if (!text) {
        alert('Текст отчёта не может быть пустым');
        return;
    }

    if (isNaN(overwork) || overwork < 0) {
        alert('Количество переработок должно быть неотрицательным числом');
        return;
    }

    // Устанавливаем флаг и блокируем кнопку
    isSubmitting = true;
    submitBtn.disabled = true;
    submitBtn.textContent = 'Отправка...';

    const reportData = {
        user_id: user?.id,
        project_id,
        text,
        overwork
    };

    try {
        const res = await fetch('/api/v1/reports', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(reportData),
        });

        if (res.ok) {
            alert('Отчёт успешно отправлен');
            reportText.value = '';
            overworkInput.value = '0';
            await loadReports(); // Перезагружаем список отчётов
        } else {
            const err = await res.json().catch(() => ({ error: 'Неизвестная ошибка' }));
            alert(`Ошибка: ${err.error}`);
        }
    } catch (err) {
        alert('Ошибка сети');
    } finally {
        // Сбрасываем флаг и разблокируем кнопку
        isSubmitting = false;
        submitBtn.disabled = false;
        submitBtn.textContent = 'Отправить отчёт';
    }
}

// === Вспомогательная функция для безопасного экранирования HTML, кроме ссылок ===
function escapeHtmlExceptLinks(text: string): string {
    if (!text) return '';

    // Сначала экранируем весь текст
    let escaped = text
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');

    // Регулярное выражение для поиска URL (http/https/www)
    const urlRegex = /(https?:\/\/[^\s<>"{}|\\^`\[\]]+|www\.[^\s<>"{}|\\^`\[\]]+)/g;

    // Функция для создания ссылки
    const makeLink = (match: string): string => {
        let url = match;
        // Если URL начинается с www, добавляем http://
        if (match.startsWith('www.')) {
            url = 'http://' + match;
        }

        return `<a href="${url}" target="_blank" rel="noopener noreferrer" class="report-link">${match}</a>`;
    };

    // Заменяем URL на ссылки
    return escaped.replace(urlRegex, makeLink);
}

// === Обновлённая функция рендера отчётов с поддержкой ссылок ===
function renderReports(): void {
    const container = document.getElementById('reportsContainer');
    if (!container) return;

    if (reports.length === 0) {
        container.innerHTML = '<p>У вас пока нет отчётов.</p>';
        return;
    }

    // Сортировка по дате (новые — сначала)
    const sortedReports = [...reports].sort((a, b) =>
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    );

    container.innerHTML = `
        <h3>Ваши отчёты</h3>
        ${sortedReports.map(r => {
            const project = projects.find(p => p.id === r.project_id);
            const date = new Date(r.created_at).toLocaleString();
            // Используем новую функцию для обработки текста с ссылками
            const formattedText = escapeHtmlExceptLinks(r.text);
            return `
                <div class="report-item" data-id="${r.id}">
                    <div class="report-field">
                        <span class="report-label">Дата:</span>
                        <span class="report-value">${date}</span>
                    </div>
                    <div class="report-field">
                        <span class="report-label">Проект:</span>
                        <span class="report-value">${project ? escapeHtml(project.name) : '—'}</span>
                    </div>
                    <div class="report-field">
                        <span class="report-label">Текст:</span>
                        <div class="report-value report-text">${formattedText}</div>
                    </div>
                    <div class="report-field">
                        <span class="report-label">Переработки:</span>
                        <span class="report-value"><span class="report-overwork">${r.overwork || 0}</span> час(ов)</span>
                    </div>
                    <div class="report-actions">
                        <button class="btn btn-neutral btn-sm edit-report-btn" data-id="${r.id}">Редактировать</button>
                        <button class="btn btn-danger btn-sm delete-report-btn" data-id="${r.id}">Удалить</button>
                    </div>
                    <!-- Форма редактирования — изначально скрыта -->
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
                        <div>
                            <button class="btn btn-primary btn-sm save-report-btn" data-id="${r.id}">Сохранить</button>
                            <button class="btn btn-neutral btn-sm cancel-edit-btn" data-id="${r.id}" style="margin-left: 8px;">Отмена</button>
                        </div>
                    </div>
                </div>
            `;
        }).join('')}
    `;

    // Назначаем обработчики для кнопок редактирования
    attachEditEventListeners();
    attachDeleteEventListeners();
}

function attachEditEventListeners(): void {
    // Кнопка "Редактировать"
    document.querySelectorAll('.edit-report-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const id = (e.target as HTMLElement).getAttribute('data-id');
            openEditForm(Number(id));
        });
    });

    // Кнопка "Сохранить"
    document.querySelectorAll('.save-report-btn').forEach(btn => {
        btn.addEventListener('click', async (e) => {
            const id = (e.target as HTMLElement).getAttribute('data-id');
            await saveReportEdit(Number(id));
        });
    });

    // Кнопка "Отмена"
    document.querySelectorAll('.cancel-edit-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const id = (e.target as HTMLElement).getAttribute('data-id');
            cancelEdit(Number(id));
        });
    });
}

function openEditForm(id: number): void {
    // Скрываем все формы и показываем только нужную
    document.querySelectorAll('.edit-report-form').forEach(form => {
        const element = form as HTMLElement;
        element.style.display = 'none';
    });

    const form = document.getElementById(`edit-form-${id}`);
    if (form) {
        form.style.display = 'block';
    }

    // Скрываем текст отчёта
    const reportItem = document.querySelector(`.report-item[data-id="${id}"]`);
    if (reportItem) {
        const textSpan = reportItem.querySelector('.report-text');
        const overworkSpan = reportItem.querySelector('.report-overwork');
        if (textSpan) textSpan.innerHTML = '<em>Редактируется...</em>';
        if (overworkSpan) overworkSpan.textContent = '...';
    }
}

async function saveReportEdit(id: number): Promise<void> {
    const form = document.getElementById(`edit-form-${id}`) as HTMLElement;
    const projectSelect = form.querySelector('.edit-project-select') as HTMLSelectElement;
    const textArea = form.querySelector('.edit-report-text') as HTMLTextAreaElement;
    const overworkInput = form.querySelector('.edit-overwork-input') as HTMLInputElement;

    const project_id = projectSelect.value ? Number(projectSelect.value) : null;
    const text = textArea.value.trim();
    const overwork = overworkInput.value ? parseInt(overworkInput.value, 10) : 0;

    if (!text) {
        alert('Текст отчёта не может быть пустым');
        return;
    }

    if (isNaN(overwork) || overwork < 0) {
        alert('Переработки должны быть неотрицательным числом');
        return;
    }

    const updatedReport = {
        project_id,
        text,
        overwork
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
            // Обновляем локальный массив
            const report = reports.find(r => r.id === id);
            if (report) {
                report.project_id = project_id;
                report.text = text;
                report.overwork = overwork;
            }
            // Перерисовываем
            renderReports();
            alert('Отчёт обновлён');
        } else {
            const err = await res.json().catch(() => ({ error: 'Ошибка при обновлении' }));
            alert(`Ошибка: ${err.error}`);
        }
    } catch (err) {
        alert('Ошибка сети при сохранении');
    }
}

function cancelEdit(id: number): void {
    const form = document.getElementById(`edit-form-${id}`);
    if (form) {
        form.style.display = 'none';
    }
    // Перерисовываем, чтобы сбросить состояние
    renderReports();
}

function attachDeleteEventListeners(): void {
    document.querySelectorAll('.delete-report-btn').forEach(btn => {
        btn.addEventListener('click', async (e) => {
            const id = Number((e.target as HTMLElement).getAttribute('data-id'));
            if (!id || !confirm('Удалить этот отчёт?')) return;

            try {
                const res = await fetch(`/api/v1/reports/${id}`, {
                    method: 'DELETE',
                    headers: { 'Authorization': `Bearer ${token}` },
                });

                if (res.ok) {
                    reports = reports.filter(report => report.id !== id);
                    renderReports();
                    return;
                }

                const err = await res.json().catch(() => ({ error: 'Не удалось удалить отчёт' }));
                alert(`Ошибка: ${err.error}`);
            } catch (err) {
                alert('Ошибка сети при удалении');
            }
        });
    });
}

function setupEventListeners(): void {
    document.getElementById('createReportForm')?.addEventListener('submit', handleCreateReport);

    const exportReportsBtn = document.getElementById('exportReportsBtn');
    if (exportReportsBtn instanceof HTMLButtonElement) {
        exportReportsBtn.addEventListener('click', () => {
            window.downloadExcelExport?.('/api/v1/export/reports', 'reports.xlsx', exportReportsBtn);
        });
    }
}

function showError(msg: string): void {
    const container = document.getElementById('reportsContainer');
    if (container) {
        container.innerHTML = `<div class="error">${escapeHtml(msg)}</div>`;
    }
}

// === Инициализация ===
document.addEventListener('DOMContentLoaded', async () => {
    token = localStorage.getItem('token');
    if (!token) return logout();

    // Сначала загружаем текущего пользователя
    await loadCurrentUser();
    await loadProjects();

    renderPage();
    // Только после этого загружаем отчёты
    await loadReports();
    setupEventListeners();
});

// === Выход ===
document.getElementById('logoutBtn')?.addEventListener('click', logout);
