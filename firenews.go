package main

import (
	"fmt"
	"hash/fnv"
	"html"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/itsjamie/gin-cors"
	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/urlshortener/v1"
)

const timeZone = "Asia/Taipei"
const dateTimeFormat0 = "2006-01-02T15:04:05Z07:00"
const dateTimeFormat1 = "Mon, 02 Jan 2006 15:04:05 -0700"

var newsSource = map[string]string{
	"twpowernews.com":        "勁報",
	"twreporter.org":         "報導者",
	"mypeople.tw":            "民眾日報",
	"chinatimes.com":         "中時電子報",
	"ettoday.net":            "ETtoday",
	"tvbs.com.tw":            "TVBS",
	"appledaily.com.tw":      "蘋果日報",
	"ftv.com.tw":             "民視新聞",
	"ltn.com.tw":             "自由時報電子報",
	"udn.com":                "聯合新聞網",
	"ebc.net.tw":             "東森新聞",
	"setn.com":               "三立新聞網",
	"pts.org.tw":             "公視新聞",
	"nownews.com":            "NOWnews",
	"mdnkids.com":            "國語日報",
	"cna.com.tw":             "中央通訊社",
	"hinet.net":              "HiNet新聞",
	"storm.mg":               "風傳媒",
	"ttv.com.tw":             "台視新聞",
	"cts.com.tw":             "華視新聞",
	"ithome.com.tw":          "iThome Online",
	"eradio.ner.gov.tw":      "國立教育廣播電台",
	"mradio.com.tw":          "全國廣播",
	"musou.tw":               "沃草國會無雙",
	"anntw.com":              "台灣醒報",
	"thenewslens.com":        "The News Lens關鍵評論網",
	"coolloud.org.tw":        "苦勞網",
	"yam.com":                "蕃新聞",
	"taiwanhot.net":          "台灣好新聞",
	"knowing.asia":           "Knowing",
	"101newsmedia.com":       "一零一傳媒",
	"peopo.org":              "公民新聞",
	"gpwb.gov.tw":            "軍事新聞網",
	"ipcf.org.tw":            "原住民族電視台",
	"ntdtv.com.tw":           "新唐人亞太電視台",
	"cnyes.com":              "鉅亨網",
	"epochtimes.com":         "大紀元",
	"bltv.tv":                "人間衛視",
	"merit-times.com.tw":     "人間福報",
	"tw.on.cc":               "on.cc東網台灣",
	"sina.com.tw":            "臺灣新浪網",
	"ntdtv.com":              "NTDTV",
	"ctitv.com.tw":           "必POTV",
	"travelnews.tw":          "宜蘭新聞網",
	"chinesetoday.com":       "國際日報",
	"gamebase.com.tw":        "遊戲基地",
	"soundofhope.org":        "希望之聲",
	"cdnews.com.tw":          "中央日報",
	"idn.com.tw":             "自立晚報",
	"rti.org.tw":             "中央廣播電台",
	"times-bignews.com":      "今日大話新聞",
	"saydigi.com":            "SayDigi點子生活",
	"e-info.org.tw":          "環境資訊中心",
	"tanews.org.tw":          "台灣動物新聞網",
	"eventsinfocus.org":      "焦點事件",
	"cmmedia.com.tw":         "信傳媒",
	"civilmedia.tw":          "公民行動影音紀錄資料庫",
	"tw.iscarmg.com":         "iscar!",
	"hk.news.yahoo.com":      "Yahoo香港",
	"hk.apple.nextmedia.com": "香港蘋果日報",
	"cdn.org.tw":             "基督教今日報",
	"kairos.news":            "Kairos風向新聞",
	"gov.taipei":             "臺北市政府",
	"yunlin.gov.tw":          "雲林縣政府",
	"taichung.gov.tw":        "臺中市政府",
	"taitung.gov.tw":         "臺東縣政府",
	"hccg.gov.tw":            "新竹市政府",
	"fat.com.tw":             "遠東航空",
	"ctust.edu.tw":           "中臺科技大學",
	"msn.com":                "msn新聞",
	"newtalk.tw":             "新頭殼",
	"cnabc.com":              "中央通訊社",
	"ey.gov.tw":              "行政院全球資訊網",
	"walkerland.com.tw":      "Walkerland",
	"tsna.com.tw":            "tsna",
	"voacantonese.com":       "美國之音",
	"ct.org.tw":              "基督教論壇報",
	"pacificnews.com.tw":     "太平洋新聞網",
	"housefun.com.tw":        "好房網",
	"msntw.com":              "主流傳媒",
	"kinmen.gov.tw":          "金門縣政府",
	"chiayi.gov.tw":          "嘉義市政府",
	"ccu.edu.tw":             "國立中正大學",
	"npa.gov.tw":             "內政部警政署",
	"tainan.gov.tw":          "台南市政府",
	"match.net.tw":           "match生活網",
	"pixnet.net":             "痞客邦",
	"ptt.cc":                 "PTT",
	"xn--4gq171p.com":        "一頁新聞",
	"taiwandaily.net":        "美洲台灣日報",
}

var blockedSource = map[string]bool{
	"bltv.tv":            true,
	"sina.com.tw":        true,
	"epochtimes.com":     true,
	"walkerland.com.tw":  true,
	"tsna.com.tw":        true,
	"pacificnews.com.tw": true,
	"cdnews.com.tw":      true,
	"msntw.com":          true,
	"match.net.tw":       true,
	"pixnet.net":         true,
	"ptt.cc":             true,
	"xn--4gq171p.com":    true,
	"taiwandaily.net":    true,
}

var activedSource = map[string]bool{
	"ltn.com.tw":        true,
	"chinatimes.com":    true,
	"appledaily.com.tw": true,
	"udn.com":           true,
}

// RssItem struct
type RssItem struct {
	Title      string `json:"title"`
	TimeText   string `json:"timeText"`
	Link       string `json:"link"`
	OriginLink string `json:"originLink"`
	Source     string `json:"source"`
	Tag        string `json:"tag"`
	Keyword    string
	Time       time.Time `json:"time"`
	Status     int       `json:"status"`
	Hash       uint32    `json:"hash"`
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
		fmt.Printf("Failed fetch and parse the feed: %s\n", url)
		return collect
	}

	for _, item := range feed.Items {
		local, dateTimeErr := time.Parse(dateTimeFormat0, item.Published)
		if dateTimeErr != nil {
			local, _ = time.Parse(dateTimeFormat1, item.Published)
		}
		location, err := time.LoadLocation(timeZone)
		if err == nil {
			local = local.In(location)
		}

		title := html.UnescapeString(p.Sanitize(item.Title))

		h := fnv.New32a()
		h.Write([]byte(title))
		hashnum := h.Sum32()

		link, originLink := GetURL(item.Link)
		source, keyword := GetNewsSource(item.Link)

		news := RssItem{
			Link:       link,
			OriginLink: originLink,
			Time:       local,
			TimeText:   local.Format("15:04"),
			Title:      title,
			Source:     source,
			Tag:        tag,
			Status:     0,
			Hash:       hashnum,
			Keyword:    keyword,
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
	tmp := make(map[string]RssItem, 0)
	for _, ele := range elements {
		ele.Title = strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, ele.Title)

		tmp[ele.Title] = ele
	}
	var i int
	for _, ele := range tmp {
		elements[i] = ele
		i++
	}
	return elements[:len(tmp)]
}

func titleIsActived(title string, andMode bool) bool {
	var found = 0
	const keywordLength = 2
	keywordList := [keywordLength]string{
		"消防",
		"竹市",
	}

	for i := 0; i < keywordLength; i++ {
		if strings.Contains(title, keywordList[i]) {
			found = found + 1
			if !andMode {
				break
			}
		} else {
			if andMode {
				found = found - 1
			}
		}
	}

	if andMode {
		if found == keywordLength {
			return true
		}
	} else {
		if found > 0 {
			return true
		}
	}

	return false
}

// ActiveElements active elements
func ActiveElements(elements []RssItem) []RssItem {
	for i, item := range elements {
		if _, found := activedSource[item.Keyword]; found {
			elements[i].Status = 1
		} else if titleIsActived(item.Title, true) {
			elements[i].Status = 1
		}
	}
	return elements
}

// CleanupElements makes elements clean
func CleanupElements(elements []RssItem) []RssItem {
	for i := len(elements) - 1; i >= 0; i-- {
		item := elements[i]
		if _, found := blockedSource[item.Keyword]; found {
			elements = append(elements[:i], elements[i+1:]...)
		} else if strings.Contains("關鍵字搜尋搜尋", item.Title) {
			elements = append(elements[:i], elements[i+1:]...)
		}
	}

	return elements
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
			var news [14]([]RssItem)
			news[0] = LoadRSS("消防", "https://www.google.com.tw/alerts/feeds/04784784225885481651/10937227332545439003")
			news[1] = LoadRSS("救護", "https://www.google.com.tw/alerts/feeds/04784784225885481651/10937227332545439311")
			news[2] = LoadRSS("火災", "https://www.google.com.tw/alerts/feeds/04784784225885481651/2277690879891404912")
			news[3] = LoadRSS("送醫", "https://www.google.com.tw/alerts/feeds/04784784225885481651/7089524768908772692")
			news[4] = LoadRSS("cpr", "https://www.google.com.tw/alerts/feeds/04784784225885481651/1999534239766046938")
			news[5] = LoadRSS("蘋果日報社會版", "https://feed.janicek.co/filter?url=http%3A%2F%2Fwww.appledaily.com.tw%2Frss%2Fcreate%2Fkind%2Frnews%2Ftype%2F102&include=%E6%B6%88%E9%98%B2%7C%E6%95%91%E8%AD%B7%7C%E7%81%AB%E8%AD%A6%7C%E7%81%AB%E7%81%BD%7C%E9%80%81%E9%86%AB%7C%E5%A4%A7%E7%81%AB%7Ccpr")
			news[6] = LoadRSS("自由時報社會版", "https://feed.janicek.co/filter?url=http%3A%2F%2Fnews.ltn.com.tw%2Frss%2Fsociety.xml&include=%E6%B6%88%E9%98%B2%7C%E6%95%91%E8%AD%B7%7C%E7%81%AB%E8%AD%A6%7C%E7%81%AB%E7%81%BD%7C%E9%80%81%E9%86%AB%7C%E5%A4%A7%E7%81%AB%7Ccpr")
			news[7] = LoadRSS("聯合新聞社會版", "https://feed.janicek.co/filter?url=http%3A%2F%2Fudn.com%2Fudnrss%2Fsocial.xml&include=%E6%B6%88%E9%98%B2%7C%E6%95%91%E8%AD%B7%7C%E7%81%AB%E8%AD%A6%7C%E7%81%AB%E7%81%BD%7C%E9%80%81%E9%86%AB%7C%E5%A4%A7%E7%81%AB%7Ccpr")
			news[8] = LoadRSS("中國時報社會版", "https://feed.janicek.co/filter?url=http%3A%2F%2Fwww.chinatimes.com%2Frss%2Frealtimenews-society.xml&include=%E6%B6%88%E9%98%B2%7C%E6%95%91%E8%AD%B7%7C%E7%81%AB%E8%AD%A6%7C%E7%81%AB%E7%81%BD%7C%E9%80%81%E9%86%AB%7C%E5%A4%A7%E7%81%AB%7Ccpr")
			news[9] = LoadRSS("蘋果日報國際版", "https://feed.janicek.co/filter?url=http%3A%2F%2Fwww.appledaily.com.tw%2Frss%2Fcreate%2Fkind%2Frnews%2Ftype%2F103&include=%E7%81%AB%E8%AD%A6%7C%E7%81%AB%E7%81%BD%7C%E5%A4%A7%E7%81%AB%7C%E6%95%91%E8%AD%B7%7C%E9%80%81%E9%86%AB%7Ccpr")
			news[10] = LoadRSS("自由時報國際版", "https://feed.janicek.co/filter?url=http%3A%2F%2Fnews.ltn.com.tw%2Frss%2Fworld.xml&include=%E7%81%AB%E8%AD%A6%7C%E7%81%AB%E7%81%BD%7C%E5%A4%A7%E7%81%AB%7C%E6%95%91%E8%AD%B7%7C%E9%80%81%E9%86%AB%7Ccpr")
			news[11] = LoadRSS("聯合新聞國際版", "https://feed.janicek.co/filter?url=http%3A%2F%2Fudn.com%2Fudnrss%2FBREAKINGNEWS4.xml&include=%E7%81%AB%E8%AD%A6%7C%E7%81%AB%E7%81%BD%7C%E5%A4%A7%E7%81%AB%7C%E6%95%91%E8%AD%B7%7C%E9%80%81%E9%86%AB%7Ccpr")
			news[12] = LoadRSS("中國時報國際版", "https://feed.janicek.co/filter?url=http%3A%2F%2Fwww.chinatimes.com%2Frss%2Frealtimenews-international.xml&include=%E7%81%AB%E8%AD%A6%7C%E7%81%AB%E7%81%BD%7C%E5%A4%A7%E7%81%AB%7C%E6%95%91%E8%AD%B7%7C%E9%80%81%E9%86%AB%7Ccpr")
			news[13] = LoadRSS("民眾日報", "https://feed.janicek.co/filter?url=http%3A%2F%2Fwww.mypeople.tw%2Frss%2F&include=%E6%B6%88%E9%98%B2")
			news[0] = append(news[0], news[1]...)
			news[0] = append(news[0], news[2]...)
			news[0] = append(news[0], news[3]...)
			news[0] = append(news[0], news[4]...)
			news[0] = append(news[0], news[5]...)
			news[0] = append(news[0], news[6]...)
			news[0] = append(news[0], news[7]...)
			news[0] = append(news[0], news[8]...)
			news[0] = append(news[0], news[9]...)
			news[0] = append(news[0], news[10]...)
			news[0] = append(news[0], news[11]...)
			news[0] = append(news[0], news[12]...)
			news[0] = append(news[0], news[13]...)
			news[0] = UinqueElements(news[0])
			news[0] = CleanupElements(news[0])
			news[0] = ActiveElements(news[0])
			sort.Sort(ByTime(news[0]))

			c.JSON(200, gin.H{
				"news": news[0],
			})
		})
		v1.GET("/city", func(c *gin.Context) {
			var news [10]([]RssItem)
			news[0] = LoadRSS("竹市", "https://www.google.com.tw/alerts/feeds/04784784225885481651/2705564241123909653")
			news[1] = LoadRSS("中國時報地方版", "https://feed.janicek.co/filter?url=http%3A%2F%2Fwww.chinatimes.com%2Frss%2Fchinatimes-local.xml&include=%E7%AB%B9%E5%B8%82")
			news[2] = LoadRSS("聯合新聞地方桃竹苗版", "https://feed.janicek.co/filter?url=http%3A%2F%2Fudn.com%2Fudnrss%2Flocal_tyhcml.xml&include=%E7%AB%B9%E5%B8%82")
			news[3] = LoadRSS("自由時報地方版", "https://feed.janicek.co/filter?url=http%3A%2F%2Fnews.ltn.com.tw%2Frss%2Flocal.xml&include=%E7%AB%B9%E5%B8%82")
			news[4] = LoadRSS("蘋果日報地方綜合", "https://feed.janicek.co/filter?url=http%3A%2F%2Fwww.appledaily.com.tw%2Frss%2Fcreate%2Fkind%2Fsec%2Ftype%2F1076&include=%E7%AB%B9%E5%B8%82")
			news[5] = LoadRSS("中國時報社會版", "https://feed.janicek.co/filter?url=http%3A%2F%2Fwww.chinatimes.com%2Frss%2Frealtimenews-society.xml&include=%E7%AB%B9%E5%B8%82")
			news[6] = LoadRSS("聯合新聞社會版", "https://feed.janicek.co/filter?url=http%3A%2F%2Fudn.com%2Fudnrss%2Fsocial.xml&include=%E7%AB%B9%E5%B8%82")
			news[7] = LoadRSS("自由時報社會版", "https://feed.janicek.co/filter?url=http%3A%2F%2Fnews.ltn.com.tw%2Frss%2Fsociety.xml&include=%E7%AB%B9%E5%B8%82")
			news[8] = LoadRSS("蘋果日報社會版", "https://feed.janicek.co/filter?url=http%3A%2F%2Fwww.appledaily.com.tw%2Frss%2Fcreate%2Fkind%2Frnews%2Ftype%2F102&include=%E7%AB%B9%E5%B8%82")
			news[9] = LoadRSS("民眾日報", "https://feed.janicek.co/filter?url=http%3A%2F%2Fwww.mypeople.tw%2Frss%2F&include=%E7%AB%B9%E5%B8%82")
			news[0] = append(news[0], news[1]...)
			news[0] = append(news[0], news[2]...)
			news[0] = append(news[0], news[3]...)
			news[0] = append(news[0], news[4]...)
			news[0] = append(news[0], news[5]...)
			news[0] = append(news[0], news[6]...)
			news[0] = append(news[0], news[7]...)
			news[0] = append(news[0], news[8]...)
			news[0] = append(news[0], news[9]...)
			news[0] = UinqueElements(news[0])
			news[0] = CleanupElements(news[0])
			news[0] = ActiveElements(news[0])
			sort.Sort(ByTime(news[0]))

			c.JSON(200, gin.H{
				"news": news[0],
			})
		})
		v1.GET("/typhon", func(c *gin.Context) {
			var news [9]([]RssItem)
			news[0] = LoadRSS("颱風", "https://www.google.com.tw/alerts/feeds/04784784225885481651/5973699102355057312")
			news[1] = LoadRSS("熱帶低氣壓", "https://www.google.com.tw/alerts/feeds/04784784225885481651/9494720717694166142")
			news[2] = LoadRSS("輕颱", "https://www.google.com.tw/alerts/feeds/04784784225885481651/13369455153026830745")
			news[3] = LoadRSS("中颱", "https://www.google.com.tw/alerts/feeds/04784784225885481651/13369455153026831531")
			news[4] = LoadRSS("強颱", "https://www.google.com.tw/alerts/feeds/04784784225885481651/13369455153026831346")
			news[5] = LoadRSS("中國時報焦點", "https://feed.janicek.co/filter?url=http%3A%2F%2Fwww.chinatimes.com%2Frss%2Frealtimenews-focus.xml&include=%E9%A2%B1%E9%A2%A8%7C%E8%BC%95%E9%A2%B1%7C%E4%B8%AD%E9%A2%B1%7C%E5%BC%B7%E9%A2%B1%7C%E7%86%B1%E5%B8%B6%E4%BD%8E%E6%B0%A3%E5%A3%93")
			news[6] = LoadRSS("聯合新聞最新", "https://feed.janicek.co/filter?url=http%3A%2F%2Fudn.com%2Fudnrss%2Flatest.xml&include=%E9%A2%B1%E9%A2%A8%7C%E8%BC%95%E9%A2%B1%7C%E4%B8%AD%E9%A2%B1%7C%E5%BC%B7%E9%A2%B1%7C%E7%86%B1%E5%B8%B6%E4%BD%8E%E6%B0%A3%E5%A3%93")
			news[7] = LoadRSS("自由時報頭版", "https://feed.janicek.co/filter?url=http%3A%2F%2Fnews.ltn.com.tw%2Frss%2Ffocus.xml&include=%E9%A2%B1%E9%A2%A8%7C%E8%BC%95%E9%A2%B1%7C%E4%B8%AD%E9%A2%B1%7C%E5%BC%B7%E9%A2%B1%7C%E7%86%B1%E5%B8%B6%E4%BD%8E%E6%B0%A3%E5%A3%93")
			news[8] = LoadRSS("蘋果日報最新", "https://feed.janicek.co/filter?url=http%3A%2F%2Fwww.appledaily.com.tw%2Frss%2Fcreate%2Fkind%2Frnews%2Ftype%2Fnew&include=%E9%A2%B1%E9%A2%A8%7C%E8%BC%95%E9%A2%B1%7C%E4%B8%AD%E9%A2%B1%7C%E5%BC%B7%E9%A2%B1%7C%E7%86%B1%E5%B8%B6%E4%BD%8E%E6%B0%A3%E5%A3%93")

			news[0] = append(news[0], news[1]...)
			news[0] = append(news[0], news[2]...)
			news[0] = append(news[0], news[3]...)
			news[0] = append(news[0], news[4]...)
			news[0] = append(news[0], news[5]...)
			news[0] = append(news[0], news[6]...)
			news[0] = append(news[0], news[7]...)
			news[0] = append(news[0], news[8]...)
			news[0] = UinqueElements(news[0])
			news[0] = CleanupElements(news[0])
			news[0] = ActiveElements(news[0])
			sort.Sort(ByTime(news[0]))

			c.JSON(200, gin.H{
				"news": news[0],
			})
		})
		v1.GET("/earthquake", func(c *gin.Context) {
			var news [5]([]RssItem)
			news[0] = LoadRSS("地震", "https://www.google.com.tw/alerts/feeds/04784784225885481651/11159700034107135548")
			news[1] = LoadRSS("中國時報總覽", "https://feed.janicek.co/filter?url=http%3A%2F%2Fwww.chinatimes.com%2Frss%2Frealtimenews.xml&include=%E5%9C%B0%E9%9C%87")
			news[2] = LoadRSS("聯合新聞最新", "https://feed.janicek.co/filter?url=http%3A%2F%2Fudn.com%2Fudnrss%2Flatest.xml&include=%E5%9C%B0%E9%9C%87")
			news[3] = LoadRSS("自由時報頭版", "https://feed.janicek.co/filter?url=http%3A%2F%2Fnews.ltn.com.tw%2Frss%2Ffocus.xml&include=%E5%9C%B0%E9%9C%87")
			news[4] = LoadRSS("蘋果日報最新", "https://feed.janicek.co/filter?url=http%3A%2F%2Fwww.appledaily.com.tw%2Frss%2Fcreate%2Fkind%2Frnews%2Ftype%2Fnew&include=%E5%9C%B0%E9%9C%87")

			news[0] = append(news[0], news[1]...)
			news[0] = append(news[0], news[2]...)
			news[0] = append(news[0], news[3]...)
			news[0] = append(news[0], news[4]...)
			news[0] = UinqueElements(news[0])
			news[0] = CleanupElements(news[0])
			news[0] = ActiveElements(news[0])
			sort.Sort(ByTime(news[0]))

			c.JSON(200, gin.H{
				"news": news[0],
			})
		})
		v1.GET("/hcfd", func(c *gin.Context) {
			var news [12]([]RssItem)
			news[0] = LoadRSS("聯合新聞網（記者王敏旭、林麒偉）", "https://feed.janicek.co/filter?url=http%3A%2F%2Fudn.com%2Frssfeed%2Fnews%2F1%2F3%3Fch%3Dnews&include=%E7%AB%B9%E5%B8%82.%2A%E6%B6%88%E9%98%B2%7C%E6%B6%88%E9%98%B2.%2A%E7%AB%B9%E5%B8%82")
			news[1] = LoadRSS("自由時報（記者王駿杰、蔡彰盛、洪美秀）", "https://feed.janicek.co/filter?url=http%3A%2F%2Fnews.ltn.com.tw%2Frss%2Fnorthern.xml&include=%E7%AB%B9%E5%B8%82.%2A%E6%B6%88%E9%98%B2%7C%E6%B6%88%E9%98%B2.%2A%E7%AB%B9%E5%B8%82")
			news[2] = LoadRSS("中時電子報（記者徐養齡、郭芝函）", "https://feed.janicek.co/filter?url=http%3A%2F%2Fwww.chinatimes.com%2Frss%2Fchinatimes-local.xml&include=%E7%AB%B9%E5%B8%82.%2A%E6%B6%88%E9%98%B2%7C%E6%B6%88%E9%98%B2.%2A%E7%AB%B9%E5%B8%82")
			news[3] = LoadRSS("中央社（記者魯鋼駿）", "https://feed.janicek.co/filter?url=http%3A%2F%2Ffeeds.feedburner.com%2Frsscna%2Flocal&include=%E7%AB%B9%E5%B8%82.%2A%E6%B6%88%E9%98%B2%7C%E6%B6%88%E9%98%B2.%2A%E7%AB%B9%E5%B8%82")
			news[4] = LoadRSS("勁報（勁報記者羅蔚舟）", "https://feed.janicek.co/filter?url=http%3A%2F%2Ftwpowernews.com%2Fhome%2Frss.php&include=%E7%AB%B9%E5%B8%82.%2A%E6%B6%88%E9%98%B2%7C%E6%B6%88%E9%98%B2.%2A%E7%AB%B9%E5%B8%82")
			//news[5] = LoadRSS("真晨報-搜尋竹市、消防等關鍵字（記者王萱）", "")
			//news[6] = LoadRSS("臺灣時報（記者鄭銘德）", "")
			news[7] = LoadRSS("ETtoday（新竹振道記者蔡文綺、記者萬世璉）", "https://feed.janicek.co/filter?url=http%3A%2F%2Ffeeds.feedburner.com%2Fettoday%2Flocal&include=%E7%AB%B9%E5%B8%82.%2A%E6%B6%88%E9%98%B2%7C%E6%B6%88%E9%98%B2.%2A%E7%AB%B9%E5%B8%82")
			news[8] = LoadRSS("民眾日報（記者方詠騰）", "https://feed.janicek.co/filter?url=http%3A%2F%2Fwww.mypeople.tw%2Frss&include=%E7%AB%B9%E5%B8%82.%2A%E6%B6%88%E9%98%B2%7C%E6%B6%88%E9%98%B2.%2A%E7%AB%B9%E5%B8%82")
			news[9] = LoadRSS("青年日報（記者余華昌）", "https://feed.janicek.co/filter?url=http%3A%2F%2Fnews.gpwb.gov.tw%2FRss%2F77&include=%E7%AB%B9%E5%B8%82.%2A%E6%B6%88%E9%98%B2%7C%E6%B6%88%E9%98%B2.%2A%E7%AB%B9%E5%B8%82")
			news[10] = LoadRSS("台灣新聞報（記者戴欣怡）", "https://feed.janicek.co/filter?url=http%3A%2F%2Fwww.twnewsdaily.com%2Fhome%2Frss.php&include=%E7%AB%B9%E5%B8%82.%2A%E6%B6%88%E9%98%B2%7C%E6%B6%88%E9%98%B2.%2A%E7%AB%B9%E5%B8%82")
			news[11] = LoadRSS("Google 快訊 - 竹市 消防", "https://www.google.com.tw/alerts/feeds/04784784225885481651/262575485178737034")
			news[0] = append(news[0], news[1]...)
			news[0] = append(news[0], news[2]...)
			news[0] = append(news[0], news[3]...)
			news[0] = append(news[0], news[4]...)
			news[0] = append(news[0], news[5]...)
			news[0] = append(news[0], news[6]...)
			news[0] = append(news[0], news[7]...)
			news[0] = append(news[0], news[8]...)
			news[0] = append(news[0], news[9]...)
			news[0] = append(news[0], news[10]...)
			news[0] = append(news[0], news[11]...)
			news[0] = UinqueElements(news[0])
			news[0] = CleanupElements(news[0])
			news[0] = ActiveElements(news[0])
			sort.Sort(ByTime(news[0]))

			c.JSON(200, gin.H{
				"news": news[0],
			})
		})
	}

	router.Run() // 0.0.0.0:8080
}
