package email

import (
	"fmt"
	"strings"
	"time"
)

func passwordResetHTML(newPassword string) string {
	return fmt.Sprintf(`
		<html>
		<body>
			<h2>Сброс пароля</h2>
			<p>Ваш новый пароль: <strong>%s</strong></p>
			<p>Рекомендуем сменить пароль после входа в систему.</p>
		</body>
		</html>
	`, newPassword)
}

func taskAssignedHTML(assigneeName, assignerName, taskName, taskDescription string, startTime, endTime *time.Time, priority int) string {
	return fmt.Sprintf(`
		<html>
		<body>
			<h2>Новая задача</h2>
			<p>Здравствуйте, %s.</p>
			<p>На вас назначена новая задача: <strong>%s</strong></p>
			<p><strong>Постановщик:</strong> %s</p>
			<p><strong>Описание:</strong> %s</p>
			<p><strong>Начало:</strong> %s</p>
			<p><strong>Крайний срок:</strong> %s</p>
			<p><strong>Приоритет:</strong> %s</p>
			<p>Во вложении — событие для вашего календаря (.ics).</p>
		</body>
		</html>
	`,
		fallbackString(assigneeName, "коллега"),
		fallbackString(taskName, "Без названия"),
		fallbackString(assignerName, "Не указан"),
		fallbackString(taskDescription, "Не указано"),
		formatTimeForEmail(startTime),
		formatTimeForEmail(endTime),
		taskPriorityLabel(priority),
	)
}

func buildTaskICS(toEmail, organizerEmail, assigneeName, taskName, taskDescription string, startTime, endTime *time.Time, taskID uint) string {
	now := time.Now().UTC()

	start := now.Add(1 * time.Hour)
	if startTime != nil {
		start = startTime.UTC()
	}

	end := start.Add(1 * time.Hour)
	if endTime != nil {
		end = endTime.UTC()
	}

	uid := fmt.Sprintf("task-%d-%d@task-manager.local", taskID, now.UnixNano())
	if taskID > 0 {
		uid = fmt.Sprintf("task-%d@task-manager.local", taskID)
	}

	return fmt.Sprintf(`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Task Manager//Task Manager//RU
CALSCALE:GREGORIAN
METHOD:REQUEST
BEGIN:VEVENT
UID:%s
DTSTAMP:%s
DTSTART:%s
DTEND:%s
SUMMARY:%s
DESCRIPTION:%s
ORGANIZER;CN=Task Manager:MAILTO:%s
ATTENDEE;CN=%s;ROLE=REQ-PARTICIPANT;RSVP=TRUE:MAILTO:%s
STATUS:CONFIRMED
SEQUENCE:0
TRANSP:OPAQUE
BEGIN:VALARM
TRIGGER:-PT15M
ACTION:DISPLAY
DESCRIPTION:Напоминание о задаче: %s
END:VALARM
END:VEVENT
END:VCALENDAR`,
		uid,
		now.Format("20060102T150405Z"),
		start.Format("20060102T150405Z"),
		end.Format("20060102T150405Z"),
		escapeICSText(fallbackString(taskName, "Без названия")),
		escapeICSText(fallbackString(taskDescription, "Не указано")),
		organizerEmail,
		escapeICSText(fallbackString(assigneeName, "Коллега")),
		toEmail,
		escapeICSText(fallbackString(taskName, "Задача")),
	)
}

func escapeICSText(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, ";", `\;`)
	value = strings.ReplaceAll(value, ",", `\,`)
	value = strings.ReplaceAll(value, "\r\n", `\n`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	return value
}

func formatTimeForEmail(value *time.Time) string {
	if value == nil {
		return "Не указано"
	}
	return value.Format("02.01.2006 15:04")
}

func fallbackString(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func taskPriorityLabel(priority int) string {
	switch priority {
	case 2:
		return "Высокий"
	case 1:
		return "Средний"
	default:
		return "Низкий"
	}
}
