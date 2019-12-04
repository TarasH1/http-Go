package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"os/signal"
	"syscall"
)

var auth smtp.Auth

func queryParams(w http.ResponseWriter, r *http.Request) {
	/*	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}*/
	switch r.Method {
	case "GET":
		for k, v := range r.URL.Query() {
			fmt.Printf("%s: %s\n", k, v)
		}
		w.Write([]byte("Received a GET request\n"))
		/*		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
				http.ServeFile(w, r, "/home/tarash/git/book/index.html")
			})*/
	case "POST":
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", reqBody)
		w.Write([]byte("Received a POST request\n"))

		emails, ok := r.URL.Query()["email"]

		if !ok || len(emails[0]) < 1 {
			log.Println("URL Param 'email' is missing")
			return
		}
		email := emails[0]
		log.Println("URL Param 'email' is: " + email)

		//Display all request params
		for k, v := range r.URL.Query() {
			fmt.Printf("%s: %s\n", k, v)
		}

		auth = smtp.PlainAuth("", "taras.h.ua@gmail.com", "mlxqtvziciulbigo", "smtp.gmail.com")

		templateData := struct {
			URL  string
		}{
			URL:  "https://three-sides.com/pdf/tsoh.pdf",
		}
		r := NewRequest([]string{email}, "Ваша електронна копія книги", "")
		if err := r.ParseTemplate("mailTemplate.html", templateData); err == nil {
			ok, _ := r.SendEmail()
			fmt.Println(ok)
		}
	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

func main() {
	/*	auth = smtp.PlainAuth("", "taras.h.ua@gmail.com", "cakeslice", "smtp.gmail.com")
		templateData := struct {
			Name string
			URL  string
		}{
			Name: "User",
			URL:  "http://wisehands.me/book.pdf",
		}*/
	/*	r := NewRequest([]string{"research010@gmail.com"}, "Hello User", "Hello")
		if err := r.ParseTemplate("mailTemplate.html", templateData); err == nil {
			ok, _ := r.SendEmail()
			fmt.Println(ok)
		}*/
	fs := http.FileServer(http.Dir("../book"))
	http.Handle("/", http.StripPrefix("/", fs))

	http.HandleFunc("/api", queryParams)
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

	if err := smtp.SendMail(addr, auth, "taras.h.ua@gmail.com", r.to, msg); err != nil {
		log.Fatal(err)
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
