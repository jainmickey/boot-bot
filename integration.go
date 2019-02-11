package main

import (
	"fmt"
	"time"

	"github.com/jainmickey/justworks_integration/environment"
	"github.com/jainmickey/justworks_integration/forecast"
	"github.com/jainmickey/justworks_integration/justworks"
	"github.com/jainmickey/justworks_integration/slacknotifier"
)

func getDateRange() (time.Time, time.Time) {
	start := time.Now()
	start = start.Add(time.Duration(8-int(start.Weekday())) * 24 * time.Hour)
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	end := start.Add(6 * 24 * time.Hour)
	return start, end
}

func main() {
	envVars, _ := environment.GetEnvironmentVars()
	start, end := getDateRange()
	eventsList, _ := justworks.GetByDateRange(start, end, envVars)
	sortedEventsList, _ := justworks.SortCalenderItems(eventsList)
	message, _ := justworks.CreateEventMessage(sortedEventsList)
	fmt.Println("Final Message", message)
	slackConn := slacknotifier.New(envVars["SlackWebhookURL"])
	slackConn.Notify(message)
	forecastPeople, _ := forecast.GetPeopleDetailsFromForecast(envVars)
	forecastPeople, _ = forecast.FilterForcastPeople(forecastPeople, eventsList)
	fmt.Println("Filtered People", forecastPeople)
	// forecast.CreateProjectAssignmentForecast(forecastPeople)
}
