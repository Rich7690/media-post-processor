package moviedb

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	//"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog/log"
	"media-web/internal/config"
	"media-web/internal/utils"
	"media-web/internal/web"
	"net/http"
)

func Search(query string) (*web.MovieSearchResults, error) {

	//key := fmt.Sprintf("movie-search-%s", query)
	var resp *http.Response
	var body []byte
	var err error
	//body, err = storage.RedisClient.Get(key).Bytes()//redis.Bytes(connection.Do("GET", key))

	//if err != nil {
	params := url.Values{}
	params.Add("api_key", config.GetMovieDBAPIKey())
	params.Add("query", query)
	params.Add("language", "en-US")
	params.Add("page", "1")
	request := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?%s", params.Encode())
	log.Debug().Msg("Calling: " + request)
	resp, body, err = utils.GetRequest(request)
	//} else {
	//	log.Debug().Msg("Redis hit on key: " + key)
	//	resp = &http.Response{StatusCode:200}
	//}

	if err != nil {
		return nil, err
	}

	response := web.MovieSearchResults{}

	if resp.StatusCode == 404 {
		return &response, nil
	}

	if resp.StatusCode < 300 {
		//storage.RedisClient.Set(key, body, time.Minute * 5)
		err = json.Unmarshal(body, &response)

		if err != nil {
			log.Err(err).Str("response", string(body)).Msg("Error parsing json")
			return nil, err
		}

		return &response, nil

	} else {
		log.Err(err).Int("status_code", resp.StatusCode).Str("response", string(body)).Msg("Error calling moviedb")
		return nil, errors.New("Failed to get movie search result")
	}
}
