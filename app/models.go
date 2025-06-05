package main

type User struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     int64  `json:"role"`
}

type Risorsa struct {
	Id  		string `json:"id"`
	Nome  		string `json:"nome"`
}

type Cons struct {
	Chi  		string `json:"chi"`
	Quando_g  	string `json:"quando_g"`
	Htot  		string `json:"htot"`
	Note  		string `json:"note"`
}

type Cons2 struct {
	Chi  		string `json:"chi"`
	Quando_g  	string `json:"quando_g"`
	Htot  		string `json:"htot"`
	Note  		string `json:"note"`
	Remote  	string `json:"remote"`
}


type Item_pg struct {
	Giorno  		string `json:"giorno"`
	Chi  			string `json:"chi"`
	Ingresso    	string `json:"ingresso"`
	Pausa_ini    	string `json:"pausa_ini"`
	Pausa_fin    	string `json:"pausa_fin"`
	Uscita    		string `json:"uscita"`
	Note    		string `json:"note"`
	Minuti_lordi	string `json:"minuti_lordi"`
	Minuti_netti    string `json:"minuti_netti"`
}


type Dup_op struct {

	Id  		string `json:"id"`
	Chi  		string `json:"chi"`
	Quando_g  	string `json:"quando_g"`
	Quando_o    string `json:"quando_o"`
	Operazione  string `json:"operazione"`
	Prev_op  	string `json:"prev_op"`
	Prev_id  	string `json:"prev_id"`
	Remote 		string `json:"remote"`
	Auto  		string `json:"auto"`
}

type Presence struct {
	Chi  			string `json:"chi"`
	Timbratura  	string `json:"timbratura"`
	Stato    		string `json:"stato attuale"`
	DataStato    	string `json:"data stato"`
	Remote    		string `json:"remote"`
}

type Covlog struct {
	Chi  			string `json:"chi"`
	Timbratura  	string `json:"timbratura"`
	Stato    		string `json:"stato attuale"`
	DataStato    	string `json:"data stato"`
	Remote    		string `json:"remote"`
	Rilevazione    			string `json:"rilevazione"`
}

/*
type Daytime struct {

	Id  		string `json:"id"`
	Quando_g  	string `json:"quando_g"`
	Quando_o    string `json:"quando_o"`
	Operazione  string `json:"operazione"`
	Chi  		string `json:"chi"`
	Nota  		string `json:"nota"`
	Auto  		string `json:"auto"`
}
*/
type Worktime struct {

	Id  		string `json:"id"`
	Chi  		string `json:"chi"`
	Operazione  string `json:"operazione"`
	Quando_g  	string `json:"quando_g"`
	Quando_o    string `json:"quando_o"`
	Correct_o	string `json:"correct_o"`
	Nota  		string `json:"nota"`
	Auto  		string `json:"auto"`
}

type Wrp struct {
		Worktimes []Worktime
		Htot string
		Ris string
		Quando_g string
	}
	
type Allegato struct {
	Id  		string `json:"id"`
	Chi  		string `json:"chi"`
	Dsc  		string `json:"dsc"`
	Quando_g  	string `json:"quando_g"`
	Filename 	string `json:"filename"`
	Data_ins    string `json:"data_ins"`
}	

type Gest_Allegati struct {
	Chi string
	Allegati []Allegato
}
