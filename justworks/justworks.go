package justworks

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jainmickey/justworks_integration/ses"
	"github.com/lestrrat-go/ical"
)

type Event struct {
	summary, eventType, name string
	startDate, endDate       time.Time
}

func (ev *Event) StartDate() time.Time {
	return ev.startDate
}

func (ev *Event) EndDate() time.Time {
	return ev.endDate
}

func (ev *Event) Summary() string {
	return ev.summary
}

func (ev *Event) EventType() string {
	return ev.eventType
}

func (ev *Event) Name() string {
	return ev.name
}

func (ev *Event) setType(evType string) {
	ev.eventType = evType
}

func (ev *Event) setNameInEvent(name string) {
	ev.name = name
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

func getNameFromEventSummary(eventSummary string, eventType string) (string, error) {
	reLeadcloseWhtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	reInsideWhtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
	eventTypeString := fmt.Sprintf("(%s)", eventType)
	strippedMessage := strings.Replace(eventSummary, "PTO", "", -1)
	strippedMessage = strings.Replace(strippedMessage, eventTypeString, "", -1)
	strippedMessage = reLeadcloseWhtsp.ReplaceAllString(strippedMessage, "")
	strippedMessage = reInsideWhtsp.ReplaceAllString(strippedMessage, "")
	return strippedMessage, nil
}

func createPTOText(event Event) (string, error) {
	startDateFormatted := formatDate(event.startDate)
	endDateFormatted := formatDate(event.endDate)
	duration := event.endDate.Sub(event.startDate).Hours()
	dateSuffix := fmt.Sprintf("till %s", endDateFormatted)
	if int(event.endDate.Weekday()) >= 5 && int(duration/24) < 7 {
		dateSuffix = "and starts again after the weekend."
	}

	dateMessage := fmt.Sprintf("from %s %s", startDateFormatted, dateSuffix)
	if int(duration) < 25 {
		dateMessage = fmt.Sprintf("on %s", startDateFormatted)
	}
	message := fmt.Sprintf("- %s %s\n", event.name, dateMessage)
	return message, nil
}

func getEventEmoji(eventType string) (string, error) {
	emojiesMapping := map[string]string{"Vacation": ":palm_tree:", "Working Remotely": ":house_with_garden:",
		"Casual Leave - Noida Team Only": ":beach_with_umbrella:", "Sick Leave": ":face_with_thermometer:",
		"Working from Home (Same Timezone": ":house_with_garden:"}
	if val, ok := emojiesMapping[eventType]; ok {
		return val, nil
	}
	errorMessage := fmt.Sprintf("Emoji for %s doesn't exists!", eventType)
	return "", errors.New(errorMessage)
}

func CreateEventMessage(sortedEventsList map[string][]Event) (string, error) {
	messaging := "Hey there :wave:, keeping you up to date on who's O.O.O. this week"
	for key, val := range sortedEventsList {
		fmt.Println("Emoji", key)
		emoji, err := getEventEmoji(key)
		if err == nil {
			message := fmt.Sprintf("\n\n%s *%s* (%d in total):\n\n", emoji, key, len(val))
			for _, ev := range val {
				fmt.Println("Event", ev.summary, ev.startDate, ev.endDate)
				ptoText, _ := createPTOText(ev)
				message = fmt.Sprintf("%s%s", message, ptoText)
			}
			messaging = fmt.Sprintf("%s%s", messaging, message)
		}
	}
	return messaging, nil
}

func CreateProductAndAccountMessage(sortedEventsList map[string][]Event) (string, error) {
	messaging := "Hey there :wave:, keeping you up to date on who's O.O.O. in Product and Accounts team today:"
	for key, val := range sortedEventsList {
		fmt.Println("Emoji", key)
		emoji, err := getEventEmoji(key)
		if err == nil {
			message := fmt.Sprintf("\n\n%s *%s* (%d in total):\n\n", emoji, key, len(val))
			for _, ev := range val {
				fmt.Println("Event", ev.summary, ev.startDate, ev.endDate)
				ptoText, _ := createPTOText(ev)
				message = fmt.Sprintf("%s%s", message, ptoText)
			}
			messaging = fmt.Sprintf("%s%s", messaging, message)
		}
	}
	return messaging, nil
}

func setTypeNameOfEvent(events []Event) ([]Event, error) {
	// re := regexp.MustCompile(`\((.*?)\)`)
	re := regexp.MustCompile(`\((.*?)\)`)
	for index := range events {
		split := re.FindStringSubmatch(events[index].summary)
		if len(split) > 1 {
			eventType := split[1]
			summary := events[index].summary
			name, _ := getNameFromEventSummary(summary, eventType)
			events[index].setType(eventType)
			events[index].setNameInEvent(name)
		}
	}
	return events, nil
}

func SortCalenderItems(events []Event) (map[string][]Event, error) {
	sortedEvents := make(map[string][]Event)
	availableTypes := []string{"Vacation", "Working Remotely", "Casual Leave - Noida Team Only",
		"Sick Leave", "Working from home (Same Timezone"}
	for _, ev := range events {
		if len(ev.eventType) > 0 {
			if sort.SearchStrings(availableTypes, ev.eventType) <= len(availableTypes) {
				if val, ok := sortedEvents[ev.eventType]; ok {
					sortedEvents[ev.eventType] = append(val, ev)
				} else {
					sortedEvents[ev.eventType] = []Event{ev}
				}
			}
		}
	}
	return sortedEvents, nil
}

func DownloadJustWorksFile(envVars map[string]string) (bool, error) {
	_, err := os.Stat("/tmp/justWorksCal.ics")
	if err == nil {
		os.Remove("/tmp/justWorksCal.ics")
	}
	out, err := os.Create("/tmp/justWorksCal.ics")
	if err != nil {
		fmt.Println("Error in creating calender file")
		return false, err
	}
	defer out.Close()

	fmt.Println("Justworks File Created!")

	resp, err := http.Get(envVars["JustWorksUrl"])
	if err != nil {
		fmt.Println("Error in fetching calender")
		emailSubject := "Error in Justworks Integration"
		emailBody := fmt.Sprintf("Justworks link expired: %s", err)
		ses.SendEmailSMTP(envVars["DefaultFromEmail"], envVars["AdminEmail"], emailSubject, emailBody, envVars)
		return false, err
	}
	defer resp.Body.Close()

	fmt.Println("Justworks File Fetched")

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Error in saving calender file")
		return false, err
	}

	fmt.Println("Justworks File Saved")
	return true, nil
}

func GetByDateRange(fromDate time.Time, toDate time.Time, envVars map[string]string) ([]Event, error) {
	var eventsList []Event

	p := ical.NewParser()
	c, err := p.ParseFile("/tmp/justWorksCal.ics")
	if err != nil {
		fmt.Println("Error", err)
		return eventsList, err
	}

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

		if prop2Time.After(fromDate) && prop2Time.Before(toDate) {
			event := Event{summary: prop.RawValue(), startDate: prop2Time, endDate: prop3Time}
			eventsList = append(eventsList, event)
		}
	}

	eventsList, _ = setTypeNameOfEvent(eventsList)
	return eventsList, nil
}

func GetByStartDate(fromDate time.Time, envVars map[string]string) ([]Event, error) {
	var eventsList []Event

	p := ical.NewParser()
	c, err := p.ParseFile("/tmp/justWorksCal.ics")
	if err != nil {
		fmt.Println("Error", err)
		return eventsList, err
	}

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

		if prop2Time.After(fromDate) {
			event := Event{summary: prop.RawValue(), startDate: prop2Time, endDate: prop3Time}
			eventsList = append(eventsList, event)
		}
	}

	eventsList, _ = setTypeNameOfEvent(eventsList)
	return eventsList, nil
}

func GetTodaysEvents(envVars map[string]string) ([]Event, error) {
	var eventsList []Event
	start := time.Now().UTC()
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	startAMinuteBefore := start.Add(-1 * time.Minute)
	end := start.Add(24 * time.Hour)

	p := ical.NewParser()
	c, err := p.ParseFile("/tmp/justWorksCal.ics")
	if err != nil {
		fmt.Println("Error", err)
		return eventsList, err
	}

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

		if (prop2Time.After(startAMinuteBefore) && prop2Time.Before(end)) || (prop2Time.Before(startAMinuteBefore) && prop3Time.After(end)) ||
			(prop3Time.After(start) && prop3Time.Before(end)) {
			event := Event{summary: prop.RawValue(), startDate: prop2Time, endDate: prop3Time}
			eventsList = append(eventsList, event)
		}
	}

	eventsList, _ = setTypeNameOfEvent(eventsList)
	return eventsList, nil
}

func FilterEventsForVacation(events []Event) ([]Event, error) {
	var filteredEventsList []Event
	availableTypes := []string{"Vacation", "Casual Leave - Noida Team Only", "Sick Leave"}
	for _, ev := range events {
		if len(ev.eventType) > 0 {
			if sort.SearchStrings(availableTypes, ev.eventType) <= len(availableTypes) {
				filteredEventsList = append(filteredEventsList, ev)
			}
		}
	}
	return filteredEventsList, nil
}

func FilterEventsForVacationAndRemote(events []Event) ([]Event, error) {
	var filteredEventsList []Event
	availableTypes := []string{"Vacation", "Working Remotely", "Casual Leave - Noida Team Only",
		"Sick Leave", "Working from home (Same Timezone"}
	for _, ev := range events {
		if len(ev.eventType) > 0 {
			if sort.SearchStrings(availableTypes, ev.eventType) <= len(availableTypes) {
				filteredEventsList = append(filteredEventsList, ev)
			}
		}
	}
	return filteredEventsList, nil
}
