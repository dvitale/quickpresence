package main

import "github.com/gorilla/sessions"

const (
	dbType                = "postgres"
	dbhost                = "HRDBHOST"
	dbuser                = " postgres "
	dbname                = " go_quickpresence "
	dbmode                = " sslmode=disable "
	connection            = "host=" + dbhost + " user=" + dbuser + " dbname=" + dbname + dbmode
	sessiontimeout uint32 = 600 //seconds
	IpWANUS               = "93.42.222.106"
	IpLANUS               = "172.16."
)

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	store      = sessions.NewCookieStore([]byte("33446a9dcf9ea060a0a6532b166da32f304af0de"))
	lastAccess = ""
)
