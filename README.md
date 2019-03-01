# Boot-bot
### Justworks Company Calendar Slack Bot

A bot that sends out a `Out of Office (O.O.O)` message for all registered `Personal Time Off (P.T.O)` requests.
Basically schedules a Cloudwatch Event and executes a Lambda Function to execute:

- Fetch ICS file from justworks and filter out events with type `Vacation` in the coming week.
- Fetch People data from Forecast and create time off for based on Justworks events.
- Sends a message to slack

Example of the slack message is:

```
Hey there :wave:, keeping you up to date on who's O.O.O. next week:

:house_with_garden: *Working Remotely* (2 in total):

- Rachel J. is out on Thursday
- Nathan J. from Wednesday and starts after the weekend

:palm_tree: *Vacation* (3 in total):

- Nathan J. on Tuesday
- Rob S. from Tuesday to Thursday
- Meredith F. from Thursday and starts after the weekend
```

## Setup

### Configuration

To fetch data from Justworks, Forecast and sending message to Slack requires some configuration in the form of environment variables:

- No Forecast api available, for the time being:
  ```
  You will need a Forecast account, accountId and authorization token.

  The easiest way to determine your accountId and authorization token is by logging in to Forecast from Google Chrome and using the web inspector > Network tab to see one of the request(s) being made.

  Observe a request and note the accoundId and authorization from the request header.
  ```

  Set Environment variables: ForeCastApiToken, ForeCastApiAccountId, ForeCastApiTimeOffProjectID
  ```
  export ForeCastApiToken={Token}
  export ForeCastApiAccountId={Account Id}
  export ForeCastApiTimeOffProjectId={Project Id}
  ```
- Justworks data fetched in the form of `ics` file. It requires justworks account `ical` url. To get that url visit:
  `https://secure.justworks.com/calendar`
  On top left click `Subscribe via iCal`. It'll show a url, copy that and set it in the environment variable `JustWorksUrl`.
- Slack integration can be setup using webhook url in environment variable `SlackWebhookURL`

To run:

#### Install the dependencies
```
go get github.com/lestrrat-go/ical
go get github.com/aws/aws-lambda-go/lambda
```

#### Build and run the binary
```
go build -o integration integration.go
./integration
```

## Note

To fetch data from Justworks, Forecast and sending message to Slack requires some configuration in the form of environment variables:

- Justworks url changes time to time. Need to add error handler to notify about this.
- To deploy build for linux instead of osx. It can be easily done using command:
  `GOARCH=amd64 GOOS=linux go build -o integration integration.go`