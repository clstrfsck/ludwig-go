package highlight

import (
	"regexp"
	"testing"
)

func TestMakeHeader(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name: "valid header",
			input: []byte(`go
\.go$
^package\s

`),
			wantErr: false,
		},
		{
			name: "valid header with all fields",
			input: []byte(`python
\.py$
^#!/usr/bin/python
import\s+\w+`),
			wantErr: false,
		},
		{
			name: "valid header with empty regexes",
			input: []byte(`text



`),
			wantErr: false,
		},
		{
			name:    "invalid - too few lines",
			input:   []byte("go\n\\.go$\n"),
			wantErr: true,
		},
		{
			name: "invalid filename regex",
			input: []byte(`go
[invalid(regex



`),
			wantErr: true,
		},
		{
			name: "invalid header regex",
			input: []byte(`go

[invalid(regex

`),
			wantErr: true,
		},
		{
			name: "invalid signature regex",
			input: []byte(`go


[invalid(regex`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header, err := MakeHeader(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && header == nil {
				t.Errorf("MakeHeader() returned nil header without error")
			}
		})
	}
}

func TestMakeHeaderYaml(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
		check   func(*Header) error
	}{
		{
			name: "valid yaml header",
			input: []byte(`filetype: go
detect:
    filename: "\\.go$"
    header: "^package\\s"
`),
			wantErr: false,
		},
		{
			name: "valid yaml with signature",
			input: []byte(`filetype: python
detect:
    filename: "\\.py$"
    header: "^#!/usr/bin/python"
    signature: "import\\s+\\w+"
`),
			wantErr: false,
		},
		{
			name: "valid yaml with empty detect",
			input: []byte(`filetype: text
detect:
    filename: ""
    header: ""
`),
			wantErr: false,
		},
		{
			name:    "invalid yaml",
			input:   []byte(`invalid: yaml: structure:`),
			wantErr: true,
		},
		{
			name: "invalid filename regex in yaml",
			input: []byte(`filetype: test
detect:
    filename: "[invalid(regex"
`),
			wantErr: true,
		},
		{
			name: "invalid header regex in yaml",
			input: []byte(`filetype: test
detect:
    filename: "\.test$"
    header: "[invalid(regex"
`),
			wantErr: true,
		},
		{
			name: "invalid signature regex in yaml",
			input: []byte(`filetype: test
detect:
    filename: "\.test$"
    header: ""
    signature: "[invalid(regex"
`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header, err := MakeHeaderYaml(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeHeaderYaml() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if header == nil {
					t.Errorf("MakeHeaderYaml() returned nil header without error")
				}
				if tt.check != nil {
					if err := tt.check(header); err != nil {
						t.Errorf("MakeHeaderYaml() check failed: %v", err)
					}
				}
			}
		})
	}
}

func TestHeaderMatchFileName(t *testing.T) {
	tests := []struct {
		name     string
		regex    string
		filename string
		want     bool
	}{
		{
			name:     "go file matches",
			regex:    `\.go$`,
			filename: "main.go",
			want:     true,
		},
		{
			name:     "go file doesn't match python regex",
			regex:    `\.py$`,
			filename: "main.go",
			want:     false,
		},
		{
			name:     "no regex",
			regex:    "",
			filename: "test.txt",
			want:     false,
		},
		{
			name:     "complex regex matches",
			regex:    `^Makefile$|\.mk$`,
			filename: "Makefile",
			want:     true,
		},
		{
			name:     "complex regex matches extension",
			regex:    `^Makefile$|\.mk$`,
			filename: "build.mk",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := &Header{}
			if tt.regex != "" {
				header.FileNameRegex = regexp.MustCompile(tt.regex)
			}
			if got := header.MatchFileName(tt.filename); got != tt.want {
				t.Errorf("Header.MatchFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHeaderMatchFileHeader(t *testing.T) {
	tests := []struct {
		name      string
		regex     string
		firstLine []byte
		want      bool
	}{
		{
			name:      "shell script shebang matches",
			regex:     `^#!/bin/(ba)?sh`,
			firstLine: []byte("#!/bin/bash"),
			want:      true,
		},
		{
			name:      "python shebang matches",
			regex:     `^#!/usr/bin/python`,
			firstLine: []byte("#!/usr/bin/python3"),
			want:      true,
		},
		{
			name:      "no match",
			regex:     `^package\s`,
			firstLine: []byte("import fmt"),
			want:      false,
		},
		{
			name:      "no regex",
			regex:     "",
			firstLine: []byte("anything"),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := &Header{}
			if tt.regex != "" {
				header.HeaderRegex = regexp.MustCompile(tt.regex)
			}
			if got := header.MatchFileHeader(tt.firstLine); got != tt.want {
				t.Errorf("Header.MatchFileHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHeaderHasFileSignature(t *testing.T) {
	tests := []struct {
		name  string
		regex string
		want  bool
	}{
		{
			name:  "has signature",
			regex: `import\s+\w+`,
			want:  true,
		},
		{
			name:  "no signature",
			regex: "",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := &Header{}
			if tt.regex != "" {
				header.SignatureRegex = regexp.MustCompile(tt.regex)
			}
			if got := header.HasFileSignature(); got != tt.want {
				t.Errorf("Header.HasFileSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHeaderMatchFileSignature(t *testing.T) {
	tests := []struct {
		name  string
		regex string
		line  []byte
		want  bool
	}{
		{
			name:  "import statement matches",
			regex: `import\s+\w+`,
			line:  []byte("import sys"),
			want:  true,
		},
		{
			name:  "no match",
			regex: `import\s+\w+`,
			line:  []byte("print('hello')"),
			want:  false,
		},
		{
			name:  "no regex",
			regex: "",
			line:  []byte("anything"),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := &Header{}
			if tt.regex != "" {
				header.SignatureRegex = regexp.MustCompile(tt.regex)
			}
			if got := header.MatchFileSignature(tt.line); got != tt.want {
				t.Errorf("Header.MatchFileSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	tests := []struct {
		name         string
		input        []byte
		wantErr      bool
		wantFileType string
	}{
		{
			name: "valid simple file",
			input: []byte(`filetype: test
rules:
    - comment: "#.*$"
`),
			wantErr:      false,
			wantFileType: "test",
		},
		{
			name: "valid file with detect",
			input: []byte(`filetype: go
detect:
    filename: "\\.go$"
rules:
    - keyword: "\\b(func|package|import)\\b"
`),
			wantErr:      false,
			wantFileType: "go",
		},
		{
			name:    "missing filetype",
			input:   []byte(`rules: []`),
			wantErr: true,
		},
		{
			name: "empty filetype",
			input: []byte(`filetype: ""
rules: []`),
			wantErr: true,
		},
		{
			name:    "invalid yaml",
			input:   []byte(`invalid: yaml: structure:`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := ParseFile(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if file == nil {
					t.Errorf("ParseFile() returned nil file without error")
					return
				}
				if file.FileType != tt.wantFileType {
					t.Errorf("ParseFile() FileType = %v, want %v", file.FileType, tt.wantFileType)
				}
			}
		})
	}
}

func TestParseDef(t *testing.T) {
	setupTest := func() {
		Groups = make(map[string]Group)
		numGroups = 0
	}

	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name: "simple pattern rule",
			input: []byte(`filetype: test
rules:
    - comment: "#.*$"
`),
			wantErr: false,
		},
		{
			name: "multiple patterns",
			input: []byte(`filetype: test
rules:
    - keyword: "\\bif\\b"
    - keyword: "\\belse\\b"
    - comment: "#.*$"
`),
			wantErr: false,
		},
		{
			name: "region rule",
			input: []byte(`filetype: test
rules:
    - string:
        start: "\""
        end: "\""
        skip: "\\\\."
`),
			wantErr: false,
		},
		{
			name: "region with nested rules",
			input: []byte(`filetype: test
rules:
    - comment:
        start: "/\\*"
        end: "\\*/"
        rules:
            - todo: "(TODO|FIXME):?"
`),
			wantErr: false,
		},
		{
			name: "include rule",
			input: []byte(`filetype: test
rules:
    - include: "other"
`),
			wantErr: false,
		},
		{
			name: "empty rules",
			input: []byte(`filetype: test
rules: []
`),
			wantErr: false,
		},
		{
			name:    "no rules section is ok",
			input:   []byte(`filetype: test`),
			wantErr: false,
		},
		{
			name: "region missing start",
			input: []byte(`filetype: test
rules:
    - string:
        end: "\""
`),
			wantErr: true,
		},
		{
			name: "region missing end",
			input: []byte(`filetype: test
rules:
    - string:
        start: "\""
`),
			wantErr: true,
		},
		{
			name: "empty pattern",
			input: []byte(`filetype: test
rules:
    - comment: ""
`),
			wantErr: true,
		},
		{
			name: "empty start",
			input: []byte(`filetype: test
rules:
    - string:
        start: ""
        end: "\""
`),
			wantErr: true,
		},
		{
			name: "empty end",
			input: []byte(`filetype: test
rules:
    - string:
        start: "\""
        end: ""
`),
			wantErr: true,
		},
		{
			name: "empty skip",
			input: []byte(`filetype: test
rules:
    - string:
        start: "\""
        end: "\""
        skip: ""
`),
			wantErr: true,
		},
		{
			name: "invalid regex in pattern",
			input: []byte(`filetype: test
rules:
    - keyword: "[invalid(regex"
`),
			wantErr: true,
		},
		{
			name: "invalid regex in start",
			input: []byte(`filetype: test
rules:
    - string:
        start: "[invalid(regex"
        end: "\""
`),
			wantErr: true,
		},
		{
			name: "invalid regex in end",
			input: []byte(`filetype: test
rules:
    - string:
        start: "\""
        end: "[invalid(regex"
`),
			wantErr: true,
		},
		{
			name: "invalid regex in skip",
			input: []byte(`filetype: test
rules:
    - string:
        start: "\""
        end: "\""
        skip: "[invalid(regex"
`),
			wantErr: true,
		},
		{
			name: "region with limit-group",
			input: []byte(`filetype: test
rules:
    - string:
        start: "\""
        end: "\""
        limit-group: "string.delimiter"
`),
			wantErr: false,
		},
		{
			name: "empty limit-group",
			input: []byte(`filetype: test
rules:
    - string:
        start: "\""
        end: "\""
        limit-group: ""
`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest()

			file, err := ParseFile(tt.input)
			if err != nil {
				t.Fatalf("ParseFile() failed: %v", err)
			}

			header, err := MakeHeaderYaml(tt.input)
			if err != nil && !tt.wantErr {
				t.Fatalf("MakeHeaderYaml() failed: %v", err)
			}

			def, err := ParseDef(file, header)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && def == nil {
				t.Errorf("ParseDef() returned nil def without error")
			}
		})
	}
}

func TestGroupString(t *testing.T) {
	Groups = make(map[string]Group)
	numGroups = 0

	Groups["keyword"] = 1
	Groups["comment"] = 2
	Groups["string"] = 3

	tests := []struct {
		name  string
		group Group
		want  string
	}{
		{
			name:  "keyword group",
			group: 1,
			want:  "keyword",
		},
		{
			name:  "comment group",
			group: 2,
			want:  "comment",
		},
		{
			name:  "string group",
			group: 3,
			want:  "string",
		},
		{
			name:  "non-existent group",
			group: 99,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.group.String(); got != tt.want {
				t.Errorf("Group.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasIncludes(t *testing.T) {
	setupTest := func() {
		Groups = make(map[string]Group)
		numGroups = 0
	}

	tests := []struct {
		name  string
		input []byte
		want  bool
	}{
		{
			name: "no includes",
			input: []byte(`filetype: test
rules:
    - comment: "#.*$"
`),
			want: false,
		},
		{
			name: "has include",
			input: []byte(`filetype: test
rules:
    - include: "other"
`),
			want: true,
		},
		{
			name: "include in region",
			input: []byte(`filetype: test
rules:
    - comment:
        start: "/\\*"
        end: "\\*/"
        rules:
            - include: "other"
`),
			want: true,
		},
		{
			name: "multiple includes",
			input: []byte(`filetype: test
rules:
    - include: "one"
    - include: "two"
`),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest()

			file, err := ParseFile(tt.input)
			if err != nil {
				t.Fatalf("ParseFile() failed: %v", err)
			}

			header, err := MakeHeaderYaml(tt.input)
			if err != nil {
				t.Fatalf("MakeHeaderYaml() failed: %v", err)
			}

			def, err := ParseDef(file, header)
			if err != nil {
				t.Fatalf("ParseDef() failed: %v", err)
			}

			if got := HasIncludes(def); got != tt.want {
				t.Errorf("HasIncludes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetIncludes(t *testing.T) {
	setupTest := func() {
		Groups = make(map[string]Group)
		numGroups = 0
	}

	tests := []struct {
		name  string
		input []byte
		want  []string
	}{
		{
			name: "no includes",
			input: []byte(`filetype: test
rules:
    - comment: "#.*$"
`),
			want: []string{},
		},
		{
			name: "single include",
			input: []byte(`filetype: test
rules:
    - include: "other"
`),
			want: []string{"other"},
		},
		{
			name: "multiple includes",
			input: []byte(`filetype: test
rules:
    - include: "one"
    - include: "two"
`),
			want: []string{"one", "two"},
		},
		{
			name: "include in region",
			input: []byte(`filetype: test
rules:
    - comment:
        start: "/\\*"
        end: "\\*/"
        rules:
            - include: "nested"
`),
			want: []string{"nested"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest()

			file, err := ParseFile(tt.input)
			if err != nil {
				t.Fatalf("ParseFile() failed: %v", err)
			}

			header, err := MakeHeaderYaml(tt.input)
			if err != nil {
				t.Fatalf("MakeHeaderYaml() failed: %v", err)
			}

			def, err := ParseDef(file, header)
			if err != nil {
				t.Fatalf("ParseDef() failed: %v", err)
			}

			got := GetIncludes(def)
			if len(got) != len(tt.want) {
				t.Errorf("GetIncludes() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("GetIncludes()[%d] = %v, want %v", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestResolveIncludes(t *testing.T) {
	setupTest := func() {
		Groups = make(map[string]Group)
		numGroups = 0
	}

	t.Run("resolve single include", func(t *testing.T) {
		setupTest()

		baseInput := []byte(`filetype: base
rules:
    - comment: "#.*$"
`)
		baseFile, err := ParseFile(baseInput)
		if err != nil {
			t.Fatalf("ParseFile(base) failed: %v", err)
		}

		mainInput := []byte(`filetype: main
rules:
    - include: "base"
    - keyword: "\\btest\\b"
`)
		mainFile, err := ParseFile(mainInput)
		if err != nil {
			t.Fatalf("ParseFile(main) failed: %v", err)
		}

		mainHeader, _ := MakeHeaderYaml(mainInput)

		mainDef, err := ParseDef(mainFile, mainHeader)
		if err != nil {
			t.Fatalf("ParseDef(main) failed: %v", err)
		}

		files := []*File{baseFile}
		ResolveIncludes(mainDef, files)

		if len(mainDef.rules.patterns) < 2 {
			t.Errorf("ResolveIncludes() did not add patterns from base, got %d patterns", len(mainDef.rules.patterns))
		}

		hasComment := false
		hasKeyword := false
		for _, p := range mainDef.rules.patterns {
			if p.group.String() == "comment" {
				hasComment = true
			}
			if p.group.String() == "keyword" {
				hasKeyword = true
			}
		}
		if !hasComment {
			t.Errorf("ResolveIncludes() did not include comment pattern from base")
		}
		if !hasKeyword {
			t.Errorf("ResolveIncludes() lost keyword pattern from main")
		}
	})
}
