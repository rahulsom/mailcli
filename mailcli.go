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
		log.Println("You need to define SENDGRID_FROM, SENDGRID_USER, SENDGRID_PASS in the environment.")
		os.Exit(1)
	}

	return from, username, password
}

func parseRecipients(input *string) []*mail.Address {
	recipients, err := mail.ParseAddressList(*input)
	if err != nil {
		return []*mail.Address{}
	}

	log.Println("Recipients are", recipients)
	return recipients
}

func main() {
	from, username, password := loadEnvironment()

	to := flag.String("to", "", "Recipients in To")
	cc := flag.String("cc", "", "Recipients in CC")
	bcc := flag.String("bcc", "", "Recipients in BCC")
	subject := flag.String("s", "", "Subject")
	isHtml := flag.Bool("html", false, "Send as HTML")
	help := flag.Bool("h", false, "Display this help message")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *to == "" && *cc == "" && *bcc == "" {
		log.Println("Need atleast one of to,cc,bcc")
		flag.Usage()
		os.Exit(2)
	}

	if *subject == "" {
		log.Println("Subject cannot be empty")
		flag.Usage()
		os.Exit(3)
	}

	sg := sendgrid.NewSendGridClient(username, password)
	message := sendgrid.NewMail()

	message.SetFrom(from)
	message.AddRecipients(parseRecipients(to))
	message.AddCcRecipients(parseRecipients(cc))
	message.AddBccRecipients(parseRecipients(bcc))
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

//	fmt.Println ("SG",sg)
	if r := sg.Send(message); r == nil {
		fmt.Println("Email sent!")
	} else {
		fmt.Println(r)
	}
}
