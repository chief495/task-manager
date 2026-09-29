package email

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"
)

type EmailService struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewEmailService(host, port, username, password, from string) *EmailService {
	return &EmailService{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *EmailService) IsConfigured() bool {
	if s == nil {
		return false
	}
	return strings.TrimSpace(s.host) != "" &&
		strings.TrimSpace(s.port) != "" &&
		strings.TrimSpace(s.username) != "" &&
		strings.TrimSpace(s.password) != "" &&
		strings.TrimSpace(s.from) != ""
}

// Отправка письма со сбросом пароля (только HTML)
func (s *EmailService) SendNewPassword(to, newPassword string) error {
	if !s.IsConfigured() {
		return fmt.Errorf("email service is not configured")
	}

	subject := "Сброс пароля"
	return s.sendHTML(to, subject, passwordResetHTML(newPassword))
}

// Отправка уведомления о задаче (HTML + .ics)
func (s *EmailService) SendTaskAssigned(to, assigneeName, assignerName string, taskName, taskDescription string, startTime, endTime *time.Time, priority int, taskID uint) error {
	if !s.IsConfigured() {
		return nil
	}

	subject := "Вам назначена новая задача"
	htmlBody := taskAssignedHTML(
		assigneeName,
		assignerName,
		taskName,
		taskDescription,
		startTime,
		endTime,
		priority,
	)

	icsBody := buildTaskICS(
		to,
		s.from,
		assigneeName,
		taskName,
		taskDescription,
		startTime,
		endTime,
		taskID,
	)

	return s.sendHTMLWithICS(to, subject, htmlBody, "task.ics", icsBody)
}

// Отправка HTML + .ics (multipart/mixed)
func (s *EmailService) sendHTMLWithICS(to, subject, htmlBody, icsFilename, icsBody string) error {
	boundary := fmt.Sprintf("----=_Part_%d", time.Now().UnixNano())

	message := strings.Join([]string{
		fmt.Sprintf("From: %s", s.from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		fmt.Sprintf(`Content-Type: multipart/mixed; boundary="%s"`, boundary),
		"",
		"--" + boundary,
		`Content-Type: text/html; charset="UTF-8"`,
		"Content-Transfer-Encoding: 8bit",
		"",
		htmlBody,
		"",
		"--" + boundary,
		fmt.Sprintf(`Content-Type: text/calendar; method=REQUEST; charset="UTF-8"; name="%s"`, icsFilename),
		"Content-Transfer-Encoding: 8bit",
		fmt.Sprintf(`Content-Disposition: attachment; filename="%s"`, icsFilename),
		"",
		icsBody,
		"",
		"--" + boundary + "--",
		"",
	}, "\r\n")

	return s.smtpSend(to, message)
}

// Отправка обычного HTML (без вложений)
func (s *EmailService) sendHTML(to, subject, body string) error {
	message := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.from, to, subject, body,
	)

	return s.smtpSend(to, message)
}

// Общая функция SMTP-отправки (без дублирования)
func (s *EmailService) smtpSend(to, message string) error {
	log.Printf("[EMAIL] sending to=%s host=%s port=%s from=%s", to, s.host, s.port, s.from)

	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.host,
		MinVersion:         tls.VersionTLS12,
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", s.host, s.port), tlsConfig)
	if err != nil {
		log.Printf("[EMAIL] connection failed: %v", err)
		return fmt.Errorf("failed to connect to smtp server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		log.Printf("[EMAIL] client creation failed: %v", err)
		return fmt.Errorf("failed to create smtp client: %w", err)
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		log.Printf("[EMAIL] auth failed: %v", err)
		return fmt.Errorf("smtp auth failed: %w", err)
	}

	if err = client.Mail(s.from); err != nil {
		log.Printf("[EMAIL] set sender failed: %v", err)
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err = client.Rcpt(to); err != nil {
		log.Printf("[EMAIL] set recipient failed: %v", err)
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		log.Printf("[EMAIL] data writer failed: %v", err)
		return fmt.Errorf("failed to create data writer: %w", err)
	}

	if _, err = writer.Write([]byte(message)); err != nil {
		_ = writer.Close()
		log.Printf("[EMAIL] write failed: %v", err)
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err = writer.Close(); err != nil {
		log.Printf("[EMAIL] server rejected message: %v", err)
		return fmt.Errorf("smtp server rejected message after DATA: %w", err)
	}

	if err = client.Quit(); err != nil {
		log.Printf("[EMAIL] quit failed: %v", err)
		return fmt.Errorf("smtp quit failed: %w", err)
	}

	log.Printf("[EMAIL] sent successfully to=%s", to)
	return nil
}
