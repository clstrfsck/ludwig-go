/*-
 * Adapted from FreeBSD getopt.h which is under the following license.
 *
 * Copyright (c) 1987, 1993, 1994
 *	The Regents of the University of California.  All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 * 3. Neither the name of the University nor the names of its contributors
 *    may be used to endorse or promote products derived from this software
 *    without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE REGENTS AND CONTRIBUTORS ``AS IS'' AND
 * ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED.  IN NO EVENT SHALL THE REGENTS OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
 * OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
 * HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
 * LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
 * OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
 * SUCH DAMAGE.
 */

package ludwig

import "strings"

var (
	LwOptInd   int
	LwOptOpt   int
	LwOptReset bool
	LwOptArg   string
	place      string
)

const (
	badCh  = int('?')
	badArg = int(':')
)

func init() {
	LwOptInd = 1
}

func LwGetOpt(nargv []string, ostr string) int {

	if LwOptReset || place == "" {
		LwOptReset = false
		if LwOptInd >= len(nargv) {
			place = ""
			return -1
		}
		place = nargv[LwOptInd]
		if len(place) == 0 || place[0] != '-' {
			place = ""
			return -1
		}
		place = place[1:]
		if len(place) > 0 && place[0] == '-' && len(place) == 1 {
			LwOptInd++
			place = ""
			return -1
		}
		if len(place) == 0 {
			place = ""
			if !strings.ContainsRune(ostr, '-') {
				return -1
			}
			LwOptOpt = '-'
		} else {
			LwOptOpt = int(place[0])
			place = place[1:]
		}
	} else {
		LwOptOpt = int(place[0])
		place = place[1:]
	}

	oli := strings.IndexRune(ostr, rune(LwOptOpt))
	if LwOptOpt == ':' || oli == -1 {
		if len(place) == 0 {
			LwOptInd++
		}
		return badCh
	}

	if oli+1 < len(ostr) && ostr[oli+1] != ':' {
		LwOptArg = ""
		if len(place) == 0 {
			LwOptInd++
		}
	} else {
		if len(place) > 0 {
			LwOptArg = place
		} else if oli+2 < len(ostr) && ostr[oli+2] == ':' {
			LwOptArg = ""
		} else if len(nargv) > LwOptInd+1 {
			LwOptInd++
			LwOptArg = nargv[LwOptInd]
		} else {
			place = ""
			if len(ostr) > 0 && ostr[0] == ':' {
				return badArg
			}
			return badCh
		}
		place = ""
		LwOptInd++
	}
	return LwOptOpt
}
