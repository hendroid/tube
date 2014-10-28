package main

import (
	tb "github.com/nsf/termbox-go"
)

type ListItem interface {
	Format(width int) string
}

type List struct {
	cur   int
	items []ListItem
}

func NewList(items []ListItem) List {
	return List{0, items}
}

func Prints(x, y, w int, fg, bg tb.Attribute, msg string) int {
	cx := x
	for _, c := range msg {
		tb.SetCell(cx, y, c, fg, bg)
		if cx++; cx > x+w {
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
		Prints(x, cy, w-x, tb.ColorDefault, bg, c.Format(w-x))
		if cy++; cy > y+h {
			break
		}
	}
	return cy
}
