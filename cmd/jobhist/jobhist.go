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
	"strconv"
	"strings"
	"unicode/utf8"
)

type accountingRow struct {
	id                 int // 'Primary table id'
	_pos               int
	_checksum          string  //  'md5_hex checksum of line in file',
	qname              string  // 'Name of the cluster queue in which the job has run.',
	hostname           string  // 'Name of the execution host.',
	ugroup             string  // 'The effective group id of the job owner when executing the job.',
	owner              string  // 'Owner of the Grid Engine job.',
	job_name           string  // 'Job name.',
	job_number         int     // 'Job identifier - job number.',
	account            string  // 'An account string as specified by the qsub(1) or qalter(1) -A option.',
	priority           int     // 'Priority value assigned to the job corresponding to the priority parameter in the queue configuration.',
	submission_time    int     // 'Submission time (GMT unix time stamp).',
	start_time         int     // 'Start time (GMT unix time stamp).',
	end_time           int     // 'End time (GMT unix time stamp).',
	failed             int     // 'Indicates the problem which occurred in case a job could not be started on the execution host.',
	exit_status        int     // 'Exit status of the job script (or Grid Engine specific status in case of certain error conditions).',
	ru_wallclock       int     // 'Difference between end_time and start_time.',
	ru_utime           float64 // 'user time used',
	ru_stime           float64 // 'system time used',
	ru_maxrss          float64 // 'maximum resident set size',
	ru_ixrss           float64 // 'integral shared memory size',
	ru_ismrss          float64
	ru_idrss           float64 // 'integral unshared data size',
	ru_isrss           float64 // 'integral unshared stack size',
	ru_minflt          float64 // 'page reclaims',
	ru_majflt          float64 // 'page faults',
	ru_nswap           float64 // 'swaps',
	ru_inblock         float64 // 'block input operations',
	ru_oublock         float64 // 'block output operations',
	ru_msgsnd          float64 // 'messages sent',
	ru_msgrcv          float64 // 'messages received',
	ru_nsignals        float64 // 'signals received',
	ru_nvcsw           float64 // 'voluntary context switches',
	ru_nivcsw          float64 // 'involuntary context switches',
	project            string  // 'The project which was assigned to the job.',
	department         string  // 'The department which was assigned to the job.',
	granted_pe         string  // 'The parallel environment which was selected for that job.',
	slots              int     // 'The number of slots which were dispatched to the job by the scheduler.',
	task_number        int     // 'Array job task index number.',
	cpu                float64 // 'The cpu time usage in seconds.',
	mem                float64 // 'The integral memory usage in Gbytes cpu seconds.',
	io                 float64 // 'The amount of data transferred in input/output operations.',
	category           string  // 'A string specifying the job category.',
	iow                float64 // 'The io wait time in seconds.',
	pe_taskid          string  // 'If this identifier is set the task was part of a parallel job and was passed to Grid Engine via the qrsh -inherit interface.',
	maxvmem            float64 // 'The maximum vmem size in bytes.',
	arid               int     // 'Advance reservation identifier. If the job used resources of an advance reservation then this field contains a positive integer identifier otherwise the value is ''0''.',
	ar_submission_time int     // 'If the job used resources of an advance reservation then this field contains the submission time (GMT unix time stamp) of the advance reservation, otherwise the value is ''0''.',
	cost               int     //,
	C__l__bonus        int     //,
	C__l__cpu          int     //,
	C__l__gpu          int     //,
	C__l__h_rss        string  //,
	C__l__h_rt         string  //,
	C__l__h_vmem       string  //,
	C__l__memory       string  //,
	C__l__penalty      float64 //,
	C__l__threads      int     //,
	// ^-- db columns, v-- statement-calculated values
	fsubtime       string  //, 'Formatted submission time',
	fstime         string  //, 'Formatted start time',
	fetime         string  //, 'Formatted end time',
	ewalltime      int     //, 'Elapsed walltime',
	waittime       int     //, 'Time between submission and starting',
	cpu_efficiency float64 //, 'Experimental efficiency calculation',
}

func accountingRowsAssign(rows *sql.Rows) []*accountingRow {
	rowArray := make([]*accountingRow, 0, 0)

	i := 0
	for rows.Next() {
		var s accountingRow
		err := rows.Scan(
			&s.id,
			&s._pos,
			&s._checksum,
			&s.qname,
			&s.hostname,
			&s.ugroup,
			&s.owner,
			&s.job_name,
			&s.job_number,
			&s.account,
			&s.priority,
			&s.submission_time,
			&s.start_time,
			&s.end_time,
			&s.failed,
			&s.exit_status,
			&s.ru_wallclock,
			&s.ru_utime,
			&s.ru_stime,
			&s.ru_maxrss,
			&s.ru_ixrss,
			&s.ru_ismrss,
			&s.ru_idrss,
			&s.ru_isrss,
			&s.ru_minflt,
			&s.ru_majflt,
			&s.ru_nswap,
			&s.ru_inblock,
			&s.ru_oublock,
			&s.ru_msgsnd,
			&s.ru_msgrcv,
			&s.ru_nsignals,
			&s.ru_nvcsw,
			&s.ru_nivcsw,
			&s.project,
			&s.department,
			&s.granted_pe,
			&s.slots,
			&s.task_number,
			&s.cpu,
			&s.mem,
			&s.io,
			&s.category,
			&s.iow,
			&s.pe_taskid,
			&s.maxvmem,
			&s.arid,
			&s.ar_submission_time,
			&s.cost,
			&s.C__l__bonus,
			&s.C__l__cpu,
			&s.C__l__gpu,
			&s.C__l__h_rss,
			&s.C__l__h_rt,
			&s.C__l__h_vmem,
			&s.C__l__memory,
			&s.C__l__penalty,
			&s.C__l__threads,
			&s.fsubtime,
			&s.fstime,
			&s.fetime,
			&s.ewalltime,
			&s.waittime,
			&s.cpu_efficiency)
		if err != nil {
			log.Println(err)
			log.Fatal("Problem line: ", rows)
		}
		rowArray = append(rowArray, &s)
		i += 1
	}
	if *debug {
		log.Printf("%d rows captured", i)
	}
	return rowArray
}

func getNamedElement(s *accountingRow, element string) string {
	switch element {
	case "id":
		return strconv.Itoa(s.id)
	case "_pos":
		return strconv.Itoa(s._pos)
	case "_checksum":
		return s._checksum
	case "qname":
		return s.qname
	case "hostname":
		return s.hostname
	case "ugroup":
		return s.ugroup
	case "owner":
		return s.owner
	case "job_name":
		return s.job_name
	case "job_number":
		return strconv.Itoa(s.job_number)
	case "account":
		return s.account
	case "priority":
		return strconv.Itoa(s.priority)
	case "submission_time":
		return strconv.Itoa(s.submission_time)
	case "start_time":
		return strconv.Itoa(s.start_time)
	case "end_time":
		return strconv.Itoa(s.end_time)
	case "failed":
		return strconv.Itoa(s.failed)
	case "exit_status":
		return strconv.Itoa(s.exit_status)
	case "ru_wallclock":
		return strconv.Itoa(s.ru_wallclock)
	case "ru_utime":
		return strconv.FormatFloat(s.ru_utime, 'G', 9, 32)
	case "ru_stime":
		return strconv.FormatFloat(s.ru_stime, 'G', 9, 32)
	case "ru_maxrss":
		return strconv.FormatFloat(s.ru_maxrss, 'G', 9, 32)
	case "ru_ixrss":
		return strconv.FormatFloat(s.ru_ixrss, 'G', 9, 32)
	case "ru_ismrss":
		return strconv.FormatFloat(s.ru_ismrss, 'G', 9, 32)
	case "ru_idrss":
		return strconv.FormatFloat(s.ru_idrss, 'G', 9, 32)
	case "ru_isrss":
		return strconv.FormatFloat(s.ru_isrss, 'G', 9, 32)
	case "ru_minflt":
		return strconv.FormatFloat(s.ru_minflt, 'G', 9, 32)
	case "ru_majflt":
		return strconv.FormatFloat(s.ru_majflt, 'G', 9, 32)
	case "ru_nswap":
		return strconv.FormatFloat(s.ru_nswap, 'G', 9, 32)
	case "ru_inblock":
		return strconv.FormatFloat(s.ru_inblock, 'G', 9, 32)
	case "ru_oublock":
		return strconv.FormatFloat(s.ru_oublock, 'G', 9, 32)
	case "ru_msgsnd":
		return strconv.FormatFloat(s.ru_msgsnd, 'G', 9, 32)
	case "ru_msgrcv":
		return strconv.FormatFloat(s.ru_msgrcv, 'G', 9, 32)
	case "ru_nsignals":
		return strconv.FormatFloat(s.ru_nsignals, 'G', 9, 32)
	case "ru_nvcsw":
		return strconv.FormatFloat(s.ru_nvcsw, 'G', 9, 32)
	case "ru_nivcsw":
		return strconv.FormatFloat(s.ru_nivcsw, 'G', 9, 32)
	case "project":
		return s.project
	case "department":
		return s.department
	case "granted_pe":
		return s.granted_pe
	case "slots":
		return strconv.Itoa(s.slots)
	case "task_number":
		return strconv.Itoa(s.task_number)
	case "cpu":
		return strconv.FormatFloat(s.cpu, 'G', 9, 32)
	case "mem":
		return strconv.FormatFloat(s.mem, 'G', 9, 32)
	case "io":
		return strconv.FormatFloat(s.io, 'G', 9, 32)
	case "category":
		return s.category
	case "iow":
		return strconv.FormatFloat(s.iow, 'G', 9, 32)
	case "pe_taskid":
		return s.pe_taskid
	case "maxvmem":
		return strconv.FormatFloat(s.maxvmem, 'G', 9, 32)
	case "arid":
		return strconv.Itoa(s.arid)
	case "ar_submission_time":
		return strconv.Itoa(s.ar_submission_time)
	case "cost":
		return strconv.Itoa(s.cost)
	case "C__l__bonus":
		return strconv.Itoa(s.C__l__bonus)
	case "C__l__cpu":
		return strconv.Itoa(s.C__l__cpu)
	case "C__l__gpu":
		return strconv.Itoa(s.C__l__gpu)
	case "C__l__h_rss":
		return s.C__l__h_rss
	case "C__l__h_rt":
		return s.C__l__h_rt
	case "C__l__h_vmem":
		return s.C__l__h_vmem
	case "C__l__memory":
		return s.C__l__memory
	case "C__l__penalty":
		return strconv.FormatFloat(s.C__l__penalty, 'G', 9, 32)
	case "C__l__threads":
		return strconv.Itoa(s.C__l__threads)
	case "fsubtime":
		return s.fsubtime
	case "fstime":
		return s.fstime
	case "fetime":
		return s.fetime
	case "ewalltime":
		return strconv.Itoa(s.ewalltime)
	case "waittime":
		return strconv.Itoa(s.waittime)
	case "cpu_efficiency":
		return strconv.FormatFloat(s.cpu_efficiency, 'f', 9, 32)
	default:
		return "(element not found)"
	}

}

func showInfoElements() {
	fmt.Println(`Possible info elements:
	qname,
	hostname,
	ugroup,
	owner,
	job_name,
	job_number,
	account,
	priority,
	submission_time,
	start_time,
	end_time,
	failed,
	exit_status,
	ru_wallclock,
	ru_utime,
	ru_stime,
	ru_maxrss,
	ru_ixrss,
	ru_ismrss,
	ru_idrss,
	ru_isrss,
	ru_minflt,
	ru_majflt,
	ru_nswap,
	ru_inblock,
	ru_oublock,
	ru_msgsnd,
	ru_msgrcv,
	ru_nsignals,
	ru_nvcsw,
	ru_nivcsw,
	project,
	department,
	granted_pe,
	slots,
	task_number,
	cpu,
	mem,
	io,
	category,
	iow,
	pe_taskid,
	maxvmem,
	arid,
	ar_submission_time,
	cost,
	C__l__bonus,
	C__l__cpu,
	C__l__gpu,
	C__l__h_rss,
	C__l__h_rt,
	C__l__h_vmem,
	C__l__memory,
	C__l__penalty,
	C__l__threads,
	fsubtime,
	fstime,
	fetime,
	ewalltime,
	waittime,
	cpu_efficiency, # Experimental!
	stdset # (shortcut for default group)`)
}

func getJobData(query string) []*accountingRow {
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
	if (r < 'a') || (r > 'z') {
		if (r < '0') || (r > '9') {
			if r != '-' {
				if *debug {
					log.Printf("unsafe character dropped: %v", r)
				}
				return -1
			}
		}
	}
	return r
}

func getLocalClusterName() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	clusterMap := map[string]string{
		"^(?:login0[12]|node-r99a-[0-9]{3})$":                     "grace",
		"^(?:login0[56789]|node-[m-qs-z][0-9]{2}[a-f]-[0-9]{3})$": "legion",
		"^(?:login0[34]|node-k98[a-t]-[0-9]{3})$":                 "thomas",
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
	searchMHost     = kingpin.Flag("host", "Search for jobs that used a given node as the master.").Short('n').PlaceHolder("<hostname>").Default("(none)").String()
	searchCluster   = kingpin.Flag("cluster", "Search jobs run in a given cluster (legion|grace|thomas)").Short('c').PlaceHolder("<cluster>").Default("auto").String()
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

	clusterDBTables := map[string]string{
		"legion": "sgelogs2",
		"grace":  "grace_sgelogs",
		"thomas": "thomas_sgelogs",
	}

	if *searchCluster == "auto" {
		*searchCluster = getLocalClusterName()
	}
	searchDB := clusterDBTables[*searchCluster]
	if searchDB == "unknown" {
		log.Fatal("Error: there is no known database for this cluster.")
	}

	query := "SELECT *, " +
		"DATE_FORMAT(FROM_UNIXTIME(submission_time), \"%Y-%m-%d %T\") AS fsubtime," +
		"DATE_FORMAT(FROM_UNIXTIME(start_time), \"%Y-%m-%d %T\") AS fstime, " +
		"DATE_FORMAT(FROM_UNIXTIME(end_time), \"%Y-%m-%d %T\") AS fetime, " +
		"end_time - start_time AS ewalltime, " +
		"CAST(start_time AS SIGNED INTEGER) - CAST(submission_time AS SIGNED INTEGER) as waittime, " +
		"(ru_utime + ru_stime) / (slots * (0.9+CAST(end_time AS SIGNED INTEGER) - CAST(start_time AS SIGNED INTEGER))) AS eff "
		// avoid div/0 errors by adding 0.9 -- works out that jobs taking less than a second take 0.9 seconds

	query = fmt.Sprintf("%s FROM %s.accounting WHERE ", query, searchDB)

	query = fmt.Sprintf("%s end_time > (UNIX_TIMESTAMP(SUBDATE(NOW(), INTERVAL %d HOUR))) ", query, (uint64)(*searchBackHours))

	if *searchUser != "*" {
		// Check for username validity
		if (utf8.RuneCountInString(*searchUser) != 7) ||
			(len(strings.Map(dropUnsafeChars, *searchUser)) < len(*searchUser)) {
			log.Fatal("Error: Invalid username.")
		}
		query = fmt.Sprintf("%s AND owner = \"%s\" ", query, *searchUser)
	}

	if *searchMHost != "(none)" {
		if len(strings.Map(dropUnsafeChars, *searchMHost)) < len(*searchMHost) {
			log.Fatal("Error: Invalid hostname.")
		}
		query = fmt.Sprintf("%s AND hostname = \"%s\" ", query, *searchMHost)
	}

	query = fmt.Sprintf("%s ORDER BY end_time ", query)

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

	// This snippet could be made more abstract, but we only want one shortcut right now.
	splitInfoEls := strings.Split(*infoEls, ",")
	var newInfoEls []string
	standardSet := []string{"fstime", "fetime", "hostname", "owner", "job_number", "task_number", "exit_status", "job_name"}

	for _, el := range splitInfoEls {
		if el != "stdset" {
			newInfoEls = append(newInfoEls, el)
		} else {
			newInfoEls = append(newInfoEls, standardSet...)
		}
	}

	printJobData(jobData, newInfoEls)
}
