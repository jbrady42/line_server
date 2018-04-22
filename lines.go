package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

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

func openFile(fname string) (io.ReadCloser, error) {
	var file io.ReadCloser
	var err error
	if fileCompressed(fname) {
		var tmp *CmdReadCloser
		tmp, err = openFileComp(fname)
		file = io.ReadCloser(tmp)
	} else {
		file, err = os.Open(fname)
	}

	return file, err
}

func (t *Server) openStream(p LineParams) (*bufio.Reader, int, error) {
	// If already open but past requested line
	if s, ok := t.sources[p.fname]; ok && s.CurrentLine > p.start {
		log.Printf("Current line %d, start line %d\n", s.CurrentLine, p.start)
		s.Stream.Close()
		delete(t.sources, p.fname)
		log.Printf("Closing stream %v\n", p.fname)
	}

	// if not open
	if _, ok := t.sources[p.fname]; !ok {
		f, err := openFile(p.fname)
		if err != nil {
			return nil, 0, err
		}
		log.Printf("Opening new stream %v\n", p.fname)
		t.sources[p.fname] = &LineSource{
			Stream:      f,
			CurrentLine: 0,
			Reader:      bufio.NewReader(f),
		}
	}

	reader := t.sources[p.fname].Reader
	ns := p.start - t.sources[p.fname].CurrentLine

	// set stream
	return reader, ns, nil
}

func (t *Server) readLines(p LineParams) ([]string, int) {
	// file, err := openFile(p.fname)
	reader, newStart, err := t.openStream(p)

	if err != nil {
		log.Println("Error opening file", err)
		return []string{}, 0
	}

	// defer file.Close()

	lines := make([]string, p.count)
	count := 0
	seekCount := 0

	// reader := t. //bufio.NewReader(file)

	// Skip lines
	for seekCount < newStart {
		reader.ReadString('\n')
		log.Println("seek")
		seekCount++
	}

	// Read data
	var line string
	for err == nil && count < p.count {
		line, err = reader.ReadString('\n')
		lines[count] = line
		count++
	}

	if err != nil {
		log.Println("Error reading file", err)
	}

	s, _ := t.sources[p.fname]
	s.CurrentLine += count + seekCount

	return lines, count
}

func fileCompressed(fname string) bool {
	for k := range compressors {
		if strings.HasSuffix(fname, k) {
			return true
		}
	}
	return false
}

func openFileComp(fname string) (*CmdReadCloser, error) {
	compType := ""
	for k := range compressors {
		if strings.HasSuffix(fname, k) {
			compType = k
		}
	}

	comp := compressors[compType]

	proc := exec.Command(comp, "-d", "-c", fname)

	outPipe, err := proc.StdoutPipe()
	if err != nil {
		log.Println("Error opening file", err)
		return nil, err
	}

	proc.Start()

	streamWrap := &CmdReadCloser{
		ReadCloser: outPipe,
		Cmd:        proc,
	}
	return streamWrap, nil
}
