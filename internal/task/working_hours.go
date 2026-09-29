package task

import (
	"time"
)

// Рабочее время: с 10:00 до 18:00 → 8 часов в день
const (
	WorkStartHour = 10
	WorkEndHour   = 18
	HoursPerDay   = 8 // WorkEndHour - WorkStartHour
)

// Список праздников (пример для России, формат: "02.01" → 2 января)
// Можно вынести в конфиг или БД
var holidays = map[string]bool{
	"01.01": true, // Новый год
	"02.01": true, // Новогодние каникулы
	"03.01": true,
	"04.01": true,
	"05.01": true,
	"06.01": true,
	"07.01": true,
	"08.01": true,
	"09.01": true,
	"23.02": true, // День защитника Отечества
	"08.03": true, // Международный женский день
	"01.05": true, // Праздник весны и труда
	"09.05": true, // День Победы
	"12.06": true, // День России
	"04.11": true, // День народного единства
}

// isHoliday проверяет, является ли дата выходным или праздником
func isHoliday(date time.Time) bool {
	// Проверка выходных
	if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		return true
	}

	// Форматируем дату как "день.месяц"
	dayMonth := date.Format("02.01")
	return holidays[dayMonth]
}

// calculateWorkingHours рассчитывает количество рабочих часов между двумя датами
// Учитывает:
// - Рабочие часы: с 10:00 до 18:00 (8 часов в день)
// - Только будние дни
// - Праздники
// - Частичные дни в начале и конце интервала
func calculateWorkingHours(start, end time.Time) float32 {
	if start.After(end) {
		return 0
	}

	loc := time.Local
	start = start.In(loc)
	end = end.In(loc)

	var total float32 = 0

	// Начало и конец первого дня
	year, month, day := start.Date()
	dayStart := time.Date(year, month, day, WorkStartHour, 0, 0, 0, loc)
	dayEnd := time.Date(year, month, day, WorkEndHour, 0, 0, 0, loc)

	// Если начало интервала до конца рабочего дня и первый день — не выходной/праздник
	if start.Before(dayEnd) && !isHoliday(start) {
		// Рабочее время в первый день = min(end, dayEnd) - max(start, dayStart)
		workStart := start
		if start.Before(dayStart) {
			workStart = dayStart
		}
		workEnd := end
		if end.After(dayEnd) {
			workEnd = dayEnd
		} else {
			workEnd = end
		}
		if workStart.Before(workEnd) {
			total += float32(workEnd.Sub(workStart).Hours())
		}
	}

	// Переходим ко второму дню
	nextDay := dayStart.AddDate(0, 0, 1)
	year, month, day = nextDay.Date()
	nextWorkDayStart := time.Date(year, month, day, WorkStartHour, 0, 0, 0, loc)
	nextWorkDayEnd := time.Date(year, month, day, WorkEndHour, 0, 0, 0, loc)

	// Полные дни между первым и последним
	for nextWorkDayStart.Before(end) {
		if !isHoliday(nextWorkDayStart) {
			// Полный рабочий день — 8 часов
			if nextWorkDayEnd.Before(end) {
				total += HoursPerDay
			} else {
				// Последний день — частичный
				if end.After(nextWorkDayStart) {
					total += float32(end.Sub(nextWorkDayStart).Hours())
				}
			}
		}
		// Переход на следующий день
		nextDay = nextWorkDayStart.AddDate(0, 0, 1)
		year, month, day = nextDay.Date()
		nextWorkDayStart = time.Date(year, month, day, WorkStartHour, 0, 0, 0, loc)
		nextWorkDayEnd = time.Date(year, month, day, WorkEndHour, 0, 0, 0, loc)
	}

	return total
}
