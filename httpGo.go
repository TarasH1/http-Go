package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var auth smtp.Auth

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func main() {
	//Logging
	f, err := os.OpenFile("/tmp/orders.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	defer f.Close()
	wrt := io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)
	log.Println("App Started..")

	fs := http.FileServer(http.Dir("../book"))
	http.Handle("/", http.StripPrefix("/", fs))

	//HTTP server endpoints
	http.HandleFunc("/api", sendFreeBookViaEmail)
	http.HandleFunc("/api/payment/done", wayForPayHandler)
	http.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../book/success.html")
		//fmt.Fprint(w, "Success Page")
	})
	http.HandleFunc("/order-book", sendPhysicalCopyOfBook)

	port := ":5446"
	reloadable()
	fmt.Println("Server is listening... on port" + port)
	start := http.ListenAndServe(port, nil)
	log.Fatal(start)
}

func reloadable() {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGHUP)
	go func() {
		for {
			<-s
			fmt.Println("Reloaded")
		}
	}()
}

//Request struct
type Request struct {
	from    string
	to      []string
	subject string
	body    string
}

func NewRequest(to []string, subject, body string) *Request {
	return &Request{
		to:      to,
		subject: subject,
		body:    body,
	}
}

func (r *Request) SendEmail() (bool, error) {
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Subject: " + r.subject + "!\n"
	msg := []byte(subject + mime + "\n" + r.body)
	addr := "smtp.gmail.com:587"

	if err := smtp.SendMail(addr, auth, "3sidesplatform@gmail.com", r.to, msg); err != nil {
		fmt.Println(err)
		return false, err
	}
	return true, nil
}

func (r *Request) ParseTemplate(templateFileName string, data interface{}) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}
	r.body = buf.String()
	return nil
}

type WayForPaySuccessResponse struct {
	orderReference string
	status         string
	time           int64
	signature      string
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
