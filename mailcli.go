package main

import (
	"flag"
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"io/ioutil"
	"log"
	"net/mail"
	"os"
)

func loadEnvironment() (string, string, string) {
	from := os.ExpandEnv("${SENDGRID_FROM}")
	username := os.ExpandEnv("${SENDGRID_USER}")
	password := os.ExpandEnv("${SENDGRID_PASS}")

	if from == "" || username == "" || password == "" {
		log.Panic("You need to define SENDGRID_FROM, SENDGRID_USER, SENDGRID_PASS in the environment.")
	}

	return from, username, password
}

func parseRecipients(input *string, optional bool) []*mail.Address {
	recipients, err := mail.ParseAddressList(*input)
	if err != nil && optional == false {
		log.Panic("Could not parse arguments")
	}

	return recipients
}

func main() {
	from, username, password := loadEnvironment()

	to := flag.String("to", "", "Recipients in To")
	cc := flag.String("cc", "", "Recipients in CC")
	bcc := flag.String("bcc", "", "Recipients in BCC")
	subject := flag.String("s", "", "Subject")
	isHtml := flag.Bool("html", false, "Send as HTML")

	flag.Parse()

	sg := sendgrid.NewSendGridClient(username, password)
	message := sendgrid.NewMail()

	message.SetFrom(from)
	message.AddRecipients(parseRecipients(to, false))
	message.AddCcRecipients(parseRecipients(cc, true))
	message.AddBccRecipients(parseRecipients(bcc, true))
	message.SetSubject(*subject)

	b, err := ioutil.ReadAll(os.Stdin)

	if err != nil {
		log.Panic("Could not read stdin")
	}

	if *isHtml {
		message.SetHTML(string(b))
		log.Println("This is HTML")
	} else {
		message.SetText(string(b))
		log.Println("This is Text")
	}

	if r := sg.Send(message); r == nil {
		fmt.Println("Email sent!")
	} else {
		fmt.Println(r)
	}
}
