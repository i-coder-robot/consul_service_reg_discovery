package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
)

type food struct {
	ID    uint64  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func RegisterService() {
	config := consulapi.DefaultConfig()
	consul, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatalln(err)
	}

	registration := new(consulapi.AgentServiceRegistration)

	registration.ID = "food-service"
	registration.Name = "food-service"
	address := hostname()
	registration.Address = address
	port, err := strconv.Atoi(port()[1:len(port())])
	if err != nil {
		log.Fatalln(err)
	}
	registration.Port = port
	registration.Check = new(consulapi.AgentServiceCheck)
	registration.Check.HTTP = fmt.Sprintf("http://%s:%v/healthcheck", address, port)
	registration.Check.Interval = "5s"
	registration.Check.Timeout = "3s"
	consul.Agent().ServiceRegister(registration)
}

func Configuration(w http.ResponseWriter, r *http.Request) {
	config := consulapi.DefaultConfig()
	consul, err := consulapi.NewClient(config)
	if err != nil {
		fmt.Fprintf(w, "Error. %s", err)
		return
	}
	kv, _, err := consul.KV().Get("food-configuration", nil)
	if err != nil {
		fmt.Fprintf(w, "Error. %s", err)
		return
	}
	if kv.Value == nil {
		fmt.Fprintf(w, "Configuration empty")
		return
	}
	val := string(kv.Value)
	fmt.Fprintf(w, "%s", val)

}

func FoodList(w http.ResponseWriter, r *http.Request) {
	foods := []food{
		{
			ID:    1,
			Name:  "牛排",
			Price: 398.00,
		},
		{
			ID:    2,
			Name:  "羊排",
			Price: 368.00,
		},
		{
			ID:    3,
			Name:  "帝王蟹",
			Price: 998.00,
		},
		{
			ID:    4,
			Name:  "波士顿龙虾",
			Price: 498.00,
		},
		{
			ID:    5,
			Name:  "牛油果",
			Price: 218.00,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&foods)
}

func main() {
	RegisterService()
	http.HandleFunc("/healthcheck", healthcheck)
	http.HandleFunc("/foods", FoodList)
	http.HandleFunc("/food-configuration", Configuration)
	fmt.Printf("food service is up on port: %s", port())
	http.ListenAndServe(port(), nil)
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `ok`)
}

func port() string {
	p := os.Getenv("FOOD_SERVICE_PORT")
	if len(strings.TrimSpace(p)) == 0 {
		return ":8100"
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
