// tasks.ts — интерфейс управления задачами
// Обновлён: добавлены кнопки перевода статуса, "Редактировать" и фильтрация по проектам

export {};
declare const marked: any;
declare const Chart: any;
declare global {
    interface Window {
        downloadExcelExport?: (endpoint: string, fallbackFilename: string, trigger?: HTMLButtonElement | null) => Promise<void>;
    }
}

if (window.location.pathname !== '/tasks') {
    throw new Error('tasks.ts loaded outside /tasks');
}
// === Интерфейсные типы ===
interface Task {
    id: number;
    name: string;
    description: string | null;
    category: string | null;
    status_id: number;
    executor_id: number;
    project_id: number;
    start_time: string | null;
    end_time: string | null;  // ISO строка: "2025-04-05T10:00:00Z"
    priority: number;
    created_at: string;
    updated_at: string;
    parentID: number | null;           // nullable
    children: Task[];
    status?: Status;
    project?: Project;
    executor?: User;
    urgency?: number;
    complexity?: number;
}

interface TaskFormData {
    name: string;
    description?: string;
    category?: string;
    project_id: number;
    executor_id: number;
    status_id: number;
    priority: number;
    start_time?: string;
    end_time?: string;
    urgency?: number;
    complexity?: number;
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
    is_active: boolean;
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
    to_status: Status;
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
}

interface TaskFormData {
    name: string;
    description?: string;
    category?: string;
    project_id: number;
    executor_id: number;
    status_id: number;
    priority: number;
    start_time?: string;
    end_time?: string;
}

let tasks: Task[] = [];
let allProjectTasks: Task[] = []; // Задачи по проектам, где пользователь — ответственный
let filteredTasks: Task[] = [];
let projects: Project[] = [];
let users: User[] = [];
let statuses: Status[] = [];
let transitions: StatusTransition[] = [];
let taskTemplates: TaskTemplate[] = [];
let user: User | null = null;
let defaultStatusId: number = 1;
let token: string | null = '';
// === Обновление state ===
type ViewMode = 'list' | 'kanban' | 'gantt';
let currentView: ViewMode = 'kanban';
let currentChart: any = null; // Для хранения экземпляра Chart.js

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

function getProjectName(id: number): string {
    const project = projects.find(p => p.id === id);
    return project ? project.name : `Проект #${id}`;
}

function getUserName(id: number): string {
    const luser = users.find(u => u.id === id);
    return luser ? luser.name : `Пользователь #${id}`;
}

function getTaskCategories(): string[] {
    return Array.from(
        new Set(
            [...tasks, ...allProjectTasks]
                .map(task => (task.category || '').trim())
                .filter(category => category.length > 0)
        )
    ).sort((a, b) => a.localeCompare(b, 'ru'));
}

function getPriorityLabel(priority: number): string {
    switch (priority) {
        case 2: return 'Высокий';
        case 1: return 'Средний';
        case 0: return 'Низкий';
        default: return 'Неизвестно';
    }
}

function isFinalTaskStatus(statusId: number): boolean {
    const status = statuses.find(item => item.id === statusId);
    if (!status) return false;
    if (status.is_final) return true;

    const normalized = status.name.toLowerCase();
    return ['done', 'completed', 'готов', 'выполн', 'заверш', 'закрыт', 'аннулир'].some(marker => normalized.includes(marker));
}

function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString();
}

function getTaskDescriptionCheckboxStorageKey(taskId: number): string {
    return `task-description-checkboxes:${taskId}`;
}

function renderTaskDescription(taskId: number, description: string | null | undefined): string {
    if (!description) {
        return '';
    }

    const wrapper = document.createElement('div');
    wrapper.innerHTML = marked.parse(description, { gfm: true, breaks: true });

    const storageKey = getTaskDescriptionCheckboxStorageKey(taskId);
    let savedStates: boolean[] = [];

    try {
        const savedStatesRaw = localStorage.getItem(storageKey);
        savedStates = savedStatesRaw ? JSON.parse(savedStatesRaw) : [];
    } catch {
        savedStates = [];
    }

    wrapper.querySelectorAll('input[type="checkbox"]').forEach((checkbox, index) => {
        const input = checkbox as HTMLInputElement;
        input.removeAttribute('disabled');
        input.classList.add('task-markdown-checkbox');
        input.dataset.storageKey = storageKey;
        input.dataset.checkboxIndex = String(index);

        if (typeof savedStates[index] === 'boolean') {
            input.checked = savedStates[index];
            if (savedStates[index]) {
                input.setAttribute('checked', 'checked');
            } else {
                input.removeAttribute('checked');
            }
        }
    });

    return wrapper.innerHTML;
}

function attachTaskDescriptionCheckboxListeners(): void {
    document.querySelectorAll('.task-markdown-checkbox').forEach(checkbox => {
        const input = checkbox as HTMLInputElement;
        if (input.dataset.listenerAttached === 'true') {
            return;
        }

        input.dataset.listenerAttached = 'true';
        input.addEventListener('click', (event) => {
            event.stopPropagation();
        });

        input.addEventListener('change', (event) => {
            event.stopPropagation();

            const target = event.currentTarget as HTMLInputElement;
            const storageKey = target.dataset.storageKey;
            const checkboxIndex = Number(target.dataset.checkboxIndex);

            if (!storageKey || Number.isNaN(checkboxIndex)) {
                return;
            }

            let savedStates: boolean[] = [];

            try {
                const savedStatesRaw = localStorage.getItem(storageKey);
                savedStates = savedStatesRaw ? JSON.parse(savedStatesRaw) : [];
            } catch {
                savedStates = [];
            }

            savedStates[checkboxIndex] = target.checked;
            localStorage.setItem(storageKey, JSON.stringify(savedStates));
        });
    });
}

function canEditTask(task: Task): boolean {
    if (!user || !task.project) return false;
    return user.role === 'admin' || user.id === task.project.responsible_id;
}

function getTaskById(taskId: number): Task | undefined {
    return allProjectTasks.find(t => t.id === taskId) || tasks.find(t => t.id === taskId);
}
// === Проверка прав ===
function canCreateTask(): boolean {
    if (!user) return false;

    // Админ может везде
    if (user.role === 'admin') return true;

    // Ответственный — только по своим проектам
    const managedProjects = projects.filter(p => p.responsible_id === user!.id);
    return managedProjects.length > 0;
}

function canAssignToProject(projectId: number): boolean {
    if (!user || !projectId) return false;

    if (user.role === 'admin') return true;

    const project = projects.find(p => p.id === projectId);
    return project?.responsible_id === user.id;
}
// === Загрузка данных ===
async function loadTasks(): Promise<void> {
    if (user == null) {
        setTimeout(function (){
            console.log(user);
        }, 1000);
    }
    try {
        // Если пользователь — админ, загружаем все задачи
        if (user?.role === 'admin') {
            const res = await fetch('/api/v1/tasks', {
                headers: { 'Authorization': `Bearer ${token}` },
            });

            if (res.status === 401) return logout();
            if (!res.ok) {
                throw new Error(`HTTP error! status: ${res.status}`);
            }
            const loadedTasks: Task[] = await res.json();
            tasks = loadedTasks;
            allProjectTasks = loadedTasks;
        } else {
            // Обычный пользователь: задачи, где он исполнитель
            const myTasksRes = await fetch('/api/v1/tasks', {
                headers: { 'Authorization': `Bearer ${token}` },
            });

            if (myTasksRes.status === 401) return logout();
            if (!myTasksRes.ok) {
                throw new Error(`HTTP error! status: ${myTasksRes.status}`);
            }

            tasks = await myTasksRes.json();

            // Загружаем задачи по проектам, где пользователь — ответственный
            const managedProjectIds = projects
                .filter(p => p.responsible_id === user?.id)
                .map(p => p.id);
            if (managedProjectIds.length > 0) {
                const projectTasksRes = await fetch(`/api/v1/projects/tasks?project_ids=${managedProjectIds.join(',')}`, {
                    headers: { 'Authorization': `Bearer ${token}` },
                });
                if (projectTasksRes.ok) {
                    allProjectTasks = await projectTasksRes.json();
                }
            }
        }

        applyFilters();
    } catch (err) {
        console.error('Ошибка загрузки задач:', err);
        showError('Не удалось загрузить задачи');
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

async function loadUsers(): Promise<void> {
    try {
        const res = await fetch('/api/v1/users', {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        users = await res.json();
    } catch (err) {
        console.warn('Пользователи не загружены');
        users = [];
    }
}

async function loadStatuses(): Promise<void> {
    try {
        const res = await fetch('/api/v1/statuses', {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        statuses = await res.json();

        const defaultStatus = statuses.find(s => s.is_default) ||
            statuses.find(s => s.name.toLowerCase().includes('выполнению') || s.name === 'To Do') ||
            statuses[0];

        defaultStatusId = defaultStatus?.id || 1;
    } catch (err) {
        console.warn('Статусы не загружены');
        defaultStatusId = 1;
    }
}

async function loadTransitions(): Promise<void> {
    try {
        const res = await fetch('/api/v1/status-transitions', {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        transitions = await res.json();
    } catch (err) {
        console.warn('Переходы не загружены');
        transitions = [];
    }
}

async function loadTaskTemplates(projectId?: number): Promise<TaskTemplate[]> {
    const query = typeof projectId === 'number'
        ? `?project_id=${projectId}`
        : '?project_id=0';

    try {
        const res = await fetch(`/api/v1/task-templates${query}`, {
            headers: { 'Authorization': `Bearer ${token}` },
        });

        if (res.status === 401) {
            logout();
            return [];
        }

        if (!res.ok) {
            throw new Error(`HTTP error! status: ${res.status}`);
        }

        taskTemplates = await res.json();
        return taskTemplates;
    } catch (err) {
        console.warn('Шаблоны задач не загружены', err);
        taskTemplates = [];
        return [];
    }
}

// === Загрузка данных ===
async function loadCurrentUser(): Promise<void> {
    try {
        const tokenUser = decodeUserFromToken(token);
        const res = await fetch('/api/v1/users', {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        const allUsers: User[] = await res.json();

        user = allUsers.find(u => u.id === tokenUser?.id)
            || allUsers.find(u => u.email === tokenUser?.email)
            || (tokenUser ? { id: tokenUser.id, email: tokenUser.email, role: tokenUser.role, name: tokenUser.email, is_active: true } : null);

        if (user) {
            localStorage.setItem('user', JSON.stringify(user));
        }
    } catch (err) {
        console.error('Не удалось определить текущего пользователя');
        const tokenUser = decodeUserFromToken(token);
        user = tokenUser ? { id: tokenUser.id, email: tokenUser.email, role: tokenUser.role, name: tokenUser.email, is_active: true } : null;
    }
}

function decodeUserFromToken(token: string | null): Pick<User, 'id' | 'email' | 'role'> | null {
    if (!token) return null;

    try {
        const payload = token.split('.')[1];
        if (!payload) return null;

        const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
        const parsed = JSON.parse(decoded);
        const id = Number(parsed.user_id);
        if (!id || !parsed.email || !parsed.role) return null;
        return { id, email: parsed.email, role: parsed.role };
    } catch (e) {
        console.warn('Не удалось декодировать JWT', e);
        return null;
    }
}

// === Фильтрация ===
function applyFilters(): void {
    const projectId = Number((document.getElementById('projectFilter') as HTMLSelectElement)?.value);
    const category = ((document.getElementById('categoryFilter') as HTMLSelectElement)?.value || '').trim();

    // Фильтрация только по задачам, где пользователь — исполнитель
    let myTasks = tasks.filter(t => t.executor_id === user?.id);
    // Применяем фильтр по проекту
    if (projectId) {
        myTasks = myTasks.filter(t => t.project_id === projectId);
    }
    if (category) {
        myTasks = myTasks.filter(t => (t.category || '').trim() === category);
    }

    filteredTasks = myTasks;
    renderTasks();
}

// === Рендеринг интерфейса ===
// === Кнопки переключения вида — обновлено ===
function renderPage(): void {
    const categories = getTaskCategories();

    content!.innerHTML = `
        <link rel="stylesheet" href="/static/css/app.css">
        <link rel="stylesheet" href="/static/css/tasks.css?v=20260710-1">
        <div class="controls tasks-controls">
            <button id="createTaskBtn" class="btn btn-primary">➕ Создать задачу</button>
            <button id="exportTasksBtn" class="btn btn-excel">Скачать Excel</button>

            <div class="view-controls">
                <button id="viewListBtn" class="view-btn">Список</button>
                <button id="viewKanbanBtn" class="view-btn">Kanban</button>
                <button id="viewGanttBtn" class="view-btn">Ганта</button>
            </div>

            <div class="task-filter-group task-filter-group-project">
                <label for="projectFilter">Фильтр по проекту:</label>
                <select id="projectFilter">
                    <option value="">Все проекты</option>
                    ${projects.map(p => `<option value="${p.id}">${escapeHtml(p.name)}</option>`).join('')}
                </select>
            </div>
            <div class="task-filter-group">
                <label for="categoryFilter">Фильтр по категории:</label>
                <select id="categoryFilter">
                    <option value="">Все категории</option>
                    ${categories.map(category => `<option value="${escapeHtml(category)}">${escapeHtml(category)}</option>`).join('')}
                </select>
            </div>
        </div>
        <div id="tasksContainer"></div>
    `;

    setupEventListeners();
    updateViewButtons();
}

// === Обновление обработчиков ===
function setupEventListeners(): void {
    document.getElementById('projectFilter')?.addEventListener('change', applyFilters);
    document.getElementById('categoryFilter')?.addEventListener('change', applyFilters);

    if (!content!.dataset.tasksControlsDelegated) {
        content!.addEventListener('click', (event) => {
            const target = event.target as HTMLElement;

            if (target.closest('#createTaskBtn')) {
                openCreateTaskModal();
                return;
            }

            const exportTasksBtn = target.closest('#exportTasksBtn');
            if (exportTasksBtn instanceof HTMLButtonElement) {
                window.downloadExcelExport?.('/api/v1/export/tasks', 'tasks.xlsx', exportTasksBtn);
                return;
            }

            if (target.closest('#viewListBtn')) {
                setView('list');
                return;
            }

            if (target.closest('#viewKanbanBtn')) {
                setView('kanban');
                return;
            }

            if (target.closest('#viewGanttBtn')) {
                setView('gantt');
            }
        });

        content!.dataset.tasksControlsDelegated = 'true';
    }
}

function updateViewButtons(): void {
    const listBtn = document.getElementById('viewListBtn');
    const kanbanBtn = document.getElementById('viewKanbanBtn');
    const ganttBtn = document.getElementById('viewGanttBtn');

    listBtn?.classList.toggle('active', currentView === 'list');
    kanbanBtn?.classList.toggle('active', currentView === 'kanban');
    ganttBtn?.classList.toggle('active', currentView === 'gantt');
}

function setView(view: ViewMode): void {
    currentView = view;
    updateViewButtons();
    renderTasks();
}

function renderTasks(): void {
    const container = document.getElementById('tasksContainer');
    if (!container) return;

    container.innerHTML = '';

    // Блок: задачи пользователя
    const myTasksSection = document.createElement('div');
    myTasksSection.className = 'tasks-section';
    myTasksSection.innerHTML = `<h3 class="tasks-section-title">Мои задачи</h3>`;
    const myTasksContainer = document.createElement('div');
    myTasksContainer.id = 'myTasksList';
    myTasksContainer.className = 'tasks-section-content';

    const activeFilteredTasks = filteredTasks.filter(task => !isFinalTaskStatus(task.status_id));

    if (activeFilteredTasks.length === 0) {
        myTasksContainer.innerHTML = '<p>У вас нет назначенных задач.</p>';
    } else {
        if (currentView === 'list') {
            renderTaskList(activeFilteredTasks, myTasksContainer);
        } else if (currentView === 'kanban') {
            renderKanbanBoard(activeFilteredTasks, myTasksContainer);
        } else if (currentView === 'gantt') {
            renderGanttChart(activeFilteredTasks, myTasksContainer);
        }
    }

    myTasksSection.appendChild(myTasksContainer);
    container.appendChild(myTasksSection);


    // Блок: задачи по проектам (только если есть управляемые проекты)
    const managedProjects = projects.filter(p => p.responsible_id === user?.id);
    if (managedProjects.length > 0 || user?.role === 'admin') {
        const projectTasksSection = document.createElement('div');
        projectTasksSection.className = 'tasks-section';
        projectTasksSection.innerHTML = `<h3 class="tasks-section-title">Задачи по проектам, где вы — ответственный</h3>`;
        const projectTasksContainer = document.createElement('div');
        projectTasksContainer.id = 'projectTasksList';
        projectTasksContainer.className = 'tasks-section-content';

        if (allProjectTasks.length === 0) {
            projectTasksContainer.innerHTML = '<p>Нет задач в ваших проектах.</p>';
        } else {
            const category = ((document.getElementById('categoryFilter') as HTMLSelectElement)?.value || '').trim();
            let visibleProjectTasks = allProjectTasks;
            if (category) {
                visibleProjectTasks = visibleProjectTasks.filter(task => (task.category || '').trim() === category);
            }

            if (visibleProjectTasks.length === 0) {
                projectTasksContainer.innerHTML = '<p>Нет задач по выбранной категории.</p>';
                projectTasksSection.appendChild(projectTasksContainer);
                container.appendChild(projectTasksSection);
                return;
            }

            if (currentView === 'list') {
                renderTaskList(visibleProjectTasks, projectTasksContainer);
            } else if (currentView === 'kanban') {
                renderKanbanBoard(visibleProjectTasks, projectTasksContainer);
            } else if (currentView === 'gantt') {
                renderGanttChart(visibleProjectTasks, projectTasksContainer);
            }
        }

        projectTasksSection.appendChild(projectTasksContainer);
        container.appendChild(projectTasksSection);
    }
}

// Обновляем функции рендера, чтобы принимали контейнер
function renderTaskList(tasksToRender: Task[], container: HTMLElement): void {
    const ul = document.createElement('ul');
    ul.className = 'task-list';
    const tasksToRenderSorted = tasksToRender.sort((a, b) => {
        const aDate = new Date(a.created_at);
        const bDate = new Date(b.created_at);
        return aDate < bDate ? 1 : -1;
    })
    ul.innerHTML = tasksToRenderSorted.map(task => {
        const status = statuses.find(s => s.id === task.status_id) || { name: 'Неизвестно' };
        const editBtn = canEditTask(task)
            ? `<button class="btn btn-neutral btn-sm edit-btn" data-id="${task.id}" style="margin-left: 10px;">Редактировать</button>`
            : '';
        const transitionsHtml = getAvailableTransitions(task).map(t => {
            const toStatus = statuses.find(s => s.id === t.to_status_id) || t.to_status;
            return `<button class="transition-btn small" data-task="${task.id}" data-to="${t.to_status_id}" style="margin: 2px;">→ ${escapeHtml(toStatus.name)}</button>`;
        }).join('');
        let startTimeInfo = '';
        if (task.start_time) {
            const date = new Date(task.start_time);
            const formattedDate = date.toLocaleDateString();
            const formattedTime = date.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
            startTimeInfo = `<div class="task-meta">Начало: ${formattedDate} ${formattedTime}</div>`;
        }
        let deadlineInfo = '';
        if (task.end_time) {
            const date = new Date(task.end_time);
            const formattedDate = date.toLocaleDateString();
            const formattedTime = date.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
            deadlineInfo = `<div class="task-meta">Срок: ${formattedDate} ${formattedTime}</div>`;
        }
        const statusClass = status.name === "В работе" ? "work" :
                                            status.name === "На проверке" ? "check":
                                                status.name === "Выполнено" ? "done":
                                                        status.name === "Пауза" ? "pause":
                                                            status.name === "Аннулировано" ? "cancelled":
                                                                status.name === "Создана" ? "created" : "";
        const formattedText = renderTaskDescription(task.id, task.description);
        return `
            <li class="task-item">
                <div class="task-header">${escapeHtml(task.name)} <span class="status-badge ${statusClass}">${escapeHtml(status.name)}</span></div>
                <div class="task-meta">Проект: ${getProjectName(task.project_id)}</div>
                <div class="task-meta">Исполнитель: ${getUserName(task.executor_id)}</div>
                <div class="task-meta">Категория: ${escapeHtml(task.category || 'Не указана')}</div>
                <div class="task-meta">Приоритет: ${getPriorityLabel(task.priority)}</div>
                ${startTimeInfo}
                ${deadlineInfo}
                ${task.description ? `
                 <div class="task-meta task-description-label">Описание:</div>
                 <div class="task-description-content task-description-block" style="padding-left: 8px;">
                    ${formattedText}
                 </div>` : ''}
                 <div class="task-actions" style="margin-top: 10px; display: flex; flex-wrap: wrap; gap: 4px;">
                    ${transitionsHtml}
                    ${editBtn}
                </div>
            </li>
        `;
    }).join('');

    container.appendChild(ul);
    // Устанавливаем делегирование сразу после рендеринга
    setTimeout(() => {
        attachTaskEventListeners();
        attachTaskDescriptionCheckboxListeners();
    }, 0);
}

function updateTaskCardStyle(taskCard: HTMLElement, task: Task): void {
    const today = new Date();
    const endDate = new Date(task.end_time!);
    today.setHours(0, 0, 0, 0);
    endDate.setHours(0, 0, 0, 0);

    // Проверяем, просрочена ли задача
    const isOverdue = task.status?.name !== 'Выполнено' && task.status?.name !== 'Пауза' && today >= endDate;

    if (isOverdue) {
        taskCard.style.backgroundColor = '#3a1818'; // Темно-красный фон
        taskCard.style.borderLeft = '4px solid #ea4335'; // Красная полоса слева
    } else {
        taskCard.style.backgroundColor = '#1e1e1e'; // Стандартный фон
        taskCard.style.borderLeft = '4px solid #1a73e8'; // Синяя полоса слева
    }
}
function renderKanbanBoard(tasksToRender: Task[], container: HTMLElement): void {
    const board = document.createElement('div');
    board.className = 'kanban-board';

    statuses.forEach(status => {
        const column = document.createElement('div');
        column.className = 'kanban-column';
        column.dataset.taskDropStatus = status.id.toString();
        column.innerHTML = `<h3>${escapeHtml(status.name)}</h3>`;

        const columnTasks = tasksToRender.filter(t => t.status_id === status.id);
        columnTasks.forEach(task => {
            const weekAgo = new Date();
            weekAgo.setDate(weekAgo.getDate() - 7);
            const endDate = new Date(task.end_time!);
            if ((task.status?.name === "Аннулировано" || task.status?.name === "Выполнено") && endDate < weekAgo) {
                return;
            }

            // Создаем карточку задачи
            const taskCard = document.createElement('div');
            taskCard.className = 'task-item kanban-task';
            taskCard.dataset.taskId = task.id.toString();
            taskCard.dataset.taskStatusId = task.status_id.toString();
            taskCard.draggable = true;
            updateTaskCardStyle(taskCard, task);

            // Краткое отображение
            taskCard.innerHTML = `
                <div class="task-header">${escapeHtml(task.name)}</div>
                <div class="task-meta">Проект: ${getProjectName(task.project_id)}</div>
                <div class="task-meta">Исполнитель: ${getUserName(task.executor_id)}</div>
                <div class="task-meta">Категория: ${escapeHtml(task.category || 'Не указана')}</div>
                <div class="task-meta">Приоритет: ${getPriorityLabel(task.priority)}</div>
            `;

            // Подробное отображение (скрыто по умолчанию)
            const details = document.createElement('div');
            details.className = 'task-details kanban-task-details';
            details.style.display = 'none';
            details.style.paddingTop = '8px';
            details.style.marginTop = '8px';
            details.style.borderTop = '1px solid #444';
            details.innerHTML = `${task.description ?
                `<div class="task-meta task-description-label">Описание:</div><div class="task-description-content task-description-block" style="padding-left: 8px;">${renderTaskDescription(task.id, task.description)}</div>` : ''}
                ${task.start_time ? `<div class="task-meta">Начало: ${formatDate(task.start_time)}</div>` : ''}
                ${task.end_time ? `<div class="task-meta">Срок: ${formatDate(task.end_time)}</div>` : ''}
                <div class="task-actions" style="margin-top: 10px; display: flex; flex-wrap: wrap; gap: 4px;">
                    ${getAvailableTransitions(task).map(t => {
                const toStatus = statuses.find(s => s.id === t.to_status_id) || t.to_status;
                return `<button class="transition-btn small" data-task="${task.id}" data-to="${t.to_status_id}" style="margin: 2px;">→ ${escapeHtml(toStatus.name)}</button>`;
            }).join('')}
                    ${canEditTask(task) ? `<button class="btn btn-neutral btn-sm edit-btn" data-id="${task.id}" style="margin-left: 10px;">Редактировать</button>` : ''}
                </div>
            `;

            taskCard.appendChild(details);
            column.appendChild(taskCard);
        });

        board.appendChild(column);
    });

    container.appendChild(board);

    // Устанавливаем обработчики после рендеринга
    setTimeout(() => {
        attachTaskEventListeners();
        attachTaskExpandListeners();
        setupTaskDragAndDrop(container);
        attachTaskDescriptionCheckboxListeners();
    }, 0);
}

function attachTaskExpandListeners(): void {
    document.querySelectorAll('.task-item').forEach(card => {
        card.addEventListener('click', (e) => {
            // Проверяем, был ли клик по кнопке внутри карточки
            const target = e.target as HTMLElement;
            if (target.closest('.transition-btn') || target.closest('.edit-btn') || target.closest('.task-markdown-checkbox') || target.closest('label')) {
                return; // Не сворачиваем при клике по кнопкам
            }

            const details = card.querySelector('.task-details') as HTMLElement;
            if (details) {
                const isExpanded = details.style.display !== 'none';
                details.style.display = isExpanded ? 'none' : 'block';

                // Добавляем/удаляем класс для визуального выделения
                card.classList.toggle('expanded');
            }
        });
    });
}

function setupTaskDragAndDrop(container: HTMLElement): void {
    container.querySelectorAll<HTMLElement>('.kanban-task[draggable="true"]').forEach(card => {
        card.addEventListener('dragstart', event => {
            const taskId = card.dataset.taskId;
            const statusId = card.dataset.taskStatusId;
            if (!taskId || !statusId || !event.dataTransfer) return;

            event.dataTransfer.effectAllowed = 'move';
            event.dataTransfer.setData('application/x-task-id', taskId);
            event.dataTransfer.setData('application/x-task-status-id', statusId);
            card.classList.add('dragging');
        });

        card.addEventListener('dragend', () => {
            card.classList.remove('dragging');
            clearTaskDropTargets(container);
        });
    });

    container.querySelectorAll<HTMLElement>('[data-task-drop-status]').forEach(column => {
        column.addEventListener('dragover', event => {
            if (!event.dataTransfer?.types.includes('application/x-task-id')) return;
            event.preventDefault();
            event.dataTransfer.dropEffect = 'move';
            column.classList.add('drag-over');
        });

        column.addEventListener('dragleave', event => {
            if (!column.contains(event.relatedTarget as Node | null)) {
                column.classList.remove('drag-over');
            }
        });

        column.addEventListener('drop', async event => {
            event.preventDefault();
            const taskId = Number(event.dataTransfer?.getData('application/x-task-id'));
            const fromStatusId = Number(event.dataTransfer?.getData('application/x-task-status-id'));
            const toStatusId = Number(column.dataset.taskDropStatus);
            clearTaskDropTargets(container);

            if (!taskId || !toStatusId || fromStatusId === toStatusId) return;
            const changed = await handleStatusTransition(taskId, toStatusId, false);
            if (!changed) {
                markInvalidDropTarget(column);
            }
        });
    });
}

function clearTaskDropTargets(container: HTMLElement): void {
    container.querySelectorAll('.kanban-column.drag-over').forEach(column => {
        column.classList.remove('drag-over');
    });
}

function markInvalidDropTarget(element: HTMLElement): void {
    element.classList.add('drop-invalid');
    window.setTimeout(() => element.classList.remove('drop-invalid'), 1200);
}
// === Функция рендера диаграммы Ганта ===
function renderGanttChart(tasksToRender: Task[], container: HTMLElement): void {
    type GanttTaskRow = {
        type: 'task';
        key: string;
        displayLabel: string;
        project: string;
        start: Date;
        end: Date;
        executor: string;
        status: string;
        id: number;
    };

    type GanttHeaderRow = {
        type: 'header';
        key: string;
        displayLabel: string;
    };

    type GanttRow = GanttTaskRow | GanttHeaderRow;

    // Определяем начальные и конечные даты
    const today = new Date();
    const defaultStartDate = new Date(today);
    const defaultEndDate = new Date(today);
    defaultEndDate.setDate(defaultEndDate.getDate() + 7);
    const formatLocalDate = (date: Date): string => {
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        return `${year}-${month}-${day}`;
    };
    const chartId = `ganttChart_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

    // Очищаем предыдущий график
    container.innerHTML = `
        <div class="gantt-chart-shell" style="width: 100%;">
            <div style="display: flex; align-items: center; margin-top: 10px; margin-bottom: 10px; gap: 8px; flex-wrap: wrap;">
                <label class="gantt-filters" style="color: #ddd; white-space: nowrap;">
                    С: <input type="date" id="${chartId}_startDate" value="${formatLocalDate(defaultStartDate)}">
                </label>
                <label class="gantt-filters" style="color: #ddd; white-space: nowrap;">
                    По: <input type="date" id="${chartId}_endDate" value="${formatLocalDate(defaultEndDate)}">
                </label>
                <span class="today-badge" style="color: #4bb4e6; font-weight: 500; white-space: nowrap;">
                    Сегодня: ${today.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' })}
                </span>
            </div>
            <div class="gantt-chart-wrap" id="${chartId}_wrap" style="position: relative; height: 520px; width: 100%;">
                <canvas id="${chartId}" class="gantt-chart-canvas"></canvas>
            </div>
        </div>
    `;

    // Отложенный рендеринг для гарантии отрисовки canvas
    setTimeout(() => {
        const canvas = container.querySelector(`canvas#${chartId}`) as HTMLCanvasElement;
        if (!canvas) {
            console.error('Canvas #ganttChart не найден после вставки');
            return;
        }

        const ctx = canvas.getContext('2d');
        if (!ctx) {
            console.error('Не удалось получить 2D контекст');
            return;
        }

        const chartWrap = container.querySelector(`#${chartId}_wrap`) as HTMLElement;
        if (!chartWrap) {
            console.error('Контейнер диаграммы не найден');
            return;
        }

        if (currentChart) {
            currentChart.destroy();
            currentChart = null;
        }

        // === Создаем изолированный экземпляр Chart.js с уникальным именем ===
        const chartInstance = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: [],
                datasets: [{
                    data: [],
                    backgroundColor: [],
                    borderWidth: 3,
                    barPercentage: 1,
                    categoryPercentage: 1,
                }]
            },
            options: {
                indexAxis: 'y',
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false },
                    tooltip: {
                        callbacks: {
                            title: () => '',
                            label: (context: any) => {
                                const task = context.raw?.task;
                                if (!task) return ['Задача'];
                                return [
                                    `Задача: ${task.label}`,
                                    `Проект: ${task.project}`,
                                    `Статус: ${task.status}`,
                                    `Исполнитель: ${task.executor}`,
                                    `С: ${task.start.toLocaleDateString()}`,
                                    `По: ${task.end.toLocaleDateString()}`
                                ];
                            }
                        }
                    },
                    annotation: {
                        annotations: {
                            todayLine: {
                                type: 'line',
                                xMin: today,
                                xMax: today,
                                borderColor: '#ff6b6b',
                                borderWidth: 2,
                                label: {
                                    display: true,
                                    content: 'Сегодня',
                                    color: '#fff',
                                    backgroundColor: '#ff6b6b',
                                    font: { size: 10, weight: 'bold' },
                                    position: 'end'
                                }
                            }
                        }
                    }
                },
                scales: {
                    x: {
                        type: 'time',
                        time: { unit: 'day', displayFormats: { day: 'dd MMM' }, tooltipFormat: 'dd MMM yyyy' },
                        title: { display: true, text: 'Время' },
                        grid: { color: '#444' },
                        ticks: { color: '#e0e0e0' },
                        min: defaultStartDate,
                        max: defaultEndDate
                    },
                    y: {
                        type: 'category',
                        title: { display: true, text: 'Исполнители и задачи' },
                        grid: { color: '#444' },
                        ticks: {
                            color: function (context: any) {
                                const key = context.chart?.data?.labels?.[context.index];
                                if (typeof key !== 'string') {
                                    return '#e0e0e0';
                                }
                                return key.startsWith('header-') ? '#ffffff' : '#b8c0cc';
                            },
                            font: function (context: any) {
                                const key = context.chart?.data?.labels?.[context.index];
                                const isHeader = typeof key === 'string' && key.startsWith('header-');
                                return {
                                    size: isHeader ? 14 : 12,
                                    weight: isHeader ? '800' : '400'
                                };
                            },
                            callback: function (_value: any, index: number) {
                                const label = chartInstance.data.labels?.[index];
                                if (typeof label === 'string') {
                                    return label;
                                }
                                return '';
                            }
                        }
                    }
                }
            }
        });
        function updateChart() {
            const startInput = container.querySelector(`#${chartId}_startDate`) as HTMLInputElement;
            const endInput = container.querySelector(`#${chartId}_endDate`) as HTMLInputElement;

            if (!startInput || !endInput) return;

            const startDate = new Date(startInput.value);
            const endDate = new Date(endInput.value);
            endDate.setHours(23, 59, 59, 999);

            const groupedTasks = tasksToRender
                .filter(t => t.start_time && t.end_time)
                .map((task): GanttTaskRow => {
                    const status = statuses.find(s => s.id === task.status_id) || { name: 'Неизвестно' };
                    return {
                        key: `task-${task.id}`,
                        displayLabel: `    ${task.name}`,
                        project: getProjectName(task.project_id),
                        start: new Date(task.start_time!),
                        executor: task.executor?.name || getUserName(task.executor_id),
                        end: new Date(task.end_time!),
                        status: status.name,
                        id: task.id,
                        type: 'task'
                    };
                })
                .filter(t => (t.start <= endDate || t.status === 'В работе')&& t.end >= startDate)
                .reduce<Record<string, GanttTaskRow[]>>((acc, task) => {
                    if (!acc[task.executor]) {
                        acc[task.executor] = [];
                    }
                    acc[task.executor].push(task);
                    return acc;
                }, {});

            const ganttRows: GanttRow[] = Object.entries(groupedTasks)
                .sort(([executorA], [executorB]) => executorA.localeCompare(executorB, 'ru'))
                .flatMap(([executor, executorTasks]) => {
                    const sortedTasks = executorTasks.sort((a, b) => a.start.getTime() - b.start.getTime());
                    const headerRow: GanttHeaderRow = {
                        type: 'header',
                        key: `header-${executor}`,
                        displayLabel: executor
                    };

                    return [headerRow, ...sortedTasks];
                });

            if (ganttRows.length === 0) {
                chartInstance.data.labels = ['Нет задач в диапазоне'];
                chartInstance.data.datasets[0].data = [{}];
                chartInstance.data.datasets[0].backgroundColor = ['#666'];
            } else {
                chartInstance.data.labels = ganttRows.map(row => row.key);
                chartInstance.data.datasets[0].data = ganttRows
                    .filter((row): row is GanttTaskRow => row.type === 'task')
                    .map(taskRow => ({
                        x: [taskRow.start, taskRow.end],
                        y: taskRow.key,
                        task: taskRow
                    }));
                chartInstance.data.datasets[0].backgroundColor = ganttRows
                    .filter((row): row is GanttTaskRow => row.type === 'task')
                    .map(taskRow => {
                        const colors: Record<string, string> = {
                            'В работе': '#1a73e8',
                            'На проверке': '#fbbc05',
                            'Выполнено': '#34a853',
                            'Создана': '#666',
                            'Аннулировано': '#ea4335',
                            'Пауза': '#fb6c05'
                        };
                        return colors[taskRow.status] || '#999';
                    });
            }

            const chartHeight = Math.max(520, ganttRows.length * 34);
            chartWrap.style.height = `${chartHeight}px`;

            // Обновляем диапазон оси X
            chartInstance.options.scales!.x!.min = startDate;
            chartInstance.options.scales!.x!.max = endDate;
            chartInstance.options.scales!.y!.ticks!.callback = function (_value: any, index: number) {
                const key = chartInstance.data.labels?.[index];
                const row = ganttRows.find(item => item.key === key);
                return row?.displayLabel || '';
            };
            chartInstance.options.elements.bar.borderSkipped = 'end';
            chartInstance.resize();
            chartInstance.update();
        }
        // Обработчики изменений дат
        const startInput = container.querySelector(`#${chartId}_startDate`) as HTMLInputElement;
        const endInput = container.querySelector(`#${chartId}_endDate`) as HTMLInputElement;

        startInput?.addEventListener('change', updateChart);
        endInput?.addEventListener('change', updateChart);

        // Инициализация графика
        updateChart();

        // Очищаем предыдущий график при перерисовке
        return () => {
            if (chartInstance) {
                chartInstance.destroy();
            }
        };
    }, 50);
}

function getAvailableTransitions(task: Task): StatusTransition[] {
    if (!user || !task.project) return [];

    return transitions
        .filter(t => t.from_status_id === task.status_id)
        .filter(t => {
            if (t.allowed_roles.includes('all')) return true;
            if (t.allowed_roles.includes(user!.role)) return true;
            return t.allowed_roles.includes('manager') && task.project?.responsible_id === user!.id;

        });
}

function attachTaskEventListeners(): void {
    const myTasksList = document.getElementById('myTasksList');
    const projectTasksList = document.getElementById('projectTasksList');

    const handler = (e: Event) => {
        const target = e.target as HTMLElement;

        if (target.classList.contains('task-markdown-checkbox')) {
            return;
        }

        const transitionButton = target.closest('.transition-btn') as HTMLElement | null;
        const editButton = target.closest('.edit-btn') as HTMLElement | null;

        if (transitionButton) {
            const taskId = Number(transitionButton.getAttribute('data-task'));
            const toStatusId = Number(transitionButton.getAttribute('data-to'));
            if (taskId && toStatusId) {
                handleStatusTransition(taskId, toStatusId);
            }
            return;
        }

        // Обработка редактирования
        if (editButton) {
            const taskId = Number(editButton.getAttribute('data-id'));
            if (taskId) {
                openEditTaskModal(taskId);
            }
        }
    };

    if (myTasksList && !myTasksList.dataset.delegated) {
        myTasksList.addEventListener('click', handler);
        myTasksList.dataset.delegated = 'true';
    }

    if (projectTasksList && !projectTasksList.dataset.delegated) {
        projectTasksList.addEventListener('click', handler);
        projectTasksList.dataset.delegated = 'true';
    }
}

// Выносим логику перехода в отдельную функцию
async function handleStatusTransition(taskId: number, toStatusId: number, askConfirmation = true): Promise<boolean> {
    if (askConfirmation && !confirm('Перевести задачу в новый статус?')) return false;

    try {
        const res = await fetch(`/api/v1/tasks/${taskId}/status`, {
            method: 'PUT',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ status_id: toStatusId }),
        });

        // В функции handleStatusTransition после успешного ответа
        if (res.ok) {
            const updatedTask = await res.json();
            console.log('Задача обновлена:', updatedTask);

            // Обновляем задачу в общем списке
            const taskIndex = tasks.findIndex(t => t.id === taskId);
            if (taskIndex !== -1) {
                tasks[taskIndex].status_id = updatedTask.status_id;
                tasks[taskIndex].status = updatedTask.status; // если в ответе приходит полный объект статуса
            }

            // Обновляем в списке задач пользователя
            const filteredIndex = filteredTasks.findIndex(t => t.id === taskId);
            if (filteredIndex !== -1) {
                filteredTasks[filteredIndex].status_id = updatedTask.status_id;
                filteredTasks[filteredIndex].status = updatedTask.status;
            }

            // Обновляем в allProjectTasks
            const projectIndex = allProjectTasks.findIndex(t => t.id === taskId);
            if (projectIndex !== -1) {
                allProjectTasks[projectIndex].status_id = updatedTask.status_id;
                allProjectTasks[projectIndex].status = updatedTask.status;
            }

            applyFilters(); // Перерисовываем с актуальными данными
            return true;
        } else {
            const err = await res.json().catch(() => ({ error: 'Неизвестная ошибка' }));
            if (err.error !== 'invalid status transition') {
                alert(`Ошибка: ${err.error || 'Не удалось изменить статус'}`);
            }
            return false;
        }
    } catch (err) {
        console.error('Ошибка сети:', err);
        alert('Не удалось подключиться к серверу');
        return false;
    }
}

function showError(msg: string): void {
    const container = document.getElementById('tasksContainer');
    if (container) {
        container.innerHTML = `<div class="error">${escapeHtml(msg)}</div>`;
    }
}
// Функция для получения локальной даты и времени в формате ISO (без мс, с локальным часовым поясом)
function getLocalISOString(date: Date): string {
    const offset = date.getTimezoneOffset() * 60000; // смещение в миллисекундах
    const localDate = new Date(date.getTime() - offset);
    return localDate.toISOString().slice(0, 16);
}

function getTemplateProjectLabel(template: TaskTemplate): string {
    if (!template.project_id) {
        return 'Общий шаблон';
    }

    const projectName = template.project?.name || getProjectName(template.project_id);
    return `Проект: ${projectName}`;
}

function buildTaskTemplateOptions(templates: TaskTemplate[]): string {
    if (templates.length === 0) {
        return '<option value="">Шаблоны не найдены</option>';
    }

    return [
        '<option value="">Без шаблона</option>',
        ...templates.map(template => `
            <option value="${template.id}">
                ${escapeHtml(template.name)} (${escapeHtml(getTemplateProjectLabel(template))})
            </option>
        `),
    ].join('');
}

function applyTemplateToCreateForm(form: HTMLFormElement, template: TaskTemplate | undefined): void {
    if (!template) {
        return;
    }

    const nameInput = form.elements.namedItem('name') as HTMLInputElement | null;
    const descriptionInput = form.elements.namedItem('description') as HTMLTextAreaElement | null;
    const categoryInput = form.elements.namedItem('category') as HTMLInputElement | null;
    const priorityInput = form.elements.namedItem('priority') as HTMLSelectElement | null;
    const urgencyInput = form.elements.namedItem('urgency') as HTMLInputElement | null;
    const complexityInput = form.elements.namedItem('complexity') as HTMLInputElement | null;

    if (nameInput) nameInput.value = template.name || '';
    if (descriptionInput) descriptionInput.value = template.description || '';
    if (categoryInput) categoryInput.value = template.category || '';
    if (priorityInput) priorityInput.value = String(template.priority ?? 0);
    if (urgencyInput) urgencyInput.value = String(template.urgency ?? 1);
    if (complexityInput) complexityInput.value = String(template.complexity ?? 1);
}

async function refreshTemplateSelect(select: HTMLSelectElement, projectId?: number): Promise<void> {
    select.disabled = true;
    select.innerHTML = '<option value="">Загрузка шаблонов...</option>';

    const templates = await loadTaskTemplates(projectId);
    select.innerHTML = buildTaskTemplateOptions(templates);
    select.disabled = templates.length === 0;
}

async function saveTaskTemplateFromCreateForm(form: HTMLFormElement): Promise<void> {
    const raw = Object.fromEntries(new FormData(form));

    if (!raw.name) {
        alert('Сначала укажите название задачи');
        return;
    }

    const projectId = raw.project_id ? Number(raw.project_id) : 0;
    if (projectId && !canAssignToProject(projectId)) {
        alert('Вы не можете сохранять шаблон для этого проекта');
        return;
    }

    const templateData = {
        name: raw.name as string,
        description: raw.description as string || '',
        category: raw.category as string || '',
        priority: Number(raw.priority) || 0,
        urgency: raw.urgency ? parseFloat(raw.urgency as string) : 1,
        complexity: raw.complexity ? parseFloat(raw.complexity as string) : 1,
        project_id: projectId || null,
    };

    try {
        const res = await fetch('/api/v1/task-templates', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(templateData),
        });

        if (res.ok) {
            alert('Шаблон сохранён');
        } else {
            const err = await res.json().catch(() => ({ error: 'Неизвестная ошибка' }));
            alert(`Ошибка: ${err.error || 'Не удалось сохранить шаблон'}`);
        }
    } catch (err) {
        alert('Ошибка сети');
    }
}

// === Функция рендера модального окна создания задачи ===
function openCreateTaskModal(): void {
    if (!canCreateTask()) {
        alert('У вас нет прав на создание задач');
        return;
    }
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <button type="button" class="modal-close-x" aria-label="Закрыть">×</button>
            <h3>Создать задачу</h3>
            <form id="createTaskForm">
                <div class="form-group">
                    <label>Название *</label>
                    <input type="text" name="name" required />
                </div>
                <div class="form-group">
                    <label>Описание</label>
                    <textarea name="description"></textarea>
                </div>
                <div class="form-group">
                    <label>Категория</label>
                    <input type="text" name="category" />
                </div>
                <div class="form-group">
                    <label>Проект *</label>
                    <select name="project_id" required>
                        <option value="">Выберите проект</option>
                        ${(user!.role === 'admin' ? projects : projects.filter(p => p.responsible_id === user?.id))
        .map(p => `<option value="${p.id}">${escapeHtml(p.name)}</option>`)
        .join('')}
                    </select>
                </div>
                <div class="form-group template-select-group">
                    <label>Шаблон задачи</label>
                    <select name="task_template_id" class="task-template-select">
                        <option value="">Загрузка шаблонов...</option>
                    </select>
                    <div class="form-hint">Выбор шаблона заполнит название, описание, категорию, приоритет, важность и сложность.</div>
                    ${user?.role === 'admin' ? '<a class="template-page-link" href="/admin/task-templates">Открыть список шаблонов</a>' : ''}
                </div>
                <div class="form-group">
                    <label>Исполнитель *</label>
                    <select name="executor_id" required>
                        <option value="">Выберите пользователя</option>
                        ${users.map(u => `<option value="${u.id}">${escapeHtml(u.name)} (${escapeHtml(u.email)})</option>`).join('')}
                    </select>
                </div>
                <div class="form-group">
                    <label>Время начала</label>
                    <input type="datetime-local" name="start_time" value="${getLocalISOString(new Date())}"/>
                </div>
                <div class="form-group">
                    <label>Крайний срок</label>
                    <input type="datetime-local" name="end_time" />
                </div>
                <div class="form-group">
                    <label>Приоритет</label>
                    <select name="priority">
                        <option value="0">Низкий</option>
                        <option value="1">Средний</option>
                        <option value="2">Высокий</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>Важность (urgency)</label>
                    <input type="number" name="urgency" step="0.1" min="0" placeholder="0.0 - 1.0" value="1.0"/>
                </div>
                <div class="form-group">
                    <label>Сложность (complexity)</label>
                    <input type="number" name="complexity" step="0.1" min="0" placeholder="0.0 - 1.0" value="1.0" />
                </div>
                <div class="form-actions">
                    <button type="button" class="secondary-action save-template-from-create">Сохранить как шаблон</button>
                    <button type="submit">Создать</button>
                </div>
            </form>
        </div>
    `;
    document.body.appendChild(modal);

    const form = modal.querySelector('#createTaskForm') as HTMLFormElement;
    const projectSelect = form.elements.namedItem('project_id') as HTMLSelectElement;
    const templateSelect = form.elements.namedItem('task_template_id') as HTMLSelectElement;

    void refreshTemplateSelect(templateSelect);

    projectSelect.addEventListener('change', async () => {
        const projectId = Number(projectSelect.value);
        await refreshTemplateSelect(templateSelect, projectId || undefined);
    });

    templateSelect.addEventListener('change', () => {
        const templateId = Number(templateSelect.value);
        const selectedTemplate = taskTemplates.find(template => template.id === templateId);
        applyTemplateToCreateForm(form, selectedTemplate);
    });

    modal.querySelector('.save-template-from-create')?.addEventListener('click', async () => {
        await saveTaskTemplateFromCreateForm(form);
    });
    modal.querySelector('.modal-close-x')?.addEventListener('click', () => modal.remove());

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(form);
        const raw = Object.fromEntries(formData);

        if (!raw.name || !raw.project_id || !raw.executor_id) {
            alert('Заполните обязательные поля');
            return;
        }

        const projectId = Number(raw.project_id);
        if (!canAssignToProject(projectId)) {
            alert('Вы не можете назначать задачи по этому проекту');
            return;
        }
        const start_time = raw.start_time ? formatInputToISO(raw.start_time as string) : undefined;
        const end_time = raw.end_time ? formatInputToISO(raw.end_time as string) : undefined;

        const taskData: TaskFormData = {
            name: raw.name as string,
            description: raw.description as string || undefined,
            category: raw.category as string || undefined,
            project_id: projectId,
            executor_id: Number(raw.executor_id),
            status_id: defaultStatusId,
            priority: Number(raw.priority) || 0,
            start_time: start_time!,
            end_time: end_time!,
            urgency: raw.urgency ? parseFloat(raw.urgency as string) : undefined,
            complexity: raw.complexity ? parseFloat(raw.complexity as string) : undefined
        };

        try {
            const res = await fetch('/api/v1/tasks', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(taskData),
            });

            if (res.ok) {
                modal.remove();
                await loadProjects();
                await loadTasks();
            } else {
                const err = await res.json();
                alert(`Ошибка: ${err.error || 'Неизвестная ошибка'}`);
            }
        } catch (err) {
            alert('Ошибка сети');
        }
    });

}

// === Функция рендера модального окна редактирования задачи ===
function openEditTaskModal(taskId: number): void {
    const task = getTaskById(taskId);
    if (!task) {
        return;
    }

    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <button type="button" class="modal-close-x" aria-label="Закрыть">×</button>
            <h3>Редактировать задачу</h3>
            <form id="editTaskForm">
                <input type="hidden" name="id" value="${task.id}" />
                <div class="form-group">
                    <label>Название *</label>
                    <input type="text" name="name" value="${escapeHtml(task.name)}" required />
                </div>
                <div class="form-group">
                    <label>Описание</label>
                    <textarea name="description">${escapeHtml(task.description || '')}</textarea>
                </div>
                <div class="form-group">
                    <label>Категория</label>
                    <input type="text" name="category" value="${escapeHtml(task.category || '')}" />
                </div>
                <div class="form-group">
                    <label>Проект *</label>
                    <select name="project_id" required>
                        ${projects.map(p => `<option value="${p.id}" ${p.id === task.project_id ? 'selected' : ''}>${escapeHtml(p.name)}</option>`).join('')}
                    </select>
                </div>
                <div class="form-group">
                    <label>Исполнитель *</label>
                    <select name="executor_id" required>
                        ${users.map(u => `<option value="${u.id}" ${u.id === task.executor_id ? 'selected' : ''}>${escapeHtml(u.name)} (${escapeHtml(u.email)})</option>`).join('')}
                    </select>
                </div>
                <div class="form-group">
                    <label>Время начала</label>
                    <input type="datetime-local" name="start_time" value="${formatDateForInput(task.start_time)}" />
                </div>
                <div class="form-group">
                    <label>Крайний срок</label>
                    <input type="datetime-local" name="end_time" value="${formatDateForInput(task.end_time)}" />
                </div>
                <div class="form-group">
                    <label>Приоритет</label>
                    <select name="priority">
                        <option value="0" ${task.priority === 0 ? 'selected' : ''}>Низкий</option>
                        <option value="1" ${task.priority === 1 ? 'selected' : ''}>Средний</option>
                        <option value="2" ${task.priority === 2 ? 'selected' : ''}>Высокий</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>Важность (urgency)</label>
                    <input type="number" name="urgency" step="0.1" min="0" placeholder="0.0 - 1.0" value="${task.urgency || 1.0}" />
                </div>
                <div class="form-group">
                    <label>Сложность (complexity)</label>
                    <input type="number" name="complexity" step="0.1" min="0" placeholder="0.0 - 1.0" value="${task.complexity || 1.0}"/>
                </div>
                <div class="form-actions">
                    <button type="submit">Сохранить</button>
                </div>
            </form>
        </div>
    `;
    document.body.appendChild(modal);

    const form = modal.querySelector('#editTaskForm') as HTMLFormElement;
    modal.querySelector('.modal-close-x')?.addEventListener('click', () => modal.remove());
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(form);
        const raw = Object.fromEntries(formData);

        // Преобразуем end_time: string | null -> string | undefined
        const end_time = raw.end_time ? formatInputToISO(raw.end_time as string) : undefined;
        const start_time = raw.start_time ? formatInputToISO(raw.start_time as string) : undefined;
        const taskData = {
            id: Number(raw.id),
            name: raw.name as string,
            description: raw.description as string || null,
            category: raw.category as string || null,
            project_id: Number(raw.project_id),
            executor_id: Number(raw.executor_id),
            priority: Number(raw.priority),
            start_time: start_time,
            end_time: end_time,
            urgency: raw.urgency ? parseFloat(raw.urgency as string) : undefined,
            complexity: raw.complexity ? parseFloat(raw.complexity as string) : undefined
        };

        try {
            const res = await fetch(`/api/v1/tasks/${taskData.id}`, {
                method: 'PUT',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(taskData),
            });

            if (res.ok) {
                modal.remove();
                await loadTasks();
            } else {
                const err = await res.json();
                alert(`Ошибка: ${err.error || 'Не удалось сохранить'}`);
            }
        } catch (err) {
            alert('Ошибка сети');
        }
    });
}
// === Вспомогательные функции ===
function formatDateForInput(dateTimeString: string | null): string {
    if (!dateTimeString) return '';

    try {
        // Преобразуем ISO строку в локальное время
        const date = new Date(dateTimeString);
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        const hours = String(date.getHours()).padStart(2, '0');
        const minutes = String(date.getMinutes()).padStart(2, '0');

        return `${year}-${month}-${day}T${hours}:${minutes}`;
    } catch (e) {
        console.error('Ошибка парсинга даты:', dateTimeString, e);
        return '';
    }
}

function formatInputToISO(localDateTime: string): string | null {
    if (!localDateTime) return null;

    // Браузер возвращает локальное время, конвертируем в ISO
    const date = new Date(localDateTime);
    return date.toISOString(); // Уже в формате, который понимает PostgreSQL
}
// === Инициализация ===
document.addEventListener('DOMContentLoaded', async () => {
    token = localStorage.getItem('token');
    if (!token) return logout();

    await Promise.all([
        loadUsers(),
        loadCurrentUser()

    ]);
    await loadProjects();
    await Promise.all([
        loadTasks(),
        loadStatuses(),
        loadTransitions()
    ]);
    renderPage();
    applyFilters(); // Первичная фильтрация
});

// === Выход ===
document.getElementById('logoutBtn')?.addEventListener('click', logout);
