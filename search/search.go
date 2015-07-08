package search

import (
	"encoding/json"
	"fmt"
	"net/http"

	"code.google.com/p/google-api-go-client/googleapi/transport"
	"code.google.com/p/google-api-go-client/youtube/v3"
)

const API_KEY = ""

const REDDIT_EP = "http://reddit.com/r/%s/hot.json"

type Result struct {
	Url   string
	Title string
}

func Youtube(term string, max int64) ([]Result, error) {
	client := &http.Client{
		Transport: &transport.APIKey{Key: API_KEY},
	}

	service, err := youtube.New(client)
	if err != nil {
		return nil, err
	}

	call := service.Search.List("id,snippet").
		Q(term).
		MaxResults(max)
	resp, err := call.Do()
	if err != nil {
		return nil, err
	}

	ret := make([]Result, len(resp.Items))
	for i, item := range resp.Items {
		ret[i] = Result{
			Url:   "http://youtube.com/watch?v=" + item.Id.VideoId,
			Title: item.Snippet.Title,
		}
	}
	return ret, nil
}

func Reddit(sub string) ([]Result, error) {
	res, err := http.Get(fmt.Sprintf(REDDIT_EP, sub))
	if err != nil {
		return nil, err
	}

	var result struct {
		Kind string
		Data struct {
			Modhash  string
			Children []struct {
				Kind string
				Data struct {
					Result
					Domain string
				}
			}
		}
	}

	var ret []Result
	err = json.NewDecoder(res.Body).Decode(&result)
	for _, child := range result.Data.Children {
		child.Data.Title = fmt.Sprintf("%s\n%s", child.Data.Title, child.Data.Domain)
		ret = append(ret, child.Data.Result)
	}
	return ret, err
}
