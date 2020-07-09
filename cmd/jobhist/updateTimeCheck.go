package main

import (
	"fmt"
	"log"
	"time"
)

func getMostRecentRowTime(clusterDB string) time.Time {
	con, err := getDBConn()
	defer con.Close()
	if err != nil {
		log.Fatal(err)
	}

	resultRows, err := con.Query(fmt.Sprintf("SELECT MAX(`submission_time`) AS `max_sub_time`, MAX(`end_time`) AS `max_end_time` FROM (SELECT * FROM `%s`.`accounting` ORDER BY id DESC LIMIT 1000) AS t", clusterDB))
	if err != nil {
		log.Fatal(err)
	}

	var maxSubTime int
	var maxEndTime int

	// Normally we'd iterate over rows but this will only ever return 1.
	resultRows.Next()
	err = resultRows.Scan(&maxSubTime, &maxEndTime)
	if err != nil {
		log.Fatal(err)
	}

	var maxTimestamp int
	if maxEndTime > maxSubTime {
		maxTimestamp = maxEndTime
	} else {
		maxTimestamp = maxSubTime
	}

	// The 0 is for nanoseconds, we don't record those.
	maxTime := time.Unix(int64(maxTimestamp), 0)

	return maxTime
}

func getDurationSinceMostRecentRow(clusterDB string) time.Duration {
	t := getMostRecentRowTime(clusterDB)
	d := time.Since(t)
	return d
}

func warnAboutDBTime(clusterDB string) {
	// NB: Hours() returns a float
	d := getDurationSinceMostRecentRow(clusterDB)
	if d.Hours() > 1 {
		log.Printf("Warning: most recent entry in database is over %.0f hours old. Job data updates may have been paused.\n", d.Hours())
	}
}
