package main

import (
	"fmt"
	"strconv"
	"strings"

	// "github.com/Chouette2100/srdblib/v3"
)

type EventAndUser struct {
	Eventid  string
	Ieventid int
	Userno   int
}

func SelectEventsThatDontHaveEventID() (eu []EventAndUser, err error) {
	// sqlst := " SELECT we.ieventid, weu.userno FROM weventuser weu "
	// sqlst += "JOIN wevent we ON we.eventid = weu.eventid WHERE weu.eventid like '@@@@%' ORDER BY we.ieventid "
	sqlst := " SELECT eventid,userno FROM weventuser WHERE eventid LIKE '@@@@%' ORDER BY eventid "

	// Select all events that don't have an eventid
	// The eventid is a 4-character string that starts with '@@@@'
	var itfc []interface{}
	itfc, err = Dbmap.Select(&EventAndUser{}, sqlst)
	if err != nil {
		err = fmt.Errorf("SelectEventsThatDontHaveEventID() failed: %v", err)
		return
	}

	// Extract the eventid from the wevent slice
	eu = make([]EventAndUser, len(itfc))
	for i, v := range itfc {
		eu[i] = EventAndUser{
			Eventid: v.(*EventAndUser).Eventid,
			Ieventid: func(str string) int {
				var r int
				r, err = strconv.Atoi(strings.Replace(str, "@@@@", "", 1))
				return r
			}(v.(*EventAndUser).Eventid),
			Userno: v.(*EventAndUser).Userno,
		}
	}

	return
}
