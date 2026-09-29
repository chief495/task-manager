package seed

import (
	"log"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// SeedLaborData заполняет базу начальными нормами трудозатрат (ГЭСН).
// Вызывается при старте приложения. Если данные уже есть — пропускает.
func SeedLaborData(db *gorm.DB) error {
	var count int64
	db.Model(&models.LaborCategory{}).Count(&count)
	if count > 0 {
		log.Println("[seed] Таблицы трудозатрат уже содержат данные, пропускаем")
		return nil
	}

	log.Println("[seed] Заполнение таблиц трудозатрат начальными нормами...")

	return db.Transaction(func(tx *gorm.DB) error {
		// ─── 1. Категории ────────────────────────────────────────────────
		categories := []models.LaborCategory{
			{Name: "Проектирование", Code: "PROJ", Description: "Проектные и изыскательские работы"},
			{Name: "Разработка", Code: "DEV", Description: "Разработка программного обеспечения и документации"},
			{Name: "Монтажные работы", Code: "MOUNT", Description: "Монтаж оборудования и систем"},
			{Name: "Пусконаладочные работы", Code: "COMMISSN", Description: "Пусконаладка оборудования и систем"},
			{Name: "Электромонтажные работы", Code: "ELEC", Description: "Электромонтажные работы"},
			{Name: "Слаботочные системы", Code: "LOW_CURR", Description: "Монтаж слаботочных систем"},
			{Name: "Системы автоматизации", Code: "AUTO", Description: "Системы автоматизации и диспетчеризации"},
			{Name: "Общестроительные работы", Code: "BUILD", Description: "Общестроительные работы"},
		}

		catIDs := make(map[string]uint)
		for i := range categories {
			if err := tx.Create(&categories[i]).Error; err != nil {
				return err
			}
			catIDs[categories[i].Code] = categories[i].ID
		}

		// ─── 2. Нормы трудозатрат ────────────────────────────────────────
		norms := []models.LaborNorm{
			// Проектирование
			{CategoryID: catIDs["PROJ"], Code: "ГЭСН 81-02-01-2020-01", Name: "Разработка технического задания", Unit: "чел.-ч", LaborHours: 40, Description: "Разработка и согласование ТЗ на проект", Source: "ГЭСН"},
			{CategoryID: catIDs["PROJ"], Code: "ГЭСН 81-02-01-2020-02", Name: "Разработка проектной документации (стадия П)", Unit: "комплект", LaborHours: 160, Description: "Стадия П — проектная документация", Source: "ГЭСН"},
			{CategoryID: catIDs["PROJ"], Code: "ГЭСН 81-02-01-2020-03", Name: "Разработка рабочей документации (стадия РД)", Unit: "комплект", LaborHours: 240, Description: "Стадия РД — рабочая документация", Source: "ГЭСН"},
			{CategoryID: catIDs["PROJ"], Code: "ГЭСН 81-02-01-2020-04", Name: "Обследование объекта", Unit: "чел.-ч", LaborHours: 24, Description: "Обследование существующего объекта", Source: "ГЭСН"},
			{CategoryID: catIDs["PROJ"], Code: "ГЭСН 81-02-01-2020-05", Name: "Согласование и экспертиза проекта", Unit: "чел.-ч", LaborHours: 32, Description: "Прохождение экспертизы", Source: "ГЭСН"},
			{CategoryID: catIDs["PROJ"], Code: "PROJ-06", Name: "Разработка сметной документации", Unit: "комплект", LaborHours: 80, Description: "Составление локальных и сводных смет", Source: "НР"},

			// Разработка ПО
			{CategoryID: catIDs["DEV"], Code: "DEV-01", Name: "Разработка ТЗ на ПО", Unit: "чел.-ч", LaborHours: 40, Description: "Разработка технического задания на ПО", Source: "ГОСТ 19.201"},
			{CategoryID: catIDs["DEV"], Code: "DEV-02", Name: "Проектирование архитектуры ПО", Unit: "чел.-ч", LaborHours: 80, Description: "Проектирование архитектуры системы", Source: "ГОСТ 34"},
			{CategoryID: catIDs["DEV"], Code: "DEV-03", Name: "Кодирование модуля ПО (средняя сложность)", Unit: "модуль", LaborHours: 120, Description: "Разработка одного модуля средней сложности", Source: "НР"},
			{CategoryID: catIDs["DEV"], Code: "DEV-04", Name: "Тестирование модуля ПО", Unit: "модуль", LaborHours: 40, Description: "Тестирование одного модуля", Source: "НР"},
			{CategoryID: catIDs["DEV"], Code: "DEV-05", Name: "Разработка эксплуатационной документации", Unit: "комплект", LaborHours: 60, Description: "Руководство пользователя, администратора", Source: "ГОСТ 19"},
			{CategoryID: catIDs["DEV"], Code: "DEV-06", Name: "Интеграционное тестирование", Unit: "чел.-ч", LaborHours: 80, Description: "Тестирование взаимодействия компонентов", Source: "НР"},

			// Монтажные работы
			{CategoryID: catIDs["MOUNT"], Code: "ГЭСН 10-01-001-01", Name: "Монтаж металлоконструкций", Unit: "т", LaborHours: 28, Description: "Монтаж стальных конструкций", Source: "ГЭСН"},
			{CategoryID: catIDs["MOUNT"], Code: "ГЭСН 10-01-002-01", Name: "Монтаж технологического оборудования (до 1т)", Unit: "шт.", LaborHours: 16, Description: "Монтаж единицы оборудования до 1т", Source: "ГЭСН"},
			{CategoryID: catIDs["MOUNT"], Code: "ГЭСН 10-01-002-02", Name: "Монтаж технологического оборудования (1-5т)", Unit: "шт.", LaborHours: 32, Description: "Монтаж единицы оборудования 1-5т", Source: "ГЭСН"},
			{CategoryID: catIDs["MOUNT"], Code: "ГЭСН 10-01-003-01", Name: "Монтаж трубопроводов стальных (Ду до 50мм)", Unit: "100 м", LaborHours: 45, Description: "Монтаж стальных трубопроводов малого диаметра", Source: "ГЭСН"},
			{CategoryID: catIDs["MOUNT"], Code: "ГЭСН 10-01-003-02", Name: "Монтаж трубопроводов стальных (Ду 50-150мм)", Unit: "100 м", LaborHours: 72, Description: "Монтаж стальных трубопроводов среднего диаметра", Source: "ГЭСН"},
			{CategoryID: catIDs["MOUNT"], Code: "MOUNT-06", Name: "Монтаж вентиляционного оборудования", Unit: "шт.", LaborHours: 24, Description: "Монтаж вентустановок", Source: "ГЭСН"},

			// Пусконаладочные работы
			{CategoryID: catIDs["COMMISSN"], Code: "ГЭСН 81-05-01-2001-01", Name: "ПНР электрооборудования до 0.4кВ", Unit: "присоединение", LaborHours: 8, Description: "Пусконаладка электрооборудования до 0.4кВ", Source: "ГЭСН"},
			{CategoryID: catIDs["COMMISSN"], Code: "ГЭСН 81-05-01-2001-02", Name: "ПНР электрооборудования 6-10кВ", Unit: "присоединение", LaborHours: 16, Description: "Пусконаладка электрооборудования 6-10кВ", Source: "ГЭСН"},
			{CategoryID: catIDs["COMMISSN"], Code: "ГЭСН 81-05-02-2001-01", Name: "ПНР систем вентиляции", Unit: "система", LaborHours: 24, Description: "Пусконаладка систем вентиляции и кондиционирования", Source: "ГЭСН"},
			{CategoryID: catIDs["COMMISSN"], Code: "ГЭСН 81-05-03-2001-01", Name: "ПНР контроллеров АСУТП", Unit: "контроллер", LaborHours: 40, Description: "Наладка одного контроллера АСУТП", Source: "ГЭСН"},
			{CategoryID: catIDs["COMMISSN"], Code: "COMMISSN-05", Name: "Комплексное опробование системы", Unit: "система", LaborHours: 80, Description: "Комплексное опробование всей системы", Source: "НР"},
			{CategoryID: catIDs["COMMISSN"], Code: "COMMISSN-06", Name: "ПНР частотных преобразователей", Unit: "шт.", LaborHours: 12, Description: "Наладка частотного преобразователя", Source: "ГЭСН"},

			// Электромонтажные работы
			{CategoryID: catIDs["ELEC"], Code: "ГЭСН 08-01-001-01", Name: "Прокладка кабеля до 35кВ в лотках", Unit: "100 м", LaborHours: 12, Description: "Прокладка силового кабеля в лотках", Source: "ГЭСН"},
			{CategoryID: catIDs["ELEC"], Code: "ГЭСН 08-01-002-01", Name: "Монтаж кабельных лотков", Unit: "100 м", LaborHours: 18, Description: "Монтаж кабельных лотков и коробов", Source: "ГЭСН"},
			{CategoryID: catIDs["ELEC"], Code: "ГЭСН 08-02-001-01", Name: "Монтаж автоматического выключателя", Unit: "шт.", LaborHours: 2.5, Description: "Установка автоматического выключателя в щит", Source: "ГЭСН"},
			{CategoryID: catIDs["ELEC"], Code: "ГЭСН 08-02-002-01", Name: "Монтаж электрощитового оборудования", Unit: "шт.", LaborHours: 16, Description: "Монтаж и подключение электрощита", Source: "ГЭСН"},
			{CategoryID: catIDs["ELEC"], Code: "ELEC-05", Name: "Монтаж системы заземления", Unit: "контур", LaborHours: 24, Description: "Устройство контура заземления", Source: "ГЭСН"},
			{CategoryID: catIDs["ELEC"], Code: "ELEC-06", Name: "Монтаж системы молниезащиты", Unit: "комплект", LaborHours: 32, Description: "Монтаж молниеприёмников и токоотводов", Source: "ГЭСН"},

			// Слаботочные системы
			{CategoryID: catIDs["LOW_CURR"], Code: "LOW-01", Name: "Прокладка кабеля UTP/FTP", Unit: "100 м", LaborHours: 8, Description: "Прокладка структурированных кабельных систем", Source: "НР"},
			{CategoryID: catIDs["LOW_CURR"], Code: "LOW-02", Name: "Монтаж видеокамеры", Unit: "шт.", LaborHours: 4, Description: "Установка и подключение IP-видеокамеры", Source: "НР"},
			{CategoryID: catIDs["LOW_CURR"], Code: "LOW-03", Name: "Монтаж точки доступа Wi-Fi", Unit: "шт.", LaborHours: 3, Description: "Установка и настройка точки доступа", Source: "НР"},
			{CategoryID: catIDs["LOW_CURR"], Code: "LOW-04", Name: "Монтаж извещателя пожарной сигнализации", Unit: "шт.", LaborHours: 1.5, Description: "Установка и подключение извещателя", Source: "ГЭСН"},
			{CategoryID: catIDs["LOW_CURR"], Code: "LOW-05", Name: "Монтаж панели пожарной сигнализации", Unit: "шт.", LaborHours: 12, Description: "Монтаж ППКП", Source: "ГЭСН"},
			{CategoryID: catIDs["LOW_CURR"], Code: "LOW-06", Name: "Монтаж СКС (рабочее место)", Unit: "точка", LaborHours: 6, Description: "Полный монтаж рабочего места СКС", Source: "НР"},

			// Системы автоматизации
			{CategoryID: catIDs["AUTO"], Code: "AUTO-01", Name: "Монтаж датчика температуры", Unit: "шт.", LaborHours: 3, Description: "Установка и подключение датчика температуры", Source: "ГЭСН"},
			{CategoryID: catIDs["AUTO"], Code: "AUTO-02", Name: "Монтаж датчика давления", Unit: "шт.", LaborHours: 4, Description: "Установка и подключение датчика давления", Source: "ГЭСН"},
			{CategoryID: catIDs["AUTO"], Code: "AUTO-03", Name: "Монтаж исполнительного механизма", Unit: "шт.", LaborHours: 8, Description: "Монтаж клапана с электроприводом", Source: "ГЭСН"},
			{CategoryID: catIDs["AUTO"], Code: "AUTO-04", Name: "Монтаж шкафа автоматики", Unit: "шт.", LaborHours: 24, Description: "Монтаж и подключение шкафа автоматики", Source: "ГЭСН"},
			{CategoryID: catIDs["AUTO"], Code: "AUTO-05", Name: "Программирование контроллера", Unit: "контроллер", LaborHours: 80, Description: "Разработка и загрузка программы контроллера", Source: "НР"},
			{CategoryID: catIDs["AUTO"], Code: "AUTO-06", Name: "Настройка SCADA-системы", Unit: "точка ввода/вывода", LaborHours: 0.5, Description: "Настройка одного тега SCADA", Source: "НР"},

			// Общестроительные работы
			{CategoryID: catIDs["BUILD"], Code: "ГЭСН 06-01-001-01", Name: "Устройство фундамента под оборудование", Unit: "м³", LaborHours: 8, Description: "Бетонный фундамент", Source: "ГЭСН"},
			{CategoryID: catIDs["BUILD"], Code: "ГЭСН 06-01-002-01", Name: "Устройство кабельного канала", Unit: "м", LaborHours: 3.5, Description: "Кабельный канал в полу", Source: "ГЭСН"},
			{CategoryID: catIDs["BUILD"], Code: "BUILD-03", Name: "Устройство фальшпола", Unit: "м²", LaborHours: 1.2, Description: "Монтаж фальшпола в серверной", Source: "НР"},
			{CategoryID: catIDs["BUILD"], Code: "BUILD-04", Name: "Устройство проёмов в стенах и перекрытиях", Unit: "шт.", LaborHours: 4, Description: "Пробивка и заделка проёмов", Source: "ГЭСН"},
		}

		for i := range norms {
			if err := tx.Create(&norms[i]).Error; err != nil {
				return err
			}
		}

		// ─── 3. Коэффициенты ─────────────────────────────────────────────
		uintPtr := func(v uint) *uint { return &v }

		coefficients := []models.LaborCoefficient{
			// Общие (без категории)
			{Name: "Стеснённые условия работы", Code: "K-TIGHT", Value: 1.15, Description: "Работа в стеснённых условиях"},
			{Name: "Работа на высоте", Code: "K-HEIGHT", Value: 1.2, Description: "Работа на высоте более 3.5 м"},
			{Name: "Работа в зимнее время", Code: "K-WINTER", Value: 1.1, Description: "Зимний период проведения работ"},
			{Name: "Действующее предприятие", Code: "K-ACTIVE", Value: 1.25, Description: "Работа на действующем предприятии"},
			{Name: "Вредные условия", Code: "K-HAZARD", Value: 1.3, Description: "Работа во вредных условиях"},
			{Name: "Срочность работ", Code: "K-URGENT", Value: 1.5, Description: "Сжатые сроки выполнения"},
			{Name: "Простой объект", Code: "K-SIMPLE", Value: 0.7, Description: "Пониженная сложность объекта"},
			{Name: "Объект средней сложности", Code: "K-MEDIUM", Value: 1.0, Description: "Средняя сложность объекта"},
			{Name: "Сложный объект", Code: "K-COMPLEX", Value: 1.3, Description: "Повышенная сложность объекта"},
			{Name: "Особо сложный объект", Code: "K-VCOMPLEX", Value: 1.5, Description: "Особо сложный объект"},
			{Name: "Удалённый объект", Code: "K-REMOTE", Value: 1.15, Description: "Удалённый от базы объект"},

			// Привязанные к категориям
			{Name: "Новое строительство (проектирование)", Code: "K-PROJ-NEW", Value: 1.0, CategoryID: uintPtr(catIDs["PROJ"]), Description: "Проектирование нового объекта"},
			{Name: "Реконструкция (проектирование)", Code: "K-PROJ-RECON", Value: 1.3, CategoryID: uintPtr(catIDs["PROJ"]), Description: "Проектирование реконструкции"},
			{Name: "Капремонт (проектирование)", Code: "K-PROJ-REPAIR", Value: 1.15, CategoryID: uintPtr(catIDs["PROJ"]), Description: "Проектирование капитального ремонта"},

			{Name: "Стандартный стек технологий", Code: "K-DEV-STD", Value: 1.0, CategoryID: uintPtr(catIDs["DEV"]), Description: "Стандартные технологии разработки"},
			{Name: "Новые технологии", Code: "K-DEV-NEW", Value: 1.4, CategoryID: uintPtr(catIDs["DEV"]), Description: "Использование новых технологий"},
			{Name: "Высоконагруженная система", Code: "K-DEV-HL", Value: 1.5, CategoryID: uintPtr(catIDs["DEV"]), Description: "Требования к высокой нагрузке"},

			{Name: "Типовое оборудование (наладка)", Code: "K-COM-STD", Value: 1.0, CategoryID: uintPtr(catIDs["COMMISSN"]), Description: "Типовое серийное оборудование"},
			{Name: "Нестандартное оборудование (наладка)", Code: "K-COM-CUSTOM", Value: 1.4, CategoryID: uintPtr(catIDs["COMMISSN"]), Description: "Нестандартное оборудование"},
			{Name: "Импортное оборудование (наладка)", Code: "K-COM-IMPORT", Value: 1.2, CategoryID: uintPtr(catIDs["COMMISSN"]), Description: "Импортное оборудование"},
		}

		for i := range coefficients {
			if err := tx.Create(&coefficients[i]).Error; err != nil {
				return err
			}
		}

		log.Printf("[seed] Загружено %d категорий, %d норм, %d коэффициентов",
			len(categories), len(norms), len(coefficients))
		return nil
	})
}
