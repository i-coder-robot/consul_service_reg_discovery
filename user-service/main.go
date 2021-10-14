package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	consulApi "github.com/hashicorp/consul/api"
)

type User struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Foods    []food `json:"foods"`
}

type food struct {
	ID    uint64  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func registerService() {
	config := consulApi.DefaultConfig()
	consul, err := consulApi.NewClient(config)
	if err != nil {
		log.Fatalln(err)
	}

	registration := new(consulApi.AgentServiceRegistration)

	registration.ID = "user-service"
	registration.Name = "user-service"
	address := hostname()
	registration.Address = address
	p, err := strconv.Atoi(port()[1:len(port())])
	if err != nil {
		log.Fatalln(err)
	}
	registration.Port = p
	registration.Check = new(consulApi.AgentServiceCheck)
	registration.Check.HTTP = fmt.Sprintf("http://%s:%v/healthcheck", address, p)
	registration.Check.Interval = "5s"
	registration.Check.Timeout = "3s"
	consul.Agent().ServiceRegister(registration)
}

func lookupService(serviceName string) (string, error) {
	config := consulApi.DefaultConfig()
	consul, err := consulApi.NewClient(config)
	if err != nil {
		return "", err
	}
	services, err := consul.Agent().Services()
	if err != nil {
		return "", err
	}
	service := services["food-service"]
	address := service.Address
	port := service.Port
	return fmt.Sprintf("http://%s:%v", address, port), nil
}

func main() {
	registerService()
	http.HandleFunc("/healthcheck", healthcheck)
	http.HandleFunc("/userFoods", UserFoods)
	fmt.Printf("user service is up on port: %s", port())
	http.ListenAndServe(port(), nil)
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `user service is good`)
}

func UserFoods(w http.ResponseWriter, r *http.Request) {
	var foods []food
	url, err := lookupService("food-service")
	fmt.Println("URL: ", url)
	if err != nil {
		fmt.Fprintf(w, "Error. %s", err)
		return
	}
	client := &http.Client{}
	resp, err := client.Get(url + "/foods")
	if err != nil {
		fmt.Fprintf(w, "Error. %s", err)
		return
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&foods); err != nil {
		fmt.Fprintf(w, "Error. %s", err)
		return
	}
	u := User{
		ID:       666,
		Username: "欢喜-《Go语言极简一本通》|B站:面向加薪学习|公众号:面向加薪学习|微信:write_code_666",
	}
	u.Foods = foods
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&u)
}

func port() string {
	p := os.Getenv("USER_SERVICE_PORT")
	if len(strings.TrimSpace(p)) == 0 {
		return ":8080"
	}
	return fmt.Sprintf(":%s", p)
}

func hostname() string {
	hn, err := os.Hostname()
	if err != nil {
		log.Fatalln(err)
	}
	return hn
}
