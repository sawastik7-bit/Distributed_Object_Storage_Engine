package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
)

var nodeAddresses = []string{

	"http://localhost:8081",
	"http://localhost:8082",
	"http://localhost:8083"}

var fileChunks = make(map[string][]string)
var chunkLocations = make(map[string]string)

var nextNodeIndex = 0



func main() {

	port:= flag.String("port","9000","coordinator port");

	flag.Parse();
	mux:=http.NewServeMux();
// first we need to build a put route here, in which we have to read a body for this 

mux.HandleFunc("PUT /files/{filename}",func(w http.ResponseWriter, r *http.Request) {
 
	        data,err:=io.ReadAll(r.Body);
			if err!=nil{
				fmt.Println(err);
				return;
			}

			fmt.Println(string(data))

			
			

			

		w.WriteHeader(http.StatusOK);
})


	log.Fatal(http.ListenAndServe(":"+*port,mux));

}