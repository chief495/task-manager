// admin-users.ts — интерфейс управления пользователями
export {};
declare const Chart: any;
interface User {
    id: number;
    name: string;
    email: string;
    role: string;
    is_active: boolean;
    created_at: string;
    updated_at: string;
    social_rating: number;
    middle_plan_time: number;
}

interface UserStats {
    user_id: number;
    total: number;
    in_progress: number;
    done: number;
}

// === Глобальные переменные ===
let users: User[] = [];
let stats: Record<number, UserStats> = {};
let token: string | null = '';
let currentChart: any = null; // Хранит текущий график

// === Интерфейс для данных графика ===
interface SocialRatingPoint {
    date: string;
    rating: number;
}

const content = document.getElementById('content');
if (!content) {
    throw new Error('Не найден элемент #content');
}

const AVAILABLE_ROLES = ['user', 'manager', 'admin'];
const STATUS_COLORS = {
    active: '#4CAF50',
    inactive: '#F44336'
};

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

function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString();
}

function getUserRoleLabel(role: string): string {
    const labels: Record<string, string> = {
        'admin': 'Администратор',
        'manager': 'Менеджер',
        'user': 'Пользователь'
    };
    return labels[role] || role;
}

// === Генерация интерфейса ===
function renderPage(): void {
    content!.innerHTML = `
        <link rel="stylesheet" href="/static/css/app.css">
        <link rel="stylesheet" href="/static/css/admin-users.css?v=20260602-2">
        
        <h2>Управление пользователями</h2>
        <div class="users-toolbar">
            <button id="createUserBtn" class="btn btn-primary">➕ Добавить пользователя</button>
            <button id="sortRating" class="btn btn-primary">Сортировать по рейтингу</button>
            <button id="sortName" class="btn btn-primary">Сортировать по имени</button>
        </div>
        <ul id="userList" class="list"></ul>

        <!-- Модальное окно -->
        <div id="userModal" class="modal" style="display: none;">
            <div class="modal-content">
                <button type="button" class="modal-close-x" aria-label="Закрыть" onclick="closeModal()">×</button>
                <h3 id="modalTitle">Добавить пользователя</h3>
                <form id="userForm">
                    <input type="hidden" id="userId" />
                    
                    <div class="form-group">
                        <label>Имя</label>
                        <input type="text" id="userName" />
                    </div>
                    
                    <div class="form-group">
                        <label>Email *</label>
                        <input type="email" id="userEmail" required />
                    </div>
                    
                    <div class="form-group">
                        <label>Пароль</label>
                        <button type="button" class="reset-password-btn">Сбросить пароль на 12345678</button>
                        ${isEditMode() ? '<small>Текущий пароль будет изменён только при нажатии кнопки "Сбросить"</small>' : '<small>Пароль по умолчанию: 12345678</small>'}
                    </div>
                    
                    <div class="form-group">
                        <label>Роль *</label>
                        <select id="userRole" required>
                            ${AVAILABLE_ROLES.map(role => `
                                <option value="${role}">${getUserRoleLabel(role)}</option>
                            `).join('')}
                        </select>
                    </div>
                    
                    <div class="form-group">
                        <label>Среднее плановое время (часы)</label>
                        <input type="number" id="middle_plan_time" step="0.1" min="0" />
                    </div>
                    
                    <div class="form-group">
                        <label>Социальный рейтинг</label>
                        <input type="number" id="social_rating" step="0.01" min="0" max="1" />
                    </div>
                    
                    <div class="form-actions">
                        <button type="submit">Сохранить</button>
                    </div>
                </form>
            </div>
        </div>
    `;

    setupEventListeners();
}

function isEditMode(): boolean {
    const userIdInput = document.getElementById('userId') as HTMLInputElement | null;
    return userIdInput?.value !== '';
}

// === Загрузка данных ===
async function loadUsers(): Promise<void> {
    try {
        const res = await fetch('/api/v1/users', {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        if (res.status === 401) return logout();
        users = await res.json();
        await loadAllStats();
        renderUserList();
    } catch (err) {
        console.error('Ошибка загрузки пользователей:', err);
        showError('Не удалось загрузить пользователей');
    }
}

async function loadAllStats(): Promise<void> {
    const promises = users.map(user =>
        fetch(`/api/v1/users/${user.id}/tasks-stats`, {
            headers: { 'Authorization': `Bearer ${token}` },
        })
            .then(res => res.json())
            .then(data => {
                stats[user.id] = data;
            })
            .catch(err => {
                console.warn(`Не удалось загрузить статистику для пользователя ${user.id}:`, err);
            })
    );
    await Promise.all(promises);
}

// === Рендеринг ===
function renderUserList(): void {
    const list = document.getElementById('userList');
    if (!list) return;

    list.innerHTML = users.map(user => {
        const userStats = stats[user.id] || { total: 0, in_progress: 0, done: 0 };
        const statusColor = user.is_active ? STATUS_COLORS.active : STATUS_COLORS.inactive;
        const statusText = user.is_active ? 'Активен' : 'Неактивен';
        const middlePlanHours = (user.middle_plan_time);

        return `
            <li class="user-item" data-id="${user.id}">
                <div class="user-info">
                    <div class="user-name">${escapeHtml(user.name || 'Без имени')}</div>
                    <div class="user-email">${escapeHtml(user.email)}</div>
                    <div class="user-role">${getUserRoleLabel(user.role)}</div>
                    <div class="user-status">
                        <span style="color: ${statusColor}; font-size: 12px;">●</span> ${statusText}
                    </div>
                    <div class="user-rating">
                        <strong>Рейтинг:</strong> ${(user.social_rating || 0).toFixed(2)}
                    </div>
                    <div class="user-plan-time">
                        <strong>Среднее время:</strong> ${middlePlanHours} ч
                    </div>
                    <div class="user-stats">
                        <small>Задачи: всего ${userStats.total}, в работе ${userStats.in_progress}, готово ${userStats.done}</small>
                    </div>
                </div>
                <div class="user-actions">
                    <button class="btn btn-neutral btn-sm edit-user" data-id="${user.id}">Редактировать</button>
                    <button class="btn btn-primary btn-sm show-rating-graph" data-id="${user.id}">График рейтинга</button>
                    ${user.is_active
            ? `<button class="btn btn-danger btn-sm deactivate" data-id="${user.id}">Деактивировать</button>`
            : `<button class="btn btn-success btn-sm activate" data-id="${user.id}">Активировать</button>`
        }
                </div>
            </li>
        `;
    }).join('');

    attachActionListeners();
}
// === Функция для загрузки истории рейтинга ===
async function loadSocialRatingHistory(userId: number, days: number = 14): Promise<SocialRatingPoint[]> {
    try {
        const res = await fetch(`/api/v1/users/${userId}/social-rating-history?days=${days}`, {
            headers: { 'Authorization': `Bearer ${token}` },
        });

        if (!res.ok) {
            throw new Error('Не удалось загрузить историю рейтинга');
        }

        return await res.json();
    } catch (err) {
        console.error('Ошибка загрузки истории рейтинга:', err);
        throw err;
    }
}
// === Отображение графика ===
async function showRatingGraph(userId: number): Promise<void> {
    const user = users.find(u => u.id === userId);
    if (!user) return;

    const history = await loadSocialRatingHistory(userId, 14);

    // Создаём модальное окно
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.style.display = 'flex';
    modal.innerHTML = `
        <div class="modal-content" style="width: 80%; max-width: 800px; height: 500px; display: flex; flex-direction: column;">
            <button type="button" class="modal-close-x close-chart" aria-label="Закрыть">×</button>
            <h3>График социального рейтинга — ${escapeHtml(user.name)}</h3>
            <div style="flex: 1; position: relative;">
                <canvas id="ratingChart"></canvas>
            </div>
            <div class="form-actions" style="justify-content: space-between; margin-top: 10px;">
                <div>
                    <button class="btn btn-neutral btn-sm" data-days="7">7 дней</button>
                    <button class="btn btn-neutral btn-sm" data-days="14">14 дней</button>
                    <button class="btn btn-neutral btn-sm" data-days="30">30 дней</button>
                    <button class="btn btn-neutral btn-sm" data-days="90">90 дней</button>
                </div>
            </div>
        </div>
    `;
    document.body.appendChild(modal);

    // Рисуем график
    const ctx = (document.getElementById('ratingChart') as HTMLCanvasElement).getContext('2d');
    if (!ctx) return;

    // Уничтожаем предыдущий график
    if (currentChart) {
        currentChart.destroy();
    }

    currentChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: history.map(p => new Date(p.date).toLocaleDateString()),
            datasets: [{
                label: 'Социальный рейтинг',
                data: history.map(p => p.rating),
                borderColor: '#1a73e8',
                backgroundColor: 'rgba(26, 115, 232, 0.1)',
                fill: true,
                tension: 0.4,
                pointRadius: 3,
                pointHoverRadius: 6
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    min: 0.6,
                    max: 1.3,
                    ticks: {
                        stepSize: 0.1
                    }
                }
            },
            plugins: {
                legend: {
                    display: true,
                    position: 'top'
                }
            }
        }
    });

    // Обработчики кнопок
    modal.querySelectorAll('.btn[data-days]').forEach(btn => {
        btn.addEventListener('click', async () => {
            const days = Number(btn.getAttribute('data-days'));
            const newHistory = await loadSocialRatingHistory(userId, days);
            currentChart.data.labels = newHistory.map(p => new Date(p.date).toLocaleDateString());
            currentChart.data.datasets[0].data = newHistory.map(p => p.rating);
            currentChart.update();
        });
    });

    modal.querySelector('.close-chart')?.addEventListener('click', () => {
        modal.remove();
        if (currentChart) {
            currentChart.destroy();
            currentChart = null;
        }
    });
}
function attachActionListeners(): void {
    // Редактирование пользователя
    document.querySelectorAll('.edit-user').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const id = Number((e.target as HTMLElement).getAttribute('data-id'));
            editUser(id);
        });
    });

    // Деактивация
    document.querySelectorAll('.deactivate').forEach(btn => {
        btn.addEventListener('click', async (e) => {
            const id = Number((e.target as HTMLElement).getAttribute('data-id'));
            if (confirm('Деактивировать пользователя?')) {
                await toggleUserStatus(id, false);
            }
        });
    });

    // Активация
    document.querySelectorAll('.activate').forEach(btn => {
        btn.addEventListener('click', async (e) => {
            const id = Number((e.target as HTMLElement).getAttribute('data-id'));
            if (confirm('Активировать пользователя?')) {
                await toggleUserStatus(id, true);
            }
        });
    });
    // График рейтинга
    document.querySelectorAll('.show-rating-graph').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const button = (e.target as HTMLElement).closest('.show-rating-graph') as HTMLElement | null;
            const id = Number(button?.getAttribute('data-id'));
            showRatingGraph(id);
        });
    });
}

function showError(msg: string): void {
    content!.innerHTML += `<p class="error">${escapeHtml(msg)}</p>`;
}

// === Модальное окно ===
function openModal(id: number | null = null): void {
    const modal = document.getElementById('userModal');
    const title = document.getElementById('modalTitle');
    const form = document.getElementById('userForm') as HTMLFormElement;

    if (id !== null) {
        title!.textContent = 'Редактировать пользователя';

        loadUserData(id);
    } else {
        title!.textContent = 'Добавить пользователя';
        form.reset();
        const userIdInput = document.getElementById('userId') as HTMLInputElement;
        userIdInput.value = '';
    }

    if (modal) modal.style.display = 'flex';
}

async function loadUserData(id: number): Promise<void> {
    try {
        const res = await fetch(`/api/v1/users/${id}`, {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        if (res.status === 404) {
            alert('Пользователь не найден');
            return;
        }
        const user: User = await res.json();

        const userIdInput = document.getElementById('userId') as HTMLInputElement;
        const userNameInput = document.getElementById('userName') as HTMLInputElement;
        const userEmailInput = document.getElementById('userEmail') as HTMLInputElement;
        const userRoleSelect = document.getElementById('userRole') as HTMLSelectElement;
        const middle_plan_timeInput = document.getElementById('middle_plan_time') as HTMLInputElement;
        const social_ratingInput = document.getElementById('social_rating') as HTMLInputElement;
        userIdInput.value = user.id.toString();
        userNameInput.value = user.name || '';
        userEmailInput.value = user.email;
        userRoleSelect.value = user.role;
        middle_plan_timeInput.value = (user.middle_plan_time).toFixed(1);
        social_ratingInput.value = (user.social_rating).toFixed(2);
    } catch (err) {
        console.error('Ошибка загрузки пользователя:', err);
        alert('Не удалось загрузить данные пользователя');
    }
}
// === Сортировка ===
function sortByRating(): void {
    users.sort((a, b) => (b.social_rating || 0) - (a.social_rating || 0));
    renderUserList();
}

function sortByName(): void {
    users.sort((a, b) => a.name.localeCompare(b.name));
    renderUserList();
}


function closeModal(): void {
    const modal = document.getElementById('userModal');
    if (modal) modal.style.display = 'none';
}

async function handleSubmit(e: Event): Promise<void> {
    e.preventDefault();

    const userIdInput = document.getElementById('userId') as HTMLInputElement;
    const userNameInput = document.getElementById('userName') as HTMLInputElement;
    const userEmailInput = document.getElementById('userEmail') as HTMLInputElement;
    const userRoleSelect = document.getElementById('userRole') as HTMLSelectElement;
    const userMiddlePlanTime = document.getElementById('middle_plan_time') as HTMLInputElement;
    const userSocialRating = document.getElementById('social_rating') as HTMLInputElement;
    const id = userIdInput?.value ? Number(userIdInput.value) : null;
    const name = userNameInput?.value || '';
    const email = userEmailInput?.value;
    const role = userRoleSelect?.value;
    const middle_plan_time = userMiddlePlanTime?.value;
    const social_rating = userSocialRating?.value;
    if (!email || !role) {
        alert('Заполните обязательные поля');
        return;
    }

    const userData: any = { email, role };
    if (name) userData.name = name;
    if (id) userData.id = id;
    if (role) userData.role = role;
    if (middle_plan_time) userData.middle_plan_time = middle_plan_time;
    if (social_rating) userData.social_rating = social_rating;
    // Добавляем пароль, только если флаг сброса установлен
    if (resetPasswordFlag) {
        userData.password = '12345678';
        resetPasswordFlag = false; // Сбрасываем флаг
        const hint = document.getElementById('passwordHint');
        if (hint) {
            hint.textContent = 'Пароль был сброшен на 12345678';
            hint.style.color = '#34a853';
            setTimeout(() => {
                if (hint) hint.style.display = 'none';
            }, 3000);
        }
    }
    try {
        const res = await fetch(id ? `/api/v1/users/${id}`: '/api/v1/users', {
            method: id ? 'PUT' : 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`,
            },
            body: JSON.stringify(userData),
        });

        if (res.ok) {
            closeModal();
            await loadUsers();
        } else {
            const error = await res.json();
            alert(`Ошибка: ${error.error || 'Не удалось сохранить'}`);
        }
    } catch (err) {
        alert('Ошибка сети');
    }
}

// === Управление статусом ===
async function toggleUserStatus(id: number, isActive: boolean): Promise<void> {
    try {
        const endpoint = isActive ? 'activate' : 'deactivate';
        const res = await fetch(`/api/v1/users/${id}/${endpoint}`, {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${token}` },
        });

        if (res.ok) {
            const msg = await res.json();
            alert(msg.message);
            await loadUsers();
        } else {
            alert(`Не удалось ${isActive ? 'активировать' : 'деактивировать'} пользователя`);
        }
    } catch (err) {
        alert('Ошибка сети');
    }
}

function editUser(id: number): void {
    openModal(id);
}

function createUser(): void {
    openModal();
}

// === Делаем функции глобальными ===
(window as any).closeModal = closeModal;
(window as any).editUser = editUser;
(window as any).createUser = createUser;

// === Инициализация ===
document.addEventListener('DOMContentLoaded', async () => {
    token = localStorage.getItem('token');
    if (!token) {
        window.location.href = '/login';
        return;
    }

    renderPage();
    await loadUsers();
});

// === Выход ===
document.getElementById('logoutBtn')?.addEventListener('click', logout);
// Глобальный флаг для сброса пароля
let resetPasswordFlag = false;

// === Добавление обработчиков событий ===
function setupEventListeners(): void {
    document.getElementById('createUserBtn')?.addEventListener('click', () => createUser());
    document.getElementById('userForm')?.addEventListener('submit', handleSubmit);
    document.getElementById('sortRating')?.addEventListener('click', sortByRating);
    document.getElementById('sortName')?.addEventListener('click', sortByName);
    // Обработчик кнопки сброса пароля
    document.querySelector('.reset-password-btn')?.addEventListener('click', () => {
        if (confirm('Установить пароль на 12345678?')) {
            resetPasswordFlag = true;
            const hint = document.getElementById('passwordHint');
            if (hint) {
                hint.textContent = 'Пароль будет сброшен на 12345678 при сохранении';
                hint.style.color = '#4bb4e6';
            }
        }
    });
}
