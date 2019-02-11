package justworks

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

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
	dateSuffix := fmt.Sprintf("till %s", endDateFormatted)
	if int(event.endDate.Weekday()) >= 5 && int(event.endDate.Sub(event.startDate).Hours()/24) < 7 {
		dateSuffix = "and starts again after the weekend."
	}

	dateMessage := fmt.Sprintf("from %s %s", startDateFormatted, dateSuffix)
	message := fmt.Sprintf("- %s %s\n", event.name, dateMessage)
	return message, nil
}

func getEventEmoji(eventType string) (string, error) {
	emojiesMapping := map[string]string{"Vacation": ":palm_tree:", "Working Remotely": ":house_with_garden:",
		"Casual Leave - Noida Team Only": ":beach_with_umbrella:"}
	if val, ok := emojiesMapping[eventType]; ok {
		return val, nil
	}
	return "", nil
}

func CreateEventMessage(sortedEventsList map[string][]Event) (string, error) {
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

func setTypeNameOfEvent(events []Event) ([]Event, error) {
	re := regexp.MustCompile(`\((.*?)\)`)
	for index := range events {
		split := re.FindStringSubmatch(events[index].summary)
		eventType := split[1]
		summary := events[index].summary
		name, _ := getNameFromEventSummary(summary, eventType)
		events[index].setType(eventType)
		events[index].setNameInEvent(name)
	}
	return events, nil
}

func SortCalenderItems(events []Event) (map[string][]Event, error) {
	sortedEvents := make(map[string][]Event)
	for _, ev := range events {
		if val, ok := sortedEvents[ev.eventType]; ok {
			sortedEvents[ev.eventType] = append(val, ev)
		} else {
			sortedEvents[ev.eventType] = []Event{ev}
		}
	}
	return sortedEvents, nil
}

func GetByDateRange(fromDate time.Time, toDate time.Time, envVars map[string]string) ([]Event, error) {
	fmt.Println("DateRange")
	var eventsList []Event
	out, err := os.Create("/tmp/justWorksCal.ics")
	if err != nil {
		fmt.Println("Error in creating calender file")
		return eventsList, err
	}
	defer out.Close()

	fmt.Println("DateRange Dir Created")

	resp, err := http.Get(envVars["JustWorksUrl"])
	if err != nil {
		fmt.Println("Error in fetching calender")
		return eventsList, err
	}
	defer resp.Body.Close()

	fmt.Println("DateRange File Fetched")

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Error in saving calender file")
		return eventsList, err
	}

	fmt.Println("DateRange File Saved")

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
