package qqvariants

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/guohuiyuan/music-lib/model"
)

const (
	qqlEndpoint = "https://api.luosu.top/api/qqmusic/index.php"
	qqaEndpoint = "https://a.aa.cab/qq.music"
	qqtEndpoint = "https://tang.api.s01s.cn/music_open_api.php"
)

var client = &http.Client{Timeout: 25 * time.Second}

type qqlSearch struct {
	Code int `json:"code"`
	Data struct {
		List []struct {
			ID       int64  `json:"id"`
			Mid      string `json:"mid"`
			Name     string `json:"name"`
			Artist   string `json:"artist"`
			Pic      string `json:"pic"`
			Duration string `json:"duration"`
			PayText  string `json:"paytext"`
		} `json:"list"`
	} `json:"data"`
}
type qqaSearch struct {
	Code int `json:"code"`
	Data []struct {
		Num       int    `json:"num"`
		Song      string `json:"song"`
		Singer    string `json:"singer"`
		Cover     string `json:"cover"`
		Mid       string `json:"mid"`
		MediaMid  string `json:"media_mid"`
		AlbumMid  string `json:"album_mid"`
		AlbumName string `json:"album_name"`
	} `json:"data"`
}
type qqtSearch struct {
	SongTitle  string `json:"song_title"`
	SongMid    string `json:"song_mid"`
	SingerName string `json:"singer_name"`
	Pay        string `json:"pay"`
}

type qqlDetail struct {
	Code int `json:"code"`
	Data struct {
		Name     string `json:"name"`
		Singer   string `json:"singer"`
		CoverURL string `json:"cover_url"`
		PlayURL  string `json:"play_url"`
		Lyric    string `json:"lyric"`
		ExtName  string `json:"extName"`
		Quality  string `json:"quality"`
		BitRate  int64  `json:"bitRate"`
		FileSize int64  `json:"fileSize"`
		Duration string `json:"duration"`
	} `json:"data"`
}
type qqaDetail struct {
	Code int `json:"code"`
	Data struct {
		Song      string `json:"song"`
		Singer    string `json:"singer"`
		Cover     string `json:"cover"`
		Music     string `json:"music"`
		Mid       string `json:"mid"`
		MediaMid  string `json:"media_mid"`
		AlbumMid  string `json:"album_mid"`
		AlbumName string `json:"album_name"`
	} `json:"data"`
}
type qqtDetail struct {
	SongName      string `json:"song_name"`
	SongTitle     string `json:"song_title"`
	AlbumName     string `json:"album_name"`
	AlbumPic      string `json:"album_pic"`
	SongMid       string `json:"song_mid"`
	SingerName    string `json:"singer_name"`
	SongPlayURL   string `json:"song_play_url"`
	SongPlayURLSQ string `json:"song_play_url_sq"`
	SongPlayURLHQ string `json:"song_play_url_hq"`
	SongPlayURLPQ string `json:"song_play_url_pq"`
	SongLyric     string `json:"song_lyric"`
	Lyric         string `json:"lyric"`
	Duration      string `json:"duration"`
	SizeSQ        int64  `json:"song_size_sq_str"`
	SizeHQ        int64  `json:"song_size_hq_str"`
	SizePQ        int64  `json:"song_size_pq_str"`
	KbpsSQ        int    `json:"kbps_sq"`
	KbpsHQ        int    `json:"kbps_hq"`
	KbpsPQ        int    `json:"kbps_pq"`
}

func get(endpoint string, params url.Values, out any) error {
	u := endpoint + "?" + params.Encode()
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://y.qq.com/")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("QQ variant HTTP %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return err
	}
	return nil
}

func durationSeconds(value string) int {
	p := strings.Split(strings.TrimSpace(value), ":")
	if len(p) == 2 {
		m, _ := strconv.Atoi(p[0])
		s, _ := strconv.Atoi(p[1])
		return m*60 + s
	}
	n, _ := strconv.Atoi(value)
	return n
}
func song(extra map[string]string, source, id, name, artist, album, cover string, duration, size int64, bitrate int, ext, lyric string) model.Song {
	return model.Song{ID: id, Name: name, Artist: artist, Album: album, Cover: cover, Duration: int(duration), Size: size, Bitrate: bitrate, Source: source, Ext: ext, Extra: extra, URL: "", Link: ""}
}

func SearchQQL(keyword string) ([]model.Song, error) {
	var p qqlSearch
	v := url.Values{"ss": {keyword}, "format": {"json"}, "t": {"15"}}
	if err := get(qqlEndpoint, v, &p); err != nil {
		return nil, err
	}
	if p.Code != 200 {
		return nil, errors.New("QQ l search failed")
	}
	out := make([]model.Song, 0, len(p.Data.List))
	for i, x := range p.Data.List {
		out = append(out, song(map[string]string{"qq_variant": "qq_l", "qq_n": strconv.Itoa(i + 1), "qq_mid": x.Mid}, "qq_l", strconv.FormatInt(x.ID, 10), x.Name, x.Artist, "", x.Pic, int64(durationSeconds(x.Duration)), 0, 0, "", ""))
	}
	return out, nil
}
func SearchQQA(keyword string) ([]model.Song, error) {
	var p qqaSearch
	v := url.Values{"msg": {keyword}, "num": {"60"}}
	if err := get(qqaEndpoint, v, &p); err != nil {
		return nil, err
	}
	if p.Code != 0 {
		return nil, errors.New("QQ a search failed")
	}
	out := make([]model.Song, 0, len(p.Data))
	for _, x := range p.Data {
		out = append(out, song(map[string]string{"qq_variant": "qq_a", "qq_n": strconv.Itoa(x.Num), "qq_mid": x.Mid, "qq_media_mid": x.MediaMid}, "qq_a", x.Mid, x.Song, x.Singer, x.AlbumName, x.Cover, 0, 0, 0, "", ""))
	}
	return out, nil
}
func SearchQQT(keyword string) ([]model.Song, error) {
	var p []qqtSearch
	v := url.Values{"msg": {keyword}, "type": {"json"}}
	if err := get(qqtEndpoint, v, &p); err != nil {
		return nil, err
	}
	out := make([]model.Song, 0, len(p))
	for _, x := range p {
		out = append(out, song(map[string]string{"qq_variant": "qq_t", "qq_mid": x.SongMid}, "qq_t", x.SongMid, x.SongTitle, x.SingerName, "", "", 0, 0, 0, "", ""))
	}
	return out, nil
}

func detailQQL(s *model.Song) (string, error) {
	if s.Extra == nil {
		s.Extra = map[string]string{}
	}
	n := s.Extra["qq_n"]
	if n == "" {
		n = "1"
	}
	var p qqlDetail
	if err := get(qqlEndpoint, url.Values{"ss": {s.Name}, "format": {"json"}, "n": {n}}, &p); err != nil {
		return "", err
	}
	if p.Code != 200 || p.Data.PlayURL == "" {
		return "", errors.New("QQ l has no playable URL")
	}
	s.Name = p.Data.Name
	s.Artist = p.Data.Singer
	s.Cover = p.Data.CoverURL
	s.Extra["lyric"] = p.Data.Lyric
	s.Size = p.Data.FileSize
	s.Bitrate = int(p.Data.BitRate / 1000)
	s.Ext = p.Data.ExtName
	return p.Data.PlayURL, nil
}
func detailQQA(s *model.Song) (string, error) {
	if s.Extra == nil {
		s.Extra = map[string]string{}
	}
	n := s.Extra["qq_n"]
	if n == "" {
		n = "1"
	}
	var p qqaDetail
	if err := get(qqaEndpoint, url.Values{"msg": {s.Name}, "n": {n}, "type": {"4"}}, &p); err != nil {
		return "", err
	}
	if p.Code != 0 || p.Data.Music == "" {
		return "", errors.New("QQ a has no playable URL")
	}
	if p.Data.Song != "" {
		s.Name = p.Data.Song
	}
	if p.Data.Singer != "" {
		s.Artist = p.Data.Singer
	}
	s.Album = p.Data.AlbumName
	s.Cover = p.Data.Cover
	s.Ext = "flac"
	return p.Data.Music, nil
}
func fetchQQTDetail(s *model.Song) (qqtDetail, error) {
	if s.Extra == nil {
		s.Extra = map[string]string{}
	}
	var raw json.RawMessage
	if err := get(qqtEndpoint, url.Values{"msg": {s.Name}, "type": {"json"}, "mid": {s.Extra["qq_mid"]}}, &raw); err != nil {
		return qqtDetail{}, err
	}
	var p qqtDetail
	if len(raw) > 0 && raw[0] == '[' {
		var list []qqtDetail
		if err := json.Unmarshal(raw, &list); err != nil || len(list) == 0 {
			return qqtDetail{}, errors.New("QQ t detail response is empty")
		}
		p = list[0]
	} else if err := json.Unmarshal(raw, &p); err != nil {
		return qqtDetail{}, err
	}
	return p, nil
}

// detailQQT only resolves metadata and lyrics. It never selects a playback URL.
func detailQQT(s *model.Song) (string, error) {
	if s.Extra == nil {
		s.Extra = map[string]string{}
	}
	p, err := fetchQQTDetail(s)
	if err != nil {
		return "", err
	}
	s.Name = first(p.SongName, p.SongTitle, s.Name)
	s.Artist = p.SingerName
	s.Album = p.AlbumName
	s.Cover = first(p.AlbumPic, s.Cover)
	s.Extra["lyric"] = first(p.SongLyric, p.Lyric, "")
	return "", nil
}

// playQQT is the separate QQ tang playback resolver.
func playQQT(s *model.Song) (string, error) {
	if err := func() error { _, err := detailQQT(s); return err }(); err != nil {
		return "", err
	}
	p, err := fetchQQTDetail(s)
	if err != nil {
		return "", err
	}
	if p.SongPlayURLSQ != "" {
		s.Ext, s.Size, s.Bitrate = "flac", p.SizeSQ, p.KbpsSQ
		return p.SongPlayURLSQ, nil
	}
	if p.SongPlayURLHQ != "" {
		s.Ext, s.Size, s.Bitrate = "mp3", p.SizeHQ, p.KbpsHQ
		return p.SongPlayURLHQ, nil
	}
	if p.SongPlayURLPQ != "" {
		s.Ext, s.Size, s.Bitrate = "mp3", p.SizePQ, p.KbpsPQ
		return p.SongPlayURLPQ, nil
	}
	return "", errors.New("QQ t has no playable URL")
}
func first(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
func Download(source string) func(*model.Song) (string, error) {
	return func(s *model.Song) (string, error) {
		switch source {
		case "qq_l":
			return detailQQL(s)
		case "qq_a":
			return detailQQA(s)
		case "qq_t":
			return playQQT(s)
		default:
			return "", errors.New("unknown QQ variant")
		}
	}
}
func Lyric(source string) func(*model.Song) (string, error) {
	return func(s *model.Song) (string, error) {
		if s.Extra == nil {
			s.Extra = map[string]string{}
		}
		if source == "qq_t" && strings.TrimSpace(s.Extra["lyric"]) == "" {
			if _, err := detailQQT(s); err != nil {
				return "", err
			}
		}
		return s.Extra["lyric"], nil
	}
}
