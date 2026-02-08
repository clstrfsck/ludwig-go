/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// ! Name:        LUDWIGHLPBLD
//
// Description: This program converts a sequential Ludwig help file into
//              an indexed file for fast access.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

const (
	entrySize = 77
	keySize   = 4

	defaultInputFile  = "ludwighlp.t"
	defaultOutputFile = "ludwighlp.idx"
)

func processFiles(in io.Reader, out io.Writer) error {
	section := "0"

	var index bytes.Buffer
	var contents bytes.Buffer
	var body bytes.Buffer

	indexLines := 0
	contentsLines := 0

	reader := bufio.NewReader(in)

	for {
		flag, err := reader.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		var line string
		if flag == '\n' {
			flag = ' '
			line = ""
		} else {
			lineBytes, err := reader.ReadBytes('\n')
			if err == io.EOF && len(lineBytes) > 0 {
				line = string(lineBytes)
			} else if err != nil {
				break
			} else {
				line = string(lineBytes[:len(lineBytes)-1])
			}

			if len(line) > entrySize {
				if flag != '!' && flag != '{' {
					fmt.Fprintf(os.Stderr, "Line too long--truncated\n")
					fmt.Fprintf(os.Stderr, "%s>>\n", line)
				}
				line = line[:entrySize]
			}
		}

		switch flag {
		case '\\':
			if len(line) > 0 {
				switch line[0] {
				case '%':
					body.WriteString("\\%\n")
				case '#':
					if section != "0" {
						fmt.Fprintf(&index, "%8d\n", body.Len())
					}
				default:
					if section != "0" {
						fmt.Fprintf(&index, "%8d\n", body.Len())
					}
					if len(line) >= keySize {
						section = line[:keySize]
					} else {
						section = line
					}
					if section != "0" {
						indexLines++
						fmt.Fprintf(&index, "%4s %8d", section, body.Len())
					}
				}
			}
		case '+':
			contentsLines++
			contents.WriteString(line + "\n")
			body.WriteString(line + "\n")
		case ' ':
			if section == "0" {
				contentsLines++
				contents.WriteString(line + "\n")
			} else {
				body.WriteString(line + "\n")
			}
		case '{', '!':
			// Ignore these flags
		default:
			fmt.Fprintf(os.Stderr, "Illegal flag character.\n")
			fmt.Fprintf(os.Stderr, "%c%s>>\n", flag, line)
		}

		if flag == '\\' && len(line) > 0 && line[0] == '#' {
			break
		}
	}

	fmt.Fprintf(out, "%d %d\n", indexLines, contentsLines)
	io.Copy(out, &index)
	io.Copy(out, &contents)
	io.Copy(out, &body)

	return nil
}

func main() {
	infile := defaultInputFile
	if len(os.Args) > 1 {
		infile = os.Args[1]
	}

	in, err := os.Open(infile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", infile, err)
		os.Exit(1)
	}
	defer in.Close()

	outfile := defaultOutputFile
	if len(os.Args) > 2 {
		outfile = os.Args[2]
	}

	out, err := os.Create(outfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", outfile, err)
		os.Exit(1)
	}
	defer out.Close()

	if err := processFiles(in, out); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing files: %v\n", err)
		os.Exit(1)
	}
}
