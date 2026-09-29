export {};

const allowedPaths = ['/manager/contacts', '/admin/contacts'];
if (!allowedPaths.includes(window.location.pathname)) {
    throw new Error('manager-contacts.ts loaded outside contacts page');
}

interface Organization {
    id: number;
    name: string;
    description?: string;
    phone?: string;
    email?: string;
    address?: string;
    website?: string;
    notes?: string;
}

interface OrganizationContact {
    id: number;
    organization_id: number;
    full_name: string;
    position?: string;
    phone?: string;
    email?: string;
    notes?: string;
}

declare const ymaps: any;

const content = document.getElementById('content');
if (!content) {
    throw new Error('Не найден элемент #content');
}

const YANDEX_MAPS_API_KEY = '15277c56-2dcb-4e64-aefe-2394e6f89c1d';
let token = '';
let organizations: Organization[] = [];
let contacts: OrganizationContact[] = [];
let selectedOrganizationID: number | null = null;
let editingContactID: number | null = null;
let addressAutocompleteInitialized = false;
const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/i;
const PHONE_ALLOWED_CHARS_PATTERN = /^\+?[0-9\s()\-]+$/;
const PHONE_INPUT_PATTERN = "\\+?[0-9\\s()\\-]{7,25}";
const PHONE_VALIDATION_MESSAGE = 'Введите корректный номер телефона: 7-15 цифр, можно использовать +, пробелы, скобки и дефисы.';

function getAuthToken(): string {
    return localStorage.getItem('token') || '';
}

function logout(): void {
    localStorage.removeItem('token');
    document.cookie = 'token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
    window.location.href = '/login';
}

function escapeHtml(value: string | undefined | null): string {
    const div = document.createElement('div');
    div.textContent = value || '';
    return div.innerHTML;
}

function isValidEmail(value: string): boolean {
    return value === '' || EMAIL_PATTERN.test(value);
}

function isValidPhone(value: string): boolean {
    if (value === '') return true;
    const digits = value.replace(/\D/g, '');
    return PHONE_ALLOWED_CHARS_PATTERN.test(value) && digits.length >= 7 && digits.length <= 15;
}

function validatePhoneAndEmail(form: HTMLFormElement): boolean {
    const phoneInput = form.elements.namedItem('phone') as HTMLInputElement | null;
    const emailInput = form.elements.namedItem('email') as HTMLInputElement | null;
    const phone = phoneInput?.value.trim() || '';
    const email = emailInput?.value.trim() || '';

    phoneInput?.setCustomValidity('');
    emailInput?.setCustomValidity('');

    if (!isValidPhone(phone)) {
        phoneInput?.setCustomValidity(PHONE_VALIDATION_MESSAGE);
    }
    if (!isValidEmail(email)) {
        emailInput?.setCustomValidity('Введите корректный email, например name@example.com.');
    }

    return form.reportValidity();
}

function setupValidationReset(form: HTMLFormElement): void {
    const phoneInput = form.elements.namedItem('phone') as HTMLInputElement | null;
    const emailInput = form.elements.namedItem('email') as HTMLInputElement | null;

    phoneInput?.addEventListener('input', () => phoneInput.setCustomValidity(''));
    phoneInput?.addEventListener('blur', () => {
        const phone = phoneInput.value.trim();
        phoneInput.setCustomValidity(isValidPhone(phone) ? '' : PHONE_VALIDATION_MESSAGE);
        if (phone) phoneInput.reportValidity();
    });

    emailInput?.addEventListener('input', () => emailInput.setCustomValidity(''));
}

async function apiFetch(path: string, options: RequestInit = {}): Promise<Response> {
    const headers = new Headers(options.headers || {});
    headers.set('Authorization', `Bearer ${token}`);
    if (options.body && !headers.has('Content-Type')) {
        headers.set('Content-Type', 'application/json');
    }

    const response = await fetch(`/api/v1${path}`, { ...options, headers });
    if (response.status === 401) {
        logout();
    }
    return response;
}

function renderPage(): void {
    content!.innerHTML = `
        <link rel="stylesheet" href="/static/css/manager-contacts.css">
        <div class="manager-contacts-page">
            <div class="page-heading">
                <div>
                    <h1>Компании и контакты</h1>
                    <p>Карточки компаний и контактные лица для менеджеров</p>
                </div>
                <button class="btn btn-primary" type="button" data-new-organization>
                    <span class="material-icons" aria-hidden="true">add_business</span>
                    Компания
                </button>
            </div>

            <div class="contacts-layout">
                <section class="company-panel" aria-label="Список компаний">
                    <div class="search-row">
                        <span class="material-icons" aria-hidden="true">search</span>
                        <input id="companySearch" type="search" placeholder="Поиск по компании, телефону, email" />
                    </div>
                    <div id="companyList" class="company-list"></div>
                </section>

                <section class="details-panel" aria-label="Карточка компании">
                    <div id="emptyState" class="empty-state">
                        <span class="material-icons" aria-hidden="true">business</span>
                        <h2>Выберите компанию</h2>
                        <p>Создайте или откройте карточку, чтобы вести контакты.</p>
                    </div>

                    <div id="detailsState" class="details-state hidden">
                        <form id="organizationForm" class="entity-form">
                            <input type="hidden" name="id" />
                            <div class="section-title-row">
                                <h2>Карточка компании</h2>
                                <div class="btn-group">
                                    <button class="btn btn-neutral btn-sm" type="button" data-reset-organization>Очистить</button>
                                    <button class="btn btn-danger btn-sm hidden" type="button" data-delete-organization>Удалить</button>
                                    <button class="btn btn-primary btn-sm" type="submit">Сохранить</button>
                                </div>
                            </div>

                            <div class="form-grid">
                                <label>
                                    <span>Название</span>
                                    <input name="name" required maxlength="200" />
                                </label>
                                <label>
                                    <span>Телефон</span>
                                    <input name="phone" type="tel" inputmode="tel" maxlength="50" pattern="${PHONE_INPUT_PATTERN}" title="${PHONE_VALIDATION_MESSAGE}" />
                                </label>
                                <label>
                                    <span>Email</span>
                                    <input name="email" type="email" maxlength="100" />
                                </label>
                                <label>
                                    <span>Сайт</span>
                                    <input name="website" maxlength="200" />
                                </label>
                                <label class="wide">
                                    <span>Адрес</span>
                                    <input id="organizationAddress" name="address" maxlength="500" autocomplete="off" />
                                </label>
                                <label class="wide">
                                    <span>Описание</span>
                                    <textarea name="description"></textarea>
                                </label>
                                <label class="wide">
                                    <span>Заметки</span>
                                    <textarea name="notes"></textarea>
                                </label>
                            </div>
                        </form>

                        <div class="contacts-section">
                            <div class="section-title-row">
                                <h2>Контакты</h2>
                                <button class="btn btn-primary btn-sm" type="button" data-new-contact>
                                    <span class="material-icons" aria-hidden="true">person_add</span>
                                    Контакт
                                </button>
                            </div>
                            <div id="contactList" class="contact-list"></div>

                            <form id="contactForm" class="entity-form contact-form hidden">
                                <input type="hidden" name="id" />
                                <h3 id="contactFormTitle">Новый контакт</h3>
                                <div class="form-grid">
                                    <label>
                                        <span>ФИО</span>
                                        <input name="full_name" required maxlength="200" />
                                    </label>
                                    <label>
                                        <span>Должность</span>
                                        <input name="position" maxlength="200" />
                                    </label>
                                    <label>
                                        <span>Телефон</span>
                                        <input name="phone" type="tel" inputmode="tel" maxlength="50" pattern="${PHONE_INPUT_PATTERN}" title="${PHONE_VALIDATION_MESSAGE}" />
                                    </label>
                                    <label>
                                        <span>Email</span>
                                        <input name="email" type="email" maxlength="100" />
                                    </label>
                                    <label class="wide">
                                        <span>Заметки</span>
                                        <textarea name="notes"></textarea>
                                    </label>
                                </div>
                                <div class="form-actions">
                                    <button class="btn btn-neutral" type="button" data-cancel-contact>Отмена</button>
                                    <button class="btn btn-primary" type="submit">Сохранить контакт</button>
                                </div>
                            </form>
                        </div>
                    </div>
                </section>
            </div>
        </div>
    `;

    setupEventListeners();
}

function loadYandexMapsAPI(): void {
    if ((window as any).ymaps) {
        initYandexMaps();
        return;
    }

    const existingScript = document.querySelector('script[data-yandex-maps-api]');
    if (existingScript) return;

    const script = document.createElement('script');
    script.src = `https://api-maps.yandex.ru/2.1/?apikey=${YANDEX_MAPS_API_KEY}&lang=ru_RU`;
    script.async = true;
    script.dataset.yandexMapsApi = 'true';
    script.onload = initYandexMaps;
    script.onerror = () => {
        console.error('Не удалось загрузить Yandex Maps API');
        alert('Карта временно недоступна. Продолжайте вводить адрес вручную.');
    };
    document.head.appendChild(script);
}

function initYandexMaps(): void {
    ymaps.ready(() => {
        initAddressAutocomplete();
    });
}

function initAddressAutocomplete(): void {
    if (addressAutocompleteInitialized) return;

    const addressInput = document.getElementById('organizationAddress') as HTMLInputElement | null;
    if (!addressInput) return;

    addressAutocompleteInitialized = true;

    const suggestionsContainer = document.createElement('div');
    suggestionsContainer.id = 'organization-address-suggestions';
    suggestionsContainer.className = 'address-suggestions';
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

    let timeoutID: number | null = null;

    addressInput.addEventListener('input', () => {
        updateSuggestionsPosition();
        addressInput.removeAttribute('data-latitude');
        addressInput.removeAttribute('data-longitude');

        const value = addressInput.value.trim();
        if (value.length < 3) {
            suggestionsContainer.style.display = 'none';
            return;
        }

        if (timeoutID) {
            window.clearTimeout(timeoutID);
        }

        timeoutID = window.setTimeout(() => {
            fetchAddressSuggestions(value, suggestionsContainer).catch(error => {
                console.error('Ошибка при получении подсказок:', error);
                suggestionsContainer.innerHTML = '<div class="suggestion-item muted">Ошибка сети</div>';
                suggestionsContainer.style.display = 'block';
            });
        }, 300);
    });

    addressInput.addEventListener('focus', updateSuggestionsPosition);
    addressInput.addEventListener('blur', () => {
        window.setTimeout(() => {
            suggestionsContainer.style.display = 'none';
        }, 200);
    });

    suggestionsContainer.addEventListener('click', (event) => {
        const target = (event.target as HTMLElement).closest<HTMLElement>('[data-address]');
        if (!target) return;

        const address = target.dataset.address || '';
        addressInput.value = address;
        suggestionsContainer.style.display = 'none';
        fetchCoordinatesFromYandex(address, addressInput).catch(error => {
            console.error('Ошибка при получении координат:', error);
        });
    });
}

async function fetchAddressSuggestions(input: string, container: HTMLElement): Promise<void> {
    const params = new URLSearchParams({
        format: 'json',
        results: '5',
        lang: 'ru_RU',
        geocode: input,
        boundedBy: '30.0,50.0~180.0,80.0',
        apikey: YANDEX_MAPS_API_KEY,
    });

    const response = await fetch(`https://geocode-maps.yandex.ru/1.x/?${params.toString()}`, {
        method: 'GET',
        headers: { Accept: 'application/json' },
    });

    if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    const data = await response.json();
    const suggestions = data?.response?.GeoObjectCollection?.featureMember || [];

    if (suggestions.length === 0) {
        container.innerHTML = '<div class="suggestion-item muted">Ничего не найдено</div>';
        container.style.display = 'block';
        return;
    }

    container.innerHTML = suggestions.map((item: any) => {
        const address = item?.GeoObject?.metaDataProperty?.GeocoderMetaData?.Address?.formatted || '';
        return `<button class="suggestion-item" type="button" data-address="${escapeHtml(address)}">${escapeHtml(address)}</button>`;
    }).join('');
    container.style.display = 'block';
}

async function fetchCoordinatesFromYandex(address: string, input: HTMLInputElement): Promise<void> {
    try {
        const response = await ymaps.geocode(address, { results: 1 });
        const firstGeoObject = response.geoObjects.get(0);

        if (!firstGeoObject) {
            input.removeAttribute('data-latitude');
            input.removeAttribute('data-longitude');
            return;
        }

        const coords = firstGeoObject.geometry.getCoordinates();
        input.setAttribute('data-latitude', coords[1].toString());
        input.setAttribute('data-longitude', coords[0].toString());
    } catch (error) {
        input.removeAttribute('data-latitude');
        input.removeAttribute('data-longitude');
        throw error;
    }
}

async function loadData(): Promise<void> {
    const [orgRes, contactRes] = await Promise.all([
        apiFetch('/organizations'),
        apiFetch('/organization-contacts'),
    ]);

    if (!orgRes.ok) throw new Error('Не удалось загрузить компании');
    if (!contactRes.ok) throw new Error('Не удалось загрузить контакты');

    organizations = await orgRes.json();
    contacts = await contactRes.json();
    organizations.sort((a, b) => a.name.localeCompare(b.name, 'ru'));

    if (selectedOrganizationID && !organizations.some(org => org.id === selectedOrganizationID)) {
        selectedOrganizationID = null;
    }
}

function filteredOrganizations(): Organization[] {
    const query = (document.getElementById('companySearch') as HTMLInputElement | null)?.value.trim().toLowerCase() || '';
    if (!query) return organizations;

    return organizations.filter(org => {
        const haystack = [org.name, org.phone, org.email, org.address, org.website].join(' ').toLowerCase();
        return haystack.includes(query);
    });
}

function renderCompanyList(): void {
    const list = document.getElementById('companyList');
    if (!list) return;

    const items = filteredOrganizations();
    if (items.length === 0) {
        list.innerHTML = '<div class="list-empty">Компании не найдены</div>';
        return;
    }

    list.innerHTML = items.map(org => {
        const count = contacts.filter(contact => contact.organization_id === org.id).length;
        const selected = org.id === selectedOrganizationID ? ' selected' : '';
        return `
            <button class="company-item${selected}" type="button" data-organization-id="${org.id}">
                <span class="company-name">${escapeHtml(org.name)}</span>
                <span class="company-meta">${count} конт.</span>
            </button>
        `;
    }).join('');
}

function fillOrganizationForm(org: Organization | null): void {
    const form = document.getElementById('organizationForm') as HTMLFormElement | null;
    if (!form) return;

    (form.elements.namedItem('id') as HTMLInputElement).value = org?.id?.toString() || '';
    (form.elements.namedItem('name') as HTMLInputElement).value = org?.name || '';
    (form.elements.namedItem('phone') as HTMLInputElement).value = org?.phone || '';
    (form.elements.namedItem('email') as HTMLInputElement).value = org?.email || '';
    (form.elements.namedItem('website') as HTMLInputElement).value = org?.website || '';
    const addressInput = form.elements.namedItem('address') as HTMLInputElement;
    addressInput.value = org?.address || '';
    addressInput.removeAttribute('data-latitude');
    addressInput.removeAttribute('data-longitude');
    (form.elements.namedItem('description') as HTMLTextAreaElement).value = org?.description || '';
    (form.elements.namedItem('notes') as HTMLTextAreaElement).value = org?.notes || '';

    const deleteButton = document.querySelector('[data-delete-organization]');
    deleteButton?.classList.toggle('hidden', !org);
}

function renderDetails(): void {
    const empty = document.getElementById('emptyState');
    const details = document.getElementById('detailsState');
    const org = organizations.find(item => item.id === selectedOrganizationID) || null;

    empty?.classList.toggle('hidden', Boolean(org));
    details?.classList.toggle('hidden', !org);

    if (org) {
        fillOrganizationForm(org);
        renderContacts();
    }
}

function renderContacts(): void {
    const list = document.getElementById('contactList');
    if (!list || !selectedOrganizationID) return;

    const items = contacts
        .filter(contact => contact.organization_id === selectedOrganizationID)
        .sort((a, b) => a.full_name.localeCompare(b.full_name, 'ru'));

    if (items.length === 0) {
        list.innerHTML = '<div class="list-empty">Контакты ещё не добавлены</div>';
        return;
    }

    list.innerHTML = items.map(contact => `
        <article class="contact-row">
            <div>
                <strong>${escapeHtml(contact.full_name)}</strong>
                <span>${escapeHtml(contact.position || 'Без должности')}</span>
            </div>
            <div class="contact-links">
                ${contact.phone ? `<a href="tel:${escapeHtml(contact.phone)}">${escapeHtml(contact.phone)}</a>` : ''}
                ${contact.email ? `<a href="mailto:${escapeHtml(contact.email)}">${escapeHtml(contact.email)}</a>` : ''}
            </div>
            <div class="row-actions">
                <button class="btn btn-neutral btn-sm" type="button" data-edit-contact="${contact.id}">Редактировать</button>
                <button class="btn btn-danger btn-sm" type="button" data-delete-contact="${contact.id}">Удалить</button>
            </div>
        </article>
    `).join('');
}

function selectOrganization(id: number): void {
    selectedOrganizationID = id;
    editingContactID = null;
    hideContactForm();
    renderCompanyList();
    renderDetails();
}

function showNewOrganization(): void {
    selectedOrganizationID = null;
    editingContactID = null;
    document.getElementById('emptyState')?.classList.add('hidden');
    document.getElementById('detailsState')?.classList.remove('hidden');
    fillOrganizationForm(null);
    const contactList = document.getElementById('contactList');
    if (contactList) contactList.innerHTML = '<div class="list-empty">Сначала сохраните компанию</div>';
    hideContactForm();
    renderCompanyList();
    (document.querySelector('[name="name"]') as HTMLInputElement | null)?.focus();
}

function resetContactForm(contact: OrganizationContact | null = null): void {
    const form = document.getElementById('contactForm') as HTMLFormElement | null;
    if (!form) return;
    editingContactID = contact?.id || null;
    (form.elements.namedItem('id') as HTMLInputElement).value = contact?.id?.toString() || '';
    (form.elements.namedItem('full_name') as HTMLInputElement).value = contact?.full_name || '';
    (form.elements.namedItem('position') as HTMLInputElement).value = contact?.position || '';
    (form.elements.namedItem('phone') as HTMLInputElement).value = contact?.phone || '';
    (form.elements.namedItem('email') as HTMLInputElement).value = contact?.email || '';
    (form.elements.namedItem('notes') as HTMLTextAreaElement).value = contact?.notes || '';
    const title = document.getElementById('contactFormTitle');
    if (title) title.textContent = contact ? 'Редактировать контакт' : 'Новый контакт';
    form.classList.remove('hidden');
    (form.elements.namedItem('full_name') as HTMLInputElement).focus();
}

function hideContactForm(): void {
    document.getElementById('contactForm')?.classList.add('hidden');
    editingContactID = null;
}

async function saveOrganization(event: SubmitEvent): Promise<void> {
    event.preventDefault();
    const form = event.currentTarget as HTMLFormElement;
    if (!validatePhoneAndEmail(form)) return;

    const formData = new FormData(form);
    const id = Number(formData.get('id')) || 0;
    const payload = {
        id,
        name: String(formData.get('name') || '').trim(),
        phone: String(formData.get('phone') || '').trim(),
        email: String(formData.get('email') || '').trim(),
        website: String(formData.get('website') || '').trim(),
        address: String(formData.get('address') || '').trim(),
        description: String(formData.get('description') || '').trim(),
        notes: String(formData.get('notes') || '').trim(),
    };

    if (!payload.name) {
        alert('Укажите название компании');
        return;
    }

    const response = await apiFetch(id ? '/organizations' : '/organizations', {
        method: id ? 'PUT' : 'POST',
        body: JSON.stringify(payload),
    });

    if (!response.ok) {
        const err = await response.json().catch(() => ({ error: 'Не удалось сохранить компанию' }));
        alert(err.error || 'Не удалось сохранить компанию');
        return;
    }

    const saved = await response.json();
    selectedOrganizationID = saved.id;
    await loadData();
    renderCompanyList();
    renderDetails();
}

async function saveContact(event: SubmitEvent): Promise<void> {
    event.preventDefault();
    if (!selectedOrganizationID) {
        alert('Сначала сохраните компанию');
        return;
    }

    const form = event.currentTarget as HTMLFormElement;
    if (!validatePhoneAndEmail(form)) return;

    const formData = new FormData(form);
    const payload = {
        organization_id: selectedOrganizationID,
        full_name: String(formData.get('full_name') || '').trim(),
        position: String(formData.get('position') || '').trim(),
        phone: String(formData.get('phone') || '').trim(),
        email: String(formData.get('email') || '').trim(),
        notes: String(formData.get('notes') || '').trim(),
    };

    if (!payload.full_name) {
        alert('Укажите ФИО контакта');
        return;
    }

    const response = await apiFetch(editingContactID ? `/organization-contacts/${editingContactID}` : '/organization-contacts', {
        method: editingContactID ? 'PUT' : 'POST',
        body: JSON.stringify(payload),
    });

    if (!response.ok) {
        const err = await response.json().catch(() => ({ error: 'Не удалось сохранить контакт' }));
        alert(err.error || 'Не удалось сохранить контакт');
        return;
    }

    hideContactForm();
    await loadData();
    renderCompanyList();
    renderContacts();
}

async function deleteContact(id: number): Promise<void> {
    if (!confirm('Удалить контакт?')) return;
    const response = await apiFetch(`/organization-contacts/${id}`, { method: 'DELETE' });
    if (!response.ok) {
        const err = await response.json().catch(() => ({ error: 'Не удалось удалить контакт' }));
        alert(err.error || 'Не удалось удалить контакт');
        return;
    }
    contacts = contacts.filter(contact => contact.id !== id);
    renderCompanyList();
    renderContacts();
}

async function deleteOrganization(): Promise<void> {
    if (!selectedOrganizationID) return;

    const org = organizations.find(item => item.id === selectedOrganizationID);
    const name = org?.name || 'компанию';
    if (!confirm(`Удалить компанию "${name}"? Все её контакты тоже будут удалены.`)) return;

    const response = await apiFetch(`/organizations/${selectedOrganizationID}`, { method: 'DELETE' });
    if (!response.ok) {
        const err = await response.json().catch(() => ({ error: 'Не удалось удалить компанию' }));
        alert(err.error || 'Не удалось удалить компанию');
        return;
    }

    selectedOrganizationID = null;
    editingContactID = null;
    hideContactForm();
    await loadData();
    renderCompanyList();
    renderDetails();
}

function setupEventListeners(): void {
    const organizationForm = document.getElementById('organizationForm') as HTMLFormElement | null;
    const contactForm = document.getElementById('contactForm') as HTMLFormElement | null;

    document.querySelector('[data-new-organization]')?.addEventListener('click', showNewOrganization);
    document.getElementById('companySearch')?.addEventListener('input', renderCompanyList);
    if (organizationForm) setupValidationReset(organizationForm);
    organizationForm?.addEventListener('submit', (event) => {
        saveOrganization(event as SubmitEvent).catch(error => {
            console.error(error);
            alert('Не удалось сохранить компанию');
        });
    });
    document.querySelector('[data-reset-organization]')?.addEventListener('click', () => {
        if (confirm('Очистить карточку компании? Несохранённые изменения будут потеряны.')) {
            showNewOrganization();
        }
    });
    document.querySelector('[data-delete-organization]')?.addEventListener('click', () => {
        deleteOrganization().catch(error => {
            console.error(error);
            alert('Не удалось удалить компанию');
        });
    });
    document.querySelector('[data-new-contact]')?.addEventListener('click', () => resetContactForm());
    document.querySelector('[data-cancel-contact]')?.addEventListener('click', hideContactForm);
    if (contactForm) setupValidationReset(contactForm);
    contactForm?.addEventListener('submit', (event) => {
        saveContact(event as SubmitEvent).catch(error => {
            console.error(error);
            alert('Не удалось сохранить контакт');
        });
    });

    document.getElementById('companyList')?.addEventListener('click', (event) => {
        const button = (event.target as HTMLElement).closest('[data-organization-id]') as HTMLElement | null;
        if (!button) return;
        selectOrganization(Number(button.dataset.organizationId));
    });

    document.getElementById('contactList')?.addEventListener('click', (event) => {
        const target = event.target as HTMLElement;
        const editButton = target.closest('[data-edit-contact]') as HTMLElement | null;
        const deleteButton = target.closest('[data-delete-contact]') as HTMLElement | null;

        if (editButton) {
            const contact = contacts.find(item => item.id === Number(editButton.dataset.editContact));
            if (contact) resetContactForm(contact);
            return;
        }

        if (deleteButton) {
            deleteContact(Number(deleteButton.dataset.deleteContact)).catch(error => {
                console.error(error);
                alert('Не удалось удалить контакт');
            });
        }
    });
}

document.addEventListener('DOMContentLoaded', async () => {
    token = getAuthToken();
    if (!token) {
        window.location.href = '/login';
        return;
    }

    renderPage();
    loadYandexMapsAPI();
    try {
        await loadData();
        renderCompanyList();
        if (organizations.length > 0) {
            selectOrganization(organizations[0].id);
        } else {
            renderDetails();
        }
    } catch (error) {
        console.error(error);
        content!.innerHTML = '<div class="loading">Не удалось загрузить справочник компаний</div>';
    }
});
