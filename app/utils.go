package main

import (
	"net/http"
	"net/http/httputil"
	"fmt"
	"time"
	"strings"
	)

func dumpRequest(w http.ResponseWriter, req *http.Request) {
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		fmt.Fprint(w, err.Error())
	} else {
		fmt.Fprint(w, string(requestDump))
	}
}


func chkErrWeb(err error, w http.ResponseWriter) bool {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		//panic("!!chkErrWeb" + err.Error())
		return true
	} 
	return false
}

func checksession(w http.ResponseWriter, r *http.Request) bool {
	
	
	return true  //----> avoid check sessions
	
	session, _ := store.Get(r, "gohrsession")
	
	fmt.Println("checksession session: ", session.Values)
	
//	fmt.Println(uint32(time.Now().Unix()), sessiontimeout, session.Values["last_secs"] )
	
	
	if  uint32(time.Now().Unix()) - session.Values["last_secs"].(uint32) > sessiontimeout {
		fmt.Println("checksession sessione scaduta")
		return false
	}

	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		fmt.Println("checksession utente non autenticato")
		return false
	}

		//refresh session
	session.Values["last"] = time.Now().Format("2006-01-02 15:04")
	session.Values["last_secs"] = uint32(time.Now().Unix())
	session.Save(r, w)
	
	return true

}


func isAuthenticated(w http.ResponseWriter, r *http.Request) {
	
	session, _ := store.Get(r, "gohrsession")
	fmt.Println("(isAuthenticated.1) session: ", session.Values)
	fmt.Println(uint32(time.Now().Unix()), sessiontimeout, session.Values["last_secs"] )
	
	
	if  uint32(time.Now().Unix()) - session.Values["last_secs"].(uint32) > sessiontimeout {
		http.ServeFile(w, r, "tmpl/login.html")
		return
	}

	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.ServeFile(w, r, "tmpl/login.html")
		return
	}
}

func yesterday() string {
	return time.Now().Add(-24*time.Hour).Format("2006-01-02")
}

func today() string {
	return time.Now().Format("2006-01-02")
}

func curryear() string {
	return fmt.Sprint(time.Now().Year())
}

func currmonth() string {
	return fmt.Sprint(int(time.Now().Month()))
}

func currday() string {
	return fmt.Sprint(time.Now().Day())
}

func esc (str string) string {

	return strings.Replace(str, "'", "''", -1)
}