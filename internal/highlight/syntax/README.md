# Syntax Files

The files in this directory have been adopted from the
[micro text editor](https://micro-editor.github.io/).

The micro editor syntax files are published under the following license:

```text
Micro's syntax files are licensed under the MIT "Expat" License:

Copyright (c) 2020: Zachary Yedidia, et al.

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
```

Each yaml file specifies how to detect the filetype based on file extension or
header (first line of the line).  In addition, a signature can be provided to
help resolving ambiguities when multiple matching filetypes are detected.
Then there are patterns and regions linked to highlight groups which tell micro
how to highlight that filetype.

## Ludwig syntax highlighting files

These syntax highlighting files are built into the executable at compile time.
