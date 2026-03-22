/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:        SYS_LINUX
//
// Description: Implementation of O/S support routines for go.

package ludwig

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"
)

const (
	nl = "\n"
)

// FileStatus represents file status information
type FileStatus struct {
	Valid bool
	Mode  int
	Mtime int64
	IsDir bool
}

// SysSuspend sends SIGTSTP to suspend the process
func SysSuspend() bool {
	pid := os.Getpid()
	return syscall.Kill(pid, syscall.SIGTSTP) == nil
}

// SysShell launches a shell
func SysShell() bool {
	// FIXME: Should really make this work
	return false
}

// SysIsTTY checks if stdin and stdout are connected to a terminal
func SysIsTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

// SysGetEnv retrieves an environment variable
func SysGetEnv(environ string) (string, bool) {
	return os.LookupEnv(environ)
}

// SysInitSig initializes signal handlers
func SysInitSig() {
	// Nothing yet
}

// SysExitSuccess exits with success status
func SysExitSuccess() {
	os.Exit(0)
}

// SysExitFailure exits with failure status
func SysExitFailure() {
	os.Exit(1)
}

// SysExpandFilename expands ~ in filenames to user home directories
func SysExpandFilename(filename *string) bool {
	if *filename == "" {
		return true
	}
	if strings.HasPrefix(*filename, "~") {
		userPart, rest, _ := strings.Cut((*filename)[1:], "/")
		var dir string
		if userPart == "" {
			var err error
			dir, err = os.UserHomeDir()
			if err != nil {
				return false
			}
		} else {
			u, err := user.Lookup(userPart)
			if err != nil {
				return false
			}
			dir = u.HomeDir
		}
		*filename = filepath.Join(dir, rest)
	}
	absPath, err := filepath.Abs(*filename)
	if err != nil {
		return false
	}
	*filename = absPath
	return *filename != ""
}

// SysCopyFilename extracts the filename from src_path and appends to dst_path
func SysCopyFilename(srcPath string, dstPath *string) bool {
	*dstPath = filepath.Join(*dstPath, filepath.Base(srcPath))
	return true
}

// SysOpenCommand opens a pipe to a command
func SysOpenCommand(cmd string) *os.File {
	command := exec.Command("sh", "-c", cmd)

	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil
	}

	command.Stderr = command.Stdout

	if err := command.Start(); err != nil {
		return nil
	}

	// Convert the pipe to a file descriptor
	if f, ok := stdout.(*os.File); ok {
		return f
	}

	return nil
}

// SysFileMask returns the current file creation mask
func SysFileMask() int {
	oldMask := syscall.Umask(0)
	syscall.Umask(oldMask)
	return ^oldMask
}

// SysFileExists checks if a file exists
func SysFileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// SysFileWriteable checks if a file is writeable
func SysFileWriteable(filename string) bool {
	f, err := os.OpenFile(filename, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

// SysWriteFilename writes a filename to a memory file
func SysWriteFilename(memory string, filename string) bool {
	if memory != "" {
		f, err := os.OpenFile(memory, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return false
		}
		defer f.Close()

		_, err1 := f.WriteString(filename)
		_, err2 := f.WriteString(nl)
		return err1 == nil && err2 == nil
	}
	return false
}

// SysReadFilename reads a filename from a memory file
func SysReadFilename(memory string) (string, bool) {
	f, err := os.Open(memory)
	if err != nil {
		return "", false
	}
	defer f.Close()

	line, err := bufio.NewReader(f).ReadString('\n')
	if err != nil && err != io.EOF {
		return "", false
	}
	filename := strings.TrimRight(line, "\r\n")
	return filename, filename != ""
}

// SysReapChildren reaps zombie child processes
func SysReapChildren() {
	for {
		var status syscall.WaitStatus
		pid, err := syscall.Wait4(-1, &status, syscall.WNOHANG, nil)
		if err != nil || pid <= 0 {
			break
		}
	}
}

// SysChmod changes file permissions
func SysChmod(filename string, mask int) bool {
	return os.Chmod(filename, os.FileMode(mask)) == nil
}

// SysSeek seeks to a position in a file
func SysSeek(osFile *os.File, where int64) bool {
	_, err := osFile.Seek(where, 0) // SEEK_SET = 0
	return err == nil
}

// SysTell returns the current position in a file
func SysTell(osFile *os.File) int64 {
	pos, err := osFile.Seek(0, 1) // SEEK_CUR = 1
	if err != nil {
		return -1
	}
	return pos
}

// SysFileStatus returns file status information
func SysFileStatus(filename string) FileStatus {
	fs := FileStatus{
		Valid: false,
		Mode:  0600,
		Mtime: -1,
		IsDir: false,
	}

	info, err := os.Stat(filename)
	if err != nil {
		return fs
	}

	fs.Valid = true
	fs.Mode = int(info.Mode() & 0777)
	fs.Mtime = info.ModTime().Unix()
	fs.IsDir = info.IsDir()
	return fs
}

// SysListBackups lists backup versions of a file
func SysListBackups(backupName string) []int64 {
	dir := filepath.Dir(backupName)
	if dir == "" {
		dir = "."
	}
	bname := filepath.Base(backupName)

	var versions []int64

	entries, err := os.ReadDir(dir)
	if err != nil {
		return versions
	}

	for _, entry := range entries {
		name := entry.Name()
		if len(name) < len(bname) {
			continue
		}
		if !strings.HasPrefix(name, bname) {
			continue
		}

		suffix := name[len(bname):]
		if suffix == "" {
			continue
		}

		v, err := strconv.ParseInt(suffix, 10, 64)
		if err != nil {
			continue
		}
		versions = append(versions, v)
	}

	slices.Sort(versions)

	return versions
}
