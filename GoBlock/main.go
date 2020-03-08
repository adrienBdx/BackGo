package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

/// -- Types --

type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
}

var Blockchain []Block

type Message struct {
	BPM int
}

// Func
func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		genesisBlock := Block{0, time.Now().String(), 0, "", ""}
		spew.Dump(genesisBlock)

		Blockchain = append(Blockchain, genesisBlock)
	}()

	log.Fatal(run())
}

func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
	hash := sha256.New()
	hash.Write([]byte(record))
	hashed := hash.Sum(nil)

	return hex.EncodeToString(hashed)
}

func generateBlock(previousBlock Block, BPM int) (Block, error) {

	var newBlock Block

	newBlock.Index = previousBlock.Index + 1
	newBlock.Timestamp = time.Now().String()
	newBlock.BPM = BPM
	newBlock.PrevHash = previousBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
}

func isBlockValid(newBlock, oldBlock Block) bool {

	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

func replaceChain(newBlocks []Block) {
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
}

// API
func run() error {
	mux := makeMuxRouter()
	httpAddr := "8080" //os.Getenv("ADDR")
	log.Println("Listening on ", "8080" /*os.Getenv("ADDR")*/)

	server := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	return muxRouter
}

func handleGetBlockchain(writer http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(Blockchain, "", "  ")

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	io.WriteString(writer, string(bytes))
}

func handleWriteBlock(writer http.ResponseWriter, request *http.Request) {
	var message Message

	decoder := json.NewDecoder(request.Body)

	if err := decoder.Decode(&message); err != nil {
		respondWithJSON(writer, request, http.StatusBadRequest, request.Body)
		return
	}
	//ToDo check
	defer request.Body.Close()

	newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], message.BPM)
	if err != nil {
		respondWithJSON(writer, request, http.StatusInternalServerError, message)
		return
	}
	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
		newBlockchain := append(Blockchain, newBlock)
		replaceChain(newBlockchain)

		// Log debug
		spew.Dump(Blockchain)
	}

	respondWithJSON(writer, request, http.StatusCreated, newBlock)

}

func respondWithJSON(writer http.ResponseWriter, request *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}

	writer.WriteHeader(code)
	writer.Write(response)
}
