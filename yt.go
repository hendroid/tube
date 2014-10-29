package main

import (
	"code.google.com/p/google-api-go-client/googleapi/transport"
	"code.google.com/p/google-api-go-client/youtube/v3"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Yt struct {
	svc *youtube.Service
}

type Chan struct {
	Description     string // NOT DISPLAYED
	Id              string // NOT DISPLAYED
	SubsHidden      bool   // NOT DISPLAYED
	SubscriberCount uint64 // 12 Subscribers
	Title           string // 10+ Title
	VideoCount      uint64 // 9 Videos
	ViewCount       uint64 // 12 Views
}

type Vid struct {
	ChannelId     string        // NOT DISPLAYED
	ChannelTitle  string        // +10 User
	CommentCount  uint64        // NOT DISPLAYED
	Description   string        // NOT DISPLAYED
	DislikeCount  uint64        // 5 Like%
	Duration      time.Duration // 9 Duration
	FavoriteCount uint64        // NOT DISPLAYED
	Id            string        // NOT DISPLAYED
	LikeCount     uint64        // 5 Like%
	PublishedAt   time.Time     // 10 Published
	Title         string        // 10+ Title
	ViewCount     uint64        // 10 Views
}

var (
	chanCache = make(map[string]interface{})
	vidCache  = make(map[string]interface{})
)

type byPrio []ConfigColumn

func (c byPrio) Len() int           { return len(c) }
func (c byPrio) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c byPrio) Less(i, j int) bool { return c[i].Priority > c[j].Priority }

func NewTube(APIKey string) Yt {
	client := &http.Client{Transport: &transport.APIKey{Key: APIKey}}
	s, err := youtube.New(client)
	if err != nil {
		log.Fatal(err)
	}
	return Yt{svc: s}
}

// getPrio finds out which the lowest prio column is, which we want to display
func getPrio(width uint, cols []ConfigColumn) uint {
	tmp := make([]ConfigColumn, len(cols), (cap(cols)+1)*2)
	copy(tmp, cols)
	sort.Sort(byPrio(tmp))

	var prio, lastPrio, clen uint = ^uint(0), ^uint(0), 0
	for _, c := range tmp {
		if c.Priority < prio {
			lastPrio = prio
		}
		prio = c.Priority
		if clen += 1 + uint(len(c.HeaderCaption)); clen > width+1 {
			return lastPrio
		}
	}
	return prio
}

func getFormats(w uint, cols []ConfigColumn, cache map[uint][]cachedFmt) (ret []cachedFmt) {
	var toPad []paddable
	var usedWidth uint = 0

	// retrieve cache if possible
	if fmts, ok := cache[w]; ok {
		return fmts
	}

	leastPrio := getPrio(w, cols)
	for _, c := range cols {
		if c.Priority < leastPrio {
			continue
		}
		usedWidth += 1 + uint(len(c.HeaderCaption))
		ret = append(ret, cachedFmt{
			fmt:    fmt.Sprintf("%%%d.%dv", len(c.HeaderCaption), len(c.HeaderCaption)),
			field:  c.FieldName,
			header: c.HeaderCaption})
		if c.Pad == "left" || c.Pad == "right" {
			toPad = append(toPad, paddable{
				entry: &ret[len(ret)-1],
				capt:  c.HeaderCaption,
			})
		}
	}

	// fix the fields that need padding
	for _, c := range toPad {
		p := ""
		if c.capt[0] != ' ' {
			p = "-"
		}
		l := (w-(usedWidth-1))/uint(len(toPad)) + uint(len(c.capt))
		c.entry.fmt = fmt.Sprintf("%%%v%d.%dv", p, l, l)
	}

	// update cache and return
	cache[w] = ret
	return
}

func (v Vid) Format(width uint, cache map[uint][]cachedFmt) string {
	var ret []string
	fmts := getFormats(width, config.VideoListColumns, cache)

	for _, t := range fmts {
		if t.field == "PublishedAt" {
			val := v.PublishedAt.Format("2006-01-02")
			ret = append(ret, fmt.Sprintf(t.fmt, val))
		} else if t.field == "ViewCount" {
			val := strconv.FormatUint(v.ViewCount, 10)
			ret = append(ret, fmt.Sprintf(t.fmt, val))
		} else if t.field == "LikePercentage" {
			val := strconv.FormatUint(v.LikeCount*100/(v.LikeCount+v.DislikeCount), 10)
			ret = append(ret, fmt.Sprintf(t.fmt, val))
		} else if t.field == "Duration" {
			ret = append(ret, fmt.Sprintf(t.fmt, v.Duration.String()))
		} else if t.field == "Title" {
			ret = append(ret, fmt.Sprintf(t.fmt, v.Title))
		} else if t.field == "ChannelTitle" {
			ret = append(ret, fmt.Sprintf(t.fmt, v.ChannelTitle))
		}
	}
	return strings.Join(ret, " ")
}

func (c Chan) Format(width uint, cache map[uint][]cachedFmt) string {
	var ret []string
	fmts := getFormats(width, config.ChannelListColumns, cache)

	for _, t := range fmts {
		if t.field == "SubscriberCount" {
			val := strconv.FormatUint(c.SubscriberCount, 10)
			ret = append(ret, fmt.Sprintf(t.fmt, val))
		} else if t.field == "ViewCount" {
			val := strconv.FormatUint(c.ViewCount, 10)
			ret = append(ret, fmt.Sprintf(t.fmt, val))
		} else if t.field == "VideoCount" {
			val := strconv.FormatUint(c.VideoCount, 10)
			ret = append(ret, fmt.Sprintf(t.fmt, val))
		} else if t.field == "Title" {
			ret = append(ret, fmt.Sprintf(t.fmt, c.Title))
		}
	}
	return strings.Join(ret, " ")
}

// retrieveCache gets all ListItem from cache with a valid id. The ids of
// entries that could not be found in Cache are appended to unmatched.
func retrieveCache(ids []string, cache map[string]interface{}, unmatched *[]string) (ret []ListItem) {
	for _, id := range ids {
		if v, ok := cache[id]; ok {
			ret = append(ret, v.(ListItem))
		} else {
			*unmatched = append(*unmatched, id)
		}
	}
	return
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
	ret := retrieveCache(ids, vidCache, &toFetch)
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
		chanCache[v.Id] = c
		ret = append(ret, c)
	}
	return ret
}

func (y *Yt) GetChannels(ids []string) []ListItem {
	var toFetch []string
	ret := retrieveCache(ids, chanCache, &toFetch)
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
		chanCache[v.Id] = c
		ret = append(ret, c)
	}
	return ret
}
