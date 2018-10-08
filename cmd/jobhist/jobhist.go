package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

func getJobData(query string) []*accountingRow {
	// Might need allowNativePasswords=True in future - need to look into it more
	//con, err := sql.Open("mysql", "ccspapp:U4Ah+fSt@tcp(mysql.external.legion.ucl.ac.uk:3306)/?allowNativePasswords=True")
	con, err := sql.Open("mysql", "ccspapp:U4Ah+fSt@tcp(mysql.external.legion.ucl.ac.uk:3306)/")
	defer con.Close()

	if err != nil {
		fmt.Println(err)
	}

	rows, err := con.Query(query)

	if err != nil {
		log.Fatal(err)
	}

	accountingRows := accountingRowsAssign(rows)

	return accountingRows
}

func printJobData(rows []*accountingRow, elements []string) {
	if len(rows) == 0 {
		fmt.Printf("No entries found. (Last %d hours searched.)\n", *searchBackHours)
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(elements)
	table.SetBorder(false)

	var rowBuffer []string
	rowBuffer = make([]string, len(elements))

	for _, row := range rows {
		for i, elementName := range elements {
			rowBuffer[i] = getNamedElement(row, elementName)
		}
		table.Append(rowBuffer)
	}

	table.Render()
}

func dropUnsafeChars(r rune) rune {
	if (r >= 'a') && (r <= 'z') {
		return r
	}

	if (r >= '0') && (r <= '9') {
		return r
	}

	if r == '-' {
		return r
	}

	if *debug {
		log.Printf("unsafe character dropped: %v", r)
	}
	return -1
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func getLocalClusterName() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	hostname = strings.SplitN(hostname, ".", 2)[0] // Gets only the first segment of the hostname
	clusterMap := map[string]string{
		"^(?:login1[23]|node-[hij]00a-[0-9]{3})$":                 "myriad",
		"^(?:login0[12]|node-r99a-[0-9]{3})$":                     "grace",
		"^(?:login0[56789]|node-[l-qs-z][0-9]{2}[a-f]-[0-9]{3})$": "legion",
		"^(?:login0[34]|node-k98[a-t]-[0-9]{3})$":                 "thomas",
		"^(?:login1[01]|node-k10[a-i]-0[0-3][0-9]|util11)$":       "michael",
	}
	for pattern, clusterName := range clusterMap {
		if regexp.MustCompile(pattern).MatchString(hostname) {
			return clusterName
		}
	}
	// TODO: proper error returns
	log.Fatal("automatic cluster matching could not determine which cluster this is. Please specify on the command-line or panic wildly.")
	return ""
}

var (
	debug           = kingpin.Flag("debug", "Enable debug mode.").Bool()
	searchBackHours = kingpin.Flag("hours", "Number of hours back in time to search.").Short('h').PlaceHolder("<hours>").Default("24").Int()
	searchUser      = kingpin.Flag("user", "User to search for jobs from.").Short('u').PlaceHolder("<username>").Default(os.Getenv("USER")).String()
	searchJob       = kingpin.Flag("job", "Job number to search for.").Short('j').PlaceHolder("<job number>").Default("-1").Int()
	searchMHost     = kingpin.Flag("host", "Search for jobs that used a given node as the master.").Short('n').PlaceHolder("<hostname>").Default("(none)").String()
	searchCluster   = kingpin.Flag("cluster", "Search jobs run in a given cluster (myriad|legion|grace|thomas|michael)").Short('c').PlaceHolder("<cluster>").Default("auto").String()
	showInfoEls     = kingpin.Flag("list-elements", "Show list of elements that can be displayed.").Short('l').Bool()
	infoEls         = kingpin.Flag("info", "Show selected info (CSV list).").Short('i').Default("fstime,fetime,hostname,owner,job_number,task_number,exit_status,job_name").String()
	// TODO: implement timeout
	//timeoutSeconds  = kingpin.Flag("timeout", "Seconds to wait for database response.").Short('t').Default("3").Int()
	commitLabel string
	buildDate   string
)

func main() {

	kingpin.Version(fmt.Sprintf("jobhist 0.0.1 commit %s built on %s", commitLabel, buildDate))
	kingpin.Parse()

	if *showInfoEls != false {
		showInfoElements()
		os.Exit(0)
	}

	// This snippet could be made more abstract, but we only want one shortcut right now.
	splitInfoEls := strings.Split(*infoEls, ",")
	var displayInfoEls []string
	standardSet := []string{"fstime", "fetime", "hostname", "owner", "job_number", "task_number", "exit_status", "job_name"}

	for _, el := range splitInfoEls {
		if el != "stdset" {
			displayInfoEls = append(displayInfoEls, el)
		} else {
			displayInfoEls = append(displayInfoEls, standardSet...)
		}
	}

	// Build SQL Query

	// First the FROM:
	clusterDBTables := map[string]string{
		"myriad":  "myriad_sgelogs",
		"legion":  "sgelogs2",
		"grace":   "grace_sgelogs",
		"thomas":  "thomas_sgelogs",
		"michael": "michael_sgelogs",
	}

	if *searchCluster == "auto" {
		*searchCluster = getLocalClusterName()
	}
	searchDB := clusterDBTables[*searchCluster]
	if searchDB == "unknown" {
		log.Fatal("Error: there is no known database for this cluster.")
	}

	queryFrom := searchDB

	// Next the SELECT:

	querySelect := "*, " +
		"DATE_FORMAT(FROM_UNIXTIME(submission_time), \"%Y-%m-%d %T\") AS fsubtime," +
		"DATE_FORMAT(FROM_UNIXTIME(start_time), \"%Y-%m-%d %T\") AS fstime, " +
		"DATE_FORMAT(FROM_UNIXTIME(end_time), \"%Y-%m-%d %T\") AS fetime, " +
		"end_time - start_time AS ewalltime, " +
		"CAST(start_time AS SIGNED INTEGER) - CAST(submission_time AS SIGNED INTEGER) as waittime, " +
		"(ru_utime + ru_stime) / (slots * (0.9+CAST(end_time AS SIGNED INTEGER) - CAST(start_time AS SIGNED INTEGER))) AS eff "
		// avoid div/0 errors by adding 0.9 -- works out that jobs taking less than a second take 0.9 seconds

	// This element is expensive to retrieve, so we want to avoid calculating it if we don't need it
	if stringInSlice("req_time", displayInfoEls) {
		querySelect += ", substr(`accounting`.`category`,(locate('h_rt=',`accounting`.`category`) + 5),(locate(',',substr(`accounting`.`category`,(locate('h_rt=',`accounting`.`category`) + 5))) - 1)) AS `req_time`"
	} else {
		querySelect += ", 0 as `req_time`"
	}

	// Finally the WHERE:
	time_condition := " (" +
		"        (end_time > (UNIX_TIMESTAMP(SUBDATE(NOW(), INTERVAL %d HOUR)))) OR " +
		"      (start_time > (UNIX_TIMESTAMP(SUBDATE(NOW(), INTERVAL %d HOUR)))) OR " +
		" (submission_time > (UNIX_TIMESTAMP(SUBDATE(NOW(), INTERVAL %d HOUR))))" +
		") "
	queryWhere := fmt.Sprintf(time_condition,
		(uint64)(*searchBackHours),
		(uint64)(*searchBackHours),
		(uint64)(*searchBackHours))

	if *searchUser != "*" {
		// Check for username validity
		if (utf8.RuneCountInString(*searchUser) != 7) ||
			(len(strings.Map(dropUnsafeChars, *searchUser)) < len(*searchUser)) {
			log.Fatal("Error: Invalid username.")
		}
		queryWhere = fmt.Sprintf("%s AND owner = \"%s\" ", queryWhere, *searchUser)
	}

	if *searchJob > 0 {
		queryWhere = fmt.Sprintf("%s AND job_number = %d ", queryWhere, *searchJob)
	}

	if *searchMHost != "(none)" {
		if len(strings.Map(dropUnsafeChars, *searchMHost)) < len(*searchMHost) {
			log.Fatal("Error: Invalid hostname.")
		}
		queryWhere = fmt.Sprintf("%s AND hostname = \"%s\" ", queryWhere, *searchMHost)
	}

	query := fmt.Sprintf("SELECT %s FROM %s.accounting WHERE %s ORDER BY end_time", querySelect, queryFrom, queryWhere)

	if *debug {
		log.Printf("Making query: %s", query)
	}

	// TODO: add timeout on DB connection
	// https://blog.golang.org/go-concurrency-patterns-timing-out-and
	//timeout := make(chan bool, 1)
	//go func() {
	//	time.Sleep(*timeoutSeconds * time.Second)
	//	timeout <- true
	//}()

	jobData := getJobData(query)

	printJobData(jobData, displayInfoEls)
}
