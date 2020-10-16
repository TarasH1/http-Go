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
	"time"
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

		//Parse request body
		amount, ok := dat["amount"].(float64)
		log.Printf("wayForPayHandler 'amount' is: %f\n", amount)

		transactionStatus, ok := dat["transactionStatus"].(string)
		logValueOrError("transactionStatus", transactionStatus, ok)

		emailParam, ok := dat["email"].(string)
		logValueOrError("email", emailParam, ok)

		orderReference, ok := dat["orderReference"].(string)
		logValueOrError("orderReference", orderReference, ok)

		isPdfCopy := amount == 99
		if isPdfCopy {
			log.Printf("wayForPayHandler PDF copy scenario, sending link to book over email...")
			sendDigitalCopyEmailLink(emailParam, transactionStatus)
		}

		isPaperBook := amount == 199
		if isPaperBook {
			log.Printf("wayForPayHandler paper book paid by card scenario, do not need to send any email...")
		}

		// Make response to WayForPay
		status := "accept"
		time := makeTimestamp()
		signature := generateSignature(orderReference, status, time)

		response := WayForPaySuccessResponse{orderReference, status, time, signature}
		js, err := json.Marshal(response)
		if err != nil {
			log.Println("wayForPayHandler JSON response from our server error: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("wayForPayHandler JSON response from our server: %+v\n", response)
		log.Println("wayForPayHandler JSON raw response from our server: " + string(js))

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)

	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

type WayForPaySuccessResponse struct {
	OrderReference string `json:"orderReference"`
	Status         string `json:"status"`
	Time           int64  `json:"time"`
	Signature      string `json:"signature"`
}

func makeTimestamp() int64 {
	now := time.Now()
	nanos := now.UnixNano()
	secs := nanos / 1000000000
	return secs
}

func generateSignature(orderReference string, status string, time int64) string {

	concatenated := fmt.Sprint(orderReference, ";"+status+";", time)

	//TODO: replace secret with WAYFORPAY secret, but do not commit to GIT!!!
	secret := "mysecret"
	log.Printf("wayForPayHandler  generateSignature Secret: %s Concatenated String: %s\n", secret, concatenated)

	h := hmac.New(md5.New, []byte(secret))

	// Write Data to it
	h.Write([]byte(concatenated))

	// Get result and encode as hexadecimal string
	signature := hex.EncodeToString(h.Sum(nil))

	log.Println("wayForPayHandler generateSignature Signature: " + signature)
	return signature
}

func sendDigitalCopyEmailLink(clientEmail string, transactionStatus string) {
	auth = smtp.PlainAuth("", "3sidesplatform@gmail.com", "hjnhrjuzaxkmxzuf", "smtp.gmail.com")

	log.Println("wayForPayHandler sendEmails 'email' is: " + clientEmail)
	templateUserData := struct {
		URL string
	}{
		URL: "https://three-sides.com/pdf/Три сторони щастя. Святосла Беш.pdf",
	}

	if transactionStatus == "Approved" {
		rm := NewRequest([]string{clientEmail}, "Книга \"Три сторони щастя\"", "")
		if err := rm.ParseTemplate("orderAndDownloadUserTemplate.html", templateUserData); err == nil {
			ok, _ := rm.SendEmail()
			log.Printf("wayForPayHandler sendEmails email for pdf copy to user sent... %t\n", ok)
		} else {
			log.Println(err)
		}
	}

}

func logValueOrError(c string, v string, ok bool) {
	if !ok {
		log.Println("wayForPayHandler ERROR dat[" + c + "]")
	} else {
		log.Println("wayForPayHandler " + c + " is " + v)
	}
}
