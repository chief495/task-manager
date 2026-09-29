// labor.ts — калькулятор трудозатрат, привязанный к проекту

if (window.location.pathname !== "/labor") {
    throw new Error("labor.ts loaded outside /labor");
}

// ─── Интерфейсы ──────────────────────────────────────────────────────────────

interface LaborCategory {
    id: number;
    name: string;
    code: string;
    description: string;
}

interface LaborNorm {
    id: number;
    category_id: number;
    code: string;
    name: string;
    unit: string;
    labor_hours: number;
    description: string;
    source: string;
    category_name?: string;
}

interface LaborCoefficient {
    id: number;
    name: string;
    code: string;
    value: number;
    category_id: number | null;
    description: string;
}

interface LaborItemCoefficient {
    coefficient_id: number;
    value: number;
}

interface LaborCalculationItemRequest {
    norm_id: number;
    quantity: number;
    custom_coeff: number;
    coefficients: LaborItemCoefficient[];
}

interface LaborGlobalCoefficientRequest {
    category_id: number;
    value: number;
    name: string;
}

interface LaborCalculationRequest {
    project_id: number;
    name: string;
    items: LaborCalculationItemRequest[];
    global_coefficients: LaborGlobalCoefficientRequest[];
}

interface LaborCalculationResultItem {
    norm_code: string;
    norm_name: string;
    category_name: string;
    unit: string;
    quantity: number;
    base_labor_hours: number;
    coefficient_mult: number;
    total_labor_hours: number;
    total_man_days: number;
}

interface LaborCategoryTotal {
    category_name: string;
    total_labor_hours: number;
    total_man_days: number;
    global_coeff: number;
    adjusted_hours: number;
    adjusted_days: number;
}

interface LaborCalculationResult {
    id: number;
    project_id: number;
    project_name: string;
    name: string;
    items: LaborCalculationResultItem[];
    category_totals: LaborCategoryTotal[];
    grand_total: number;
    calculated_at: string;
}

interface LaborCalculationSummary {
    id: number;
    name: string;
    grand_total: number;
    items_count: number;
    created_by: string;
    created_at: string;
}

// ─── Константы и состояние ───────────────────────────────────────────────────

const LABOR_API      = "/api/v1/labor";
const EXPORT_EXCEL_URL = "/api/v1/export/labor/excel";
const EXPORT_PDF_URL   = "/api/v1/export/labor/pdf";

let token        = "";
let projectId    = 0;
let projectName  = "";
let categories: LaborCategory[]   = [];
let norms: LaborNorm[]            = [];
let coefficients: LaborCoefficient[] = [];
let lastResult: LaborCalculationResult | null = null;

const content = document.getElementById("content") as HTMLElement | null;

// ─── Точка входа ─────────────────────────────────────────────────────────────

document.addEventListener("DOMContentLoaded", () => {
    if (window.location.pathname !== "/labor") return;
    if (!content) { console.error("Не найден #content"); return; }

    token = localStorage.getItem("token") || "";

    // Читаем project_id из URL: /labor?project_id=5
    const params = new URLSearchParams(window.location.search);
    projectId = parseInt(params.get("project_id") || "0", 10);

    if (!projectId) {
        content.innerHTML = `
            <div style="padding:40px;text-align:center;color:#f87171">
                <h2>Проект не указан</h2>
                <p>Откройте калькулятор из карточки проекта.</p>
                <a href="/admin/projects" style="color:#6a9bff">← К проектам</a>
            </div>`;
        return;
    }

    void initLaborPage();
});

async function initLaborPage(): Promise<void> {
    content!.style.visibility = "hidden";

    try {
        // Загружаем название проекта
        await loadProjectName();

        renderPage();
        bindStaticEvents();

        await Promise.all([
            loadCategories(),
            loadNorms(),
            loadCoefficients(),
        ]);

        buildGlobalCoeffGrid();
        addCalcItem();
        await loadCalculationHistory();
    } catch (err) {
        console.error("Ошибка инициализации labor:", err);
        showToast("Ошибка загрузки страницы", "error");
    } finally {
        content!.style.visibility = "visible";
    }
}

// ─── Загрузка названия проекта ───────────────────────────────────────────────

async function loadProjectName(): Promise<void> {
    try {
        const res = await fetch(`/api/v1/projects/${projectId}`, {
            headers: buildHeaders(false),
            credentials: "same-origin",
        });
        if (res.ok) {
            const project = await res.json();
            projectName = project.name || `Проект #${projectId}`;
        } else {
            projectName = `Проект #${projectId}`;
        }
    } catch {
        projectName = `Проект #${projectId}`;
    }
}

// ─── Рендер страницы ─────────────────────────────────────────────────────────

function renderPage(): void {
    content!.innerHTML = `
        <link rel="stylesheet" href="/static/css/labor.css">

        <div class="labor-page">

            <div class="labor-toolbar">
                <div class="labor-title-block">
                    <div class="labor-breadcrumb">
                        <a href="/admin/projects" class="labor-breadcrumb-link">← Проекты</a>
                        <span class="labor-breadcrumb-sep">/</span>
                        <span class="labor-breadcrumb-current">${esc(projectName)}</span>
                    </div>
                    <h2>Калькулятор трудозатрат</h2>
                </div>
                <div class="labor-toolbar-actions">
                    <button class="btn btn-outline btn-sm" id="exportExcelBtn">Excel</button>
                    <button class="btn btn-outline btn-sm" id="exportPdfBtn">PDF</button>
                </div>
            </div>

            <div class="labor-tabs">
                <button class="labor-tab active" data-tab="calculator">Расчёт</button>
                <button class="labor-tab" data-tab="history">История расчётов</button>
                <button class="labor-tab" data-tab="norms">Нормы</button>
                <button class="labor-tab" data-tab="categories">Категории</button>
                <button class="labor-tab" data-tab="coefficients">Коэффициенты</button>
            </div>

            <div id="toastContainer" class="toast-container"></div>

            <!-- ══ TAB: CALCULATOR ══ -->
            <div class="labor-tab-content active" id="tab-calculator">

                <div class="card">
                    <div class="card-header">Параметры расчёта</div>
                    <div class="card-body">
                        <div class="labor-params-grid">
                            <div class="form-group">
                                <label>Проект</label>
                                <input class="form-control" type="text"
                                       value="${esc(projectName)}" disabled>
                            </div>
                            <div class="form-group">
                                <label>Название расчёта</label>
                                <input id="calcName" class="form-control" type="text"
                                       placeholder="Например: Этап 1 — проектирование">
                            </div>
                        </div>
                    </div>
                </div>

                <div class="card">
                    <div class="card-header">Глобальные коэффициенты по категориям</div>
                    <div class="card-body">
                        <div id="globalCoeffGrid" class="global-coeff-grid"></div>
                    </div>
                </div>

                <div class="card">
                    <div class="card-header card-header-with-action">
                        <span>Позиции расчёта</span>
                        <button class="btn btn-success btn-sm" id="addCalcItemBtn">
                            Добавить позицию
                        </button>
                    </div>
                    <div class="card-body">
                        <div id="calcItems"></div>
                        <div id="emptyCalcState" class="empty-state">
                            Нет позиций. Нажмите «Добавить позицию».
                        </div>
                        <div class="calc-actions">
                            <button class="btn btn-primary" id="calculateBtn">
                                Рассчитать и сохранить
                            </button>
                        </div>
                    </div>
                </div>

                <div id="resultsContainer" style="display:none">
                    <div class="card">
                        <div class="card-header card-header-with-action">
                            <span>Результаты расчёта</span>
                            <span id="resultCalcName" class="labor-result-name"></span>
                        </div>
                        <div class="card-body">
                            <div id="resultSummary" class="result-summary"></div>

                            <div class="table-wrapper">
                                <table>
                                    <thead>
                                        <tr>
                                            <th>№</th><th>Код</th><th>Наименование</th>
                                            <th>Категория</th><th>Ед. изм.</th><th>Кол-во</th>
                                            <th>База (чел.-ч)</th><th>Коэфф.</th>
                                            <th>Итого (чел.-ч)</th><th>Итого (чел.-дн.)</th>
                                        </tr>
                                    </thead>
                                    <tbody id="resultTableBody"></tbody>
                                </table>
                            </div>

                            <h3 style="margin:20px 0 12px">Итого по категориям</h3>
                            <div class="table-wrapper">
                                <table>
                                    <thead>
                                        <tr>
                                            <th>Категория</th><th>Чел.-часы</th>
                                            <th>Глоб. коэфф.</th>
                                            <th>С коэфф. (чел.-ч)</th>
                                            <th>С коэфф. (чел.-дн.)</th>
                                        </tr>
                                    </thead>
                                    <tbody id="categoryTotalsBody"></tbody>
                                </table>
                            </div>

                            <div class="export-bar">
                                <button class="btn btn-success" id="exportExcelResultBtn">
                                    Экспорт в Excel
                                </button>
                                <button class="btn btn-danger" id="exportPdfResultBtn">
                                    Экспорт в PDF
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- ══ TAB: HISTORY ══ -->
            <div class="labor-tab-content" id="tab-history">
                <div class="card">
                    <div class="card-header">История расчётов — ${esc(projectName)}</div>
                    <div class="card-body">
                        <div class="table-wrapper">
                            <table>
                                <thead>
                                    <tr>
                                        <th>Название</th>
                                        <th>Итого (чел.-ч)</th>
                                        <th>Позиций</th>
                                        <th>Автор</th>
                                        <th>Дата</th>
                                        <th>Действия</th>
                                    </tr>
                                </thead>
                                <tbody id="historyTableBody"></tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>

            <!-- ══ TAB: NORMS ══ -->
            <div class="labor-tab-content" id="tab-norms">
                <div class="card">
                    <div class="card-header card-header-with-action">
                        <span>Нормы трудозатрат</span>
                        <button class="btn btn-success btn-sm" id="addNormBtn">Добавить норму</button>
                    </div>
                    <div class="card-body">
                        <div class="search-bar">
                            <input id="normSearch" class="form-control" type="text"
                                   placeholder="Поиск по коду или названию">
                            <select id="normCategoryFilter" class="form-control">
                                <option value="">Все категории</option>
                            </select>
                        </div>
                        <div class="table-wrapper">
                            <table>
                                <thead>
                                    <tr>
                                        <th>Код</th><th>Наименование</th><th>Категория</th>
                                        <th>Ед. изм.</th><th>Норма</th><th>Источник</th>
                                        <th>Действия</th>
                                    </tr>
                                </thead>
                                <tbody id="normsTableBody"></tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>

            <!-- ══ TAB: CATEGORIES ══ -->
            <div class="labor-tab-content" id="tab-categories">
                <div class="card">
                    <div class="card-header card-header-with-action">
                        <span>Категории</span>
                        <button class="btn btn-success btn-sm" id="addCategoryBtn">
                            Добавить категорию
                        </button>
                    </div>
                    <div class="card-body">
                        <div class="table-wrapper">
                            <table>
                                <thead>
                                    <tr>
                                        <th>Код</th><th>Название</th>
                                        <th>Описание</th><th>Действия</th>
                                    </tr>
                                </thead>
                                <tbody id="categoriesTableBody"></tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>

            <!-- ══ TAB: COEFFICIENTS ══ -->
            <div class="labor-tab-content" id="tab-coefficients">
                <div class="card">
                    <div class="card-header card-header-with-action">
                        <span>Коэффициенты</span>
                        <button class="btn btn-success btn-sm" id="addCoeffBtn">
                            Добавить коэффициент
                        </button>
                    </div>
                    <div class="card-body">
                        <div class="table-wrapper">
                            <table>
                                <thead>
                                    <tr>
                                        <th>Код</th><th>Название</th><th>Значение</th>
                                        <th>Категория</th><th>Описание</th><th>Действия</th>
                                    </tr>
                                </thead>
                                <tbody id="coefficientsTableBody"></tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>

            <!-- ══ MODALS ══ -->

            <div class="modal-overlay" id="categoryModal">
                <div class="modal">
                    <div class="modal-header">
                        <h3 id="categoryModalTitle">Добавить категорию</h3>
                        <button class="modal-close" data-close-modal="categoryModal">&times;</button>
                    </div>
                    <div class="modal-body">
                        <input type="hidden" id="categoryEditId">
                        <div class="form-group">
                            <label>Код</label>
                            <input type="text" id="categoryCode" class="form-control">
                        </div>
                        <div class="form-group">
                            <label>Название</label>
                            <input type="text" id="categoryName" class="form-control">
                        </div>
                        <div class="form-group">
                            <label>Описание</label>
                            <textarea id="categoryDescription" class="form-control"></textarea>
                        </div>
                    </div>
                    <div class="modal-footer">
                        <button class="btn btn-primary" id="saveCategoryBtn">Сохранить</button>
                    </div>
                </div>
            </div>

            <div class="modal-overlay" id="normModal">
                <div class="modal">
                    <div class="modal-header">
                        <h3 id="normModalTitle">Добавить норму</h3>
                        <button class="modal-close" data-close-modal="normModal">&times;</button>
                    </div>
                    <div class="modal-body">
                        <input type="hidden" id="normEditId">
                        <div class="form-group">
                            <label>Код</label>
                            <input type="text" id="normCode" class="form-control">
                        </div>
                        <div class="form-group">
                            <label>Категория</label>
                            <select id="normCategoryId" class="form-control"></select>
                        </div>
                        <div class="form-group">
                            <label>Наименование</label>
                            <input type="text" id="normName" class="form-control">
                        </div>
                        <div class="form-group">
                            <label>Единица измерения</label>
                            <input type="text" id="normUnit" class="form-control" value="чел.-ч">
                        </div>
                        <div class="form-group">
                            <label>Норма трудозатрат (чел.-ч)</label>
                            <input type="number" step="0.01" id="normLaborHours" class="form-control">
                        </div>
                        <div class="form-group">
                            <label>Источник</label>
                            <input type="text" id="normSource" class="form-control" value="ГЭСН">
                        </div>
                        <div class="form-group">
                            <label>Описание</label>
                            <textarea id="normDescription" class="form-control"></textarea>
                        </div>
                    </div>
                    <div class="modal-footer">
                        <button class="btn btn-primary" id="saveNormBtn">Сохранить</button>
                    </div>
                </div>
            </div>

            <div class="modal-overlay" id="coeffModal">
                <div class="modal">
                    <div class="modal-header">
                        <h3 id="coeffModalTitle">Добавить коэффициент</h3>
                        <button class="modal-close" data-close-modal="coeffModal">&times;</button>
                    </div>
                    <div class="modal-body">
                        <input type="hidden" id="coeffEditId">
                        <div class="form-group">
                            <label>Код</label>
                            <input type="text" id="coeffCode" class="form-control">
                        </div>
                        <div class="form-group">
                            <label>Название</label>
                            <input type="text" id="coeffName" class="form-control">
                        </div>
                        <div class="form-group">
                            <label>Значение</label>
                            <input type="number" step="0.01" id="coeffValue" class="form-control" value="1.0">
                        </div>
                        <div class="form-group">
                            <label>Категория</label>
                            <select id="coeffCategoryId" class="form-control">
                                <option value="0">Общий (все категории)</option>
                            </select>
                        </div>
                        <div class="form-group">
                            <label>Описание</label>
                            <textarea id="coeffDescription" class="form-control"></textarea>
                        </div>
                    </div>
                    <div class="modal-footer">
                        <button class="btn btn-primary" id="saveCoeffBtn">Сохранить</button>
                    </div>
                </div>
            </div>

        </div><!-- /labor-page -->
    `;
}

// ─── Привязка событий ─────────────────────────────────────────────────────────

function bindStaticEvents(): void {
    document.querySelectorAll<HTMLElement>(".labor-tab").forEach((tab) => {
        tab.addEventListener("click", () => switchTab(tab.dataset.tab || "calculator"));
    });

    getById<HTMLButtonElement>("addCalcItemBtn")?.addEventListener("click", addCalcItem);
    getById<HTMLButtonElement>("calculateBtn")?.addEventListener("click", () => void calculate());

    getById<HTMLButtonElement>("exportExcelBtn")?.addEventListener("click", () => void exportExcel());
    getById<HTMLButtonElement>("exportPdfBtn")?.addEventListener("click", () => void exportPDF());
    getById<HTMLButtonElement>("exportExcelResultBtn")?.addEventListener("click", () => void exportExcel());
    getById<HTMLButtonElement>("exportPdfResultBtn")?.addEventListener("click", () => void exportPDF());

    getById<HTMLInputElement>("normSearch")?.addEventListener("input", renderNorms);
    getById<HTMLSelectElement>("normCategoryFilter")?.addEventListener("change", renderNorms);

    getById<HTMLButtonElement>("addCategoryBtn")?.addEventListener("click", showCategoryModal);
    getById<HTMLButtonElement>("saveCategoryBtn")?.addEventListener("click", () => void saveCategory());

    getById<HTMLButtonElement>("addNormBtn")?.addEventListener("click", showNormModal);
    getById<HTMLButtonElement>("saveNormBtn")?.addEventListener("click", () => void saveNorm());

    getById<HTMLButtonElement>("addCoeffBtn")?.addEventListener("click", showCoeffModal);
    getById<HTMLButtonElement>("saveCoeffBtn")?.addEventListener("click", () => void saveCoefficient());

    getById("categoriesTableBody")?.addEventListener("click", onCategoriesTableClick);
    getById("normsTableBody")?.addEventListener("click", onNormsTableClick);
    getById("coefficientsTableBody")?.addEventListener("click", onCoefficientsTableClick);
    getById("historyTableBody")?.addEventListener("click", onHistoryTableClick);

    getById("calcItems")?.addEventListener("change", onCalcItemsChange);
    getById("calcItems")?.addEventListener("click", onCalcItemsClick);

    document.addEventListener("click", onDocumentClick);

    // Запрет прокрутки числовых полей
    content!.addEventListener("wheel", (e) => {
        const t = e.target as HTMLElement;
        if (t.matches("input[type='number']") && document.activeElement === t) {
            e.preventDefault();
            (t as HTMLInputElement).blur();
        }
    }, { passive: false });
}

// ─── Вкладки ─────────────────────────────────────────────────────────────────

function switchTab(name: string): void {
    document.querySelectorAll(".labor-tab").forEach(t => t.classList.remove("active"));
    document.querySelectorAll(".labor-tab-content").forEach(c => c.classList.remove("active"));
    document.querySelector(`.labor-tab[data-tab="${name}"]`)?.classList.add("active");
    getById(`tab-${name}`)?.classList.add("active");
}

// ─── Загрузка данных ─────────────────────────────────────────────────────────

async function loadCategories(): Promise<void> {
    try {
        categories = (await apiGet<LaborCategory[]>("/categories")) ?? [];
        renderCategories();
        updateCategorySelects();
    } catch (e) {
        showToast("Ошибка загрузки категорий", "error");
        console.error(e);
    }
}

async function loadNorms(): Promise<void> {
    try {
        norms = (await apiGet<LaborNorm[]>("/norms")) ?? [];
        renderNorms();
        refreshCalcNormSelects();
    } catch (e) {
        showToast("Ошибка загрузки норм", "error");
        console.error(e);
    }
}

async function loadCoefficients(): Promise<void> {
    try {
        coefficients = (await apiGet<LaborCoefficient[]>("/coefficients")) ?? [];
        renderCoefficients();
        refreshAllCalcCoeffBlocks();
    } catch (e) {
        showToast("Ошибка загрузки коэффициентов", "error");
        console.error(e);
    }
}

async function loadCalculationHistory(): Promise<void> {
    try {
        const data = await apiGet<LaborCalculationSummary[]>(
            `/calculations?project_id=${projectId}`
        );
        renderHistory(data ?? []);
    } catch (e) {
        console.error("Ошибка загрузки истории расчётов:", e);
    }
}

// ─── История расчётов ────────────────────────────────────────────────────────

function renderHistory(list: LaborCalculationSummary[]): void {
    const tbody = getById<HTMLTableSectionElement>("historyTableBody");
    if (!tbody) return;

    if (!list.length) {
        tbody.innerHTML = `
            <tr>
                <td colspan="6" style="text-align:center;color:#666;padding:24px">
                    Расчётов пока нет. Создайте первый на вкладке «Расчёт».
                </td>
            </tr>`;
        return;
    }

    tbody.innerHTML = list.map(c => `
        <tr>
            <td>${esc(c.name)}</td>
            <td><strong>${format2(c.grand_total)}</strong></td>
            <td>${c.items_count}</td>
            <td>${esc(c.created_by)}</td>
            <td>${esc(c.created_at)}</td>
            <td>
                <div class="btn-group">
                    <button class="btn btn-primary btn-sm"
                            data-action="view-calc" data-id="${c.id}">
                        Открыть
                    </button>
                    <button class="btn btn-success btn-sm"
                            data-action="export-calc-excel" data-id="${c.id}">
                        Excel
                    </button>
                    <button class="btn btn-danger btn-sm"
                            data-action="delete-calc" data-id="${c.id}">
                        Удалить
                    </button>
                </div>
            </td>
        </tr>
    `).join("");
}

function onHistoryTableClick(event: Event): void {
    const btn = (event.target as HTMLElement).closest("button[data-action]") as HTMLButtonElement | null;
    if (!btn) return;

    const id   = parseInt(btn.dataset.id || "0", 10);
    const action = btn.dataset.action;

    if (action === "view-calc")         void viewCalculation(id);
    if (action === "export-calc-excel") void exportCalculationExcel(id);
    if (action === "delete-calc")       void deleteCalculation(id);
}

async function viewCalculation(id: number): Promise<void> {
    try {
        const result = await fetchCalculationResult(id);
        if (!result) return;
        lastResult = result;
        renderResults(result);
        switchTab("calculator");
        getById("resultsContainer")?.scrollIntoView({ behavior: "smooth" });
        showToast(`Открыт расчёт: ${result.name}`);
    } catch (e) {
        console.error(e);
        showToast("Ошибка загрузки расчёта", "error");
    }
}

async function exportCalculationExcel(id: number): Promise<void> {
    try {
        const result = await fetchCalculationResult(id);
        if (!result) return;
        const blob = await exportFile(EXPORT_EXCEL_URL, result);
        downloadBlob(blob, `labor_${sanitizeFileName(result.name)}.xlsx`);
        showToast("Excel файл загружен");
    } catch (e) {
        console.error(e);
        showToast("Ошибка экспорта", "error");
    }
}

async function deleteCalculation(id: number): Promise<void> {
    if (!window.confirm("Удалить расчёт? Это действие необратимо.")) return;
    try {
        await apiDelete(`/calculations/${id}`);
        showToast("Расчёт удалён");
        await loadCalculationHistory();
    } catch (e) {
        console.error(e);
        showToast("Ошибка удаления расчёта", "error");
    }
}

async function fetchCalculationResult(id: number): Promise<LaborCalculationResult | null> {
    const res = await fetch(`${LABOR_API}/calculations/${id}`, {
        headers: buildHeaders(false),
        credentials: "same-origin",
    });
    if (res.status === 401) { redirectToLogin(); return null; }
    if (!res.ok) throw new Error("failed to load calculation");
    return res.json() as Promise<LaborCalculationResult>;
}

// ─── Рендер справочников ─────────────────────────────────────────────────────

function renderCategories(): void {
    const tbody = getById<HTMLTableSectionElement>("categoriesTableBody");
    if (!tbody) return;

    if (!categories.length) {
        tbody.innerHTML = `<tr><td colspan="4">Нет категорий</td></tr>`;
        return;
    }

    tbody.innerHTML = categories.map(c => `
        <tr>
            <td>${esc(c.code)}</td>
            <td>${esc(c.name)}</td>
            <td>${esc(c.description)}</td>
            <td>
                <div class="btn-group">
                    <button class="btn btn-warning btn-sm"
                            data-action="edit-category" data-id="${c.id}">Изменить</button>
                    <button class="btn btn-danger btn-sm"
                            data-action="delete-category" data-id="${c.id}">Удалить</button>
                </div>
            </td>
        </tr>
    `).join("");
}

function renderNorms(): void {
    const tbody  = getById<HTMLTableSectionElement>("normsTableBody");
    if (!tbody) return;

    const search  = (getById<HTMLInputElement>("normSearch")?.value || "").trim().toLowerCase();
    const catF    = getById<HTMLSelectElement>("normCategoryFilter")?.value || "";

    let list = [...norms];
    if (catF)   list = list.filter(n => String(n.category_id) === catF);
    if (search) list = list.filter(n =>
        n.code.toLowerCase().includes(search) ||
        n.name.toLowerCase().includes(search)
    );

    if (!list.length) {
        tbody.innerHTML = `<tr><td colspan="7">Нормы не найдены</td></tr>`;
        return;
    }

    tbody.innerHTML = list.map(n => `
        <tr>
            <td>${esc(n.code)}</td>
            <td>${esc(n.name)}</td>
            <td>${esc(n.category_name || getCategoryName(n.category_id))}</td>
            <td>${esc(n.unit)}</td>
            <td>${format2(n.labor_hours)}</td>
            <td>${esc(n.source)}</td>
            <td>
                <div class="btn-group">
                    <button class="btn btn-warning btn-sm"
                            data-action="edit-norm" data-id="${n.id}">Изменить</button>
                    <button class="btn btn-danger btn-sm"
                            data-action="delete-norm" data-id="${n.id}">Удалить</button>
                </div>
            </td>
        </tr>
    `).join("");
}

function renderCoefficients(): void {
    const tbody = getById<HTMLTableSectionElement>("coefficientsTableBody");
    if (!tbody) return;

    if (!coefficients.length) {
        tbody.innerHTML = `<tr><td colspan="6">Нет коэффициентов</td></tr>`;
        return;
    }

    tbody.innerHTML = coefficients.map(c => `
        <tr>
            <td>${esc(c.code)}</td>
            <td>${esc(c.name)}</td>
            <td>${format3(c.value)}</td>
            <td>${c.category_id ? esc(getCategoryName(c.category_id)) : "Общий"}</td>
            <td>${esc(c.description)}</td>
            <td>
                <div class="btn-group">
                    <button class="btn btn-warning btn-sm"
                            data-action="edit-coeff" data-id="${c.id}">Изменить</button>
                    <button class="btn btn-danger btn-sm"
                            data-action="delete-coeff" data-id="${c.id}">Удалить</button>
                </div>
            </td>
        </tr>
    `).join("");
}

function updateCategorySelects(): void {
    const selects: Array<{ id: string; prefix: string }> = [
        { id: "normCategoryId",     prefix: "" },
        { id: "normCategoryFilter", prefix: "" },
        { id: "coeffCategoryId",    prefix: "" },
    ];

    selects.forEach(({ id }) => {
        const el = getById<HTMLSelectElement>(id);
        if (!el) return;
        const cur = el.value;

        if (id === "normCategoryFilter") {
            el.innerHTML = `<option value="">Все категории</option>`;
        } else if (id === "coeffCategoryId") {
            el.innerHTML = `<option value="0">Общий (все категории)</option>`;
        } else {
            el.innerHTML = "";
        }

        categories.forEach(c => {
            el.innerHTML += `<option value="${c.id}">${esc(c.code)} — ${esc(c.name)}</option>`;
        });

        el.value = cur || (id === "coeffCategoryId" ? "0" : "");
    });
}

function buildGlobalCoeffGrid(): void {
    const grid = getById<HTMLDivElement>("globalCoeffGrid");
    if (!grid) return;

    grid.innerHTML = categories.map(c => `
        <div class="global-coeff-item">
            <label>${esc(c.name)}</label>
            <input type="number" step="0.01" min="0.1" max="10" value="1.00"
                   class="global-coeff-input" data-category-id="${c.id}">
        </div>
    `).join("");
}

// ─── Позиции расчёта ─────────────────────────────────────────────────────────

function addCalcItem(): void {
    const container  = getById<HTMLDivElement>("calcItems");
    const emptyState = getById<HTMLDivElement>("emptyCalcState");
    if (!container || !emptyState) return;

    const row = document.createElement("div");
    row.className = "calc-item";
    row.innerHTML = `
        <div class="calc-item-main">
            <div class="item-number"></div>
            <div class="calc-item-fields">
                <div class="form-group">
                    <label>Норма</label>
                    <select class="form-control labor-norm-select">
                        ${buildNormOptionsHtml("")}
                    </select>
                </div>
                <div class="form-group">
                    <label>Количество</label>
                    <input type="number" step="0.1" min="0.1" value="1"
                           class="form-control labor-qty-input">
                </div>
                <div class="form-group">
                    <label>Доп. коэфф.</label>
                    <input type="number" step="0.01" min="0.01" value="1.00"
                           class="form-control labor-custom-coeff-input">
                </div>
            </div>
        </div>
        <div class="coeff-selector">
            <span class="coeff-selector-title">Коэффициенты:</span>
            <div class="coeff-selector-body">
                <span class="muted">Сначала выберите норму</span>
            </div>
        </div>
        <div style="margin-top:10px">
            <button class="btn btn-danger btn-sm" data-action="remove-calc-item">
                Удалить
            </button>
        </div>
    `;

    container.appendChild(row);
    emptyState.style.display = "none";
    renumberCalcItems();
}

function renumberCalcItems(): void {
    const items = document.querySelectorAll<HTMLElement>("#calcItems .calc-item");
    items.forEach((item, i) => {
        const num = item.querySelector<HTMLElement>(".item-number");
        if (num) num.textContent = String(i + 1);
    });

    const es = getById<HTMLDivElement>("emptyCalcState");
    if (es) es.style.display = items.length ? "none" : "block";
}

function refreshCalcNormSelects(): void {
    document.querySelectorAll<HTMLSelectElement>(".labor-norm-select").forEach(sel => {
        const cur = sel.value;
        sel.innerHTML = buildNormOptionsHtml(cur);
    });
}

function refreshAllCalcCoeffBlocks(): void {
    document.querySelectorAll<HTMLElement>("#calcItems .calc-item").forEach(item => {
        renderCalcCoeffSelector(item);
    });
}

function buildNormOptionsHtml(selectedValue: string): string {
    const grouped = new Map<number, LaborNorm[]>();
    norms.forEach(n => {
        if (!grouped.has(n.category_id)) grouped.set(n.category_id, []);
        grouped.get(n.category_id)!.push(n);
    });

    let html = `<option value="">— Выберите норму —</option>`;
    categories.forEach(cat => {
        const group = grouped.get(cat.id) || [];
        if (!group.length) return;
        html += `<optgroup label="${esc(cat.name)}">`;
        group.forEach(n => {
            const sel = String(n.id) === selectedValue ? "selected" : "";
            html += `<option value="${n.id}" ${sel}>
                [${esc(n.code)}] ${esc(n.name)} (${n.labor_hours} ${esc(n.unit)})
            </option>`;
        });
        html += `</optgroup>`;
    });
    return html;
}

function renderCalcCoeffSelector(item: HTMLElement): void {
    const select = item.querySelector<HTMLSelectElement>(".labor-norm-select");
    const body   = item.querySelector<HTMLDivElement>(".coeff-selector-body");
    if (!select || !body) return;

    const normId = parseInt(select.value, 10);
    if (!normId) {
        body.innerHTML = `<span class="muted">Сначала выберите норму</span>`;
        return;
    }

    const norm = norms.find(n => n.id === normId);
    if (!norm) {
        body.innerHTML = `<span class="muted">Норма не найдена</span>`;
        return;
    }

    const applicable = coefficients.filter(c =>
        c.category_id === null || c.category_id === norm.category_id
    );

    if (!applicable.length) {
        body.innerHTML = `<span class="muted">Подходящих коэффициентов нет</span>`;
        return;
    }

    body.innerHTML = applicable.map(c => `
        <label class="coeff-chip" title="${esc(c.description)}">
            <input type="checkbox" data-coeff-id="${c.id}" data-coeff-value="${c.value}">
            <span>${esc(c.name)} (${format3(c.value)})</span>
        </label>
    `).join("");
}

// ─── Расчёт ──────────────────────────────────────────────────────────────────

async function calculate(): Promise<void> {
    const calcName = (getById<HTMLInputElement>("calcName")?.value || "").trim() ||
        `Расчёт от ${new Date().toLocaleDateString("ru-RU")}`;

    const items: LaborCalculationItemRequest[] = [];

    document.querySelectorAll<HTMLElement>("#calcItems .calc-item").forEach(item => {
        const normId = parseInt(
            item.querySelector<HTMLSelectElement>(".labor-norm-select")?.value || "0", 10
        );
        if (!normId) return;

        const quantity    = parseFloat(
            item.querySelector<HTMLInputElement>(".labor-qty-input")?.value || "1"
        ) || 1;
        const customCoeff = parseFloat(
            item.querySelector<HTMLInputElement>(".labor-custom-coeff-input")?.value || "1"
        ) || 1;

        const coeffs: LaborItemCoefficient[] = [];
        item.querySelectorAll<HTMLInputElement>(
            ".coeff-selector-body input[type='checkbox']:checked"
        ).forEach(cb => {
            coeffs.push({
                coefficient_id: parseInt(cb.dataset.coeffId || "0", 10),
                value:          parseFloat(cb.dataset.coeffValue || "1"),
            });
        });

        items.push({ norm_id: normId, quantity, custom_coeff: customCoeff, coefficients: coeffs });
    });

    if (!items.length) {
        showToast("Добавьте хотя бы одну позицию с выбранной нормой", "warning");
        return;
    }

    const globalCoeffs: LaborGlobalCoefficientRequest[] = [];
    document.querySelectorAll<HTMLInputElement>(".global-coeff-input").forEach(input => {
        const val = parseFloat(input.value || "1") || 1;
        if (val !== 1.0) {
            globalCoeffs.push({
                category_id: parseInt(input.dataset.categoryId || "0", 10),
                value: val,
                name: "",
            });
        }
    });

    const payload: LaborCalculationRequest = {
        project_id:          projectId,
        name:                calcName,
        items,
        global_coefficients: globalCoeffs,
    };

    const btn = getById<HTMLButtonElement>("calculateBtn");
    if (btn) { btn.disabled = true; btn.textContent = "Считаем..."; }

    try {
        // POST /api/v1/labor/calculate — сохраняет и возвращает результат
        const res = await fetch(`${LABOR_API}/calculate`, {
            method:      "POST",
            headers:     buildHeaders(true),
            credentials: "same-origin",
            body:        JSON.stringify(payload),
        });

        if (res.status === 401) { redirectToLogin(); return; }

        if (!res.ok) {
            const err = await res.json().catch(() => ({ error: "Ошибка сервера" }));
            throw new Error(err.error || "Ошибка сервера");
        }

        const result = (await res.json()) as LaborCalculationResult;
        lastResult = result;
        renderResults(result);
        showToast("Расчёт сохранён");

        // Обновляем историю в фоне
        void loadCalculationHistory();
    } catch (e) {
        console.error(e);
        showToast(`Ошибка расчёта: ${(e as Error).message}`, "error");
    } finally {
        if (btn) { btn.disabled = false; btn.textContent = "Рассчитать и сохранить"; }
    }
}

// ─── Рендер результатов ──────────────────────────────────────────────────────

function renderResults(result: LaborCalculationResult): void {
    const container  = getById<HTMLDivElement>("resultsContainer");
    const summary    = getById<HTMLDivElement>("resultSummary");
    const itemsBody  = getById<HTMLTableSectionElement>("resultTableBody");
    const totalsBody = getById<HTMLTableSectionElement>("categoryTotalsBody");
    const nameEl     = getById<HTMLSpanElement>("resultCalcName");

    if (!container || !summary || !itemsBody || !totalsBody) return;

    container.style.display = "block";
    if (nameEl) nameEl.textContent = result.name;

    const totalDays   = result.grand_total / 8;
    const totalMonths = totalDays / 22;

    summary.innerHTML = `
        <div class="summary-card">
            <div class="value">${format2(result.grand_total)}</div>
            <div class="label">Чел.-часы</div>
        </div>
        <div class="summary-card">
            <div class="value">${format2(totalDays)}</div>
            <div class="label">Чел.-дни</div>
        </div>
        <div class="summary-card">
            <div class="value">${format2(totalMonths)}</div>
            <div class="label">Чел.-месяцы</div>
        </div>
        <div class="summary-card">
            <div class="value">${result.items.length}</div>
            <div class="label">Позиций</div>
        </div>
    `;

    itemsBody.innerHTML = result.items.map((item, i) => `
        <tr>
            <td>${i + 1}</td>
            <td>${esc(item.norm_code)}</td>
            <td>${esc(item.norm_name)}</td>
            <td>${esc(item.category_name)}</td>
            <td>${esc(item.unit)}</td>
            <td>${format2(item.quantity)}</td>
            <td>${format2(item.base_labor_hours)}</td>
            <td>${format3(item.coefficient_mult)}</td>
            <td><strong>${format2(item.total_labor_hours)}</strong></td>
            <td>${format2(item.total_man_days)}</td>
        </tr>
    `).join("");

    const totalHours = result.category_totals.reduce((s, r) => s + r.total_labor_hours, 0);

    totalsBody.innerHTML =
        result.category_totals.map(row => `
            <tr>
                <td>${esc(row.category_name)}</td>
                <td>${format2(row.total_labor_hours)}</td>
                <td>${format3(row.global_coeff)}</td>
                <td><strong>${format2(row.adjusted_hours)}</strong></td>
                <td><strong>${format2(row.adjusted_days)}</strong></td>
            </tr>
        `).join("") +
        `<tr class="result-total-row">
            <td><strong>ИТОГО</strong></td>
            <td><strong>${format2(totalHours)}</strong></td>
            <td>—</td>
            <td><strong>${format2(result.grand_total)}</strong></td>
            <td><strong>${format2(result.grand_total / 8)}</strong></td>
        </tr>`;

    container.scrollIntoView({ behavior: "smooth", block: "start" });
}

// ─── Экспорт ─────────────────────────────────────────────────────────────────

async function exportExcel(): Promise<void> {
    if (!lastResult) { showToast("Сначала выполните расчёт", "warning"); return; }
    try {
        const blob = await exportFile(EXPORT_EXCEL_URL, lastResult);
        downloadBlob(blob, `labor_${sanitizeFileName(lastResult.name)}.xlsx`);
        showToast("Excel файл загружен");
    } catch (e) {
        console.error(e);
        showToast("Ошибка экспорта Excel", "error");
    }
}

async function exportPDF(): Promise<void> {
    if (!lastResult) { showToast("Сначала выполните расчёт", "warning"); return; }
    try {
        const blob = await exportFile(EXPORT_PDF_URL, lastResult);
        downloadBlob(blob, `labor_${sanitizeFileName(lastResult.name)}.pdf`);
        showToast("PDF файл загружен");
    } catch (e) {
        console.error(e);
        showToast("Ошибка экспорта PDF", "error");
    }
}

// ─── CRUD справочников ───────────────────────────────────────────────────────

function showCategoryModal(): void {
    setValue("categoryEditId", ""); setValue("categoryCode", "");
    setValue("categoryName", "");   setValue("categoryDescription", "");
    setText("categoryModalTitle", "Добавить категорию");
    openModal("categoryModal");
}

function editCategory(id: number): void {
    const c = categories.find(x => x.id === id);
    if (!c) return;
    setValue("categoryEditId", String(c.id));
    setValue("categoryCode", c.code);
    setValue("categoryName", c.name);
    setValue("categoryDescription", c.description);
    setText("categoryModalTitle", "Редактировать категорию");
    openModal("categoryModal");
}

async function saveCategory(): Promise<void> {
    const editId  = getValue("categoryEditId");
    const payload = {
        code:        getValue("categoryCode").trim(),
        name:        getValue("categoryName").trim(),
        description: getValue("categoryDescription").trim(),
    };
    if (!payload.code || !payload.name) {
        showToast("Заполните код и название", "warning"); return;
    }
    try {
        editId
            ? await apiPut(`/categories/${editId}`, payload)
            : await apiPost("/categories", payload);
        showToast(editId ? "Категория обновлена" : "Категория создана");
        closeModal("categoryModal");
        await loadCategories();
        buildGlobalCoeffGrid();
    } catch (e) { console.error(e); showToast("Ошибка сохранения категории", "error"); }
}

async function deleteCategory(id: number): Promise<void> {
    if (!window.confirm("Удалить категорию?")) return;
    try {
        await apiDelete(`/categories/${id}`);
        showToast("Категория удалена");
        await loadCategories();
        await loadNorms();
        buildGlobalCoeffGrid();
    } catch (e) { console.error(e); showToast("Ошибка удаления", "error"); }
}

function showNormModal(): void {
    setValue("normEditId", "");  setValue("normCode", "");
    setValue("normName", "");    setValue("normUnit", "чел.-ч");
    setValue("normLaborHours", ""); setValue("normSource", "ГЭСН");
    setValue("normDescription", "");
    if (categories.length) setValue("normCategoryId", String(categories[0].id));
    setText("normModalTitle", "Добавить норму");
    openModal("normModal");
}

function editNorm(id: number): void {
    const n = norms.find(x => x.id === id);
    if (!n) return;
    setValue("normEditId",      String(n.id));
    setValue("normCode",        n.code);
    setValue("normName",        n.name);
    setValue("normUnit",        n.unit);
    setValue("normLaborHours",  String(n.labor_hours));
    setValue("normSource",      n.source);
    setValue("normDescription", n.description);
    setValue("normCategoryId",  String(n.category_id));
    setText("normModalTitle", "Редактировать норму");
    openModal("normModal");
}

async function saveNorm(): Promise<void> {
    const editId  = getValue("normEditId");
    const payload = {
        category_id:  parseInt(getValue("normCategoryId") || "0", 10),
        code:         getValue("normCode").trim(),
        name:         getValue("normName").trim(),
        unit:         getValue("normUnit").trim() || "чел.-ч",
        labor_hours:  parseFloat(getValue("normLaborHours") || "0"),
        description:  getValue("normDescription").trim(),
        source:       getValue("normSource").trim() || "ГЭСН",
    };
    if (!payload.category_id || !payload.code || !payload.name || payload.labor_hours <= 0) {
        showToast("Заполните обязательные поля нормы", "warning"); return;
    }
    try {
        editId
            ? await apiPut(`/norms/${editId}`, payload)
            : await apiPost("/norms", payload);
        showToast(editId ? "Норма обновлена" : "Норма создана");
        closeModal("normModal");
        await loadNorms();
    } catch (e) { console.error(e); showToast("Ошибка сохранения нормы", "error"); }
}

async function deleteNorm(id: number): Promise<void> {
    if (!window.confirm("Удалить норму?")) return;
    try {
        await apiDelete(`/norms/${id}`);
        showToast("Норма удалена");
        await loadNorms();
    } catch (e) { console.error(e); showToast("Ошибка удаления", "error"); }
}

function showCoeffModal(): void {
    setValue("coeffEditId", "");  setValue("coeffCode", "");
    setValue("coeffName", "");    setValue("coeffValue", "1.0");
    setValue("coeffCategoryId", "0"); setValue("coeffDescription", "");
    setText("coeffModalTitle", "Добавить коэффициент");
    openModal("coeffModal");
}

function editCoefficient(id: number): void {
    const c = coefficients.find(x => x.id === id);
    if (!c) return;
    setValue("coeffEditId",      String(c.id));
    setValue("coeffCode",        c.code);
    setValue("coeffName",        c.name);
    setValue("coeffValue",       String(c.value));
    setValue("coeffCategoryId",  String(c.category_id || 0));
    setValue("coeffDescription", c.description);
    setText("coeffModalTitle", "Редактировать коэффициент");
    openModal("coeffModal");
}

async function saveCoefficient(): Promise<void> {
    const editId   = getValue("coeffEditId");
    const rawCat   = parseInt(getValue("coeffCategoryId") || "0", 10);
    const payload: { code: string; name: string; value: number; category_id: number | null; description: string } = {
        code:        getValue("coeffCode").trim(),
        name:        getValue("coeffName").trim(),
        value:       parseFloat(getValue("coeffValue") || "1"),
        category_id: rawCat > 0 ? rawCat : null,
        description: getValue("coeffDescription").trim(),
    };
    if (!payload.code || !payload.name) {
        showToast("Заполните код и название", "warning"); return;
    }
    try {
        editId
            ? await apiPut(`/coefficients/${editId}`, payload)
            : await apiPost("/coefficients", payload);
        showToast(editId ? "Коэффициент обновлён" : "Коэффициент создан");
        closeModal("coeffModal");
        await loadCoefficients();
    } catch (e) { console.error(e); showToast("Ошибка сохранения", "error"); }
}

async function deleteCoefficient(id: number): Promise<void> {
    if (!window.confirm("Удалить коэффициент?")) return;
    try {
        await apiDelete(`/coefficients/${id}`);
        showToast("Коэффициент удалён");
        await loadCoefficients();
    } catch (e) { console.error(e); showToast("Ошибка удаления", "error"); }
}

// ─── Обработчики событий ─────────────────────────────────────────────────────

function onCalcItemsChange(event: Event): void {
    const target = event.target as HTMLElement;
    if (!(target instanceof HTMLSelectElement)) return;
    if (!target.classList.contains("labor-norm-select")) return;
    const item = target.closest(".calc-item") as HTMLElement | null;
    if (item) renderCalcCoeffSelector(item);
}

function onCalcItemsClick(event: Event): void {
    const btn = (event.target as HTMLElement)
        .closest("[data-action='remove-calc-item']") as HTMLButtonElement | null;
    if (!btn) return;
    btn.closest(".calc-item")?.remove();
    renumberCalcItems();
}

function onCategoriesTableClick(event: Event): void {
    const btn = (event.target as HTMLElement).closest("button[data-action]") as HTMLButtonElement | null;
    if (!btn) return;
    const id = parseInt(btn.dataset.id || "0", 10);
    if (btn.dataset.action === "edit-category")   editCategory(id);
    if (btn.dataset.action === "delete-category") void deleteCategory(id);
}

function onNormsTableClick(event: Event): void {
    const btn = (event.target as HTMLElement).closest("button[data-action]") as HTMLButtonElement | null;
    if (!btn) return;
    const id = parseInt(btn.dataset.id || "0", 10);
    if (btn.dataset.action === "edit-norm")   editNorm(id);
    if (btn.dataset.action === "delete-norm") void deleteNorm(id);
}

function onCoefficientsTableClick(event: Event): void {
    const btn = (event.target as HTMLElement).closest("button[data-action]") as HTMLButtonElement | null;
    if (!btn) return;
    const id = parseInt(btn.dataset.id || "0", 10);
    if (btn.dataset.action === "edit-coeff")   editCoefficient(id);
    if (btn.dataset.action === "delete-coeff") void deleteCoefficient(id);
}

function onDocumentClick(event: Event): void {
    const target = event.target as HTMLElement;
    if (target.classList.contains("modal-overlay")) {
        target.classList.remove("active"); return;
    }
    const closeBtn = target.closest("[data-close-modal]") as HTMLElement | null;
    if (closeBtn) closeModal(closeBtn.dataset.closeModal || "");
}

// ─── Модалки ─────────────────────────────────────────────────────────────────

function openModal(id: string): void  { getById(id)?.classList.add("active"); }
function closeModal(id: string): void { getById(id)?.classList.remove("active"); }

// ─── Toast ────────────────────────────────────────────────────────────────────

function showToast(message: string, type: "success" | "error" | "warning" = "success"): void {
    const container = getById<HTMLDivElement>("toastContainer");
    if (!container) return;

    const toast = document.createElement("div");
    toast.className = `toast ${type}`;
    toast.textContent = message;
    container.appendChild(toast);

    setTimeout(() => {
        toast.style.opacity   = "0";
        toast.style.transform = "translateX(100px)";
        toast.style.transition = "0.3s ease";
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

// ─── API helpers ─────────────────────────────────────────────────────────────

async function apiGet<T>(path: string): Promise<T | null> {
    const res = await fetch(LABOR_API + path, {
        headers: buildHeaders(false), credentials: "same-origin",
    });
    return handleJsonResponse<T>(res);
}

async function apiPost<T>(path: string, data: unknown): Promise<T> {
    const res = await fetch(LABOR_API + path, {
        method: "POST", headers: buildHeaders(true),
        credentials: "same-origin", body: JSON.stringify(data),
    });
    return handleJsonResponse<T>(res) as Promise<T>;
}

async function apiPut<T>(path: string, data: unknown): Promise<T | null> {
    const res = await fetch(LABOR_API + path, {
        method: "PUT", headers: buildHeaders(true),
        credentials: "same-origin", body: JSON.stringify(data),
    });
    return handleJsonResponse<T>(res);
}

async function apiDelete<T>(path: string): Promise<T | null> {
    const res = await fetch(LABOR_API + path, {
        method: "DELETE", headers: buildHeaders(false), credentials: "same-origin",
    });
    return handleJsonResponse<T>(res);
}

async function exportFile(url: string, data: unknown): Promise<Blob> {
    const res = await fetch(url, {
        method: "POST", headers: buildHeaders(true),
        credentials: "same-origin", body: JSON.stringify(data),
    });
    if (res.status === 401) { redirectToLogin(); throw new Error("unauthorized"); }
    if (!res.ok) throw new Error("export failed");
    return res.blob();
}

function buildHeaders(json: boolean): HeadersInit {
    const h: HeadersInit = {};
    if (json) h["Content-Type"] = "application/json";
    if (token) h["Authorization"] = `Bearer ${token}`;
    return h;
}

async function handleJsonResponse<T>(res: Response): Promise<T | null> {
    if (res.status === 401) { redirectToLogin(); return null; }
    if (res.status === 204) return null;
    if (!res.ok) {
        let msg = "server error";
        try { const d = await res.json(); msg = d.error || msg; } catch { /* ignore */ }
        throw new Error(msg);
    }
    const text = await res.text();
    return text ? JSON.parse(text) as T : null;
}

// ─── Вспомогательные ─────────────────────────────────────────────────────────

function redirectToLogin(): void { window.location.href = "/login"; }

function downloadBlob(blob: Blob, fileName: string): void {
    const url = URL.createObjectURL(blob);
    const a   = Object.assign(document.createElement("a"), { href: url, download: fileName });
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
}

function getCategoryName(id: number): string {
    return categories.find(c => c.id === id)?.name || `Категория #${id}`;
}

function getById<T extends HTMLElement = HTMLElement>(id: string): T | null {
    return document.getElementById(id) as T | null;
}

function getValue(id: string): string {
    return (getById<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>(id)?.value) ?? "";
}

function setValue(id: string, val: string): void {
    const el = getById<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>(id);
    if (el) el.value = val;
}

function setText(id: string, val: string): void {
    const el = getById(id);
    if (el) el.textContent = val;
}

function format2(v: number): string { return Number(v).toFixed(2); }
function format3(v: number): string { return Number(v).toFixed(3); }

function sanitizeFileName(v: string): string {
    return v.replace(/[\\/:*?"<>|]+/g, "_");
}

function esc(v: string): string {
    const d = document.createElement("div");
    d.textContent = v || "";
    return d.innerHTML;
}
