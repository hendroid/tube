package main

import (
	"io/ioutil"
	"encoding/json"
	"log"
	"reflect"
	"os"
	"os/user"
	"path"
	tb "github.com/nsf/termbox-go"
)

type Config struct {
	APIKey string
	Subscriptions []string
}

var (
	yt Yt
	subs List
)

var config Config
var defaultcfg Config = Config{
	APIKey: "Put your google API key here",
	Subscriptions: []string{"zimbel", "auto"},
}
var running = true

var evhandlers = map[tb.EventType]func(tb.Event){
	tb.EventKey: keydown,
	tb.EventResize: resize,
}

func keydown(ev tb.Event) {
	if ev.Key == tb.KeyEsc {
		running = false
	} else if ev.Key == tb.KeyArrowDown {
		subs.SelectRel(+1)
	} else if ev.Key == tb.KeyArrowUp {
		subs.SelectRel(-1)
	}
	redraw()
}

func resize(ev tb.Event) {
//	tb.Clear(tb.ColorDefault, tb.ColorDefault)
//	s := fmt.Sprintf("%dx%d", ev.Width, ev.Height)
//	Prints(1, 1, tb.ColorDefault, tb.ColorDefault, s)
//	tb.Flush()
	redraw()
}

func redraw() {
	tb.Clear(tb.ColorDefault, tb.ColorDefault)
	w, h := tb.Size()
	Prints(0, 0, w, tb.ColorDefault, tb.ColorDefault, "Hi there")
	subs.Draw(0, 1, w, h-1)
	tb.Flush()
}

//func drawlist(y int, items []Chan) {
//	w, h := tb.Size()
//	wchanmin := len("Channel")
//	wsub := len("Subscribers")
//	wvid := len("Videos") + 3
//	wview := len("Views") + 7
//	wchan := w - 1 - wsub - 1 - wvid - 1 - wview
//	if wchan < wchanmin {
//		Prints(0, 0, w, tb.ColorDefault, tb.ColorDefault, "term to small")
//		return
//	}
//
//	f := fmt.Sprintf("%%-%d.%dv %%%d.%dv %%+%d.%dv %%%d.%dv", wchan, wchan, wsub, wsub, wvid, wvid, wview, wview)
//	line := fmt.Sprintf(f, "Channel", "Subscribers", "Videos", "Views")
//	Prints(0, y, w, tb.ColorDefault, tb.ColorBlack, line)
//	y++
//	for _, c := range items {
//		line = fmt.Sprintf(f, c.Title, strconv.FormatUint(c.SubscriberCount, 10), strconv.FormatUint(c.VideoCount, 10), strconv.FormatUint(c.ViewCount, 10))
//		Prints(0, y, w, tb.ColorDefault, tb.ColorDefault, line)
//		y++
//		if y >= h {
//			break
//		}
//	}
//}

func configsave(filename string, cfg *Config) error{
	j, err := json.MarshalIndent(cfg, "", "\t")
	if err == nil {
		err = ioutil.WriteFile(filename, j, 0600)
		return err
	}
	return err
}

func configload(filename string, cfg *Config) error {
	b, err := ioutil.ReadFile(filename)
	if err == nil {
		err = json.Unmarshal(b, cfg)
		return err
	}
	return err
}

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	fname := path.Join(usr.HomeDir, ".tuberc")
	err = configload(fname, &config)
	if reflect.TypeOf(err) == reflect.TypeOf(new(os.PathError)) {
		if err := configsave(fname, &defaultcfg); err != nil {
			log.Fatal(err)
		}
		config = defaultcfg
	} else if err != nil {
		log.Fatal("could not parse .tuberc:", err)
	}

	yt = NewTube(config.APIKey)
//	fmt.Println(yt.GetChannels(config.Subscriptions))

	if err := tb.Init(); err != nil {
		log.Fatal(err)
	}
	defer tb.Close()
	tb.SetInputMode(tb.InputEsc)

	go func() {subs = NewList(yt.GetChannels(config.Subscriptions))}()
	for running {
//		redraw()
		ev := tb.PollEvent()
		if handler, ok := evhandlers[ev.Type]; ok {
			handler(ev)
		}
	}

	if err := configsave(fname, &config); err != nil {
		log.Fatal(err)
	}
}
