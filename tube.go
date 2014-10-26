package main

import (
	"fmt"
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

func prints(x, y int, fg, bg tb.Attribute, msg string) int {
	for _, c := range msg {
		tb.SetCell(x, y, c, fg, bg)
		x++
	}
	return x
}

func keydown(ev tb.Event) {
	if ev.Key == tb.KeyEsc {
		running = false
	}
}

func resize(ev tb.Event) {
	tb.Clear(tb.ColorDefault, tb.ColorDefault)
	s := fmt.Sprintf("%dx%d", ev.Width, ev.Height)
	prints(1, 1, tb.ColorDefault, tb.ColorDefault, s)
	tb.Flush()
}

func redraw() {
	tb.Clear(tb.ColorDefault, tb.ColorDefault)
	tb.Flush()
}

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
	fmt.Println(yt.GetChannels(config.Subscriptions))
//	if err := tb.Init(); err != nil {
//		log.Fatal(err)
//	}
//	defer tb.Close()
//	tb.SetInputMode(tb.InputEsc)
//
//	redraw()
//	for running {
//		ev := tb.PollEvent()
//		if handler, ok := evhandlers[ev.Type]; ok {
//			handler(ev)
//		}
//	}

	if err := configsave(fname, &config); err != nil {
		log.Fatal(err)
	}
}
