package main

import (
	"code.google.com/p/google-api-go-client/googleapi/transport"
	"code.google.com/p/google-api-go-client/youtube/v3"
	"log"
	"fmt"
	"strconv"
	"net/http"
	"strings"
)

type Yt struct {
	svc *youtube.Service
}

type Chan struct {
	Id string
	Title string
	Description string
	SubscriberCount uint64
	SubsHidden bool
	VideoCount uint64
	ViewCount uint64
}

func NewTube(APIKey string) Yt {
	client := &http.Client{Transport: &transport.APIKey{Key: APIKey}}
	s, err := youtube.New(client)
	if err != nil {
		log.Fatal(err)
	}
	return Yt{svc: s}
}

//	search := svc.Search.List("id,snippet").MaxResults(5).Q("auto")
//
//	results, err := search.Do()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	for res := range results.Items {
//		_ = res
////		(results.Items[res].Id.VideoId, results.Items[res].Snippet)
//	}

func (c Chan) Format(width int) string {
	wchanmin := len("Channel")
	wsub := len("Subscribers")
	wvid := len("Videos") + 3
	wview := len("Views") + 7
	wchan := width - 1 - wsub - 1 - wvid - 1 - wview
	if wchan < wchanmin {
		return "term too small"
	}

	f := fmt.Sprintf("%%-%d.%dv %%%d.%dv %%+%d.%dv %%%d.%dv", wchan, wchan,
	wsub, wsub, wvid, wvid, wview, wview)
//	line := fmt.Sprintf(f, "Channel", "Subscribers", "Videos", "Views")
//	Prints(0, y, width, tb.ColorDefault, tb.ColorBlack, line)
	return fmt.Sprintf(f, c.Title, strconv.FormatUint(c.SubscriberCount, 10),
	strconv.FormatUint(c.VideoCount, 10), strconv.FormatUint(c.ViewCount, 10))
}

func (y Yt) GetChannels(ids []string) []ListItem {
	ret := make([]ListItem, 0)
	search := y.svc.Channels.List("id,snippet,statistics")
	search = search.Id(strings.Join(ids, ","))

	results, err := search.Do()
	if err != nil {
		log.Println("could not retrieve channel list:", err)
		return nil
	}

	for _, v := range results.Items {
		ret = append(ret, Chan{Id: v.Id,
		                       Title: v.Snippet.Title,
		                       Description: v.Snippet.Description,
		                       SubscriberCount: v.Statistics.SubscriberCount,
		                       SubsHidden: v.Statistics.HiddenSubscriberCount,
		                       VideoCount: v.Statistics.VideoCount,
		                       ViewCount: v.Statistics.ViewCount})
	}
	return ret
}
