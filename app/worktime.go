package main
import (
	"time"
	"fmt"
	"strings"
	"strconv"
)


/*
/// gestione entrata
1. tutte le giornate cominciano con IN e finiscono con OUT (bisogna aspettare l'inizio del nuovo giorno per il calcolo, causa OUT cron-based)
2. tutti gli OUT minori o uguali a 8.00 annullano i relativi IN (ai fini del conteggio del tempo lavorato) [[LOOP1]]
3. eventuali IN prima delle 8.00 non annullati vengono riportati alle 8.00 [[LOOP2]]

///gestione pausa
4. il primo OUT >= 13.00 e <= 14.00 viene riportato a 13.00  [[LOOP1]]
5. eventuali ulteriori OUT nell'intervallo 13.00-14.00 si azzerano coi precedenti IN (x conteggio del tempo lavorato) [[LOOP2]]
6. l'eventuale IN non annullato nell'intervallo 13.00-14.00 viene riportato a 14.00 [[LOOP3]]
7. se non esiste un OUT nell'intervallo 13.00-14.00 viene sottratta un'ora dal totale  [[LOOP1]]
7.bis se esiste un OUT prima delle 13.00 seguito da un IN dopo le 14.00 non viene sottratta un'ora dal totale  [[LOOP1]]
7.ter se esiste un OUT prima delle 13.00 seguito da un IN tra le 13 e le 14.00 non viene sottratta un'ora dal totale ma l'IN viene riportato alle 14:00 [[LOOP1]]

///gestione uscita
8. il primo OUT > 18.30 viene riportato a 18.30 insieme ad eventuali IN > 18:30 [[LOOP3]]
9. eventuali successivi OUT annullano i relativi IN (x conteggio del tempo lavorato) [[LOOP4]]
10. gli intervalli di uscita entro i 5 minuti non vengono calcolati   [[LOOP5]]

///aggiornamento campo di calcolo per tibrature non modificate 
98. corretto <- effettivo  [[LOOP98]]

///calcolo ore lavorate
99. somma(uscite) - somma(entrate)  [[LOOP99]]
*/

const (
	
	wt1_s="08:00"
	wt2_s="13:00"
	wt3_s="14:00"
	wt4_s="18:30"
)

func recalcWt(wrp Wrp)  Wrp {
	
	var next_o time.Time 
		
	w1 := str2time(wt1_s)
	w2 := str2time(wt2_s)
	w3 := str2time(wt3_s)
	w4 := str2time(wt4_s)
	pausa := 60 // durata pausa da sottrarre anche nel caso di mancata pausa
	wt := 0 //totale minuti lavorati
	
	fmt.Printf("recalcWt input: =====\n %v \n====\n\n", wrp.Worktimes)
	
	for i, v := range wrp.Worktimes {		  //LOOP1

//		fmt.Println("L1 - id=", wrp.Worktimes[i].Id," - i=", i, " - quando=", v.Quando_o)

		quando_o := str2time(v.Quando_o)

		if i < len(wrp.Worktimes) - 1 { // esiste almeno una timbratura successiva
			next_o = str2time(wrp.Worktimes[i+1].Quando_o)
		} else {
			next_o = str2time("00:00")
		}
		
		if v.Operazione == "OUT" &&  quando_o.Before(w1) && len(v.Correct_o) == 0 { //punto 2.
			wrp.Worktimes[i].Correct_o="0:0" //annulla OUT
			wrp.Worktimes[i-1].Correct_o="0:0" //annulla IN precedente	
			fmt.Println("punto 2 - i=", i, wrp.Worktimes[i])
		}
		if v.Operazione == "OUT" &&  (quando_o.After(w2) || quando_o == w2 ) &&  quando_o.Before(w3) && len(v.Correct_o) == 0 { //punto 4.

			if i>0 && str2time(wrp.Worktimes[i-1].Quando_o).After(w2) { //***correzione per uscite multiple durante pausa
			} else {
				wrp.Worktimes[i].Correct_o=wt2_s //correggi OUT a w2 
			} 

			fmt.Println("punto 4 , punto 7 - i=", i, wrp.Worktimes[i])
			pausa = 0					//punto 7.
		}
		if v.Operazione == "OUT" &&  quando_o.Before(w2) &&  !next_o.Before(w2) { // 7.bis e 7.ter
			fmt.Println("punto 7bis - i=", i, wrp.Worktimes[i])
			pausa = 0					
		}
	}

//	fmt.Println("dopo L1",  wrp.Worktimes)

	for i, v := range wrp.Worktimes {		//LOOP2
		quando_o := str2time(v.Quando_o)
		if v.Operazione == "IN" &&  quando_o.Before(w1) { //punto 3.
			wrp.Worktimes[i].Correct_o=wt1_s //correggi IN entrata
			fmt.Println("punto 3 - i=", i, wrp.Worktimes[i])
		}	else {
//			fmt.Println("not punto 3 (IN before w1) - i=", i, wrp.Worktimes[i])
		}
		if v.Operazione == "OUT" &&  quando_o.After(w2) &&  quando_o.Before(w3) && len(v.Correct_o) == 0 { //punto 5.
			wrp.Worktimes[i].Correct_o="0:0" //annulla OUT
			wrp.Worktimes[i-1].Correct_o="0:0" //annulla IN precedente	
			fmt.Println("punto 5 - i=", i, wrp.Worktimes[i])
		}
	}
	
//	fmt.Println("dopo L2",  wrp.Worktimes)

	for i, v := range wrp.Worktimes {		//LOOP3
		quando_o := str2time(v.Quando_o)
		if v.Operazione == "IN" && quando_o.After(w2) && quando_o.Before(w3) && len(v.Correct_o) == 0 { //punto 6.
			wrp.Worktimes[i].Correct_o=wt3_s //correggi IN ritorno pausa
			fmt.Println("punto 8 - i=", i, wrp.Worktimes[i])
		}	
		if /* v.Operazione == "OUT" && */ quando_o.After(w4) && len(v.Correct_o) == 0 { //punto 8. : IN e OUT
			wrp.Worktimes[i].Correct_o=wt4_s //correggi uscita
			fmt.Println("punto 8 - i=", i, wrp.Worktimes[i])
		}	
	}

//	fmt.Printf("\ndopo L3 :\n %v+\n\n", wrp.Worktimes)
	
	for i, v := range wrp.Worktimes {		  //LOOP4
		quando_o := str2time(v.Quando_o)
		if v.Operazione == "OUT" &&  quando_o.After(w4) && len(v.Correct_o) == 0 { //punto 9.
			wrp.Worktimes[i].Correct_o="0:0" //annulla OUT
			wrp.Worktimes[i-1].Correct_o="0:0" //annulla IN precedente	
			fmt.Println("punto 2 - i=", i, wrp.Worktimes[i])
		}
	}

//	fmt.Printf("dopo L4 :\n %v+\n\n", wrp.Worktimes)
		
	for i, v := range wrp.Worktimes {		  //LOOP5
		if v.Operazione == "IN" && i>0 && len(v.Correct_o) == 0 { //punto 10. per ogni Rientro (i>0)
			in_o := wrp.Worktimes[i].Quando_o
			out_o  := wrp.Worktimes[i-1].Quando_o
			if strtm2min(in_o) - strtm2min(out_o) < 6 {
				wrp.Worktimes[i].Correct_o = wrp.Worktimes[i-1].Quando_o //annulla intervallo con OUT=in
				fmt.Println("punto 10 - i=", i, wrp.Worktimes[i])
			}
		}
	}

//	fmt.Println("dopo L5", wrp.Worktimes)

	for i, v := range wrp.Worktimes {		//LOOP98
		if len(v.Correct_o) == 0 { //copia nel campo "orario corretto"  non coinvolto nelle correzioni il valore "orario effettivo"
			wrp.Worktimes[i].Correct_o = wrp.Worktimes[i].Quando_o 
		}
		
		if i == len(wrp.Worktimes) -1 && wrp.Worktimes[i].Operazione == "IN" && i != 0{ // se l'ultima operazione è IN si deve annullare 
			wrp.Worktimes[i].Correct_o="0:0" 
		}
	}

//	fmt.Println("dopo L98", wrp.Worktimes)

	for i, v := range wrp.Worktimes {		//LOOP99
	
		if v.Operazione == "IN" {
			fmt.Println(i, "##IN - wt prima:", wt, "v.Correct_o:",v.Correct_o)
			wt -= strtm2min(v.Correct_o)
			fmt.Println(i, "##IN - wt correct", wt)
		} else {
			wt += strtm2min(v.Correct_o)			
			//fmt.Println(i, "OUT : add correct_o", wt)
		}
		fmt.Println("L99 - i:", i, v.Correct_o, v.Operazione, " - subtot: ", wt)
	}
	
	
		//##5 nov 2021 -  da implementare ## if (lastout < inipausa) OR (firstin > finpausa ) NON SOTTRARRE PAUSA 
	last:= len(wrp.Worktimes) - 1
	first := 0
	if str2time(wrp.Worktimes[last].Quando_o).Before(w2) || str2time(wrp.Worktimes[first].Quando_o).After(w3) {
		pausa = 0
	}
	
	wrp.Htot=fmt.Sprintf("%2.2d:%2.2d", (wt - pausa)/60, (wt - pausa)%60)
	fmt.Println("recalcWt() - totale=: ",  wrp.Htot)
	
	return wrp
}

func str2time(str string) time.Time {
	
	layout := "15:04"
	t, err := time.Parse(layout, str)
	if err != nil {
		fmt.Println("str2time", err)
	}
	return t
}

func strtm2min(in string) int {
	
	
	s := strings.Split(in, ":")
	h,_ := strconv.Atoi(s[0])
	m,_ := strconv.Atoi(s[1])
	ret:=h * 60 + m
	//fmt.Println("strtm2min in: ", in, " ret:", ret)
	return ret
}
