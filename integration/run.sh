#!/bin/sh
# Integration checks for yup-paste, run inside a Debian (GNU coreutils) container.
#
# parity "<args...>"  — yup-paste reading stdin must match GNU `paste` byte for
#                       byte on a single input stream.
# assert "<want>" ARGS... — yup-paste must produce exactly <want> (for the
#                       file-operand modes, where this gloo paste intentionally
#                       diverges from GNU and has no matching reference).
set -eu

fails=0
sample='alpha
beta
gamma'

# parity feeds the same stdin to yup-paste and GNU paste with identical args and
# asserts byte-identical stdout. Both read the single stdin stream (no operand),
# so the default (parallel) mode is a passthrough that matches GNU.
parity() {
	args=$1
	# shellcheck disable=SC2086
	ours=$(printf '%s\n' "$sample" | yup-paste $args 2>/dev/null || true)
	# shellcheck disable=SC2086
	gnu=$(printf '%s\n' "$sample" | paste $args 2>/dev/null || true)
	if [ "$ours" = "$gnu" ]; then
		printf 'ok    parity  paste %s < stdin\n' "$args"
	else
		printf 'FAIL  parity  paste %s < stdin\n        gnu:  %s\n        ours: %s\n' "$args" "$gnu" "$ours"
		fails=$((fails + 1))
	fi
}

# assert checks yup-paste stdout exactly, for the documented divergences from
# GNU (multi-file parallel concatenation and whole-stream serial joins).
assert() {
	want=$1
	shift
	got=$(yup-paste "$@" 2>/dev/null || true)
	if [ "$got" = "$want" ]; then
		printf 'ok    assert  paste %s\n' "$*"
	else
		printf 'FAIL  assert  paste %s\n        want: %s\n        got:  %s\n' "$*" "$want" "$got"
		fails=$((fails + 1))
	fi
}

# --- parity on a single stdin stream (matches GNU paste exactly) ---
# Default parallel mode: each input line is its own row (passthrough).
parity ''
# Serial mode (-s): collapse the whole stream into one tab-joined row.
parity '-s'
# Serial mode with a custom single-character delimiter.
parity '-s -d ,'
# Serial mode with a multi-character delimiter list, cycled byte by byte.
printf 'w\nx\ny\nz\n' > /tmp/cyc.txt
ours_cyc=$(yup-paste -s -d '-=' < /tmp/cyc.txt 2>/dev/null || true)
gnu_cyc=$(paste -s -d '-=' < /tmp/cyc.txt 2>/dev/null || true)
if [ "$ours_cyc" = "$gnu_cyc" ]; then
	printf 'ok    parity  paste -s -d %s < stdin\n' '-='
else
	printf 'FAIL  parity  paste -s -d %s\n        gnu:  %s\n        ours: %s\n' '-=' "$gnu_cyc" "$ours_cyc"
	fails=$((fails + 1))
fi

# --- documented divergences with file operands (no GNU-equivalent reference) ---
# This gloo paste treats multiple files as one concatenated stream rather than
# side-by-side columns, so file-operand behavior is asserted against its own
# contract, not GNU paste.
printf 'a\nb\nc\n' > /tmp/a.txt
printf '1\n2\n3\n' > /tmp/b.txt
# Default parallel mode over two files: concatenation, not column merge.
# (GNU would emit "a\t1\nb\t2\nc\t3".)
assert 'a
b
c
1
2
3' /tmp/a.txt /tmp/b.txt
# Serial mode over two files: the whole concatenated stream becomes one row.
# (GNU would emit one row per file: "a\tb\tc\n1\t2\t3".)
assert 'a	b	c	1	2	3' -s /tmp/a.txt /tmp/b.txt

if [ "$fails" -ne 0 ]; then
	printf '\n%s check(s) failed\n' "$fails"
	exit 1
fi
printf '\nall checks passed\n'
