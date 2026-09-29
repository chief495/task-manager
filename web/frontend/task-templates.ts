export {};

if (window.location.pathname !== '/admin/task-templates') {
    throw new Error('task-templates.ts loaded outside /admin/task-templates');
}

interface User {
    id: number;
    name: string;
    email: string;
    role: string;
}

interface Project {
    id: number;
    name: string;
    responsible_id: number;
}

interface TaskTemplate {
    id: number;
    name: string;
    description: string;
    category: string;
    priority: number;
    urgency: number;
    complexity: number;
    project_id: number | null;
    project?: Project | null;
    created_by_id?: number;
    created_by?: User;
}

let token: string | null = '';
let user: User | null = null;
let projects: Project[] = [];
let allTemplates: TaskTemplate[] = [];
let templates: TaskTemplate[] = [];
let currentFilter = 'all';

const content = document.getElementById('content');
if (!content) {
    throw new Error('Не найден элемент #content');
}

function logout(): void {
    localStorage.removeItem('token');
    window.location.href = '/login';
}

function escapeHtml(text: string): string {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function decodeEmailFromToken(activeToken: string | null): string | null {
    if (!activeToken) return null;

    try {
        const payload = activeToken.split('.')[1];
        if (!payload) return null;
        const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
        const parsed = JSON.parse(decoded);
        return parsed.email || null;
    } catch {
        return null;
    }
}

function getProjectName(projectId: number | null): string {
    if (!projectId) {
        return 'Общий шаблон';
    }

    return projects.find(project => project.id === projectId)?.name || `Проект #${projectId}`;
}

function getPriorityLabel(priority: number): string {
    switch (priority) {
        case 2: return 'Высокий';
        case 1: return 'Средний';
        case 0: return 'Низкий';
        default: return 'Неизвестно';
    }
}

function getManageableProjects(): Project[] {
    if (user?.role === 'admin') {
        return projects;
    }

    return projects.filter(project => project.responsible_id === user?.id);
}

function canManageTemplate(template: TaskTemplate): boolean {
    if (user?.role === 'admin') {
        return true;
    }

    if (!template.project_id) {
        return false;
    }

    return getManageableProjects().some(project => project.id === template.project_id);
}

function buildProjectFilterOptions(): string {
    const manageableProjects = getManageableProjects();
    const visibleProjects = user?.role === 'admin' ? projects : manageableProjects;

    return [
        '<option value="all">Все шаблоны</option>',
        '<option value="global">Только общие</option>',
        ...visibleProjects.map(project => `<option value="${project.id}">${escapeHtml(project.name)}</option>`),
    ].join('');
}

function renderPage(): void {
    content!.innerHTML = `
        <link rel="stylesheet" href="/static/css/app.css">
        <link rel="stylesheet" href="/static/css/task-templates.css?v=20260710-1">
        <div class="templates-page">
            <div class="templates-toolbar">
                <button id="createTemplateBtn" class="btn btn-primary">➕ Добавить шаблон</button>
                <div class="templates-filter">
                    <label for="templateProjectFilter">Фильтр</label>
                    <select id="templateProjectFilter">
                        ${buildProjectFilterOptions()}
                    </select>
                </div>
            </div>
            <div id="templatesList" class="templates-list"></div>
        </div>
    `;

    const filter = document.getElementById('templateProjectFilter') as HTMLSelectElement | null;
    if (filter) {
        filter.value = currentFilter;
    }

    setupEventListeners();
}

function setupEventListeners(): void {
    document.getElementById('createTemplateBtn')?.addEventListener('click', () => openTemplateModal());

    document.getElementById('templateProjectFilter')?.addEventListener('change', async (event) => {
        const target = event.target as HTMLSelectElement;
        currentFilter = target.value;
        applyTemplateFilter(currentFilter);
        renderTemplatesList();
    });

    document.getElementById('templatesList')?.addEventListener('click', async (event) => {
        const target = event.target as HTMLElement;
        const editButton = target.closest('.edit-template-btn') as HTMLButtonElement | null;
        const deleteButton = target.closest('.delete-template-btn') as HTMLButtonElement | null;

        if (editButton) {
            const templateId = Number(editButton.dataset.id);
            const template = templates.find(item => item.id === templateId);
            if (template) {
                openTemplateModal(template);
            }
        }

        if (deleteButton) {
            const templateId = Number(deleteButton.dataset.id);
            await deleteTemplate(templateId);
        }
    });
}

async function loadCurrentUser(): Promise<void> {
    try {
        const res = await fetch('/api/v1/users', {
            headers: { 'Authorization': `Bearer ${token}` },
        });

        if (res.status === 401) {
            logout();
            return;
        }

        const users: User[] = await res.json();
        const email = decodeEmailFromToken(token);
        user = email ? users.find(item => item.email === email) || users[0] : users[0];
    } catch (err) {
        console.error('Не удалось определить пользователя', err);
    }
}

async function loadProjects(): Promise<void> {
    try {
        const res = await fetch('/api/v1/projects', {
            headers: { 'Authorization': `Bearer ${token}` },
        });

        if (res.status === 401) {
            logout();
            return;
        }

        projects = await res.json();
    } catch (err) {
        console.error('Не удалось загрузить проекты', err);
        projects = [];
    }
}

async function fetchTemplatesByQuery(query: string): Promise<TaskTemplate[]> {
    const res = await fetch(`/api/v1/task-templates${query}`, {
        headers: { 'Authorization': `Bearer ${token}` },
    });

    if (res.status === 401) {
        logout();
        return [];
    }

    if (!res.ok) {
        throw new Error(`HTTP error ${res.status}`);
    }

    return await res.json();
}

function applyTemplateFilter(filterValue: string): void {
    if (filterValue === 'all') {
        templates = [...allTemplates];
        return;
    }

    if (filterValue === 'global') {
        templates = allTemplates.filter(template => !template.project_id);
        return;
    }

    const projectId = Number(filterValue);
    templates = allTemplates.filter(template => template.project_id === projectId);
}

async function loadTemplates(): Promise<void> {
    try {
        if (user?.role === 'admin') {
            allTemplates = await fetchTemplatesByQuery('');
        } else {
            const manageableProjects = getManageableProjects();
            const results = await Promise.all([
                fetchTemplatesByQuery('?project_id=0'),
                ...manageableProjects.map(project => fetchTemplatesByQuery(`?project_id=${project.id}`)),
            ]);

            allTemplates = Array.from(
                new Map(results.flat().map(template => [template.id, template])).values()
            );
        }

        applyTemplateFilter(currentFilter);
    } catch (err) {
        console.error('Не удалось загрузить шаблоны задач', err);
        allTemplates = [];
        templates = [];
    }
}

function renderTemplatesList(): void {
    const list = document.getElementById('templatesList');
    if (!list) return;

    if (templates.length === 0) {
        list.innerHTML = '<div class="templates-empty">Шаблоны задач не найдены.</div>';
        return;
    }

    list.innerHTML = templates.map(template => {
        const actions = canManageTemplate(template)
            ? `
                <button class="btn btn-neutral edit-template-btn" data-id="${template.id}">Редактировать</button>
                <button class="btn btn-danger delete-template-btn" data-id="${template.id}">Удалить</button>
            `
            : '';

        return `
            <article class="template-card">
                <div class="template-card-header">
                    <div>
                        <h3>${escapeHtml(template.name)}</h3>
                        <div class="template-badges">
                            <span class="template-badge">${escapeHtml(getProjectName(template.project_id))}</span>
                            <span class="template-badge priority-${template.priority}">${escapeHtml(getPriorityLabel(template.priority))}</span>
                        </div>
                    </div>
                    <div class="template-actions">
                        ${actions}
                    </div>
                </div>
                <div class="template-meta">Категория: ${escapeHtml(template.category || 'Не указана')}</div>
                <div class="template-meta">Важность: ${template.urgency ?? 1} · Сложность: ${template.complexity ?? 1}</div>
                <div class="template-description-block">
                    <div class="template-description-label">Описание</div>
                    <div class="template-description">${escapeHtml(template.description || 'Описание отсутствует')}</div>
                </div>
            </article>
        `;
    }).join('');
}

function openTemplateModal(template?: TaskTemplate): void {
    const manageableProjects = getManageableProjects();
    const isAdmin = user?.role === 'admin';

    if (!isAdmin && manageableProjects.length === 0) {
        alert('У вас нет проектов, для которых можно создавать шаблоны');
        return;
    }

    const selectedScope = template?.project_id ? 'project' : 'global';
    const selectedProjectId = template?.project_id || manageableProjects[0]?.id || 0;

    const scopeOptions = isAdmin
        ? `
            <option value="global" ${selectedScope === 'global' ? 'selected' : ''}>Общий шаблон</option>
            <option value="project" ${selectedScope === 'project' ? 'selected' : ''}>Шаблон проекта</option>
        `
        : '<option value="project" selected>Шаблон проекта</option>';

    const projectOptions = manageableProjects
        .map(project => `<option value="${project.id}" ${project.id === selectedProjectId ? 'selected' : ''}>${escapeHtml(project.name)}</option>`)
        .join('');

    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content template-modal-content">
            <button type="button" class="modal-close-x" aria-label="Закрыть">×</button>
            <h3>${template ? 'Редактировать шаблон' : 'Создать шаблон'}</h3>
            <form id="taskTemplateForm">
                <div class="form-group">
                    <label>Название *</label>
                    <input type="text" name="name" value="${escapeHtml(template?.name || '')}" required />
                </div>
                <div class="form-group">
                    <label>Описание</label>
                    <textarea name="description">${escapeHtml(template?.description || '')}</textarea>
                </div>
                <div class="form-group">
                    <label>Категория</label>
                    <input type="text" name="category" value="${escapeHtml(template?.category || '')}" />
                </div>
                <div class="form-group">
                    <label>Приоритет</label>
                    <select name="priority">
                        <option value="0" ${(template?.priority ?? 0) === 0 ? 'selected' : ''}>Низкий</option>
                        <option value="1" ${(template?.priority ?? 0) === 1 ? 'selected' : ''}>Средний</option>
                        <option value="2" ${(template?.priority ?? 0) === 2 ? 'selected' : ''}>Высокий</option>
                    </select>
                </div>
                <div class="form-row">
                    <div class="form-group">
                        <label>Важность</label>
                        <input type="number" name="urgency" min="0" step="0.1" value="${template?.urgency ?? 1}" />
                    </div>
                    <div class="form-group">
                        <label>Сложность</label>
                        <input type="number" name="complexity" min="0" step="0.1" value="${template?.complexity ?? 1}" />
                    </div>
                </div>
                <div class="form-group">
                    <label>Область шаблона</label>
                    <select name="scope" id="templateScopeSelect">${scopeOptions}</select>
                </div>
                <div class="form-group" id="templateProjectGroup">
                    <label>Проект</label>
                    <select name="project_id">${projectOptions}</select>
                </div>
                <div class="form-actions">
                    <button type="submit">${template ? 'Сохранить' : 'Создать'}</button>
                </div>
            </form>
        </div>
    `;

    document.body.appendChild(modal);

    const form = modal.querySelector('#taskTemplateForm') as HTMLFormElement;
    const scopeSelect = form.elements.namedItem('scope') as HTMLSelectElement;
    const projectGroup = modal.querySelector('#templateProjectGroup') as HTMLElement;
    modal.querySelector('.modal-close-x')?.addEventListener('click', () => modal.remove());

    const updateProjectVisibility = (): void => {
        projectGroup.style.display = scopeSelect.value === 'project' ? 'block' : 'none';
    };

    updateProjectVisibility();
    scopeSelect.addEventListener('change', updateProjectVisibility);

    form.addEventListener('submit', async (event) => {
        event.preventDefault();
        const formData = new FormData(form);
        const raw = Object.fromEntries(formData);

        if (!raw.name) {
            alert('Укажите название шаблона');
            return;
        }

        const payload = {
            name: raw.name as string,
            description: raw.description as string || '',
            category: raw.category as string || '',
            priority: Number(raw.priority) || 0,
            urgency: raw.urgency ? parseFloat(raw.urgency as string) : 1,
            complexity: raw.complexity ? parseFloat(raw.complexity as string) : 1,
            project_id: raw.scope === 'project' ? Number(raw.project_id) : null,
        };

        const endpoint = template ? `/api/v1/task-templates/${template.id}` : '/api/v1/task-templates';
        const method = template ? 'PUT' : 'POST';

        try {
            const res = await fetch(endpoint, {
                method,
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(payload),
            });

            if (res.status === 401) {
                logout();
                return;
            }

            if (!res.ok) {
                const err = await res.json().catch(() => ({ error: 'Неизвестная ошибка' }));
                alert(`Ошибка: ${err.error || 'Не удалось сохранить шаблон'}`);
                return;
            }

            modal.remove();
            await loadTemplates();
            renderTemplatesList();
        } catch (err) {
            console.error('Ошибка сохранения шаблона', err);
            alert('Ошибка сети');
        }
    });
}

async function deleteTemplate(templateId: number): Promise<void> {
    if (!confirm('Удалить шаблон задачи?')) {
        return;
    }

    try {
        const res = await fetch(`/api/v1/task-templates/${templateId}`, {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${token}` },
        });

        if (res.status === 401) {
            logout();
            return;
        }

        if (!res.ok) {
            const err = await res.json().catch(() => ({ error: 'Неизвестная ошибка' }));
            alert(`Ошибка: ${err.error || 'Не удалось удалить шаблон'}`);
            return;
        }

        await loadTemplates();
        renderTemplatesList();
    } catch (err) {
        console.error('Ошибка удаления шаблона', err);
        alert('Ошибка сети');
    }
}

document.addEventListener('DOMContentLoaded', async () => {
    token = localStorage.getItem('token');
    if (!token) {
        window.location.href = '/login';
        return;
    }

    await loadCurrentUser();
    await loadProjects();
    renderPage();
    await loadTemplates();
    renderTemplatesList();
});
