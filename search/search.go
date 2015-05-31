package search

import (
	"net/http"

	"code.google.com/p/google-api-go-client/googleapi/transport"
	"code.google.com/p/google-api-go-client/youtube/v3"
)

const API_KEY = "AIzaSyBIM-dPY4ky7YAk4jLSgkf4axlTzzvFSlU"

type Result struct {
	Url   string
	Title string
}

func Do(term string, max int64) ([]Result, error) {
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
