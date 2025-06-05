package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
	"database/sql"
	_ "github.com/lib/pq"

)



func covidHandler(w http.ResponseWriter, r *http.Request) {

	templ_file:="covid.html"
	covlogqry := `
	select 
		reg.chi, 
		reg.timbratura, 
		reg."stato attuale", 
		reg."da quando", 
		reg.remote,
		coalesce(cov.rilevazione,'')
		from view_registro_presenze_new reg  
		left join log_covid cov on reg.chi=cov.chi and left(reg."da quando", 10)=cov.quando_g
		where left(reg."da quando", 10) =  TO_CHAR(NOW()::DATE, 'yyyy-mm-dd')
		order by 1`


	fmt.Println("covidHandler - url: ", r.URL.Path)

	var covlogs []Covlog
	var covlog Covlog
	
	//isAuthenticated(w, r)
	
	db, err := sql.Open(dbType, connection)	
	if err != nil {
		panic(err.Error()) 
	}
	defer db.Close()
		rows, err := db.Query(covlogqry)	
	if chkErrWeb(err, w) {
		return
	}

	for rows.Next() {
		err = rows.Scan(&covlog.Chi, &covlog.Timbratura, &covlog.Stato, &covlog.DataStato,  &covlog.Remote, &covlog.Rilevazione)
		if chkErrWeb(err, w) {
			return
		}
		covlogs = append(covlogs, covlog)
	}

	t, err := template.New(templ_file).Funcs(template.FuncMap{
    "isOld": func(s string) bool {
		today := time.Now().Format("2006-01-02")
		return !strings.Contains(s, today)
    },
  }).ParseFiles("./tmpl/" + templ_file)

	_=chkErrWeb(err, w)
	
	err=t.Execute(w, covlogs) 	
	_=chkErrWeb(err, w)
		
}


func covsaveHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("update_timeHandler - url: ", r.URL.Path)
	
	ins_tpl :=  `insert into log_covid (chi, quando_g, rilevazione ) values ('%s', TO_CHAR(NOW()::DATE, 'yyyy-mm-dd'), '%s')`

	r.ParseForm()
	rilevazione := r.FormValue("rilevazione")
	chi := r.FormValue("chi")
	
	if strings.TrimSpace(rilevazione) != "" {

		db, err := sql.Open(dbType, connection)		
		if err != nil {
			panic(err.Error()) 
		}
		defer db.Close()

		insqry := fmt.Sprintf(ins_tpl, chi, rilevazione) 	
		fmt.Println("### covsav ", insqry)				
		db.QueryRow(insqry)
	}	
	http.Redirect(w, r, "/covid", 301)
}
