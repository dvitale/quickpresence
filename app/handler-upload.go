package main

import (
//	"crypto/rand"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"html/template"
	"os"
	"time"
	"path/filepath"
	"database/sql"
	_ "github.com/lib/pq"
)

const maxUploadSize = 10 * 1024 * 1024 // 10 mb
const uploadPath = "./uploadedfiles"
const allegatiPath = "./allegatifiles"

func allegatiHandler(w http.ResponseWriter, r *http.Request) {
//wrapper pagina allegati.html	
	var appo struct { Chi string}
	
	fmt.Println("allegatiHandler - url: ", r.URL.Path)

	r.ParseForm()

	appo.Chi =  r.FormValue("chi")
	
	t := template.Must(template.ParseFiles("./tmpl/allegati.html"))
	
	err=t.Execute(w, appo) 
	
	_=chkErrWeb(err, w)
}

func lista_allegatiHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("lista_allegatiHandler - url: ", r.URL.Path)

	var gest_Allegati Gest_Allegati
	var allegato Allegato 
	
	qrytpl := `select chi, id::text, dsc, quando_g, filename, to_char(data_ins,'yyyy-mm-dd HH24:MM') data_ins from allegati  where chi ='%s'	`

	r.ParseForm()

	chi :=  r.FormValue("chi")
			
	db, err := sql.Open(dbType, connection)	
	if err != nil {
		panic(err.Error()) 
	} 
	defer db.Close()
	
	qry := fmt.Sprintf(qrytpl,chi)
	fmt.Println("lista_allegatiHandler" ,qry)

	rows, err := db.Query(qry)
	if chkErrWeb(err, w) {
		return
	}
	for rows.Next() {
		err = rows.Scan(&allegato.Chi, &allegato.Id, &allegato.Dsc, &allegato.Quando_g, &allegato.Filename, &allegato.Data_ins)
		if chkErrWeb(err, w) {
			return
		}
		gest_Allegati.Allegati = append(gest_Allegati.Allegati, allegato)
	}

	t := template.Must(template.ParseFiles("./tmpl/lista_allegati.html"))
	
	gest_Allegati.Chi = chi
	err=t.Execute(w, gest_Allegati) 
	
	_=chkErrWeb(err, w)
	
}

func rimuovi_allegatoHandler(w http.ResponseWriter, r *http.Request) {

	var result string
	
//	dumpRequest(w, r)

	del_tpl :=  `delete from allegati where id=%s returning id` 
	
	fmt.Println("rimuovi_allegatoHandler: \n" + formatRequest(r))
	
	r.ParseForm()
	
	id := r.FormValue("id")
	chi := r.FormValue("chi")
	
	db, err := sql.Open(dbType, connection)		
	if err != nil {
		panic(err.Error()) 
	}
	defer db.Close()

	del_qry := fmt.Sprintf(del_tpl, id)		
	fmt.Println("rimuovi_allegatoHandler: ", del_qry)
	db.QueryRow(del_qry).Scan(&result)
	//redirect to lista_allegati
	http.Redirect(w, r, "/lista_allegati?chi=" +  chi, 301)
}

func salva_allegatoHandler(w http.ResponseWriter, r *http.Request) {

	var result string

	ins_tpl :=  `select * from ins_allegato ('%s', 'tmp-file-name', '%s', '%s')` //ins_allegato (dsc, filename, quando_g, chi)
	upd_tpl :=  `update allegati set filename = '%s' where id= %s returning id` 
	
	fmt.Println("salva_allegatoHandler: \n" + formatRequest(r))
	
	r.ParseMultipartForm(maxUploadSize)
	
	dsc := r.FormValue("dsc")
	quando_g := r.FormValue("quando_g")
	chi := r.FormValue("chi")
	
	//session, _ := store.Get(r, "gohrsession")	
	
	fileName := chi + "_" + quando_g

	fmt.Println("salva_allegatoHandler fileName:", fileName)
	// validate file size
	
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
		return
	}

	// parse and validate file and post parameters
//	fileType := r.PostFormValue("type")

	file, _, err := r.FormFile("uploadFile")
	if err != nil {
		renderError(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		renderError(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}

	// check file type, detectcontenttype only needs the first 512 bytes
	filetype := http.DetectContentType(fileBytes)

	switch filetype {
	case "image/jpeg", "image/jpg":
	case "image/gif", "image/png":
	case "application/pdf":
		break
	default:
		renderError(w, "INVALID_FILE_TYPE", http.StatusBadRequest)
		return
	}
	
//		fileEndings, err := mime.ExtensionsByType(fileType)
	fileEndings, err := mime.ExtensionsByType(filetype)
	if err != nil {
		renderError(w, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
		return
	}
	
	db, err := sql.Open(dbType, connection)		
	if err != nil {
		panic(err.Error()) 
	}
	defer db.Close()

	ins_qry := fmt.Sprintf(ins_tpl,dsc, quando_g, chi)		
	fmt.Println("salva_allegatoHandler", ins_qry)
	db.QueryRow(ins_qry).Scan(&result)
	
	fileName = result + "_" + fileName
	upd_qry := fmt.Sprintf(upd_tpl, fileName + fileEndings[0], result)		
	fmt.Println("salva_allegatoHandler", upd_qry)
	db.QueryRow(upd_qry).Scan(&result)

	newPath := filepath.Join(allegatiPath, fileName+fileEndings[0])
	fmt.Printf("salva_allegatoHandler  Filepath: %s\n",  newPath)

	// write file
	newFile, err := os.Create(newPath)
	if err != nil {
		renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
		return
	}
	defer newFile.Close() // idempotent, okay to call twice
	if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("SUCCESS"))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "gohrsession")
	
	fileName := session.Values["lastop"].(string) + "_" + time.Now().Format("2006-01-02 15:04:05")
	
	fmt.Println("uploadFileHandler fileName:", fileName)
	// validate file size
	
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
		return
	}

	// parse and validate file and post parameters
//	fileType := r.PostFormValue("type")

	file, _, err := r.FormFile("uploadFile")
	if err != nil {
		renderError(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		renderError(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}

	// check file type, detectcontenttype only needs the first 512 bytes
	filetype := http.DetectContentType(fileBytes)
	
	fmt.Println("uploadFileHandler filetype:", filetype)

	switch filetype {
	case "image/jpeg", "image/jpg":
	case "image/gif", "image/png":
	case "application/pdf":
		break
	default:
		renderError(w, "INVALID_FILE_TYPE", http.StatusBadRequest)
		return
	}
	
//		fileEndings, err := mime.ExtensionsByType(fileType)
	fileEndings, err := mime.ExtensionsByType(filetype)
	if err != nil {
		renderError(w, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
		return
	}
	newPath := filepath.Join(uploadPath, fileName+fileEndings[0])
	fmt.Printf("uploadFileHandler  Filepath: %s\n",  newPath)

	// write file
	newFile, err := os.Create(newPath)
	if err != nil {
		renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
		return
	}
	defer newFile.Close() // idempotent, okay to call twice
	if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("SUCCESS"))
}

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

func randToken(len int) string {
/*	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b) */
	return fmt.Sprintf(time.Now().Format("2006-01-02 15:04:05"))
}
