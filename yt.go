package main

import (
	"code.google.com/p/google-api-go-client/googleapi/transport"
	"code.google.com/p/google-api-go-client/youtube/v3"
	"log"
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

func (y Yt) GetChannels(ids []string) []Chan {
	ret := make([]Chan, 0)
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
