---
title: Google Calendar CLI
date: 2020-04-13
description: A simple tool to access Google Calendar API from CLI.
categories:
    - Programming
tags:
  - golang
  - cobra
---
Our company allows us to work from home some days a week. To do that, we have to create an event in Google Calendar.

I created this tool to run it from CLI.

First, take a look at this [quickstart](https://developers.google.com/calendar/quickstart/go).

Create initical code by running:

```shell script
$ cobra init

$ tree -L 2
.
├── LICENSE
├── cmd
│   └── root.go
├── main.go
```

Create `event` command:

```shell script
$ cobra add event
$ cobra add insert -p 'eventCmd'

$ tree -L 2
.
├── LICENSE
├── cmd
│   ├── event.go
│   ├── event_insert.go
│   └── root.go
├── main.go
```

Create a new Calendar service:

```go
func newCalendar() (*calendar.Service, error) {
	b, err := ioutil.ReadFile(viper.ConfigFileUsed())
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file from config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.New(client)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve calendar client: %v", err)
	}

	return srv, nil
}
```

Get calendar ID from summary:

```go
func getCalendarID(srv *calendar.Service, calendarSummary string) (string, error) {
	if calendarSummary == "" {
		return defaultCalendar, nil
	}

	var calendarID string
	for {
		calendarList, err := srv.CalendarList.List().Do()
		if err != nil {
			return "", fmt.Errorf("Cannot get calendars list: %v", err)
		}
		for _, item := range calendarList.Items {
			if item.Summary == calendarSummary {
				calendarID = item.Id
				break
			}
		}
		pageToken := calendarList.NextPageToken
		if pageToken == "" {
			break
		}
	}
	return calendarID, nil
}
```

Get flag values:

```go
		calendarSummary, err := cmd.Flags().GetString("calendar")
		if err != nil {
			log.Fatalf("Unable to get value of calendar flag: %v", err)
		}

		calendarID, err := getCalendarID(srv, calendarSummary)
		if err != nil {
			log.Fatalf("Unable to get calendar ID of %s: %v", calendarSummary, err)
		}

		eventTitle, err := cmd.Flags().GetString("title")
		if err != nil {
			log.Fatalf("Unable to get value of title flag: %v", err)
		}
```

And create an event:

```go
		event := &calendar.Event{
			Summary: eventTitle,
		}
		event, err = srv.Events.Insert(calendarID, event).Do()
		if err != nil {
			log.Fatalf("Unable to create event: %v\n", err)
		}

```

Whenever I want to WFH, just run:

```shell script
$ gcal event insert -c "Company" -t "Quan Tong - WFH"
```

GitHub repository: https://github.com/quantonganh/gcal