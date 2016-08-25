package main

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itsjamie/gin-cors"
	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/urlshortener/v1"
)

const timeZone = "Asia/Taipei"
const dateTimeFormat = "2006-01-02T15:04:05Z07:00"

var newsSource = map[string]string{
	"chinatimes.com":    "中時電子報",
	"ettoday.net":       "ETtoday",
	"tvbs.com.tw":       "TVBS",
	"appledaily.com.tw": "蘋果日報",
	"ftv.com.tw":        "民視新聞",
	"ltn.com.tw":        "自由時報電子報",
	"udn.com":           "聯合新聞網",
	"ebc.net.tw":        "東森新聞",
	"setn.com":          "三立新聞網",
	"pts.org.tw":        "公視新聞",
	"nownews.com":       "NOWnews今日新聞",
	"cna.com.tw":        "中央通訊社",
	"hinet.net":         "自由時報電子報",
	"ttv.com.tw":        "台視新聞",
	"cts.com.tw":        "華視新聞",
	"musou.tw":          "沃草國會無雙",
	"yam.com":           "蕃新聞",
	"taiwanhot.net":     "台灣好新聞",
	"knowing.asia":      "Knowing",
	"101newsmedia.com":  "一零一傳媒",
	"gpwb.gov.tw":       "軍事新聞網",
	"ipcf.org.tw":       "原住民族電視台",
	"epochtimes.com":    "大紀元",
	"tw.on.cc":          "on.cc東網台灣",
	"sina.com.tw":       "臺灣新浪網",
	"ntdtv.com":         "NTDTV",
	"ctitv.com.tw":      "必POTV",
	"travelnews.tw":     "宜蘭新聞網",
}

// RssItem struct
type RssItem struct {
	Title      string    `json:"title"`
	TimeText   string    `json:"timeText"`
	Link       string    `json:"link"`
	OriginLink string    `json:"originLink"`
	Source     string    `json:"source"`
	Tag        string    `json:"tag"`
	Time       time.Time `json:"time"`
	Status     int       `json:"status"`
}

// ByTime implements sort.Interface for []RssItem based on
// the Time field.
type ByTime []RssItem

func (a ByTime) Len() int           { return len(a) }
func (a ByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTime) Less(i, j int) bool { return a[i].Time.After(a[j].Time) }

// LoadRSS loads rss from an url
func LoadRSS(tag string, url string) []RssItem {
	collect := []RssItem{}
	p := bluemonday.NewPolicy()
	parser := gofeed.NewParser()
	feed, err := parser.ParseURL(url)
	if err != nil {
		panic("Failed fetch and parse the feed!")
	}

	for _, item := range feed.Items {
		local, _ := time.Parse(dateTimeFormat, item.Published)
		location, err := time.LoadLocation(timeZone)
		if err == nil {
			local = local.In(location)
		}

		link, originLink := GetURL(item.Link)
		news := RssItem{
			Link:       link,
			OriginLink: originLink,
			Time:       local,
			TimeText:   local.Format("15:04"),
			Title:      p.Sanitize(item.Title),
			Source:     GetNewsSource(item.Link),
			Tag:        tag,
			Status:     0,
		}

		collect = append(collect, news)
	}
	return collect
}

// GetNewsSource detects news source from urls
func GetNewsSource(str string) string {
	for keyword := range newsSource {
		if strings.Contains(str, keyword) {
			return newsSource[keyword]
		}
	}
	return "!未知的來源!"
}

// CleanURL cuts a stting as url
func CleanURL(str string) string {
	return strings.Split(strings.Split(str, "&url=")[1], "&ct=")[0]
}

// GetURL cuts a string as url and makes short url
func GetURL(str string) (string, string) {
	developerKey := "AIzaSyBW-K5dEyqgBRCP5AWZyh61EbZLP4QkniA"
	longURL := CleanURL(str)
	client := &http.Client{
		Transport: &transport.APIKey{Key: developerKey},
	}
	svc, err := urlshortener.New(client)
	if err != nil {
		panic("Unable to create UrlShortener service!")
	}
	url, err := svc.Url.Insert(&urlshortener.Url{LongUrl: longURL}).Do()
	if err != nil {
		panic("Unable to get shortUrl!")
	}
	return url.Id, longURL
}

func main() {
	router := gin.Default()

	router.LoadHTMLGlob("firenewsweb/dist/*.html")
	router.Static("/static", "firenewsweb/dist/static/")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	router.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET, PUT, POST, DELETE",
		RequestHeaders:  "Origin, Authorization, Content-Type",
		ExposedHeaders:  "",
		MaxAge:          50 * time.Second,
		Credentials:     true,
		ValidateHeaders: false,
	}))

	v1 := router.Group("/api/news/v1")
	{
		v1.GET("/main", func(c *gin.Context) {
			news0 := LoadRSS("消防", "https://www.google.com.tw/alerts/feeds/04784784225885481651/10937227332545439003")
			news1 := LoadRSS("救護", "https://www.google.com.tw/alerts/feeds/04784784225885481651/10937227332545439311")
			news2 := LoadRSS("颱風 新竹市", "https://www.google.com.tw/alerts/feeds/04784784225885481651/12744255099028442939")
			news3 := LoadRSS("熱帶低氣壓 新竹市", "https://www.google.com.tw/alerts/feeds/04784784225885481651/10937227332545437714")
			news0 = append(news0, news1...)
			news0 = append(news0, news2...)
			news0 = append(news0, news3...)
			sort.Sort(ByTime(news0))

			c.JSON(200, gin.H{
				"news": news0,
			})
		})
		v1.GET("/city", func(c *gin.Context) {
			news := LoadRSS("新竹市", "https://www.google.com.tw/alerts/feeds/04784784225885481651/13141838524979976729")
			sort.Sort(ByTime(news))

			c.JSON(200, gin.H{
				"news": news,
			})
		})
		v1.GET("/typhon", func(c *gin.Context) {
			news0 := LoadRSS("颱風", "https://www.google.com.tw/alerts/feeds/04784784225885481651/12744255099028442939")
			news1 := LoadRSS("熱帶低氣壓", "https://www.google.com.tw/alerts/feeds/04784784225885481651/10937227332545437714")
			news0 = append(news0, news1...)
			sort.Sort(ByTime(news0))

			c.JSON(200, gin.H{
				"news": news0,
			})
		})
	}

	router.Run() // 0.0.0.0:8080
}
