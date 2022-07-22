package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "wss://www7.jungleerummyuat.com/ws", "http service address")

var socketConn = 10

func main() {

	// this is to interuupt and block main thread

	socketConn = processArg(os.Args)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	sockets := make([]*websocket.Conn, socketConn)
	for i := 0; i < socketConn; i++ {
		sockets[i] = NewWebSocket(i)
	}

	dataChan := readFile()
	for j := 0; j < socketConn; j++ {
		processRequest(dataChan, sockets[j], j)
	}
	<-c
	close(dataChan)
	os.Exit(1)
}

func processArg(arg []string) int {

	if len(arg) > 0 {
		var err error
		val := arg[1:]

		if len(val) > 0 {
			socketConn, err = strconv.Atoi(val[0])
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return socketConn
}

func processRequest(userDetail chan string, socket *websocket.Conn, id int) {

	go func() {
		for eachEmail := range userDetail {
			time.Sleep(5 * time.Second)
			tmp := strings.Split(eachEmail, ",")
			finalEmail := tmp[1]
			fmt.Println(finalEmail)
			err := socket.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("{'type':'cm-login-pass','network':'JUNGLEERUMMY','login':'%s','password':'Perform@1234','extras':{'adkey':null,'deviceInfo':{'deviceName':'Samsung Galaxy S8+','deviceType':'pwaa'},'utmParams':{'utm_source':'','utm_medium':'','utm_campaign':'','utm_term':'','utm_content':'','gclid':'','fbclid':'','fbp':'','fbc':''},'osName':'Android','osVersion':'8.0.0','referralId':0,'browserName':'Chrome','browserVersion':'87','userAgentType':'Mozilla/5.0 (Linux; Android 8.0.0; SM-G955U Build/R16NW) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.141 Mobile Safari/537.36','userAgent':'Mozilla/5.0 (Linux; Android 8.0.0; SM-G955U Build/R16NW) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.141 Mobile Safari/537.36','otpTracking':{'pageUrl':'LOGIN','activity':'FORGOT_PASSWORD','otpMobileNo':'','resend':0,'mobile_number_detection_permission':'NO','mobile_number_capturing':'MANUAL','otp_capturing_permission':'NO','otp_capturing':'MANUAL'},'pageName':'/login','mobileDRMID':'23518ad9edfaaae83ad3ba384d2b614a','gcaptkn':'','latitude':null,'longitude':null,'fireBaseAppInstanceId':null}}", finalEmail)))
			if err != nil {
				log.Println("write: from socket conn %d", err, id)
			}
		}
	}()
}

func NewWebSocket(id int) *websocket.Conn {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	log.Printf("connecting to %s with socketcnnn id %d", u.String(), id)

	c, _, err := websocket.DefaultDialer.Dial("wss://www7.jungleerummyuat.com/ws", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	done := make(chan struct{})

	go func(c *websocket.Conn) {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read: for conn %d", err, id)
				return
			}
			log.Printf("recv: %s fom conn %d", message, id)
		}
	}(c)

	go func(c *websocket.Conn) {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				err := c.WriteMessage(websocket.TextMessage, []byte("{type:'cm-rp'}"))
				if err != nil {
					log.Println("write:", err)
					return
				}
			case <-interrupt:
				log.Println("interrupt")

				// Cleanly close the connection by sending a close message and then
				// waiting (with timeout) for the server to close the connection.
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close:", err)
					return
				}
				select {
				case <-done:
				case <-time.After(time.Second):
				}
				return
			}
		}
	}(c)

	return c
}

func readFile() chan string {

	userData := make(chan string)

	f, err := os.Open("data.csv")

	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(f)

	go func(f *os.File) {
		defer f.Close()
		defer close(userData)

		for scanner.Scan() {
			userData <- scanner.Text()
		}

	}(f)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return userData
}
