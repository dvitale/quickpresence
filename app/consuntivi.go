package main

import (
	"fmt"
	"strings"
	"database/sql"
	_ "github.com/lib/pq"
)

func upd_day(params ...string) {

// usage: upd_day(quando_g, chi) - default: (yesterday, all)
	
	notes := ""
	ris := "" 
	quando_g := yesterday()
	
	//fmt.Printf("upd_day - len(params): (%d): params[0]: %s, params[1]: %s \n", len(params), params[0], params[1])

	if len(params) > 0 {
		quando_g = params[0]
	
		if quando_g > today() { 
			return 
		}
	}
	
	if len(params) > 1 {
		ris = params[1]
	}
	
	fmt.Printf("upd_day - params (ris, quando_g):(%s, %s)\n", ris, quando_g)

	var worktime Worktime
	var qry string
	
	ristpl := `select distinct chi from timetable_all where chi ilike '%%%s' and quando_g='%s' order by 1`

	qrytpl := `select chi, operazione, quando_g, quando_o, nota  from (
		select id,chi, operazione, quando_g, quando_o, nota from getdaytimes_note('%s', '%s')
		/* union
		select id, chi, operazione, quando_g, quando_o, nota from getdaytimes_fpm('%s', '%s') */
		) q order by chi, quando_g, quando_o, id	`

	instpl_notes := `select * from set_cons_notes('%s', '%s', '%s', '%s')` 
//	instpl := `select * from set_cons('%s', '%s', '%s')` 
				
	db, err := sql.Open(dbType, connection)	
	if err != nil {
		panic(err.Error()) 
	} 
	defer db.Close()
	
	risqry := fmt.Sprintf(ristpl, ris, quando_g)
	fmt.Println("upd_day risqry: ", risqry)
	
	risrows, err := db.Query(risqry)
	if err != nil {
		panic(err.Error()) 
	} 

	for risrows.Next() {

		var wrp Wrp 
		var localChi string
		err = risrows.Scan(&localChi)
		if err != nil {
			panic(err.Error()) 
		} 

		qry = fmt.Sprintf(qrytpl, localChi, quando_g, localChi, quando_g)
		fmt.Println("upd_day wtqry: ", qry)

		rows, err := db.Query(qry)
		if err != nil {
			panic(err.Error()) 
		} 
		
		notes=""

		for rows.Next() {
			err = rows.Scan(&worktime.Chi, &worktime.Operazione, &worktime.Quando_g, &worktime.Quando_o, &worktime.Nota)
			fmt.Println("upd_day worktime: ", worktime)

			if err != nil {
				panic(err.Error()) 
			} 
			if len(worktime.Nota) > 0 {
				notes += "{" + worktime.Nota + "}"
			}
			wrp.Worktimes = append(wrp.Worktimes, worktime)
		}
		
		wrp.Ris = localChi
		wrp.Quando_g = quando_g
		
		fmt.Println("upd_day wrp-in: ", wrp.Ris, wrp.Quando_g, wrp.Htot)

		retwrp := recalcWt(wrp)
		
		fmt.Println("upd_day wrp-out: ", retwrp.Ris, retwrp.Quando_g, retwrp.Htot)
		
		qry_notes := fmt.Sprintf(instpl_notes, localChi, quando_g, retwrp.Htot,  strings.Replace(notes, "'", "''", -1))
		fmt.Println("upd_day setcons: ", qry_notes)
		db.QueryRow(qry_notes)
	}
}