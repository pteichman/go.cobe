#!/usr/bin/env python

"""ungutenberg converts a Project Gutenberg etext to learnable text"""

import re
import sys

def indexp(predicate, items):
    """Return the first index into items that satisfies predicate(item)."""
    for i, item in enumerate(items):
        if predicate(item):
            return i

    return -1

def main():
    if len(sys.argv) < 2:
        sys.exit("Usage: ungutenberg.py <text file>")

    filename = sys.argv[1]
    with open(filename, "r") as fd:
        text = fd.read()

    # Convert DOS line endings to Unix for convenience, and split all
    # the paragraphs into separate strings.
    text = text.replace("\r\n", "\n")
    paragraphs = [line.strip() for line in text.split("\n\n")]

    start_str = "*** START OF THIS PROJECT GUTENBERG EBOOK"
    end_str = "*** END OF THIS PROJECT GUTENBERG EBOOK"

    start_pos = indexp(lambda s: s.startswith(start_str), paragraphs)
    end_pos = indexp(lambda s: s.startswith(end_str), paragraphs)

    for paragraph in paragraphs[start_pos+1:end_pos]:
        s = paragraph.replace("\n", " ")
        if len(s) > 0:
            print s

if __name__ == "__main__":
    main()
