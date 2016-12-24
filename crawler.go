package main 

import (
  "fmt"
  "io/ioutil"
  "net/http"
  "net/url"
  "regexp"
  "runtime"
  "strings"
)

const NCPU = 8

type filterFunc func(string, Crawler) bool

type Crawler struct {
  host string 
  urls chan string
  filteredUrls chan string
  filters []filterFunc
  re *regexp.Regexp
  count int
}

func (c *Crawler) start() {
  // url filtering thread, send filtered urls to urls channel
  go func() {
    for n := range c.urls {
      go c.filter(n)
    }
  }()
  // crawling thread, waiting for filtered urls from urls channel
  go func() {
    for s := range c.filteredUrls {
      fmt.Println(s)
      c.count++
      fmt.Println(c.count)
      go c.crawl(s)
    }
  }()
}

func (c *Crawler) filter(url string) {
  temp := false
  for _, fn := range c.filters {
    temp = fn(url, c)
    if temp != true {
      return 
    }
  }
  c.filteredUrls <- url
}

// find all text on page
func (c *Crawler) crawl(url String) {
  resp, err := http.Get(url)
  if err != nil {
    fmt.Println("An Error has ocurred")
    fmt.Println(err)
  } else {
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
      fmt.Println("An Error has ocurred in reading url response")
    } else {
      strBody := string(body)
      c.extractUrls(url, strBody)
    }
  }
}

// find all links on page, pass the absolute urls to channel (or make absolute)
func (c *Crawler) extractUrls(Url, body string) {
  newUrls := c.re.FindAllStringSubmatch(body, -1)
  u := ""
  baseUrl, _ := url.Parse(Url)
  if newUrls != nil {
    for _, z := range newUrls {
      u = z[1]
      ur, err := url.Parse(z[1])
      if err == nil {
        // pass url to chan only when it is absolute
        if ur.IsAbs() == true {
          c.urls <- u 
        } else if ur.IsAbs() == false {
          c.urls <- baseUrl.ResolveReference(ur).String()
        } else if strings.HasPrefix(u, "//") {
          c.urls <- "http:" + u
        } else if strings.HasPrefix(u, "/") {
          c.urls <- c.host + u
        } else {
          c.urls <- Url + u
        }
      }
    }
  }
}

func (c *Crawler) addFilter(filter filterFunc) Crawler {
  c.filters = append(c.filters, filter)
  return c
}

func (c *Crawler) stop() {
  close(c.urls)
  close(c.filteredUrls)
}

func main() {
  runtime.GOMAXPROCS(NCPU)

  c := Crawler{
    "http://www.thesaurus.com/", 
    make(chan string),
    make(chan string),
    make([]filterFunc, 0),
    regexp.MustCompile("(?s)<a[ t]+.*?href="(http.*?)".*?>.*?</a>"),
    0,
  }

  c.addFilter(func(Url string, c Crawler) bool {
    return strings.Contains(Url, c.host)
  }).start()

  c.urls <- c.host

  var input string
  fmt.Scanln(&input)
}
