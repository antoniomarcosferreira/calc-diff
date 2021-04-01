package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nsf/jsondiff"
)

// CalcTests is struct to base of json file
type CalcTests struct {
	CalcTests []CalcTest `json:"calcTests"`
}

// CalcTest is struct for the itens in json file
type CalcTest struct {
	ID       string `json:"id"`
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

func readData(fileName string) ([][]string, error) {

	f, err := os.Open(fileName)

	if err != nil {
		return [][]string{}, err
	}
	defer f.Close()

	r := csv.NewReader(f)

	// skip first line
	if _, err := r.Read(); err != nil {
		return [][]string{}, err
	}

	records, err := r.ReadAll()

	if err != nil {
		return [][]string{}, err
	}

	return records, nil
}

func calc(s string, u string) string {
	var jsonStr = []byte(s)
	req, err := http.NewRequest("POST", u, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.Status != "200 OK" {
		panic("Service not return a status code: " + resp.Status)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}

// Clean remove json keys indiferent to the test
func Clean(b string) []byte {
	var i interface{}
	if err := json.Unmarshal([]byte(b), &i); err != nil {
		panic(err)
	}

	if m, ok := i.(map[string]interface{}); ok {
		delete(m, "dtES")
		delete(m, "dtEmissao")
		delete(m, "dtPagto")
		delete(m, "log")
	}

	out, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}

	return out
}

// Compare calc the scenarios in two server
//func Compare(id string, s string, u1 string, u2 string, w *csv.Writer, wg *sync.WaitGroup) {
func Compare(id string, s string, u1 string, u2 string, w *csv.Writer) {
	//defer wg.Done()

	retA := Clean(calc(s, u1))
	retB := Clean(calc(s, u2))

	opts := jsondiff.DefaultJSONOptions()
	diff, text := jsondiff.Compare(retA, retB, &opts)
	res := ""

	if diff != jsondiff.FullMatch {
		w.Write([]string{id, res, text})
		res = "divergent"
	} else {
		res = "ok"
	}

	fmt.Printf("Test ID: %v, %s\n", id, res)
}

// Run this with three methods
// url := "https://rt-prod.taxweb.com.br:443/taxgateway/webapi/taxrules/calctaxes"
//url := "https://rt-extrafarma.taxweb.com.br:443/taxgateway/webapi/taxrules/calctaxes"

func main() {
	/*
		        In this aplication we define a path for json file and
				test it in two servers

				Example to run:

				comparTest https://rt-prod.taxweb.com.br:443/taxgateway/webapi/taxrules/calctaxes https://rt-extrafarma.taxweb.com.br:443/taxgateway/webapi/taxrules/calctaxes
	*/
	start := time.Now()
	filePath := os.Args[1]
	server1 := os.Args[2]
	server2 := os.Args[3]
	outPath := strings.Replace(filePath, ".csv", "-out.csv", 1)

	records, err := readData(filePath)

	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(outPath)
	defer f.Close()

	if err != nil {
		log.Fatalln("failed to open file", err)
	}
	w := csv.NewWriter(f)
	defer w.Flush()
	reader := []string{"CALCTEST ID", "TEST", "DIFERENCES"}

	w.Write(reader)

	//var wg sync.WaitGroup
	//wg.Add(len(records))

	for _, record := range records {
		calcTest := CalcTest{
			ID:       record[0],
			Input:    record[1],
			Expected: record[2],
		}
		//go Compare(calcTest.ID, calcTest.Input, server1, server2, w, &wg)
		Compare(calcTest.ID, calcTest.Input, server1, server2, w)
	}
	//wg.Wait()
	endTime := time.Since(start)
	fmt.Printf("Test completed, time: %s\n", endTime.String())
	w.Write([]string{"END", endTime.String(), "", ""})
}
