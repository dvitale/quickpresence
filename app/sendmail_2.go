package main

import (
    "crypto/tls"
	gomail "gopkg.in/gomail.v2" 
)

func sendmail2(to, subj, body string) {

	m := gomail.NewMessage()
	m.SetHeader("From", "personale@usmail.it")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subj)
	m.SetBody("text/html", body)

//	d := gomail.Dialer{Host: "localhost", Port: 25}	
	

	d := gomail.NewDialer("mail.usmail.it", 465, "personale@usmail.it","4ANY8eNEBULEjAp10")

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
	    panic(err)
	}
}

