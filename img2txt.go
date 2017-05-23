package main

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"github.com/go-redis/redis"
)

func callShellScript(s string) string {
	cmd := exec.Command("/bin/sh", "-c", s)
	var out bytes.Buffer

	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return out.String()
}

func img2txt(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, err := template.ParseFiles("./img2txt.gptl")
		if err != nil {
			err.Error()
		}
		t.Execute(w, nil)
	} else {
		result, _:= ioutil.ReadAll(r.Body)
		r.Body.Close()
		filename := "./upload/" + r.Header.Get("filename")
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
		f.Write(result)
		if err != nil {
			err.Error()
		}
		f.Close()
		txt := callShellScript("tesseract " + filename + " stdout")
		response := strings.TrimSpace(txt)
		w.Write([]byte(response))
		log.Println("img2txt " + filename + " " + response)
	}
}

func getcid(w http.ResponseWriter, r *http.Request) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", 	// no password set
		DB:       0,  	// use default DB
	})

	if err := client.Incr("app_img2txt_cid").Err(); err != nil {
		panic(err)
	}

	cid, _ := client.Get("app_img2txt_cid").Result()
	w.Write([]byte(cid))
}

func main() {
	http.HandleFunc("/getcid", getcid)
	http.HandleFunc("/img2txt", img2txt)
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal("listenAndServe: ", err)
	}
}
