package main

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
)

func wayForPayHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("wayForPayHandler")
		fmt.Printf("wayForPayHandler request body %s\n", reqBody)

		//Display all request params
		for k, v := range r.URL.Query() {
			log.Printf("wayForPayHandler request param %s: %s\n", k, v)
		}

		//Parse JSON
		jsonData := reqBody
		var dat map[string]interface{}

		if err := json.Unmarshal([]byte(jsonData), &dat); err != nil {
			panic(err)
		}

		emailParam, ok := dat["email"].(string)
		fmt.Println("wayForPayHandler " + emailParam)
		transactionStatus := dat["transactionStatus"].(string)
		fmt.Println("wayForPayHandler " + transactionStatus)

		orderReference := dat["orderReference"].(string)
		fmt.Println("wayForPayHandler" + orderReference)
		phone := dat["phone"].(string)
		clientName := dat["clientName"].(string)

		if !ok || len(emailParam) < 1 {
			log.Println("URL Param 'email' is missing")
			return
		}
		log.Println("wayForPayHandler URL Param 'email' is: " + emailParam)

		//Mail authorization
		//TODO: TARAS what da fuck password in plain text doing here???
		auth = smtp.PlainAuth("", "3sidesplatform@gmail.com", "hjnhrjuzaxkmxzuf", "smtp.gmail.com")

		templateUserToAdminData := struct {
			Email             string
			Phone             string
			ClientName        string
			TransactionStatus string
		}{
			Email:             emailParam,
			Phone:             phone,
			ClientName:        clientName,
			TransactionStatus: transactionStatus,
		}

		rm := NewRequest([]string{"3sidesplatform@gmail.com"}, "Нове замовлення на книгу", "")
		if err := rm.ParseTemplate("orderAndDownloadAdminTemplate.html", templateUserToAdminData); err == nil {
			ok, _ := rm.SendEmail()
			fmt.Printf("SendFreeBookViaEmail email for pdf copy to admin sent... %t\n", ok)
		} else {
			log.Println(err)
		}

		templateUserData := struct {
			URL string
		}{
			URL: "https://three-sides.com/pdf/Три сторони щастя. Святосла Беш.pdf",
		}

		if transactionStatus == "Approved" {
			rm := NewRequest([]string{emailParam}, "Книга \"Три сторони щастя\"", "")
			if err := rm.ParseTemplate("orderAndDownloadUserTemplate.html", templateUserData); err == nil {
				ok, _ := rm.SendEmail()
				fmt.Printf("SendFreeBookViaEmail email for pdf copy to user sent... %t\n", ok)
			} else {
				log.Println(err)
			}
		}
		status := "accept"
		time := makeTimestamp()
		concatenated := fmt.Sprint(orderReference, ";accept;", time)

		//TODO: TARAS, replase secret with WAYFORPAY secret, but do not commit to GIT!!!
		secret := "mysecret"
		data := concatenated
		fmt.Printf("SendFreeBookViaEmail Secret: %s Data: %s\n", secret, data)

		h := hmac.New(md5.New, []byte(secret))

		// Write Data to it
		h.Write([]byte(data))

		// Get result and encode as hexadecimal string
		signature := hex.EncodeToString(h.Sum(nil))

		fmt.Println("SendFreeBookViaEmail Result: " + signature)

		response := WayForPaySuccessResponse{orderReference, status, time, signature}
		js, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)

	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}
