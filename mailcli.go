package main

import (
	"flag"
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"github.com/vanng822/go-premailer/premailer"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/mail"
	"os"
	"regexp"
	"strings"
	"path/filepath"
)

type multistring []string

func (i *multistring) String() string {
	return fmt.Sprintf("%s", *i)
}

func (i *multistring) Set(value string) error {
	fmt.Printf("%s\n", value)
	*i = append(*i, value)
	return nil
}

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

func extractTitle(input *string) *string {
	r := strings.NewReader(*input)

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil
	}

	retval := doc.Find("title").First().Text()

	return &retval
}

func parseRecipients(input *string) []*mail.Address {
	recipients, err := mail.ParseAddressList(*input)
	if err != nil {
		return []*mail.Address{}
	}

	log.Println("Recipients are", recipients)
	return recipients
}

func setTextBody(message *sendgrid.SGMail, inputString string) {
	message.SetText(inputString)
	log.Println("This is Text")
}

func setHtmlBody(message *sendgrid.SGMail, inputString string) {
	// This fixes a bug in andybalholm/cascadia in dealing with :: in css for somethings.
	regex := regexp.MustCompile("(?m)^.*::.*$")
	inputString = regex.ReplaceAllLiteralString(inputString, "")

	// This turns stylesheets into inline styles so email clients respect the css.
	prem := premailer.NewPremailerFromString(inputString, premailer.NewOptions())
	htmlString, err := prem.Transform()
	if err != nil {
		log.Fatal(err)
	}

	message.SetHTML(htmlString)
	log.Println("This is HTML")
}

func attach(message *sendgrid.SGMail, inputString multistring) {
	for _, elem := range inputString {
		file, err := os.Open(elem)
		if err != nil {
			log.Println(err)
		}
		log.Println("Attaching ", filepath.Base(elem))
		message.AddAttachment(filepath.Base(elem), file)

	}
}

func main() {
	from, username, password := loadEnvironment()

	var attachments multistring

	to := flag.String("to", "", "Recipients in To")
	cc := flag.String("cc", "", "Recipients in CC")
	bcc := flag.String("bcc", "", "Recipients in BCC")
	subject := flag.String("s", "", "Subject; in case of html, can be inferred from title")
	isHtml := flag.Bool("html", false, "Send as HTML")
	cleanHtml := flag.Bool("clean", false, "Clean HTML")
	isText := flag.Bool("text", false, "Send as Text")
	flag.Var(&attachments, "a", "Attach file")
	help := flag.Bool("h", false, "Display this help message")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *isHtml && *isText {
		log.Println("Can't be both html and text")
		flag.Usage()
		os.Exit(4)
	}

	if *to == "" && *cc == "" && *bcc == "" {
		log.Println("Need atleast one of to,cc,bcc")
		flag.Usage()
		os.Exit(2)
	}

	sg := sendgrid.NewSendGridClient(username, password)
	message := sendgrid.NewMail()

	message.SetFrom(from)
	message.AddRecipients(parseRecipients(to))
	message.AddCcRecipients(parseRecipients(cc))
	message.AddBccRecipients(parseRecipients(bcc))

	b, err := ioutil.ReadAll(os.Stdin)

	if err != nil {
		log.Panic("Could not read stdin")
	}

	inputString := string(b)

	if !*isHtml && !*isText {
		log.Println("html or text not specified. Will determine that...")
		couldBeHtml := inputString[0:1] == "<"
		log.Println("isHtml is: ", couldBeHtml)
		isHtml = &couldBeHtml
	}

	if *isHtml && (*subject == "") {
		subject = extractTitle(&inputString)
	}

	log.Println("extract subject")

	if *subject == "" {
		log.Println("Subject cannot be empty")
		flag.Usage()
		os.Exit(3)
	}
	message.SetSubject(*subject)

	if *isHtml {
		if (*cleanHtml) {
			setHtmlBody(message, inputString)
		} else {
			message.SetHTML(inputString)
		}
	}

	if *isText {
		setTextBody(message, inputString)
	}

	if len(attachments) > 0 {
		attach(message, attachments)
	}

	if r := sg.Send(message); r == nil {
		fmt.Println("Email sent!")
	} else {
		fmt.Println(r)
	}
}
