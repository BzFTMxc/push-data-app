package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	shortuuid "github.com/lithammer/shortuuid"
)

var topicTokens = sync.Map{}

type resp struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func marshal(s *resp) []byte {
	bytes, err := json.Marshal(s)
	if err != nil {
		log.Print(err)
		return make([]byte, 0)
	}
	return bytes

}

func error(data string) []byte {
	s := &resp{
		Status: "ERR",
		Data:   data,
	}
	return marshal(s)
}

func success(data string) []byte {
	s := &resp{
		Status: "OK",
		Data:   data,
	}
	return marshal(s)
}

func main() {

	port, ok := os.LookupEnv("PORT")

	if !ok {
		port = "8080"
	}

	log.Printf("Starting server on port %s\n", port)

	http.HandleFunc("/token", GetToken)
	http.HandleFunc("/push", PushData)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func GetToken(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	topic := r.Form.Get("topic")
	if len(topic) == 0 {
		fmt.Fprintf(w, "%s", error("InvalidTopic"))
		return
	}
	token, exists := topicTokens.LoadOrStore(topic, shortuuid.New())
	if !exists {
		go func() {
			waitTime := (1 * time.Second) + time.Duration(rand.Int63n(time.Duration(2*time.Second).Nanoseconds()))
			time.Sleep(waitTime)
			topicTokens.Delete(topic)
		}()
	}
	fmt.Fprintf(w, "%s", success(token.(string)))
}

func PushData(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	topic := r.Form.Get("topic")
	token := r.Form.Get("token")

	if topic == "piidata" {
		fmt.Fprintf(w, "%s", error("TopicRejected"))
		return
	}

	savedToken, exists := topicTokens.Load(topic)
	if !exists {
		fmt.Fprintf(w, "%s", error("TokenExpired"))
		return
	}

	if savedToken.(string) != token {
		fmt.Fprintf(w, "%s", error("TokenMismatch"))
		return
	}
	select {
	case <-time.After(time.Duration(rand.Int63n(time.Duration(10 * time.Millisecond).Nanoseconds()))):
		fmt.Fprintf(w, "%s", error("Retry"))
		return
	case <-time.After(time.Duration(rand.Int63n(time.Duration(10 * time.Millisecond).Nanoseconds()))):
		fmt.Fprintf(w, "%s", success("ACK"))
		return
	case <-time.After(time.Duration(rand.Int63n(time.Duration(10 * time.Millisecond).Nanoseconds()))):
		fmt.Fprintf(w, "%s", error("RateLimit"))
		return
	case <-time.After(time.Duration(rand.Int63n(time.Duration(10 * time.Millisecond).Nanoseconds()))):
		fmt.Fprintf(w, "%s", success("ACK"))
		return
	}
}
