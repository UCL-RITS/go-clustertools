package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
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
	ru_idrss           float64       // 'integral unshared data size',
	ru_isrss           float64       // 'integral unshared stack size',
	ru_minflt          float64       // 'page reclaims',
	ru_majflt          float64       // 'page faults',
	ru_nswap           float64       // 'swaps',
	ru_inblock         float64       // 'block input operations',
	ru_oublock         float64       // 'block output operations',
	ru_msgsnd          float64       // 'messages sent',
	ru_msgrcv          float64       // 'messages received',
	ru_nsignals        float64       // 'signals received',
	ru_nvcsw           float64       // 'voluntary context switches',
	ru_nivcsw          float64       // 'involuntary context switches',
	project            string        // 'The project which was assigned to the job.',
	department         string        // 'The department which was assigned to the job.',
	granted_pe         string        // 'The parallel environment which was selected for that job.',
	slots              int           // 'The number of slots which were dispatched to the job by the scheduler.',
	task_number        int           // 'Array job task index number.',
	cpu                float64       // 'The cpu time usage in seconds.',
	mem                float64       // 'The integral memory usage in Gbytes cpu seconds.',
	io                 float64       // 'The amount of data transferred in input/output operations.',
	category           string        // 'A string specifying the job category.',
	iow                float64       // 'The io wait time in seconds.',
	pe_taskid          string        // 'If this identifier is set the task was part of a parallel job and was passed to Grid Engine via the qrsh -inherit interface.',
	maxvmem            float64       // 'The maximum vmem size in bytes.',
	arid               int           // 'Advance reservation identifier. If the job used resources of an advance reservation then this field contains a positive integer identifier otherwise the value is ''0''.',
	ar_submission_time int           // 'If the job used resources of an advance reservation then this field contains the submission time (GMT unix time stamp) of the advance reservation, otherwise the value is ''0''.',
	cost               sql.NullInt64 // This is a special type because at some point something broke and now there are NULL entries for it.
	C__l__bonus        int           //,
	C__l__cpu          int           //,
	C__l__gpu          int           //,
	C__l__h_rss        string        //,
	C__l__h_rt         string        //,
	C__l__h_vmem       string        //,
	C__l__memory       string        //,
	C__l__penalty      float64       //,
	C__l__threads      int           //,
	// ^-- db columns, v-- statement-calculated values
	fsubtime       string          //, 'Formatted submission time',
	fstime         string          //, 'Formatted start time',
	fetime         string          //, 'Formatted end time',
	slowdown       float64         //, '(Bad) slowdown metric',
	ewalltime      int             //, 'Elapsed walltime',
	waittime       int             //, 'Time between submission and starting',
	cpu_efficiency float64         //, 'Experimental efficiency calculation',
	req_time       string          //, 'Requested job time (stored) -- this is a string because someone made a mess in the past :(',
	req_time_calc  int             //, 'Requested job time (extracted from the category field)',
	req_slowdown   sql.NullFloat64 //, 'Slowdown metric using requested time (stored) instead of run time. A special type because some rows have invalid req_time data.',
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
			&s.slowdown,
			&s.ewalltime,
			&s.waittime,
			&s.cpu_efficiency,
			&s.req_time,
			&s.req_time_calc,
			&s.req_slowdown,
		)
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

// Remove DNS suffix from hostname
func unqdn(s string) string {
	if i := strings.Index(s, "."); i < 0 {
		return s
	} else {
		return s[0:i]
	}
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
		return unqdn(s.hostname)
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
		if s.cost.Valid == true {
			return strconv.FormatInt(s.cost.Int64, 10)
		} else {
			return "(null)"
		}
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
	case "slowdown":
		return strconv.FormatFloat(s.slowdown, 'f', 1, 32)
	case "ewalltime":
		return strconv.Itoa(s.ewalltime)
	case "waittime":
		return strconv.Itoa(s.waittime)
	case "req_time":
		return s.req_time
	case "req_time_calc":
		return strconv.Itoa(s.req_time_calc)
	case "req_slowdown":
		if s.req_slowdown.Valid == true {
			return strconv.FormatFloat(s.req_slowdown.Float64, 'f', 1, 32)
		} else {
			return "(null)"
		}
	case "cpu_efficiency":
		return strconv.FormatFloat(s.cpu_efficiency, 'f', 9, 32)
	default:
		return "(element not found)"
	}

}

type elementDesc struct {
	Label       string
	Description string
}

func showInfoElements() {
	elementDescriptions := []elementDesc{
		{"qname", "the name of the internal queue this job used"},
		{"hostname", "the hostname of the master node this job ran on"},
		{"ugroup", "the effective group id of the job owner"},
		{"owner", "the user who owns the job"},
		{"job_name", "the name of the job in the scheduler"},
		{"job_number", "the job ID"},
		{"account", "a string used to calculate Gold spending"},
		{"priority", "priority value assigned to the job, by the queue"},
		{"submission_time", "the time the job was submitted, in seconds since the UNIX epoch"},
		{"start_time", "the time the job started, in seconds since the UNIX epoch (0 if failed to start)"},
		{"end_time", "the time the job ended, in seconds since the UNIX epoch (0 if failed to start)"},
		{"failed", "a numeric error code indicated whether and why a job failed at the scheduler level"},
		{"exit_status", "the exit status of the job, or an additional error code from the scheduler in case of failure"},
		{"ewalltime", "elapsed time for the job"},
		// I don't trust the ru_ ones to mean anything sensible
		// {"ru_wallclock", ""},
		// {"ru_utime,", ""},
		// {"ru_stime,", ""},
		// {"ru_maxrss,", ""},
		// {"ru_ixrss,", ""},
		// {"ru_ismrss,", ""},
		// {"ru_idrss,", ""},
		// {"ru_isrss,", ""},
		// {"ru_minflt,", ""},
		// {"ru_majflt,", ""},
		// {"ru_nswap,", ""},
		// {"ru_inblock,", ""},
		// {"ru_oublock,", ""},
		// {"ru_msgsnd,", ""},
		// {"ru_msgrcv,", ""},
		// {"ru_nsignals,", ""},
		// {"ru_nvcsw,", ""},
		// {"ru_nivcsw,", ""},
		{"slots", "'slots' granted to the job by the scheduler"},
		{"cost", "number of cores blocked out by the job (virtual cores on clusters with hyperthreading)"},
		{"task_number", "the task ID, for array jobs"},
		// I don't trust the cpu, mem, or io ones either
		// {"cpu", ""},
		// {"mem", ""},
		// {"io", ""},
		// {"iow", ""},
		// {"maxvmem", ""},
		{"category", "some stuck-together info about the job"},
		// Then there's some other stuff which doesn't apply to any of our jobs
		// {"pe_taskid", "this would only be populated if we had the accounting_summary setting turned off in the scheduler. See `man accounting`."},
		// {"arid", "advanced reservation ID. We never use these."},
		// {"ar_submission_time", "advanced reservation submission time. We never use these."},
		// And then the subset of category break-outs that actually seem useful
		{"C__l__threads", "whether the job requested use of all hyperthreaded cores"},
		{"C__l__gpu", "number of GPUs requested"},
		{"C__l__memory", "RAM per core requested"},
		// And the ones that don't
		// {"C__l__bonus", "Don't know"},
		// {"C__l__cpu", "Don't know"},
		// {"C__l__h_rt", "The same as req_time"},
		// {"C__l__penalty", "Don't know"},
		// {"C__l__h_rss", "A memory resource request I don't think we use"},
		// {"C__l__h_vmem", "Ditto"},
		// And then some add-ons, calculated rather than stored (except req_time, which used to be calculated)
		{"fsubtime", "submission time, converted into readable format"},
		{"fstime", "start time, converted into readable format"},
		{"fetime", "end time, converted into readable format"},
		{"waittime", "how long the job spent waiting (0 if failed to start)"},
		{"cpu_efficiency", "experimental: number of CPU processing seconds divided by elapsed walltime"},
		{"req_time", "maximum walltime requested by job"},
		{"req_time_calc", "maximum walltime requested by job (calculated, for jobs before this was stored)"},
		{"slowdown", "wait time + run time / run_time"},
		{"req_slowdown", "slowdown, calculated from time requested rather than run time"},
		{"stdset", "a shortcut for the default set of printed fields"},
	}

	fmt.Println(`Possible info elements:`)
	for _, v := range elementDescriptions {
		fmt.Printf("  %15s     %s\n", v.Label, v.Description)
	}
}
