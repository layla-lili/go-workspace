package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Block struct {
	Position   int
	Data       BookCheckout
	TimeStamp  string
	Hash       string
	PrevHash   string 

}

type BookCheckout struct {
	BookID        string  `json:"book_id"`
	User          string   `json: "user"`
	CheckoutDate  string `json: "checkout_date"`
	IsGenesis     bool  `json: "is_genesis"`
}

 type Book struct {
	ID           string   `json:"id"`
    Title        string   `json:"title"`
	Author       string   `json:"author"`
	PublishDate  string   `json:"publish_date"`
	ISBN         string   `json:"isbn:`
}

 type Blockhain struct{
	blocks [] *Block
 }

 var BlockChain *Blockhain

func (bc *Blockhain) AddBlock(data BookCheckout){
	prevBlock := bc.blocks[len(bc.blocks)-1]
	block := CreateBlock(prevBlock, data)

	if validBlock(block, prevBlock){
		bc.blocks = append(bc.blocks, block)
	}
}

func writeBlock(w http.ResponseWriter, r *http.Request){
	var checkoutitem BookCheckout

	if err := json.NewDecoder(r.Body).Decode(&checkoutitem); err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not write block: %v", err)
		w.Write([]byte("could not write block"))
		return
	}
	BlockChain.AddBlock(checkoutitem)
	resp, err := json.MarshalIndent((checkoutitem), "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not marshal payload: %v", err)
		w.Write([]byte("could not write block"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)

}

 func newBook(w http.ResponseWriter,r *http.Request){
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("cloud not create: %v", err)
		w.Write([]byte("could not create new book"))
	    return
	}
	h := md5.New()
	io.WriteString(h, book.ISBN+book.PublishDate)
	book.ID = fmt.Sprintf("%x", h.Sum(nil))

	resp, err := json.MarshalIndent(book, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not marshel payload: %v", err)
		w.Write([]byte("could not save book data"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
 }

 func (b *Block)generateHash(){
	bytes, _ := json.Marshal(b.Data)

	data := string(b.Position)+ b.TimeStamp + string(bytes)+b.PrevHash
	hash := sha256.New()
	hash.Write([]byte(data))
	b.Hash = hex.EncodeToString(hash.Sum(nil))
 }

func CreateBlock(prevBlock *Block, checkoutitem BookCheckout) *Block {
       block := &Block{}
	   block.Position = prevBlock.Position + 1
	   block.TimeStamp = time.Now().String()
	   block.Data = checkoutitem
	   block.PrevHash = prevBlock.Hash
	   block.generateHash()
	   return block
}

func validBlock(block, prevBlock *Block) bool{
   if prevBlock.Hash != block.PrevHash {
	return false 
   }
   if !block.validateHash(block.Hash){
	return false
   }
   if prevBlock.Position+1 !=  block.Position{
	return false
   }
   return true
}

func (b *Block)validateHash(hash string) bool{
	b.generateHash()
	if b.Hash != hash {
		return false
	} else {
		return true
	}

}

func GenesisBlock() *Block{
	return CreateBlock(&Block{}, BookCheckout{IsGenesis: true})
}

func newBlockchain() *Blockhain{
    return &Blockhain{[]*Block{GenesisBlock()}}
}
func getBlockchain(w http.ResponseWriter, r *http.Request){
	jbytes, err := json.MarshalIndent(BlockChain.blocks,"", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
io.WriteString(w,  string(jbytes))
}

 func main(){
	BlockChain = newBlockchain()
	r := mux.NewRouter()
	r.HandleFunc("/", getBlockchain).Methods("GET")
	r.HandleFunc("/", writeBlock).Methods("POST")
	r.HandleFunc("/new", newBook).Methods("POST")
	go func(){
		for _, block := range BlockChain.blocks{
			fmt.Printf(":Prev, hash: %x\n", block.PrevHash)
			bytes, _ := json.MarshalIndent(block.Data, "", " ")
			fmt.Printf("Data:%v\n", string(bytes))
			fmt.Printf("Hash:%x\n", block.Hash)
			fmt.Println()
		}
	}()
	log.Println("Listening on port 3000")
	log.Fatal(http.ListenAndServe(":3000", r))
 }