package main

import (
	"database/sql"
	"fmt"
	"github.com/UCL-RITS/go-clustertools/internal/clusters"
	_ "github.com/go-sql-driver/mysql"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"os"
	"strings"
	"unicode/utf8"
)

var dbConnString = "ccspapp:U4Ah+fSt@tcp(db.rc.ucl.ac.uk:3306)/"

func getDBConn() (*sql.DB, error) {
	// Might need allowNativePasswords=True in future - need to look into it more
	//con, err := sql.Open("mysql", "ccspapp:U4Ah+fSt@tcp(mysql.rc.ucl.ac.uk:3306)/?allowNativePasswords=True")
	return sql.Open("mysql", dbConnString)
}

func getJobData(query string) []*accountingRow {
	con, err := getDBConn()
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
		if *searchBackHours > -1 {
			fmt.Printf("No entries found. (Last %d hours searched.)\n", *searchBackHours)
		} else {
			fmt.Printf("No entries found.\n")
		}
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	if *hideHeader == false {
		table.SetHeader(elements)
	}
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

var (
	debug           = kingpin.Flag("debug", "Enable debug mode.").Bool()
	hideHeader      = kingpin.Flag("no-header", "Don't print the column headings.").Short('q').Default("false").Bool()
	searchBackHours = kingpin.Flag("hours", "Number of hours back in time to search. (Default: 48)").Short('h').PlaceHolder("<hours>").Default("-1").Int()
	searchLast      = kingpin.Flag("last", "Search for the user's <num> previous jobs. (Removes time limit.) (Default: no limit)").PlaceHolder("<num>").Default("-1").Int()
	searchNoLimits  = kingpin.Flag("all", "Do not limit results by time or number.").Short('a').Bool()
	searchUser      = kingpin.Flag("user", "User to search for jobs from. ('*' -> any) (Default: yourself)").Short('u').PlaceHolder("<username>").Default("").String()
	searchJob       = kingpin.Flag("job", "Single specific job number to search for.").Short('j').PlaceHolder("<job number>").Default("-1").Int()
	searchMHost     = kingpin.Flag("host", "Search for jobs that used a given node as the master.").Short('n').PlaceHolder("<hostname>").Default("(none)").String()
	searchCluster   = kingpin.Flag("cluster", "Search jobs run in a given cluster (myriad|legion|grace|thomas|michael|kathleen) (Default: this cluster)").Short('c').PlaceHolder("<cluster>").Default("auto").String()
	showInfoEls     = kingpin.Flag("list-elements", "Show list of elements that can be displayed.").Short('l').Bool()
	infoEls         = kingpin.Flag("info", "Show selected info (CSV list).").Short('i').Default("fstime,fetime,hostname,owner,job_number,task_number,exit_status,job_name").String()
	// TODO: implement timeout
	//timeoutSeconds  = kingpin.Flag("timeout", "Seconds to wait for database response.").Short('t').Default("3").Int()
	commitLabel string
	buildDate   string
)

func main() {

	kingpin.Version(fmt.Sprintf("jobhist commit %s built on %s", commitLabel, buildDate))
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
	if *searchCluster == "auto" {
		var err error
		*searchCluster, err = clusters.GetLocalClusterName()
		if err != nil {
			log.Fatal(err)
		}
	}
	searchDB, err := clusters.GetClusterAccountingDBName(*searchCluster)
	if err != nil {
		log.Fatalf("Error: %s.", err)
	}

	warnAboutDBTime(searchDB)

	queryFrom := searchDB

	// Next the SELECT:
	// There's a bunch of extra derived fields we want to add here.

	querySelect := "*, " +
		"DATE_FORMAT(FROM_UNIXTIME(submission_time), \"%Y-%m-%d %T\") AS fsubtime," +
		"DATE_FORMAT(FROM_UNIXTIME(start_time), \"%Y-%m-%d %T\") AS fstime, " +
		"DATE_FORMAT(FROM_UNIXTIME(end_time), \"%Y-%m-%d %T\") AS fetime, " +
		"end_time - start_time AS ewalltime, " +
		"CAST(start_time AS SIGNED INTEGER) - CAST(submission_time AS SIGNED INTEGER) as waittime, " +
		"(ru_utime + ru_stime) / (GREATEST(slots,1) * (0.9+CAST(end_time AS SIGNED INTEGER) - CAST(start_time AS SIGNED INTEGER))) AS eff "
		// avoid div/0 errors by adding 0.9 -- works out that jobs taking less than a second take 0.9 seconds
		// also avoid div/0 by using greatest(slots,1): if the shepherd fails, the job has slots = 0

	// This element is expensive to retrieve, so we want to avoid calculating it if we don't need it
	if stringInSlice("req_time", displayInfoEls) {
		querySelect += ", substr(`accounting`.`category`,(locate('h_rt=',`accounting`.`category`) + 5),(locate(',',substr(`accounting`.`category`,(locate('h_rt=',`accounting`.`category`) + 5))) - 1)) AS `req_time`"
	} else {
		querySelect += ", 0 as `req_time`"
	}

	// Finally the WHERE:
	var conditions []string

	// Searching for a specific job is fast enough and specific enough that we should
	//  ignore the time bounds unless explicitly specified
	// We also disable the default time limit if a specific number of jobs is searched for
	if (*searchJob < 0) && (*searchBackHours == -1) && (*searchLast < 0) {
		*searchBackHours = 48
	}
	if (*searchBackHours > -1) && (!*searchNoLimits) {
		time_condition := " (" +
			"        (end_time > (UNIX_TIMESTAMP(SUBDATE(NOW(), INTERVAL %d HOUR)))) OR " +
			"      (start_time > (UNIX_TIMESTAMP(SUBDATE(NOW(), INTERVAL %d HOUR)))) OR " +
			" (submission_time > (UNIX_TIMESTAMP(SUBDATE(NOW(), INTERVAL %d HOUR))))" +
			") "
		time_condition_composed := fmt.Sprintf(time_condition,
			(uint64)(*searchBackHours),
			(uint64)(*searchBackHours),
			(uint64)(*searchBackHours))
		conditions = append(conditions, time_condition_composed)
	}

	// If no explicit user-to-search-for has been specified, and we're searching for a specific job ID,
	//  assume any user is fine.
	// (Otherwise we default to searching for the current user, below.)
	if (*searchJob > 0) && (*searchUser == "") {
		*searchUser = "*"
	}

	if *searchUser != "*" {
		// Default to current user
		if *searchUser == "" {
			*searchUser = os.Getenv("USER")
		}

		// Check for username validity
		if (utf8.RuneCountInString(*searchUser) != 7) ||
			(len(strings.Map(dropUnsafeChars, *searchUser)) < len(*searchUser)) {
			log.Fatal("Error: Invalid username.")
		}
		conditions = append(conditions, fmt.Sprintf("owner = \"%s\" ", *searchUser))
	}

	if *searchJob > 0 {
		conditions = append(conditions, fmt.Sprintf("job_number = %d ", *searchJob))
	}

	if *searchMHost != "(none)" {
		if len(strings.Map(dropUnsafeChars, *searchMHost)) < len(*searchMHost) {
			log.Fatal("Error: Invalid hostname.")
		}
		conditions = append(conditions, fmt.Sprintf("hostname = \"%s\" ", *searchMHost))
	}

	// We don't need a where clause if there are no conditions
	queryWhere := ""
	if len(conditions) > 0 {
		queryWhere = " WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s.accounting %s ORDER BY end_time", querySelect, queryFrom, queryWhere)
	if (*searchLast >= 0) && (!*searchNoLimits) {
		// We need to flip the order to get only the last rows by end_time,
		//   but then we want the order to be flipped *back* for display
		query = fmt.Sprintf("SELECT * FROM (%s DESC LIMIT %d) AS t1 ORDER BY end_time", query, *searchLast)
	}

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
