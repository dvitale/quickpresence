package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
	"io"
	"io/ioutil"
	"os"
    "strconv"
    "log"
)

const (
	key="itopinonavevanonipoti"
)

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func encrypt(data []byte, passphrase string) []byte {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

func decrypt(data []byte, passphrase string) []byte {

	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Println(err.Error())
	}
	return plaintext
}

func encryptFile(filename string, data []byte, passphrase string) {
	f, _ := os.Create(filename)
	defer f.Close()
	f.Write(encrypt(data, passphrase))
}

func decryptFile(filename string, passphrase string) []byte {
	data, _ := ioutil.ReadFile(filename)
	return decrypt(data, passphrase)
}

func getToken() string {
	timestamp := fmt.Sprint(time.Now().Unix())
	token := fmt.Sprintf("%x",encrypt([]byte(timestamp), key))
	fmt.Println("getToken()", token)
	return token
}

func chkToken(token string) string {
	
	ciphertext, _ := hex.DecodeString(token)
	fmt.Println("chkToken() ciphertext:", ciphertext)
	fmt.Println("chkToken() key:", key)
	decrtext := decrypt(ciphertext, key)
	fmt.Println("chkToken() decrypted:", decrtext)
	plaindata, err := strconv.ParseUint(string(decrtext), 10, 32)
	fmt.Println("chkToken() plaindata", plaindata, " - now: ", time.Now().Unix())

    if err != nil {
        log.Println(err)
    }
	
	if uint32(time.Now().Unix()) - uint32(plaindata) < 180 {
		return "OK"
	} else {
		return "NOK"
	}
}

func maintest() {
	
	timestamp := fmt.Sprint(time.Now().Unix())
	fmt.Println("Starting the application...")
	ciphertext := encrypt([]byte(timestamp), key)
	fmt.Printf("Encrypted: %x\n", ciphertext)
	
	plaindata, err := strconv.ParseUint(string(decrypt(ciphertext, key)), 10, 32)
    if err != nil {
        fmt.Println(err)
    }
	
	fmt.Printf("Decrypted time spent: %d\n", uint32(time.Now().Unix()) - uint32(plaindata))
	fmt.Printf("Decrypted time spent: %d\n", uint32(time.Now().Unix()) - uint32(1561724465))
}

