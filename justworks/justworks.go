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

var vacation = "Vacation"
var workingRemotely = "Working Remotely"
var workingHome = "Working from Home (Same Timezone"
var casualLeave = "Casual Leave - Noida Team Only"
var sickLeave = "Sick Leave"
var availableTypes = []string{vacation, workingRemotely, casualLeave, sickLeave, workingHome}
var vacationTypes = []string{vacation, casualLeave, sickLeave}
var productAccountsVacationTypes = []string{vacation}

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
	return t.Format("Mon, 2" + suffix + " January")
	// return t.Format("Monday, January 2" + suffix)
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

func getEventEmoji(eventType string) (string, error) {
	emojiesMapping := map[string]string{vacation: ":beach_with_umbrella:", workingRemotely: ":house_with_garden:",
		casualLeave: ":beach_with_umbrella:", sickLeave: ":face_with_thermometer:", workingHome: ":house_with_garden:"}
	if val, ok := emojiesMapping[eventType]; ok {
		return val, nil
	}
	errorMessage := fmt.Sprintf("Emoji for %s doesn't exists!", eventType)
	return "", errors.New(errorMessage)
}

func createPTOText(event Event, upcoming bool) (string, error) {
	startDateFormatted := formatDate(event.startDate)
	endDateFormatted := formatDate(event.endDate)
	duration := event.endDate.Sub(event.startDate).Hours()
	if int(duration)%24 == 0 {
		endDateFormatted = formatDate(event.endDate.Add(-24 * time.Hour))
	}

	dateMessage := fmt.Sprintf("%s ↔︎ %s", startDateFormatted, endDateFormatted)
	if int(duration) < 25 {
		dateMessage = fmt.Sprintf("%s", startDateFormatted)
	}

	message := fmt.Sprintf("- %s - %s\n", event.name, dateMessage)
	if upcoming == true {
		eventType := event.eventType
		if eventType == workingHome {
			eventType = workingRemotely
		}
		emoji, err := getEventEmoji(eventType)
		if err == nil {
			message = fmt.Sprintf("%s %s - %s\n", emoji, event.name, dateMessage)
		}
	}
	return message, nil
}

func CreateEventMessage(sortedEventsList map[string][]Event) (string, error) {
	messaging := "Hey there :wave:, keeping you up to date on who's O.O.O. this week"
	for key, val := range sortedEventsList {
		fmt.Println("Emoji", key)
		emoji, err := getEventEmoji(key)
		if err == nil {
			message := ""
			if len(val) > 0 {
				message = fmt.Sprintf("\n\n%s *%s* (%d in total):\n\n", emoji, key, len(val))
				for _, ev := range val {
					fmt.Println("Event", ev.summary, ev.startDate, ev.endDate)
					// Bool specify its upcoming event or not
					ptoText, _ := createPTOText(ev, false)
					message = fmt.Sprintf("%s%s", message, ptoText)
				}
			}
			messaging = fmt.Sprintf("%s%s", messaging, message)
		}
	}
	return messaging, nil
}

func CreateProductAndAccountMessage(sortedEventsList map[string][]Event, upcoming bool) (string, error) {
	messaging := "Hey there :wave:, keeping you up to date on who's *OOO* in *Product and Accounts* team today:"
	upcomigEvents := true
	if upcoming == true {
		messaging = "\n*Upcoming OOOs (all Fueled employees)*:\n\n"
		upcomigEvents = false
	}
	for key, val := range sortedEventsList {
		emoji, err := getEventEmoji(key)
		if err == nil {
			message := ""
			if upcoming == false {
				message = fmt.Sprintf("\n\n%s *%s*:\n\nNo one is on *%s* today!!", emoji, key, key)
				if len(val) > 0 {
					message = fmt.Sprintf("\n\n%s *%s* (%d in total):\n\n", emoji, key, len(val))
				}
			}
			for _, ev := range val {
				fmt.Println("Event", ev.summary, ev.startDate, ev.endDate)
				ptoText, _ := createPTOText(ev, upcoming)
				message = fmt.Sprintf("%s%s", message, ptoText)
				upcomigEvents = true
			}
			messaging = fmt.Sprintf("%s%s", messaging, message)
		}
	}
	if upcomigEvents == false {
		messaging = fmt.Sprintf("%s\n\nNothing for the upcoming week yet!!\n", messaging)

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

func SortCalenderItems(events []Event, forProductAccountPeople, upcoming bool) (map[string][]Event, error) {
	sortedEvents := map[string][]Event{
		vacation:        []Event{},
		workingRemotely: []Event{},
	}
	if forProductAccountPeople == false {
		sortedEvents[casualLeave] = []Event{}
		sortedEvents[sickLeave] = []Event{}
	}
	for _, ev := range events {
		if len(ev.eventType) > 0 {
			if sort.SearchStrings(availableTypes, ev.eventType) <= len(availableTypes) {
				eventType := ev.eventType
				if eventType == workingHome {
					eventType = workingRemotely
				}
				if val, ok := sortedEvents[eventType]; ok {
					if upcoming == true {
						// ---- For upcoming events merge all in one group to be sorted together ----------
						sortedEvents[vacation] = append(sortedEvents[vacation], ev)
					} else {
						sortedEvents[eventType] = append(val, ev)
					}
				}
			}
		}
	}

	// ------- Sorting by Start Date ----------------------
	for _, val := range sortedEvents {
		sort.Slice(val, func(i, j int) bool {
			return val[j].startDate.After(val[i].startDate)
		})
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

func GetUpcomingEvents(envVars map[string]string) ([]Event, error) {
	var eventsList []Event
	start := time.Now().UTC()
	start = start.Add(24 * time.Hour)
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	startAMinuteBefore := start.Add(-1 * time.Minute)
	daysForUpcoming := time.Duration(7 + (6 - int(start.Weekday())))
	end := start.Add(daysForUpcoming * 24 * time.Hour)

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

		if prop2Time.After(startAMinuteBefore) && prop2Time.Before(end) {
			event := Event{summary: prop.RawValue(), startDate: prop2Time, endDate: prop3Time}
			eventsList = append(eventsList, event)
		}
	}

	eventsList, _ = setTypeNameOfEvent(eventsList)
	return eventsList, nil
}

func FilterEventsForVacation(events []Event) ([]Event, error) {
	var filteredEventsList []Event
	for _, ev := range events {
		if len(ev.eventType) > 0 {
			if sort.SearchStrings(vacationTypes, ev.eventType) <= len(vacationTypes) {
				filteredEventsList = append(filteredEventsList, ev)
			}
		}
	}
	return filteredEventsList, nil
}

func FilterEventsForVacationAndRemote(events []Event) ([]Event, error) {
	var filteredEventsList []Event
	for _, ev := range events {
		if len(ev.eventType) > 0 {
			if sort.SearchStrings(availableTypes, ev.eventType) <= len(availableTypes) {
				filteredEventsList = append(filteredEventsList, ev)
			}
		}
	}
	return filteredEventsList, nil
}
