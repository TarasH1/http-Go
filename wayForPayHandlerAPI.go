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
		log.Println("wayForPayHandler")
		log.Printf("wayForPayHandler request body %s\n", reqBody)

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
		logValueOrError("email", emailParam, ok)

		transactionStatus, ok := dat["transactionStatus"].(string)
		logValueOrError("transactionStatus", transactionStatus, ok)

		orderReference, ok := dat["orderReference"].(string)
		logValueOrError("orderReference", orderReference, ok)

		phone, ok := dat["phone"].(string)
		logValueOrError("phone", phone, ok)

		clientName, ok := dat["clientName"].(string)
		logValueOrError("clientName", clientName, ok)

		amount, ok := dat["amount"].(float64)
		log.Printf("wayForPayHandler URL Param 'amount' is: %f\n", amount)

		if !ok || len(emailParam) < 1 {
			log.Println("URL Param 'email' is missing")
			return
		}
		log.Println("wayForPayHandler URL Param 'email' is: " + emailParam)

		//Mail authorization
		//TODO: TARAS what da fuck password in plain text doing here???
		auth = smtp.PlainAuth("", "3sidesplatform@gmail.com", "hjnhrjuzaxkmxzuf", "smtp.gmail.com")

		isPaperBook := amount == 199
		log.Printf("wayForPayHandler isPaperBook %t", isPaperBook)
		if isPaperBook {
			log.Printf("wayForPayHandler paper book paid by card scenario, continue...")
		}

		isPdfCopy := amount == 99
		log.Printf("wayForPayHandler isPdfCopy %t", isPaperBook)
		if isPdfCopy {
			log.Printf("wayForPayHandler PDF copy scenario, quiting...")
			return
		}

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
			log.Printf("wayForPayHandler email for pdf copy to admin sent... %t\n", ok)
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
				log.Printf("wayForPayHandler email for pdf copy to user sent... %t\n", ok)
			} else {
				log.Println(err)
			}
		}
		status := "accept"
		time := makeTimestamp()
		concatenated := fmt.Sprint(orderReference, ";accept;", time)

		//TODO: TARAS, replase secret with WAYFORPAY secret, but do not commit to GIT!!!
		secret := "mysecret"
		log.Printf("wayForPayHandler Secret: %s Data: %s\n", secret, concatenated)

		h := hmac.New(md5.New, []byte(secret))

		// Write Data to it
		h.Write([]byte(concatenated))

		// Get result and encode as hexadecimal string
		signature := hex.EncodeToString(h.Sum(nil))

		log.Println("wayForPayHandler Result: " + signature)

		response := WayForPaySuccessResponse{orderReference, status, time, signature}
		js, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Println("wayForPayHandler JSON response from our server: " + string(js))

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)

	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

func logValueOrError(c string, v string, ok bool) {
	if !ok {
		log.Println("wayForPayHandler ERROR dat[" + c + "]")
	} else {
		log.Println("wayForPayHandler " + c + " is " + v)
	}
}
