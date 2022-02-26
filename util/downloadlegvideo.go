package util

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"inspector/model"

	"github.com/grafov/m3u8"
)

func DownloadVideo(video model.Video, outputDir string) error {
	resp, err := http.Get(video.PlaylistUrl)
	p, listType, err := m3u8.DecodeFrom(resp.Body, true)
	if err != nil {
		return err
	} else if listType != m3u8.MASTER {
		return fmt.Errorf("%s is not a master playlist", video.PlaylistUrl)
	}

	masterpl := p.(*m3u8.MasterPlaylist)
	mediaUrl := strings.ReplaceAll(video.PlaylistUrl, "playlist.m3u8", masterpl.Variants[0].URI)
	res, err := http.Get(mediaUrl)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	p, listType, err = m3u8.DecodeFrom(res.Body, true)
	if err != nil {
		return err
	}
	mediapl := p.(*m3u8.MediaPlaylist)

	dir := outputDir + "/" + strconv.Itoa(video.Id)
	os.MkdirAll(dir, os.ModePerm)
	tempDir, err := os.MkdirTemp(dir, "segments")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	var wg sync.WaitGroup
	for _, segment := range mediapl.Segments {
		if segment == nil {
			continue
		}
		wg.Add(1)
		go func(segment *m3u8.MediaSegment) {
			if segment.URI == "" {
				return
			}
			resp, err := http.Get(strings.ReplaceAll(video.PlaylistUrl, "playlist.m3u8", segment.URI))
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			fileName := strings.ReplaceAll(segment.URI, "/", "_")
			file, err := os.Create(tempDir + "/" + fileName)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			_, err = io.Copy(file, resp.Body)
			if err != nil {
				log.Fatal(err)
			}

			wg.Done()
		}(segment)
	}

	wg.Wait()

	mergeFile, err := os.Create(tempDir + "/" + strconv.Itoa(video.Id) + ".ts")
	if err != nil {
		return err
	}
	defer mergeFile.Close()

	for _, segment := range mediapl.Segments {
		if segment == nil {
			continue
		}
		fileName := strings.ReplaceAll(segment.URI, "/", "_")
		file, err := os.Open(tempDir + "/" + fileName)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(mergeFile, file)
		if err != nil {
			return err
		}
	}

	cmd := exec.Command("ffmpeg", "-i", tempDir+"/"+strconv.Itoa(video.Id)+".ts", dir+"/"+strconv.Itoa(video.Id)+".mp4")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
