package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var auth smtp.Auth

func queryParams(w http.ResponseWriter, r *http.Request) {
	/*	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}*/
	switch r.Method {
	case "GET":
		//Display all request params
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
		//Mail authorization
		auth = smtp.PlainAuth("", "3sidesplatform@gmail.com", "hjnhrjuzaxkmxzuf", "smtp.gmail.com")

		templateAdminData := struct {
			UserEmail string
		}{
			UserEmail: email,
		}
		rm := NewRequest([]string{"3sidesplatform@gmail.com"}, "Нове завантаження книги", "")
		if err := rm.ParseTemplate("downloadFreeAdminTemplate.html", templateAdminData); err == nil {
			ok, _ := rm.SendEmail()
			fmt.Println(ok)
		}

		templateData := struct {
			URL string
		}{
			URL: "https://three-sides.com/pdf/Три сторони щастя (з реквізитами).pdf",
		}
		r := NewRequest([]string{email}, "Книга \"Три сторони щастя\"", "")
		if err := r.ParseTemplate("mailTemplate.html", templateData); err == nil {
			ok, _ := r.SendEmail()
			fmt.Println(ok)
		}
	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

func receiveData(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", reqBody)
		w.Write([]byte("Received a POST request\n"))

		//Display all request params
		for k, v := range r.URL.Query() {
			log.Printf("%s: %s\n", k, v)
		}

		//Parse JSON
		jsonData := reqBody
		var dat map[string]interface{}

		if err := json.Unmarshal([]byte(jsonData), &dat); err != nil {
			panic(err)
		}

		emailParam, ok := dat["email"].(string)
		fmt.Println(emailParam)
		transactionStatus := dat["transactionStatus"].(string)
		fmt.Println(transactionStatus)

		orderReference := dat["orderReference"].(string)
		fmt.Println(orderReference)
		phone := dat["phone"].(string)
		clientName := dat["clientName"].(string)

		if !ok || len(emailParam) < 1 {
			log.Println("URL Param 'email' is missing")
			return
		}
		log.Println("URL Param 'email' is: " + emailParam)

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
			fmt.Println(ok)
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
				fmt.Println(ok)
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
		fmt.Printf("Secret: %s Data: %s\n", secret, data)

		h := hmac.New(md5.New, []byte(secret))

		// Write Data to it
		h.Write([]byte(data))

		// Get result and encode as hexadecimal string
		signature := hex.EncodeToString(h.Sum(nil))

		fmt.Println("Result: " + signature)

		response := WayForPaySuccessResponse{orderReference, status, time, signature}
		js, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)

	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

//CORS
func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func indexHandler(w http.ResponseWriter, req *http.Request) {

	// process the request...
}

func sendData(w http.ResponseWriter, r *http.Request) {
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
		fmt.Printf("%s\n", reqBody)
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
		newPostDepartmentNumber := r.URL.Query()["newPostDepartmentNumber"][0]
		paymentType := r.URL.Query()["paymentType"][0]

		//Mail authorization
		auth = smtp.PlainAuth("", "3sidesplatform@gmail.com", "hjnhrjuzaxkmxzuf", "smtp.gmail.com")

		templateData := struct {
			FullName                string
			PhoneNumber             string
			UserEmail               string
			Address                 string
			NewPostDepartmentNumber string
			PaymentType             string
		}{
			FullName:                fullName,
			PhoneNumber:             phoneNumber,
			UserEmail:               userEmail,
			Address:                 address,
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
			NewPostDepartmentNumber string
			PaymentType             string
		}{
			Address:                 address,
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

func main() {
	fs := http.FileServer(http.Dir("../book"))
	http.Handle("/", http.StripPrefix("/", fs))

	//HTTP server endpoints
	http.HandleFunc("/api", queryParams)
	http.HandleFunc("/api/payment/done", receiveData)
	http.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../book/success.html")
		//fmt.Fprint(w, "Success Page")
	})
	http.HandleFunc("/order-book", sendData)

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
