const token = localStorage.getItem("token");
const API_BASE = "/api/v1";
let suggestView = null;
let projectsState = [];
let stagesState = [];
let usersState = [];
let statusesState = [];
let statusTransitionsState = [];
let organizationsState = [];
let selectedProjectId = null;
if (!token) {
    window.location.href = "/login.html";
}
// === Инициализация Yandex Maps API ===
function initYandexMaps() {
    ymaps.ready(() => {
        console.log('Yandex Maps API загружена');
        initAddressAutocomplete();
    });
}
function initAddressAutocomplete() {
    const addressInput = document.getElementById('projectAddress');
    if (!addressInput)
        return;
    // Контейнер для подсказок
    const suggestionsContainer = document.createElement('div');
    suggestionsContainer.id = 'address-suggestions';
    suggestionsContainer.style.position = 'absolute';
    suggestionsContainer.style.maxHeight = '200px';
    suggestionsContainer.style.overflowY = 'auto';
    suggestionsContainer.style.backgroundColor = '#2d2d2d';
    suggestionsContainer.style.border = '1px solid #444';
    suggestionsContainer.style.borderTop = 'none';
    suggestionsContainer.style.borderRadius = '0 0 6px 6px';
    suggestionsContainer.style.zIndex = '1000';
    suggestionsContainer.style.display = 'none';
    document.body.appendChild(suggestionsContainer);
    const updateSuggestionsPosition = () => {
        const rect = addressInput.getBoundingClientRect();
        suggestionsContainer.style.top = `${window.scrollY + rect.bottom}px`;
        suggestionsContainer.style.left = `${window.scrollX + rect.left}px`;
        suggestionsContainer.style.width = `${rect.width}px`;
    };
    updateSuggestionsPosition();
    window.addEventListener('resize', updateSuggestionsPosition);
    window.addEventListener('scroll', updateSuggestionsPosition, true);
    let timeoutId = null;
    addressInput.addEventListener('input', () => {
        updateSuggestionsPosition();
        const value = addressInput.value.trim();
        if (value.length < 3) {
            suggestionsContainer.style.display = 'none';
            return;
        }
        if (timeoutId) {
            window.clearTimeout(timeoutId);
        }
        timeoutId = window.setTimeout(() => {
            fetchSuggestions(value, suggestionsContainer);
        }, 300);
    });
    // Скрываем при потере фокуса
    addressInput.addEventListener('focus', updateSuggestionsPosition);
    addressInput.addEventListener('blur', () => {
        setTimeout(() => {
            suggestionsContainer.style.display = 'none';
        }, 200);
    });
    // Обработка выбора
    suggestionsContainer.addEventListener('click', (e) => {
        const target = e.target;
        if (target.classList.contains('suggestion-item')) {
            const address = target.textContent || '';
            addressInput.value = address;
            suggestionsContainer.style.display = 'none';
            // Получаем координаты после выбора
            fetchCoordinatesFromYandex(address, addressInput);
        }
    });
}
// === Получение подсказок через HTTP Geocoder API ===
async function fetchSuggestions(input, container) {
    try {
        // Кодируем параметры отдельно
        const params = new URLSearchParams({
            format: 'json',
            results: '5',
            lang: 'ru_RU',
            geocode: input,
            boundedBy: '30.0,50.0~180.0,80.0',
            apikey: '15277c56-2dcb-4e64-aefe-2394e6f89c1d' // Добавляем ключ как параметр
        });
        const response = await fetch(`https://geocode-maps.yandex.ru/1.x/?${params.toString()}`, {
            method: 'GET',
            headers: {
                'Accept': 'application/json'
            }
        });
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        const data = await response.json();
        // Проверяем наличие данных
        if (!data.response || !data.response.GeoObjectCollection) {
            console.warn('Некорректный ответ от Яндекс.Геокодера:', data);
            container.innerHTML = '<div class="suggestion-item" style="padding: 8px; color: #888;">Нет данных</div>';
            container.style.display = 'block';
            return;
        }
        const suggestions = data.response.GeoObjectCollection.featureMember;
        if (suggestions.length === 0) {
            container.innerHTML = '<div class="suggestion-item" style="padding: 8px; color: #888;">Ничего не найдено</div>';
            container.style.display = 'block';
            return;
        }
        container.innerHTML = suggestions.map((item) => {
            const address = item.GeoObject.metaDataProperty.GeocoderMetaData.Address.formatted;
            return `<div class="suggestion-item" style="padding: 8px 12px; cursor: pointer; border-bottom: 1px solid #444; color: #e0e0e0;">${address}</div>`;
        }).join('');
        container.style.display = 'block';
    }
    catch (err) {
        console.error('Ошибка при получении подсказок:', err);
        container.innerHTML = '<div class="suggestion-item" style="padding: 8px; color: #888;">Ошибка сети</div>';
        container.style.display = 'block';
    }
}
async function fetchCoordinatesFromYandex(address, input) {
    try {
        const response = await ymaps.geocode(address, {
            results: 1
        });
        const firstGeoObject = response.geoObjects.get(0);
        if (firstGeoObject) {
            const coords = firstGeoObject.geometry.getCoordinates(); // [longitude, latitude]
            const latitude = coords[1];
            const longitude = coords[0];
            // Сохраняем координаты в data-атрибуты
            input.setAttribute('data-latitude', latitude.toString());
            input.setAttribute('data-longitude', longitude.toString());
            console.log('Координаты найдены:', { latitude, longitude });
        }
        else {
            console.warn('Не удалось получить координаты для адреса:', address);
            input.removeAttribute('data-latitude');
            input.removeAttribute('data-longitude');
        }
    }
    catch (err) {
        console.error('Ошибка при получении координат:', err);
        input.removeAttribute('data-latitude');
        input.removeAttribute('data-longitude');
    }
}
// === DOM Elements ===
const content = document.getElementById("content");
if (!content) {
    throw new Error("Не найден элемент #content");
}
// === Вспомогательные функции ===
function logout() {
    localStorage.removeItem("token");
    window.location.href = "/login.html";
}
function escapeHtml(text) {
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
}
function formatDate(value) {
    if (!value)
        return "—";
    return new Date(value).toLocaleDateString("ru-RU");
}
function formatDateTime(value) {
    if (!value)
        return "—";
    return new Date(value).toLocaleString("ru-RU", {
        day: "2-digit",
        month: "2-digit",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
    });
}
function getPriorityLabel(priority) {
    if (priority === 2)
        return "Высокий";
    if (priority === 1)
        return "Средний";
    return "Низкий";
}
function toDateInputValue(value) {
    return value ? value.substring(0, 10) : "";
}
function parseDateOnly(value) {
    if (!value)
        return null;
    const date = new Date(`${value.substring(0, 10)}T00:00:00`);
    return Number.isNaN(date.getTime()) ? null : date;
}
function addDays(date, days) {
    const next = new Date(date);
    next.setDate(next.getDate() + days);
    return next;
}
function daysBetween(start, end) {
    return Math.max(0, signedDaysBetween(start, end));
}
function signedDaysBetween(start, end) {
    const msPerDay = 24 * 60 * 60 * 1000;
    return Math.round((end.getTime() - start.getTime()) / msPerDay);
}
function clamp(value, min, max) {
    return Math.min(max, Math.max(min, value));
}
function getLocalDateTimeValue(date = new Date()) {
    const offset = date.getTimezoneOffset();
    const local = new Date(date.getTime() - offset * 60000);
    return local.toISOString().slice(0, 16);
}
function dateTimeInputToISO(value) {
    return value ? new Date(value).toISOString() : undefined;
}
function dateTimeToInputValue(value) {
    if (!value)
        return "";
    const date = new Date(value);
    const offset = date.getTimezoneOffset();
    const local = new Date(date.getTime() - offset * 60000);
    return local.toISOString().slice(0, 16);
}
function dateInputToISO(value) {
    if (!value || typeof value !== "string")
        return null;
    return value ? new Date(`${value}T00:00:00`).toISOString() : null;
}
function getDefaultStatusId() {
    return statusesState.find(status => status.is_default)?.id || statusesState[0]?.id || 0;
}
function getUserName(id) {
    return usersState.find(user => user.id === id)?.name || `Пользователь #${id}`;
}
function getProjectAccessUserIDs(project) {
    return (project.access_users || []).map(item => item.user_id);
}
function getProjectMemberUserIDs(project) {
    return (project.members || []).map(item => item.user_id);
}
function renderProjectAccessPicker(selectedUserIDs = []) {
    const selected = new Set(selectedUserIDs);
    return `
        <div class="project-access-list">
            ${usersState.map(user => `
                <label class="project-access-option">
                    <input type="checkbox" name="access_user_ids" value="${user.id}" ${selected.has(user.id) ? "checked" : ""} />
                    <span>${escapeHtml(user.name)}${user.email ? ` (${escapeHtml(user.email)})` : ""}</span>
                </label>
            `).join("")}
        </div>
    `;
}
function renderProjectMemberPicker(project) {
    const selected = new Set(getProjectMemberUserIDs(project));
    return `
        <div class="project-access-list">
            ${usersState
        .filter(user => user.id !== project.responsible_id)
        .map(user => `
                <label class="project-access-option">
                    <input type="checkbox" name="project_member_user_ids" value="${user.id}" ${selected.has(user.id) ? "checked" : ""} />
                    <span>${escapeHtml(user.name)}${user.email ? ` (${escapeHtml(user.email)})` : ""}</span>
                </label>
            `).join("")}
        </div>
    `;
}
function getCheckedProjectAccessUserIDs(root) {
    return Array.from(root.querySelectorAll('input[name="access_user_ids"]:checked'))
        .map(input => Number(input.value))
        .filter(Boolean);
}
function getCheckedProjectMemberUserIDs(root) {
    return Array.from(root.querySelectorAll('input[name="project_member_user_ids"]:checked'))
        .map(input => Number(input.value))
        .filter(Boolean);
}
async function syncProjectMembers(projectId, existingMembers, selectedUserIDs, responsibleID) {
    const selected = new Set(selectedUserIDs.filter(userID => userID && userID !== responsibleID));
    const existingByUserID = new Map(existingMembers.map(member => [member.user_id, member]));
    await Promise.all(existingMembers
        .filter(member => member.user_id !== responsibleID && !selected.has(member.user_id))
        .map(member => apiFetch(`/project-members/${member.id}`, { method: "DELETE" })));
    await Promise.all(Array.from(selected)
        .filter(userID => !existingByUserID.has(userID))
        .map(userID => apiFetch("/project-members", {
        method: "POST",
        body: JSON.stringify({ project_id: projectId, user_id: userID, role: "member" }),
    })));
}
function setupHiddenProjectControls(root) {
    const hiddenInput = root.querySelector('[data-project-hidden]');
    const accessBlock = root.querySelector('[data-project-access-block]');
    const sync = () => {
        if (accessBlock && hiddenInput) {
            accessBlock.hidden = !hiddenInput.checked;
        }
    };
    hiddenInput?.addEventListener("change", sync);
    sync();
}
async function apiFetch(path, options = {}) {
    const headers = new Headers(options.headers || {});
    headers.set("Authorization", `Bearer ${token}`);
    if (options.body && !headers.has("Content-Type")) {
        headers.set("Content-Type", "application/json");
    }
    const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
    if (res.status === 401) {
        logout();
    }
    return res;
}
// === Генерация разметки ===
function renderPage() {
    const canManageProjects = checkUserRole("admin") || checkUserRole("manager");
    content.innerHTML = `
        
        <link rel="stylesheet" href="/static/css/app.css">
        <link rel="stylesheet" href="/static/css/admin-projects.css">
        <div class="projects-toolbar"> 
            ${canManageProjects ? `<button id="createProjectBtn" class="btn btn-primary">+ Добавить проект</button>` : ""}
            ${canManageProjects ? `<button id="exportProjectsBtn" class="btn btn-excel export-btn">Скачать Excel</button>` : ""}
        </div>
        <ul id="projectList"></ul>

        <!-- Модальное окно -->
        <div id="projectModal" class="modal" style="display: none;">
            <div class="modal-content">
                <button type="button" class="modal-close-x" aria-label="Закрыть" onclick="closeModal()">×</button>
                <h3 id="modalTitle">Добавить проект</h3>
                <form id="projectForm">
                    <input type="hidden" id="projectID" />
                    
                    <div class="form-group">
                        <label for="projectName">Название</label>
                        <input type="text" id="projectName" required />
                    </div>
                    
                    <div class="form-group">
                        <label for="projectDescription">Описание</label>
                        <textarea id="projectDescription"></textarea>
                    </div>

                    <div class="form-group">
                        <label for="projectAddress">Адрес</label>
                        <input type="text" id="projectAddress" />
                    </div>

                    <div class="form-row">
                        <div class="form-group">
                            <label for="projectStartDate">Дата начала</label>
                            <input type="date" id="projectStartDate" />
                        </div>
                        <div class="form-group">
                            <label for="projectDeadlineDate">Дата конца</label>
                            <input type="date" id="projectDeadlineDate" />
                        </div>
                    </div>

                    <div class="form-row">
                        <div class="form-group">
                            <label for="projectCost">Стоимость</label>
                            <input type="number" step="0.01" id="projectCost" />
                        </div>
                    </div>

                    <div class="form-group">
                        <label for="projectComments">Комментарии</label>
                        <textarea id="projectComments"></textarea>
                    </div>
                    
                    <div class="form-group">
                        <label for="projectStage">Стадия</label>
                        <select id="projectStage" required>
                            <option value="">Выберите стадию...</option>
                        </select>
                    </div>
                    
                    <div class="form-group">
                        <label for="projectResponsible">Ответственный</label>
                        <select id="projectResponsible" required>
                            <option value="">Выберите пользователя...</option>
                        </select>
                    </div>

                    ${checkUserRole("admin") ? `
                        <div class="form-group">
                            <label class="project-hidden-toggle">
                                <input type="checkbox" id="projectIsHidden" data-project-hidden />
                                <span>Скрытый проект</span>
                            </label>
                        </div>
                        <div class="form-group" data-project-access-block hidden>
                            <label>Пользователи с доступом</label>
                            <div id="projectAccessUsers"></div>
                        </div>
                    ` : ""}
                    
                    <div class="form-actions">
                        <button type="submit">Сохранить</button>
                    </div>
                </form>
            </div>
        </div>
    `;
    // Привязка событий
    const logoutBtn = document.getElementById("logoutBtn");
    const createProjectBtn = document.getElementById("createProjectBtn");
    const exportProjectsBtn = document.getElementById("exportProjectsBtn");
    const form = document.getElementById("projectForm");
    if (logoutBtn)
        logoutBtn.addEventListener("click", logout);
    if (createProjectBtn)
        createProjectBtn.addEventListener("click", () => openModal());
    if (exportProjectsBtn instanceof HTMLButtonElement) {
        exportProjectsBtn.addEventListener("click", () => {
            window.downloadExcelExport?.('/api/v1/export/projects', 'projects.xlsx', exportProjectsBtn);
        });
    }
    if (form)
        form.addEventListener("submit", handleSubmit);
    if (checkUserRole("admin")) {
        setupHiddenProjectControls(document);
    }
}
// === Загрузка данных ===
async function loadProjects() {
    try {
        const res = await apiFetch("/projects");
        if (res.status === 401)
            return logout();
        const projects = await res.json();
        projectsState = projects;
        renderProjectList(projects);
    }
    catch (err) {
        console.error("Ошибка загрузки проектов:", err);
        content.innerHTML += `<p>Не удалось загрузить проекты.</p>`;
    }
}
async function loadStages() {
    const select = document.getElementById("projectStage");
    try {
        const res = await apiFetch("/stages");
        const stages = await res.json();
        stagesState = stages;
        if (select) {
            select.innerHTML = `<option value="">Выберите стадию...</option>`;
            stages.forEach((stage) => {
                const option = document.createElement("option");
                option.value = stage.id.toString();
                option.textContent = stage.name;
                select.appendChild(option);
            });
        }
    }
    catch (err) {
        console.error("Ошибка загрузки стадий:", err);
    }
}
async function loadUsers() {
    const select = document.getElementById("projectResponsible");
    try {
        const res = await apiFetch("/users");
        const users = await res.json();
        usersState = users;
        if (select) {
            select.innerHTML = `<option value="">Выберите пользователя...</option>`;
            users.forEach((user) => {
                const option = document.createElement("option");
                option.value = user.id.toString();
                option.textContent = user.name;
                select.appendChild(option);
            });
        }
        const accessUsers = document.getElementById("projectAccessUsers");
        if (accessUsers) {
            accessUsers.innerHTML = renderProjectAccessPicker();
        }
    }
    catch (err) {
        console.error("Ошибка загрузки пользователей:", err);
    }
}
async function loadStatuses() {
    try {
        const res = await apiFetch("/statuses");
        statusesState = await res.json();
    }
    catch (err) {
        console.error("Ошибка загрузки статусов:", err);
    }
}
async function loadStatusTransitions() {
    try {
        const res = await apiFetch("/status-transitions");
        statusTransitionsState = res.ok ? await res.json() : [];
    }
    catch (err) {
        console.error("Ошибка загрузки переходов статусов:", err);
        statusTransitionsState = [];
    }
}
async function loadOrganizations() {
    try {
        const res = await apiFetch("/organizations");
        organizationsState = res.ok ? await res.json() : [];
    }
    catch (err) {
        console.error("Ошибка загрузки организаций:", err);
        organizationsState = [];
    }
}
// === Рендеринг ===
function renderProjectList(projects) {
    const list = document.getElementById("projectList");
    if (!list)
        return;
    list.innerHTML = "";
    const isUserAdmin = checkUserRole("admin");
    const currentUserId = getCurrentUserId();
    projects.forEach((project) => {
        const canEdit = isUserAdmin || project.responsible_id === currentUserId;
        const canDelete = isUserAdmin;
        const li = document.createElement("li");
        li.className = "project-item";
        li.dataset.projectId = project.id.toString();
        li.innerHTML = `
            <button type="button" class="project-details project-open-btn" data-project-id="${project.id}">
                <h3>
                    ${escapeHtml(project.name)}
                    ${project.is_hidden ? `<span class="project-hidden-badge">Скрытый</span>` : ""}
                </h3>
                <div class="project-info">
                    Ответственный: ${escapeHtml(project.responsible?.name || "—")}
                </div>
                <div class="project-info">
                    Стадия: ${escapeHtml(project.stage?.name || "—")}
                </div>
                ${project.address
            ? `<div class="project-info">Адрес: ${escapeHtml(project.address)}</div>`
            : ""}
                ${project.cost != null
            ? `<div class="project-info">Стоимость: ${project.cost}</div>`
            : ""}
                <div class="project-info">
                    ${escapeHtml(project.description || "")}
                </div>
            </button>
            <div class="project-actions">
                <!-- Кнопка трудозатрат — доступна всем у кого есть доступ к проекту -->
                <a href="/labor?project_id=${project.id}" class="btn btn-primary">Трудозатраты</a>
                ${canEdit ? `<button class="btn btn-neutral" onclick="editProject(${project.id})">Редактировать</button>` : ""}
                ${canDelete ? `<button class="btn btn-danger"  onclick="deleteProject(${project.id})">Удалить</button>` : ""}
            </div>
        `;
        list.appendChild(li);
    });
    list.querySelectorAll(".project-open-btn").forEach((button) => {
        button.addEventListener("click", () => {
            const projectId = Number(button.dataset.projectId);
            if (projectId) {
                openProjectDetail(projectId);
            }
        });
    });
}
async function openProjectDetail(projectId) {
    selectedProjectId = projectId;
    try {
        const [projectRes, phasesRes, membersRes, tasksRes, statusesRes, transitionsRes, messagesRes] = await Promise.all([
            apiFetch(`/projects/${projectId}`),
            apiFetch(`/project-phases?project_id=${projectId}`),
            apiFetch(`/project-members?project_id=${projectId}`),
            apiFetch(`/tasks/by-project?project_id=${projectId}`),
            apiFetch("/statuses"),
            apiFetch("/status-transitions"),
            apiFetch(`/projects/${projectId}/messages`),
        ]);
        if (!projectRes.ok) {
            alert("Не удалось загрузить проект");
            return;
        }
        const project = await projectRes.json();
        project.members = membersRes.ok ? await membersRes.json() : [];
        const phases = sortProjectPhases(phasesRes.ok ? await phasesRes.json() : []);
        const tasks = tasksRes.ok ? await tasksRes.json() : [];
        statusesState = statusesRes.ok ? await statusesRes.json() : statusesState;
        statusTransitionsState = transitionsRes.ok ? await transitionsRes.json() : statusTransitionsState;
        const messages = messagesRes.ok ? await messagesRes.json() : [];
        renderProjectDetail(project, phases, tasks, messages);
    }
    catch (err) {
        console.error("Ошибка загрузки карточки проекта:", err);
        alert("Не удалось открыть проект");
    }
}
async function refreshProjectDetailPreservingScroll(projectId) {
    const scrollX = window.scrollX;
    const scrollY = window.scrollY;
    await openProjectDetail(projectId);
    window.scrollTo({ top: scrollY, left: scrollX, behavior: "auto" });
    requestAnimationFrame(() => {
        window.scrollTo({ top: scrollY, left: scrollX, behavior: "auto" });
    });
    window.setTimeout(() => {
        window.scrollTo({ top: scrollY, left: scrollX, behavior: "auto" });
    }, 50);
}
async function refreshProjectChat(projectId, anchor = "scroll") {
    const chat = document.querySelector(".project-chat");
    const scrollX = window.scrollX;
    const scrollY = window.scrollY;
    const formTop = document.getElementById("projectChatForm")?.getBoundingClientRect().top ?? null;
    const res = await apiFetch(`/projects/${projectId}/messages`);
    if (!res.ok || !chat) {
        await refreshProjectDetailPreservingScroll(projectId);
        return;
    }
    const messages = await res.json();
    chat.outerHTML = renderProjectChat(messages);
    setupProjectChat(projectId, messages);
    const restoreFormPosition = () => {
        if (anchor === "scroll") {
            window.scrollTo({ top: scrollY, left: scrollX, behavior: "auto" });
            return;
        }
        if (formTop == null) {
            window.scrollTo({ top: scrollY, left: scrollX, behavior: "auto" });
            return;
        }
        const nextTop = document.getElementById("projectChatForm")?.getBoundingClientRect().top;
        if (nextTop == null)
            return;
        window.scrollBy({ top: nextTop - formTop, left: 0, behavior: "auto" });
    };
    restoreFormPosition();
    requestAnimationFrame(restoreFormPosition);
    window.setTimeout(restoreFormPosition, 50);
}
function renderProjectDetail(project, phases, tasks, messages) {
    const canEdit = checkUserRole("admin") || project.responsible_id === getCurrentUserId();
    content.innerHTML = `
        <link rel="stylesheet" href="/static/css/app.css">
        <link rel="stylesheet" href="/static/css/tasks.css">
        <link rel="stylesheet" href="/static/css/admin-projects.css">
        <div class="project-detail">
            <div class="project-detail-header">
                <div class="project-detail-title">
                    <button id="backToProjectsBtn" class="project-back-btn" type="button">
                        <span class="material-icons" aria-hidden="true">arrow_back</span>
                        Проекты
                    </button>
                    <h2>
                        ${escapeHtml(project.name)}
                        ${project.is_hidden ? `<span class="project-hidden-badge">Скрытый</span>` : ""}
                    </h2>
                    <div class="project-detail-meta">
                        <span>Ответственный: ${escapeHtml(project.responsible?.name || getUserName(project.responsible_id))}</span>
                        <span>Стадия: ${escapeHtml(project.stage?.name || "—")}</span>
                        ${project.deadline_date ? `<span>Дата конца: ${formatDate(project.deadline_date)}</span>` : ""}
                    </div>
                </div>
                <div class="project-detail-actions">
                    ${canEdit ? `<button id="changeProjectStageBtn" class="btn btn-neutral" type="button">Стадия</button>` : ""}
                    <div class="create-menu">
                        <button id="projectCreateMenuBtn" class="btn btn-primary" type="button">+ Создать</button>
                        <div id="projectCreateMenu" class="create-menu-list" hidden>
                            <button type="button" data-create="task">Задачу</button>
                            <button type="button" data-create="phase">Этап</button>
                        </div>
                    </div>
                </div>
            </div>

            <div class="project-tabs" role="tablist" aria-label="Разделы проекта">
                <button class="project-tab active" type="button" data-project-tab="general">Общие</button>
                <button class="project-tab" type="button" data-project-tab="tasks">Задачи</button>
                <button class="project-tab" type="button" data-project-tab="gantt">Гант</button>
                <button class="project-tab" type="button" data-project-tab="team">Команда</button>
            </div>

            <section class="project-tab-panel active" data-project-panel="general">
                ${renderProjectGeneralTab(project, phases, canEdit)}
            </section>

            <section class="project-tab-panel" data-project-panel="tasks" hidden>
                ${renderProjectTasksTab(phases, tasks)}
            </section>

            <section class="project-tab-panel" data-project-panel="gantt" hidden>
                ${renderProjectGanttTab(project, phases, tasks)}
            </section>

            <section class="project-tab-panel" data-project-panel="team" hidden>
                ${renderProjectTeamTab(project)}
            </section>

            ${renderProjectChat(messages)}
        </div>
    `;
    document.getElementById("backToProjectsBtn")?.addEventListener("click", async () => {
        selectedProjectId = null;
        renderPage();
        await Promise.all([loadProjects(), loadStages(), loadUsers(), loadStatuses(), loadStatusTransitions(), loadOrganizations()]);
        loadYandexMapsAPI();
    });
    document.getElementById("changeProjectStageBtn")?.addEventListener("click", () => openStageModal(project));
    document.getElementById("projectCreateMenuBtn")?.addEventListener("click", () => {
        const menu = document.getElementById("projectCreateMenu");
        if (menu)
            menu.hidden = !menu.hidden;
    });
    document.getElementById("projectCreateMenu")?.addEventListener("click", (event) => {
        const target = event.target;
        const action = target.dataset.create;
        const menu = document.getElementById("projectCreateMenu");
        if (menu)
            menu.hidden = true;
        if (action === "task") {
            openTaskModal(project, phases);
        }
        if (action === "phase") {
            openPhaseModal(project);
        }
    });
    document.querySelectorAll(".project-section-heading [data-create='phase']").forEach((button) => {
        button.addEventListener("click", () => openPhaseModal(project));
    });
    document.querySelectorAll("[data-open-phase]").forEach((button) => {
        button.addEventListener("click", () => {
            const phaseId = Number(button.getAttribute("data-open-phase"));
            if (phaseId) {
                openPhaseCard(project, phases, phaseId);
            }
        });
    });
    document.querySelectorAll("[data-open-task]").forEach((button) => {
        button.addEventListener("click", (event) => {
            const target = event.target;
            const current = event.currentTarget;
            if (target.closest("[data-delete-task]") || (current.classList.contains("project-task-card") && target.closest(".task-actions")))
                return;
            event.stopPropagation();
            const taskId = Number(current.getAttribute("data-open-task"));
            const task = tasks.find(item => item.id === taskId);
            if (task)
                openTaskCard(task, phases);
        });
    });
    document.querySelectorAll("[data-delete-task]").forEach((button) => {
        button.addEventListener("click", async () => {
            const taskId = Number(button.getAttribute("data-delete-task"));
            if (taskId) {
                await deleteProjectTask(project.id, taskId);
            }
        });
    });
    document.querySelectorAll("[data-transition-task]").forEach((button) => {
        button.addEventListener("click", async (event) => {
            event.stopPropagation();
            const element = button;
            const taskId = Number(element.getAttribute("data-transition-task"));
            const statusId = Number(element.getAttribute("data-to-status"));
            if (taskId && statusId) {
                await updateProjectTaskStatus(project.id, taskId, statusId);
            }
        });
    });
    setupProjectTaskDragAndDrop(project.id);
    document.querySelector("[data-edit-project-details]")?.addEventListener("click", () => {
        if (canEdit)
            openProjectDetailsModal(project);
    });
    document.querySelector("[data-manage-client]")?.addEventListener("click", () => {
        if (canEdit)
            openClientModal(project);
    });
    document.querySelector("[data-add-client-contact]")?.addEventListener("click", (event) => {
        event.stopPropagation();
        if (canEdit)
            openClientContactModal(project);
    });
    document.querySelectorAll("[data-edit-project-details], [data-manage-client]").forEach((card) => {
        card.addEventListener("keydown", (event) => {
            const keyboardEvent = event;
            if (keyboardEvent.key === "Enter" || keyboardEvent.key === " ") {
                keyboardEvent.preventDefault();
                card.click();
            }
        });
    });
    document.querySelectorAll("[data-project-tab]").forEach((button) => {
        button.addEventListener("click", () => {
            const tab = button.dataset.projectTab;
            setProjectTab(tab || "general");
        });
    });
    setupProjectChat(project.id, messages);
}
function setProjectTab(tab) {
    document.querySelectorAll("[data-project-tab]").forEach((button) => {
        button.classList.toggle("active", button.dataset.projectTab === tab);
    });
    document.querySelectorAll("[data-project-panel]").forEach((panel) => {
        const isActive = panel.dataset.projectPanel === tab;
        panel.classList.toggle("active", isActive);
        panel.hidden = !isActive;
    });
}
function renderProjectGeneralTab(project, phases, canEdit) {
    const sortedPhases = sortProjectPhases(phases);
    return `
        <div class="project-info-grid">
            <div class="project-info-card ${canEdit ? "project-info-card-clickable" : ""}" ${canEdit ? `data-edit-project-details role="button" tabindex="0" aria-label="Редактировать сведения проекта"` : ""}>
                <h3>Сведения</h3>
                <dl class="project-details-list">
                    <div><dt>Ответственный</dt><dd>${escapeHtml(project.responsible?.name || getUserName(project.responsible_id))}</dd></div>
                    <div><dt>Начало</dt><dd>${formatDate(project.start_date)}</dd></div>
                    <div><dt>Дата конца</dt><dd>${formatDate(project.deadline_date)}</dd></div>
                    <div><dt>Стоимость</dt><dd>${project.cost != null ? project.cost : "—"}</dd></div>
                    <div><dt>Адрес</dt><dd>${escapeHtml(project.address || "—")}</dd></div>
                </dl>
                <div class="project-description-block">
                    <h3>Описание</h3>
                    <div class="project-detail-description">${escapeHtml(project.description || "Описание не заполнено")}</div>
                </div>
            </div>
            <div class="project-info-card ${canEdit ? "project-info-card-clickable" : ""}" ${canEdit ? `data-manage-client role="button" tabindex="0" aria-label="Управление клиентом проекта"` : ""}>
                <div class="project-card-heading">
                    <h3>Клиент</h3>
                </div>
                ${renderProjectClient(project)}
            </div>
        </div>
        <section class="project-section">
            <div class="project-section-heading">
                <h3>Этапы проекта</h3>
                ${canEdit ? `<button class="btn btn-neutral" type="button" data-create="phase">+ Этап</button>` : ""}
            </div>
            <div class="phase-list">
                ${sortedPhases.length > 0 ? sortedPhases.map(renderPhaseSummaryCard).join("") : `<div class="empty-state">Этапы ещё не созданы</div>`}
            </div>
        </section>
    `;
}
function sortProjectPhases(phases) {
    return [...phases].sort((a, b) => {
        const startA = a.start_date ? new Date(a.start_date).getTime() : Number.MAX_SAFE_INTEGER;
        const startB = b.start_date ? new Date(b.start_date).getTime() : Number.MAX_SAFE_INTEGER;
        if (startA !== startB)
            return startA - startB;
        const positionDiff = (a.position || 0) - (b.position || 0);
        if (positionDiff !== 0)
            return positionDiff;
        return a.id - b.id;
    });
}
function renderPhaseSummaryCard(phase, index) {
    return `
        <article class="phase-card phase-card-clickable" data-open-phase="${phase.id}" tabindex="0" role="button" aria-label="Открыть этап ${escapeHtml(phase.name)}">
            <div class="phase-card-header">
                <div class="phase-sequence">${index + 1}</div>
                <div>
                    <h4>${escapeHtml(phase.name)}</h4>
                    <div class="phase-meta">
                        <span>Ответственный: ${escapeHtml(phase.responsible?.name || getUserName(phase.responsible_id))}</span>
                        <span>${formatDate(phase.start_date)} - ${formatDate(phase.end_date)}</span>
                    </div>
                </div>
            </div>
            ${phase.description ? `<div class="phase-description">${escapeHtml(phase.description)}</div>` : ""}
        </article>
    `;
}
async function openPhaseCard(project, phases, phaseId) {
    try {
        const res = await apiFetch(`/project-phases/${phaseId}`);
        if (!res.ok) {
            alert("Не удалось открыть этап");
            return;
        }
        const phase = await res.json();
        if (usersState.length === 0)
            await loadUsers();
        const modal = createActionModal(`
            <h3>${escapeHtml(phase.name)}</h3>
            <div class="phase-detail-card">
                <form id="phaseEditForm" class="phase-description-form">
                    <div class="form-group">
                        <label>Название *</label>
                        <input type="text" name="name" value="${escapeHtml(phase.name)}" required />
                    </div>
                    <div class="form-group">
                        <label>Описание</label>
                        <textarea name="description" rows="5" placeholder="Добавьте описание этапа">${escapeHtml(phase.description || "")}</textarea>
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label>Дата начала</label>
                            <input type="date" name="start_date" value="${toDateInputValue(phase.start_date)}" />
                        </div>
                        <div class="form-group">
                            <label>Дата конца</label>
                            <input type="date" name="end_date" value="${toDateInputValue(phase.end_date)}" />
                        </div>
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label>Позиция</label>
                            <input type="number" name="position" min="0" value="${phase.position}" />
                        </div>
                        <div class="form-group">
                            <label>Ответственный *</label>
                            <select name="responsible_id" required>
                                <option value="">Выберите пользователя</option>
                                ${usersState.map(user => `<option value="${user.id}" ${user.id === phase.responsible_id ? "selected" : ""}>${escapeHtml(user.name)}${user.email ? ` (${escapeHtml(user.email)})` : ""}</option>`).join("")}
                            </select>
                        </div>
                    </div>
                    <div class="form-actions">
                        <button type="button" class="btn btn-neutral" data-add-task-for-phase="${phase.id}">Добавить задачу</button>
                        <button type="submit">Сохранить</button>
                    </div>
                </form>
                <div class="project-section">
                    <div class="project-section-heading">
                        <h3>Задачи этапа</h3>
                    </div>
                    <div class="task-list compact">
                        ${(phase.tasks || []).length > 0
            ? (phase.tasks || []).map(task => `
                                <div class="task-row">
                                    <div class="task-row-main">
                                        <div class="task-row-title">${escapeHtml(task.name)}</div>
                                        <div class="task-row-meta">
                                            <span>${escapeHtml(task.status?.name || "Без статуса")}</span>
                                            <span>Исполнитель: ${escapeHtml(task.executor?.name || getUserName(task.executor_id))}</span>
                                            ${task.end_time ? `<span>До ${formatDateTime(task.end_time)}</span>` : ""}
                                        </div>
                                    </div>
                                </div>
                            `).join("")
            : `<div class="empty-state">В этапе пока нет задач</div>`}
                    </div>
                </div>
            </div>
        `);
        modal.querySelector("[data-add-task-for-phase]")?.addEventListener("click", () => {
            modal.remove();
            openTaskModal(project, phases, phase.id);
        });
        const form = modal.querySelector("#phaseEditForm");
        form.addEventListener("submit", async (event) => {
            event.preventDefault();
            const raw = Object.fromEntries(new FormData(form));
            const payload = {
                project_id: project.id,
                name: String(raw.name || ""),
                description: String(raw.description || ""),
                start_date: dateInputToISO(raw.start_date),
                end_date: dateInputToISO(raw.end_date),
                position: Number(raw.position) || 0,
                responsible_id: Number(raw.responsible_id),
            };
            if (!payload.name || !payload.responsible_id) {
                alert("Заполните все обязательные поля");
                return;
            }
            const updateRes = await apiFetch(`/project-phases/${phase.id}`, {
                method: "PUT",
                body: JSON.stringify(payload),
            });
            if (updateRes.ok) {
                modal.remove();
                await openProjectDetail(project.id);
                setProjectTab("general");
                return;
            }
            const err = await updateRes.json().catch(() => ({ error: "Не удалось сохранить этап" }));
            alert(`Ошибка: ${err.error}`);
        });
    }
    catch (err) {
        alert("Ошибка сети");
    }
}
function getProjectClient(project) {
    return (project.participants || []).find(item => item.role === "client") || (project.participants || [])[0];
}
function renderProjectClient(project) {
    const client = getProjectClient(project);
    if (!client) {
        return `<div class="empty-state">Клиент не указан</div>`;
    }
    const org = client.organization;
    return `
        <div class="person-card-list">
            <div class="person-card">
                <div class="person-card-title">${escapeHtml(org?.name || `Компания #${client.organization_id}`)}</div>
                <div class="person-card-meta">
                    ${org?.phone ? `<span>${escapeHtml(org.phone)}</span>` : ""}
                    ${org?.email ? `<span>${escapeHtml(org.email)}</span>` : ""}
                    ${org?.website ? `<span>${escapeHtml(org.website)}</span>` : ""}
                </div>
                ${org?.address ? `<div class="person-card-note">${escapeHtml(org.address)}</div>` : ""}
                <div class="person-card-actions">
                    <button class="btn btn-neutral btn-sm" type="button" data-add-client-contact="${client.organization_id}">Добавить контакт</button>
                </div>
            </div>
        </div>
    `;
}
function renderProjectTasksTab(phases, tasks) {
    const phaseByID = new Map(phases.map(phase => [phase.id, phase.name]));
    return `
        <section class="project-section">
            <div class="project-section-heading">
                <h3>Задачи проекта</h3>
            </div>
            <div class="project-task-board" style="--project-task-status-count: ${Math.max(statusesState.length, 1)}">
                ${statusesState.length > 0 ? statusesState.map(status => renderProjectTaskColumn(status, tasks.filter(task => task.status_id === status.id), phaseByID)).join("") : renderProjectTaskList(tasks, phaseByID)}
            </div>
        </section>
    `;
}
function renderProjectGanttTab(project, phases, tasks) {
    const phaseRows = phases
        .map((phase) => {
        const start = parseDateOnly(phase.start_date);
        const end = parseDateOnly(phase.end_date);
        return start && end ? { phase, start, end: end < start ? start : end } : null;
    })
        .filter((row) => row !== null)
        .sort((a, b) => a.start.getTime() - b.start.getTime() || a.phase.position - b.phase.position);
    if (phases.length === 0) {
        return `<div class="empty-state">Этапы ещё не созданы</div>`;
    }
    if (phaseRows.length === 0) {
        return `<div class="empty-state">У этапов не заполнены даты начала и конца</div>`;
    }
    const projectStart = parseDateOnly(project.start_date);
    const projectEnd = parseDateOnly(project.deadline_date);
    const chartStart = projectStart || phaseRows[0].start;
    const chartEnd = new Date(Math.max(...phaseRows.map(row => row.end.getTime()), projectEnd?.getTime() ?? phaseRows[phaseRows.length - 1].end.getTime()));
    const totalDays = Math.max(daysBetween(chartStart, chartEnd), 1);
    const ticks = buildGanttTicks(chartStart, chartEnd, totalDays);
    const todayOffsetDays = signedDaysBetween(chartStart, new Date());
    const todayLeft = todayOffsetDays >= 0 && todayOffsetDays <= totalDays
        ? (todayOffsetDays / totalDays) * 100
        : null;
    return `
        <section class="project-section">
            <div class="project-section-heading">
                <h3>Диаграмма Ганта</h3>
                <div class="gantt-project-range">Старт проекта: ${formatDate(chartStart.toISOString())}</div>
            </div>
            <div class="gantt-chart" style="--gantt-days: ${totalDays};">
                <div class="gantt-scale">
                    <div class="gantt-scale-spacer"></div>
                    <div class="gantt-timeline">
                        ${todayLeft != null ? `<span class="gantt-today" style="left: ${todayLeft}%;">Сегодня</span>` : ""}
                        ${ticks.map(tick => `
                            <span class="gantt-tick" style="left: ${tick.left}%;">
                                ${escapeHtml(tick.label)}
                            </span>
                        `).join("")}
                    </div>
                </div>
                <div class="gantt-rows">
                    ${phaseRows.map(row => renderGanttPhaseRow(row.phase, row.start, row.end, chartStart, totalDays, tasks, todayLeft)).join("")}
                </div>
            </div>
        </section>
    `;
}
function buildGanttTicks(chartStart, chartEnd, totalDays) {
    const tickCount = Math.min(6, totalDays + 1);
    if (tickCount <= 1) {
        return [{ label: formatDate(chartStart.toISOString()), left: 0 }];
    }
    return Array.from({ length: tickCount }, (_, index) => {
        const offset = Math.round((totalDays * index) / (tickCount - 1));
        const date = index === tickCount - 1 ? chartEnd : addDays(chartStart, offset);
        return {
            label: formatDate(date.toISOString()),
            left: (offset / totalDays) * 100,
        };
    });
}
function renderGanttPhaseRow(phase, start, end, chartStart, totalDays, tasks, todayLeft) {
    const phaseTasks = tasks.filter(task => task.project_phase_id === phase.id);
    const progress = getPhaseProgress(phaseTasks);
    const offsetDays = signedDaysBetween(chartStart, start);
    const durationDays = Math.max(daysBetween(start, end), 1);
    const left = clamp((offsetDays / totalDays) * 100, 0, 100);
    const width = clamp((durationDays / totalDays) * 100, 2, 100 - left);
    return `
        <div class="gantt-row">
            <div class="gantt-phase-label">
                <div class="gantt-phase-name">${escapeHtml(phase.name)}</div>
                <div class="gantt-phase-meta">
                    <span>${formatDate(phase.start_date)} - ${formatDate(phase.end_date)}</span>
                    <span>Выполнено ${progress}%</span>
                </div>
            </div>
            <div class="gantt-track" aria-label="${escapeHtml(phase.name)}: ${progress}%">
                ${todayLeft != null ? `<span class="gantt-today-line" style="left: ${todayLeft}%;"></span>` : ""}
                <div class="gantt-bar" style="left: ${left}%; width: ${width}%;">
                    <div class="gantt-bar-progress" style="width: ${progress}%;"></div>
                </div>
            </div>
        </div>
    `;
}
function getPhaseProgress(tasks) {
    if (tasks.length === 0)
        return 0;
    const completed = tasks.filter(isTaskCompleted).length;
    return Math.round((completed / tasks.length) * 100);
}
function isTaskCompleted(task) {
    const status = task.status || statusesState.find(item => item.id === task.status_id);
    if (!status)
        return Boolean(task.completed);
    if (status.is_final)
        return true;
    const normalized = status.name.toLowerCase();
    return ["done", "completed", "готов", "выполн", "заверш", "закрыт"].some(marker => normalized.includes(marker));
}
function renderProjectTaskColumn(status, tasks, phaseByID) {
    return `
        <div class="project-task-column" data-task-drop-status="${status.id}">
            <h4>${escapeHtml(status.name)}</h4>
            ${tasks.length > 0 ? renderProjectTaskList(tasks, phaseByID) : `<div class="empty-state">Нет задач</div>`}
        </div>
    `;
}
function renderProjectTaskList(tasks, phaseByID) {
    const sortedTasks = [...tasks].sort((a, b) => Number(new Date(b.end_time || 0)) - Number(new Date(a.end_time || 0)));
    return `
        <div class="task-list">
            ${sortedTasks.map(task => renderProjectTaskCard(task, phaseByID)).join("") || `<div class="empty-state">Нет задач</div>`}
        </div>
    `;
}
function renderProjectTaskCard(task, phaseByID) {
    const status = task.status || statusesState.find(item => item.id === task.status_id);
    const phaseName = task.project_phase_id ? phaseByID.get(task.project_phase_id) || `Этап #${task.project_phase_id}` : "Без этапа";
    const transitions = getAvailableStatusTransitions(task);
    const actionButtonClass = "btn btn-sm project-task-action-btn";
    return `
        <article class="task-item project-task-card" data-open-task="${task.id}" data-task-id="${task.id}" data-task-status-id="${task.status_id}" draggable="true">
            <div class="task-header">
                ${escapeHtml(task.name)}
                <span class="status-badge">${escapeHtml(status?.name || "Без статуса")}</span>
            </div>
            <div class="task-meta">Проект: ${escapeHtml(selectedProjectId ? projectsState.find(project => project.id === selectedProjectId)?.name || "" : "")}</div>
            <div class="task-meta">Этап: ${escapeHtml(phaseName)}</div>
            <div class="task-meta">Исполнитель: ${escapeHtml(task.executor?.name || getUserName(task.executor_id))}</div>
            <div class="task-meta">Приоритет: ${getPriorityLabel(task.priority)}</div>
            ${task.end_time ? `<div class="task-meta">Срок: ${formatDateTime(task.end_time)}</div>` : ""}
            <div class="task-actions">
                ${transitions.map(transition => {
        const toStatus = statusesState.find(item => item.id === transition.to_status_id) || transition.to_status;
        return `<button class="${actionButtonClass} btn-neutral" type="button" data-transition-task="${task.id}" data-to-status="${transition.to_status_id}">→ ${escapeHtml(toStatus?.name || "Статус")}</button>`;
    }).join("")}
                <button class="${actionButtonClass} btn-neutral" type="button" data-open-task="${task.id}">Открыть</button>
                <button class="${actionButtonClass} btn-danger" type="button" data-delete-task="${task.id}">Удалить</button>
            </div>
        </article>
    `;
}
function setupProjectTaskDragAndDrop(projectId) {
    document.querySelectorAll(".project-task-card[draggable='true']").forEach((card) => {
        card.addEventListener("dragstart", (event) => {
            const taskId = card.dataset.taskId;
            const statusId = card.dataset.taskStatusId;
            if (!taskId || !statusId || !event.dataTransfer)
                return;
            event.dataTransfer.effectAllowed = "move";
            event.dataTransfer.setData("application/x-task-id", taskId);
            event.dataTransfer.setData("application/x-task-status-id", statusId);
            card.classList.add("dragging");
        });
        card.addEventListener("dragend", () => {
            card.classList.remove("dragging");
            clearTaskDropTargets();
        });
    });
    document.querySelectorAll("[data-task-drop-status]").forEach((column) => {
        column.addEventListener("dragover", (event) => {
            if (!event.dataTransfer?.types.includes("application/x-task-id"))
                return;
            event.preventDefault();
            event.dataTransfer.dropEffect = "move";
            column.classList.add("drag-over");
        });
        column.addEventListener("dragleave", (event) => {
            if (!column.contains(event.relatedTarget)) {
                column.classList.remove("drag-over");
            }
        });
        column.addEventListener("drop", async (event) => {
            event.preventDefault();
            const taskId = Number(event.dataTransfer?.getData("application/x-task-id"));
            const fromStatusId = Number(event.dataTransfer?.getData("application/x-task-status-id"));
            const toStatusId = Number(column.dataset.taskDropStatus);
            clearTaskDropTargets();
            if (!taskId || !toStatusId || fromStatusId === toStatusId)
                return;
            const changed = await updateProjectTaskStatus(projectId, taskId, toStatusId);
            if (!changed) {
                markInvalidDropTarget(column);
            }
        });
    });
}
function clearTaskDropTargets() {
    document.querySelectorAll(".project-task-column.drag-over").forEach((column) => {
        column.classList.remove("drag-over");
    });
}
function markInvalidDropTarget(element) {
    element.classList.add("drop-invalid");
    window.setTimeout(() => element.classList.remove("drop-invalid"), 1200);
}
function getAvailableStatusTransitions(task) {
    return statusTransitionsState.filter(transition => transition.from_status_id === task.status_id);
}
function renderProjectTeamTab(project) {
    const members = project.members || [];
    const memberCards = [
        renderProjectUserCard(project.responsible || usersState.find(user => user.id === project.responsible_id), "Ответственный"),
        ...members
            .filter(member => member.user_id !== project.responsible_id)
            .map(member => renderProjectUserCard(member.user || usersState.find(user => user.id === member.user_id), member.role)),
    ];
    return `
        <section class="project-section">
            <div class="project-section-heading">
                <h3>Команда проекта</h3>
            </div>
            <div class="project-user-grid">
                ${memberCards.join("")}
            </div>
        </section>
    `;
}
function renderProjectChat(messages) {
    return `
        <section class="project-section project-chat" aria-label="Чат проекта">
            <div class="project-section-heading">
                <h3>Чат проекта</h3>
            </div>
            <div class="project-chat-list">
                ${messages.length > 0 ? messages.map(renderProjectChatMessage).join("") : `<div class="empty-state">Сообщений пока нет</div>`}
            </div>
            <form id="projectChatForm" class="project-chat-form">
                <div id="projectChatReply" class="project-chat-reply" hidden></div>
                <textarea name="content" rows="3" placeholder="Напишите сообщение" required></textarea>
                <div class="form-actions">
                    <button type="button" class="btn btn-neutral" data-cancel-chat-reply hidden>Отменить ответ</button>
                    <button type="submit" class="btn btn-primary">Отправить</button>
                </div>
            </form>
        </section>
    `;
}
function renderProjectChatMessage(message) {
    const currentUserId = getCurrentUserId();
    const canModify = checkUserRole("admin") || message.user_id === currentUserId;
    const author = message.user?.name || getUserName(message.user_id);
    const reply = message.reply_to;
    const createdAt = message.created_at || message.CreatedAt;
    const updatedAt = message.updated_at || message.UpdatedAt;
    return `
        <article class="project-chat-message" data-message-id="${message.id}">
            ${reply ? `
                <button class="project-chat-quote" type="button" data-scroll-message="${reply.id}">
                    <span>${escapeHtml(reply.user?.name || getUserName(reply.user_id))}</span>
                    ${escapeHtml(reply.content)}
                </button>
            ` : ""}
            <div class="project-chat-message-head">
                <div>
                    <strong>${escapeHtml(author)}</strong>
                    <span>${formatDateTime(createdAt)}</span>
                    ${updatedAt && createdAt && updatedAt !== createdAt ? `<span>изменено</span>` : ""}
                </div>
                <div class="project-chat-actions">
                    <button type="button" class="btn btn-neutral btn-sm" data-reply-message="${message.id}">Ответить</button>
                    ${canModify ? `<button type="button" class="btn btn-neutral btn-sm" data-edit-message="${message.id}">Редактировать</button>` : ""}
                    ${canModify ? `<button type="button" class="btn btn-danger btn-sm" data-delete-message="${message.id}">Удалить</button>` : ""}
                </div>
            </div>
            <div class="project-chat-message-content">${escapeHtml(message.content)}</div>
        </article>
    `;
}
function setupProjectChat(projectId, messages) {
    const form = document.getElementById("projectChatForm");
    const textarea = form?.querySelector("textarea[name='content']");
    const replyBlock = document.getElementById("projectChatReply");
    const cancelReplyButton = form?.querySelector("[data-cancel-chat-reply]");
    let replyToId = null;
    const clearReply = () => {
        replyToId = null;
        if (replyBlock) {
            replyBlock.hidden = true;
            replyBlock.innerHTML = "";
        }
        if (cancelReplyButton)
            cancelReplyButton.hidden = true;
    };
    document.querySelectorAll("[data-reply-message]").forEach((button) => {
        button.addEventListener("click", () => {
            replyToId = Number(button.dataset.replyMessage) || null;
            const message = messages.find(item => item.id === replyToId);
            if (replyBlock && message) {
                replyBlock.hidden = false;
                replyBlock.innerHTML = `
                    <span>Ответ на ${escapeHtml(message.user?.name || getUserName(message.user_id))}</span>
                    ${escapeHtml(message.content)}
                `;
            }
            if (cancelReplyButton)
                cancelReplyButton.hidden = false;
            textarea?.focus();
        });
    });
    cancelReplyButton?.addEventListener("click", clearReply);
    document.querySelectorAll("[data-scroll-message]").forEach((button) => {
        button.addEventListener("click", () => {
            const messageId = Number(button.dataset.scrollMessage);
            document.querySelector(`[data-message-id="${messageId}"]`)?.scrollIntoView({ behavior: "smooth", block: "center" });
        });
    });
    document.querySelectorAll("[data-edit-message]").forEach((button) => {
        button.addEventListener("click", () => {
            const messageId = Number(button.dataset.editMessage);
            const message = messages.find(item => item.id === messageId);
            const card = document.querySelector(`[data-message-id="${messageId}"]`);
            const content = card?.querySelector(".project-chat-message-content");
            if (!message || !content)
                return;
            content.innerHTML = `
                <form class="project-chat-edit-form" data-message-edit-form="${message.id}">
                    <textarea name="content" rows="3" required>${escapeHtml(message.content)}</textarea>
                    <div class="form-actions">
                        <button type="button" class="btn btn-neutral btn-sm" data-cancel-message-edit>Отмена</button>
                        <button type="submit" class="btn btn-primary btn-sm">Сохранить</button>
                    </div>
                </form>
            `;
            const editForm = content.querySelector("[data-message-edit-form]");
            editForm?.querySelector("textarea")?.focus();
            editForm?.addEventListener("submit", async (event) => {
                event.preventDefault();
                const raw = Object.fromEntries(new FormData(editForm));
                const nextContent = String(raw.content || "").trim();
                if (!nextContent)
                    return;
                const res = await apiFetch(`/projects/${projectId}/messages/${message.id}`, {
                    method: "PUT",
                    body: JSON.stringify({ content: nextContent }),
                });
                if (res.ok) {
                    await refreshProjectChat(projectId, "scroll");
                    return;
                }
                const err = await res.json().catch(() => ({ error: "Не удалось сохранить сообщение" }));
                alert(`Ошибка: ${err.error}`);
            });
            editForm?.querySelector("[data-cancel-message-edit]")?.addEventListener("click", () => {
                content.innerHTML = escapeHtml(message.content);
            });
        });
    });
    document.querySelectorAll("[data-delete-message]").forEach((button) => {
        button.addEventListener("click", async () => {
            const messageId = Number(button.dataset.deleteMessage);
            if (!messageId || !confirm("Удалить сообщение?"))
                return;
            const res = await apiFetch(`/projects/${projectId}/messages/${messageId}`, { method: "DELETE" });
            if (res.ok) {
                await refreshProjectChat(projectId, "scroll");
            }
            else {
                alert("Не удалось удалить сообщение");
            }
        });
    });
    form?.addEventListener("submit", async (event) => {
        event.preventDefault();
        const content = textarea?.value.trim() || "";
        if (!content)
            return;
        const res = await apiFetch(`/projects/${projectId}/messages`, {
            method: "POST",
            body: JSON.stringify({ content, reply_to_id: replyToId }),
        });
        if (res.ok) {
            clearReply();
            if (textarea)
                textarea.value = "";
            await refreshProjectChat(projectId, "form");
            return;
        }
        const err = await res.json().catch(() => ({ error: "Не удалось отправить сообщение" }));
        alert(`Ошибка: ${err.error}`);
    });
}
function renderProjectUserCard(user, role) {
    return `
        <div class="project-user-card">
            <div class="project-user-avatar">${escapeHtml((user?.name || "?").slice(0, 1).toUpperCase())}</div>
            <div class="project-user-body">
                <div class="person-card-title">${escapeHtml(user?.name || "Пользователь")}</div>
                <div class="person-card-meta">
                    <span>${escapeHtml(role || "Участник")}</span>
                    ${user?.email ? `<span>${escapeHtml(user.email)}</span>` : ""}
                    ${user?.role ? `<span>${escapeHtml(user.role)}</span>` : ""}
                </div>
            </div>
        </div>
    `;
}
function openStageModal(project) {
    const modal = createActionModal(`
        <h3>Сменить стадию проекта</h3>
        <form id="stageChangeForm">
            <div class="form-group">
                <label for="projectDetailStage">Стадия</label>
                <select id="projectDetailStage" name="stage_id" required>
                    ${stagesState.map(stage => `<option value="${stage.id}" ${stage.id === project.stage_id ? "selected" : ""}>${escapeHtml(stage.name)}</option>`).join("")}
                </select>
            </div>
            <div class="form-actions">
                <button type="submit">Сохранить</button>
            </div>
        </form>
    `);
    const form = modal.querySelector("#stageChangeForm");
    form.addEventListener("submit", async (event) => {
        event.preventDefault();
        const stageId = Number(form.elements.namedItem("stage_id").value);
        const res = await apiFetch(`/projects/${project.id}`, {
            method: "PUT",
            body: JSON.stringify({ stage_id: stageId }),
        });
        if (res.ok) {
            modal.remove();
            await openProjectDetail(project.id);
            return;
        }
        const err = await res.json().catch(() => ({ error: "Не удалось сменить стадию" }));
        alert(`Ошибка: ${err.error}`);
    });
}
async function openProjectDetailsModal(project) {
    if (stagesState.length === 0)
        await loadStages();
    if (usersState.length === 0)
        await loadUsers();
    const modal = createActionModal(`
        <h3>Редактировать сведения о проекте</h3>
        <form id="projectDetailsEditForm">
            <div class="form-group">
                <label>Название *</label>
                <input type="text" name="name" value="${escapeHtml(project.name)}" required />
            </div>
            <div class="form-group">
                <label>Описание</label>
                <textarea name="description" rows="4">${escapeHtml(project.description || "")}</textarea>
            </div>
            <div class="form-group">
                <label>Адрес</label>
                <input type="text" name="address" value="${escapeHtml(project.address || "")}" />
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label>Дата начала</label>
                    <input type="date" name="start_date" value="${toDateInputValue(project.start_date)}" />
                </div>
                <div class="form-group">
                    <label>Дата конца</label>
                    <input type="date" name="deadline_date" value="${toDateInputValue(project.deadline_date)}" />
                </div>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label>Стоимость</label>
                    <input type="number" step="0.01" name="cost" value="${project.cost != null ? project.cost : ""}" />
                </div>
                <div class="form-group">
                    <label>Стадия *</label>
                    <select name="stage_id" required>
                        <option value="">Выберите стадию</option>
                        ${stagesState.map(stage => `<option value="${stage.id}" ${stage.id === project.stage_id ? "selected" : ""}>${escapeHtml(stage.name)}</option>`).join("")}
                    </select>
                </div>
            </div>
            <div class="form-group">
                <label>Ответственный *</label>
                <select name="responsible_id" required>
                    <option value="">Выберите пользователя</option>
                    ${usersState.map(user => `<option value="${user.id}" ${user.id === project.responsible_id ? "selected" : ""}>${escapeHtml(user.name)}${user.email ? ` (${escapeHtml(user.email)})` : ""}</option>`).join("")}
                </select>
            </div>
            <div class="form-group">
                <label>Команда проекта</label>
                ${renderProjectMemberPicker(project)}
            </div>
            ${checkUserRole("admin") ? `
                <div class="form-group">
                    <label class="project-hidden-toggle">
                        <input type="checkbox" name="is_hidden" data-project-hidden ${project.is_hidden ? "checked" : ""} />
                        <span>Скрытый проект</span>
                    </label>
                </div>
                <div class="form-group" data-project-access-block hidden>
                    <label>Пользователи с доступом</label>
                    ${renderProjectAccessPicker(getProjectAccessUserIDs(project))}
                </div>
            ` : ""}
            <div class="form-group">
                <label>Оценка</label>
                <textarea name="assessment" rows="3">${escapeHtml(project.assessment || "")}</textarea>
            </div>
            <div class="form-group">
                <label>Комментарии</label>
                <textarea name="comments" rows="3">${escapeHtml(project.comments || "")}</textarea>
            </div>
            <div class="form-actions">
                <button type="submit">Сохранить</button>
            </div>
        </form>
    `);
    const form = modal.querySelector("#projectDetailsEditForm");
    setupHiddenProjectControls(modal);
    form.addEventListener("submit", async (event) => {
        event.preventDefault();
        const raw = Object.fromEntries(new FormData(form));
        const payload = {
            name: String(raw.name || ""),
            description: String(raw.description || ""),
            address: String(raw.address || ""),
            start_date: String(raw.start_date || ""),
            deadline_date: String(raw.deadline_date || ""),
            cost: raw.cost ? Number(raw.cost) : 0,
            stage_id: Number(raw.stage_id),
            responsible_id: Number(raw.responsible_id),
            assessment: String(raw.assessment || ""),
            comments: String(raw.comments || ""),
        };
        if (checkUserRole("admin")) {
            const isHidden = Boolean(form.elements.namedItem("is_hidden")?.checked);
            payload.is_hidden = isHidden;
            payload.access_user_ids = isHidden ? getCheckedProjectAccessUserIDs(form) : [];
        }
        if (!payload.name || !payload.stage_id || !payload.responsible_id) {
            alert("Заполните все обязательные поля");
            return;
        }
        const res = await apiFetch(`/projects/${project.id}`, {
            method: "PUT",
            body: JSON.stringify(payload),
        });
        if (res.ok) {
            await syncProjectMembers(project.id, project.members || [], getCheckedProjectMemberUserIDs(form), payload.responsible_id);
            modal.remove();
            await openProjectDetail(project.id);
            return;
        }
        const err = await res.json().catch(() => ({ error: "Не удалось сохранить сведения проекта" }));
        alert(`Ошибка: ${err.error}`);
    });
}
function openPhaseModal(project) {
    const modal = createActionModal(`
        <h3>Создать этап</h3>
        <form id="phaseCreateForm">
            <div class="form-group">
                <label>Название *</label>
                <input type="text" name="name" required />
            </div>
            <div class="form-group">
                <label>Описание</label>
                <textarea name="description"></textarea>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label>Дата начала</label>
                    <input type="date" name="start_date" />
                </div>
                <div class="form-group">
                    <label>Дата окончания</label>
                    <input type="date" name="end_date" />
                </div>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label>Позиция</label>
                    <input type="number" name="position" min="0" value="0" />
                </div>
                <div class="form-group">
                    <label>Ответственный *</label>
                    <select name="responsible_id" required>
                        <option value="">Выберите пользователя</option>
                        ${usersState.map(user => `<option value="${user.id}" ${user.id === project.responsible_id ? "selected" : ""}>${escapeHtml(user.name)}</option>`).join("")}
                    </select>
                </div>
            </div>
            <div class="form-actions">
                <button type="submit">Создать</button>
            </div>
        </form>
    `);
    const form = modal.querySelector("#phaseCreateForm");
    form.addEventListener("submit", async (event) => {
        event.preventDefault();
        const raw = Object.fromEntries(new FormData(form));
        const payload = {
            project_id: project.id,
            name: raw.name,
            description: raw.description || "",
            start_date: dateInputToISO(raw.start_date),
            end_date: dateInputToISO(raw.end_date),
            position: Number(raw.position) || 0,
            responsible_id: Number(raw.responsible_id),
        };
        const res = await apiFetch("/project-phases", {
            method: "POST",
            body: JSON.stringify(payload),
        });
        if (res.ok) {
            modal.remove();
            await openProjectDetail(project.id);
            return;
        }
        const err = await res.json().catch(() => ({ error: "Не удалось создать этап" }));
        alert(`Ошибка: ${err.error}`);
    });
}
function openTaskModal(project, phases, selectedPhaseId) {
    const defaultStatusId = getDefaultStatusId();
    if (!defaultStatusId) {
        alert("Сначала настройте статусы задач");
        return;
    }
    const modal = createActionModal(`
        <h3>Создать задачу</h3>
        <form id="projectTaskCreateForm">
            <div class="form-group">
                <label>Название *</label>
                <input type="text" name="name" required />
            </div>
            <div class="form-group">
                <label>Описание</label>
                <textarea name="description"></textarea>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label>Этап</label>
                    <select name="project_phase_id">
                        <option value="">Без этапа</option>
                        ${phases.map(phase => `<option value="${phase.id}" ${phase.id === selectedPhaseId ? "selected" : ""}>${escapeHtml(phase.name)}</option>`).join("")}
                    </select>
                </div>
                <div class="form-group">
                    <label>Статус *</label>
                    <select name="status_id" required>
                        ${statusesState.map(status => `<option value="${status.id}" ${status.id === defaultStatusId ? "selected" : ""}>${escapeHtml(status.name)}</option>`).join("")}
                    </select>
                </div>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label>Исполнитель *</label>
                    <select name="executor_id" required>
                        <option value="">Выберите пользователя</option>
                        ${usersState.map(user => `<option value="${user.id}">${escapeHtml(user.name)}${user.email ? ` (${escapeHtml(user.email)})` : ""}</option>`).join("")}
                    </select>
                </div>
                <div class="form-group">
                    <label>Приоритет</label>
                    <select name="priority">
                        <option value="0">Низкий</option>
                        <option value="1">Средний</option>
                        <option value="2">Высокий</option>
                    </select>
                </div>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label>Начало *</label>
                    <input type="datetime-local" name="start_time" value="${getLocalDateTimeValue()}" required />
                </div>
                <div class="form-group">
                    <label>Окончание *</label>
                    <input type="datetime-local" name="end_time" required />
                </div>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label>Важность</label>
                    <input type="number" name="urgency" min="0" step="0.1" value="1" />
                </div>
                <div class="form-group">
                    <label>Сложность</label>
                    <input type="number" name="complexity" min="0" step="0.1" value="1" />
                </div>
            </div>
            <div class="form-actions">
                <button type="submit">Создать</button>
            </div>
        </form>
    `);
    const form = modal.querySelector("#projectTaskCreateForm");
    const submitButton = form.querySelector('button[type="submit"]');
    let isSubmitting = false;
    form.addEventListener("submit", async (event) => {
        event.preventDefault();
        if (isSubmitting)
            return;
        isSubmitting = true;
        if (submitButton) {
            submitButton.disabled = true;
            submitButton.textContent = "Создание...";
        }
        try {
            const raw = Object.fromEntries(new FormData(form));
            const payload = {
                name: raw.name,
                description: raw.description || "",
                project_id: project.id,
                project_phase_id: raw.project_phase_id ? Number(raw.project_phase_id) : null,
                executor_id: Number(raw.executor_id),
                status_id: Number(raw.status_id),
                priority: Number(raw.priority) || 0,
                start_time: dateTimeInputToISO(raw.start_time),
                end_time: dateTimeInputToISO(raw.end_time),
                urgency: raw.urgency ? Number(raw.urgency) : 1,
                complexity: raw.complexity ? Number(raw.complexity) : 1,
            };
            const res = await apiFetch("/tasks", {
                method: "POST",
                body: JSON.stringify(payload),
            });
            if (res.ok) {
                modal.remove();
                await openProjectDetail(project.id);
                return;
            }
            const err = await res.json().catch(() => ({ error: "Не удалось создать задачу" }));
            alert(`Ошибка: ${err.error}`);
        }
        catch (err) {
            alert("Ошибка сети");
        }
        isSubmitting = false;
        if (submitButton) {
            submitButton.disabled = false;
            submitButton.textContent = "Создать";
        }
    });
}
async function deleteProjectTask(projectId, taskId) {
    if (!confirm("Удалить эту задачу?"))
        return;
    try {
        const res = await apiFetch(`/tasks/${taskId}`, {
            method: "DELETE",
        });
        if (res.ok) {
            await openProjectDetail(projectId);
            setProjectTab("tasks");
            return;
        }
        const err = await res.json().catch(() => ({ error: "Не удалось удалить задачу" }));
        alert(`Ошибка: ${err.error}`);
    }
    catch (err) {
        alert("Ошибка сети");
    }
}
async function updateProjectTaskStatus(projectId, taskId, statusId) {
    try {
        const res = await apiFetch(`/tasks/${taskId}/status`, {
            method: "PUT",
            body: JSON.stringify({ status_id: statusId }),
        });
        if (res.ok) {
            await openProjectDetail(projectId);
            setProjectTab("tasks");
            return true;
        }
        const err = await res.json().catch(() => ({ error: "Не удалось изменить статус задачи" }));
        if (err.error !== "invalid status transition") {
            alert(`Ошибка: ${err.error || "Не удалось изменить статус задачи"}`);
        }
        return false;
    }
    catch (err) {
        alert("Ошибка сети");
        return false;
    }
}
function openTaskCard(task, phases) {
    const canEdit = checkUserRole("admin");
    const project = projectsState.find(item => item.id === task.project_id);
    const modal = createActionModal(`
        <button type="button" class="modal-close-x" aria-label="Закрыть">×</button>
        <div class="task-card-detail task-card-formatted">
            <div class="task-card-heading">
                <div>
                    <div class="task-card-kicker">${escapeHtml(project?.name || `Проект #${task.project_id}`)}</div>
                    <h3>${escapeHtml(task.name)}</h3>
                </div>
                <span class="status-badge">${escapeHtml(task.status?.name || "Без статуса")}</span>
            </div>
            <form id="projectTaskEditForm">
                <div class="task-card-grid">
                    <div class="form-group task-card-field-wide">
                        <label>Название</label>
                        <input type="text" name="name" value="${escapeHtml(task.name)}" ${canEdit ? "required" : "disabled"} />
                    </div>
                    <div class="form-group">
                        <label>Статус</label>
                        <select name="status_id" ${canEdit ? "" : "disabled"}>
                            ${statusesState.map(status => `<option value="${status.id}" ${status.id === task.status_id ? "selected" : ""}>${escapeHtml(status.name)}</option>`).join("")}
                        </select>
                    </div>
                    <div class="form-group">
                        <label>Этап</label>
                        <select name="project_phase_id" ${canEdit ? "" : "disabled"}>
                            <option value="" ${task.project_phase_id ? "" : "selected"}>Без этапа</option>
                            ${phases.map(phase => `<option value="${phase.id}" ${phase.id === task.project_phase_id ? "selected" : ""}>${escapeHtml(phase.name)}</option>`).join("")}
                        </select>
                    </div>
                    <div class="form-group">
                        <label>Исполнитель</label>
                        <select name="executor_id" ${canEdit ? "" : "disabled"}>
                            ${usersState.map(user => `<option value="${user.id}" ${user.id === task.executor_id ? "selected" : ""}>${escapeHtml(user.name)}${user.email ? ` (${escapeHtml(user.email)})` : ""}</option>`).join("")}
                        </select>
                    </div>
                    <div class="form-group">
                        <label>Приоритет</label>
                        <select name="priority" ${canEdit ? "" : "disabled"}>
                            <option value="0" ${task.priority === 0 ? "selected" : ""}>Низкий</option>
                            <option value="1" ${task.priority === 1 ? "selected" : ""}>Средний</option>
                            <option value="2" ${task.priority === 2 ? "selected" : ""}>Высокий</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label>Категория</label>
                        <input type="text" name="category" value="${escapeHtml(task.category || "")}" ${canEdit ? "" : "disabled"} />
                    </div>
                    <div class="form-group">
                        <label>Начало</label>
                        <input type="datetime-local" name="start_time" value="${dateTimeToInputValue(task.start_time)}" ${canEdit ? "" : "disabled"} />
                    </div>
                    <div class="form-group">
                        <label>Срок</label>
                        <input type="datetime-local" name="end_time" value="${dateTimeToInputValue(task.end_time)}" ${canEdit ? "" : "disabled"} />
                    </div>
                    <div class="form-group">
                        <label>Важность</label>
                        <input type="number" name="urgency" min="0" step="0.1" value="${task.urgency ?? 1}" ${canEdit ? "" : "disabled"} />
                    </div>
                    <div class="form-group">
                        <label>Сложность</label>
                        <input type="number" name="complexity" min="0" step="0.1" value="${task.complexity ?? 1}" ${canEdit ? "" : "disabled"} />
                    </div>
                    <div class="form-group task-card-field-wide">
                        <label>Описание</label>
                        <textarea name="description" rows="7" ${canEdit ? "" : "disabled"}>${escapeHtml(task.description || "")}</textarea>
                    </div>
                </div>
                ${canEdit ? `
                    <div class="form-actions">
                        <button type="submit">Сохранить</button>
                    </div>
                ` : ""}
            </form>
        </div>
    `);
    const closeButton = modal.querySelector(".modal-close-x");
    closeButton?.addEventListener("click", () => modal.remove());
    const form = modal.querySelector("#projectTaskEditForm");
    if (canEdit && form) {
        form.addEventListener("submit", async (event) => {
            event.preventDefault();
            const raw = Object.fromEntries(new FormData(form));
            const res = await apiFetch(`/tasks/${task.id}`, {
                method: "PUT",
                body: JSON.stringify({
                    name: raw.name,
                    description: raw.description || "",
                    category: raw.category || "",
                    status_id: Number(raw.status_id),
                    project_phase_id: raw.project_phase_id ? Number(raw.project_phase_id) : null,
                    executor_id: Number(raw.executor_id),
                    priority: Number(raw.priority) || 0,
                    start_time: dateTimeInputToISO(raw.start_time),
                    end_time: dateTimeInputToISO(raw.end_time),
                    urgency: raw.urgency ? Number(raw.urgency) : 1,
                    complexity: raw.complexity ? Number(raw.complexity) : 1,
                }),
            });
            if (res.ok) {
                modal.remove();
                await openProjectDetail(task.project_id);
                setProjectTab("tasks");
                return;
            }
            const err = await res.json().catch(() => ({ error: "Не удалось сохранить задачу" }));
            alert(`Ошибка: ${err.error}`);
        });
    }
}
function openClientModal(project) {
    const client = getProjectClient(project);
    const modal = createActionModal(`
        <h3>Клиент проекта</h3>
        <form id="clientForm">
            <div class="form-row">
                <div class="form-group">
                    <label>Компания из справочника</label>
                    <select name="organization_id" data-client-organization>
                        <option value="">Новая компания</option>
                        ${organizationsState.map(org => `<option value="${org.id}" ${org.id === client?.organization_id ? "selected" : ""}>${escapeHtml(org.name)}</option>`).join("")}
                    </select>
                </div>
                <div class="form-group">
                    <label>Новая компания</label>
                    <input type="text" name="organization_name" placeholder="Название, если нет в списке" />
                </div>
            </div>
            <div class="form-group">
                <label>Описание компании</label>
                <textarea name="organization_description" rows="3"></textarea>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label>Телефон компании</label>
                    <input type="text" name="organization_phone" />
                </div>
                <div class="form-group">
                    <label>Email компании</label>
                    <input type="email" name="organization_email" />
                </div>
            </div>
            <div class="form-group">
                <label>Адрес компании</label>
                <input type="text" name="organization_address" />
            </div>
            <div class="form-actions">
                <button type="submit">Сохранить</button>
            </div>
        </form>
    `);
    modal.classList.add("client-modal");
    const form = modal.querySelector("#clientForm");
    form.addEventListener("submit", async (event) => {
        event.preventDefault();
        const raw = Object.fromEntries(new FormData(form));
        let organizationID = Number(raw.organization_id) || 0;
        const organizationName = String(raw.organization_name || "").trim();
        try {
            if (!organizationID) {
                if (!organizationName) {
                    alert("Выберите компанию или укажите новую");
                    return;
                }
                const orgRes = await apiFetch("/organizations", {
                    method: "POST",
                    body: JSON.stringify({
                        name: organizationName,
                        description: String(raw.organization_description || ""),
                        phone: String(raw.organization_phone || ""),
                        email: String(raw.organization_email || ""),
                        address: String(raw.organization_address || ""),
                    }),
                });
                if (!orgRes.ok) {
                    const err = await orgRes.json().catch(() => ({ error: "Не удалось создать компанию" }));
                    alert(`Ошибка: ${err.error}`);
                    return;
                }
                const org = await orgRes.json();
                organizationID = org.id;
                await loadOrganizations();
            }
            const participantPayload = {
                id: client?.id,
                project_id: project.id,
                organization_id: organizationID,
                role: "client",
            };
            const res = await apiFetch(client ? `/participants/${client.id}` : "/participants", {
                method: client ? "PUT" : "POST",
                body: JSON.stringify(participantPayload),
            });
            if (!res.ok) {
                const err = await res.json().catch(() => ({ error: "Не удалось сохранить клиента" }));
                alert(`Ошибка: ${err.error}`);
                return;
            }
            modal.remove();
            await openProjectDetail(project.id);
        }
        catch (err) {
            alert("Ошибка сети");
        }
    });
}
function openClientContactModal(project) {
    const client = getProjectClient(project);
    if (!client) {
        alert("Сначала добавьте клиента");
        return;
    }
    const modal = createActionModal(`
        <h3>Добавить контакт клиента</h3>
        <form id="clientContactForm">
            <div class="form-group">
                <label>Компания</label>
                <input type="text" value="${escapeHtml(client.organization?.name || `Компания #${client.organization_id}`)}" disabled />
            </div>
            <div class="form-group">
                <label>ФИО *</label>
                <input type="text" name="full_name" required />
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label>Должность</label>
                    <input type="text" name="position" />
                </div>
                <div class="form-group">
                    <label>Телефон</label>
                    <input type="text" name="phone" />
                </div>
            </div>
            <div class="form-group">
                <label>Email</label>
                <input type="email" name="email" />
            </div>
            <div class="form-group">
                <label>Заметки</label>
                <textarea name="notes"></textarea>
            </div>
            <div class="form-actions">
                <button type="submit">Добавить</button>
            </div>
        </form>
    `);
    const form = modal.querySelector("#clientContactForm");
    form.addEventListener("submit", async (event) => {
        event.preventDefault();
        const raw = Object.fromEntries(new FormData(form));
        const res = await apiFetch("/organization-contacts", {
            method: "POST",
            body: JSON.stringify({
                organization_id: client.organization_id,
                full_name: String(raw.full_name || "").trim(),
                position: String(raw.position || ""),
                phone: String(raw.phone || ""),
                email: String(raw.email || ""),
                notes: String(raw.notes || ""),
            }),
        });
        if (res.ok) {
            modal.remove();
            return;
        }
        const err = await res.json().catch(() => ({ error: "Не удалось добавить контакт" }));
        alert(`Ошибка: ${err.error}`);
    });
}
function createActionModal(innerHtml) {
    const modal = document.createElement("div");
    modal.className = "modal project-action-modal";
    const closeButtonHtml = innerHtml.includes("modal-close-x")
        ? ""
        : `<button type="button" class="modal-close-x" aria-label="Закрыть">×</button>`;
    modal.innerHTML = `<div class="modal-content">${closeButtonHtml}${innerHtml}</div>`;
    document.body.appendChild(modal);
    modal.querySelectorAll(".form-actions .cancel").forEach(button => button.remove());
    modal.querySelector(".modal-close-x")?.addEventListener("click", () => modal.remove());
    modal.addEventListener("click", (event) => {
        if (event.target === modal) {
            modal.remove();
        }
    });
    return modal;
}
// === Вспомогательные функции ===
function checkUserRole(role) {
    const user = localStorage.getItem('user');
    if (user) {
        return JSON.parse(user).role === role;
    }
    try {
        const payload = JSON.parse(atob((token || "").split(".")[1] || ""));
        return payload.role === role;
    }
    catch {
        return false;
    }
}
function getCurrentUserId() {
    const user = localStorage.getItem('user');
    if (user) {
        return JSON.parse(user).id;
    }
    try {
        const payload = JSON.parse(atob((token || "").split(".")[1] || ""));
        return Number(payload.user_id) || null;
    }
    catch {
        return null;
    }
}
// === Модальное окно ===
function openModal(id = null) {
    const modal = document.getElementById("projectModal");
    const title = document.getElementById("modalTitle");
    const form = document.getElementById("projectForm");
    if (id !== null) {
        title.textContent = "Редактировать проект";
        loadProjectData(id);
    }
    else {
        title.textContent = "Добавить проект";
        form.reset();
        document.getElementById("projectID").value = "";
        const accessUsers = document.getElementById("projectAccessUsers");
        if (accessUsers)
            accessUsers.innerHTML = renderProjectAccessPicker();
        setupHiddenProjectControls(document);
    }
    if (modal)
        modal.style.display = "flex";
}
async function loadProjectData(id) {
    try {
        const res = await fetch(`${API_BASE}/projects/${id}`, {
            headers: { Authorization: `Bearer ${token}` },
        });
        const project = await res.json();
        // Проверка прав на редактирование
        const isUserAdmin = checkUserRole('admin');
        const currentUserId = getCurrentUserId();
        if (!isUserAdmin && project.responsible_id !== currentUserId) {
            alert('У вас нет прав на редактирование этого проекта');
            closeModal();
            return;
        }
        const idInput = document.getElementById("projectID");
        const nameInput = document.getElementById("projectName");
        const descInput = document.getElementById("projectDescription");
        const stageSelect = document.getElementById("projectStage");
        const respSelect = document.getElementById("projectResponsible");
        const addressInput = document.getElementById("projectAddress");
        const startInput = document.getElementById("projectStartDate");
        const deadlineInput = document.getElementById("projectDeadlineDate");
        const costInput = document.getElementById("projectCost");
        const commentsInput = document.getElementById("projectComments");
        const hiddenInput = document.getElementById("projectIsHidden");
        const accessUsers = document.getElementById("projectAccessUsers");
        if (idInput)
            idInput.value = project.id.toString();
        if (nameInput)
            nameInput.value = project.name;
        if (descInput)
            descInput.value = project.description || "";
        if (stageSelect)
            stageSelect.value = project.stage_id.toString();
        if (respSelect)
            respSelect.value = project.responsible_id.toString();
        if (addressInput)
            addressInput.value = project.address || "";
        if (startInput)
            startInput.value = project.start_date ? project.start_date.substring(0, 10) : "";
        if (deadlineInput)
            deadlineInput.value = project.deadline_date ? project.deadline_date.substring(0, 10) : "";
        if (costInput)
            costInput.value = project.cost != null ? project.cost.toString() : "";
        if (commentsInput)
            commentsInput.value = project.comments || "";
        if (hiddenInput)
            hiddenInput.checked = Boolean(project.is_hidden);
        if (accessUsers)
            accessUsers.innerHTML = renderProjectAccessPicker(getProjectAccessUserIDs(project));
        setupHiddenProjectControls(document);
    }
    catch (err) {
        console.error("Ошибка загрузки проекта:", err);
        alert("Не удалось загрузить данные проекта");
    }
}
function closeModal() {
    const modal = document.getElementById("projectModal");
    if (modal)
        modal.style.display = "none";
}
async function handleSubmit(e) {
    e.preventDefault();
    const idInput = document.getElementById("projectID");
    const nameInput = document.getElementById("projectName");
    const descInput = document.getElementById("projectDescription");
    const stageSelect = document.getElementById("projectStage");
    const respSelect = document.getElementById("projectResponsible");
    const addressInput = document.getElementById("projectAddress");
    const startInput = document.getElementById("projectStartDate");
    const deadlineInput = document.getElementById("projectDeadlineDate");
    const costInput = document.getElementById("projectCost");
    const commentsInput = document.getElementById("projectComments");
    const hiddenInput = document.getElementById("projectIsHidden");
    const id = idInput?.value;
    const name = nameInput?.value;
    const description = descInput?.value || null;
    const stageId = Number(stageSelect?.value);
    const responsibleId = Number(respSelect?.value);
    const address = addressInput?.value || null;
    const startDate = startInput?.value || null;
    const deadlineDate = deadlineInput?.value || null;
    const cost = costInput?.value ? Number(costInput.value) : null;
    const comments = commentsInput?.value || null;
    if (!name || !stageSelect?.value || !respSelect?.value) {
        alert("Заполните все обязательные поля");
        return;
    }
    const method = id ? "PUT" : "POST";
    const endpoint = id ? `/projects/${id}` : "/projects";
    const body = {
        name,
        description,
        stage_id: stageId,
        responsible_id: responsibleId,
        address,
        start_date: startDate,
        deadline_date: deadlineDate,
        cost,
        comments,
    };
    if (checkUserRole("admin")) {
        const isHidden = Boolean(hiddenInput?.checked);
        body.is_hidden = isHidden;
        body.access_user_ids = isHidden ? getCheckedProjectAccessUserIDs(document) : [];
    }
    const latitude = addressInput?.getAttribute('data-latitude');
    const longitude = addressInput?.getAttribute('data-longitude');
    if (latitude && longitude) {
        body.latitude = parseFloat(latitude);
        body.longitude = parseFloat(longitude);
    }
    if (id)
        body.id = Number(id);
    try {
        const res = await fetch(`${API_BASE}${endpoint}`, {
            method,
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify(body),
        });
        if (res.ok) {
            closeModal();
            if (id && selectedProjectId === Number(id)) {
                await openProjectDetail(Number(id));
                return;
            }
            await loadProjects();
        }
        else {
            const err = await res.json();
            alert(`Ошибка: ${err.error}`);
        }
    }
    catch (err) {
        alert("Ошибка сети");
    }
}
// === Загрузка Yandex Maps API ===
function loadYandexMapsAPI() {
    const script = document.createElement('script');
    script.src = 'https://api-maps.yandex.ru/2.1/?apikey=15277c56-2dcb-4e64-aefe-2394e6f89c1d&lang=ru_RU';
    script.async = true;
    script.onload = () => {
        initYandexMaps();
    };
    script.onerror = () => {
        console.error('Не удалось загрузить Yandex Maps API');
        alert('Карта временно недоступна. Продолжайте вводить адрес вручную.');
    };
    document.head.appendChild(script);
}
async function deleteProject(id) {
    if (!confirm("Вы уверены, что хотите удалить этот проект?"))
        return;
    try {
        const res = await fetch(`${API_BASE}/projects/${id}`, {
            method: "DELETE",
            headers: { Authorization: `Bearer ${token}` },
        });
        if (res.ok) {
            await loadProjects();
        }
        else {
            alert("Не удалось удалить проект");
        }
    }
    catch (err) {
        alert("Ошибка сети");
    }
}
function editProject(id) {
    openModal(id);
}
// === Делаем функции глобальными ===
window.editProject = editProject;
window.closeModal = closeModal;
window.deleteProject = deleteProject;
// === Инициализация ===
window.addEventListener("DOMContentLoaded", () => {
    renderPage();
    loadProjects();
    loadStages();
    loadUsers();
    loadStatuses();
    loadStatusTransitions();
    loadOrganizations();
    loadYandexMapsAPI();
});
export {};
