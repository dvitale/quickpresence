package main

import (
	"net/http"

	//	"github.com/peterhellberg/acr122u"
	"database/sql"
	"fmt"

	"github.com/alexandrevicenzi/go-sse"
)

var (
	err           error
	authenticated = false
)

func main() {

	mux := http.NewServeMux()

	s := &http.Server{
		Handler: mux,
	}

	mux.HandleFunc("/", rootHandler)                   //entry point telelav
	mux.HandleFunc("/remote", remoteHandler)           //entry point telelav
	mux.HandleFunc("/showlogin", showloginHandler)     //entry point telelav
	mux.HandleFunc("/remchk", remchkHandler)           //validazione utente telelav 1
	mux.HandleFunc("/remtoken", remtokenHandler)       //validazione utente telelav 2
	mux.HandleFunc("/mylist", mylistHandler)           //validazione utente telelav 2
	mux.HandleFunc("/list_mytime", list_mytimeHandler) //validazione utente telelav 2
	mux.HandleFunc("/wt", wtHandler)                   //validazione utente telelav 2
	mux.HandleFunc("/wtfm", wtfmHandler)               //validazione utente telelav 2
	mux.HandleFunc("/wtfm2", wtfm2Handler)             //validazione utente telelav 2
	mux.HandleFunc("/update_nota", update_notaHandler) //validazione utente telelav 2
	mux.HandleFunc("/set_fullday", set_fulldayHandler) //validazione utente telelav 2
	mux.HandleFunc("/allegati", allegatiHandler)
	mux.HandleFunc("/salva_allegato", salva_allegatoHandler)
	mux.HandleFunc("/rimuovi_allegato", rimuovi_allegatoHandler)
	mux.HandleFunc("/lista_allegati", lista_allegatiHandler)

	mux.HandleFunc("/wdcal", wdcalHandler)               //entry point for all month working days
	mux.HandleFunc("/wdprevweek", wdprevweekHandler)     //entry point for previous week working days
	mux.HandleFunc("/setcons_mese", setcons_meseHandler) //entry point for all month consuntivi
	mux.HandleFunc("/setcons", setconsHandler)           //entry point for consuntivi
	mux.HandleFunc("/forbidden", adminHandler)           //entry point admin **** only on a particular domain!!
	//	mux.HandleFunc("qp2.u-s.it/forbidden", adminHandler) //entry point admin **** only on a particular domain!!
	mux.HandleFunc("/checkuser", checkuserHandler)       //validazione super utente
	mux.HandleFunc("/home", homeHandler)                 //pagina iniziale super utente
	mux.HandleFunc("/list_pres", list_presHandler)       //logbook
	mux.HandleFunc("/covid", covidHandler)               //covidlogbook
	mux.HandleFunc("/cov_save", covsaveHandler)          //covidlogbook	 save
	mux.HandleFunc("/list_daytime", list_daytimeHandler) // lista timbrature
	mux.HandleFunc("/update_time", update_timeHandler)   //modifica timbratura

	mux.HandleFunc("/get_time", get_timeHandler)     //gestione QR
	mux.HandleFunc("/get_time_sec", get_time_secHandler)     //gestione QR senza controllo IP
	mux.HandleFunc("/check_byqr", check_byqrHandler) // verifica QR
	mux.HandleFunc("/time_byqr", time_byqrHandler)   // rilevatione timbratura tramite QR

	mux.HandleFunc("/debug", debugHandler)
	mux.HandleFunc("/upload", uploadHandler)
	mux.HandleFunc("/diag", diagHandler)
	mux.HandleFunc("/nfc", nfcHandler)

	mux.HandleFunc("/richiedi-ferie", richiestaFerieHandler)
	mux.HandleFunc("/calendario", calendarioHandler)
	mux.HandleFunc("/elenco-ferie", elencoFerieHandler)
	mux.HandleFunc("/elim-ferie", elimFerieHandler)

	mux.Handle("/statics/",
		http.StripPrefix("/statics/", http.FileServer(http.Dir("./statics"))),
	)

	mux.Handle("/js/",
		http.StripPrefix("/js/", http.FileServer(http.Dir("./js"))),
	)

	mux.Handle("/dev_test/",
		http.StripPrefix("/dev_test/", http.FileServer(http.Dir("./dev_test"))),
	)

	///////////// SSE NFC

	db, err := sql.Open("postgres", connection)
	if err != nil {
		panic("could not connect to database")
	}
	defer func() {
		_ = db.Close()
	}()

	es := sse.NewServer(nil)
	defer es.Shutdown()

	mux.Handle("/nfc_sse", es)

	fmt.Println("ATTENZIONE -> VERSIONE QP2 SENZA LETTORE NFC")

	/*ctx, err := acr122u.EstablishContext()
	if err == nil {  //presente la gestione rfid-nfc
	_pollqry := `select Upper(cognome) as chi from risorse where nfctag='%s'`
	var chi string

		go ctx.ServeFunc(func(c acr122u.Card) {

				nfctag := fmt.Sprintf("%x", c.UID())
				fmt.Printf("nfc tag: %v+\t-\t%s\n", c, nfctag)
				pollqry:=fmt.Sprintf(_pollqry, nfctag)
				row := db.QueryRow(pollqry)
				fmt.Println("nfc pollqry: ", pollqry)
				err := row.Scan(&chi)
				fmt.Println("nfc chi: ", chi)
				if err != nil && err != sql.ErrNoRows {
					panic("could not query database")
				}
				es.SendMessage("/nfc_sse", sse.SimpleMessage(chi))
				chi=""
			})
		}

	*/
	/////////////

	s.Addr = ":8443"
	s.ListenAndServeTLS("./certificate.crt", "./private.key")

}
