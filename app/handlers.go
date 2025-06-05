package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"bytes"
	"time"
	"database/sql"
	_ "github.com/lib/pq"

)

const (
)

func set_fulldayHandler(w http.ResponseWriter, r *http.Request) {
	
	ins_tpl :=  `select * from setfullday('%s', '%s', '%s') `
	fmt.Println("set_fulldayHandler - url: ", r.URL.Path)
	db, err := sql.Open(dbType, connection)	
	if err != nil {
		panic(err.Error()) 
	} 
	defer db.Close()
	r.ParseForm()
	ris := r.FormValue("ris")
	quando_g := r.FormValue("quando_g")
	causale := r.FormValue("causale")
	qry := fmt.Sprintf(ins_tpl, ris, quando_g, causale)
	fmt.Println("set_fulldayHandler" ,qry)
	db.QueryRow(qry)
	
	http.Redirect(w, r, "/wt?quando_g=" +  quando_g + "&ris=" + ris, 301)

}

func wdprevweekHandler(w http.ResponseWriter, r *http.Request) {

	var consuntivo Cons
	var consuntivi []Cons
	
	qrytpl := `
	select  coalesce(chi,'%s') chi, days quando_g, coalesce(htot, '0') htot, case when htot is null then '!!FERIE AUTOCALC!!' else note end as note  
	from get_prevweekworkdays() wd
	left join (select  quando_g, chi, htot, coalesce(note,'') note from consuntivi where chi ilike '%s%%' )c 
	on c.quando_g=wd.days 
	order by 2 desc
	`	
	r.ParseForm()
	chi :=  r.FormValue("ris")
	to :=  r.FormValue("to")

	fmt.Printf("prevweekHandler - url: %s  - chi = %s  to=%s\n", r.URL.Path, chi, to)
	
	db, err := sql.Open(dbType, connection)	
	if err != nil {
		panic(err.Error()) 
	} 
	defer db.Close()
	qry := fmt.Sprintf(qrytpl, chi, chi )
	fmt.Println("prevweekHandler" ,qry)

	rows, err := db.Query(qry)
	if chkErrWeb(err, w) {
		return
	}
	for rows.Next() {
		err = rows.Scan(&consuntivo.Chi, &consuntivo.Quando_g, &consuntivo.Htot, &consuntivo.Note)
		if chkErrWeb(err, w) {
			return
		}
		consuntivi = append(consuntivi, consuntivo)
	}

	t := template.Must(template.ParseFiles("./tmpl/_wdprevweek.html"))
	
	var tpl bytes.Buffer
	if e := t.Execute(&tpl, consuntivi); e != nil {
		_=chkErrWeb(err, w)	
	}
	sendmail2(to, "QP2 - Consuntivo Settimanale", tpl.String())
}


func wdcalHandler(w http.ResponseWriter, r *http.Request) {

	var consuntivo Cons
	var consuntivi []Cons
	
	r.ParseForm()
	chi :=  r.FormValue("ris")
	quando_g := r.FormValue("quando_g")
	m := strings.Split(quando_g, "-")[1]
	qry := ""
	
	if len(chi) == 0 {
		qry="select '-', days ,'-','-' from get_monthworkdays('" + m + "')"
	} else {
	
		qrytpl := `
		select  coalesce(chi,'%s') chi, days quando_g,  coalesce(htot, '0') htot, case when htot is null then '!!FERIE AUTOCALC!!' else note end as note  
		from get_monthworkdays('%s') wd
		left join (select  quando_g, chi, htot, coalesce(note,'') note from consuntivi where quando_g ilike '2019-%s%%' and chi ilike '%s%%' )c 
		on c.quando_g=wd.days 
		order by 2 
		`
		qry = fmt.Sprintf(qrytpl, chi, m, m, chi )
	}
	
	fmt.Println("wdcalHandler - url: ", r.URL.Path)

	fmt.Printf("wdcalHandler - url: %s  - quando_g: %s\n\n", r.URL.Path, quando_g)
	
	db, err := sql.Open(dbType, connection)	
	if err != nil {
		panic(err.Error()) 
	} 
	defer db.Close()
	
	fmt.Println("wdcalHandler" ,qry)

	rows, err := db.Query(qry)
	if chkErrWeb(err, w) {
		return
	}
	for rows.Next() {
		err = rows.Scan(&consuntivo.Chi, &consuntivo.Quando_g, &consuntivo.Htot, &consuntivo.Note)
		if chkErrWeb(err, w) {
			return
		}
		consuntivi = append(consuntivi, consuntivo)
	}

	t := template.Must(template.ParseFiles("./tmpl/wdcal.html"))
	

	err=t.Execute(w, consuntivi)
	
	_=chkErrWeb(err, w)
		
}

func wtfm2Handler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("wtfm2Handler - url: ", r.URL.Path)

	var consuntivo Cons2
	var consuntivi []Cons2
	
	qrytpl := `select c.chi, c.quando_g, htot, note, remote from  vw_presenze_in_us c
				where c.quando_g ilike '%s' and c.chi ilike '%%%s' order by c.chi, c.quando_g`

	session, _ := store.Get(r, "gohrsession")	
	fmt.Printf("\nwtfm2Handler session:%v+\n\n", session)

//	isAuthenticated(w, r)

	r.ParseForm()
	quando_g := r.FormValue("quando_g")
		
	ris :=  r.FormValue("ris")
			
	db, err := sql.Open(dbType, connection)	
	if err != nil {
		panic(err.Error()) 
	} 
	defer db.Close()
	
	qry := fmt.Sprintf(qrytpl, quando_g, ris)
	fmt.Println("wtfm2Handler" ,qry)

	rows, err := db.Query(qry)
	if chkErrWeb(err, w) {
		return
	}
	for rows.Next() {
		err = rows.Scan(&consuntivo.Chi, &consuntivo.Quando_g, &consuntivo.Htot, &consuntivo.Note, &consuntivo.Remote)
		if chkErrWeb(err, w) {
			return
		}
		consuntivi = append(consuntivi, consuntivo)
	}

	t := template.Must(template.ParseFiles("./tmpl/worktimefm2.html"))
	
	err=t.Execute(w, consuntivi)
	
	_=chkErrWeb(err, w)
		
}

func wtfmHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("wtfmHandler - url: ", r.URL.Path)

	var consuntivo Cons
	var consuntivi []Cons 
	
	qrytpl := `select chi, quando_g, htot, coalesce(note,'')  from  consuntivi where quando_g ilike '%s' and chi ilike '%%%s' order by chi, quando_g`

	session, _ := store.Get(r, "gohrsession")	
	fmt.Printf("\nwtfmHandler session:%v+\n\n", session)

//	isAuthenticated(w, r)

	r.ParseForm()
	quando_g := r.FormValue("quando_g")
		
	ris :=  r.FormValue("ris")
			
	db, err := sql.Open(dbType, connection)	
	if err != nil {
		panic(err.Error()) 
	} 
	defer db.Close()
	
	qry := fmt.Sprintf(qrytpl, quando_g, ris)
	fmt.Println("wtfmHandler" ,qry)

	rows, err := db.Query(qry)
	if chkErrWeb(err, w) {
		return
	}
	for rows.Next() {
		err = rows.Scan(&consuntivo.Chi, &consuntivo.Quando_g, &consuntivo.Htot, &consuntivo.Note)
		if chkErrWeb(err, w) {
			return
		}
		consuntivi = append(consuntivi, consuntivo)
	}

	t := template.Must(template.ParseFiles("./tmpl/worktimefm.html"))
	

	err=t.Execute(w, consuntivi)
	
	_=chkErrWeb(err, w)
		
}


func wtHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("wtHandler - url: ", r.URL.Path)

	//var worktimes []Worktime
	var worktime Worktime
	var wrp Wrp 
	
	qrytpl := ` select chi, operazione, quando_g, quando_o, nota  from (
		select id,chi, operazione, quando_g, quando_o, nota from getdaytimes_note('%s', '%s')
		/*union
		select id, chi, operazione, quando_g, quando_o, nota from getdaytimes_fpm('%s', '%s') */ 
		) q  order by chi, quando_g, quando_o, id	`

//	isAuthenticated(w, r)

	r.ParseForm()
	quando_g := r.FormValue("quando_g")
	
	/********* controllo per evitare il calcolo su giornata corrente ********
	dt := time.Now()  
    today := fmt.Sprint(dt.Format("2006-01-02"))

	if  quando_g == today {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, "<h3 style=' text-align: center;margin: 0px 50px 0 50px;'><a href='#' onclick='window.history.back();'>Funzione non disponibile per la giornata corrente</a></h3>")
		return
	}
	*************************************************************************/
	
	ris :=  r.FormValue("ris")
			
	db, err := sql.Open(dbType, connection)	
	if err != nil {
		panic(err.Error()) 
	} 
	defer db.Close()
	
	qry := fmt.Sprintf(qrytpl, ris, quando_g, ris, quando_g)
	fmt.Println("wtHandler" ,qry)

	rows, err := db.Query(qry)
	if chkErrWeb(err, w) {
		return
	}
	for rows.Next() {
		err = rows.Scan(&worktime.Chi, &worktime.Operazione, &worktime.Quando_g, &worktime.Quando_o, &worktime.Nota)
		if chkErrWeb(err, w) {
			return
		}
		wrp.Worktimes = append(wrp.Worktimes, worktime)
	}

	t := template.Must(template.ParseFiles("./tmpl/worktime.html"))
	

	wrp.Ris = ris
	wrp.Quando_g = quando_g
	err=t.Execute(w, recalcWt(wrp)) 
	
	_=chkErrWeb(err, w)
		
}


func update_notaHandler(w http.ResponseWriter, r *http.Request) {

	var result string
	
	fmt.Println("update_notaHandler - url: ", r.URL.Path)

	// 1. modifica record corrente
	upd_tpl :=  `select * from update_nota(%s, '%s') `

	//isAuthenticated(w, r)

	db, err := sql.Open(dbType, connection)		
	if err != nil {
		panic(err.Error()) 
	}
	defer db.Close()

	r.ParseForm()
	id := r.FormValue("id")
	nota := esc(r.FormValue("nota"))
	
	updqry := fmt.Sprintf(upd_tpl, id, nota)		
	fmt.Println("update_notaHandler: ", updqry)
	
	db.QueryRow(updqry).Scan(&result)	

	if result == "TRUE" {
		redir := "/list_mytime?quando_g=" +  r.FormValue("quando_g") + "&ris=" + r.FormValue("ris")
		fmt.Println("update_notaHandler: ", redir)
		http.Redirect(w, r, redir, 301)
	} else {

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, "<h3 style=' text-align: center;margin: 0px 50px 0 50px;'><a href='#' onclick='window.history.back();'>Errore sul DB.</a></h3>") 
	}
}


func mylistHandler(w http.ResponseWriter, r *http.Request) {

	var ris Risorsa
	
	fmt.Println("mylistHandler - url: ", r.URL.Path)
		
	r.ParseForm()

	ris.Nome = r.FormValue("lastChi")
	
	t := template.Must(template.ParseFiles("./tmpl/my_input.html"))
	
	err := t.Execute(w, ris) 
	
	_=chkErrWeb(err, w)

}


func check_byqrHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("check_byqrHandler - url: ", r.URL.Path)
	
	var qry, status string

	//chkname_tpl := `select * from  getname('%s')`
	chkname_tpl := `select chi || '|' || quando_g || 'T' || quando_o || '|' || operazione from view_last_time_new where chi ilike getname('%s')`
		
	db, err := sql.Open(dbType, connection)		
	if err != nil {
		panic(err.Error()) 
	}
	defer db.Close()
	
	r.ParseForm()

	chi := r.FormValue("chi")
	qry = fmt.Sprintf(chkname_tpl, chi)	
	
	fmt.Println("check_byqrHandler() -> getname: ", qry)
	
	err = db.QueryRow(qry).Scan(&status)
	if chkErrWeb(err, w) {
		return
	}
	
	if  status != ""{
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		session, _ := store.Get(r, "gohrsession")
		session.Values["lastop"] = chi + "-chkQR"
		session.Save(r, w)
		fmt.Fprint(w, status)
		return
	}

}

func chk_remHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("chk_remHandler - url: ", r.URL.Path)	

	userIp := get_userIp(r.RemoteAddr)
	myIpByCli := r.FormValue("myip")
	isUS:= (myIpByCli == IpWANUS)
	fmt.Printf("chk_remHandler - userIp: >%s< IpWANUS: >%s<  isUS %t  myIpByCli: %s\n", userIp, IpWANUS, isUS, myIpByCli)
//	if userIp != "127.0.0.1" && userIp != IpWANUS && userIp != IpLANUS { //ip remoto: timbratura virtuale in telelavoro
	remote := "S"
	if isUS  { //ip remoto: timbratura virtuale in telelavoro
		fmt.Println("chk_remHandler - userIp is remote: ", myIpByCli)
		remote = "N"
	}
	fmt.Fprintln(w,"remote: ", remote)
}
 
func time_byqrHandler(w http.ResponseWriter, r *http.Request) {
	
	fmt.Println("time_byqrHandler - url: ", r.URL.Path)	
	remote := "?"
	myIpByCli := r.FormValue("myIpByCli")
	isUS:= (myIpByCli == IpWANUS)
	fmt.Printf("time_byqrHandler  IpWANUS: >%s<  isUS %t  myIpByCli: %s\n", IpWANUS, isUS, myIpByCli)
	
	if isUS  { //ip remoto: timbratura virtuale in telelavoro
		fmt.Println("time_byqrHandler - Ip IS NOT remote: ", myIpByCli)
		remote = "N"
	} else {
		fmt.Println("time_byqrHandler - Ip IS remote: ", myIpByCli)
		remote = "S"
	}
	
	var qry, esito, name string

	chkname_tpl := `select * from  getname('%s')`
	
	instime_tpl :=  `select * from  settime_ext(upper('%s'), '%s', '%s')` // settime_ext(name, op, remote)
	
	db, err := sql.Open(dbType, connection)		
	if err != nil {
		panic(err.Error()) 
	}
	defer db.Close()
	
	r.ParseForm()

	
	chi := r.FormValue("chi")
	operazione := r.FormValue("operazione")
	qry = fmt.Sprintf(chkname_tpl, chi)	
	
	err = db.QueryRow(qry).Scan(&name)
	if chkErrWeb(err, w) {
		return
	}
	
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if  name == "-1" {
		fmt.Fprint(w, "ERRORE rilevazione: QR code non riconducibile ad alcuna risorsa anagrafata")
		return
	}

	//lastAccess = name
	fmt.Println("time_byqrHandler() -> getname: ", qry, name)

	qry = fmt.Sprintf(instime_tpl, name, operazione, remote)			
	fmt.Println("time_byqrHandler() -> settime_ext: ", qry)
	
	err = db.QueryRow(qry).Scan(&esito)
	if chkErrWeb(err, w) {
		return
	}
	resp := "Chi: " + name + " - Operazione: " + operazione + " - Time: " + time.Now().Format("2006-01-02 15:04")
	
	if esito=="TRUE" {
		resp += "<h3 class='w3-red'> !!Attenzione!!  L'operazione richiesta è fuori sequenza. Applicata forzatura automatica.</h3>"
	}
	
	session, _ := store.Get(r, "gohrsession")
	session.Values["lastop"] = name + "-" + operazione
	session.Save(r, w)

	fmt.Fprintln(w, resp)
}

func update_timeHandler(w http.ResponseWriter, r *http.Request) {

	var result string
	
	fmt.Println("update_timeHandler - url: ", r.URL.Path)

	// 1. modifica record corrente
	upd_tpl :=  `select * from update_time(%s, '%s') `

	// 2. salvataggio dati in tab.log
	ins_tpl :=  `insert into log_mod_orari (id_timetable, oldval, newval, note, modified ) values (%s, '%s', '%s', '%s', current_timestamp)`

	isAuthenticated(w, r)

	db, err := sql.Open(dbType, connection)		
	if err != nil {
		panic(err.Error()) 
	}
	defer db.Close()

	r.ParseForm()
	fmt.Println("update_timeHandler" ,r.Form)
	id := r.FormValue("id")
	ris := r.FormValue("ris")
	quando_g := r.FormValue("quando_g")
	note := r.FormValue("note")
	quando_o := r.FormValue("quando_o")
	quando_o_old := r.FormValue("quando_o_old")
	
//1.
	updqry := fmt.Sprintf(upd_tpl, id, quando_o)		
	fmt.Println("### 1. ", updqry)
	db.QueryRow(updqry).Scan(&result)	
	fmt.Println("### 1. -> result ->", result)

	if result == "TRUE" {

//2.
		insqry := fmt.Sprintf(ins_tpl, id, quando_o_old, quando_o, note) 	
		fmt.Println("### 2. ", insqry)				
		db.QueryRow(insqry)
		
		http.Redirect(w, r, "/list_daytime?quando_g=" + quando_g + "&ris=" + ris, 301)
	} else {

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, "<h3 style=' text-align: center;margin: 0px 50px 0 50px;'><a href='#' onclick='window.history.back();'>La modifica richiesta risulta essere fuori sequenza e non può essere applicata.</a></h3>") 
	}
}

func list_daytimeHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("list_daytimeHandler - url: ", r.URL.Path)
	
	var wrp Wrp 
	
	//qrytpl := "select id, operazione, quando_g, quando_o, chi, coalesce(auto,'') from getdaytimes('%s', '%s') "
	qrytpl := "select id, operazione, quando_g, quando_o, chi, coalesce(auto,''), coalesce(nota,'') from getdaytimes_note('%s', '%s') "

	//isAuthenticated(w, r)

	r.ParseForm()

	quando_g := r.FormValue("quando_g")
	ris :=  r.FormValue("ris")
			
	db, err := sql.Open(dbType, connection)	
	if err != nil {
		panic(err.Error()) 
	}
	defer db.Close()
	
	qry := fmt.Sprintf(qrytpl, ris, quando_g)
	fmt.Println("list_daytime" ,qry)

	rows, err := db.Query(qry)
	if chkErrWeb(err, w) {
		return
	}

	for rows.Next() {
		var worktime Worktime
		err = rows.Scan(&worktime.Id, &worktime.Operazione, &worktime.Quando_g, &worktime.Quando_o, &worktime.Chi, &worktime.Auto,  &worktime.Nota)
		if chkErrWeb(err, w) {
			return
		}
		wrp.Worktimes = append(wrp.Worktimes , worktime)
	}

	t := template.Must(template.ParseFiles("./tmpl/list_daytime.html"))
	
	wrp.Ris = ris
	wrp.Quando_g = quando_g
	err=t.Execute(w, wrp) 
	
	_=chkErrWeb(err, w)
		
}

func list_mytimeHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("list_mytimeHandler - url: ", r.URL.Path)

	var wrp Wrp 
	
	qrytpl := `select chi, id, operazione, quando_g, quando_o, auto, nota from getdaytimes_note('%s', '%s') `

//	isAuthenticated(w, r)

	r.ParseForm()

	quando_g := r.FormValue("quando_g")
	ris :=  r.FormValue("ris")
			
	db, err := sql.Open(dbType, connection)	
	if err != nil {
		panic(err.Error()) 
	} 
	defer db.Close()
	
	qry := fmt.Sprintf(qrytpl, ris, quando_g)
	fmt.Println("list_mytime" ,qry)

	rows, err := db.Query(qry)
	if chkErrWeb(err, w) {
		return
	}

	for rows.Next() {
		var worktime Worktime
		err = rows.Scan(&worktime.Chi, &worktime.Id, &worktime.Operazione, &worktime.Quando_g, &worktime.Quando_o, &worktime.Auto, &worktime.Nota)
		if chkErrWeb(err, w) {
			return
		}
		wrp.Worktimes = append(wrp.Worktimes , worktime)
	}
	t := template.Must(template.ParseFiles("./tmpl/list_mydaytime.html"))
	
	wrp.Ris = ris
	wrp.Quando_g = quando_g
	err=t.Execute(w, wrp) 
	
	_=chkErrWeb(err, w)
		
}


func get_timeHandler(w http.ResponseWriter, r *http.Request) {

msg := `<br><br><div style='text-align:center;vertical-align: middle;background: yellow'><h2 style='color:blue'><br><br>Attenzione: c'è un problema col tuo entry point di timbratura <br><br>
		Se sei in ufficio dovresti usare l'apposito dispositivo situato all'ingresso<br><br>
		Se invece stai timbrando da remoto dovresti usare </h2> <h3>https://qp2.u-s.it:8443 </h3><br></div>
		`

	enabledIP:="172.16.0.32"
	enabledIP2:="93.42.222.106"
	
	fmt.Println("get_timeHandler - url: ", r.URL.Path)
//	session, _ := store.Get(r, "gohrsession")	
//	fmt.Printf("\nget_timeHandler session:%v+\n\n", session)

	userIp := get_userIp(r.RemoteAddr)
	
	fmt.Println("get_timeHandler ip: ", userIp)
	
	if fmt.Sprint(userIp) != enabledIP &&  fmt.Sprint(userIp) != enabledIP2 {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		
		fmt.Fprintf(w, msg)	
			
	}  else {

		http.ServeFile(w, r, "tmpl/get_time.html")
	}
}


func get_time_secHandler(w http.ResponseWriter, r *http.Request) {

	http.ServeFile(w, r, "tmpl/get_time.html")
}



func list_presHandler(w http.ResponseWriter, r *http.Request) {


	presenceqry := `select chi, timbratura, "stato attuale", "da quando", remote from view_registro_presenze_new  order by 3,1`


	fmt.Println("list_presHandler - url: ", r.URL.Path)

	var presences []Presence
	var presence Presence
	
	//isAuthenticated(w, r)
	
	db, err := sql.Open(dbType, connection)	
	if err != nil {
		panic(err.Error()) 
	}
	defer db.Close()
	
	rows, err := db.Query(presenceqry)	
	if chkErrWeb(err, w) {
		return
	}

	for rows.Next() {
		err = rows.Scan(&presence.Chi, &presence.Timbratura, &presence.Stato, &presence.DataStato,  &presence.Remote)
		if chkErrWeb(err, w) {
			return
		}
		presences = append(presences, presence)

	}

/*	t := template.Must(template.ParseFiles("./tmpl/list_pres.html"))

	t.Funcs(template.FuncMap{
		"isOld": func(s string) bool {
			today := time.Now().Format("2006-01-02")
			return !strings.Contains(s, today)
		}
	})
*/	
	t, err := template.New("list_pres.html").Funcs(template.FuncMap{
    "isOld": func(s string) bool {
		today := time.Now().Format("2006-01-02")
		return !strings.Contains(s, today)
    },
  }).ParseFiles("./tmpl/list_pres.html")

	_=chkErrWeb(err, w)
	
	err=t.Execute(w, presences) 	
	_=chkErrWeb(err, w)
		
}
