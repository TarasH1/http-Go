package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
)

func sendPhysicalCopyOfBook(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}
	switch r.Method {
	case "POST":
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s\n", reqBody)
		w.Write([]byte("Received a POST request\n"))

		//Display all request params
		for k, v := range r.URL.Query() {
			log.Printf("%s: %s\n", k, v)
		}

		adminEmail := "3sidesplatform@gmail.com"
		log.Println("URL Param 'email' is: " + adminEmail)

		fullName := r.URL.Query()["fullName"][0]
		phoneNumber := r.URL.Query()["phoneNumber"][0]
		userEmail := r.URL.Query()["userEmail"][0]
		address := r.URL.Query()["address"][0]
		comment := r.URL.Query()["comment"][0]
		newPostDepartmentNumber := r.URL.Query()["newPostDepartmentNumber"][0]
		paymentType := r.URL.Query()["paymentType"][0]

		//Mail authorization
		auth = smtp.PlainAuth("", "3sidesplatform@gmail.com", "hjnhrjuzaxkmxzuf", "smtp.gmail.com")

		templateData := struct {
			FullName                string
			PhoneNumber             string
			UserEmail               string
			Address                 string
			Comment                 string
			NewPostDepartmentNumber string
			PaymentType             string
		}{
			FullName:                fullName,
			PhoneNumber:             phoneNumber,
			UserEmail:               userEmail,
			Address:                 address,
			Comment:                 comment,
			NewPostDepartmentNumber: newPostDepartmentNumber,
			PaymentType:             paymentType,
		}

		r := NewRequest([]string{adminEmail}, "Нове замовлення на книгу", "")
		if err := r.ParseTemplate("orderTemplate.html", templateData); err == nil {
			ok, _ := r.SendEmail()
			fmt.Println(ok)
		} else {
			log.Println(err)
		}

		templateUserData := struct {
			Address                 string
			Comment                 string
			NewPostDepartmentNumber string
			PaymentType             string
		}{
			Address:                 address,
			Comment:                 comment,
			NewPostDepartmentNumber: newPostDepartmentNumber,
			PaymentType:             paymentType,
		}

		rm := NewRequest([]string{userEmail}, "Книга \"Три сторони щастя\"", "")
		if err := rm.ParseTemplate("orderUserTemplate.html", templateUserData); err == nil {
			ok, _ := rm.SendEmail()
			fmt.Println(ok)
		} else {
			log.Println(err)
		}

	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}
