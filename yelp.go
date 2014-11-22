package goyelp

import (
	"errors"
	"github.com/garyburd/go-oauth/oauth"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

var YelpDefaultLimit = 20

type YelpClient struct {
	consumerKey       string
	consumerSecret    string
	token             string
	tokenSecret       string
	httpClient        *http.Client
	credentials       oauth.Credentials
	clientCredentials oauth.Credentials
	apiUrl            string
}

type YelpSearchCriteria struct {
	Term           string
	Location       string
	LatLng         GeoLocation
	Limit          int
	Offset         int
	Sort           int
	CategoryFilter string
	RadiusFilter   int
}

type YelpSearchResult struct {
	Total      int            `json:"total"`
	Businesses []YelpBusiness `json:"businesses"`
}

type YelpBusiness struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Rating      float32 `json:"rating"`
	ReviewCount int     `json:"review_count"`
	Location    struct {
		Address       []string    `json:"address"`
		City          string      `json:"city"`
		State         string      `json:"state_code"`
		PostalCode    string      `json:"postal_code"`
		Country       string      `json:"country_code"`
		Neighborhoods []string    `json:"neighborhoods"`
		Coordinate    GeoLocation `json:"coordinate"`
	} `json:"location"`
	Phone      string     `json:"phone"`
	Url        string     `json:"url"`
	MobileUrl  string     `json:"mobile_url"`
	Categories [][]string `json:"categories"`
}

type GeoLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func NewYelpClient(consumerKey string, consumerSecret string, token string, tokenSecret string, httpClient *http.Client) *YelpClient {

	var client = YelpClient{
		consumerKey:    consumerKey,
		consumerSecret: consumerSecret,
		token:          token,
		tokenSecret:    tokenSecret,
		httpClient:     httpClient,
	}
	client.apiUrl = "http://api.yelp.com/v2/"
	client.credentials = oauth.Credentials{Token: client.token, Secret: client.tokenSecret}
	client.clientCredentials = oauth.Credentials{Token: client.consumerKey, Secret: client.consumerSecret}
	if client.httpClient == nil {
		client.httpClient = http.DefaultClient
	}
	return &client
}

func getLimit(criteria YelpSearchCriteria) int {
	if criteria.Limit == 0 {
		return YelpDefaultLimit
	}
	return criteria.Limit
}

func (y *YelpClient) Search(criteria YelpSearchCriteria) ([]byte, error) {

	var searchUrl = y.apiUrl + "search"
	var query = url.Values{"term": {criteria.Term},
		"sort":     {strconv.Itoa(criteria.Sort)},
		"location": {criteria.Location},
		"limit":    {strconv.Itoa(getLimit(criteria))},
		"offset":   {strconv.Itoa(criteria.Offset)},
	}
	if criteria.CategoryFilter != "" {
		query.Add("category_filter", criteria.CategoryFilter)
	}

	var client = oauth.Client{Credentials: y.clientCredentials}

	resp, err := client.Get(y.httpClient, &y.credentials, searchUrl, query)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

func (y *YelpClient) Business(yelpId string) ([]byte, error) {

	var businessUrl = y.apiUrl + "business/" + yelpId
	var client = oauth.Client{Credentials: y.clientCredentials}

	resp, err := client.Get(y.httpClient, &y.credentials, businessUrl, url.Values{})
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("not found")
	}

	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}
