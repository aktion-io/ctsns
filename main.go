package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aktion-io/ctsns/certstream"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/glassechidna/awsctx/service/snsctx"
	"github.com/gorilla/websocket"
	"github.com/honeycombio/libhoney-go"
	"github.com/pkg/errors"
	"golang.org/x/net/publicsuffix"
	"os"
	"strings"
)

func main() {
	libhoney.Init(libhoney.Config{
		APIKey:  os.Getenv("HONEYCOMB_APIKEY"),
		Dataset: os.Getenv("HONEYCOMB_DATASET"),
	})

	events := make(chan certstream.Event)
	ctx := context.Background()

	go read(ctx, events)
	write(ctx, events)
}

func read(ctx context.Context, events chan certstream.Event) {
	for {
		c, _, err := websocket.DefaultDialer.Dial("wss://certstream.calidog.io", nil)
		if err != nil {
			panic(err)
		}

		err = certstream.EventStream(ctx, c, events)
		if err != nil {
			fmt.Printf("error here: %+v\n", err)
		}
	}
}

func write(ctx context.Context, events chan certstream.Event) {
	sess := session.Must(session.NewSession())
	api := snsctx.New(sns.New(sess), nil)
	//api := &printer{}

	topicArn := os.Getenv("TOPIC_ARN")

	for {
		err := writeOne(ctx, api, topicArn, events)
		if err != nil {
			fmt.Printf("error: %+v\n", err)
		}
	}
}

func writeOne(ctx context.Context, api snsctx.SNS, topicArn string, events chan certstream.Event) error {
	event := <-events
	honeycomb(event)

	_, err := api.PublishWithContext(ctx, &sns.PublishInput{
		TopicArn: &topicArn,
		Message:  aws.String(string(event.Data)),
		MessageAttributes: map[string]*sns.MessageAttributeValue{
			"reverse.dns": reverseDnsAttribute(event.AllDomains),
		},
	})

	return errors.WithStack(err)
}

func honeycomb(cse certstream.Event) {
	for _, domain := range cse.AllDomains {
		etld, _ := publicsuffix.EffectiveTLDPlusOne(domain)
		suffix, icann := publicsuffix.PublicSuffix(domain)

		//prefix := strings.TrimSuffix(domain, suffix)
		//prefix = prefix[:len(prefix)-1] // chomp off trailing period

		ev := libhoney.NewEvent()
		ev.Add(map[string]interface{}{
			"domain": domain,
			//"prefix": prefix,
			"etld":   etld,
			"suffix": suffix,
			"icann":  icann,
		})

		parts := strings.Split(domain, ".")
		for idx := 0; idx < len(parts) && idx < 8; idx++ {
			ev.AddField(fmt.Sprintf("nth_%d", idx), parts[len(parts)-1-idx])
		}

		ev.Send()
	}

	//ev := libhoney.NewEvent()
	//ev.AddField("cert", cse.Data)
	//ev.Send()
}

func reverseDnsAttribute(input []string) *sns.MessageAttributeValue {
	var reversed []string

	for _, dns := range input {
		rejoined := reverseDns(dns)
		reversed = append(reversed, rejoined)
	}

	val, _ := json.Marshal(reversed)
	return &sns.MessageAttributeValue{
		DataType:    aws.String("String.Array"),
		StringValue: aws.String(string(val)),
	}
}

func reverseDns(dns string) string {
	s := strings.Split(dns, ".")

	for left, right := 0, len(s)-1; left < right; left, right = left+1, right-1 {
		s[left], s[right] = s[right], s[left]
	}

	rejoined := strings.Join(s, ".")
	return rejoined
}

type printer struct {
	snsctx.SNS
}

func (n *printer) PublishWithContext(ctx context.Context, input *sns.PublishInput, opts ...request.Option) (*sns.PublishOutput, error) {
	attr := input.MessageAttributes["reverse.dns"].StringValue

	var slice []string
	_ = json.Unmarshal([]byte(*attr), &slice)

	fmt.Printf("%+v\n", slice)
	return &sns.PublishOutput{}, nil
}
