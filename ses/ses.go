package ses

import (
	"fmt"
	"log"
	"net/smtp"
)

func SendEmailSMTP(from string, to string, subject string, body string,
	envVars map[string]string) (string, error) {

	host := fmt.Sprintf("%s:%s", envVars["EmailHost"], envVars["EmailPort"])
	fmt.Println("Testing", host, envVars["EmailHostUser"], envVars["EmailHostPassword"], envVars["EmailPort"])
	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	err := smtp.SendMail(host,
		smtp.PlainAuth("", envVars["EmailHostUser"], envVars["EmailHostPassword"], envVars["EmailHost"]),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return "", err
	}

	log.Print("Sent Successfully!")
	return "Sent Successfully!", nil
}
