package main

import (
//	"os/exec"
//	 "fmt"
)


func sendMail( to, msg string) {

sendmail2(to, "QP2 token", msg)

/*	
	cmd := "/bin/mailx"
	parms := " -s token -r personale@usmail.it "

	fmt.Println("sendmail(): \n echo " + msg + " | " + cmd + parms  +  to )
	
	sendmail := exec.Command(cmd, parms, to)
	stdin, err := sendmail.StdinPipe()
	if err != nil {
			panic(err)
	}

	sendmail.Start()
	stdin.Write([]byte(msg))
	stdin.Close()
	//sentBytes, _ := ioutil.ReadAll(stdout)
	sendmail.Wait()

*/
}

/*
func sendmail(to, subj, msg string) {

	body_ini := `<<EOF
		`
	body_fin := `
EOF
`
	body := body_ini + msg +body_fin
	
	fmt.Println("sendmail2() - executing: \n" + "mail", `-s "` + subj + `" ` + to + " " + body)
	
	sendmail := exec.Command("mail", `-s "` + subj + `" ` + to + " " + body)
//	stdin, err := sendmail.StdinPipe()
	if err != nil {
		panic(err)
	}
	sendmail.Start()
//	stdin.Write([]byte(msg))
//	stdin.Close()
	sendmail.Wait()
}
*/