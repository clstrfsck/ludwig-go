/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:        FILESYS
//
// Description: The File interface for Go.

package ludwig

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	filesysNL     = "\n"
	filesysNLSize = 1
)

// removeBackupFiles removes backup files in the specified range
func removeBackupFiles(backupFile string, versions []int64, start, end int) {
	for i := start; i < end; i++ {
		fileName := backupFile + strconv.FormatInt(versions[i], 10)
		SysUnlink(fileName)
	}
}

// startsWith checks if haystack starts with needle
func startsWith(haystack, needle string) bool {
	return len(haystack) >= len(needle) && haystack[:len(needle)] == needle
}

// toArgv converts a command line string to an argv-style slice
func toArgv(cmdline string) []string {
	return strings.Fields(cmdline)
}

// FilesysCreateOpen opens a file for input or creates a file for output
func FilesysCreateOpen(fyle *FileObject, rfyle *FileObject, ordinaryOpen bool) bool {
	// Check the file has a 'Z' in the right place (debug check)
	if fyle.Zed != 'Z' {
		ScreenMessage("FILE and FILESYS definition of file_object disagree.")
		return false
	}

	fyle.LCounter = 0
	if !fyle.OutputFlag { // open file for reading
		if fyle.Filename == "" {
			return false
		}
		if !ordinaryOpen { // really executing a command
			fd := SysOpenCommand(fyle.Filename)
			if fd == -1 {
				ScreenMessage("Cannot create pipe")
				return false
			}
			fyle.Fd = fd
		} else {
			if !SysExpandFilename(&fyle.Filename) {
				ScreenMessage(fmt.Sprintf("Error in filename (%s)", fyle.Filename))
				return false
			}
			fs := SysFileStatus(fyle.Filename)
			if !fs.Valid || fs.IsDir {
				return false
			}
			fyle.Fd = SysOpenFile(fyle.Filename)
			if fyle.Fd < 0 {
				return false
			}
			fyle.Mode = fs.Mode
			fyle.PreviousFileId = fs.Mtime
		}
		fyle.Idx = 0
		fyle.Len = 0
		fyle.Eof = false
	} else { // otherwise open new file for output
		var related string
		if rfyle != nil {
			related = rfyle.Filename
		}
		if fyle.Filename == "" { // default to related filename
			fyle.Filename = related
		}
		if fyle.Filename == "" {
			return false
		} else if !SysExpandFilename(&fyle.Filename) {
			ScreenMessage(fmt.Sprintf("Error in filename (%s)", fyle.Filename))
			return false
		}
		// if the file given is a directory create a filename using the input
		// filename in the given directory
		fs := SysFileStatus(fyle.Filename)
		if fs.Valid && fs.IsDir && related != "" {
			// Get the filename from the related name
			SysCopyFilename(related, &fyle.Filename)
			fs = SysFileStatus(fyle.Filename)
		}
		if fs.Valid {
			// if we wanted to create a new file then complain
			if fyle.Create {
				ScreenMessage(fmt.Sprintf("File (%s) already exists", fyle.Filename))
				return false
			}
			// check that the file we may overwrite is not a directory
			if fs.IsDir {
				ScreenMessage(fmt.Sprintf("File (%s) is a directory", fyle.Filename))
				return false
			}
			// check that we can write over the current version
			if !SysFileWriteable(fyle.Filename) {
				ScreenMessage(fmt.Sprintf("Write access to file (%s) is denied", fyle.Filename))
				return false
			}
			fyle.Mode = fs.Mode
			fyle.PreviousFileId = fs.Mtime
		} else {
			fyle.Mode = 0666 & SysFileMask()
			fyle.PreviousFileId = 0
		}
		// now create the temporary name
		uniq := 0
		fyle.Tnm = fyle.Filename + "-lw"
		for SysFileExists(fyle.Tnm) {
			uniq++
			fyle.Tnm = fyle.Filename + "-lw" + strconv.Itoa(uniq)
		}
		fyle.Fd = SysCreateFile(fyle.Tnm)
		if fyle.Fd < 0 {
			ScreenMessage(fmt.Sprintf("Error opening (%s) as output", fyle.Tnm))
			return false
		}
	}
	return true
}

// FilesysClose closes a file
// Action is an integer interpreted as follows:
//
//	0 : close
//	1 : close and delete
//	2 : process the file as if closing it, but do not close it
//
// The msgs parameter indicates whether we want to be told about what is going on
func FilesysClose(fyle *FileObject, action int, msgs bool) bool {
	if !fyle.OutputFlag {
		// reap any children
		SysReapChildren()
		// an ordinary input file, just close
		if SysClose(fyle.Fd) < 0 {
			return false
		}
		if msgs {
			plural := "s"
			if fyle.LCounter == 1 {
				plural = ""
			}
			ScreenMessage(fmt.Sprintf("File %s closed (%d line%s read).",
				fyle.Filename, fyle.LCounter, plural))
		}
		return true
	}

	// an output file to close
	if action != 2 && SysClose(fyle.Fd) < 0 {
		return false
	}
	if action == 1 {
		// remove the file
		if SysUnlink(fyle.Tnm) {
			if msgs {
				ScreenMessage(fmt.Sprintf("Output file %s deleted.", fyle.Tnm))
			}
			return true
		} else {
			return false
		}
	}

	// check that another file hasn't been created while we were editing
	// with the name we are going to use as the output name
	fs := SysFileStatus(fyle.Filename)
	if fs.Valid && fs.Mtime != fyle.PreviousFileId {
		ScreenMessage(fmt.Sprintf("%s was modified by another process", fyle.Filename))
	}

	tname := fyle.Filename + "~"
	var maxVnum int64 = 0
	if fyle.Purge {
		versions := SysListBackups(tname)
		if fyle.Versions <= 0 {
			// Just remove all of them in this case
			removeBackupFiles(tname, versions, 0, len(versions))
		} else {
			// Work out how many to retain, and how many to remove
			toRetain := fyle.Versions - 1
			if len(versions) > toRetain {
				toRemove := len(versions) - toRetain
				removeBackupFiles(tname, versions, 0, toRemove)
			}
		}
		if len(versions) > 0 {
			maxVnum = versions[len(versions)-1]
		}
	} else {
		versions := SysListBackups(tname)
		if len(versions) > 0 && len(versions) >= fyle.Versions {
			// Remove just one file
			removeBackupFiles(tname, versions, 0, 1)
		}
		if len(versions) > 0 {
			maxVnum = versions[len(versions)-1]
		}
	}

	// Perform Backup on original file if -
	//    a. fyle->versions != 0
	// or b. doing single backup and backup already done at least once
	if fyle.Versions != 0 || (!fyle.Purge && maxVnum != 0) {
		// does the real file already exist?
		if SysFileExists(fyle.Filename) {
			// try to rename current file to backup
			temp := tname + strconv.FormatInt(maxVnum+1, 10)
			SysRename(fyle.Filename, temp)
		}
	}

	// now rename the temp file to the real thing
	SysChmod(fyle.Tnm, fyle.Mode&0777)
	if !SysRename(fyle.Tnm, fyle.Filename) {
		ScreenMessage(fmt.Sprintf("Cannot rename %s to %s", fyle.Tnm, fyle.Filename))
		return false
	} else {
		if msgs {
			plural := "s"
			if fyle.LCounter == 1 {
				plural = ""
			}
			ScreenMessage(fmt.Sprintf("File %s created (%d line%s written).",
				fyle.Filename, fyle.LCounter, plural))
		}
		// Time to set the memory, if it's required and we aren't writing in
		// one of the global tmp directories
		if !startsWith(fyle.Filename, "/tmp/") &&
			!startsWith(fyle.Filename, "/usr/tmp/") &&
			!startsWith(fyle.Filename, "/var/tmp/") {
			SysWriteFilename(fyle.Memory, fyle.Filename)
		}
		return true
	}
}

// FilesysRead reads a line from a file
// Attempts to read MAX_STRLEN characters into buffer
// Number of characters read is returned in outlen
func FilesysRead(fyle *FileObject, outputBuffer *StrObject, outlen *int) bool {
	*outlen = 0
	for {
		if fyle.Idx >= fyle.Len {
			fyle.Buf = make([]byte, MaxStrLen)
			fyle.Len = int(SysRead(fyle.Fd, fyle.Buf))
			fyle.Idx = 0
		}
		if fyle.Len <= 0 {
			fyle.Eof = true
			// If the last line is not terminated properly,
			// the buffer is not empty and we must return the buffer
			if *outlen > 0 {
				break
			}
			return false
		}
		ch := int(fyle.Buf[fyle.Idx]) & 0x7F // toascii
		fyle.Idx++
		if unicode.IsPrint(rune(ch)) {
			*outlen++
			outputBuffer.Set(*outlen, byte(ch))
		} else if ch == '\t' { // expand the tab
			exp := 8 - (*outlen % 8)
			if *outlen+exp > MaxStrLen {
				exp = MaxStrLen - *outlen
			}
			for ; exp > 0; exp-- {
				*outlen++
				outputBuffer.Set(*outlen, ' ')
			}
		} else if ch == '\n' || ch == '\r' || ch == '\v' || ch == '\f' {
			break // finished if newline or carriage return
		} // forget other control characters
		if *outlen >= MaxStrLen {
			break
		}
	}
	fyle.LCounter++
	return true
}

// FilesysRewind rewinds file described by the fyle pointer
func FilesysRewind(fyle *FileObject) bool {
	if !SysSeek(fyle.Fd, 0) {
		return false
	}
	fyle.Idx = 0
	fyle.Len = 0
	fyle.Eof = false
	fyle.LCounter = 0
	return true
}

// FilesysWrite writes a line to a file
// Attempts to write bufsiz characters from buffer to the file
func FilesysWrite(fyle *FileObject, buffer *StrObject, bufsiz int) bool {
	if bufsiz > 0 {
		offset := 0
		tabs := 0
		if fyle.Entab {
			// FIXME: This is not known to be correct for all cases
			i := 1
			for i <= bufsiz {
				if buffer.Get(i) != ' ' {
					break
				}
				i++
			}
			tabs = i / 8
			offset = tabs * 7
			for i := 1; i <= tabs; i++ {
				buffer.Set(offset+i, '\t')
			}
		}
		count := SysWrite(fyle.Fd, []byte(buffer.Slice(offset+1, bufsiz)))
		if tabs > 0 {
			for i := 1; i <= tabs; i++ {
				buffer.Set(offset+i, ' ')
			}
		}
		if count != int64(bufsiz-offset) {
			return false
		}
	}
	ok := SysWrite(fyle.Fd, []byte(filesysNL)) == filesysNLSize
	fyle.LCounter++
	return ok
}

// FilesysSave implements part of the File Save command
func FilesysSave(iFyle *FileObject, oFyle *FileObject, copyLines int) bool {
	var fyle FileObject

	var inputEof bool
	var inputPosition int64
	var line StrObject
	var lineLen int

	if iFyle != nil {
		// remember things to be restored
		inputEof = iFyle.Eof
		inputPosition = SysTell(oFyle.Fd)

		// copy unread portion of input file to output file
		for {
			if !FilesysRead(iFyle, &line, &lineLen) {
				break
			}
			FilesysWrite(oFyle, &line, lineLen)
			if iFyle.Eof {
				break
			}
		}

		// close input file
		FilesysClose(iFyle, 0, true)
	}

	// rename temporary file to output file
	// process backup options, but do not close the file
	FilesysClose(oFyle, 2, true)

	// make the old output file the new input file
	if iFyle == nil {
		iFyle = &fyle
		iFyle.OutputFlag = false
		iFyle.FirstLine = nil
		iFyle.LastLine = nil
		iFyle.LineCount = 0
	}
	iFyle.Filename = oFyle.Filename
	iFyle.Fd = oFyle.Fd

	// rewind the input file
	FilesysRewind(iFyle)

	// open a new output file
	oFyle.Create = false
	if !FilesysCreateOpen(oFyle, nil, true) {
		return false
	}

	// copy lines from the input file to the output file
	for i := 0; i < copyLines; i++ {
		if !FilesysRead(iFyle, &line, &lineLen) {
			return false
		}
		if !FilesysWrite(oFyle, &line, lineLen) {
			return false
		}
	}

	// reposition or close the input file
	if iFyle == &fyle {
		SysClose(iFyle.Fd)
	} else {
		iFyle.Eof = inputEof
		SysSeek(iFyle.Fd, inputPosition)
		iFyle.Idx = 0
		iFyle.Len = 0
	}
	return true
}

// FilesysParse parses command line arguments for file operations
func FilesysParse(
	commandLine string,
	parseType ParseType,
	fileData *FileDataType,
	input *FileObject,
	output *FileObject,
) bool {
	const usage = "usage : ludwig [-c] [-r] [-i value] [-I] " +
		"[-s value] [-m file] [-M] [-t] [-T] " +
		"[-b value] [-B value] [-o] [-O] [-u] " +
		"[file [file]]"
	const fileUsage = "usage : [-m file] [-t] [-T] [-b value] " +
		"[-B value] [file [file]]"

	if parseType == ParseStdin {
		input.Valid = true
		input.Fd = 0
		input.Eof = false
		input.LCounter = 0
		return true
	}

	// create an argc and argv from the "command_line"
	argv := []string{"Ludwig"}
	cl := toArgv(commandLine)
	argv = append(argv, cl...)

	entab := fileData.Entab
	space := fileData.Space
	purge := fileData.Purge
	versions := fileData.Versions

	createFlag := false
	readOnlyFlag := false
	spaceFlag := false
	usageFlag := false
	versionFlag := false

	errors := 0
	checkInput := false

	var initialize string
	var memory string

	if parseType == ParseCommand {
		var home string
		if !SysGetEnv("HOME", &home) {
			home = "."
		}
		initialize = home + "/.ludwigrc"
		memory = home + "/.lud_memory"
	} else {
		initialize = ""
		memory = ""
	}

	// Parse command line options using a simple parser
	optind := 1
	for optind < len(argv) {
		arg := argv[optind]
		if !strings.HasPrefix(arg, "-") {
			break
		}
		optind++

		for _, c := range arg[1:] {
			optarg := ""
			if optind < len(argv) && !strings.HasPrefix(argv[optind], "-") {
				optarg = argv[optind]
			}

			switch c {
			case 'c':
				if readOnlyFlag {
					errors++
				} else {
					createFlag = true
				}
			case 'r':
				if createFlag {
					errors++
				} else {
					readOnlyFlag = true
				}
			case 'i':
				initialize = optarg
				optind++
			case 'I':
				initialize = ""
			case 's':
				val, err := strconv.Atoi(optarg)
				if err != nil {
					errors++
				} else {
					space = val
					spaceFlag = true
					optind++
				}
			case 'm':
				memory = optarg
				optind++
			case 'M':
				memory = ""
			case 't':
				entab = true
			case 'T':
				entab = false
			case 'b':
				val, err := strconv.Atoi(optarg)
				if err != nil {
					errors++
				} else {
					versions = val
					purge = false
					optind++
				}
			case 'B':
				val, err := strconv.Atoi(optarg)
				if err != nil {
					errors++
				} else {
					versions = val
					purge = true
					optind++
				}
			case 'o':
				versionFlag = true
				fileData.OldCmds = true
			case 'O':
				versionFlag = true
				fileData.OldCmds = false
			case 'u':
				usageFlag = true
			}
		}
	}

	if usageFlag || errors > 0 {
		if parseType == ParseCommand {
			ScreenMessage(usage)
		} else {
			ScreenMessage(fileUsage)
		}
		return false
	}

	if parseType == ParseCommand {
		fileData.Initial = initialize
		fileData.Space = space
		fileData.Entab = entab
		fileData.Purge = purge
		fileData.Versions = versions
	} else if createFlag || readOnlyFlag || initialize != "" || spaceFlag || versionFlag {
		return false
	}

	var file []string
	for filesCount := 0; optind < len(argv); filesCount++ {
		if filesCount >= 2 {
			ScreenMessage("More than two files specified")
			return false
		}
		file = append(file, argv[optind])
		optind++
	}

	if len(file) == 2 {
		checkInput = true
		if parseType == ParseInput || parseType == ParseOutput ||
			parseType == ParseExecute || createFlag || readOnlyFlag {
			ScreenMessage("Only one file name can be specified")
			return false
		}
	}

	switch parseType {
	case ParseCommand, ParseEdit:
		if len(file) > 0 {
			input.Filename = file[0]
		} else if memory != "" {
			if SysReadFilename(memory, &input.Filename) {
				checkInput = true
			} else {
				if parseType == ParseCommand {
					input.Filename = ""
					return true
				}
				ScreenMessage(fmt.Sprintf("Error opening memory file (%s)", memory))
				return false
			}
		} else if parseType == ParseCommand {
			input.Filename = ""
			return true
		} else {
			input.Filename = ""
			return false
		}

		if len(file) > 1 {
			output.Filename = file[1]
		} else {
			output.Filename = input.Filename
		}
		output.Memory = memory
		output.Entab = entab
		output.Purge = purge
		output.Versions = versions

		if readOnlyFlag {
			input.Create = false
			if !FilesysCreateOpen(input, nil, true) {
				ScreenMessage(fmt.Sprintf("Error opening (%s) as input", input.Filename))
				return false
			}
			input.Valid = true
		} else if createFlag {
			output.Create = true
			if !FilesysCreateOpen(output, nil, true) {
				return false
			}
			output.Valid = true
		} else {
			input.Create = false
			output.Create = false
			if FilesysCreateOpen(input, nil, true) {
				input.Valid = true
			} else if checkInput || parseType == ParseEdit {
				ScreenMessage(fmt.Sprintf("Error opening (%s) as input", input.Filename))
				return false
			}
			if FilesysCreateOpen(output, input, true) {
				output.Valid = true
			} else {
				return false
			}
		}

	case ParseInput:
		if len(file) == 1 {
			input.Filename = file[0]
		} else if memory == "" || !SysReadFilename(memory, &input.Filename) {
			input.Filename = ""
			return false
		}
		input.Create = false
		if input.Filename == "" || !FilesysCreateOpen(input, nil, true) {
			ScreenMessage(fmt.Sprintf("Error opening (%s) as input", input.Filename))
			return false
		}
		input.Valid = true

	case ParseExecute:
		if len(file) == 1 {
			input.Filename = file[0]
		} else {
			input.Filename = ""
		}
		input.Create = false
		if input.Filename == "" || !FilesysCreateOpen(input, nil, true) {
			return false
		}
		input.Valid = true

	case ParseOutput:
		if len(file) == 1 {
			output.Filename = file[0]
		} else if input != nil {
			output.Filename = input.Filename
		} else {
			output.Filename = ""
		}
		output.Memory = memory
		output.Entab = entab
		output.Purge = purge
		output.Versions = versions
		output.Create = false
		if output.Filename == "" || !FilesysCreateOpen(output, input, true) {
			return false
		}
		output.Valid = true

	case ParseStdin:
		// Handled at the top of this function
		return false
	}

	return true
}
