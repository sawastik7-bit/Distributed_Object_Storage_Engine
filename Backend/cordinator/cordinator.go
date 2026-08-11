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
	"http://localhost:8083",
}

var fileChunks = make(map[string][]string)   // file name -> ["chunk1","chunk2","chunk3"]
var chunkLocations = make(map[string]string) // chunk id -> node address

var nextNodeIndex = 0

func main() {
	port := flag.String("port", "9000", "coordinator port")
	flag.Parse()

	mux := http.NewServeMux()

	mux.HandleFunc("PUT /files/{filename}", func(w http.ResponseWriter, r *http.Request) {
		fileName := r.PathValue("filename")
		fmt.Println("filename detected:", fileName)

		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		reader := bytes.NewReader(data)
		chunks, err := chunker.Split(reader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("Split into %d chunks\n", len(chunks))

		client := &http.Client{Timeout: 10 * time.Second}
		var orderedChunkIDs []string

		for i := 0; i < len(chunks); i++ {
			c := chunks[i]

			node := nodeAddresses[nextNodeIndex%len(nodeAddresses)]
			nextNodeIndex++

			url := node + "/chunks/" + c.Meta.ID

			req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(c.Data))
			if err != nil {
				http.Error(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				http.Error(w, "Failed to send chunk: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if resp.StatusCode != http.StatusCreated {
				resp.Body.Close()
				http.Error(w, fmt.Sprintf("node %s rejected chunk, status: %d", node, resp.StatusCode), http.StatusInternalServerError)
				return
			}

			fmt.Println("Sent chunk", c.Meta.ID, "to", node, "- status:", resp.StatusCode)
			resp.Body.Close()

			
			orderedChunkIDs = append(orderedChunkIDs, c.Meta.ID)
			chunkLocations[c.Meta.ID] = node
		}

		fileChunks[fileName] = orderedChunkIDs

		fmt.Println("fileChunks now:", fileChunks)
		fmt.Println("chunkLocations now:", chunkLocations)

		w.WriteHeader(http.StatusOK)
	})



mux.HandleFunc("GET /files/{filename}",func(w http.ResponseWriter, r *http.Request) {

	fileName:=r.PathValue("filename"); // fetching the file from the url 

	client:= &http.Client{
		Timeout: 10*time.Second,
	}

	fmt.Println(fileName);


	          chunkArr:=   fileChunks[fileName];

      for i:=0;i<len(chunkArr);i++{

		         port:=chunkLocations[chunkArr[i]];
           chunkId:=chunkArr[i];

				 url:=port + "/chunks/"+chunkId;

		req,err:=http.NewRequest(http.MethodGet,url,nil);
		
		if err!=nil{

http.Error(w,err.Error(),http.StatusInternalServerError);
			return;
		}

		resp, err:=client.Do(req);

		if err!=nil{
			http.Error(w,err.Error(),http.StatusInternalServerError);
			return;
		}

		fmt.Println("Sent the request to the url :", url);
				
       fmt.Println(resp.StatusCode);

	   w.WriteHeader(http.StatusOK);

	  }



	w.WriteHeader(http.StatusOK);


})





	log.Fatal(http.ListenAndServe(":"+*port, mux))
}