// admin-stages.ts — динамический интерфейс управления стадиями проектов
let stages = [];
let transitions = [];
let token = '';
let selectedStageId = null;
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
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
// === Генерация интерфейса ===
function renderPage() {
    content.innerHTML = `
        <link rel="stylesheet" href="/static/css/app.css">
        <link rel="stylesheet" href="/static/css/admin-statuses.css?v=5">

        <div class="admin-statuses-container">
            <button id="createStageBtn" class="btn btn-primary">➕ Добавить стадию</button>

            <ul id="stageList" class="list"></ul>

            <div id="transitionsSection" style="display: none; margin-top: 20px;">
                <h3>Переходы из <span id="currentStageName"></span></h3>

                <div class="form-group">
                    <label>Добавить разрешённый переход</label>
                    <div class="transition-controls-row">
                        <select id="toStageSelect"></select>
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
    await loadStages();
    await loadTransitions();
    renderStageList();
});
async function loadStages() {
    try {
        const res = await fetch('/api/v1/stages', {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        if (res.status === 401)
            return logout();
        stages = await res.json();
    }
    catch (err) {
        console.error('Ошибка загрузки стадий:', err);
        alert('Не удалось загрузить стадии');
    }
}
async function loadTransitions() {
    try {
        const res = await fetch('/api/v1/stage-transitions', {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        if (res.status === 401)
            return logout();
        transitions = await res.json();
    }
    catch (err) {
        console.error('Ошибка загрузки переходов между стадиями:', err);
        alert('Не удалось загрузить переходы');
    }
}
function setupEventListeners() {
    document.getElementById('createStageBtn')?.addEventListener('click', openCreateStageModal);
    document.getElementById('stageList')?.addEventListener('click', (e) => {
        const target = e.target;
        if (target.classList.contains('edit-transitions')) {
            const id = Number(target.getAttribute('data-id'));
            selectStage(id);
        }
    });
    document.getElementById('addTransitionBtn')?.addEventListener('click', () => {
        addTransition().catch(err => {
            console.error('Ошибка добавления перехода:', err);
            alert('Ошибка при добавлении перехода');
        });
    });
}
function renderStageList() {
    const list = document.getElementById('stageList');
    if (!list)
        return;
    list.innerHTML = stages.map(stage => {
        return `
            <li class="status-item" data-id="${stage.id}">
                <span><strong>#${stage.position}</strong> ${escapeHtml(stage.name)}</span>
                <div class="actions">
                    <button class="btn btn-neutral btn-sm edit-transitions" data-id="${stage.id}">Редактировать переходы</button>
                </div>
            </li>
        `;
    }).join('');
}
function selectStage(stageId) {
    selectedStageId = stageId;
    const stage = stages.find(s => s.id === stageId);
    if (!stage)
        return;
    document.getElementById('currentStageName').textContent = stage.name;
    document.getElementById('transitionsSection').style.display = 'block';
    renderOutgoingTransitions(stageId);
    renderIncomingTransitions(stageId);
    renderToStageSelect(stageId);
}
function renderOutgoingTransitions(fromId) {
    const container = document.getElementById('outgoingTransitions');
    container.innerHTML = '';
    const outgoing = transitions.filter(t => t.from_stage_id === fromId);
    if (outgoing.length === 0) {
        container.innerHTML = '<div class="transition-row empty"><em>Нет разрешённых переходов</em></div>';
        return;
    }
    outgoing.forEach(t => {
        const toStage = stages.find(s => s.id === t.to_stage_id);
        const roles = t.allowed_roles.length ? `(${t.allowed_roles.join(', ')})` : '(все)';
        const row = document.createElement('div');
        row.className = 'transition-row';
        row.innerHTML = `
            <div class="transition-row-content">
                <strong>${escapeHtml(t.from_stage?.name || 'Неизвестно')}</strong>
                <span class="transition-arrow">→</span>
                <span>${escapeHtml(toStage?.name || 'Неизвестно')} ${roles}</span>
            </div>
            <button class="btn btn-danger btn-sm" data-id="${t.id}">Удалить</button>
        `;
        row.querySelector('.btn-danger')?.addEventListener('click', async (e) => {
            const id = Number(e.target.getAttribute('data-id'));
            await deleteTransition(id);
        });
        container.appendChild(row);
    });
}
function renderIncomingTransitions(toId) {
    const container = document.getElementById('incomingTransitions');
    container.innerHTML = '';
    const incoming = transitions.filter(t => t.to_stage_id === toId);
    if (incoming.length === 0) {
        container.innerHTML = '<div class="transition-row empty"><em>Никто не может перейти в эту стадию</em></div>';
        return;
    }
    incoming.forEach(t => {
        const fromStage = stages.find(s => s.id === t.from_stage_id);
        const row = document.createElement('div');
        row.className = 'transition-row';
        row.innerHTML = `
            <div class="transition-row-content">
                <span>${escapeHtml(fromStage?.name || 'Неизвестно')}</span>
                <span class="transition-arrow">→</span>
                <strong>${escapeHtml(t.to_stage?.name || 'Неизвестно')}</strong>
            </div>
        `;
        container.appendChild(row);
    });
}
function renderToStageSelect(fromId) {
    const select = document.getElementById('toStageSelect');
    select.innerHTML = '<option value="">Выберите стадию</option>';
    stages
        .filter(s => s.id !== fromId)
        .forEach(s => {
        select.innerHTML += `<option value="${s.id}">${escapeHtml(s.name)}</option>`;
    });
    const container = document.getElementById('rolesContainer');
    if (!container)
        return;
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
async function addTransition() {
    if (!selectedStageId)
        return;
    const select = document.getElementById('toStageSelect');
    const toId = Number(select.value);
    if (toId <= 0) {
        alert('Выберите целевую стадию');
        return;
    }
    const roleCheckboxes = document.querySelectorAll('input[name="allowed_roles"]:checked');
    const allowedRoles = Array.from(roleCheckboxes).map(el => el.value);
    if (allowedRoles.length === 0) {
        alert('Выберите хотя бы одну роль');
        return;
    }
    const transitionData = {
        from_stage_id: selectedStageId,
        to_stage_id: toId,
        allowed_roles: allowedRoles,
    };
    try {
        const res = await fetch('/api/v1/stage-transitions', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(transitionData),
        });
        if (res.ok) {
            const newTrans = await res.json();
            transitions.push(newTrans);
            renderOutgoingTransitions(selectedStageId);
            select.value = '';
            roleCheckboxes.forEach(el => el.checked = false);
        }
        else {
            const error = await res.json().catch(() => ({ error: 'Неизвестная ошибка' }));
            alert(`Ошибка: ${error.error || 'Не удалось добавить переход'}`);
        }
    }
    catch (err) {
        alert('Ошибка сети');
    }
}
async function deleteTransition(id) {
    if (!confirm('Удалить переход?'))
        return;
    try {
        const res = await fetch(`/api/v1/stage-transitions/${id}`, {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${token}` },
        });
        if (res.ok) {
            transitions = transitions.filter(t => t.id !== id);
            if (selectedStageId) {
                renderOutgoingTransitions(selectedStageId);
            }
        }
        else {
            alert('Ошибка удаления');
        }
    }
    catch (err) {
        alert('Ошибка сети');
    }
}
function openCreateStageModal() {
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <button type="button" class="modal-close-x" aria-label="Закрыть">×</button>
            <h3>Создать стадию</h3>
            <form id="createStageForm">
                <div class="form-group">
                    <label>Название *</label>
                    <input type="text" name="name" required />
                </div>
                <div class="form-group">
                    <label>Описание</label>
                    <textarea name="description"></textarea>
                </div>
                <div class="form-group">
                    <label>Позиция *</label>
                    <input type="number" name="position" min="0" required />
                </div>
                <div class="form-actions">
                    <button type="submit">Создать</button>
                </div>
            </form>
        </div>
    `;
    document.body.appendChild(modal);
    const form = modal.querySelector('#createStageForm');
    modal.querySelector('.modal-close-x')?.addEventListener('click', () => modal.remove());
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(form);
        const data = {
            name: formData.get('name').trim(),
            description: formData.get('description') || null,
            position: Number(formData.get('position')),
        };
        if (!data.name) {
            alert('Введите название');
            return;
        }
        try {
            const res = await fetch('/api/v1/stages', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data),
            });
            if (res.ok) {
                const newStage = await res.json();
                stages.push(newStage);
                renderStageList();
                modal.remove();
            }
            else {
                alert('Ошибка создания стадии');
            }
        }
        catch (err) {
            alert('Ошибка сети');
        }
    });
}
// === Выход ===
document.getElementById('logoutBtn')?.addEventListener('click', logout);
export {};
