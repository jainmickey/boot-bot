package environment

import (
	"fmt"
	"log"
	"os"
)

func getEnvWithDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetEnvironmentVars() (map[string]string, error) {
	envVars := map[string]string{}
	envVars["JustWorksUrl"] = getEnvWithDefault("JustWorksUrl", "")
	envVars["ForeCastApiUrl"] = getEnvWithDefault("ForeCastApiUrl", "https://api.forecastapp.com")
	envVars["ForeCastApiToken"] = getEnvWithDefault("ForeCastApiToken", "")
	envVars["ForeCastApiAccountId"] = getEnvWithDefault("ForeCastApiAccountId", "")
	envVars["ForeCastApiTimeOffProjectID"] = getEnvWithDefault("ForeCastApiTimeOffProjectID", "")
	envVars["SlackWebhookURL"] = getEnvWithDefault("SlackWebhookURL", "")
	envVars["ProductAndAccountSlackWebhookURL"] = getEnvWithDefault("ProductAndAccountSlackWebhookURL", "")
	envVars["AWS_STORAGE_BUCKET_NAME"] = getEnvWithDefault("AWS_STORAGE_BUCKET_NAME", "")
	envVars["DefaultFromEmail"] = getEnvWithDefault("DefaultFromEmail", "")
	envVars["AdminEmail"] = getEnvWithDefault("AdminEmail", "backend@fueled.com")
	envVars["EmailHost"] = getEnvWithDefault("EmailHost", "")
	envVars["EmailHostPassword"] = getEnvWithDefault("EmailHostPassword", "")
	envVars["EmailHostUser"] = getEnvWithDefault("EmailHostUser", "")
	envVars["EmailPort"] = getEnvWithDefault("EmailPort", "")
	fmt.Println("Env", envVars)

	for k := range envVars {
		if envVars[k] == "" {
			log.Fatal(fmt.Sprintf("$%s must be set", k))
		}
	}

	return envVars, nil
}
