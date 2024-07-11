# Paginator

## Description
This program receives a file with only one line as an input and returns an output file that is paginated in the following manner:
- All pages have 25 lines
- Each line as 80 characters
- Everything is coded in extended ASCII (0-255)
- Each line has a new line symbol at the end
- There is an end of page symbol at the end of the page

## Usage
1. Open terminal: ``Ctrl+Alt+T``
2. Clone this repository: ``git clone https://github.com/NoveliaYuki4982/paginator.git``
3. Go to the folder: ``cd paginator``
4. Execute program: ``go run paginator.go [filename]``

### Example
``go run paginator.go document.txt``

## Additional information
- Page number occupies 1 unsigned integer (4 bytes), meaning the maximum number of pages allowed for this program is 2³²-1-
- New line symbol occupies 1 byte (0x0A in ASCII)
- Line 25 of each page has 5 less bytes of information (used up by end of page symbol + page number)
- Lines 1-24 have 79 bytes of information from the original file, line 25 has only 75 bytes

## Logic flow

1. Make sure the file exists:
   - We can not read a file if it does not exist.
2. Read chunk of input file:
   - We read chunks of 1995 characters and store them in a byte array, which is the maximum amount of characters that fits into one page.
   - Offset is recalculated to not miss any word or cut it
   - If chunk it bigger than the remaining bytes to read, we handle the error by using another function to beautify the byte array.
3. Adapt byte array to our paramters and write it until there are no more chunks to read:
   - Check if words are cut in each line. If so, we write it into the next line.
   - Update offset so we do not miss any word the next time we read from the input file
4. Adapt byte array to our paramters and write the remaining bytes:
   - It is different because this byte array will always have a shorter or equal length. This way, we make this byte array's size dynamic.