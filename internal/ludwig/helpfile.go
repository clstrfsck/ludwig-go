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
	"io"
	"os"
	"strconv"
	"strings"
)

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
	helpfile   *os.File
	helpReader *bufio.Reader
	table      map[string]keyType
	currentKey keyType
)

// helpOpenFile opens a help file
func helpOpenFile(filename string) bool {
	var err error
	helpfile, err = os.Open(filename)
	if err != nil {
		return false
	}
	helpReader = bufio.NewReader(helpfile)
	return true
}

// helpTryOpenFile tries to open a help file from environment or default location
func helpTryOpenFile(envName string, defaultFilename string) bool {
	helpPath := os.Getenv(envName)
	if helpPath == "" || !helpOpenFile(helpPath) {
		return helpOpenFile(defaultFilename)
	}
	return true
}

// readIndex reads the help file index
func readIndex() bool {
	// Read in the size of the index, and the size in lines of the contents
	indexSizeLine, err := helpReader.ReadString('\n')
	if err != nil {
		return false
	}
	indexSize, err := strconv.ParseInt(strings.TrimSpace(indexSizeLine), 10, 64)
	if err != nil {
		return false
	}

	contentsLinesLine, err := helpReader.ReadString('\n')
	if err != nil {
		return false
	}
	contentsLines, err := strconv.ParseInt(strings.TrimSpace(contentsLinesLine), 10, 64)
	if err != nil {
		return false
	}

	// Read in indexSize keys
	table = make(map[string]keyType)
	for i := int64(0); i < indexSize; i++ {
		var k keyType

		keyLine, err := helpReader.ReadString('\n')
		if err != nil {
			return false
		}
		k.key = strings.TrimSpace(keyLine)

		startPosLine, err := helpReader.ReadString('\n')
		if err != nil {
			return false
		}
		k.startPos, err = strconv.ParseInt(strings.TrimSpace(startPosLine), 10, 64)
		if err != nil {
			return false
		}

		endPosLine, err := helpReader.ReadString('\n')
		if err != nil {
			return false
		}
		k.endPos, err = strconv.ParseInt(strings.TrimSpace(endPosLine), 10, 64)
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
	currentPos, err := helpfile.Seek(0, io.SeekCurrent)
	if err != nil {
		return false
	}
	// Account for buffered data
	currentPos -= int64(helpReader.Buffered())
	contents.startPos = currentPos

	for i := int64(0); i < contentsLines; i++ {
		_, err := helpReader.ReadString('\n')
		if err != nil {
			return false
		}
	}

	currentPos, err = helpfile.Seek(0, io.SeekCurrent)
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
	if helpfile != nil {
		helpfile.Close()
		helpfile = nil
		helpReader = nil
	}
	currentKey.key = ""
	table = nil
}

// HelpfileOpen opens a help file (old or new version)
func HelpfileOpen(oldVersion bool) bool {
	if helpfile != nil { // we have done this before, don't do it again!
		return true
	}
	if oldVersion {
		if !helpTryOpenFile(oldHelpfileEnv, oldDefaultHlpFile) {
			return false
		}
	} else {
		if !helpTryOpenFile(newHelpfileEnv, newDefaultHlpFile) {
			return false
		}
	}
	return readIndex()
}

// HelpfileOpenFile opens a specific help file
func HelpfileOpenFile(filename string) bool {
	if helpfile != nil { // we have done this before, don't do it again!
		return true
	}
	if !helpOpenFile(filename) {
		return false
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
	_, err := helpfile.Seek(currentKey.startPos, io.SeekStart)
	if err != nil {
		return false
	}
	// Need to create a new reader after seeking
	helpReader = bufio.NewReader(helpfile)

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
	currentPos, err := helpfile.Seek(0, io.SeekCurrent)
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
