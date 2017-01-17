package main

import (
	"encoding/xml"
	"fmt"
	"hash/fnv"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	fb "github.com/huandu/facebook"
	"github.com/itsjamie/gin-cors"
	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/urlshortener/v1"
)

const timeZone = "Asia/Taipei"
const dateTimeFormat0 = "2006-01-02T15:04:05Z07:00"
const dateTimeFormat1 = "Mon, 02 Jan 2006 15:04:05 -0700"
const dateTimeFormat2 = "Mon, 02 Jan 2006 15:04:05 GMT"
const dateTimeFormat3 = "2006-01-02 15:04:05"
const dateTimeFormat4 = "Mon,02 Jan 2006 15:04:05  -0700"
const dateTimeFormat5 = "2006-01-02 15:04:05 -0700 UTC"
const dateTimeFormat6 = "2006-01-02T15:04:05-07:00"
const dateTimeFormat7 = "Mon, 2 Jan 2006 15:04:05 GMT"
const dateTimeFormatFB = "2006-01-02T15:04:05-0700"

var newsSource = map[string]string{
	"twpowernews.com":                 "勁報",
	"greatnews.com.tw":                "大成報",
	"www.cdns.com.tw":                 "中華日報",
	"tssdnews.com.tw":                 "台灣新生報",
	"tynews.com.tw":                   "天眼日報",
	"fingermedia.tw":                  "指傳媒",
	"5550555.com":                     "真晨報",
	"twnewsdaily.com":                 "台灣新聞報",
	"newstaiwan.com.tw":               "台灣好報",
	"twtimes.com.tw":                  "台灣時報",
	"twreporter.org":                  "報導者",
	"mypeople.tw":                     "民眾日報",
	"chinatimes.com":                  "中時電子報",
	"ettoday.net":                     "ETtoday",
	"feedproxy.google.com/~r/ettoday": "ETtoday",
	"tvbs.com.tw":                     "TVBS",
	"appledaily.com.tw":               "蘋果日報",
	"ftv.com.tw":                      "民視新聞",
	"ltn.com.tw":                      "自由時報電子報",
	"udn.com":                         "聯合新聞網",
	"ebc.net.tw":                      "東森新聞",
	"setn.com":                        "三立新聞網",
	"pts.org.tw":                      "公視新聞",
	"nownews.com":                     "NOWnews",
	"mdnkids.com":                     "國語日報",
	"cna.com.tw":                      "中央通訊社",
	"feedproxy.google.com/~r/rsscna":  "中央通訊社",
	"hinet.net":                       "HiNet新聞",
	"storm.mg":                        "風傳媒",
	"ttv.com.tw":                      "台視新聞",
	"cts.com.tw":                      "華視新聞",
	"ithome.com.tw":                   "iThome Online",
	"technews.tw":                     "科技新報",
	"eradio.ner.gov.tw":               "國立教育廣播電台",
	"mradio.com.tw":                   "全國廣播",
	"musou.tw":                        "沃草國會無雙",
	"anntw.com":                       "台灣醒報",
	"thenewslens.com":                 "The News Lens關鍵評論網",
	"coolloud.org.tw":                 "苦勞網",
	"yam.com":                         "蕃新聞",
	"taiwanhot.net":                   "台灣好新聞",
	"knowing.asia":                    "Knowing",
	"101newsmedia.com":                "一零一傳媒",
	"peopo.org":                       "公民新聞",
	"gpwb.gov.tw":                     "青年日報",
	"ipcf.org.tw":                     "原住民族電視台",
	"ntdtv.com.tw":                    "新唐人亞太電視台",
	"cnyes.com":                       "鉅亨網",
	"epochtimes.com":                  "大紀元",
	"bltv.tv":                         "人間衛視",
	"merit-times.com.tw":              "人間福報",
	"tw.on.cc":                        "on.cc東網台灣",
	"sina.com.tw":                     "臺灣新浪網",
	"sina.com.hk":                     "香港新浪",
	"ntdtv.com":                       "NTDTV",
	"ctitv.com.tw":                    "必POTV",
	"travelnews.tw":                   "宜蘭新聞網",
	"chinesetoday.com":                "國際日報",
	"gamebase.com.tw":                 "遊戲基地",
	"soundofhope.org":                 "希望之聲",
	"cdnews.com.tw":                   "中央日報",
	"idn.com.tw":                      "自立晚報",
	"rti.org.tw":                      "中央廣播電台",
	"times-bignews.com":               "今日大話新聞",
	"saydigi.com":                     "SayDigi點子生活",
	"e-info.org.tw":                   "環境資訊中心",
	"tanews.org.tw":                   "台灣動物新聞網",
	"eventsinfocus.org":               "焦點事件",
	"cmmedia.com.tw":                  "信傳媒",
	"civilmedia.tw":                   "公民行動影音紀錄資料庫",
	"tw.iscarmg.com":                  "iscar!",
	"hk.news.yahoo.com":               "Yahoo香港",
	"hk.celebrity.yahoo.com":          "Yahoo香港",
	"hk.apple.nextmedia.com":          "香港蘋果日報",
	"s.nextmedia.com":                 "香港蘋果日報",
	"cdn.org.tw":                      "基督教今日報",
	"kairos.news":                     "Kairos風向新聞",
	"gov.taipei":                      "臺北市政府",
	"yunlin.gov.tw":                   "雲林縣政府",
	"taichung.gov.tw":                 "臺中市政府",
	"taitung.gov.tw":                  "臺東縣政府",
	"hccg.gov.tw":                     "新竹市政府",
	"fat.com.tw":                      "遠東航空",
	"ctust.edu.tw":                    "中臺科技大學",
	"msn.com":                         "msn新聞",
	"newtalk.tw":                      "新頭殼",
	"cnabc.com":                       "中央通訊社",
	"ey.gov.tw":                       "行政院全球資訊網",
	"walkerland.com.tw":               "Walkerland",
	"tsna.com.tw":                     "tsna",
	"voacantonese.com":                "美國之音",
	"ct.org.tw":                       "基督教論壇報",
	"pacificnews.com.tw":              "太平洋新聞網",
	"housefun.com.tw":                 "好房網",
	"msntw.com":                       "主流傳媒",
	"kinmen.gov.tw":                   "金門縣政府",
	"chiayi.gov.tw":                   "嘉義市政府",
	"ccu.edu.tw":                      "國立中正大學",
	"npa.gov.tw":                      "內政部警政署",
	"tainan.gov.tw":                   "台南市政府",
	"match.net.tw":                    "match生活網",
	"pixnet.net":                      "痞客邦",
	"ptt.cc":                          "PTT",
	"xn--4gq171p.com":                 "一頁新聞",
	"taiwandaily.net":                 "美洲台灣日報",
	"koreastardaily.com":              "韓星網",
	"digitimes.com.tw":                "DIGITIMES",
	"money-link.com.tw":               "富聯網",
	"qoos.com":                        "Qoos",
	"wenweipo.com":                    "文匯報",
	"hk.on.cc":                        "東網即時",
	"chinapress.com.my":               "China Press",
	"now.com":                         "now新聞",
	"singtao.ca":                      "星島日報",
	"moi.gov.tw":                      "中華民國內政部",
	"brain.com.tw":                    "動腦新聞",
	"mobile01.com":                    "mobile01",
	"hk01.com":                        "香港01",
	"mingpao.com":                     "明報新聞網",
	"passiontimes.hk":                 "熱血時報",
	"hkej.com":                        "信報",
	"thestandnews.com":                "立場新聞",
	"info.gov.hk":                     "香港特別行政區政府新聞公報",
	"macaodaily.com":                  "澳門日報",
	"hcfdrss.blogspot.com":            "爆料公社",
	"cwb.gov.tw":                      "中央氣象局",
	"yes-news.com":                    "熱新聞 YesNews",
	"hk.crntt.com":                    "中國評論新聞網",
	"dronesplayer.com":                "飛新聞",
	"hket.com":                        "香港經濟日報",
	"mygonews.com":                    "MyGoNews",
	"singtaousa.com":                  "星島日報",
	"secretchina.com":                 "看中國",
	"cardu.com.tw":                    "卡優新聞網",
	"stheadline.com":                  "頭條日報",
	"greenpeace.org":                  "綠色和平",
}

var blockedSource = map[string]bool{
	"bltv.tv":                true,
	"sina.com.tw":            true,
	"epochtimes.com":         true,
	"walkerland.com.tw":      true,
	"tsna.com.tw":            true,
	"pacificnews.com.tw":     true,
	"cdnews.com.tw":          true,
	"msntw.com":              true,
	"match.net.tw":           true,
	"pixnet.net":             true,
	"ptt.cc":                 true,
	"xn--4gq171p.com":        true,
	"taiwandaily.net":        true,
	"koreastardaily.com":     true,
	"qoos.com":               true,
	"hk.on.cc":               true,
	"wenweipo.com":           true,
	"brain.com.tw":           true,
	"mobile01.com":           true,
	"hk.apple.nextmedia.com": true,
	"s.nextmedia.com":        true,
	"sina.com.hk":            true,
	"yes-news.com":           true,
	"hk.crntt.com":           true,
	"hinet.net":              true,
	"hket.com":               true,
	"singtaousa.com":         true,
	"secretchina.com":        true,
	"greenpeace.org":         true,
}

var activedSource = map[string]bool{
	"ltn.com.tw":        true,
	"chinatimes.com":    true,
	"appledaily.com.tw": true,
	"udn.com":           true,
}

// ResultCap struct
type ResultCap struct {
	Language  string      `xml:"info>language"`
	Web       string      `xml:"info>web"`
	Parameter []Parameter `xml:"info>parameter"`
	Resource  []Resource  `xml:"info>resource"`
}

// Parameter struct
type Parameter struct {
	ValueName string `xml:"valueName"`
	Value     string `xml:"value"`
}

// Resource struct
type Resource struct {
	ResourceDesc string `xml:"resourceDesc"`
	MimeType     string `xml:"mimeType"`
	URI          string `xml:"uri"`
}

// FacebookItem struct
type FacebookItem struct {
	Gid        string    `json:"id"`
	Pid        string    `json:"id"`
	Message    string    `json:"message"`
	Story      string    `json:"story"`
	Time       time.Time `json:"time"`
	TimeText   string    `json:"timeText"`
	Link       string    `json:"link"`
	OriginLink string    `json:"originLink"`
	Source     string    `json:"source"`
}

// RssItem struct
type RssItem struct {
	Title       string `json:"title"`
	TimeText    string `json:"timeText"`
	Link        string `json:"link"`
	OriginLink  string `json:"originLink"`
	Source      string `json:"source"`
	Tag         string `json:"tag"`
	Keyword     string
	Time        time.Time `json:"time"`
	Status      int       `json:"status"`
	Hash        uint32    `json:"hash"`
	Description string    `json:"description"`
}

// ByTime implements sort.Interface for []RssItem based on
// the Time field.
type ByTime []RssItem

func (a ByTime) Len() int           { return len(a) }
func (a ByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTime) Less(i, j int) bool { return a[i].Time.After(a[j].Time) }

func fixedLink(link string, tag string) string {
	if strings.HasPrefix(link, "news_pagein.php?") {
		switch tag {
		case "大成報":
			link = "http://www.greatnews.com.tw/home/" + link
		case "勁報（勁報記者羅蔚舟）":
			link = "http://www.twpowernews.com/home/" + link
		case "台灣新聞報（記者戴欣怡）":
			link = "http://twnewsdaily.com/home/" + link
		}
	}

	if strings.HasSuffix(link, "//") {
		link = strings.TrimSuffix(link, "/")
	}

	return link
}

func fetchXML(url string) []byte {
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := netClient.Get(url)
	if err != nil {
		fmt.Printf("fetchXML http.Get error: %v", err)
		os.Exit(1)
	}

	xmldata, ioErr := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if ioErr != nil {
		fmt.Printf("fetchXML ioutil.ReadAll error: %v", err)
		return nil
	}

	return xmldata
}

func loadLocal(timetext string, tag string) time.Time {
	location, loadLocationErr := time.LoadLocation(timeZone)

	local, dateTimeErr := time.Parse(dateTimeFormat0, timetext)
	if dateTimeErr != nil {
		local, dateTimeErr = time.Parse(dateTimeFormat1, timetext)
	}
	if dateTimeErr != nil {
		local, dateTimeErr = time.Parse(dateTimeFormat2, timetext)
	}
	if dateTimeErr != nil {
		local, dateTimeErr = time.Parse(dateTimeFormat3, timetext)
	}
	if dateTimeErr != nil {
		local, dateTimeErr = time.Parse(dateTimeFormat4, timetext)
	}
	if dateTimeErr != nil {
		local, dateTimeErr = time.Parse(dateTimeFormat5, timetext)
	}
	if dateTimeErr != nil {
		local, dateTimeErr = time.Parse(dateTimeFormat6, timetext)
	}
	if dateTimeErr != nil {
		local, dateTimeErr = time.Parse(dateTimeFormat7, timetext)
	}

	if dateTimeErr != nil {
		fmt.Printf("Failed parse dateTime: %v\n", timetext)
	}

	if loadLocationErr == nil {
		local = local.In(location)
	}

	switch tag {
	case "民眾日報（記者方詠騰）":
		local = local.Add(-14 * time.Hour)
	case "勁報（勁報記者羅蔚舟）":
		local = local.Add(-8 * time.Hour)
	case "大成報":
		local = local.Add(-8 * time.Hour)
	case "台灣新聞報（記者戴欣怡）":
		local = local.Add(-8 * time.Hour)
	}

	return local
}

// LoadRSS loads rss from an url
func LoadRSS(tag string, url string) []RssItem {
	collect := []RssItem{}
	p := bluemonday.NewPolicy()
	parser := gofeed.NewParser()
	feed, err := parser.ParseURL(url)
	if err != nil {
		fmt.Printf("Failed fetch and parse the feed: %s\n", url)
		return collect
	}

	for _, item := range feed.Items {
		local := loadLocal(item.Published, tag)

		title := html.UnescapeString(p.Sanitize(item.Title))

		h := fnv.New32a()
		h.Write([]byte(title))
		hashnum := h.Sum32()

		item.Link = fixedLink(item.Link, tag)
		link, originLink := GetURL(item.Link)
		source, keyword := GetNewsSource(item.Link)

		news := RssItem{
			Link:        link,
			OriginLink:  originLink,
			Time:        local,
			TimeText:    local.Format("15:04"),
			Title:       title,
			Source:      source,
			Tag:         tag,
			Status:      0,
			Hash:        hashnum,
			Keyword:     keyword,
			Description: item.Description,
		}

		collect = append(collect, news)
	}
	return collect
}

// GetNewsSource detects news source from urls
func GetNewsSource(str string) (string, string) {
	for k := range newsSource {
		if strings.Contains(str, k) {
			return newsSource[k], k
		}
	}
	return "!未知的來源!", ""
}

// URLDecode encodes a string
func URLDecode(str string) (string, error) {
	return url.QueryUnescape(str)
}

// CleanURL cuts a stting as url
func CleanURL(str string) string {
	var tmp0 = strings.Split(str, "&url=")
	if len(tmp0) == 2 {
		var tmp1 = strings.Split(tmp0[1], "&ct=")
		if len(tmp1) == 2 {
			return tmp1[0]
		}
	}

	return str
}

// GetURL cuts a string as url and makes short url
func GetURL(str string) (string, string) {
	developerKey := "AIzaSyBW-K5dEyqgBRCP5AWZyh61EbZLP4QkniA"
	cleanedURL := CleanURL(str)
	longURL, _ := URLDecode(cleanedURL)
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

// UinqueElements removes duplicates
func UinqueElements(elements []RssItem) []RssItem {

	source0 := "中央通訊社"
	source1 := "中時電子報"
	cnaTitles := tagSources(elements, source0)
	chinatimesTitles := tagSources(elements, source1)

	var duplicated bool
	tmp := make(map[string]RssItem, 0)
	sort.Strings(cnaTitles)
	for _, ele := range elements {
		ele.Title = strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, ele.Title)

		if strings.HasSuffix(ele.Title, "...") {
			ele.Title = ele.Title[:len(ele.Title)-3]
		}

		duplicated = false
		n := sort.SearchStrings(cnaTitles, ele.Title)
		if n < len(cnaTitles) && cnaTitles[n] == ele.Title && ele.Source != source0 {
			if ele.Source != source0 {
				duplicated = true
			} else {
				ele.Status = 1
			}
		}

		if !duplicated {
			tmp[ele.Title+ele.Source] = ele
		}
	}

	tmp2 := make(map[string]RssItem, 0)
	sort.Strings(chinatimesTitles)
	var duplicatedList []string
	for _, ele := range tmp {
		duplicated = false
		n := sort.SearchStrings(chinatimesTitles, ele.Title)
		if n < len(chinatimesTitles) && chinatimesTitles[n] == ele.Title {
			if ele.Source != source1 {
				duplicated = true
				duplicatedList = append(duplicatedList, ele.Title)
			}
		}

		if !duplicated {
			tmp2[ele.Link] = ele
		}
	}

	sort.Strings(duplicatedList)
	var i int
	for _, ele := range tmp2 {
		n := sort.SearchStrings(duplicatedList, ele.Title)
		if n < len(duplicatedList) && duplicatedList[n] == ele.Title {
			ele.Source = source0
		}
		elements[i] = ele
		i++
	}

	return elements[:len(tmp2)]
}

// tagSources get titles from a specific source
func tagSources(elements []RssItem, source string) []string {
	var titles []string
	for _, ele := range elements {
		if ele.Source == source {
			ele.Title = strings.Map(func(r rune) rune {
				if unicode.IsSpace(r) {
					return -1
				}
				return r
			}, ele.Title)
			titles = append(titles, ele.Title)
		}
	}

	return titles
}

func titleIsActived(title string, keywords string) bool {
	rp := regexp.MustCompile(keywords)
	found := rp.MatchString(title)

	return found
}

// ActiveElements active elements
func ActiveElements(elements []RssItem) []RssItem {
	for i, item := range elements {
		if _, found := activedSource[item.Keyword]; found {
			elements[i].Status = 1
		} else if titleIsActived(item.Title, "竹市.*消防|消防.*竹市|竹市.*義消|義消.*竹市") {
			elements[i].Status = 1
		}
	}
	return elements
}

// ActiveAllElements active elements
func ActiveAllElements(elements []RssItem) []RssItem {
	for i := range elements {
		elements[i].Status = 1
	}
	return elements
}

// CleanupElements makes elements clean
func CleanupElements(elements []RssItem) []RssItem {
	keywords := "關鍵字搜尋搜尋|行善|廟|寺|地震.*局勢|局勢.*地震|地震.*經濟|經濟.*地震|收成|收成|價格|貸款|保險|價揚|治安|人壽|訂單|投資|火燒心|daily|價高|價低|量多|量少|民調|希望|拉麵|豪華|救護車模型|演技|綠能產業|冷門市場|人事地震|事業|表特|品牌|保險"
	rp := regexp.MustCompile(keywords)

	for i := len(elements) - 1; i >= 0; i-- {
		item := elements[i]
		if _, found := blockedSource[item.Keyword]; found {
			elements = append(elements[:i], elements[i+1:]...)
		} else if rp.MatchString(CJKnorm(item.Title)) {
			elements = append(elements[:i], elements[i+1:]...)
		}
	}

	return elements
}

// CJKnorm fixed ambiguous cjk text
func CJKnorm(s string) string {
	var str string
	str = strings.Replace(s, "巿", "市", -1)

	return str
}

func newsFetcher(feeds map[string]string) []RssItem {
	var news []RssItem

	wgFeeds := make(chan []RssItem)
	var wg sync.WaitGroup
	wg.Add(len(feeds))

	for tag, url := range feeds {
		go func(tag string, url string) {
			defer wg.Done()
			data := LoadRSS(tag, url)
			wgFeeds <- []RssItem(data)
		}(tag, url)
	}

	go func() {
		for feed := range wgFeeds {
			news = append(news, feed...)
		}
	}()

	wg.Wait()

	news = UinqueElements(news)
	news = CleanupElements(news)
	news = ActiveElements(news)
	sort.Sort(ByTime(news))

	return news
}

func main() {
	var filterAPIPoint string
	if os.Getenv("GIN_MODE") == "release" {
		filterAPIPoint = "http://localhost/firenews/api/util/v1/"
	} else {
		filterAPIPoint = "http://localhost:1234/api/util/v1/"
	}

	router := gin.Default()
	router.Use(gzip.Gzip(gzip.DefaultCompression))

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

	utilv1 := router.Group("/api/util/v1")
	{
		utilv1.GET("/filter", func(c *gin.Context) {
			sURL := c.Query("url")
			include := c.Query("include")
			dataType := c.Query("type")
			parser := gofeed.NewParser()
			feed, err := parser.ParseURL(sURL)
			if err != nil {
				return
			}

			var foundItems = make([]*gofeed.Item, 0)

			rp := regexp.MustCompile(include)
			for _, item := range feed.Items {
				foundTitle := rp.MatchString(CJKnorm(item.Title))
				foundDescription := rp.MatchString(CJKnorm(item.Description))
				foundContent := rp.MatchString(CJKnorm(item.Content))

				if foundTitle || foundDescription || foundContent {
					foundItems = append(foundItems, item)
				}
			}

			newFeed := &feeds.Feed{
				Title:       feed.Title,
				Link:        &feeds.Link{Href: feed.Link},
				Description: feed.Description,
			}

			if feed.PublishedParsed != nil {
				newFeed.Created = *feed.PublishedParsed
			} else {
				if feed.UpdatedParsed != nil {
					newFeed.Created = *feed.UpdatedParsed
				}
			}

			if feed.Author != nil {
				newFeed.Author = &feeds.Author{
					Name:  feed.Author.Name,
					Email: feed.Author.Email,
				}
			}

			newFeed.Items = make([]*feeds.Item, 0)
			for _, item := range foundItems {
				var created time.Time
				if item.PublishedParsed != nil {
					created = *item.PublishedParsed
				} else {
					created = *item.UpdatedParsed
				}

				if dataType == "cap" {
					v := ResultCap{}
					xmldata := fetchXML(item.Link)
					xmlErr := xml.Unmarshal([]byte(xmldata), &v)
					if xmlErr != nil {
						return
					}
					newFeed.Items = append(newFeed.Items, &feeds.Item{
						Title:       v.Resource[1].URI,
						Link:        &feeds.Link{Href: v.Web},
						Description: item.Description,
						Created:     created,
					})
				} else {
					newFeed.Items = append(newFeed.Items, &feeds.Item{
						Title:       item.Title,
						Link:        &feeds.Link{Href: item.Link},
						Description: item.Description,
						Created:     created,
					})
				}
			}

			rss, err := newFeed.ToRss()
			if err != nil {
				log.Fatal(err)
			}

			c.String(http.StatusOK, "%v", rss)
		})
	}

	v1 := router.Group("/api/news/v1")
	{
		v1.GET("/main", func(c *gin.Context) {
			includeText := "%E6%B6%88%E9%98%B2%7C%E7%81%AB%E8%AD%A6%7C%E7%81%AB%E7%81%BD%7C%E7%81%AB%E8%AD%A6%7C%E7%81%AB%E7%87%92%7C%E5%A4%A7%E7%81%AB%7C%E6%95%91%E8%AD%B7%7C%E6%95%91%E7%81%BD%7C%E9%80%81%E9%86%AB%7C%E8%AD%A6%E6%B6%88%7C%E7%BE%A9%E6%B6%88%7C%E8%90%BD%E8%BB%8C%7C%E8%B7%B3%E8%BB%8C%7C%E4%BD%8F%E8%AD%A6%E5%99%A8%7C%E4%BD%8F%E5%AE%85%E8%AD%A6%E5%A0%B1%E5%99%A8%7C%E4%BD%8F%E5%AE%85%E7%81%AB%E8%AD%A6%E5%99%A8%7C%E5%8F%B0%E9%90%B5%E9%A6%99%E5%B1%B1%7C%E9%A6%99%E5%B1%B1%E7%81%AB%E8%BB%8A%E7%AB%99%7C%E9%A6%99%E5%B1%B1%E7%AB%99%7C%E9%9B%B2%E6%A2%AF%7C%E6%B6%88%E9%98%B2.%2A%E9%A6%99%E5%B1%B1%7C%E9%A6%99%E5%B1%B1.%2A%E6%B6%88%E9%98%B2%7CCPR"

			feeds := map[string]string{
				"消防": "https://www.google.com.tw/alerts/feeds/04784784225885481651/1432933957568832221",
				"火燒||火警||火災||大火||住警器||住宅警報器||住宅火警器||義消||落軌||跳軌||台鐵香山||香山火車站||香山站||雲梯||打火": "https://www.google.com.tw/alerts/feeds/04784784225885481651/11834919735038606131",
				"救護":  "https://www.google.com.tw/alerts/feeds/04784784225885481651/10937227332545439311",
				"救災":  "https://www.google.com.tw/alerts/feeds/04784784225885481651/15512682411139935187",
				"送醫":  "https://www.google.com.tw/alerts/feeds/04784784225885481651/7089524768908772692",
				"cpr": "https://www.google.com.tw/alerts/feeds/04784784225885481651/1999534239766046938",
				"蘋果日報社會版":     filterAPIPoint + "filter?url=http%3A%2F%2Fwww.appledaily.com.tw%2Frss%2Fcreate%2Fkind%2Frnews%2Ftype%2F102&include=" + includeText,
				"自由時報社會版":     filterAPIPoint + "filter?url=http%3A%2F%2Fnews.ltn.com.tw%2Frss%2Fsociety.xml&include=" + includeText,
				"聯合新聞社會版":     filterAPIPoint + "filter?url=http%3A%2F%2Fudn.com%2Fudnrss%2Fsocial.xml&include=" + includeText,
				"中國時報社會版":     filterAPIPoint + "filter?url=http%3A%2F%2Fwww.chinatimes.com%2Frss%2Frealtimenews-society.xml&include=" + includeText,
				"蘋果日報國際版":     filterAPIPoint + "filter?url=http%3A%2F%2Fwww.appledaily.com.tw%2Frss%2Fcreate%2Fkind%2Frnews%2Ftype%2F103&include=" + includeText,
				"自由時報國際版":     filterAPIPoint + "filter?url=http%3A%2F%2Fnews.ltn.com.tw%2Frss%2Fworld.xml&include=" + includeText,
				"聯合新聞國際版":     filterAPIPoint + "filter?url=http%3A%2F%2Fudn.com%2Fudnrss%2FBREAKINGNEWS4.xml&include=" + includeText,
				"中國時報國際版":     filterAPIPoint + "filter?url=http%3A%2F%2Fwww.chinatimes.com%2Frss%2Frealtimenews-international.xml&include=" + includeText,
				"民眾日報（記者方詠騰）": filterAPIPoint + "filter?url=http%3A%2F%2Fwww.mypeople.tw%2Frss%2F&include=" + includeText,
				"台灣新生報 地方綜合":  filterAPIPoint + "filter?url=http%3A%2F%2Ffeeds.feedburner.com%2Ftssdnews&include=" + includeText,
				"蘋果日報即時":      filterAPIPoint + "filter?url=http%3A%2F%2Fwww.appledaily.com.tw%2Frss%2Fcreate%2Fkind%2Frnews%2Ftype%2Fnew&include=" + includeText,
			}

			c.JSON(200, gin.H{
				"news": newsFetcher(feeds),
			})
		})
		v1.GET("/city", func(c *gin.Context) {
			includeText := "%E7%AB%B9%E5%B8%82"
			feeds := map[string]string{
				"Google 快訊 竹市||台鐵香山||香山火車站||香山站": "https://www.google.com.tw/alerts/feeds/04784784225885481651/2705564241123909653",
				"中國時報地方版":                        filterAPIPoint + "filter?url=http%3A%2F%2Fwww.chinatimes.com%2Frss%2Fchinatimes-local.xml&include=" + includeText,
				"聯合新聞地方桃竹苗版":                     filterAPIPoint + "filter?url=http%3A%2F%2Fudn.com%2Frssfeed%2Fnews%2F2%2F6641%2F7324%3Fch%3Dnews&include=" + includeText,
				"自由時報地方版":                        filterAPIPoint + "filter?url=http%3A%2F%2Fnews.ltn.com.tw%2Frss%2Flocal.xml&include=" + includeText,
				"蘋果日報地方綜合":                       filterAPIPoint + "filter?url=http%3A%2F%2Fwww.appledaily.com.tw%2Frss%2Fcreate%2Fkind%2Fsec%2Ftype%2F1076&include=" + includeText,
				"中國時報社會版":                        filterAPIPoint + "filter?url=http%3A%2F%2Fwww.chinatimes.com%2Frss%2Frealtimenews-society.xml&include=" + includeText,
				"聯合新聞社會版":                        filterAPIPoint + "filter?url=http%3A%2F%2Fudn.com%2Fudnrss%2Fsocial.xml&include=" + includeText,
				"自由時報社會版":                        filterAPIPoint + "filter?url=http%3A%2F%2Fnews.ltn.com.tw%2Frss%2Fsociety.xml&include=" + includeText,
				"蘋果日報社會版":                        filterAPIPoint + "filter?url=http%3A%2F%2Fwww.appledaily.com.tw%2Frss%2Fcreate%2Fkind%2Frnews%2Ftype%2F102&include=" + includeText,
				"民眾日報（記者方詠騰）":                    filterAPIPoint + "filter?url=http%3A%2F%2Fwww.mypeople.tw%2Frss%2F&include=" + includeText,
				"台灣新生報 地方綜合":                     filterAPIPoint + "filter?url=http%3A%2F%2Ffeeds.feedburner.com%2Ftssdnews&include=" + includeText,
				"蘋果日報 要聞":                        filterAPIPoint + "filter?url=http%3A%2F%2Fwww.appledaily.com.tw%2Frss%2Fcreate%2Fkind%2Fsec%2Ftype%2F11&include=" + includeText,
			}

			c.JSON(200, gin.H{
				"news": newsFetcher(feeds),
			})
		})
		v1.GET("/typhon", func(c *gin.Context) {
			feeds := map[string]string{
				"颱風":     "https://www.google.com.tw/alerts/feeds/04784784225885481651/5973699102355057312",
				"熱帶低氣壓":  "https://www.google.com.tw/alerts/feeds/04784784225885481651/9494720717694166142",
				"輕颱":     "https://www.google.com.tw/alerts/feeds/04784784225885481651/13369455153026830745",
				"中颱":     "https://www.google.com.tw/alerts/feeds/04784784225885481651/13369455153026831531",
				"強颱":     "https://www.google.com.tw/alerts/feeds/04784784225885481651/13369455153026831346",
				"中國時報焦點": filterAPIPoint + "filter?url=http%3A%2F%2Fwww.chinatimes.com%2Frss%2Frealtimenews-focus.xml&include=%E9%A2%B1%E9%A2%A8%7C%E8%BC%95%E9%A2%B1%7C%E4%B8%AD%E9%A2%B1%7C%E5%BC%B7%E9%A2%B1%7C%E7%86%B1%E5%B8%B6%E4%BD%8E%E6%B0%A3%E5%A3%93",
				"聯合新聞最新": filterAPIPoint + "filter?url=http%3A%2F%2Fudn.com%2Fudnrss%2Flatest.xml&include=%E9%A2%B1%E9%A2%A8%7C%E8%BC%95%E9%A2%B1%7C%E4%B8%AD%E9%A2%B1%7C%E5%BC%B7%E9%A2%B1%7C%E7%86%B1%E5%B8%B6%E4%BD%8E%E6%B0%A3%E5%A3%93",
				"自由時報頭版": filterAPIPoint + "filter?url=http%3A%2F%2Fnews.ltn.com.tw%2Frss%2Ffocus.xml&include=%E9%A2%B1%E9%A2%A8%7C%E8%BC%95%E9%A2%B1%7C%E4%B8%AD%E9%A2%B1%7C%E5%BC%B7%E9%A2%B1%7C%E7%86%B1%E5%B8%B6%E4%BD%8E%E6%B0%A3%E5%A3%93",
				"蘋果日報最新": filterAPIPoint + "filter?url=http%3A%2F%2Fwww.appledaily.com.tw%2Frss%2Fcreate%2Fkind%2Frnews%2Ftype%2Fnew&include=%E9%A2%B1%E9%A2%A8%7C%E8%BC%95%E9%A2%B1%7C%E4%B8%AD%E9%A2%B1%7C%E5%BC%B7%E9%A2%B1%7C%E7%86%B1%E5%B8%B6%E4%BD%8E%E6%B0%A3%E5%A3%93",
			}

			c.JSON(200, gin.H{
				"news": newsFetcher(feeds),
			})
		})
		v1.GET("/earthquake", func(c *gin.Context) {
			feeds := map[string]string{
				"地震":     "https://www.google.com.tw/alerts/feeds/04784784225885481651/11159700034107135548",
				"中國時報總覽": filterAPIPoint + "filter?url=http%3A%2F%2Fwww.chinatimes.com%2Frss%2Frealtimenews.xml&include=%E5%9C%B0%E9%9C%87",
				"聯合新聞最新": filterAPIPoint + "filter?url=http%3A%2F%2Fudn.com%2Fudnrss%2Flatest.xml&include=%E5%9C%B0%E9%9C%87",
				"自由時報頭版": filterAPIPoint + "filter?url=http%3A%2F%2Fnews.ltn.com.tw%2Frss%2Ffocus.xml&include=%E5%9C%B0%E9%9C%87",
				"蘋果日報最新": filterAPIPoint + "filter?url=http%3A%2F%2Fwww.appledaily.com.tw%2Frss%2Fcreate%2Fkind%2Frnews%2Ftype%2Fnew&include=%E5%9C%B0%E9%9C%87",
			}

			c.JSON(200, gin.H{
				"news": newsFetcher(feeds),
			})
		})
		v1.GET("/ncdr", func(c *gin.Context) {
			feeds := map[string]string{
				"地震": filterAPIPoint + "filter?type=cap&url=https://alerts.ncdr.nat.gov.tw/RssAtomFeed.ashx?AlertType=6",
			}

			c.JSON(200, gin.H{
				"news": newsFetcher(feeds),
			})
		})
		v1.GET("/hcfd", func(c *gin.Context) {
			includeText := "竹市.*火勢|火勢.*竹市|竹市.*大火|大火.*竹市|竹市.*火災|火災.*竹市|竹市.*火警|火警.*竹市|竹市.*消防|消防.*竹市|竹市.*住警器|住警器.*竹市|竹市.*住宅火警器|住宅火警器.*竹市|竹市.*雲梯|雲梯.*竹市|林智堅.*雲梯|雲梯.*林智堅|消防.*香山|香山.*消防|消防.*林智堅|林智堅.*消防|竹市.*義消|義消.*竹市|義消.*林智堅|林智堅.*義消|竹市.*防災|防災.*竹市|新竹.*淹水|淹水.*新竹|竹市.*淹水|淹水.*竹市|竹市.*CPR|CPR.*竹市|竹市.*AED|AED.*竹市|竹市.*救護|救護.*竹市|竹市.*特搜|特搜.*竹市|竹市.*搶救|搶救.*竹市|竹市.*救援|救援.*竹市|竹市.*警消|警消.*竹市|竹市.*鳳凰志工|鳳凰志工.*竹市|消安.*竹市|竹市.*消安|防火.*竹市|竹市.*防火|竄火.*竹市|竹市.*竄火|被燒.*竹市|竹市.*被燒|中毒.*竹市|竹市.*中毒|竹市.*臥軌|臥軌.*竹市|竹市.*跳軌|跳軌.*竹市|竹市.*落軌|落軌.*竹市|新竹.*臥軌|臥軌.*新竹|新竹.*跳軌|跳軌.*新竹|新竹.*落軌|落軌.*新竹"
			feeds := map[string]string{
				"消防": filterAPIPoint + "filter?url=https%3A%2F%2Fwww.google.com.tw%2Falerts%2Ffeeds%2F04784784225885481651%2F1432933957568832221&include=" + includeText,
				"聯合新聞網（記者王敏旭、林麒偉）":                                                                      filterAPIPoint + "filter?url=http%3A%2F%2Fudn.com%2Frssfeed%2Fnews%2F1%2F2%3Fch%3Dnews&include=" + includeText,
				"自由時報（記者王駿杰、蔡彰盛、洪美秀）":                                                                   filterAPIPoint + "filter?url=http%3A%2F%2Fnews.ltn.com.tw%2Frss%2Fnorthern.xml&include=" + includeText,
				"中時電子報（記者徐養齡、郭芝函）":                                                                      filterAPIPoint + "filter?url=http%3A%2F%2Fwww.chinatimes.com%2Frss%2Frealtimenews-society.xml&include=" + includeText,
				"中時電子報生活版":                                                                              filterAPIPoint + "filter?url=http%3A%2F%2Fwww.chinatimes.com%2Frss%2Frealtimenews-life.xml&include=" + includeText,
				"中時電子報地方版":                                                                              filterAPIPoint + "filter?url=http%3A%2F%2Fwww.chinatimes.com%2Frss%2Frealtimenews-local.xml&include=" + includeText,
				"中央社（記者魯鋼駿）":                                                                            filterAPIPoint + "filter?url=http%3A%2F%2Ffeeds.feedburner.com%2Frsscna%2Flocal&include=" + includeText,
				"勁報（勁報記者羅蔚舟）":                                                                           filterAPIPoint + "filter?url=http%3A%2F%2Fwww.twpowernews.com%2Fhome%2Frss.php&include=" + includeText,
				"真晨報（記者王萱）":                                                                             filterAPIPoint + "filter?url=http%3A%2F%2Ffeeds.feedburner.com%2F5550555&include=" + includeText,
				"臺灣時報（記者鄭銘德）":                                                                           filterAPIPoint + "filter?url=http%3A%2F%2Ffeeds.feedburner.com%2Ftwtimesrss&include=" + includeText,
				"ETtoday（新竹振道記者蔡文綺、記者萬世璉）":                                                              filterAPIPoint + "filter?url=http%3A%2F%2Ffeeds.feedburner.com%2Fettoday%2Flocal&include=" + includeText,
				"民眾日報（記者方詠騰）":                                                                           filterAPIPoint + "filter?url=http%3A%2F%2Fwww.mypeople.tw%2Frss&include=" + includeText,
				"青年日報（記者余華昌）":                                                                           filterAPIPoint + "filter?url=http%3A%2F%2Ffeeds.feedburner.com%2Fgov%2FckHD&include=" + includeText,
				"台灣新聞報（記者戴欣怡）":                                                                          filterAPIPoint + "filter?url=http%3A%2F%2Ffeeds.feedburner.com%2Ftwnewsdaily&include=" + includeText,
				"Google 快訊 竹市||台鐵香山||香山火車站||香山站":                                                        filterAPIPoint + "filter?url=https%3A%2F%2Fwww.google.com.tw%2Falerts%2Ffeeds%2F04784784225885481651%2F2705564241123909653&include=" + includeText,
				"Google 快訊 竹市消防局||勤務派遣科":                                                                filterAPIPoint + "filter?url=https%3A%2F%2Fwww.google.com.tw%2Falerts%2Ffeeds%2F04784784225885481651%2F7890686135979287740&include=" + includeText,
				"Google 快訊 火燒||火警||火災||大火||住警器||住宅警報器||住宅火警器||義消||落軌||跳軌||臥軌||台鐵香山||香山火車站||香山站||雲梯||打火": filterAPIPoint + "filter?url=https%3A%2F%2Fwww.google.com.tw%2Falerts%2Ffeeds%2F04784784225885481651%2F11834919735038606131&include=" + includeText,
				"Google 快訊 竹市 義消": filterAPIPoint + "filter?url=https%3A%2F%2Fwww.google.com.tw%2Falerts%2Ffeeds%2F04784784225885481651%2F18304303068024362009&include=" + includeText,
				"Google 快訊 竹市 雲梯": filterAPIPoint + "filter?url=https%3A%2F%2Fwww.google.com.tw%2Falerts%2Ffeeds%2F04784784225885481651%2F10993434182923813560&include=" + includeText,
				"指傳媒":             filterAPIPoint + "filter?url=http%3A%2F%2Fwww.fingermedia.tw%3Ffeed%3Drss2%26cat%3D2650&include=" + includeText,
				"台灣好報 地方新聞":       filterAPIPoint + "filter?url=http%3A%2F%2Ffeeds.feedburner.com%2Fnewstaiwan&include=" + includeText,
				"台灣新生報 地方綜合":      filterAPIPoint + "filter?url=http%3A%2F%2Ffeeds.feedburner.com%2Ftssdnews&include=" + includeText,
				"天眼日報 警消新聞":       filterAPIPoint + "filter?url=http%3A%2F%2Ffeeds.feedburner.com%2Ftynews3&include=" + includeText,
				//"新竹市政府": filterAPIPoint + "filter?url=http%3A%2F%2Fwww.hccg.gov.tw%2FMunicipalNews%3Flanguage%3Dchinese%26websitedn%3Dou%3Dhccg%2Cou%3Dap_root%2Co%3Dhccg%2Cc%3Dtw&include="+includeText,
				"大成報":          filterAPIPoint + "filter?url=http%3A%2F%2Fwww.greatnews.com.tw%2Fhome%2Frss.php&include=" + includeText,
				"聯合新聞網 地方桃竹苗版": filterAPIPoint + "filter?url=http%3A%2F%2Fudn.com%2Frssfeed%2Fnews%2F2%2F6641%2F7324%3Fch%3Dnews&include=" + includeText,
				"中華新聞網":        filterAPIPoint + "filter?url=http%3A%2F%2Ffeeds.feedburner.com%2Fcdns&include=" + includeText,
				//"蕃新聞": filterAPIPoint + "filter?url=http%3A%2F%2Fn.yam.com%2FRSS%2FRss_society.xml&include="+includeText,
				"蘋果日報 要聞":     filterAPIPoint + "filter?url=http%3A%2F%2Fwww.appledaily.com.tw%2Frss%2Fcreate%2Fkind%2Fsec%2Ftype%2F11&include=" + includeText,
				"自由時報社會版":     filterAPIPoint + "filter?url=http%3A%2F%2Fnews.ltn.com.tw%2Frss%2Fsociety.xml&include=" + includeText,
				"聯合新聞網 即時 地方": filterAPIPoint + "filter?url=http%3A%2F%2Fudn.com%2Frssfeed%2Fnews%2F1%2F3%3Fch%3Dnews&include=" + includeText,
				"風傳媒 新竹頻道":    filterAPIPoint + "filter?url=http%3A%2F%2Fwww.storm.mg%2Ffeeds%2Fs36303&include=" + includeText,
				"自由時報生活版":     filterAPIPoint + "filter?url=http%3A%2F%2Fnews.ltn.com.tw%2Frss%2Flife.xml&include=" + includeText,
				"聯合新聞網 即時 社會": filterAPIPoint + "filter?url=http%3A%2F%2Fudn.com%2Frssfeed%2Fnews%2F1%2F2%3Fch%3Dnews&include=" + includeText,
				"台灣好新聞":       filterAPIPoint + "filter?url=https%3A%2F%2Fwww.google.com.tw%2Falerts%2Ffeeds%2F04784784225885481651%2F3504523367051993014&include=" + includeText,
				"中時電子報 即時 社會": filterAPIPoint + "filter?url=http%3A%2F%2Fwww.chinatimes.com%2Frss%2Frealtimenews-society.xml&include=" + includeText,
				"聯合新聞網 地方":    filterAPIPoint + "filter?url=https%3A%2F%2Fudn.com%2Frssfeed%2Fnews%2F2%2F6641%3Fch%3Dnews&include=" + includeText,
				"蘋果日報 即時":     filterAPIPoint + "filter?url=http%3A%2F%2Fwww.appledaily.com.tw%2Frss%2Fcreate%2Fkind%2Frnews%2Ftype%2Fnew&include=" + includeText,
			}

			c.JSON(200, gin.H{
				"news": newsFetcher(feeds),
			})
		})
	}

	facebookv1 := router.Group("/api/facebook/v1")
	{
		facebookv1.GET("/feed/:id", func(c *gin.Context) {
			include := c.Query("include")
			fbType := c.Query("type")
			appID := "1154770827904156"
			appSecret := "dc0cc2d41255119776b6a9a82ef568c9"

			app := fb.New(appID, appSecret)
			accessToken := appID + "|" + appSecret
			session := app.Session(accessToken)
			id := c.Param("id")
			res, fbErr := session.Get("/"+id+"/feed", nil)
			if fbErr != nil {
				fmt.Println(fbErr)
				return
			}

			paging, _ := res.Paging(session)
			results := paging.Data()

			res, fbErr = session.Get("/"+id, nil)
			if fbErr != nil {
				fmt.Println(fbErr)
				return
			}

			source := res["name"].(string)

			location, loadLocationErr := time.LoadLocation(timeZone)

			collect := []FacebookItem{}
			for _, result := range results {
				local, parseErr := time.Parse(dateTimeFormatFB, fmt.Sprint(result["created_time"]))
				if parseErr != nil {
					local, _ = time.Parse(dateTimeFormatFB, fmt.Sprint(result["updated_time"]))
				}
				if loadLocationErr == nil {
					local = local.In(location)
				}

				id := strings.Split(result["id"].(string), "_")

				var link string
				if fbType == "pg" {
					link = "https://www.facebook.com/permalink.php?story_fbid=" + id[1] + "&id=" + id[0]
				} else {
					link = "https://www.facebook.com/groups/" + id[0] + "/permalink/" + id[1] + "/"
				}

				var story string
				if result["story"] != nil {
					story = result["story"].(string)
				}

				var message string
				if result["message"] != nil {
					message = result["message"].(string)
				}

				fb := FacebookItem{
					Gid:        id[0],
					Pid:        id[1],
					Story:      story,
					Message:    message,
					Link:       link,
					OriginLink: link,
					Time:       local,
					TimeText:   local.Format("15:04"),
					Source:     source,
				}

				collect = append(collect, fb)
			}

			var foundItems = []FacebookItem{}

			rp := regexp.MustCompile(include)
			for _, item := range collect {
				foundMessage := rp.MatchString(CJKnorm(item.Message))

				if foundMessage {
					foundItems = append(foundItems, item)
				}
			}

			c.JSON(200, gin.H{
				"fb": foundItems,
			})
		})
	}

	bloggerv1 := router.Group("/api/blogger/v1")
	{
		bloggerv1.GET("/feed/:id", func(c *gin.Context) {
			includeText := c.Query("include")
			feeds := map[string]string{
				"爆料公社": filterAPIPoint + "filter?url=http%3A%2F%2Fhcfdrss.blogspot.com%2Ffeeds%2Fposts%2Fdefault&include=" + includeText,
			}

			c.JSON(200, gin.H{
				"news": newsFetcher(feeds),
			})
		})
	}

	router.Run() // 0.0.0.0:8080
}
