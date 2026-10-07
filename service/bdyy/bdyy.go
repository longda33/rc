package bdyy

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

const endpoint = "https://api.xcvts.cn/api/music/bdyy"

var client = &http.Client{Timeout: 25 * time.Second}

type searchResponse struct {
	Code int `json:"code"`
	Data []struct {
		Name       string   `json:"name"`
		Artist     string   `json:"artist"`
		Singers    []string `json:"_singer"`
		DetailPage string   `json:"detail_page"`
		Cover      string   `json:"pic"`
	} `json:"data"`
}

type detailResponse struct {
	Code int `json:"code"`
	Data struct {
		Name       string `json:"name"`
		Artist     string `json:"artist"`
		Cover      string `json:"cover"`
		DetailPage string `json:"detail_page"`
		PlayURL    string `json:"play_url"`
		Lyric      string `json:"lrc"`
	} `json:"data"`
}

func request(values url.Values, out any) error {
	target := endpoint + "?" + values.Encode()
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("bdyy HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func quality(extra map[string]string) string {
	if value := strings.TrimSpace(extra["bdyy_quality"]); value != "" {
		return value
	}
	return "2000kflac"
}

func Search(keyword string) ([]model.Song, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, errors.New("bdyy empty keyword")
	}
	var payload searchResponse
	err := request(url.Values{
		"msg": {keyword}, "br": {"2000kflac"}, "type": {"json"},
		"p": {"1"}, "sc": {"50"},
	}, &payload)
	if err != nil {
		return nil, err
	}
	if payload.Code != 200 {
		return nil, errors.New("bdyy search failed")
	}
	result := make([]model.Song, 0, len(payload.Data))
	for index, item := range payload.Data {
		if item.Name == "" || item.DetailPage == "" {
			continue
		}
		artist := item.Artist
		if artist == "" && len(item.Singers) > 0 {
			artist = strings.Join(item.Singers, " / ")
		}
		result = append(result, model.Song{
			ID: item.DetailPage, Name: item.Name, Artist: artist,
			Cover: item.Cover, Source: "bdyy",
			Extra: map[string]string{
				"bdyy_n":       strconv.Itoa(index + 1),
				"bdyy_quality": "2000kflac",
			},
		})
	}
	return result, nil
}

func detail(song *model.Song) (string, error) {
	if song.Extra == nil {
		song.Extra = map[string]string{}
	}
	name := strings.TrimSpace(song.Name)
	if name == "" {
		return "", errors.New("bdyy empty song name")
	}
	n := song.Extra["bdyy_n"]
	if n == "" {
		n = "1"
	}
	var payload detailResponse
	err := request(url.Values{
		"msg": {name}, "br": {quality(song.Extra)}, "type": {"json"},
		"n": {n}, "p": {"1"}, "sc": {"50"},
	}, &payload)
	if err != nil {
		return "", err
	}
	if payload.Code != 200 || payload.Data.PlayURL == "" {
		return "", errors.New("bdyy has no playable URL")
	}
	if payload.Data.Name != "" {
		song.Name = payload.Data.Name
	}
	if payload.Data.Artist != "" {
		song.Artist = payload.Data.Artist
	}
	if payload.Data.Cover != "" {
		song.Cover = payload.Data.Cover
	}
	song.Ext = "flac"
	song.Extra["lyric"] = payload.Data.Lyric
	song.Extra["bdyy_detail_page"] = payload.Data.DetailPage
	return payload.Data.PlayURL, nil
}

func Download(song *model.Song) (string, error) {
	return detail(song)
}

func Lyric(song *model.Song) (string, error) {
	if value := strings.TrimSpace(song.Extra["lyric"]); value != "" {
		return value, nil
	}
	_, err := detail(song)
	if err != nil {
		return "", err
	}
	return song.Extra["lyric"], nil
}
