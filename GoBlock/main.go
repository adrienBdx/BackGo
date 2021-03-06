package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

/// -- Types --

type Block struct {
	Index      int
	Timestamp  string
	BPM        int
	Hash       string
	PrevHash   string
	Difficulty int
	Nonce      string
}

type Message struct {
	BPM int
}

var mutex = &sync.Mutex{}
var blockServer chan []Block
var Blockchain []Block

// Func
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	blockServer = make(chan []Block)

	go func() {
		genesisBlock := Block{0, time.Now().String(), 0, "", "", 1, ""}
		spew.Dump(genesisBlock)

		Blockchain = append(Blockchain, genesisBlock)
	}()

	server, err := net.Listen("tcp", ":"+os.Getenv("port"))
	//defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleConn(conn)
	}

	log.Fatal(run())
}

func handleConn(conn net.Conn) {

	// Data input (BPM) part
	io.WriteString(conn, "Enter a new BPM:")

	scanner := bufio.NewScanner(conn)

	go func() {
		for scanner.Scan() {
			bpm, err := strconv.Atoi(scanner.Text())
			if err != nil {
				log.Printf("%v not a number: %v", scanner.Text(), err)
				continue
			}
			newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], bpm)
			if err != nil {
				log.Println(err)
				continue
			}
			if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
				newBlockchain := append(Blockchain, newBlock)
				replaceChain(newBlockchain)
			}

			blockServer <- Blockchain
			io.WriteString(conn, "\nEnter a new BPM:")
		}
	}()

	// simulate receiving broadcast
	go func() {
		for {
			time.Sleep(30 * time.Second)
			output, err := json.Marshal(Blockchain)
			if err != nil {
				log.Fatal(err)
			}
			io.WriteString(conn, string(output))
		}
	}()

	for _ = range blockServer {
		spew.Dump(Blockchain)
	}

	defer conn.Close()
}

func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash + block.Nonce
	hash := sha256.New()
	hash.Write([]byte(record))
	hashed := hash.Sum(nil)

	return hex.EncodeToString(hashed)
}

func isHashValid(hash string, difficulty int) bool {

	prefix := strings.Repeat("0", difficulty)

	return strings.HasPrefix(hash, prefix)
}

func generateBlock(previousBlock Block, BPM int) (Block, error) {

	var newBlock Block

	newBlock.Index = previousBlock.Index + 1
	newBlock.Timestamp = time.Now().String()
	newBlock.BPM = BPM
	newBlock.PrevHash = previousBlock.Hash
	newBlock.Difficulty = 1

	for i := 0; ; i++ {
		hex := fmt.Sprintf("%x", i)
		newBlock.Nonce = hex

		newHashAttempt := calculateHash(newBlock)

		if !isHashValid(newHashAttempt, newBlock.Difficulty) {
			fmt.Println(newHashAttempt, "Do more work.")
			//time.Sleep(time.Second) time consuming simulation

			continue
		}

		fmt.Println(newHashAttempt, "Valid hash, block done.")
		newBlock.Hash = newHashAttempt
		break
	}

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
	httpAddr := os.Getenv("port")
	log.Println("Listening on", os.Getenv("port"))

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

	defer request.Body.Close()

	mutex.Lock()
	newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], message.BPM)
	mutex.Unlock()

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
