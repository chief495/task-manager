// admin-statuses.ts — динамический интерфейс управления статусами

export {};

if (window.location.pathname !== '/admin/statuses') {
    throw new Error('admin-statuses.ts loaded outside /admin/statuses');
}

interface Status {
    id: number;
    name: string;
    description: string;
    is_default: boolean;
    is_final: boolean;
}

interface StatusTransition {
    id: number;
    from_status_id: number;
    to_status_id: number;
    condition: string | null;
    allowed_roles: string[];
    from_status?: Status;
    to_status?: Status;
}

let statuses: Status[] = [];
let transitions: StatusTransition[] = [];
let token: string | null = '';
let selectedStatusId: number | null = null;

const AVAILABLE_ROLES = ['admin', 'manager', 'user'];

// === DOM Elements ===
const content = document.getElementById('content');
if (!content) {
    throw new Error('Не найден элемент #content');
}

// === Вспомогательные функции ===
function logout() {
    localStorage.removeItem('token');
    window.location.href = '/login';
}

function escapeHtml(text: string): string {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// === Генерация интерфейса ===
function renderPage(): void {
    content!.innerHTML = `
        <link rel="stylesheet" href="/static/css/app.css">
        <link rel="stylesheet" href="/static/css/admin-statuses.css?v=5">

        <div class="admin-statuses-container">
            <button id="createStatusBtn" class="btn btn-primary">➕ Добавить статус</button>
            <ul id="statusList" class="list"></ul>

            <div id="transitionsSection" style="display: none; margin-top: 20px;">
                <h3>Переходы из <span id="currentStatusName"></span></h3>

                <div class="form-group">
                    <label>Добавить разрешённый переход</label>
                    <div class="transition-controls-row">
                        <select id="toStatusSelect"></select>
                        <button id="addTransitionBtn" class="btn btn-primary btn-sm">Добавить</button>
                    </div>
                    <div id="rolesContainer"></div>
                </div>

                <div class="transitions-block">
                    <h4>Исходящие переходы</h4>
                    <div id="outgoingTransitions" class="transition-list" style="list-style: none; margin-left: 0; padding-left: 0;"></div>
                </div>

                <div class="transitions-block">
                    <h4>Входящие переходы</h4>
                    <div id="incomingTransitions" class="transition-list" style="list-style: none; margin-left: 0; padding-left: 0;"></div>
                </div>
            </div>
        </div>
    `;

    setupEventListeners();
}

// === Инициализация ===
document.addEventListener('DOMContentLoaded', async () => {
    token = localStorage.getItem('token');
    if (!token) {
        window.location.href = '/login';
        return;
    }

    renderPage();
    await loadStatuses();
    await loadTransitions();
    renderStatusList();
});

async function loadStatuses(): Promise<void> {
    try {
        const res = await fetch('/api/v1/statuses', {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        if (res.status === 401) return logout();
        const data = await res.json();
        statuses = Array.isArray(data) ? data : [];
    } catch (err) {
        console.error('Ошибка загрузки статусов:', err);
        alert('Не удалось загрузить статусы');
    }
}

async function loadTransitions(): Promise<void> {
    try {
        const res = await fetch('/api/v1/status-transitions', {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        if (res.status === 401) return logout();
        const data = await res.json();
        transitions = Array.isArray(data) ? data : [];
    } catch (err) {
        console.error('Ошибка загрузки переходов:', err);
        alert('Не удалось загрузить переходы');
    }
}

function setupEventListeners(): void {
    // Кнопка "Добавить статус"
    document.getElementById('createStatusBtn')?.addEventListener('click', openCreateStatusModal);

    // Делегирование: "Редактировать переходы"
    document.getElementById('statusList')?.addEventListener('click', (e) => {
        const target = e.target as HTMLElement;
        if (target.classList.contains('edit-transitions')) {
            const id = Number(target.getAttribute('data-id'));
            selectStatus(id);
        }
    });

    // Кнопка "Добавить переход"
    document.getElementById('addTransitionBtn')?.addEventListener('click', () => {
        addTransition().catch(err => {
            console.error('Ошибка добавления перехода:', err);
            alert('Ошибка при добавлении перехода');
        });
    });
}

function renderStatusList(): void {
    const list = document.getElementById('statusList');
    if (!list) return;

    list.innerHTML = statuses.map(status => {
        const isDefault = status.is_default ? ' (по умолчанию)' : '';
        const isFinal = status.is_final ? ' (финальный)' : '';
        return `
            <li class="status-item" data-id="${status.id}">
                <span>${escapeHtml(status.name)}${isDefault}${isFinal}</span>
                <div class="actions">
                    <button class="btn btn-neutral btn-sm edit-transitions" data-id="${status.id}">Редактировать переходы</button>
                </div>
            </li>
        `;
    }).join('');
}

function getStatusName(id: number, fallback = 'Неизвестно'): string {
    return statuses.find(status => status.id === id)?.name || fallback;
}

function selectStatus(statusId: number): void {
    selectedStatusId = statusId;
    const status = statuses.find(s => s.id === statusId);
    if (!status) return;

    document.getElementById('currentStatusName')!.textContent = status.name;
    document.getElementById('transitionsSection')!.style.display = 'block';

    renderOutgoingTransitions(statusId);
    renderIncomingTransitions(statusId);
    renderToStatusSelect(statusId);
}

function renderOutgoingTransitions(fromId: number): void {
    const list = document.getElementById('outgoingTransitions')!;
    list.innerHTML = '';

    const outgoing = transitions.filter(t => t.from_status_id === fromId);
    if (outgoing.length === 0) {
        list.innerHTML = '<div class="transition-row empty"><em>Нет разрешённых переходов</em></div>';
        return;
    }

    outgoing.forEach(t => {
        const roles = t.allowed_roles.length ? `(${t.allowed_roles.join(', ')})` : '(все)';

        list.innerHTML += `
            <div class="transition-row">
                <div class="transition-row-content">
                    <strong>${escapeHtml(t.from_status?.name || getStatusName(t.from_status_id))}</strong>
                    <span>→ ${escapeHtml(t.to_status?.name || getStatusName(t.to_status_id))} ${roles}</span>
                </div>
                <button class="btn btn-danger btn-sm delete-transition" data-id="${t.id}">Удалить</button>
            </div>
        `;
    });

    list.querySelectorAll('.delete-transition').forEach(btn => {
        btn.addEventListener('click', async (e) => {
            const id = Number((e.currentTarget as HTMLElement).getAttribute('data-id'));
            await deleteTransition(id);
        });
    });
}

function renderIncomingTransitions(toId: number): void {
    const list = document.getElementById('incomingTransitions')!;
    list.innerHTML = '';

    const incoming = transitions.filter(t => t.to_status_id === toId);
    if (incoming.length === 0) {
        list.innerHTML = '<div class="transition-row empty"><em>Никто не может перейти в этот статус</em></div>';
        return;
    }

    incoming.forEach(t => {
        list.innerHTML += `
            <div class="transition-row">
                <div class="transition-row-content">
                    ${escapeHtml(t.from_status?.name || getStatusName(t.from_status_id))} → <strong>${escapeHtml(t.to_status?.name || getStatusName(t.to_status_id))}</strong>
                </div>
            </div>
        `;
    });
}

function renderToStatusSelect(fromId: number): void {
    const select = document.getElementById('toStatusSelect') as HTMLSelectElement;
    select.innerHTML = '<option value="">Выберите статус</option>';
    if (statuses.length <= 1) {
        select.innerHTML = '<option value="">Нет доступных статусов</option>';
        return;
    }

    statuses
        .filter(s => s.id !== fromId)
        .forEach(s => {
            select.innerHTML += `<option value="${s.id}">${escapeHtml(s.name)}</option>`;
        });

    const container = document.getElementById('rolesContainer');
    if (!container) return;

    container.innerHTML = `
        <div class="form-group">
            <label>Разрешённые роли *</label>
            <div class="roles-group">
                ${AVAILABLE_ROLES.map(role => `
                    <label class="role-checkbox">
                        <input type="checkbox" name="allowed_roles" value="${role}">
                        ${escapeHtml(role)}
                    </label>
                `).join('')}
            </div>
        </div>
    `;
}

async function addTransition(): Promise<void> {
    if (!selectedStatusId) return;

    const select = document.getElementById('toStatusSelect') as HTMLSelectElement;
    const toId = Number(select.value);
    if (toId <= 0) {
        alert('Выберите целевой статус');
        return;
    }

    const roleCheckboxes = document.querySelectorAll('input[name="allowed_roles"]:checked');
    const allowedRoles: string[] = Array.from(roleCheckboxes).map(el => (el as HTMLInputElement).value);

    if (allowedRoles.length === 0) {
        alert('Выберите хотя бы одну роль');
        return;
    }

    const transitionData = {
        from_status_id: selectedStatusId,
        to_status_id: toId,
        allowed_roles: allowedRoles,
    };

    try {
        const res = await fetch('/api/v1/status-transitions', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(transitionData),
        });

        if (res.ok) {
            const newTrans: StatusTransition = await res.json();
            transitions.push(newTrans);
            renderOutgoingTransitions(selectedStatusId);
            select.value = '';
            roleCheckboxes.forEach(el => (el as HTMLInputElement).checked = false);
        } else {
            const error = await res.json().catch(() => ({ error: 'Неизвестная ошибка' }));
            alert(`Ошибка: ${error.error || 'Не удалось добавить переход'}`);
        }
    } catch (err) {
        alert('Ошибка сети');
    }
}

async function deleteTransition(id: number): Promise<void> {
    if (!confirm('Удалить переход?')) return;

    try {
        const res = await fetch(`/api/v1/status-transitions/${id}`, {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${token}` },
        });

        if (res.ok) {
            transitions = transitions.filter(t => t.id !== id);
            if (selectedStatusId) {
                selectStatus(selectedStatusId);
            }
        } else {
            alert('Ошибка удаления');
        }
    } catch (err) {
        alert('Ошибка сети');
    }
}

function openCreateStatusModal(): void {
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <button type="button" class="modal-close-x" aria-label="Закрыть">×</button>
            <h3>Создать статус</h3>
            <form id="createStatusForm">
                <div class="form-group">
                    <label>Название *</label>
                    <input type="text" name="name" required />
                </div>
                <div class="form-group">
                    <label>Описание</label>
                    <textarea name="description"></textarea>
                </div>
                <div class="form-group">
                    <label><input type="checkbox" name="is_default"> По умолчанию</label>
                </div>
                <div class="form-group">
                    <label><input type="checkbox" name="is_final"> Финальный статус</label>
                </div>
                <div class="form-actions">
                    <button type="submit">Создать</button>
                </div>
            </form>
        </div>
    `;
    document.body.appendChild(modal);

    const form = modal.querySelector('#createStatusForm') as HTMLFormElement;
    modal.querySelector('.modal-close-x')?.addEventListener('click', () => modal.remove());
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(form);
        const data = {
            name: (formData.get('name') as string).trim(),
            description: formData.get('description') as string || null,
            is_default: !!formData.get('is_default'),
            is_final: !!formData.get('is_final'),
        };

        if (!data.name) {
            alert('Введите название');
            return;
        }

        try {
            const res = await fetch('/api/v1/statuses', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data),
            });

            if (res.ok) {
                const newStatus: Status = await res.json();
                statuses.push(newStatus);
                renderStatusList();
                modal.remove();
            } else {
                alert('Ошибка создания статуса');
            }
        } catch (err) {
            alert('Ошибка сети');
        }
    });
}

// === Выход ===
document.getElementById('logoutBtn')?.addEventListener('click', logout);
