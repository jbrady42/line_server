package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
)

type LineResp struct {
	Lines []map[string]interface{}
}

func handleLineRequest(w http.ResponseWriter, r *http.Request, dataDir string) {
	q := r.URL.Query()
	fname := q.Get("name")
	start := q.Get("start")
	lc := q.Get("count")

	if start == "" || fname == "" || lc == "" {
		http.Error(w, "Specify name, start and count", http.StatusInternalServerError)
		return
	}

	start_line, e1 := strconv.Atoi(start)
	line_count, e2 := strconv.Atoi(lc)

	if e1 != nil || e2 != nil {
		http.Error(w, "Specify an integer start and count", http.StatusInternalServerError)
		return
	}

	full_path := path.Join(dataDir, fname)
	if _, err := os.Stat(full_path); err != nil {
		http.Error(w, "Data file not found", http.StatusInternalServerError)
		return
	}

	log.Printf("Processing %d lines at %d for %v", line_count, start_line, full_path)

	lines, count := readLines(full_path, start_line, line_count)
	obj := lineObj(lines[0:count])

	json.NewEncoder(w).Encode(obj)
}

func main() {
	log.Println("Starting up..")
	dataDir := "."
	if len(os.Args) >= 2 {
		dataDir = os.Args[1]
	}
	log.Println("Data dir ", dataDir)

	handleLines := func(w http.ResponseWriter, r *http.Request) {
		handleLineRequest(w, r, dataDir)
	}

	http.HandleFunc("/", handleLines)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func lineObj(lines []string) LineResp {
	var ret LineResp
	ret.Lines = make([]map[string]interface{}, len(lines))

	for i, _ := range lines {
		err := json.Unmarshal([]byte(lines[i]), &ret.Lines[i])
		if err != nil {
			log.Println("json parse: ", err)
		}
	}
	return ret
}

func readLines(fname string, start, difLines int) ([]string, int) {
	file, err := os.Open(fname)
	if err != nil {
		log.Println("Error opening file", err)
		return []string{}, 0
	}
	defer file.Close()

	lines := make([]string, difLines)
	count := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() && count < difLines {
		lines[count] = scanner.Text()
		count++
	}

	if err = scanner.Err(); err != nil {
		log.Println("Error reading file", err)
	}

	return lines, count
}