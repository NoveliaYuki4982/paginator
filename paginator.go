package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
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

func writePage(file *os.File, buffer []byte, pageNumber uint32) {
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
	buffBeauty := make([]byte, AVAILABLE_CHARACTERS_PER_PAGE)
	lastEndReading := 0

	// Add new line at the end of each line of the page
	for line := 0; line < LINES_PER_PAGE; line++ {
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
				buffBeauty[line*CHARACTERS_PER_LINE+i] = buffer[lastEndReading+i]
			} else {
				break
			}
			i++
		}
		buffBeauty[line*CHARACTERS_PER_LINE+i] = NEW_LINE
		lastEndReading = pos
	}
	return buffBeauty, lastEndReading, nil
}

func beautifyEnd(buffer []byte) ([]byte, int, error) {
	buffBeauty := make([]byte, AVAILABLE_CHARACTERS_PER_PAGE)
	lastEndReading := 0
	finished := false
	bytesWritten := 0

	for line := 0; line < LINES_PER_PAGE && !finished; line++ {
		takenCharacters := 0

		pos, err := findLastSpace(buffer, lastEndReading, lastEndReading+CHARACTERS_PER_LINE-1-takenCharacters)
		if err != nil {
			return buffer, -1, err
		}

		i := 0
		for i < CHARACTERS_PER_LINE-takenCharacters-1 {
			if lastEndReading+i < pos {
				
				buffBeauty[line*CHARACTERS_PER_LINE+i] = buffer[lastEndReading+i]
				if buffer[lastEndReading+i] == 0 {
					break
				}
			}
			i++
			bytesWritten++
		}
		buffBeauty[line*CHARACTERS_PER_LINE+i] = NEW_LINE
		bytesWritten++
		lastEndReading = pos
	}
	fmt.Println(string(buffBeauty))

	// Eliminate useless characters
	finalBuff := buffBeauty[0:bytesWritten]

	return finalBuff, lastEndReading, nil
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
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				
				// Adapt byte array to our paramters and write the remaining bytes
				beautifiedBuffer, _, err2 := beautifyEnd(buffer)				
				if err2 != nil {
					displayError(writeError, err2)
				}

				writePage(output, beautifiedBuffer, uint32(pageNumber))
				break
			}
		}
		
		// Adapt byte array to our paramters and write it until there are no more chunks to read
		beautifiedBuffer, offsetBuf, err := beautify(buffer)
		if err != nil {
			displayError(writeError, err)
		}

		writePage(output, beautifiedBuffer, uint32(pageNumber))

		offset += offsetBuf
		pageNumber++
	}

	fmt.Printf("File is ready")
}
