package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Chouette2100/srapi/v2"
	"github.com/Chouette2100/srdblib/v2"
)

// SetEventIDofEvents sets the eventid of events that don't have an eventid
func SetEventIDofEvents(eu []EventAndUser) (err error) {

	var executed = make(map[int]bool)

	// for i, v := range eu {
	for _, v := range eu {

		// Get the eventid of the event
		client := &http.Client{}
		var pranking *srapi.Contribution_ranking
		pranking, err = srapi.ApiEventContribution_ranking(client, v.Ieventid, v.Userno)
		if err != nil {
			err = fmt.Errorf("ApiEventContribution_ranking() failed: %v", err)
			return
		}
		time.Sleep(100 * time.Millisecond)
		urla := strings.Split(pranking.Event.EventURL, "/")
		eventid := urla[len(urla)-1]

		// Set the eventid of the event

		// Set the eventid of the event
		var result sql.Result
		sqlst := "UPDATE wevent SET eventid = ? WHERE ieventid = ?"
		result, err = srdblib.Dbmap.Exec(sqlst, eventid, v.Ieventid)
		if err != nil {
			err = fmt.Errorf("Update() failed: %v", err)
			return
		}
		var affected int64
		affected, err = result.RowsAffected()
		if err != nil {
			err = fmt.Errorf("RowsAffected() failed: %v", err)
			return
		}
		if affected != 1 {
			ok := executed[v.Ieventid]
			if affected != 0 {
				err = fmt.Errorf("Update() failed: affected=%d", affected)
				return
			} else  if ok {
				log.Printf("SetEventIDofEvents() ieventid=%6d eventid=%s already set(C) ( userno=%d )\n", v.Ieventid, eventid, v.Userno)
			} else {
				log.Printf("SetEventIDofEvents() ieventid=%6d eventid=%s already set(P) ( userno=%d )\n", v.Ieventid, eventid, v.Userno)
			}
		}
		executed[v.Ieventid] = true

		// Set the eventid of the eventuser
		sqlst = "UPDATE weventuser SET eventid = ? WHERE eventid = ? AND userno = ?"
		result, err = srdblib.Dbmap.Exec(sqlst, eventid, fmt.Sprintf("@@@@%06d", v.Ieventid), v.Userno)
		if err != nil {
			err = fmt.Errorf("Update() failed: %v", err)
			return
		}
		affected, err = result.RowsAffected()
		if err != nil {
			err = fmt.Errorf("RowsAffected() failed: %v", err)
			return
		}
		if affected != 1 {
			err = fmt.Errorf("Update() failed: affected=%d", affected)
			return
		}

		log.Printf("SetEventIDofEvents() userno=%6d ieventid=%6d eventid=%s\n", v.Userno, v.Ieventid, eventid)

		// debug
		// if i == 0 {
		// 	break
		// }
	}
	return

}
