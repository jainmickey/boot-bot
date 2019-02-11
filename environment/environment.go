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
	fmt.Println("Env", envVars)

	for k := range envVars {
		if envVars[k] == "" {
			log.Fatal(fmt.Sprintf("$%s must be set", k))
		}
	}

	return envVars, nil
}