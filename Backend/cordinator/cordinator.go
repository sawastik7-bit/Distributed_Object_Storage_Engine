package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/sawastik7-bit/FileStorage/chunker"
)

var nodeAddresses = []string{
	"http://localhost:8081",
	"http://localhost:8082",
	"http://localhost:8083",
}

var mu sync.Mutex;
var wg sync.WaitGroup;

var fileChunks = make(map[string][]string)   // file name -> ["chunk1","chunk2","chunk3"]
var chunkLocations = make(map[string]string) // chunk id -> node address

var nextNodeIndex = 0

func main() {
	port := flag.String("port", "9000", "coordinator port")
	flag.Parse()

	mux := http.NewServeMux()
	client := &http.Client{Timeout: 10 * time.Second}

	// ---------- UPLOAD ----------
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

		var orderedChunkIDs []string

		// making a channel to handle errors inside each go routine if it gets failed
		errCh:= make(chan error, len(chunks));
		for i := 0; i < len(chunks); i++ {
			c := chunks[i]

			wg.Add(1);
			go func(index int,chunk chunker.Chunk){
				defer wg.Done();
			mu.Lock();
		node := nodeAddresses[nextNodeIndex%len(nodeAddresses)]
			nextNodeIndex++
			mu.Unlock();

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

			mu.Lock();  // this is the first mutex lock so that multiple go routines might not write to same location 
			orderedChunkIDs = append(orderedChunkIDs, c.Meta.ID)
			chunkLocations[c.Meta.ID] = node  // this is one map 
			mu.Unlock();
		}(i,c);
		}

		wg.Wait()

		close(errCh)
 // printing the total errors in go routines

		for err := range errCh {
			fmt.Println("upload error:", err)
			http.Error(w, "one or more chunks failed to upload: "+err.Error(), http.StatusInternalServerError)
			return
		}

mu.Lock()  // same mutex lock for the other map , to avoid overriding the place
		fileChunks[fileName] = orderedChunkIDs // this is the second map
		mu.Unlock(); 

		fmt.Println("fileChunks now:", fileChunks)
		fmt.Println("chunkLocations now:", chunkLocations)

		w.WriteHeader(http.StatusOK)
	})

	// ---------- DOWNLOAD ----------
	mux.HandleFunc("GET /files/{filename}", func(w http.ResponseWriter, r *http.Request) {
		fileName := r.PathValue("filename")

		chunkArr, exists := fileChunks[fileName]
		if !exists || len(chunkArr) == 0 {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}

		var reassembled bytes.Buffer // grows as we append each chunk's bytes, in order

		for i := 0; i < len(chunkArr); i++ {
			chunkID := chunkArr[i]
			node, ok := chunkLocations[chunkID]
			if !ok {
				http.Error(w, "missing location for chunk "+chunkID, http.StatusInternalServerError)
				return
			}

			resp, err := client.Get(node + "/chunks/" + chunkID)
			if err != nil {
				http.Error(w, "failed to fetch chunk: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				http.Error(w, fmt.Sprintf("node %s returned status %d for chunk %s", node, resp.StatusCode, chunkID), http.StatusInternalServerError)
				return
			}

			chunkData, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				http.Error(w, "failed to read chunk body: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// verify integrity before trusting the bytes
			if !chunker.Verify(chunkData, chunkID) {
				http.Error(w, "checksum mismatch for chunk "+chunkID, http.StatusInternalServerError)
				return
			}

			reassembled.Write(chunkData)
			fmt.Println("Fetched chunk", chunkID, "from", node, "-", len(chunkData), "bytes")
		}

		w.Write(reassembled.Bytes())
	})

	log.Fatal(http.ListenAndServe(":"+*port, mux))
}