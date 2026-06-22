package activatemessaging

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

func generateMessageID() string {
	b := make([]byte, 16)
	rand.Read(b)
	randomPart := hex.EncodeToString(b)
	timestamp := time.Now().UnixNano()

	domain := "pingvi.bot"
	if envDomain := os.Getenv("EMAIL_DOMAIN"); envDomain != "" {
		domain = envDomain
	}

	return fmt.Sprintf("<%d.%s@%s>", timestamp, randomPart, domain)
}

func PushVerifyMessageToEmail(ctx context.Context, verificationCode string, userMail string) (int, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	if verificationCode == "" {
		log.Println("verification code is empty")
		return http.StatusBadRequest, fmt.Errorf("verification code is empty")
	}

	templatePath := filepath.Join("templates", "verification_later.html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Printf("error parsing template from %s: %v", templatePath, err)
		return http.StatusInternalServerError, err
	}

	var htmlBody bytes.Buffer
	data := struct {
		VerificationCode string
	}{
		VerificationCode: verificationCode,
	}

	if err = tmpl.Execute(&htmlBody, data); err != nil {
		log.Printf("error executing template: %v", err)
		return http.StatusInternalServerError, err
	}

	from := "pingvi.bot@gmail.com"

	password := os.Getenv("GMAIL_APP_PASSWORD")

	if password == "" {
		log.Println("GMAIL_APP_PASSWORD environment variable is not set")
		log.Println("Как получить пароль приложения Gmail:")
		log.Println("1. Включите двухфакторную аутентификацию в аккаунте Google")
		log.Println("2. Перейдите на страницу паролей приложений")
		log.Println("3. Выберите 'Почта' и 'Другое', назовите 'PingVi'")
		log.Println("4. Скопируйте полученный 16-значный пароль")
		log.Println("5. Установите переменную окружения: export GMAIL_APP_PASSWORD='xxxx xxxx xxxx xxxx'")
		return http.StatusInternalServerError, fmt.Errorf("Gmail app password not configured")
	}

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	log.Printf("Attempting to send email from %s to %s via %s:%s", from, userMail, smtpHost, smtpPort)

	messageID := generateMessageID()

	htmlContent := htmlBody.String()

	encodedSubject := base64.StdEncoding.EncodeToString([]byte("Код подтверждения PingVi"))

	var emailBuffer bytes.Buffer

	emailBuffer.WriteString(fmt.Sprintf("From: PingVi Bot <%s>\r\n", from))
	emailBuffer.WriteString(fmt.Sprintf("To: %s\r\n", userMail))
	emailBuffer.WriteString(fmt.Sprintf("Subject: =?utf-8?B?%s?=\r\n", encodedSubject))
	emailBuffer.WriteString(fmt.Sprintf("Message-ID: %s\r\n", messageID))
	emailBuffer.WriteString("MIME-Version: 1.0\r\n")
	emailBuffer.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	emailBuffer.WriteString("X-Mailer: PingVi-Mailer/1.0\r\n")

	boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())
	emailBuffer.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
	emailBuffer.WriteString("\r\n")

	emailBuffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	emailBuffer.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	emailBuffer.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	emailBuffer.WriteString("\r\n")
	emailBuffer.WriteString("\r\n")
	emailBuffer.WriteString("\r\n")

	emailBuffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	emailBuffer.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	emailBuffer.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	emailBuffer.WriteString("\r\n")
	emailBuffer.WriteString(htmlContent)
	emailBuffer.WriteString("\r\n")
	emailBuffer.WriteString("\r\n")

	emailBuffer.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	resultChan := make(chan error, 1)

	go func() {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", smtpHost, smtpPort))
		if err != nil {
			resultChan <- fmt.Errorf("TCP connection failed: %w", err)
			return
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, smtpHost)
		if err != nil {
			resultChan <- fmt.Errorf("failed to create SMTP client: %w", err)
			return
		}
		defer client.Close()

		if err = client.Hello("mail.pingvi.bot"); err != nil {
			log.Printf("Warning: failed to set HELO: %v", err)
		}

		tlsConfig := &tls.Config{
			ServerName: smtpHost,
		}
		if err = client.StartTLS(tlsConfig); err != nil {
			resultChan <- fmt.Errorf("STARTTLS failed: %w", err)
			return
		}

		auth := smtp.PlainAuth("", from, password, smtpHost)
		if err = client.Auth(auth); err != nil {
			resultChan <- fmt.Errorf("authentication failed: %w", err)
			return
		}

		if err = client.Mail(from); err != nil {
			resultChan <- fmt.Errorf("MAIL FROM failed: %w", err)
			return
		}

		if err = client.Rcpt(userMail); err != nil {
			resultChan <- fmt.Errorf("RCPT TO failed: %w", err)
			return
		}

		w, err := client.Data()
		if err != nil {
			resultChan <- fmt.Errorf("DATA command failed: %w", err)
			return
		}

		_, err = w.Write(emailBuffer.Bytes())
		if err != nil {
			resultChan <- fmt.Errorf("failed to write message: %w", err)
			return
		}

		err = w.Close()
		if err != nil {
			resultChan <- fmt.Errorf("failed to close DATA: %w", err)
			return
		}

		err = client.Quit()
		resultChan <- err
	}()

	select {
	case err := <-resultChan:
		if err != nil {
			log.Printf("error sending email via Gmail: %v", err)
			return http.StatusServiceUnavailable, err
		}
	case <-ctx.Done():
		log.Printf("email sending timed out")
		return http.StatusGatewayTimeout, fmt.Errorf("smtp timeout: %w", ctx.Err())
	}

	log.Printf("Email sent successfully to %s via Gmail (Message-ID: %s)", userMail, messageID)
	return http.StatusOK, nil
}

func PushPasswordRecoveryMessage(ctx context.Context, link string, userMail string) (int, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	if link == "" {
		log.Println("reset link is empty")
		return http.StatusBadRequest, fmt.Errorf("reset link is empty")
	}

	if userMail == "" {
		log.Println("user email is empty")
		return http.StatusBadRequest, fmt.Errorf("user email is empty")
	}

	templatePath := filepath.Join("templates", "password_recovery.html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Printf("error parsing template from %s: %v", templatePath, err)
		return http.StatusInternalServerError, err
	}

	var htmlBody bytes.Buffer
	data := struct {
		ResetLink string
	}{
		ResetLink: link,
	}

	if err = tmpl.Execute(&htmlBody, data); err != nil {
		log.Printf("error executing template: %v", err)
		return http.StatusInternalServerError, err
	}

	from := "pingvi.bot@gmail.com"
	password := os.Getenv("GMAIL_APP_PASSWORD")

	if password == "" {
		log.Println("GMAIL_APP_PASSWORD environment variable is not set")
		log.Println("Как получить пароль приложения Gmail:")
		log.Println("1. Включите двухфакторную аутентификацию в аккаунте Google")
		log.Println("2. Перейдите на страницу паролей приложений")
		log.Println("3. Выберите 'Почта' и 'Другое', назовите 'PingVi'")
		log.Println("4. Скопируйте полученный 16-значный пароль")
		log.Println("5. Установите переменную окружения: export GMAIL_APP_PASSWORD='xxxx xxxx xxxx xxxx'")
		return http.StatusInternalServerError, fmt.Errorf("gmail app password not configured")
	}

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	log.Printf("Attempting to send password recovery email from %s to %s via %s:%s", from, userMail, smtpHost, smtpPort)

	messageID := generateMessageID()

	htmlContent := htmlBody.String()

	encodedSubject := base64.StdEncoding.EncodeToString([]byte("Восстановление пароля PingVi"))

	var emailBuffer bytes.Buffer

	emailBuffer.WriteString(fmt.Sprintf("From: PingVi Bot <%s>\r\n", from))
	emailBuffer.WriteString(fmt.Sprintf("To: %s\r\n", userMail))
	emailBuffer.WriteString(fmt.Sprintf("Subject: =?utf-8?B?%s?=\r\n", encodedSubject))
	emailBuffer.WriteString(fmt.Sprintf("Message-ID: %s\r\n", messageID))
	emailBuffer.WriteString("MIME-Version: 1.0\r\n")
	emailBuffer.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	emailBuffer.WriteString("X-Mailer: PingVi-Mailer/1.0\r\n")
	emailBuffer.WriteString("X-Priority: 3\r\n")
	emailBuffer.WriteString("X-MSMail-Priority: Normal\r\n")

	boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())
	emailBuffer.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
	emailBuffer.WriteString("\r\n")

	emailBuffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	emailBuffer.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	emailBuffer.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	emailBuffer.WriteString("\r\n")
	plainTextBody := fmt.Sprintf(`Здравствуйте!

Вы запросили восстановление пароля для аккаунта PingVi.

Для установки нового пароля перейдите по ссылке:
%s

Ссылка действительна в течение 1 часа.

Если вы не запрашивали сброс пароля, просто проигнорируйте это письмо.

--
С уважением,
Команда PingVi`, link)
	emailBuffer.WriteString(plainTextBody)
	emailBuffer.WriteString("\r\n")
	emailBuffer.WriteString("\r\n")

	emailBuffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	emailBuffer.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	emailBuffer.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	emailBuffer.WriteString("\r\n")
	emailBuffer.WriteString(htmlContent)
	emailBuffer.WriteString("\r\n")
	emailBuffer.WriteString("\r\n")

	emailBuffer.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	resultChan := make(chan error, 1)

	go func() {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", smtpHost, smtpPort))
		if err != nil {
			resultChan <- fmt.Errorf("tcp connection failed: %w", err)
			return
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, smtpHost)
		if err != nil {
			resultChan <- fmt.Errorf("failed to create SMTP client: %w", err)
			return
		}
		defer client.Close()

		if err = client.Hello("mail.pingvi.bot"); err != nil {
			log.Printf("Warning: failed to set HELO: %v", err)
		}

		tlsConfig := &tls.Config{
			ServerName: smtpHost,
		}
		if err = client.StartTLS(tlsConfig); err != nil {
			resultChan <- fmt.Errorf("STARTTLS failed: %w", err)
			return
		}

		auth := smtp.PlainAuth("", from, password, smtpHost)
		if err = client.Auth(auth); err != nil {
			resultChan <- fmt.Errorf("authentication failed: %w", err)
			return
		}

		if err = client.Mail(from); err != nil {
			resultChan <- fmt.Errorf("MAIL FROM failed: %w", err)
			return
		}

		if err = client.Rcpt(userMail); err != nil {
			resultChan <- fmt.Errorf("RCPT TO failed: %w", err)
			return
		}

		w, err := client.Data()
		if err != nil {
			resultChan <- fmt.Errorf("DATA command failed: %w", err)
			return
		}

		_, err = w.Write(emailBuffer.Bytes())
		if err != nil {
			resultChan <- fmt.Errorf("failed to write message: %w", err)
			return
		}

		err = w.Close()
		if err != nil {
			resultChan <- fmt.Errorf("failed to close DATA: %w", err)
			return
		}

		err = client.Quit()
		resultChan <- err
	}()

	select {
	case err := <-resultChan:
		if err != nil {
			log.Printf("error sending password recovery email via Gmail: %v", err)
			return http.StatusServiceUnavailable, err
		}
	case <-ctx.Done():
		log.Printf("password recovery email sending timed out")
		return http.StatusGatewayTimeout, fmt.Errorf("smtp timeout: %w", ctx.Err())
	}

	log.Printf("Password recovery email sent successfully to %s via Gmail (Message-ID: %s)", userMail, messageID)
	return http.StatusOK, nil
}
