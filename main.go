package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var etcdClient *clientv3.Client

/* =========================
   ETCD INITIALIZATION
========================= */

func initEtcd() {
	var err error

	etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		log.Fatal("Failed to connect to etcd:", err)
	}

	log.Println("Connected to etcd successfully")
}

/* =========================
   HTTP HANDLERS
========================= */

func formHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "index.html")
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	userData := map[string]string{
		"name":  r.FormValue("name"),
		"email": r.FormValue("email"),
		"phone": r.FormValue("phone"),
	}

	jsonData, err := json.Marshal(userData)
	if err != nil {
		http.Error(w, "Failed to serialize data", http.StatusInternalServerError)
		return
	}

	key := fmt.Sprintf("/users/%d", time.Now().Unix())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = etcdClient.Put(ctx, key, string(jsonData))
	if err != nil {
		http.Error(w, "Failed to store data in etcd", http.StatusInternalServerError)
		return
	}

	log.Println("Stored user data in etcd with key:", key)
	fmt.Fprintln(w, "User data stored successfully!")
}

/* =========================
   MAIN
========================= */

func main() {
	initEtcd()

	http.HandleFunc("/", formHandler)
	http.HandleFunc("/submit", submitHandler)

	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
