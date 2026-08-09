package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)



func main(){

	port:=flag.String("port","8080","The port you want to access for storage");
	storageDir:=flag.String("storage","./storage","folder to store chunks");
	flag.Parse();
	
fmt.Println("DEBUG storageDir value is:", *storageDir)



	os.MkdirAll(*storageDir,0755);

mux:=http.NewServeMux()
mux.HandleFunc("GET /",func(w http.ResponseWriter, r *http.Request) {

	
	   _,_=  w.Write([]byte("Hello this is go "));
})

mux.HandleFunc("PUT /chunks/{id}",func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HERE SOMEONE MADE THE REQUEST");

	id:=r.PathValue("id");
 
	body,err:=io.ReadAll(r.Body);

	if err!=nil{
		http.Error(w,err.Error(),http.StatusInternalServerError);
		return;
	}

	filePath:=*storageDir + "/" + id;

	    err=os.WriteFile(filePath,body,0644);
		if err!=nil{
			http.Error(w,err.Error(),http.StatusInternalServerError);
			return;
		}

	fmt.Printf("[node on port %s] stored chunks %s (%d bytes): \n",*port , id ,len(body));

	fmt.Println(body);
	w.WriteHeader(http.StatusCreated);

})


mux.HandleFunc("GET /chunks/{id}",func(w http.ResponseWriter, r *http.Request) {
	id:=r.PathValue("id");

	buildPath:=*storageDir + "/" + id;

	

	 data,err:=os.ReadFile(buildPath);
	
	 if err!=nil{
		if os.IsNotExist(err){
			http.Error(w,"chunk not found",http.StatusNotFound);
			return;
		}

		http.Error(w,err.Error(),http.StatusInternalServerError);
		return;
	 }

	 fmt.Println("here someone made the request in the server");

	 w.Write(data);

	 

})

fmt.Printf("Node starting on port %s, storing chunks in %s \n", *port,*storageDir);
log.Fatal(http.ListenAndServe(":"+*port,mux));



}