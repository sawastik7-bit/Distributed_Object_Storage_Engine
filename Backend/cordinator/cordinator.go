package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/sawastik7-bit/FileStorage/chunker"
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
 
fileName:= r.PathValue("filename");
fmt.Println(fileName);

fmt.Println("filename detected :",fileName);


	        data,err:=io.ReadAll(r.Body);
			fmt.Println(data);
			reader:=bytes.NewReader(data);
			

			chunks, err:=chunker.Split(reader);

			if err!=nil{
				fmt.Println(err);
				return;
			}

			

			
			
fmt.Printf("Split into %d chunks\n", len(chunks))
for _, c := range chunks {
    fmt.Printf("  chunk %d: id=%s size=%d\n", c.Meta.Index, c.Meta.ID, c.Meta.Size)
}
	client:=&http.Client{
		Timeout: 10 * time.Second,
	}		

for i:=0;i<len(chunks);i++{
c:=chunks[i];

node:=nodeAddresses[nextNodeIndex%len(nodeAddresses)];
nextNodeIndex++;

url:=node + "/chunks/" + c.Meta.ID;

req, err:=http.NewRequest(http.MethodPut,url,bytes.NewReader(c.Data));

if err!=nil{
	http.Error(w,"Failed to create request : " + err.Error(), http.StatusInternalServerError);
	return;
}

resp, err:= client.Do(req);

if err!=nil{
	http.Error(w,"Failed to send chunk : " + err.Error(), http.StatusInternalServerError);
	return;
}

fmt.Println("Sent chunk", c.Meta.ID, "to", node , "-status:", resp.StatusCode);

defer resp.Body.Close();
}


		w.WriteHeader(http.StatusOK);
})


	log.Fatal(http.ListenAndServe(":"+*port,mux));

}