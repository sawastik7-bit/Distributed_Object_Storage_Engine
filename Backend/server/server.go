package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const storageDir="./storage"

func main(){


	os.MkdirAll(storageDir,0755);

mux:=http.NewServeMux()
mux.HandleFunc("GET /api/fetchstorage",func(w http.ResponseWriter, r *http.Request) {

	
	   _,_=  w.Write([]byte("Hello this is go "));
})

mux.HandleFunc("PUT /chunks/{id}",func(w http.ResponseWriter, r *http.Request) {

	id:=r.PathValue("id");

	body,err:=io.ReadAll(r.Body);

	if err!=nil{
		http.Error(w,err.Error(),http.StatusInternalServerError);
		return;
	}

	filePath:=storageDir + "/" + id;

	    err=os.WriteFile(filePath,body,0644);
		if err!=nil{
			http.Error(w,err.Error(),http.StatusInternalServerError);
			return;
		}

	fmt.Printf("Received a request body : %s \n",body);

	fmt.Println(body);
	w.WriteHeader(http.StatusCreated);

})

log.Fatal(http.ListenAndServe(":8080",mux));



}