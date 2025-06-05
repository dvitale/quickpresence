package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	_ "github.com/lib/pq"
)

type Dipendente struct {
	ID                int
	Nome              string
	Cognome           string
	Email             string
	GiorniDisponibili int
}

type RichiestaFerie struct {
	ID           int
	DipendenteID string
	DataInizio   string
	DataFine     string
	Stato        string
	Note         string
}

type PageData struct {
	Title   string
	Content interface{}
}

type GiornoCalendario struct {
	Giorno   int
	Classe   string
	HasFerie bool
	HasSmwork bool
	HasUswork bool
	Quando_g string
}

type DatiCalendario struct {
	Anno              int
	Mese              string
	MesePrecedente    string
	MeseSuccessivo    string
	Settimane         [][]GiornoCalendario
	DipendenteID      int
	DipendenteCognome string
}

type Ferie struct {
	Data       string
	Dipendente string
	Note       string
}

type FeriePageData struct {
	Mese string 
	Title string
	Ferie []Ferie
}

var db *sql.DB

var nomiMesi = []string{
	"Gennaio", "Febbraio", "Marzo", "Aprile", "Maggio", "Giugno",
	"Luglio", "Agosto", "Settembre", "Ottobre", "Novembre", "Dicembre",
}

/*
func main() {
	var err error
	// Inizializziamo la connessione al database
	db, err = sql.Open("postgres", "postgres://postgres:postgres@qp2.u-s.it/go_quickpresence?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Verifichiamo che la connessione funzioni
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", handleHome)
	http.HandleFunc("/richiedi-ferie", handleRichiestaFerie)
	http.HandleFunc("/calendario", handleCalendario)

	log.Println("Server in ascolto sulla porta 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   "Home - Gestione Ferie",
		Content: nil,
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/home.html",
	))

	if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
*/

func elencoFerieHandler(w http.ResponseWriter, r *http.Request) {
    db, err := sql.Open(dbType, connection)
    if err != nil {
        panic(err.Error())
    }
    defer db.Close()

    mese := r.URL.Query().Get("mese")

    q := `SELECT quando_g, cognome, note FROM vw_cal_ferie_all where quando_g like '%s%%' ORDER BY quando_g `
	

    query := fmt.Sprintf(q, mese)
	
	fmt.Println("elencoFerieHandler: " , query)
	rows, err := db.Query(query)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()
    
    var ferie []Ferie
    for rows.Next() {
        var f Ferie
        err := rows.Scan(&f.Data, &f.Dipendente, &f.Note)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        ferie = append(ferie, f)
    }
    
    data := FeriePageData{
        Title: "Elenco Ferie",
		Mese: mese,
        Ferie: ferie,
    }
    
    tmpl := template.Must(template.ParseFiles(
        "tmpl/layout.html",
        "tmpl/elencoferie.html",
    ))
    
    if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}


func richiestaFerieHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("richiestaFerieHandler - url: ", r.URL.Path)

	db, err := sql.Open(dbType, connection)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "Metodo non permesso", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Errore nel parsing del form", http.StatusBadRequest)
		return
	}

	dataInizio := r.FormValue("data_inizio")
	dataFine := r.FormValue("data_fine")
	note := r.FormValue("note")
	dipendente := r.FormValue("dipendente")

	// Convertiamo le stringhe in date
	dInizio, err := time.Parse("2006-01-02", dataInizio)
	if err != nil {
		http.Error(w, "Data inizio non valida", http.StatusBadRequest)
		return
	}

	dFine, err := time.Parse("2006-01-02", dataFine)
	if err != nil {
		http.Error(w, "Data fine non valida", http.StatusBadRequest)
		return
	}

	// Validazione base delle date
	if dFine.Before(dInizio) {
		w.Write([]byte(`
			<div class="alert alert-danger mt-3">
				La data di fine non può essere precedente alla data di inizio!
			</div>
		`))
		return
	}

	// Inserimento nel database
	q := `select * from aggiungi_ferie('%s', '%s', '%s', '%s');`

	query := fmt.Sprintf(q, dipendente, dataInizio, dataFine, note)

	fmt.Println(query)

	_, err = db.Exec(query)
	if err != nil {
		log.Printf("Errore nell'inserimento della richiesta: %v", err)
		w.Write([]byte(`
			<div class="alert alert-danger mt-3">
				Errore nell'inserimento della richiesta. Riprova più tardi.
			</div>
		`))
		return
	}

	w.Write([]byte(`
		<div class="alert alert-success mt-3">
			Richiesta ferie inviata con successo!
		</div>
	`))
}

func calendarioHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("\ncalendarioHandler()")

	db, err := sql.Open(dbType, connection)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	dipendente := r.URL.Query().Get("dipendente")

	var now time.Time
	// Ottieni il parametro mese dalla querystring
	meseQuery := r.URL.Query().Get("mese")

	if meseQuery != "" {
		// Se il parametro mese è presente, parsing della data
		parsedTime, err := time.Parse("2006-01", meseQuery)
		if err != nil {
			http.Error(w, "Formato mese non valido", http.StatusBadRequest)
			return
		}
		now = parsedTime
	} else {
		// Altrimenti usa il mese corrente
		now = time.Now()
	}

	// Aggiorna il calcolo del mese precedente e successivo
	mesePrecedente := now.AddDate(0, -1, 0).Format("2006-01")
	meseSuccessivo := now.AddDate(0, 1, 0).Format("2006-01")

	// Creazione calendario
	primoDelMese := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	ultimoDelMese := primoDelMese.AddDate(0, 1, -1)

	settimane := [][]GiornoCalendario{}
	settimanaCorrente := []GiornoCalendario{}

	// Riempi i giorni vuoti all'inizio

	blankDays := (int(primoDelMese.Weekday()) + 6) % 7

	fmt.Println("mese: ", nomiMesi[now.Month()-1], "blankDays:", blankDays)

	//	for i := 1; i < int(primoDelMese.Weekday()); i++ {
	for i := 1; i <= blankDays; i++ {
		settimanaCorrente = append(settimanaCorrente, GiornoCalendario{
			Giorno: 0,
			Classe: "bg-secondary",
		})
	}
	var descr string
	// Riempi i giorni del mese
	for giorno := 1; giorno <= ultimoDelMese.Day(); giorno++ {
		dataCorrente := time.Date(now.Year(), now.Month(), giorno, 0, 0, 0, 0, time.Local)

		quando_g := fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), giorno)
		q := `select descr from vw_cal_all where quando_g='%s' and chi ilike '%s';`
		query := fmt.Sprintf(q, quando_g, dipendente)
		fmt.Println(query)

		giornoCalendario := GiornoCalendario{
			Giorno:   giorno,
			Quando_g: fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(),giorno) ,
			Classe:   "",
		}

		err := db.QueryRow(query).Scan(&descr)
		if err == nil {
			fmt.Println("descr: ", descr)
			switch descr {
				case "in_us":
					giornoCalendario.HasUswork = true	
				case "smart":
					giornoCalendario.HasSmwork = true	
				case "ferie":
					giornoCalendario.HasFerie = true					
			}
		}
		

		// Aggiungi classe per weekend
		if dataCorrente.Weekday() == time.Saturday || dataCorrente.Weekday() == time.Sunday {
			giornoCalendario.Classe = "bg-secondary"
		}

		settimanaCorrente = append(settimanaCorrente, giornoCalendario)

		// Se è domenica o ultimo giorno del mese, chiudi la settimana
		if dataCorrente.Weekday() == time.Sunday || giorno == ultimoDelMese.Day() {
			settimane = append(settimane, settimanaCorrente)
			settimanaCorrente = []GiornoCalendario{}
		}
		fmt.Println(settimanaCorrente)
	}

	data := DatiCalendario{
		Anno:              now.Year(),
		Mese:              nomiMesi[now.Month()-1],
		MesePrecedente:    mesePrecedente,
		MeseSuccessivo:    meseSuccessivo,
		Settimane:         settimane,
		DipendenteID:      -1,
		DipendenteCognome: dipendente,
	}

	//fmt.Println("data:", data)

	tmpl := template.Must(template.ParseFiles(
		"./tmpl/layout.html",
		"./tmpl/calendario.html",
	))

	if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}




func elimFerieHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("\nelimFerieHandler()")

	db, err := sql.Open(dbType, connection)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	chi := r.URL.Query().Get("chi")
	quando_g := r.URL.Query().Get("quando_g")
	
		q := `select * from elimina_ferie('%s' ,'%s');`
		query := fmt.Sprintf(q, chi, quando_g)
		fmt.Println(query)


	_, err = db.Exec(query)
	if err != nil {
		log.Printf("Errore nell'inserimento della richiesta: %v", err)
		w.Write([]byte(`
			<div class="alert alert-danger mt-3">
				Errore nell'inserimento della richiesta. Riprova più tardi.
			</div>
		`))
		return
	}
	
	w.Write([]byte(`
		<div class="alert alert-success mt-3">
			Eliminazione giornata di ferie eseguita
		</div>
		<script> 
		setTimeout(function(){ window.location = document.referrer; }, 2000);
		</script>
	`))
	
}
