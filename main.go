package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

type Request struct {
	Records []struct {
		SNS struct {
			Type           string `json:"Type"`
			MessageId      string `json:"MessageId"`
			TopicArn       string `json:"TopicArn"`
			Subject        string `json:"Subject"`
			SNSMessage     string `json:"Message"`
			Timestamp      string `json:"Timestamp"`
			UnsubscribeUrl string `json:"UnsubscribeUrl"`
		} `json:"Sns"`
	} `json:"Records"`
}

type SNSMessage struct {
	AlarmName        string `json:"AlarmName"`
	AlarmDescription string `json:"AlarmDescription"`
	NewStateValue    string `json:"NewStateValue"`
	NewStateReason   string `json:"NewStateReason"`
	AWSAccountId     string `json:"AWSAccountId"`
	StateChangeTime  string `json:"StateChangeTime"`
	Region           string `json:"Region"`
	OldStateValue    string `json:"OldStateValue"`
}

type SlackMessage struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	Text           string   `json:"text"`
	PreText        string   `json:"pretext"`
	AuthorName     string   `json:"author_name"`
	AuthorLink     string   `json:"author_link"`
	AuthorIcon     string   `json:"author_icon"`
	FallBack       string   `json:"fallback"`
	CallBackId     string   `json:"callback_id"`
	Color          string   `json:"color"`
	Title          string   `json:"title"`
	TitleLink      string   `json:"title_link"`
	AttachmentType string   `json:"attachment_type"`
	Actions        []Action `json:"actions"`
	Footer         string   `json:"footer"`
	FooterIcon     string   `json:"footer_icon"`
	TimeStamp      string   `json:"ts"`
}

type Action struct {
	Name  string `json:"name"`
	Text  string `json:"text"`
	Type  string `json:"type"`
	Value string `json:"value"`
	Style string `json:"style"`
	Url   string `json:"url"`
}

func handler(request Request) error {
	log.Printf("New Event - %s", request.Records[0].SNS.Subject)

	slackMessage := buildSlackMessage(request)
	postToSlack(slackMessage)

	log.Println("Notification has been sent")
	return nil
}

func buildSlackMessage(request Request) SlackMessage {
	var message = getSNSMessage(request)
	var region = os.Getenv("AWS_DEFAULT_REGION")
	var lambda_function_name = lambdacontext.FunctionName
	var log_group = os.Getenv("LOG_GROUP")

	return SlackMessage{
		Text: "Amazon CloudWatch Logs Event",
		Attachments: []Attachment{
			Attachment{
				Text:           message.NewStateReason,
				AuthorName:     "AWS Lambda",
				AuthorLink:     fmt.Sprintf("https://%s.console.aws.amazon.com/lambda/home?region=%s#/functions/%s?tab=graph", region, region, lambda_function_name),
				AuthorIcon:     "https://www.ng-docs.de/static/img/icons/Compute_AWSLambda_LambdaFunction.png",
				Color:          "danger",
				Title:          request.Records[0].SNS.Subject,
				AttachmentType: "default",
				Actions: []Action{
					Action{
						Type: "button",
						Text: "Cloudwatch Log",
						Url:  fmt.Sprintf("https://%s.console.aws.amazon.com/cloudwatch/home?region=%s#logStream:group=%s;streamFilter=typeLogStreamPrefix", region, region, log_group),
					},
				},
				Footer:     "Cloudwatch - Slack API",
				FooterIcon: "https://www.ng-docs.de/static/img/icons/ManagementTools_AmazonCloudWatch_alarm.png",
				TimeStamp:  strconv.FormatInt(getUnixTS(request.Records[0].SNS.Timestamp), 10),
			},
		},
	}
}

func postToSlack(message SlackMessage) error {
	var encryptedSlackWebhook string = os.Getenv("SLACK_WEBHOOK")
	var decryptedSlackWebhook string = decrypt(encryptedSlackWebhook)
	var SlackWebhookUrl = "https://hooks.slack.com/services/" + decryptedSlackWebhook

	client := &http.Client{}
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", SlackWebhookUrl, bytes.NewBuffer(data))
	//	req, err := http.NewRequest("POST", os.Getenv("SLACK_WEBHOOK"), bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println(resp.StatusCode)
		return err
	}

	return nil
}

func getSNSMessage(request Request) SNSMessage {
	var snsMessage SNSMessage
	err := json.Unmarshal([]byte(request.Records[0].SNS.SNSMessage), &snsMessage)
	if err != nil {
		log.Println(err)
	}

	return snsMessage
}

func getUnixTS(rfc3339TS string) int64 {
	t, err := time.Parse(time.RFC3339, rfc3339TS)
	if err != nil {
		log.Println(err)
		return 0
	}

	return t.Unix()
}

func decrypt(encryptedText string) string {
	kmsClient := kms.New(session.New())
	decodedBytes, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		panic(err)
	}
	input := &kms.DecryptInput{
		CiphertextBlob: decodedBytes,
	}
	response, err := kmsClient.Decrypt(input)
	if err != nil {
		panic(err)
	}
	// Plaintext is a byte array, so convert to string
	return string(response.Plaintext[:])
}

func main() {
	lambda.Start(handler)
}
