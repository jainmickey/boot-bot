package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/lestrrat-go/ical"
)

type Event struct {
	summary, eventType string
	startDate, endDate time.Time
}

type EventsSorted struct {
	eventType string
	events    []Event
}

var envVars = map[string]string{}

func AddAndValidateEnvVars() {
	envVars["JustWorksUrl"] = os.Getenv("JustWorksUrl")
	fmt.Println("Env", envVars)

	for k := range envVars {
		if envVars[k] == "" {
			log.Fatal(fmt.Sprintf("$%s must be set", k))
		}
	}
}

func (ev *Event) setType(evType string) {
	ev.eventType = evType
}

func formatDate(t time.Time) string {
	suffix := "th"
	switch t.Day() {
	case 1, 21, 31:
		suffix = "st"
	case 2, 22:
		suffix = "nd"
	case 3, 23:
		suffix = "rd"
	}
	return t.Format("Monday, January 2" + suffix)
}

func createPTOText(event Event) (string, error) {
	reLeadcloseWhtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	reInsideWhtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
	startDateFormatted := formatDate(event.startDate)
	endDateFormatted := formatDate(event.endDate)
	dateSuffix := fmt.Sprintf("till %s", endDateFormatted)
	if int(event.endDate.Weekday()) >= 5 && int(event.endDate.Sub(event.startDate).Hours()/24) < 7 {
		dateSuffix = "and starts again after the weekend."
	}

	dateMessage := fmt.Sprintf(" from %s %s", startDateFormatted, dateSuffix)
	strippedMessage := strings.Replace(event.summary, "PTO", "", -1)
	eventTypeString := fmt.Sprintf("(%s)", event.eventType)
	strippedMessage = strings.Replace(strippedMessage, eventTypeString, "", -1)
	strippedMessage = reLeadcloseWhtsp.ReplaceAllString(strippedMessage, "")
	strippedMessage = reInsideWhtsp.ReplaceAllString(strippedMessage, " ")
	message := fmt.Sprintf("- %s %s\n", strippedMessage, dateMessage)
	return message, nil
}

func getEventEmoji(eventType string) (string, error) {
	emojiesMapping := map[string]string{"Vacation": ":palm_tree:", "Working Remotely": ":house_with_garden:"}
	if val, ok := emojiesMapping[eventType]; ok {
		return val, nil
	}
	return "", nil
}

func createEventMessage(sortedEventsList map[string][]Event) (string, error) {
	messaging := "Hey there :wave:, keeping you up to date on who's O.O.O. next week:"
	for key, val := range sortedEventsList {
		emoji, _ := getEventEmoji(key)
		message := fmt.Sprintf("\n\n%s *%s* (%d in total):\n\n", emoji, key, len(val))
		for _, ev := range val {
			fmt.Println("Event", ev.startDate, ev.endDate)
			ptoText, _ := createPTOText(ev)
			message = fmt.Sprintf("%s%s", message, ptoText)
		}
		messaging = fmt.Sprintf("%s%s", messaging, message)
	}
	return messaging, nil
}

func sortCalenderItems(events []Event) (map[string][]Event, error) {
	re := regexp.MustCompile(`\((.*?)\)`)
	sortedEvents := make(map[string][]Event)
	for index := range events {
		split := re.FindStringSubmatch(events[index].summary)
		events[index].setType(split[1])
	}
	for _, ev := range events {
		if val, ok := sortedEvents[ev.eventType]; ok {
			sortedEvents[ev.eventType] = append(val, ev)
		} else {
			sortedEvents[ev.eventType] = []Event{ev}
		}
	}
	return sortedEvents, nil
}

func getByDateRange(fromDate time.Time, toDate time.Time) (string, error) {
	fmt.Println("DateRange")
	out, err := os.Create("/tmp/justWorksCal.ics")
	if err != nil {
		fmt.Println("Error in creating calender file")
		return "Test", err
	}
	defer out.Close()

	fmt.Println("DateRange Dir Created")

	resp, err := http.Get(envVars["JustWorksUrl"])
	if err != nil {
		fmt.Println("Error in fetching calender")
		return "Test", err
	}
	defer resp.Body.Close()

	fmt.Println("DateRange File Fetched")

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Error in saving calender file")
		return "Test", err
	}

	fmt.Println("DateRange File Saved")

	p := ical.NewParser()
	c, err := p.ParseFile("/tmp/justWorksCal.ics")
	fmt.Println("Testing", c, err)

	var eventsList []Event
	for e := range c.Entries() {
		ev, ok := e.(*ical.Event)
		if !ok {
			continue
		}

		layout := "20060102T000000"
		prop, ok := ev.GetProperty("summary")
		if !ok {
			continue
		}
		prop2, ok := ev.GetProperty("dtstart")
		if !ok {
			continue
		}
		prop2Time, err := time.Parse(layout, prop2.RawValue())
		if err != nil {
			fmt.Println("Error", err)
			continue
		}
		prop3, ok := ev.GetProperty("dtend")
		if !ok {
			continue
		}
		prop3Time, err := time.Parse(layout, prop3.RawValue())
		if err != nil {
			continue
		}

		if prop2Time.After(fromDate) && prop3Time.Before(toDate) {
			event := Event{summary: prop.RawValue(), startDate: prop2Time, endDate: prop3Time}
			eventsList = append(eventsList, event)
		}
	}
	fmt.Println("Dates", fromDate, toDate)

	fmt.Println("DateRange Parsed", err)
	sortedEventsList, _ := sortCalenderItems(eventsList)
	fmt.Println("Events List", sortedEventsList)
	message, err := createEventMessage(sortedEventsList)
	if err != nil {
		fmt.Println("Error", err)
	} else {
		fmt.Println("message", message)
	}

	return message, nil
}

func main() {
	AddAndValidateEnvVars()
	start := time.Now()
	start = start.Add(time.Duration(8-int(start.Weekday())) * 24 * time.Hour)
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	end := start.Add(6 * 24 * time.Hour)
	message, _ := getByDateRange(start, end)
	fmt.Println("Final Message", message)
}
