package main

import (
	tb "github.com/nsf/termbox-go"
)

type ListItem interface {
	Format(width uint) string
}

type List struct {
	cur   int
	items []ListItem
}

func NewList(items []ListItem) List {
	return List{0, items}
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
	for i, c := range l.items {
		bg := tb.ColorDefault
		if i == l.cur {
			bg = tb.ColorBlack
		}
		if w-x > 0 {
			line := c.Format(uint(w - x))
			if line == "" {
				line = "term too small"
			}
			Prints(x, cy, uint(w-x), tb.ColorDefault, bg, line)
		}
		if cy++; cy > y+h {
			break
		}
	}
	return cy
}
