package main

import (
	"fmt"
	"strconv"

	//	"html/template"
	"net"
	"net/http"

	//	"net/http/httputil"
	"database/sql"
	"html/template"
	"time"

	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
)

func nfcHandler(w http.ResponseWriter, r *http.Request) {

	pollqry := `
		select Upper(cognome) as Chi 
		from risorse 
		where nfctag is not null and lastread > current_timestamp - interval '10 sec' 
		order by lastread desc 
		limit 1
	`

	var chi string

	db, err := sql.Open(dbType, connection)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	fmt.Println("nfcHandler - url: ", r.URL.Path)

	rows, err := db.Query(pollqry)
	if chkErrWeb(err, w) {
		return
	}

	for rows.Next() {
		err = rows.Scan(&chi)
		if chkErrWeb(err, w) {
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Println("nfcHandler - chi: ", chi)

	fmt.Fprint(w, chi)

}

func diagHandler(w http.ResponseWriter, r *http.Request) {

	var dup_ops []Dup_op
	var dup_op Dup_op
	qry := "select id, chi, quando_g, quando_o, operazione, prev_id, remote, auto from view_chk_dup_op"

	db, err := sql.Open(dbType, connection)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	fmt.Println("diagHandler - url: ", r.URL.Path)

	rows, err := db.Query(qry)
	if chkErrWeb(err, w) {
		return
	}

	for rows.Next() {
		err = rows.Scan(&dup_op.Id, &dup_op.Chi, &dup_op.Quando_g, &dup_op.Quando_o, &dup_op.Operazione, &dup_op.Prev_id, &dup_op.Remote, &dup_op.Auto)
		if chkErrWeb(err, w) {
			return
		}
		dup_ops = append(dup_ops, dup_op)
	}

	t := template.Must(template.ParseFiles("./tmpl/diag.html"))

	err = t.Execute(w, dup_ops)

	_ = chkErrWeb(err, w)

}

func homeHandler(w http.ResponseWriter, r *http.Request) {

	var risorse []Risorsa
	var risorsa Risorsa

	if !checksession(w, r) {
		http.ServeFile(w, r, "tmpl/login.html")
	}

	fmt.Println("homeHandler")
	qry := `select id, cognome from risorse order by 2`

	db, err := sql.Open(dbType, connection)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	rows, err := db.Query(qry)
	if chkErrWeb(err, w) {
		return
	}

	for rows.Next() {
		err = rows.Scan(&risorsa.Id, &risorsa.Nome)
		//fmt.Println(risorsa)
		if chkErrWeb(err, w) {
			return
		}
		risorse = append(risorse, risorsa)
	}

	t := template.Must(template.ParseFiles("./tmpl/home.html"))

	err = t.Execute(w, risorse)

	_ = chkErrWeb(err, w)

}

func checkuserHandler(w http.ResponseWriter, r *http.Request) {

	qry_locked_tpl := `select check_locked('%s', '%s', 30, 3)`
	var locked bool

	db, err := sql.Open(dbType, connection)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	fmt.Println("checkuserHandler - url: ", r.URL.Path)

	session, err := store.Get(r, "gohrsession")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("(checkuserHandler.1) session: ", session.Values)

	//get client IP
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Print("checkuserHandler userIP: [", r.RemoteAddr, "] is not IP:port")
	}
	userIP := net.ParseIP(ip)
	if userIP == nil {
		fmt.Print("checkuserHandler userIP: [", r.RemoteAddr, "] is not IP:port")
		return
	}

	fmt.Println("checkuserHandler userIP: [", userIP, "]")

	r.ParseForm()
	username := r.FormValue("usrname")
	password := r.FormValue("psw")

	fmt.Printf("checkuserHandler tentativo di login da user: %s - userIP: %v\n", username, userIP)

	qry_locked := fmt.Sprintf(qry_locked_tpl, userIP, username)
	fmt.Println(qry_locked)
	err = db.QueryRow(qry_locked).Scan(&locked)
	if chkErrWeb(err, w) {
		return
	}

	if locked {
		fmt.Println("checkuserHandler access locked")
		http.ServeFile(w, r, "tmpl/401.html")
		return
	}

	if password != "pll0312.2019" && password != "admin.quickpresence" {
		fmt.Println("checkuserHandler Password non valida")
		db.QueryRow(fmt.Sprintf("select inc_lock('%s','%s')", userIP, username))
		http.ServeFile(w, r, "tmpl/login.html")
		return
	}

	db.QueryRow(fmt.Sprintf("select reset_lock('%s','%s')", userIP, username))

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   0,
		HttpOnly: false,
	}

	session.Values["authenticated"] = true
	session.Values["username"] = username
	session.Values["lastop"] = username + "-login"
	session.Values["last"] = time.Now().Format("2006-01-02 15:04")
	session.Values["last_secs"] = uint32(time.Now().Unix())
	session.Save(r, w)

	http.Redirect(w, r, "/home", 301)

}

func logoutHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("logoutHandler - url: ", r.URL.Path)

	session, _ := store.Get(r, "gohrsession")
	session.Values["authenticated"] = false
	session.Save(r, w)
	http.ServeFile(w, r, "tmpl/login.html")

}

func remtokenHandler(w http.ResponseWriter, r *http.Request) {

	var qry, name string
	chkname_tpl := `select * from  getname('%s')`
	instime_tpl := `select * from  settime_ext(upper('%s'), '%s', '%s')` // (name, op, remote)

	db, err := sql.Open(dbType, connection)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	fmt.Println("remtokenHandler() ", r.URL.Path)

	r.ParseForm()
	token := r.FormValue("token")
	op := r.FormValue("op")
	chi := r.FormValue("chi")

	qry = fmt.Sprintf(chkname_tpl, chi)
	fmt.Println("##Remote operation: ", qry)
	err = db.QueryRow(qry).Scan(&name)
	if chkErrWeb(err, w) {
		return
	}

	// verifica token - if ok go to rilevazione presenze
	if chkToken(token) == "OK" {
		session, _ := store.Get(r, "gohrsession")
		session.Values["username"] = name
		session.Save(r, w)

		if op == "IN" || op == "OUT" {
			qry = fmt.Sprintf(instime_tpl, name, op, "S")
			fmt.Println("##Remote operation: ", qry)
			db.QueryRow(qry)
			session, _ := store.Get(r, "gohrsession")
			session.Values["lastop"] = name + "-" + op
			session.Save(r, w)
			http.ServeFile(w, r, "tmpl/oktoken.html")
		} else {
			http.ServeFile(w, r, "tmpl/get_time.html")
		}

	} else {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "Attenzione: token non riconosciuto <script>setTimeout(function(){ location='/';}, 3000);</script>")
	}

}

func remchkHandler(w http.ResponseWriter, r *http.Request) {

	qry := `select coalesce(upper(cognome),'') from  risorse where email = '%s'`

	var chi string

	db, err := sql.Open(dbType, connection)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	fmt.Println("remchkHandler ", r.URL.Path)

	r.ParseForm()
	email := r.FormValue("email")

	qry2exec := fmt.Sprintf(qry, email)

	fmt.Println("remchkHandler executing:", qry2exec)

	err = db.QueryRow(qry2exec).Scan(&chi)
	if chkErrWeb(err, w) {
		return
	}

	if chi == "" {
		fmt.Fprintf(w, "email non riconosciuta")
	} else { //genera token e invia per email
		var ris Risorsa
		sendMail(email, getToken())

		ris.Nome = chi

		t := template.Must(template.ParseFiles("./tmpl/remtoken.html"))
		err := t.Execute(w, ris)
		_ = chkErrWeb(err, w)

	}

}

func get_userIp(rawIp string) string {

	//get client IP
	ip, _, err := net.SplitHostPort(rawIp)
	if err != nil {
		fmt.Println("get_userIp error: [", rawIp, "] is not IP:port")
	}
	userIp := net.ParseIP(ip)
	if userIp == nil {
		fmt.Println("get_userIp error: [", rawIp, "] is not IP:port")
		return ""
	}
	//str_userIp := fmt.Sprint(userIp)

	fmt.Println("get_userIp IP: ", ip)
	return ip
}

func remoteHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "tmpl/remchk.html")
}


func _remchkHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "tmpl/_remchk.html")
}



func rootHandler(w http.ResponseWriter, r *http.Request) {

	avviso_html := `
	<!DOCTYPE html>
<html lang="it">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Avviso Timbrature Remote</title>
    <style>
        :root {
            --primary-color: #3498db;
            --card-bg: #2c3e50;
            --body-bg: #1a1a1a;
            --light-text: #ffffff;
            --warning-bg: #f1c40f;
            --warning-text: #2c3e50;
        }

        body {
            background-color: var(--body-bg);
            color: var(--light-text);
            margin: 0;
            padding: 0;
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            font-family: system-ui, -apple-system, sans-serif;
        }

        .warning-container {
            background: var(--warning-bg);
            border-radius: 15px;
            padding: 2rem;
            margin: 2rem;
            box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
            max-width: 800px;
            text-align: center;
            animation: pulse 2s infinite;
        }

        .warning-title {
            color: var(--warning-text);
            font-size: 1.8rem;
            margin: 1rem 0;
            line-height: 1.4;
        }

        .warning-icon {
            font-size: 3rem;
            margin-bottom: 1rem;
            color: var(--warning-text);
        }

        @keyframes pulse {
            0% { transform: scale(1); }
            50% { transform: scale(1.02); }
            100% { transform: scale(1); }
        }
    </style>
</head>
<body>
    <div class="warning-container">
        <div class="warning-icon">⚠️</div>
        <h2 class="warning-title">
            Attenzione: Questo è l'entry point per le timbrature remote
            <br><br>
            Se sei in ufficio dovresti usare l'apposito dispositivo situato all'ingresso
        </h2>
    </div>
</body>
</html>`

	fmt.Println("rootHandler", r.URL.Path)

	userIp := get_userIp(r.RemoteAddr)

	fmt.Println("rootHandler ip: ", userIp)

	if fmt.Sprint(userIp) == IpWANUS || fmt.Sprint(userIp)[0:len(IpLANUS)-1] == IpLANUS {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, avviso_html)

	} else {
		http.ServeFile(w, r, "tmpl/_remote.html")
	}
}

func showloginHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("showloginHandler", r.URL.Path)

	http.ServeFile(w, r, "tmpl/login.html")
}

func adminHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("adminHandler", r.URL.Path)

	session, err := store.Get(r, "gohrsession")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["authenticated"] = false
	session.Save(r, w)
	fmt.Println("startHandler session: ", session.Values)

	http.ServeFile(w, r, "tmpl/_login.html")
}

func setcons_meseHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	mese := r.FormValue("mese")

	currmonth_int, _ := strconv.Atoi(currmonth())
	mese_int, _ := strconv.Atoi(mese)

	fmt.Printf("Consuntivi aggiornamento  mese corrente:%d, mese richiesto:%d\n", currmonth_int, mese_int)

	if currmonth_int < mese_int {
		fmt.Fprintf(w, "Consuntivi non aggiornabili per  mese futuro (%s)", mese)
	} else {
		for i := 31; i > 0; i-- {

			quando_g := fmt.Sprintf("%s-%02s-%02d", curryear(), mese, i)
			//		fmt.Printf("upd_day() %s \n\n", quando_g)
			upd_day(quando_g)
		}
		fmt.Fprintln(w, "Consuntivi aggiornati per mese", mese)
	}

}

func setconsHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("setconsHandler", r.URL.Path)

	/*	dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}

		fmt.Printf("%q", dump)
	*/
	r.ParseForm()
	chi := r.FormValue("chi")
	quando_g := r.FormValue("quando_g")
	if len(quando_g) > 0 {
		if len(chi) > 0 {
			upd_day(quando_g, chi)
			fmt.Fprintf(w, "Aggiornamento consuntivi per data = %s e risorsa = %s \n", quando_g, chi)
		} else {
			upd_day(quando_g)
			fmt.Fprintf(w, "Aggiornamento consuntivi per data = %s \n", quando_g)
		}
	} else {
		upd_day()
		fmt.Fprintln(w, "Aggiornamento consuntivi per tutti i giorni e tutte le risorse")
	}
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Fprintln(w, r.Form)
}
