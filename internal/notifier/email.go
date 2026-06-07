package notifier

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"os"
	"time"

	"github.com/Wizzerin/immogucker-go/internal/excel"
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

	// 1. Generate the Excel file in memory
	excelBuf, err := excel.GenerateResults(apartments)
	if err != nil {
		return fmt.Errorf("failed to generate excel attachment: %w", err)
	}

	// 2. Define a unique boundary for the multipart message
	boundary := fmt.Sprintf("immogucker-boundary-%d", time.Now().UnixNano())

	// 3. Build email headers
	var body bytes.Buffer
	body.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	body.WriteString("Subject: Immogucker: New apartments found for your search\r\n")
	body.WriteString("MINE-Version: 1.0\r\n")
	body.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
	body.WriteString("\r\n")

	// 4. Append HTML body part
	body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	body.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")

	htmlContent := "<h2>Search Results:</h2><ul>"
	for _, apt := range apartments {
		htmlContent += fmt.Sprintf("<li><a href='%s'>%s</a> - <b>%d €</b></li>", apt.Link, apt.Title, apt.Price)
	}
	htmlContent += "</ul>\r\n\r\n"
	body.WriteString(htmlContent)

	// 5. Append Excel attachment part
	body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	body.WriteString("Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet; name=\"results.xlsx\"\r\n")
	body.WriteString("Content-Disposition: attachment; filename=\"results.xlsx\"\r\n")
	body.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")

	// Encode the Excel buffer to base64 and split into 76-character lines (RFC 2045 standard)
	encodedExcel := base64.StdEncoding.EncodeToString(excelBuf.Bytes())
	for i := 0; i < len(encodedExcel); i += 76 {
		end := i + 76
		if end > len(encodedExcel) {
			end = len(encodedExcel)
		}
		body.WriteString(encodedExcel[i:end] + "\r\n")
	}
	body.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))

	// 3. Authenticate with the SMTP server
	auth := smtp.PlainAuth("", senderEmail, senderPass, smtpHost)

	// 4. Send the email
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, senderEmail, []string{toEmail}, body.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func SendVerificationEmail(toEmail, username, token string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	senderEmail := os.Getenv("SMTP_EMAIL")
	senderPass := os.Getenv("SMTP_PASSWORD")

	baseURL := os.Getenv("BASE_URL")

	if senderEmail == "" || senderPass == "" {
		return fmt.Errorf("SMTP_EMAIL or SMTP_PASSWORD enviroment variables are not set")
	}

	var body bytes.Buffer
	body.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	body.WriteString("Subject: Immogucker: Please verify your email\r\n")
	body.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")

	verifyLink := fmt.Sprintf("%s/api/v1/auth/verify?token=%s", baseURL, token)

	htmlContent := fmt.Sprintf(`
		<h2>Welcome to Immogucker, %s!</h2>
		<p>Please click the link below to verify your email address and unlock the scraper:</p>
		<p><a href="%s">Verify My Email</a></p>
	`, username, verifyLink)
	body.WriteString(htmlContent)

	auth := smtp.PlainAuth("", senderEmail, senderPass, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, senderEmail, []string{toEmail}, body.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}
