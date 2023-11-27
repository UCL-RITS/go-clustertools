package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf/agent"
	"github.com/headzoo/surf/browser"
	"gopkg.in/headzoo/surf.v1"
)

var addUserFlag string
var showUserFlag string

type VaspListConnector struct {
	VaspCreds
	Browser  *browser.Browser
	UserList *[]VaspUser
}

type VaspCreds struct {
	Username   string
	Password   string
	LicenceNum string
}

func (vc *VaspCreds) MustHaveCredentials() error {
	unsetCreds := []string{}
	if vc.Username == "" {
		unsetCreds = append(unsetCreds, "username")
	}
	if vc.Password == "" {
		unsetCreds = append(unsetCreds, "password")
	}
	if vc.LicenceNum == "" {
		unsetCreds = append(unsetCreds, "licence number")
	}
	if len(unsetCreds) == 0 {
		return nil
	}
	s := strings.Join(unsetCreds, " and ")
	return fmt.Errorf("%s are unset", s)
}

var VC VaspCreds

func init() {
	flag.StringVar(&addUserFlag, "add", "", "attempt to add a user (by email address) to the list")
	flag.StringVar(&showUserFlag, "show", "", "show a single user (by email address)")
	flag.Parse()

	VC.Username = os.Getenv("VASPTOOL_USERNAME")
	VC.Password = os.Getenv("VASPTOOL_PASSWORD")
	VC.LicenceNum = os.Getenv("VASPTOOL_LICNUM")
}

func main() {
	b, err := getAuthedBrowser()
	if err != nil {
		log.Fatalln("could not get authed browser: ", err)
	}

	if addUserFlag != "" {
		err := addVaspUserToList(b, addUserFlag)
		if err != nil {
			log.Fatalln("could not add vasp user to list: ", err)
		}
		vul, err := getVaspUserList(b)
		if err != nil {
			log.Fatalln("could not get user list: ", err)
		}
		vu := getSingleVUfromVUL(vul, addUserFlag)
		if vu == nil {
			log.Fatalln("added user but then they did not appear in the list")
		}
		printSingleVU(vu)
		return
	}
	if showUserFlag != "" {
		vul, err := getVaspUserList(b)
		if err != nil {
			log.Fatalln("could not get user list: ", err)
		}
		vu := getSingleVUfromVUL(vul, showUserFlag)
		if vu == nil {
			log.Fatalln("no match for that address")
			return
		}
		printSingleVU(vu)
		return
	}

	vul, err := getVaspUserList(b)
	if err != nil {
		log.Fatalln("could not get user list: ", err)
	}
	tabulateVUs(vul)
}

func addVaspUserToList(bow *browser.Browser, userEmail string) error {
	err := bow.Open("https://vasp.at/vasp-portal/manage_users/")
	if err != nil {
		return err
	}
	manageUsersResponse := new(bytes.Buffer)
	bow.Download(manageUsersResponse)
	// ^-- this stashes the data for debugging, you don't actually need to call this otherwise

	// This is a JQuery-style selector
	fm, err := bow.Form("#add-hpc-user div div form")
	fm.Input("email", userEmail)
	fm.Input("license_id", VC.LicenceNum)
	// The license_id field seems to be autofilled and hidden in the page in the browser.
	// I'm guessing that's a work of JS

	// then submit it... somehow
	err = fm.Submit()
	if err != nil {
		return fmt.Errorf("error when submitting form: %w", err)
	}
	rc := bow.StatusCode()
	if rc != 200 {
		return fmt.Errorf("error when adding user: status code %d returned", rc)
	}

	// Examples:
	// <div class="alert alert-success" role="alert">
	// User 'ccaaabc@ucl.ac.uk' added to license 'License AB01-0012 1-234'
	// <div class="alert alert-danger" role="alert">
	// User 'ccaa123@ucl.ac.uk' already member of license 'License AB01-0012 1-234'
	// <div class="alert alert-danger" role="alert">
	// No user with email 'ccaa123@ucl.ac.uk' found!
	reSuccess := regexp.MustCompile(`^\s*User '(.*)' added to license '(.*)'\s*$`)
	reAlready := regexp.MustCompile(`^\s*User '(.*)' already member of license '(.*)'\s*$`)
	reNotFound := regexp.MustCompile(`^\s*No user with email '(.*)' found!\s*$`)
	reNotAllowed := regexp.MustCompile(`^\s*Not allowed to add user '(.*)' to license '(.*)'\s*$`)

	// Get the div with class "alert"
	// Check whether it's an alert-success or an alert-danger? Nah
	alertField := bow.Find("div.alert")
	alertText := strings.Trim(alertField.Text(), " \t\n")

	// Then check whether it matches any of these patterns
	if reSuccess.MatchString(alertText) {
		log.Println(alertText)
		log.Println("user added \\o/")
		return nil
	}
	if reAlready.MatchString(alertText) {
		return fmt.Errorf("%s", alertText)
	}
	if reNotFound.MatchString(alertText) {
		return fmt.Errorf("%s", alertText)
	}
	if reNotAllowed.MatchString(alertText) {
		return fmt.Errorf("%s", alertText)
	}

	return fmt.Errorf("%s", "no match found for alert text")
}

func getAuthedBrowser() (*browser.Browser, error) {
	err := (&VC).MustHaveCredentials()
	if err != nil {
		log.Fatalln("no credentials provided: ", err)
	}

	bow := surf.NewBrowser()
	bow.SetUserAgent(agent.Chrome()) // hide our automation

	// Authenticate
	err = bow.Open("https://vasp.at/vasp-portal/weblogin/")
	if err != nil {
		log.Fatalln("could not open authentication page: ", err)
	}

	fm, err := bow.Form("form")
	fm.Input("username", VC.Username)
	fm.Input("password", VC.Password)

	err = fm.Submit()
	if err != nil {
		log.Fatalln("could not submit auth form: ", err)
	}

	rc := bow.StatusCode()
	if rc != 200 {
		return bow, fmt.Errorf("error during authorisation: status code %d returned", rc)
	}

	firstAlert := bow.Find(".alert").First()
	if len(firstAlert.Nodes) != 0 {
		alertText := strings.Trim(firstAlert.Text(), " \t\n")
		err = fmt.Errorf("%s", alertText)
		// I'm not 100% sure this is correctly mapping the flow
		// I had to correct this at some point because I think I
		//  switched from a mapping over all alerts to just picking the
		//  first alert, without completing the switch and checking it worked
		// So now it's a bit of a guess
		log.Fatalln("alert found, something went wrong: ", err)
	}

	authSubmitResponse := new(bytes.Buffer)
	bow.Download(authSubmitResponse)
	// ^-- this stashes the data for debugging, you don't actually need to call this otherwise

	return bow, nil
}

func getVaspUserList(bow *browser.Browser) (*[]*VaspUser, error) {
	// Requires a browser already authenticated

	// Get user list
	err := bow.Open("https://vasp.at/vasp-portal/manage_users/")
	if err != nil {
		return nil, err
	}
	rc := bow.StatusCode()
	if rc != 200 {
		return nil, fmt.Errorf("error when retrieving user list: status code %d returned", rc)
	}

	manageUsersResponse := new(bytes.Buffer)
	bow.Download(manageUsersResponse)
	// ^-- this stashes the data for debugging, you don't actually need to call this otherwise

	var vaspUsers []*VaspUser

	bow.Find("#license_overview").Each(func(j int, s *goquery.Selection) {
		s.Find("tr").Each(func(i int, t *goquery.Selection) {
			if vu := parseRow(t); vu != nil {
				vaspUsers = append(vaspUsers, vu)
			}
		})
	})

	return &vaspUsers, nil
}

func getSingleVUfromVUL(vus *[]*VaspUser, userEmail string) *VaspUser {
	for _, vu := range *vus {
		if vu.EmailAddress == userEmail {
			return vu
		}
	}
	return nil
}

func printSingleVU(vu *VaspUser) {
	fmt.Printf("Name: %s %s\nAddress: %s\nValid To: %s\nKind: %s\nLicence: %s\n", vu.FormerNames, vu.LatterNames, vu.EmailAddress, vu.ValidToTimeString(), vu.EntryKind, vu.LicencedForString())
}

func tabulateVUs(vus *[]*VaspUser) {
	for _, vu := range *vus {
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", vu.FormerNames, vu.LatterNames, vu.EmailAddress, vu.ValidToTimeString(), vu.EntryKind, vu.LicencedForString())
	}
}

func parseRow(rowElement *goquery.Selection) *VaspUser {
	vu := new(VaspUser)
	l, err := rowElement.Find("th").First().Html()
	if err != nil {
		log.Fatalln("could not parse user list table: ", err)
	}
	if l == "Last name" {
		// Then this is the header row
		return nil
	}
	vu.LatterNames = l
	rowElement.Find("td").Each(
		func(i int, s *goquery.Selection) {
			switch i {
			case 0:
				vu.FormerNames, err = s.Html()
			case 1:
				vu.EmailAddress, err = s.Html()
			case 2:
				vu.ValidToString, err = s.Html()
			case 3:
				vu.EntryKind, err = s.Html()
			case 4:
				// This is the delete user button, see comment in the struct
			default:
				//panic("unexpected table format scraped")
				log.Fatalln("could not parse user list table: ", err)
			}
			if err != nil {
				//panic(err)
				log.Fatalln("could not parse user list table: ", err)
			}
		})

	// Parse the valid-to time if we have one
	const timeFormat = "Jan. 2, 2006"
	loc, err := time.LoadLocation("Europe/Vienna")
	if err != nil {
		//panic(err)
		log.Fatalln("error creating timezone context: ", err)
	}
	if vu.ValidToString != "" {
		// Because of their "humanistic" date format, we have to do a little parsing anyway
		var fixedMonth string
		month, rest, found := strings.Cut(vu.ValidToString, " ")
		if found == false {
			//panic("date could not be split for checking")
			log.Fatalln("date could not be split for checking: ", vu.ValidToString)
		}
		switch month {
		case "March":
			fixedMonth = "Mar."
		case "April":
			fixedMonth = "Apr."
		case "May":
			fixedMonth = "May."
		case "June":
			fixedMonth = "Jun."
		case "July":
			fixedMonth = "Jul."
		case "Sept.":
			fixedMonth = "Sep."
		default:
			fixedMonth = month
		}

		timeString := fixedMonth + " " + rest

		validToTime, err := time.ParseInLocation(timeFormat, timeString, loc)
		vu.ValidToTime = &validToTime
		if err != nil {
			//panic(err)
			log.Fatalln("could not parse date: ", timeString)
		}
	}
	return vu
}
