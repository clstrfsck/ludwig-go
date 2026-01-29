/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         HELP
//
// Description:  Ludwig HELP facility.

package ludwig

import (
	"strings"
)

const (
	helpIndex = "0" // The index key
)

// toUpper converts a string to uppercase
func toUpper(s string) string {
	return strings.ToUpper(s)
}

// askUser prompts the user for input and returns the reply in uppercase
func askUser(prompt string) string {
	ScreenWriteln()
	reply := ScreenHelpPrompt(prompt)
	// Note that all characters not overwritten by the user will be spaces!
	ScreenWriteln()
	reply = toUpper(reply)
	return reply
}

// HelpHelp displays the help system
// The argument selects a particular part of the help file to read e.g. SD
func HelpHelp(selection string) {
	ScreenUnload()
	ScreenHome(true)

	var topic string
	if selection == "" {
		topic = helpIndex
	} else {
		topic = selection
	}

	if !HelpfileOpen(FileData.OldCmds) {
		ScreenWriteStr(3, "Can't open HELP file")
		ScreenWriteln()
		ScreenPause()
		topic = ""
	}

	for topic != "" {
		ScreenHome(true)
		var buf HelpRecord
		continu := HelpfileRead(topic, &buf)
		if !continu {
			ScreenWriteStr(3, "Can't find Command or Section in HELP")
			ScreenWriteln()
			topic = ""
		}

		for continu {
			if len(buf.Txt) >= 2 && buf.Txt[0] == '\\' && buf.Txt[1] == '%' {
				reply := askUser("<space> for more, <return> to exit : ")
				if TtControlC {
					reply = ""
				}
				if reply == "" || reply[0] != ' ' {
					continu = false
					topic = reply
				}
			} else {
				ScreenWriteStr(2, buf.Txt)
				ScreenWriteln()
			}

			if continu {
				continu = HelpfileNext(&buf) && buf.Key == topic
				if !continu {
					topic = ""
				}
			}

			if TtControlC {
				continu = false
			}
		}

		if topic == "" || (len(topic) > 0 && topic[0] == ' ') {
			topic = askUser("Command or Section or <return> to exit : ")
			if topic != "" && topic[0] == ' ' {
				topic = helpIndex
			}
			if TtControlC {
				topic = ""
			}
		}
	}
}
