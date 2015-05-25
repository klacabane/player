package search

import (
	"errors"
	"fmt"
	"net/http"

	"code.google.com/p/google-api-go-client/googleapi/transport"
	"code.google.com/p/google-api-go-client/youtube/v3"
)

const API_KEY = "AIzaSyBIM-dPY4ky7YAk4jLSgkf4axlTzzvFSlU"

var MAX_RESULTS int64 = 10

type Result struct {
	Url   string
	Title string
}

func Do(term string) ([]Result, error) {
	client := &http.Client{
		Transport: &transport.APIKey{Key: API_KEY},
	}

	service, err := youtube.New(client)
	if err != nil {
		return nil, err
	}

	call := service.Search.List("id,snippet").
		Q(term).
		MaxResults(MAX_RESULTS)
	resp, err := call.Do()
	if err != nil {
		return nil, err
	}
	for _, item := range resp.Items {
		fmt.Println(item.Snippet.Title)
	}
	return nil, errors.New("not implemented")
}
