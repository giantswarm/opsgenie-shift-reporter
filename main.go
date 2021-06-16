package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nlopes/slack"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/opsgenie/opsgenie-go-sdk-v2/schedule"
)

const (
	START_DATE    = 20
	HOUR_TO_CHECK = 22
)

var (
	scheduleNames = []string{
		"cloud_kaas_schedule",
		"empowerment_schedule",
		"onprem_kaas_schedule",
	}

	opsgenieKey string
	slackKey    string
)

func init() {
	flag.StringVar(&opsgenieKey, "opsgenie-key", "", "API key for Opsgenie")
	flag.StringVar(&slackKey, "slack-key", "", "API key for Slack")
}

// getDates takes a year, month, hour, and start date (e.g: the 20th),
// and returns time.Times at that year, month, and hour for each day in that month.
// e.g: 2021, 5, 22, 20 -> [(2021, 5, 20, 22), (2021, 5, 21, 22), ..., (2021, 6, 19, 22)]
func getDates(year int, month time.Month, hour int, startDate int) ([]time.Time, error) {
	times := []time.Time{
		time.Date(year, month, startDate, hour, 0, 0, 0, time.UTC),
	}

	for {
		times = append(times, times[len(times)-1].AddDate(0, 0, 1))

		last := times[len(times)-1]
		if last.Month() == month+1 && last.Day() == startDate-1 {
			break
		}
	}

	return times, nil
}

func getOncaller(client *schedule.Client, scheduleName string, date time.Time) (string, error) {
	schedule, err := client.GetTimeline(context.Background(), &schedule.GetTimelineRequest{
		IdentifierType:  schedule.Name,
		IdentifierValue: scheduleName,
		IntervalUnit:    schedule.Days,
		Date:            &date,
	})
	if err != nil {
		return "", err
	}

	if len(schedule.FinalTimeline.Rotations) != 1 || len(schedule.FinalTimeline.Rotations[0].Periods) == 0 {
		return "", nil
	}

	name := schedule.FinalTimeline.Rotations[0].Periods[0].Recipient.Name

	if strings.HasSuffix(name, "@giantswarm.io") {
		name = strings.TrimSuffix(name, "@giantswarm.io")
	}

	return name, nil
}

func printSummary(client *slack.Client, shifts map[string]int) error {
	names := []string{}
	for name := range shifts {
		names = append(names, name)
	}
	sort.Strings(names)

	attachment := slack.Attachment{}
	for _, name := range names {
		if name == "" {
			continue
		}

		attachment.Fields = append(attachment.Fields, slack.AttachmentField{
			Title: name,
			Value: strconv.Itoa(shifts[name]),
			Short: true,
		})
	}

	if _, _, err := client.PostMessage(
		"#noise-shift-count",
		slack.MsgOptionAsUser(true),
		slack.MsgOptionText(
			fmt.Sprintf("Oncall shifts counts (%v of %v to %v of %v)", START_DATE, time.Now().Month()-1, START_DATE-1, time.Now().Month()),
			false,
		),
		slack.MsgOptionAttachments(attachment),
	); err != nil {
		return err
	}

	return nil
}

func main() {
	flag.Parse()

	if opsgenieKey == "" {
		log.Fatalf("opsgenie key cannot be empty")
	}
	if slackKey == "" {
		log.Fatalf("slack key cannot be empty")
	}

	opsgenieClient, err := schedule.NewClient(&client.Config{ApiKey: opsgenieKey})
	if err != nil {
		log.Fatalf("%v", err)
	}

	slackClient := slack.New(slackKey)

	times, err := getDates(time.Now().Year(), time.Now().Month()-1, HOUR_TO_CHECK, START_DATE)
	if err != nil {
		log.Fatalf("%v", err)
	}

	shifts := map[string]int{}

	for _, schedule := range scheduleNames {
		for _, t := range times {
			oncaller, err := getOncaller(opsgenieClient, schedule, t)
			if err != nil {
				log.Fatalf("%v", err)
			}

			shifts[oncaller]++

			time.Sleep(1 * time.Second) // To avoid rate limiting
		}
	}

	fmt.Println(shifts)

	if err := printSummary(slackClient, shifts); err != nil {
		log.Fatalf("%v", err)
	}
}
