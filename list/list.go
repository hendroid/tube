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
	data    []ListItem
	sel     uint
	// How many lines are displayed above the currently selected one.
	pageOffset uint
	sortFuncs  map[rune]By
}

type Area struct {
	X      int
	Y      int
	Width  int
	Height int
	Done   chan<- int
}

type Event struct {
	ev            tb.Event
	nextReceivers []chan<- Event
}

type By func(i1, i2 *ListItem) bool

type sorter struct {
	items []ListItem
	by    By
}

func (s *sorter) Len() int           { return len(s.items) }
func (s *sorter) Swap(i, j int)      { s.items[i], s.items[j] = s.items[j], s.items[i] }
func (s *sorter) Less(i, j int) bool { return s.by(&s.items[i], &s.items[j]) }

func (by By) sort(items []ListItem) {
	is := &sorter{
		items: items,
		by:    by,
	}
	sort.Sort(is)
}

func New(header ListItem) (ret *List) {
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
	if line == "" {
		line = "term too small"
	}
	printLine(a.X, a.Y, uint(a.Width), tb.ColorDefault, tb.ColorDefault, line)
	if ly++; ly > a.Height || len(l.data) == 0 {
		return a.Y + ly
	}

	// print entries
	var i = l.sel - l.pageOffset
	for ; i < uint(len(l.data)) && ly < a.Height; i, ly = i+1, ly+1 {
		bg := tb.ColorDefault
		if i == l.sel {
			bg = tb.ColorBlack
		}
		line := l.data[i].String(uint(a.Width))
		if line == "" {
			line = "term too small"
		}
		printLine(a.X, a.Y+ly, uint(a.Width), tb.ColorDefault, bg, line)
	}
	for ; ly < a.Height; ly++ {
		printLine(a.X, a.Y+ly, uint(a.Width), tb.ColorDefault, tb.ColorDefault, "padding")
	}
	return a.Y + ly
}
