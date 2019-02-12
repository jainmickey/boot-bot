# Boot-bot
### Justworks Company Calendar Slack Bot

A bot that sends out a O.O.O. message for all registered P.T.O requests, this for currently only the Justworks ICS File. Basically schedules a Cloudwatch Event and executes a Lambda Function to chat on Slack. Gets next week entries and calculates who is out in that week and shows it i.e.:

Hey there :wave:, keeping you up to date on who's O.O.O. next week:

:house_with_garden: *Working Remotely* (2 in total):

- Rachel J. is out on Thursday
- Nathan J. from Wednesday and starts after the weekend

:palm_tree: *Vacation* (3 in total):

- Nathan J. on Tuesday
- Rob S. from Tuesday to Thursday
- Meredith F. from Thursday and starts after the weekend

## Installation

To run:
```
go get github.com/lestrrat-go/ical
export JustWorksUrl={Url}
ForeCastApiToken={Token}
ForeCastApiAccountId={Account id}
ForeCastApiTimeOffProjectID={Time off project id}
SlackWebhookURL={Webhook Url}
go build -o integration integration.go
./integration
```
