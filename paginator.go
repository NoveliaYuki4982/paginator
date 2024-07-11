package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

const MAX_BYTES = 4
const CHARACTERS_PER_LINE = 80
const LINES_PER_PAGE = 25
const BITS_PER_BYTE = 8
const END_OF_PAGE = 12
const AVAILABLE_CHARACTERS_PER_PAGE = 1995
const NEW_LINE = 0x0A
const SPACE = 0x20
const writeError = "error reading from or writing to the file"

func usage() {
	fmt.Println("Usage: go run paginator.go [filename]")
}

func displayError(text string, err error) {
	fmt.Println(text)
	log.Fatal(err)
}

func writePage(file *os.File, buffer []byte, pageNumber uint32, waitGroup *sync.WaitGroup) {
	
	if _, err := file.Seek(int64((pageNumber-1)*LINES_PER_PAGE*CHARACTERS_PER_LINE), 0); err != nil {
		displayError(writeError, err)
	}
	if _, err := file.Write(buffer); err != nil {
		displayError(writeError, err)
	}
	if err := binary.Write(file, binary.LittleEndian, uint8(END_OF_PAGE)); err != nil {
		displayError(writeError, err)
	}
	if err := binary.Write(file, binary.LittleEndian, pageNumber); err != nil {
		displayError(writeError, err)
	}
	defer waitGroup.Done()
}

func findLastSpace(buffer []byte, startRange int, endRange int) (int, error) {
	for i := endRange - 1; i >= startRange; i-- {
		if buffer[i] == ' ' {
			return i + 1, nil
		}
	}
	return endRange, nil
}

func beautify(buffer []byte) ([]byte, int, error) {
	buffBeauty := []byte{}
	lastEndReading := 0
	finished := false

	// Add new line at the end of each line of the page
	for line := 0; line < LINES_PER_PAGE && !finished; line++ {
		takenCharacters := 0

		if line == 24 {
			takenCharacters = 5 // page number
		}

		pos, err := findLastSpace(buffer, lastEndReading, lastEndReading+CHARACTERS_PER_LINE-1-takenCharacters)
		if err != nil {
			return buffer, -1, err
		}

		i := 0
		for i < CHARACTERS_PER_LINE-takenCharacters-1 {
			if lastEndReading+i < pos {
				buffBeauty = append(buffBeauty, buffer[lastEndReading+i])
			} else if buffer[lastEndReading+i] == 0 {
				finished = true
				break
			} else {
				break
			}
			i++
		}
		if buffer[lastEndReading+i] != 0 {
			buffBeauty = append(buffBeauty, NEW_LINE)
		}
		lastEndReading = pos
	}
	return buffBeauty, lastEndReading, nil
}

func main() {

	// Make sure the file exists
	if len(os.Args) < 2 {
		usage()
		displayError("", errors.New("missing argument: filename"))
	}
	filename := os.Args[1]
	_, err := os.Stat(filename)
	if err != nil {
		usage()
		fmt.Printf("File %s", filename)
		fmt.Println(" not found. Please, provide a valid filename")
		log.Fatal(err)
	}
	input, err := os.Open(filename)
	if err != nil {
		usage()
		fmt.Printf("File %s", filename)
		fmt.Println(" could not be opened. Please, check permissions or any other cause that might be making the program fail")
		log.Fatal(err)
	}
	defer input.Close()

	pageNumber := 1
	offset := 0
	waitGroup := sync.WaitGroup{}

	output, err := os.Create("output.txt")
	if err != nil {
		displayError(writeError, err)
	}

	for {
		// Read chunk of input file
		buffer := make([]byte, AVAILABLE_CHARACTERS_PER_PAGE)
		if _, err = input.Seek(int64(offset), 0); err != nil {
			displayError(writeError, err)
			log.Fatal(err)
		}
		_, err = io.ReadFull(input, buffer)
		var offsetBuf int
		
		if err == nil || err == io.EOF || err == io.ErrUnexpectedEOF {

			// Adapt read bytes to the parameters requested
			beautifiedBuffer, offsetBuffer, err2 := beautify(buffer)
			if err2 != nil {
				displayError(writeError, err2)
			}
			offsetBuf = offsetBuffer

			// Write in parallel
			waitGroup.Add(1)
			go writePage(output, beautifiedBuffer, uint32(pageNumber), &waitGroup)

			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
		} else {
			displayError(writeError, err)
		}

		offset += offsetBuf
		pageNumber++
	}

	waitGroup.Wait()

	fmt.Println("File is ready")
}
