package main

import (
	"code.google.com/p/google-api-go-client/googleapi/transport"
	"code.google.com/p/google-api-go-client/youtube/v3"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Yt struct {
	svc *youtube.Service
}

type Chan struct {
	Description     string
	Id              string
	SubsHidden      bool
	SubscriberCount uint64
	Title           string
	VideoCount      uint64
	ViewCount       uint64
}

type Vid struct {
	ChannelId     string
	ChannelTitle  string
	CommentCount  uint64
	Description   string
	DislikeCount  uint64
	Duration      time.Duration
	FavoriteCount uint64
	Id            string
	LikeCount     uint64
	PublishedAt   time.Time
	Title         string
	ViewCount     uint64
}

var (
	chancache = make(map[string]interface{})
	vidcache  = make(map[string]interface{})
)

func NewTube(APIKey string) Yt {
	client := &http.Client{Transport: &transport.APIKey{Key: APIKey}}
	s, err := youtube.New(client)
	if err != nil {
		log.Fatal(err)
	}
	return Yt{svc: s}
}

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

func (v Vid) Format(width int) string {
	return "hi there"
}

func retrieveCache(ids []string, cache map[string]interface{}, unmatched *[]string) []ListItem {
	var ret []ListItem
	for _, id := range ids {
		if v, ok := cache[id]; ok {
			ret = append(ret, v.(ListItem))
		} else {
			*unmatched = append(*unmatched, id)
		}
	}
	return ret
}

func (y *Yt) VideosFromChannel(channel string) []ListItem {
	search := y.svc.Search.List("snippet")
	search = search.ChannelId(channel)
	search = search.MaxResults(50)
	search = search.Order("date")
	search = search.Type("video")

	results, err := search.Do()
	if err != nil {
		log.Println("could not retrieve video list:", err)
		return nil
	}

	var toFetch []string
	for _, v := range results.Items {
		toFetch = append(toFetch, v.Id.VideoId)
	}
	return y.GetVideos(toFetch)
}

func (y *Yt) GetVideos(ids []string) []ListItem {
	var toFetch []string
	ret := retrieveCache(ids, vidcache, &toFetch)
	if 0 == len(toFetch) {
		return ret
	}

	search := y.svc.Videos.List("contentDetails,id,snippet,statistics")
	search = search.Id(strings.Join(toFetch, ","))

	results, err := search.Do()
	if err != nil {
		log.Println("could not retrieve channel list:", err)
		return nil
	}

	for _, v := range results.Items {
		dur, err := time.ParseDuration(strings.ToLower(v.ContentDetails.Duration[2:]))
		if err != nil {
			log.Println("problem parsing video duration:", err)
		}
		pubat, err := time.Parse("2006-01-02T15:04:05.000Z", v.Snippet.PublishedAt)
		if err != nil {
			log.Println("problem parsing video publishing date:", err)
		}
		c := Vid{ChannelId: v.Snippet.ChannelId,
			ChannelTitle:  v.Snippet.ChannelTitle,
			CommentCount:  v.Statistics.CommentCount,
			Description:   v.Snippet.Description,
			DislikeCount:  v.Statistics.DislikeCount,
			Duration:      dur,
			FavoriteCount: v.Statistics.FavoriteCount,
			Id:            v.Id,
			LikeCount:     v.Statistics.LikeCount,
			PublishedAt:   pubat,
			Title:         v.Snippet.Title,
			ViewCount:     v.Statistics.ViewCount}
		chancache[v.Id] = c
		ret = append(ret, c)
	}
	return ret
}

func (y *Yt) GetChannels(ids []string) []ListItem {
	var toFetch []string
	ret := retrieveCache(ids, chancache, &toFetch)
	if 0 == len(toFetch) {
		return ret
	}

	search := y.svc.Channels.List("id,snippet,statistics")
	search = search.Id(strings.Join(toFetch, ","))

	results, err := search.Do()
	if err != nil {
		log.Println("could not retrieve channel list:", err)
		return nil
	}

	for _, v := range results.Items {
		c := Chan{Id: v.Id,
			Title:           v.Snippet.Title,
			Description:     v.Snippet.Description,
			SubscriberCount: v.Statistics.SubscriberCount,
			SubsHidden:      v.Statistics.HiddenSubscriberCount,
			VideoCount:      v.Statistics.VideoCount,
			ViewCount:       v.Statistics.ViewCount}
		chancache[v.Id] = c
		ret = append(ret, c)
	}
	return ret
}
