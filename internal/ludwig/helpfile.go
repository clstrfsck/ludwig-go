/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         HELPFILE
//
// Description: Load and support indexed help file under unix.

package ludwig

import (
	"bufio"
	"bytes"
	_ "embed"
	"io"
	"os"
	"strconv"
	"strings"
)

//go:embed data/ludwighlp.idx
var oldHelpFileData []byte

//go:embed data/ludwignewhlp.idx
var newHelpFileData []byte

const (
	newHelpfileEnv    = "LUD_NEWHELPFILE"
	oldHelpfileEnv    = "LUD_HELPFILE"
	newDefaultHlpFile = "/usr/local/help/ludwignewhlp.idx"
	oldDefaultHlpFile = "/usr/local/help/ludwighlp.idx"
	skipMax           = 81 // No good reason for this value, but it's a lot of blank lines
)

// keyType represents an entry in the help index
type keyType struct {
	startPos int64
	endPos   int64
	key      string
}

var (
	helpSeeker io.ReadSeeker
	helpReader *bufio.Reader
	table      map[string]keyType
	currentKey keyType
)

// helpOpenFile opens a help file
func helpOpenFile(filename string) bool {
	var err error
	helpSeeker, err = os.Open(filename)
	if err != nil {
		return false
	}
	helpReader = bufio.NewReader(helpSeeker)
	return true
}

func helpOpenEmbedded(data []byte) bool {
	helpSeeker = bytes.NewReader(data)
	helpReader = bufio.NewReader(helpSeeker)
	return true
}

func helpTryOpenFile(envName string, defaultFilename string, embeddedData []byte) bool {
	// First try environment variable (for development/override)
	helpPath := os.Getenv(envName)
	if helpPath != "" && helpOpenFile(helpPath) {
		return true
	}
	// Then try embedded data
	if len(embeddedData) > 0 {
		return helpOpenEmbedded(embeddedData)
	}
	// Fallback to filesystem
	return helpOpenFile(defaultFilename)
}

// readIndex reads the help file index
func readIndex() bool {
	// Read in the size of the index, and the size in lines of the contents
	// Both values are on the same line, space-separated
	firstLine, err := helpReader.ReadString('\n')
	if err != nil {
		return false
	}
	parts := strings.Fields(strings.TrimSpace(firstLine))
	if len(parts) != 2 {
		return false
	}
	indexSize, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return false
	}
	contentsLines, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false
	}

	// Read in indexSize keys
	table = make(map[string]keyType)
	for range indexSize {
		indexLine, err := helpReader.ReadString('\n')
		if err != nil {
			return false
		}

		fields := strings.Fields(indexLine)
		if len(fields) != 3 {
			return false
		}

		var k keyType
		k.key = strings.TrimSpace(fields[0])
		k.startPos, err = strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return false
		}
		k.endPos, err = strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			return false
		}

		table[k.key] = k
	}

	// The key "0" is special, it's the contents page, it does NOT appear
	// in the index as it must be created on the fly while creating the
	// index file. Hence it's entry appears after the index but before the
	// bulk of the entries.
	var contents keyType
	contents.key = "0"
	currentPos, err := helpSeeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return false
	}
	// Account for buffered data
	currentPos -= int64(helpReader.Buffered())
	contents.startPos = currentPos

	for range contentsLines {
		_, err := helpReader.ReadString('\n')
		if err != nil {
			return false
		}
	}

	currentPos, err = helpSeeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return false
	}
	currentPos -= int64(helpReader.Buffered())
	contents.endPos = currentPos

	// Correct the positions in the index to point at the real offset in the file
	for key, entry := range table {
		entry.startPos += contents.endPos
		entry.endPos += contents.endPos
		table[key] = entry
	}

	// Add the contents to the table
	table[contents.key] = contents

	return true
}

// HelpfileClose closes the help file and clears the index
func HelpfileClose() {
	if helpSeeker != nil {
		if closer, ok := helpSeeker.(io.Closer); ok {
			closer.Close()
		}
		helpSeeker = nil
		helpReader = nil
	}
	currentKey.key = ""
	table = nil
}

// HelpfileOpen opens a help file (old or new version)
func HelpfileOpen(oldVersion bool) bool {
	if helpSeeker != nil { // we have done this before, don't do it again!
		return true
	}
	if oldVersion {
		if !helpTryOpenFile(oldHelpfileEnv, oldDefaultHlpFile, oldHelpFileData) {
			return false
		}
	} else {
		if !helpTryOpenFile(newHelpfileEnv, newDefaultHlpFile, newHelpFileData) {
			return false
		}
	}
	return readIndex()
}

// HelpfileRead reads the first line of a help entry for the given key
func HelpfileRead(keystr string, buffer *HelpRecord) bool {
	currentKey.key = keystr

	entry, found := table[currentKey.key]
	if !found {
		return false
	}

	// Find the offset in the helpfile for the entry and go there now
	currentKey = entry
	_, err := helpSeeker.Seek(currentKey.startPos, io.SeekStart)
	if err != nil {
		return false
	}
	// Need to create a new reader after seeking
	helpReader = bufio.NewReader(helpSeeker)

	// Get the first line in the entry and package it up in a HelpRecord
	line, err := helpReader.ReadString('\n')
	if err != nil && line == "" {
		return false
	}
	buffer.Txt = strings.TrimRight(line, "\n")
	buffer.Key = currentKey.key
	return true
}

// HelpfileNext reads the next line of the current help entry
func HelpfileNext(buffer *HelpRecord) bool {
	// if the current position >= the end of the entry return false,
	// else give back the next line nicely packaged.
	currentPos, err := helpSeeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return false
	}
	// Account for buffered data
	currentPos -= int64(helpReader.Buffered())

	if currentPos >= currentKey.endPos {
		return false
	}

	line, err := helpReader.ReadString('\n')
	if err != nil && line == "" {
		return false
	}
	buffer.Txt = strings.TrimRight(line, "\n")
	buffer.Key = currentKey.key
	return true
}
