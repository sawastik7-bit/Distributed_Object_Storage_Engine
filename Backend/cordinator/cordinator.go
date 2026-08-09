package main

import (
	"flag"
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



	log.Fatal(http.ListenAndServe(":"+*port,mux));

}