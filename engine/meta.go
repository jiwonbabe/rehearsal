package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Meta data as the Graph API returns it (or the mock in data/meta/<handle>/). Loaded by cmd/, never fetched here.

type MetaAccount struct {
	Source               string                        `json:"source"`
	Username             string                        `json:"username"`
	FollowersCount       float64                       `json:"followers_count"`
	FollowerDemographics map[string]map[string]float64 `json:"follower_demographics"`
}

type MediaInsights struct {
	Reach    float64 `json:"reach"`
	Views    float64 `json:"views"`
	Saved    float64 `json:"saved"`
	Shares   float64 `json:"shares"`
	Likes    float64 `json:"likes"`
	Comments float64 `json:"comments"`
	LinkTaps float64 `json:"link_taps"`
	Replies  float64 `json:"replies"`
	Reposts  float64 `json:"reposts"`
}

type MediaItem struct {
	ID               string        `json:"id"`
	Timestamp        string        `json:"timestamp"`
	MediaProductType string        `json:"media_product_type"`
	PostType         string        `json:"post_type"`
	Insights         MediaInsights `json:"insights"`
}

type MediaList struct {
	Source string      `json:"source"`
	Days   int         `json:"days"`
	Data   []MediaItem `json:"data"`
}

type LabelledPost struct {
	PostType      string  `json:"post_type"`
	LikeCount     float64 `json:"like_count"`
	CommentsCount float64 `json:"comments_count"`
}

type HashtagTop struct {
	Hashtag  string         `json:"hashtag"`
	TopMedia []LabelledPost `json:"top_media"`
}

type Account struct {
	Username       string         `json:"username"`
	FollowersCount float64        `json:"followers_count"`
	MediaCount     float64        `json:"media_count"`
	Biography      string         `json:"biography"`
	Recent         []LabelledPost `json:"recent"`
}

type MetaData struct {
	Account  MetaAccount
	Media    MediaList
	Threads  MediaList
	Hashtags struct {
		Data []HashtagTop `json:"data"`
	}
	Peers struct {
		Data []Account `json:"data"`
	}
	Creators struct {
		Data []Account `json:"data"`
	}
	Reach struct {
		Data map[string]any `json:"data"`
	}
}

// LoadMeta reads the seven JSON files of one connected account.
func LoadMeta(dir string) (MetaData, error) {
	var m MetaData
	for name, into := range map[string]any{"account": &m.Account, "media": &m.Media, "threads": &m.Threads,
		"hashtags": &m.Hashtags, "peers": &m.Peers, "creators": &m.Creators, "reach_estimate": &m.Reach} {
		raw, err := os.ReadFile(filepath.Join(dir, name+".json"))
		if err != nil {
			return m, err
		}
		if err := json.Unmarshal(raw, into); err != nil {
			return m, err
		}
	}
	return m, nil
}
