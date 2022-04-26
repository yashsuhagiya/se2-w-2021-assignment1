package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

var hbase_host string = "hbase"
var name = os.Getenv("name")

func main() {

	conn := connect()
	defer conn.Close()

	for conn.State() != zk.StateHasSession {
		fmt.Printf(" %s is loading Zookeeper from gserve ...\n", name)
		second := time.Second
		time.Sleep(30 * second)
	}

	fmt.Printf(" %s is connected with Zookeeper...\n", name)
	flags := int32(zk.FlagEphemeral)
	acl := zk.WorldACL(zk.PermAll)

	gserv, err := conn.Create("/grproxy/"+name, []byte(name+":9091"), flags, acl)
	errorfunc(err)
	fmt.Printf("create ephemeral node: %+v\n", gserv)

	startServer()
}

func connect() *zk.Conn {
	zksStr := "zookeeper:2181"
	zks := strings.Split(zksStr, ",")
	conn, _, err := zk.Connect(zks, time.Second)
	errorfunc(err)
	return conn
}

func errorfunc(err error) {
	if err != nil {
		fmt.Printf("ERROR: %+v\n", err)
	}
}

func encoder(unencodedJSON []byte) string {
	var unencodedRows RowsType
	json.Unmarshal(unencodedJSON, &unencodedRows)
	encodedRows := unencodedRows.encode()
	encodedJSON, _ := json.Marshal(encodedRows)
	return string(encodedJSON)
}

func decoder(encodedJSON []byte) string {
	var encodedRows EncRowsType
	json.Unmarshal(encodedJSON, &encodedRows)
	decodedRows, err := encodedRows.decode()
	if err != nil {
		fmt.Printf("%+v", err)
	}
	deCodedJSON, _ := json.Marshal(&decodedRows)
	return string(deCodedJSON)
}

func startServer() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":9091", nil))
}

func handler(writer http.ResponseWriter, req *http.Request) {

	if req.Method == "POST" || req.Method == "PUT" {

		encodedJsonByte, err := ioutil.ReadAll(req.Body)
		errorfunc(err)

		encodedJSON := encoder(encodedJsonByte)
		fmt.Println("encodedJSON : ", string(encodedJSON))

		req.Header.Set("Content-type", "application/json")
		addBook(encodedJSON)
		fmt.Fprintf(writer, "an %s\n", "POST")

	} else if req.Method == "GET" {
		req.Header.Set("Accept", "application/json")
		responseData := getBooks()
		fmt.Fprintf(writer, "Response from hbase:\n\n %s\n", string(responseData))
	} else {
		fmt.Fprintf(writer, "Invalid Request from Client")
	}
	fmt.Fprintf(writer, "    proudly served by %s", name)

}

func addBook(encodedJSON string) {

	req_url := "http://" + hbase_host + ":8080/se2:library/fakerow"
	resp, err := http.Post(req_url, "application/json", bytes.NewBuffer([]byte(encodedJSON)))

	if err != nil {
		fmt.Printf("Error from response: %+v", err)
		return
	}
	defer resp.Body.Close()
}

func getBooks() string {

	req_url := "http://" + hbase_host + ":8080/se2:library/*"
	req, _ := http.NewRequest("GET", req_url, nil)
	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, getErr := client.Do(req)
	errorfunc(getErr)
	encodedJsonByte, err := ioutil.ReadAll(resp.Body)
	errorfunc(err)
	decodedJSON := decoder(encodedJsonByte)

	defer resp.Body.Close()
	return decodedJSON
}
