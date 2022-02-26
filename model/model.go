package model

type Legislator struct {
	Name   string  `json:"name"`
	Id     int     `json:"id"`
	Videos []Video `json:"videos"`
}

type Video struct {
	Id          int    `json:"id"`
	PlaylistUrl string `json:"playlistUrl"`

	Term      int    `json:"term"`
	Session   int    `json:"session"`
	Committee string `json:"committee"`
	Desc      string `json:"desc"`

	Timestamp int64 `json:"timestamp"`
	IsHD      bool  `json:"isHD"`
}
