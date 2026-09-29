let user;
let token = '';
// Меню приложения (заменено поле short на icon)
const menuItems = [
    {
        title: "Задачи",
        path: "/tasks",
        icon: "task",
        roles: ["user", "admin", "manager", "guest"],
        badge: async () => {
            const tasks = await fetchTasks();
            return tasks.filter(t => (!t.completed && t.executor_id === user?.id)).length;
        },
    },
    {
        title: "Проекты",
        path: "/admin/projects",
        icon: "folder",
        roles: ["user", "admin", "manager"],
        visible: async () => {
            if (user?.role === 'admin')
                return true;
            const projects = await fetchVisibleProjects();
            return projects.length > 0;
        },
    },
    {
        title: "Компании",
        path: "/manager/contacts",
        icon: "business",
        roles: ["admin", "manager"],
    },
    {
        title: "Пользователи",
        path: "/admin/users",
        icon: "people",
        roles: ["admin"],
    },
    {
        title: "Статусы",
        path: "/admin/statuses",
        icon: "flag",
        roles: ["admin"],
    },
    {
        title: "Стадии",
        path: "/admin/stages",
        icon: "sync",
        roles: ["admin"],
    },
    {
        title: "Отчёт за сегодня",
        path: "/reports",
        icon: "description",
        roles: ["user", "admin", "manager", "guest"],
    },
    {
        title: "Отчёты",
        path: "/admin/reports",
        icon: "bar_chart",
        roles: ["admin", "manager"],
        visible: async () => {
            if (user?.role === 'admin')
                return true;
            const projects = await fetchUserProjects(user?.id);
            return projects.length > 0;
        },
    },
    {
        title: "Шаблоны задач",
        path: "/admin/task-templates",
        icon: "library_books",
        roles: ["admin"],
    },
    {
        title: "Табель",
        path: "/admin/timesheet",
        icon: "schedule",
        roles: ["admin"],
    },
];
// === Функция открытия модального окна смены пароля ===
function openChangePasswordModal() {
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <button type="button" class="modal-close-x" aria-label="Закрыть">×</button>
            <h3>Изменить пароль</h3>
            <form id="changePasswordForm">
                <div class="form-group">
                    <label>Текущий пароль</label>
                    <input type="password" name="currentPassword" required />
                </div>
                <div class="form-group">
                    <label>Новый пароль</label>
                    <input type="password" name="newPassword" minlength="6" required />
                </div>
                <div class="form-group">
                    <label>Подтвердите новый пароль</label>
                    <input type="password" name="confirmPassword" minlength="6" required />
                </div>
                <div class="form-actions">
                    <button type="submit">Сменить пароль</button>
                </div>
            </form>
        </div>
    `;
    document.body.appendChild(modal);
    const form = modal.querySelector('#changePasswordForm');
    modal.querySelector('.modal-close-x')?.addEventListener('click', () => modal.remove());
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(form);
        const currentPassword = formData.get('currentPassword');
        const newPassword = formData.get('newPassword');
        const confirmPassword = formData.get('confirmPassword');
        if (newPassword !== confirmPassword) {
            alert('Новые пароли не совпадают');
            return;
        }
        try {
            const response = await fetch('/api/v1/users/change-password', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ user_id: user?.id, old_password: currentPassword, new_password: newPassword }),
            });
            if (response.ok) {
                alert('Пароль успешно изменён');
                modal.remove();
            }
            else {
                const error = await response.json().catch(() => ({ error: 'Ошибка при смене пароля' }));
                alert(`Ошибка: ${error.error}`);
            }
        }
        catch (err) {
            alert('Ошибка сети');
        }
    });
}
// === Добавление кнопки "Сменить пароль" в меню профиля ===
function setupUserMenu() {
    const header = document.querySelector('.header');
    if (!header)
        return;
    const userMenuContainer = document.createElement('div');
    userMenuContainer.className = 'user-menu';
    userMenuContainer.style.display = 'flex';
    userMenuContainer.style.alignItems = 'center';
    userMenuContainer.style.gap = '12px';
    const userNameSpan = document.createElement('span');
    userNameSpan.textContent = user?.name || 'Пользователь';
    userNameSpan.style.color = '#e0e0e0';
    userNameSpan.style.fontSize = '14px';
    const changePasswordBtn = document.createElement('button');
    changePasswordBtn.className = 'btn-small';
    changePasswordBtn.textContent = 'Сменить пароль';
    changePasswordBtn.style.fontSize = '12px';
    changePasswordBtn.style.padding = '6px 10px';
    changePasswordBtn.addEventListener('click', openChangePasswordModal);
    const logoutBtn = document.createElement('button');
    logoutBtn.className = 'logout-btn';
    logoutBtn.textContent = 'Выйти';
    logoutBtn.id = 'logoutBtn';
    userMenuContainer.appendChild(userNameSpan);
    userMenuContainer.appendChild(changePasswordBtn);
    userMenuContainer.appendChild(logoutBtn);
    header.appendChild(userMenuContainer);
}
// === Получение задач (упрощённо) ===
async function fetchTasks() {
    try {
        token = getAuthToken();
        const res = await fetch("/api/v1/tasks", {
            headers: { "Authorization": `Bearer ${token}` },
        });
        if (res.ok)
            return await res.json();
        return [];
    }
    catch {
        return [];
    }
}
// === Получение токена ===
function getAuthToken() {
    return localStorage.getItem("token") || "";
}
async function downloadExcelExport(endpoint, fallbackFilename, trigger) {
    const activeToken = getAuthToken();
    if (!activeToken) {
        window.location.href = '/login';
        return;
    }
    const originalText = trigger?.textContent ?? '';
    if (trigger) {
        trigger.disabled = true;
        trigger.textContent = 'Скачивание...';
    }
    try {
        const response = await fetch(endpoint, {
            headers: { Authorization: `Bearer ${activeToken}` },
        });
        if (response.status === 401) {
            localStorage.removeItem('token');
            document.cookie = 'token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
            window.location.href = '/login';
            return;
        }
        if (!response.ok) {
            let message = 'Не удалось скачать файл';
            try {
                const errorData = await response.json();
                if (typeof errorData?.error === 'string' && errorData.error.trim()) {
                    message = errorData.error;
                }
            }
            catch {
                // Игнорируем, если сервер вернул не JSON.
            }
            throw new Error(message);
        }
        const blob = await response.blob();
        const contentDisposition = response.headers.get('Content-Disposition') || '';
        const filenameMatch = contentDisposition.match(/filename="?([^"]+)"?/i);
        const filename = filenameMatch?.[1] || fallbackFilename;
        const downloadUrl = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = downloadUrl;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        link.remove();
        URL.revokeObjectURL(downloadUrl);
    }
    catch (error) {
        console.error('Ошибка скачивания экспорта:', error);
        alert(error instanceof Error ? error.message : 'Не удалось скачать файл');
    }
    finally {
        if (trigger) {
            trigger.disabled = false;
            trigger.textContent = originalText;
        }
    }
}
// === Загрузка данных ===
async function getCurrentUser() {
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
        localStorage.setItem('user', JSON.stringify(user));
    }
    catch (err) {
        console.error('Не удалось определить текущего пользователя');
        user = null;
    }
}
// === Вспомогательная функция для извлечения email из JWT ===
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
// === Получение проектов пользователя ===
async function fetchUserProjects(userID) {
    if (!userID)
        return [];
    try {
        const token = getAuthToken();
        const res = await fetch(`/api/v1/projects/by-user?user_id=${userID}`, {
            headers: { "Authorization": `Bearer ${token}` },
        });
        if (res.ok)
            return await res.json();
        return [];
    }
    catch {
        return [];
    }
}
async function fetchVisibleProjects() {
    try {
        const token = getAuthToken();
        const res = await fetch('/api/v1/projects', {
            headers: { "Authorization": `Bearer ${token}` },
        });
        if (res.ok)
            return await res.json();
        return [];
    }
    catch {
        return [];
    }
}
// === Проверка роли ===
function getUserRole() {
    const token = getAuthToken();
    if (!token)
        return "guest";
    try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        return payload.role || "user";
    }
    catch {
        return "user";
    }
}
// === Рендер меню ===
async function renderMenu() {
    const menuEl = document.getElementById("sidebarMenu");
    if (!menuEl)
        return;
    await getCurrentUser();
    const userRole = getUserRole();
    const fragment = document.createDocumentFragment();
    for (const item of menuItems) {
        // Проверяем роль
        if (!item.roles.includes(userRole) && !item.visible)
            continue;
        // Проверяем условие видимости
        if (item.visible) {
            let isVisible = false;
            try {
                isVisible = await Promise.resolve(item.visible());
            }
            catch (err) {
                console.error("Failed to evaluate visibility:", err);
                isVisible = false;
            }
            if (!isVisible)
                continue;
        }
        const li = document.createElement("li");
        const a = document.createElement("a");
        a.href = item.path;
        // Создаём иконку Material Icons
        const iconSpan = document.createElement('span');
        iconSpan.className = 'material-icons';
        iconSpan.style.marginRight = '12px';
        iconSpan.style.fontSize = '20px';
        iconSpan.style.display = 'flex';
        iconSpan.style.alignItems = 'center';
        iconSpan.style.justifyContent = 'center';
        iconSpan.textContent = item.icon;
        // Создаём текст
        const textSpan = document.createElement('span');
        textSpan.textContent = item.title;
        // Добавляем элементы в ссылку
        a.appendChild(iconSpan);
        a.appendChild(textSpan);
        if (window.location.pathname === item.path) {
            a.classList.add("active");
        }
        if (item.badge) {
            const badge = document.createElement("span");
            badge.classList.add("badge");
            badge.classList.add("hidden");
            a.appendChild(badge);
            try {
                const count = await item.badge();
                if (count !== null) {
                    badge.textContent = count.toString();
                    badge.classList.remove("hidden");
                }
            }
            catch (err) {
                console.error("Failed to load badge:", err);
            }
        }
        li.appendChild(a);
        fragment.appendChild(li);
    }
    menuEl.appendChild(fragment);
}
// === Кнопка сворачивания меню ===
// === Кнопка сворачивания меню ===
function setupToggleSidebar() {
    const toggleBtn = document.createElement('button');
    toggleBtn.classList.add('toggle-sidebar-btn');
    toggleBtn.textContent = 'menu';
    toggleBtn.title = 'Открыть меню';
    toggleBtn.setAttribute('aria-label', 'Toggle menu');
    toggleBtn.style.fontFamily = 'Material Icons';
    const header = document.querySelector('.header');
    if (header) {
        header.prepend(toggleBtn);
    }
    const sidebar = document.querySelector('.sidebar');
    const layout = document.querySelector('.main-layout');
    if (!sidebar || !layout)
        return;
    // Функция для обновления отображения текста
    const updateMenuDisplay = (collapsed) => {
        const menuLinks = document.querySelectorAll('.sidebar-menu a');
        menuLinks.forEach(link => {
            const textSpan = link.querySelector('span:last-child');
            const iconSpan = link.querySelector('span:first-child');
            if (textSpan && iconSpan) {
                textSpan.style.display = collapsed ? 'none' : 'inline';
                iconSpan.style.marginRight = collapsed ? '0' : '12px';
            }
        });
    };
    const updateLayout = () => {
        const isMobile = window.innerWidth <= 768;
        const isCollapsed = sidebar.classList.contains('collapsed');
        if (isMobile) {
            // На мобильных: меню поверх контента
            sidebar.classList.toggle('collapsed');
            toggleBtn.textContent = sidebar.classList.contains('collapsed') ? 'menu' : 'close';
            toggleBtn.title = sidebar.classList.contains('collapsed') ? 'Открыть меню' : 'Закрыть меню';
            // Всегда показываем полный текст на мобильных
            updateMenuDisplay(false);
        }
        else {
            // На ПК: скрываем текст и сдвигаем контент
            sidebar.classList.toggle('collapsed');
            layout.classList.toggle('sidebar-collapsed');
            // Обновляем отображение текста
            const nextCollapsed = sidebar.classList.contains('collapsed');
            localStorage.setItem('sidebarCollapsed', String(nextCollapsed));
            updateMenuDisplay(nextCollapsed);
            toggleBtn.textContent = 'menu';
            toggleBtn.title = nextCollapsed ? 'Развернуть меню' : 'Свернуть меню';
        }
    };
    toggleBtn.addEventListener('click', updateLayout);
    // Обработка изменения размера окна
    window.addEventListener('resize', () => {
        const isMobile = window.innerWidth <= 768;
        const isCollapsed = sidebar.classList.contains('collapsed');
        if (isMobile) {
            layout.classList.remove('sidebar-collapsed');
            if (!isCollapsed) {
                toggleBtn.textContent = 'menu';
                toggleBtn.title = 'Открыть меню';
            }
        }
        else {
            if (isCollapsed) {
                layout.classList.add('sidebar-collapsed');
                updateMenuDisplay(true);
                toggleBtn.textContent = 'menu';
                toggleBtn.title = 'Развернуть меню';
            }
            else {
                layout.classList.remove('sidebar-collapsed');
                updateMenuDisplay(false);
                toggleBtn.textContent = 'menu';
                toggleBtn.title = 'Свернуть меню';
            }
        }
    });
    // Начальная инициализация
    const isMobile = window.innerWidth <= 768;
    const saved = localStorage.getItem('sidebarCollapsed');
    if (isMobile) {
        sidebar.classList.add('collapsed');
        toggleBtn.textContent = 'menu';
        toggleBtn.title = 'Открыть меню';
        updateMenuDisplay(false);
    }
    else {
        if (saved === 'true') {
            sidebar.classList.add('collapsed');
            layout.classList.add('sidebar-collapsed');
            updateMenuDisplay(true);
            toggleBtn.textContent = 'menu';
            toggleBtn.title = 'Развернуть меню';
        }
        else {
            updateMenuDisplay(false);
            toggleBtn.textContent = 'menu';
            toggleBtn.title = 'Свернуть меню';
        }
    }
}
function setupLogout() {
    const btn = document.getElementById("logoutBtn");
    btn?.addEventListener("click", () => {
        localStorage.removeItem("token");
        document.cookie = "token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;";
        window.location.href = "/login";
    });
}
// === Инициализация ===
document.addEventListener("DOMContentLoaded", () => {
    token = localStorage.getItem('token');
    if (!token) {
        window.location.href = '/login';
        return;
    }
    renderMenu().then(() => {
        setupToggleSidebar();
    });
    getCurrentUser().then(() => {
        setupUserMenu(); // Отображаем после загрузки пользователя
    }).then(() => {
        setupLogout();
    });
});
window.downloadExcelExport = downloadExcelExport;
export {};
