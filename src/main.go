package main

import (
	"encoding/json"
	"fmt"
	"github.com/kurrik/oauth1a"
	"github.com/kurrik/twittergo"
	"io/ioutil"
	"net/http"
	
	"net/url"
	"os"
	"sort"

)

// Default Request Handler
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	
	userInfo := twitHandler()
	//sorting
	sort.Sort(ByFollowersCount{userInfo})
	//displaying 
	fmt.Fprintf(w, "<h1>Hello %s!</h1>", r.URL.Path[1:])
	fmt.Fprintf(w, "<table><tr><th>Screen Name</th><th>Image</th><th>Followers count</th></tr>")
	for _, ele := range userInfo {
		fmt.Fprintf(w, "<tr><td>"+ ele.Screen_name+ "</td><td><img src='"+ele.Profile_image_url+"'/></td><td>%d</td></tr>",ele.Followers_count)
	}
	fmt.Fprintf(w, "</table>")
}
//type to use with sorter interface
type UsersInfo []UserInfo
//idiomatic go implementing sorter
func (s UsersInfo) Len() int      { return len(s) }
func (s UsersInfo) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
//to sort  by followers count
type ByFollowersCount struct{ UsersInfo }
//making the highest first
func (s ByFollowersCount) Less(i, j int) bool {
	return s.UsersInfo[i].Followers_count > s.UsersInfo[j].Followers_count
}

func main() {
	//start handling the requests at root with a handler function
	http.HandleFunc("/", defaultHandler)
	//starting the web server at port 8080
	http.ListenAndServe(":8080", nil)
	
}
//this includes all twitter interactions and returns the required list of data in []UserInfo
func twitHandler() []UserInfo {
	var (
		
		client *twittergo.Client

	)
	//initializing twitter client configuration
	config := &oauth1a.ClientConfig{
		ConsumerKey:    "XpxtW8CTzEu2zWB1jC1uQ",
		ConsumerSecret: "eowFB6GdzlEca88SATTqpTUcvmd3UiIig4Mi6yaHej8",
	}
	
	client = twittergo.NewClient(config, nil)
	//set the screen name of the user 
	screen_name := "spolsky"
	//ids := getFollowersIds(screen_name, client)
	//fmt.Println(ids)
	//getInfo(ids, client)
	
	//get the tweets of the user by using screen name
	tweets := getRetweets(screen_name, client)
	//get user ids who retweeted those tweets
	ids := getRetweeters(tweets, client)
	//get user inforamtion like followers count etc usinng the ids
	userInfo := getInfo(ids, client)
	return userInfo

}
/**
--getRetweeters(tline twittergo.Timeline, client *twittergo.Client) []uint
--Returns user ids by querying retweeters api 1.1
*/
func getRetweeters(tline twittergo.Timeline, client *twittergo.Client) []uint {
	var (
		req  *http.Request
		err  error
		resp *twittergo.APIResponse
		//flw  Followers
	)
	
	query := url.Values{}

	query.Set("count", "50")
	ids := make([]uint, 0)
	//loop for each retweeted tweet and get the user ids
	for i, tweet := range tline {
		query.Set("id", tweet.IdStr())
		url := fmt.Sprintf("/1.1/statuses/retweeters/ids.json?%v", query.Encode())
		fmt.Println(url)
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("Could not parse request: %v\n", err)
			os.Exit(1)
		}
		resp, err = client.SendRequest(req)
		if err != nil {
			fmt.Printf("Could not send request: %v\n", err)
			os.Exit(1)
		}
		//fmt.Println(resp)
		b, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(b))

		flw := new(Followers)
		json.Unmarshal(b, &flw)
		ids = append(ids, flw.Ids...)
		fmt.Println("retweeted ids", i, ids)
	}


	fmt.Printf("Rate limit: %v\n", resp.RateLimit())
	fmt.Printf("Rate limit remaining: %v\n", resp.RateLimitRemaining())
	fmt.Printf("Rate limit reset: %v\n", resp.RateLimitReset())
	return ids

}
/**
--getRetweets(screen_name string, client *twittergo.Client) []twittergo.Tweet
--Returns filtered timeline (a collection of tweets) by using /1.1/statuses/user_timeline.json api 1.1
*/
func getRetweets(screen_name string, client *twittergo.Client) []twittergo.Tweet {
	var (
		req  *http.Request
		err  error
		resp *twittergo.APIResponse
		//flw  Followers
	)
	//set query params
	query := url.Values{}
	query.Set("screen_name", screen_name)
	query.Set("count", "20")
	query.Set("exclude_replies", "true")

	query.Set("trim_user", "true")
	url := fmt.Sprintf("/1.1/statuses/user_timeline.json?%v", query.Encode())
	//	query.Set("q", "twitterapi")
	//url := fmt.Sprintf("/1.1/search/tweets.json?%v", query.Encode())
	fmt.Println(url)
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Could not parse request: %v\n", err)
		os.Exit(1)
	}
	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Printf("Could not send request: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(resp)
	b, _ := ioutil.ReadAll(resp.Body)

	tline := new(twittergo.Timeline)
	json.Unmarshal(b, &tline)

	for i, ele := range *tline {
		fmt.Println(i, ele.Text())
	}
	fmt.Printf("Rate limit: %v\n", resp.RateLimit())
	fmt.Printf("Rate limit remaining: %v\n", resp.RateLimitRemaining())
	fmt.Printf("Rate limit reset: %v\n", resp.RateLimitReset())
	return *tline
}

type Followers struct {
	Ids []uint
}
type UserInfo struct {
	Screen_name       string
	Profile_image_url string
	Followers_count   int
}
/**
--getFollowersIds(screen_name string, client *twittergo.Client) []uint
--Returns follower ids by using /1.1/followers/ids.json? api 1.1
*/
func getFollowersIds(screen_name string, client *twittergo.Client) []uint {
	var (
		req  *http.Request
		err  error
		resp *twittergo.APIResponse
		
	)

	query := url.Values{}
	query.Set("screen_name", screen_name)
	query.Set("count", "10")
	query.Set("cursor", "-1")

	url := fmt.Sprintf("/1.1/followers/ids.json?%v", query.Encode())
	//	query.Set("q", "twitterapi")
	//url := fmt.Sprintf("/1.1/search/tweets.json?%v", query.Encode())
	fmt.Println(url)
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Could not parse request: %v\n", err)
		os.Exit(1)
	}
	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Printf("Could not send request: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(resp)
	b, _ := ioutil.ReadAll(resp.Body)

	flw := new(Followers)
	json.Unmarshal(b, &flw)

	fmt.Println(string(b))
	fmt.Printf("Rate limit: %v\n", resp.RateLimit())
	fmt.Printf("Rate limit remaining: %v\n", resp.RateLimitRemaining())
	fmt.Printf("Rate limit reset: %v\n", resp.RateLimitReset())
	return flw.Ids
}
/**
--getInfo(ids []uint, client *twittergo.Client) []UserInfo 
--Returns a slice of UserInfo by using /1.1/users/lookup.json api 1.1
*/
func getInfo(ids []uint, client *twittergo.Client) []UserInfo {
	var (
		req  *http.Request
		err  error
		resp *twittergo.APIResponse
		
	)
	//set query string
	query := url.Values{}
	idStr := ""
	for _, e := range ids {
		idStr = fmt.Sprintf("%d", e) + "," + idStr
	}
	query.Set("user_id", idStr)
	fmt.Println(idStr)

	url := fmt.Sprintf("/1.1/users/lookup.json?%v", query.Encode())
	
	fmt.Println(url)
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Could not parse request: %v\n", err)
		os.Exit(1)
	}
	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Printf("Could not send request: %v\n", err)
		os.Exit(1)
	}
	//fmt.Println(resp)
	b, _ := ioutil.ReadAll(resp.Body)

	uInfo := make([]UserInfo, 0)
	json.Unmarshal(b, &uInfo)
	fmt.Println("uInfo", uInfo)
	//fmt.Println(string(b))
	fmt.Printf("Rate limit: %v\n", resp.RateLimit())
	fmt.Printf("Rate limit remaining: %v\n", resp.RateLimitRemaining())
	fmt.Printf("Rate limit reset: %v\n", resp.RateLimitReset())
	return uInfo
}
