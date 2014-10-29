package main

import (
	"fmt"
	tb "github.com/nsf/termbox-go"
	"log"
	"strings"
)

type cachedFmt struct {
	fmt    string
	field  string
	header string
}

type paddable struct {
	entry *cachedFmt
	capt  string
}

type ListItem interface {
	Format(width uint, cache map[uint][]cachedFmt) string
}

type List struct {
	cur         int
	items       []ListItem
	conf        []ConfigColumn
	formatCache map[uint][]cachedFmt
}

func NewList(items []ListItem, cfg []ConfigColumn) (ret *List) {
	ret = new(List)
	ret.cur = 0
	ret.items = items
	ret.conf = cfg
	ret.formatCache = make(map[uint][]cachedFmt)
	return
}

func FormatHeader(width uint, conf []ConfigColumn, cache map[uint][]cachedFmt) string {
	var ret []string
	log.Println(conf, cache)
	for _, t := range getFormats(width, conf, cache) {
		ret = append(ret, fmt.Sprintf(t.fmt, t.header))
	}
	return strings.Join(ret, " ")
}

func Prints(x, y int, w uint, fg, bg tb.Attribute, msg string) int {
	cx := x
	for _, c := range msg {
		tb.SetCell(cx, y, c, fg, bg)
		if cx++; cx > x+int(w) {
			break
		}
	}
	return cx
}

func (l *List) SelectRel(h int) {
	if ret := l.cur + h; ret > len(l.items)-1 {
		l.cur = len(l.items) - 1
	} else if ret < 0 {
		l.cur = 0
	} else {
		l.cur = ret
	}
}

func (l *List) Draw(x, y, w, h int) int {
	cy := y

	// print header
	line := FormatHeader(uint(w), l.conf, l.formatCache)
	if line == "" {
		line = "term too small"
	}
	Prints(x, y, uint(w), tb.ColorBlue, tb.ColorBlack, line)
	if cy++; cy > y+h {
		return cy
	}

	// print entries
	for i, c := range l.items {
		bg := tb.ColorDefault
		if i == l.cur {
			bg = tb.ColorBlack
		}
		if w > 0 {
			line := c.Format(uint(w), l.formatCache)
			if line == "" {
				line = "term too small"
			}
			Prints(x, cy, uint(w), tb.ColorDefault, bg, line)
		}
		if cy++; cy > y+h {
			break
		}
	}
	return cy
}
