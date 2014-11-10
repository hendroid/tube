package list

import (
	tb "github.com/nsf/termbox-go"
	"sort"
)

type ListItem interface {
	String(width uint) string
}

type List struct {
	Sort    chan<- rune
	NewData chan<- []ListItem
	Draw    chan<- Area
	Event   chan<- Event
	header  ListItem
	cols    Colors
	data    []ListItem
	sel     uint
	// How many lines are displayed above the currently selected one.
	pageOffset uint
	sortFuncs  map[rune]By
}

type Colors struct {
	Fg     tb.Attribute
	Bg     tb.Attribute
	FgHead tb.Attribute
	BgHead tb.Attribute
	FgSel  tb.Attribute
	BgSel  tb.Attribute
}

type Event struct {
	ev            tb.Event
	nextReceivers []chan<- Event
}

type Area struct {
	X      int
	Y      int
	Width  int
	Height int
	Done   chan<- int
}

type By func(i1, i2 *ListItem) bool

type sorter struct {
	it []ListItem
	by By
}

func (s *sorter) Len() int           { return len(s.it) }
func (s *sorter) Swap(i, j int)      { s.it[i], s.it[j] = s.it[j], s.it[i] }
func (s *sorter) Less(i, j int) bool { return s.by(&s.it[i], &s.it[j]) }

func (by By) sort(items []ListItem) {
	is := &sorter{
		it: items,
		by: by,
	}
	sort.Sort(is)
}

func New(header ListItem, colors Colors) (ret *List) {
	cSort := make(chan rune)
	cData := make(chan []ListItem)
	cDraw := make(chan Area)
	cEvent := make(chan Event, 1)
	ret = &List{
		Sort:       cSort,
		NewData:    cData,
		Draw:       cDraw,
		Event:      cEvent,
		header:     header,
		cols:       colors,
		data:       make([]ListItem, 0),
		sel:        0,
		pageOffset: 0,
		sortFuncs:  make(map[rune]By),
	}

	go func() {
		for {
			select {
			case key := <-cSort:
				if col, ok := ret.sortFuncs[key]; ok {
					By(col).sort(ret.data)
				}
			case data := <-cData:
				ret.data = append(ret.data, data...)
			case area := <-cDraw:
				area.Done <- ret.draw(area)
			case ev := <-cEvent:
				if !ret.handle(ev.ev) && len(ev.nextReceivers) > 0 {
					nextReceiver := ev.nextReceivers[0]
					ev.nextReceivers = ev.nextReceivers[1:]
					nextReceiver <- ev
				}
			}
		}
	}()
	return
}

// handle reacts to an event and returns true if the event has been handled.
func (l *List) handle(ev tb.Event) (ret bool) {
	switch ev.Type {
	case tb.EventKey:
		if ev.Ch == 0 {
			switch ev.Key {
			case tb.KeyArrowDown:
				l.sel += 1
				//TODO: redraw!
			case tb.KeyArrowUp:
				l.sel -= 1
				//TODO: redraw!
			default:
				return false
			}
			return true
		} else if key, ret := l.sortFuncs[ev.Ch]; ret {
			//TODO: go routine + redraw
			By(key).sort(l.data)
		}
		return
	case tb.EventMouse:
	}
	return false
}

func printLine(x, y int, w uint, fg, bg tb.Attribute, msg string) {
	lx := 0
	for _, c := range msg {
		if lx >= int(w) {
			return
		}
		tb.SetCell(x+lx, y, c, fg, bg)
		lx++
	}
	for ; lx <= int(w); lx++ {
		tb.SetCell(x+lx, y, ' ', fg, bg)
	}
}

func (l List) draw(a Area) int {
	if !tb.IsInit {
		//TODO: log it?
		return a.Y
	}

	ly := 0

	// print header
	line := l.header.String(uint(a.Width))
	printLine(a.X, a.Y, uint(a.Width), l.cols.FgHead, l.cols.BgHead, line)
	if ly++; ly > a.Height || len(l.data) == 0 {
		return a.Y + ly
	}

	// print entries
	var i = l.sel - l.pageOffset
	for ; i < uint(len(l.data)) && ly < a.Height; i, ly = i+1, ly+1 {
		fg, bg := l.cols.Fg, l.cols.Bg
		if i == l.sel {
			fg, bg = l.cols.FgSel, l.cols.BgSel
		}
		line := l.data[i].String(uint(a.Width))
		printLine(a.X, a.Y+ly, uint(a.Width), fg, bg, line)
	}
	for ; ly < a.Height; ly++ {
		printLine(a.X, a.Y+ly, uint(a.Width), l.cols.Fg, l.cols.Bg, "padding")
	}
	return a.Y + ly
}
