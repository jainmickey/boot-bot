package forecast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/jainmickey/justworks_integration/justworks"
	"github.com/jainmickey/justworks_integration/ses"
	"github.com/jainmickey/justworks_integration/utils"
)

type ForecastPerson struct {
	id, updatedByID, harvesUserID, personalFeedTokenID int
	firstName, lastName, email, login                  string
	admin, archived, subscribed, colorBlind            bool
	avatarURL, updatedAt                               string
	workingDays                                        map[string]bool
	roles                                              []interface{}
	event                                              justworks.Event
}

func CheckProductOrAccountPerson(person ForecastPerson) bool {
	if person.login == "enabled" {
		if utils.Contains(person.roles, []string{"Product", "Accounts"}) == true {
			return true
		}
	}
	return false
}

func (fp *ForecastPerson) setUpdatedByID(updatedByID float64) {
	fp.updatedByID = int(updatedByID)
}

func (fp *ForecastPerson) setHarvestUserID(harvestUserID float64) {
	fp.harvesUserID = int(harvestUserID)
}

func (fp *ForecastPerson) setPersonalFeedTokenID(personalFeedTokenID float64) {
	fp.personalFeedTokenID = int(personalFeedTokenID)
}

func (fp *ForecastPerson) setFirstName(firstName string) {
	fp.firstName = firstName
}

func (fp *ForecastPerson) setLastName(lastName string) {
	fp.lastName = lastName
}

func (fp *ForecastPerson) setEmail(email string) {
	fp.email = email
}

func (fp *ForecastPerson) setLogin(login string) {
	fp.login = login
}

func (fp *ForecastPerson) setUpdatedAt(updatedAt string) {
	fp.updatedAt = updatedAt
}

func (fp *ForecastPerson) setAvatarURL(avatarURL string) {
	fp.avatarURL = avatarURL
}

func (fp *ForecastPerson) setAdmin(admin bool) {
	fp.admin = admin
}

func (fp *ForecastPerson) setArchived(archived bool) {
	fp.archived = archived
}

func (fp *ForecastPerson) setSubscribed(subscribed bool) {
	fp.subscribed = subscribed
}

func (fp *ForecastPerson) setColorBlind(colorBlind bool) {
	fp.colorBlind = colorBlind
}

func (fp *ForecastPerson) setRoles(roles []interface{}) {
	fp.roles = roles
}

func (fp *ForecastPerson) setEvent(event justworks.Event) {
	fp.event = event
}

func CreateProjectAssignmentForecast(forecastPeople []ForecastPerson, envVars map[string]string) {
	dateLayout := "2006-01-02"
	assignmentURL := fmt.Sprintf("%s/assignments", envVars["ForeCastApiUrl"])
	client := &http.Client{}
	for _, fp := range forecastPeople {
		endDate := fp.event.EndDate()
		if int(endDate.Weekday()) == 0 {
			endDate = endDate.Add(time.Duration(-2 * 24 * time.Hour))
		} else if int(endDate.Weekday()) == 6 {
			endDate = endDate.Add(time.Duration(-1 * 24 * time.Hour))
		}
		var jsonStr = []byte(fmt.Sprintf(`{"assignment":{"start_date":"%s","end_date":"%s","allocation":null,"active_on_days_off":false,
										 "repeated_assignment_set_id":null, "project_id":"%s","person_id":"%d","placeholder_id":null}}`,
			fp.event.StartDate().Format(dateLayout), endDate.Format(dateLayout), envVars["ForeCastApiTimeOffProjectID"], fp.id))
		req, _ := http.NewRequest("POST", assignmentURL, bytes.NewBuffer(jsonStr))
		req.Header.Add("authorization", fmt.Sprintf("Bearer %s", envVars["ForeCastApiToken"]))
		req.Header.Add("forecast-account-id", envVars["ForeCastApiAccountId"])
		req.Header.Add("content-type", "application/json; charset=UTF-8")
		fmt.Println("Requesting Forecast!!", req)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error in Forecast Assignment", err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		fmt.Println("Assignment Response", fp.email, string(body))
	}
}

func FilterForcastPeople(forcastPeople []ForecastPerson, filteredEvents []justworks.Event) ([]ForecastPerson, error) {
	var filteredForecastPeople []ForecastPerson
	vacation := "Vacation"
	casualLeave := "Casual Leave - Noida Team Only"
	for _, ev := range filteredEvents {
		for _, fp := range forcastPeople {
			fpName := fmt.Sprintf("%s %c.", fp.firstName, fp.lastName[0])
			if (fpName == ev.Name()) && ((ev.EventType() == vacation) || (ev.EventType() == casualLeave)) {
				fp.setEvent(ev)
				filteredForecastPeople = append(filteredForecastPeople, fp)
			}
		}
	}
	return filteredForecastPeople, nil
}

func FilterEventsForProductAndAccountsPeople(forcastPeople []ForecastPerson, events []justworks.Event) ([]justworks.Event, error) {
	var filteredEvents []justworks.Event
	for _, ev := range events {
		for _, fp := range forcastPeople {
			fpName := fmt.Sprintf("%s %c.", fp.firstName, fp.lastName[0])
			if CheckProductOrAccountPerson(fp) == true && fpName == ev.Name() {
				fp.setEvent(ev)
				filteredEvents = append(filteredEvents, ev)
			}
		}
	}
	return filteredEvents, nil
}

func GetPeopleDetailsFromForecast(envVars map[string]string) ([]ForecastPerson, error) {
	fmt.Println("Forecast People")
	var forcastPeople []ForecastPerson

	peopleURL := fmt.Sprintf("%s/people", envVars["ForeCastApiUrl"])
	client := &http.Client{}
	req, _ := http.NewRequest("GET", peopleURL, nil)
	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", envVars["ForeCastApiToken"]))
	req.Header.Add("forecast-account-id", envVars["ForeCastApiAccountId"])

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error in fetching Forecast People", err)
		emailSubject := "Error in Forecast Integration"
		emailBody := fmt.Sprintf("Forecast token expired: %s", err)
		ses.SendEmailSMTP(envVars["DefaultFromEmail"], envVars["AdminEmail"], emailSubject, emailBody, envVars)
		return forcastPeople, nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var raw map[string][]map[string]interface{}
	json.Unmarshal([]byte(body), &raw)
	for index := range raw["people"] {
		person := ForecastPerson{
			id: int(raw["people"][index]["id"].(float64)),
		}
		roles := raw["people"][index]["roles"]
		if roles != nil {
			person.setRoles(roles.([]interface{}))
		}
		updatedByID := raw["people"][index]["updated_by_id"]
		if updatedByID != nil {
			person.setUpdatedByID(updatedByID.(float64))
		}
		harvestUserID := raw["people"][index]["harvest_user_id"]
		if harvestUserID != nil {
			person.setHarvestUserID(harvestUserID.(float64))
		}
		personalFeedTokenID := raw["people"][index]["personal_feed_token_id"]
		if personalFeedTokenID != nil {
			person.setPersonalFeedTokenID(personalFeedTokenID.(float64))
		}
		firstName := raw["people"][index]["first_name"]
		if firstName != nil {
			person.setFirstName(firstName.(string))
		}
		lastName := raw["people"][index]["last_name"]
		if lastName != nil {
			person.setLastName(lastName.(string))
		}
		email := raw["people"][index]["email"]
		if email != nil {
			person.setEmail(email.(string))
		}
		login := raw["people"][index]["login"]
		if login != nil {
			person.setLogin(login.(string))
		}
		updatedAt := raw["people"][index]["updated_at"]
		if updatedAt != nil {
			person.setUpdatedAt(updatedAt.(string))
		}
		avatarURL := raw["people"][index]["avatar_url"]
		if avatarURL != nil {
			person.setAvatarURL(avatarURL.(string))
		}
		admin := raw["people"][index]["admin"]
		if admin != nil {
			person.setAdmin(admin.(bool))
		}
		archived := raw["people"][index]["archived"]
		if archived != nil {
			person.setArchived(archived.(bool))
		}
		subscribed := raw["people"][index]["subscribed"]
		if subscribed != nil {
			person.setSubscribed(subscribed.(bool))
		}
		colorBlind := raw["people"][index]["color_blind"]
		if colorBlind != nil {
			person.setColorBlind(colorBlind.(bool))
		}
		forcastPeople = append(forcastPeople, person)
	}

	// fmt.Println("Forecast People: ", forcastPeople)
	return forcastPeople, nil
}
