package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
)

func sendFreeBookViaEmail(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("SendFreeBookViaEmail")
		fmt.Printf("%s\n", reqBody)
		w.Write([]byte("Received a POST request\n"))

		emails, ok := r.URL.Query()["email"]

		if !ok || len(emails[0]) < 1 {
			log.Println("SendFreeBookViaEmail URL Param 'email' is missing")
			return
		}
		email := emails[0]
		log.Println("\n SendFreeBookViaEmail URL Param 'email' is: " + email)

		//Display all request params
		for k, v := range r.URL.Query() {
			fmt.Printf("SendFreeBookViaEmail %s: %s\n", k, v)
		}
		//Mail authorization
		auth = smtp.PlainAuth("", "3sidesplatform@gmail.com", "password", "smtp.gmail.com")

		templateAdminData := struct {
			UserEmail string
		}{
			UserEmail: email,
		}
		rm := NewRequest([]string{"3sidesplatform@gmail.com"}, "Нове завантаження книги", "")
		if err := rm.ParseTemplate("downloadFreeAdminTemplate.html", templateAdminData); err == nil {
			ok, _ := rm.SendEmail()
			fmt.Printf("SendFreeBookViaEmail email to admin sent... %t\n", ok)
		}

		templateData := struct {
			URL string
		}{
			URL: "https://three-sides.com/pdf/Три сторони щастя (з реквізитами).pdf",
		}

		r := NewRequest([]string{email}, "Книга \"Три сторони щастя\"", "")
		if err := r.ParseTemplate("mailTemplate.html", templateData); err == nil {
			ok, _ := r.SendEmail()
			fmt.Printf("SendFreeBookViaEmail email to user sent... %t\n", ok)
		}
	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}
