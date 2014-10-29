package main

import (
	"encoding/json"
	tb "github.com/nsf/termbox-go"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"reflect"
)

type ConfigColumn struct {
	HeaderCaption string
	Pad           string
	FieldName     string
	Priority      uint
}

type Config struct {
	APIKey           string
	Subscriptions    []string
	VideoListColumns []ConfigColumn
	ListChannels     []ConfigColumn
	ListPlaylists    []ConfigColumn
}

var (
	config   Config
	yt       Yt
	subs     List
	vids     List
	listVids bool = false
	running       = true
)

var defaultcfg Config = Config{
	APIKey:        "Put your google API key here",
	Subscriptions: []string{"UC3XTzVzaHQEd30rQbuvCtTQ", "UC-lHJZR3Gqxm24_Vd_AJ5Yw"},
	VideoListColumns: []ConfigColumn{
		{HeaderCaption: " Published",
			Pad:       "",
			FieldName: "PublishedAt",
			Priority:  8},
		{HeaderCaption: "     Views",
			Pad:       "",
			FieldName: "Views",
			Priority:  6},
		{HeaderCaption: "Like%",
			Pad:       "",
			FieldName: "LikePercentage",
			Priority:  4},
		{HeaderCaption: " Duration",
			Pad:       "",
			FieldName: "Duration",
			Priority:  10},
		{HeaderCaption: "Title     ",
			Pad:       "right",
			FieldName: "Title",
			Priority:  10},
		{HeaderCaption: "      User",
			Pad:       "left",
			FieldName: "ChannelTitle",
			Priority:  2},
	},
}

var evhandlers = map[tb.EventType]func(tb.Event){
	tb.EventKey:    keydown,
	tb.EventResize: resize,
}

func keydown(ev tb.Event) {
	if ev.Key == tb.KeyEsc {
		running = false
	} else if ev.Key == tb.KeyArrowDown {
		if listVids {
			vids.SelectRel(+1)
		} else {
			subs.SelectRel(+1)
		}
	} else if ev.Key == tb.KeyArrowUp {
		if listVids {
			vids.SelectRel(-1)
		} else {
			subs.SelectRel(-1)
		}
	} else if ev.Key == tb.KeyArrowRight {
		go func() { vids = NewList(yt.VideosFromChannel(subs.items[subs.cur].(Chan).Id)); listVids = true }()
	} else if ev.Key == tb.KeyArrowLeft {
		listVids = false
	}
	redraw()
}

func resize(ev tb.Event) {
	redraw()
}

func redraw() {
	tb.Clear(tb.ColorDefault, tb.ColorDefault)
	w, h := tb.Size()
	Prints(0, 0, w, tb.ColorDefault, tb.ColorDefault, "Hi there")
	if listVids {
		vids.Draw(0, 1, w, h-1)
	} else {
		subs.Draw(0, 1, w, h-1)
	}
	tb.Flush()
}

func configsave(filename string, cfg *Config) error {
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

	go func() { subs = NewList(yt.GetChannels(config.Subscriptions)) }()
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
