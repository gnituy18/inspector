package util

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"inspector/model"

	"github.com/PuerkitoBio/goquery"
	"github.com/avast/retry-go"
)

var (
	maxPageCount    = 300
	videoListUrlPfx = "https://ivod.ly.gov.tw/Demand/ListLgltVideo"
	videoPageUrlPfx = "https://ivod.ly.gov.tw/Play/Clip"
)

func GetLegVideos(legId int, workerCount int) []model.Video {
	videos := []model.Video{}

	mutex := &sync.Mutex{}
	taskStart := make(chan struct{}, workerCount)
	taskDone := make(chan struct{})
	for i := 0; i < maxPageCount; i++ {
		go func(pn int) {
			taskStart <- struct{}{}
			log.Printf("vid page:%d: starting", pn)
			defer log.Printf("vid page:%d: done", pn)

			url := videoListUrlPfx + "/" + strconv.Itoa(legId) + "?page=" + strconv.Itoa(pn)
			resp, err := http.Get(url)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				log.Fatal(err)
			}

			clipUls := []*goquery.Selection{}
			doc.Find("#clipUl").Each(func(_ int, s *goquery.Selection) {
				clipUls = append(clipUls, s)
			})

			if len(clipUls) == 0 {
				taskDone <- struct{}{}
				return
			}

			for i, clipUl := range clipUls {
				// check out 'https://ivod.ly.gov.tw/Demand/ListLgltVideo/4790' for html structure
				thumbnailBtn := clipUl.Find(".thumbnail-btn").First()

				if strings.Contains(thumbnailBtn.Text(), "準備") {
					log.Printf("page:%d i:%d skipping...", pn, i)
					continue
				}

				a := thumbnailBtn.Find("a").First()
				isHD := a.Text() == "寬頻"

				pathStrs := strings.Split(a.AttrOr("href", ""), "/")
				videoId, err := strconv.Atoi(pathStrs[len(pathStrs)-1])
				if err != nil {
					log.Fatal(err)
				}

				clipListText := clipUl.Find(".clip-list-text")
				term, session, committee := 0, 0, ""
				fmt.Sscanf(clipListText.Find("h5").Text(), "第%d屆 第%d會期　主辦單位：%s", &term, &session, &committee)

				timeStr := strings.ReplaceAll(clipListText.Find("p").Eq(3).Text(), "會議時間：", "")
				t, err := time.Parse("2006-01-02 15:04", timeStr)
				if err != nil {
					log.Fatal(err)
				}

				desc := clipListText.Find("span").Text()

				vidQuality := "300K"
				if isHD {
					vidQuality = "1M"
				}

				videoPageUrl := videoPageUrlPfx + "/" + vidQuality + "/" + strconv.Itoa(videoId)
				masterPlaylistURL := ""
				if err = retry.Do(func() error {
					resp, err := http.Get(videoPageUrl)
					if err != nil {
						return err
					}
					defer resp.Body.Close()
					doc, err := goquery.NewDocumentFromReader(resp.Body)
					if err != nil {
						return err
					}

					strs := strings.Split(doc.Find("#fPlayer").Next().Text(), "\"")
					if len(strs) < 2 {
						log.Printf("page:%d i:%d: failed\ncontinue...\n", pn, i)
						return errors.New("failed to get master playlist url")
					}
					masterPlaylistURL = strs[1]

					return nil
				}, retry.Attempts(10), retry.Delay(time.Second)); err != nil {
					log.Fatal(err)
				}

				mutex.Lock()
				videos = append(videos, model.Video{
					Id:          videoId,
					PlaylistUrl: masterPlaylistURL,
					Desc:        desc,
					Term:        term,
					Session:     session,
					Committee:   committee,
					Timestamp:   t.Unix(),
					IsHD:        isHD,
				})
				mutex.Unlock()
			}

			taskDone <- struct{}{}
		}(i + 1)
	}

	for i := 0; i < maxPageCount; i++ {
		<-taskDone
		<-taskStart
	}

	return SortVideos(videos)
}

func SortVideos(videos []model.Video) []model.Video {
	sort.Slice(videos, func(i, j int) bool {
		if videos[i].Timestamp == videos[j].Timestamp {
			return videos[i].Id > videos[j].Id
		}
		return videos[i].Timestamp > videos[j].Timestamp
	})
	return videos
}
