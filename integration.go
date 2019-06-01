package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/jainmickey/justworks_integration/environment"
	"github.com/jainmickey/justworks_integration/forecast"
	"github.com/jainmickey/justworks_integration/justworks"
	"github.com/jainmickey/justworks_integration/s3"
	"github.com/jainmickey/justworks_integration/slacknotifier"

	"github.com/aws/aws-lambda-go/lambda"
)

type GlobalState struct {
	DailyRunTime  time.Time `json:"daily_run_time"`
	WeeklyRunTime time.Time `json:"weekly_run_time"`
}

func readGlobalStateFile(filename string) (GlobalState, error) {
	file, _ := ioutil.ReadFile(filename)
	fmt.Println("Testing", file)

	data := GlobalState{}
	var rawStrings map[string]string

	err := json.Unmarshal([]byte(file), &rawStrings)
	if err != nil {
		fmt.Println("Error in json unmarshell error: ", err)
		return data, err
	}

	dailyTime, err := time.Parse(time.RFC3339, rawStrings["daily_run_time"])
	if err != nil {
		fmt.Println("Error in time parsing: ", err)
		return data, err
	}
	data.DailyRunTime = dailyTime

	weeklyTime, err := time.Parse(time.RFC3339, rawStrings["weekly_run_time"])
	if err != nil {
		fmt.Println("Error in time parsing: ", err)
		return data, err
	}
	data.WeeklyRunTime = weeklyTime

	fmt.Println("Time", data.DailyRunTime, data.WeeklyRunTime)
	return data, nil
}

func getDateRange() (time.Time, time.Time) {
	start := time.Now()
	if int(start.Weekday()) != 1 {
		start = start.Add(time.Duration(8-int(start.Weekday())) * 24 * time.Hour)
	}
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	end := start.Add(6 * 24 * time.Hour)
	return start, end
}

func weeklySlackMessage(envVars map[string]string) {
	start, end := getDateRange()
	fmt.Println("Start End", start, end)
	eventsList, _ := justworks.GetByDateRange(start, end, envVars)
	eventsList, _ = justworks.FilterEventsForVacationAndRemote(eventsList)
	sortedEventsList, _ := justworks.SortCalenderItems(eventsList)
	message, _ := justworks.CreateEventMessage(sortedEventsList)
	fmt.Println("Final Message", message)
	slackConn := slacknotifier.New(envVars["SlackWebhookURL"])
	slackConn.Notify(message)
}

func dailyProductAccountsSlackMessage(envVars map[string]string) {
	eventsList, _ := justworks.GetTodaysEvents(envVars)
	eventsList, _ = justworks.FilterEventsForVacationAndRemote(eventsList)
	forecastPeople, _ := forecast.GetPeopleDetailsFromForecast(envVars)
	eventsList, _ = forecast.FilterEventsForProductAndAccountsPeople(forecastPeople, eventsList)
	sortedEventsList, _ := justworks.SortCalenderItems(eventsList)
	message, _ := justworks.CreateProductAndAccountMessage(sortedEventsList)
	fmt.Println("Final Message", message)
	slackConn := slacknotifier.New(envVars["ProductAndAccountSlackWebhookURL"])
	slackConn.Notify(message)
}

func dailyForecast(envVars map[string]string) {
	start := time.Now()
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())

	eventsList, _ := justworks.GetByStartDate(start, envVars)
	eventsList, _ = justworks.FilterEventsForVacation(eventsList)
	forecastPeople, _ := forecast.GetPeopleDetailsFromForecast(envVars)
	forecastPeople, _ = forecast.FilterForcastPeople(forecastPeople, eventsList)
	forecast.CreateProjectAssignmentForecast(forecastPeople, envVars)
}

func HandleLambdaEvent() (string, error) {
	envVars, _ := environment.GetEnvironmentVars()
	globalStateFile := "/tmp/globalState.json"
	s3.DownloadFile(envVars["AWS_STORAGE_BUCKET_NAME"], globalStateFile)
	globalData, err := readGlobalStateFile(globalStateFile)
	dailyDuration := 0
	weeklyDuration := 0
	if err != nil {
		globalData.DailyRunTime = time.Now()
		globalData.WeeklyRunTime = time.Now()
	} else {
		dailyDuration = int(time.Now().Sub(globalData.DailyRunTime).Hours())
		weeklyDuration = int(time.Now().Sub(globalData.WeeklyRunTime).Hours())
		fmt.Println("Duration", dailyDuration, weeklyDuration)
		if dailyDuration < 23 {
			fmt.Println("Ran Already!")
			return "Ran Already!", nil
		}

	}

	justworksFileStatus, err := justworks.DownloadJustWorksFile(envVars)
	if justworksFileStatus == false {
		fmt.Println("Error in fetching justworks file: ", err)
	} else {
		if weeklyDuration == 0 || weeklyDuration > 150 {
			weeklySlackMessage(envVars)
			globalData.WeeklyRunTime = time.Now()
		}
		dailyProductAccountsSlackMessage(envVars)
		globalData.DailyRunTime = time.Now()

		data := struct {
			DailyRunTime  string `json:"daily_run_time"`
			WeeklyRunTime string `json:"weekly_run_time"`
		}{
			DailyRunTime:  globalData.DailyRunTime.Format(time.RFC3339),
			WeeklyRunTime: globalData.WeeklyRunTime.Format(time.RFC3339),
		}
		file, _ := json.Marshal(data)
		_ = ioutil.WriteFile(globalStateFile, file, 0777)

		s3FileUploadStatus, err := s3.UploadFile(envVars["AWS_STORAGE_BUCKET_NAME"], globalStateFile)
		if s3FileUploadStatus == false {
			fmt.Println("Error in uploading s3 file: ", err)
		}

		dailyForecast(envVars)
	}
	return "Executed Successfully!", nil
	// return message, nil
}

func main() {
	HandleLambdaEvent()
	lambda.Start(HandleLambdaEvent)
}
