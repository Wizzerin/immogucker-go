package notifier

import (
	"fmt"
	"net/smtp"
	"os"

	"github.com/Wizzerin/immogucker-go/internal/models"
)

// SendResults formats and sends an HTML email with the parsed apartments
func SendResults(toEmail string, apartments []models.Apartment) error {
	if len(apartments) == 0 {
		return nil
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	senderEmail := os.Getenv("SMTP_EMAIL")
	senderPass := os.Getenv("SMTP_PASSWORD")

	if senderEmail == "" || senderPass == "" {
		return fmt.Errorf("SMTP_EMAIL or SMTP_PASSWORD environment variables are not set")
	}

	// 1. Build email headers (Subject and Content-Type for HTML)
	subject := "Subject: Immogucker: New apartments found for your search\n"
	toHeader := "To: " + toEmail + "\n"
	mime := "MIME-Version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	// 2. Generate the HTML body
	htmlBody := "<h2>Search Results:</h2><ul>"
	for _, apt := range apartments {
		htmlBody += fmt.Sprintf("<li><a href='%s'>%s</a> - <b>%d €</b></li>", apt.Link, apt.Title, apt.Price)
	}
	htmlBody += "</ul>"

	// Combine headers and body into a single byte array
	msg := []byte(toHeader + subject + mime + htmlBody)

	// 3. Authenticate with the SMTP server
	auth := smtp.PlainAuth("", senderEmail, senderPass, smtpHost)

	// 4. Send the email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, senderEmail, []string{toEmail}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
