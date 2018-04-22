package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
)

type LineSource struct {
	CurrentLine int
	Stream      io.ReadCloser
	Reader      *bufio.Reader
}

type Server struct {
	sources map[string]*LineSource
	dataDir string
}

type LineResp struct {
	Lines []map[string]interface{}
}

var compressors = map[string]string{
	"xz": "xz",
	"gz": "gzip",
}

type CmdReadCloser struct {
	io.ReadCloser
	*exec.Cmd
}

type LineParams struct {
	fname string
	start int
	count int
}

func (t *CmdReadCloser) Close() error {
	// Close stream
	t.ReadCloser.Close()
	// Signal process cleanup
	return t.Cmd.Wait()
}

func (t *Server) CloseAll() {
	for _, v := range t.sources {
		v.Stream.Close()
	}
}

func (t *Server) handleLineRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

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

	full_path := path.Join(t.dataDir, fname)
	if _, err := os.Stat(full_path); err != nil {
		http.Error(w, "Data file not found", http.StatusInternalServerError)
		return
	}

	log.Printf("Processing %d lines at %d for %v", line_count, start_line, full_path)

	lines, count := t.readLines(LineParams{
		fname: full_path,
		start: start_line,
		count: line_count,
	})
	obj := lineObj(lines[0:count])

	json.NewEncoder(w).Encode(obj)
}

func NewServer(dir string) *Server {
	return &Server{
		sources: make(map[string]*LineSource),
		dataDir: dir,
	}
}

func main() {
	log.Println("Starting up..")
	dataDir := "."
	if len(os.Args) >= 2 {
		dataDir = os.Args[1]
	}
	log.Println("Data dir ", dataDir)

	server := NewServer(dataDir)
	defer server.CloseAll()

	http.HandleFunc("/", server.handleLineRequest)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
