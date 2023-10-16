package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
)

type Timestamp struct {
	Hours        int
	Minutes      int
	Seconds      int
	Milliseconds int
}

// Takes a timestamp of the form byte(hh:mm:ss,mss) and returns a Timestamp, or an err
func timeToTimestamp(time []byte) (Timestamp, error) {
	timestamp := Timestamp{}
	if len(time) != 12 {
		return timestamp, errors.New("time must have the following format: byte(hh:mm:ss,mss)")
	}

	var err error
	timestamp.Hours, err = bytesToInt(time[:2])
	if err != nil {
		return timestamp, errors.New("error converting hours to int")
	}
	timestamp.Minutes, err = bytesToInt(time[3:5])
	if err != nil {
		return timestamp, errors.New("error converting minutes to int")
	}
	timestamp.Seconds, err = bytesToInt(time[6:8])
	if err != nil {
		return timestamp, errors.New("error converting seconds to int")
	}
	timestamp.Milliseconds, err = bytesToInt(time[10:])
	if err != nil {
		return timestamp, errors.New("error converting milliseconds to int")
	}
	return timestamp, nil

}

type Subtitle struct {
	Id    int
	Start Timestamp
	End   Timestamp
	Subs  string
}

type flag int

const (
	EOF flag = iota
	ID
	START
	END
	SUBS
)

type Parser struct {
	NextIndex int    // points to the next byte to be read
	Stream    []byte // holds the contents of the file
	Flag      flag   // state of the parser
}

func bytesToInt(toInt []byte) (int, error) {
	num, err := strconv.Atoi(string(toInt))
	if err != nil {
		return 0, err
	}
	return num, nil
}

func (p *Parser) Parse() ([]Subtitle, error) {
	subtitles := make([]Subtitle, 0)
	currSub := Subtitle{}
	eralyEOFErr := errors.New("unexpected EOF")
Loop:
	for {
		switch p.Flag {
		case ID:
			id, err := p.parseId()
			if err != nil {
				return subtitles, err
			}
			if p.Flag == EOF {
				return subtitles, eralyEOFErr
			}
			p.Flag = START
			currSub.Id = id

		case START:
			start, err := p.parseTime()
			if err != nil {
				return subtitles, err
			}
			if p.Flag == EOF {
				return subtitles, eralyEOFErr
			}
			p.Flag = END
			currSub.Start = start
		case END:
			end, err := p.parseTime()
			if err != nil {
				return subtitles, err
			}
			if p.Flag == EOF {
				return subtitles, eralyEOFErr
			}
			p.Flag = SUBS
			currSub.End = end
		case SUBS:
			subs := p.parseSubs()
			currSub.Subs = subs
			subtitles = append(subtitles, currSub)
			// flag could've been set to EOF in call to parsesubs
			if p.Flag != EOF {
				p.Flag = ID
				currSub = Subtitle{}
			}
		case EOF:
			break Loop
		}
	}
	return subtitles, nil
}

// parses and returns the ID of the current subtitle
// returns err if trouble converting id to int, or the id and nil if no error
func (p *Parser) parseId() (int, error) {
	id := make([]byte, 0)
	for curr := p.read(); curr != '\n' && p.Flag != EOF; curr = p.read() {
		id = append(id, curr)
	}
	// sometimes weird whitespaces get added to end of id? Probs bug?
	idInt, err := bytesToInt(bytes.TrimSpace(id))
	if err != nil {
		return idInt, errors.New("error parsing Id")
	}
	return idInt, nil

}

// returns timestamp if everything goes great, or err if not everything goes great. Perhaps things go badly, then an error will be returned.
func (p *Parser) parseTime() (Timestamp, error) {
	time := make([]byte, 0)
	for curr := p.read(); curr != '\n' && p.Flag != EOF; curr = p.read() {
		if curr == '-' || curr == '>' {
			continue
		}
		if curr == ' ' {
			if p.Flag == START {
				break
			} else {
				continue
			}
		}
		time = append(time, curr)
	}

	// sometimes weird whitespaces get added to end of time? Probs bug?
	return timeToTimestamp(bytes.TrimSpace(time))

}

func (p *Parser) parseSubs() string {
	subs := make([]byte, 0)
	for curr := p.read(); p.Flag != EOF; curr = p.read() {
		if curr == '\n' {
			next := p.peek()
			if next == '\n' {
				// discard the \n char
				p.read()
				break
			}
		}
		subs = append(subs, curr)
	}
	return string(subs)

}

// returns the next byte in the stream without incrementing p.NextIndex
func (p *Parser) peek() byte {
	if p.isEOF() {
		return 0
	}
	return p.Stream[p.NextIndex]
}

// returns the next byte in the stream and increments p.NextIndex
func (p *Parser) read() byte {
	next := p.peek()
	p.NextIndex++
	return next
}

func (p *Parser) isEOF() bool {
	if p.NextIndex >= len(p.Stream) {
		p.Flag = EOF
		return true
	}
	return false
}

func NewParser(srtPath string) (*Parser, error) {
	p := Parser{
		Flag:      ID,
		NextIndex: 0,
	}
	// opens file and returns error if one appears
	file, err := os.Open(srtPath)
	if err != nil {
		return &p, err
	}

	defer file.Close()

	// checks to make sure file extension is srt
	if matched, _ := regexp.MatchString(`.srt$`, file.Name()); !matched {
		return &p, err
	}

	// reads file into Parser's stream
	p.Stream, err = io.ReadAll(file)
	if err != nil {
		log.Println("Error reading file.", err)
		return &p, err
	}

	return &p, nil
}

func main() {
	parser, err := NewParser("./srt/test-EN.srt")
	if err != nil {
		fmt.Println("ERROR creating parser: ", err)
	}

	subs, err := parser.Parse()
	if err != nil {
		fmt.Println("error: ", err)
	}

	for _, sub := range subs {
		fmt.Printf("%d: %s\n\n", sub.Id, sub.Subs)
	}

}
