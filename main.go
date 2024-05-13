package main

import (
	"fmt"
	"log"

	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Proxy struct {
	IPAddress    string
	Port         string
	Code         string
	Country      string
	Anonymity    string
	Google       string
	Https        string
	LastChecked  string
	Working      bool
	ResponseTime time.Duration
}

func main() {
	bot, err := tgbotapi.NewBotAPI("6070162821:AAEj_TzX1D0SrcFrcXeEnwBSny5C4FxxBoc")
	if err != nil {
		log.Fatal("Failed to create new Bot API: ", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates, err := bot.GetUpdatesChan(updateConfig)
	if err != nil {
		log.Fatal("Failed to get updates channel: ", err)
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		documents := fetchProxies()

		proxyListRaw := parseProxiesFromDocuments(documents)

		sendMessage(bot, update.Message.Chat.ID, fmt.Sprintf("%s %d %s", "Found", len(proxyListRaw), "proxies"))

		sentMessageDel := sendMessage(bot, update.Message.Chat.ID, "Finding working proxies...")

		proxyList := checkProxiesConcurrently(proxyListRaw, 500)

		deleteMessage(bot, update.Message.Chat.ID, sentMessageDel.MessageID)

		var proxyInfo string
		for i, proxy := range proxyList {
			proxyInfo += fmt.Sprintf("%d\n%s:%s\n%s %v\nhttps: %s\n\n",
				i+1,
				proxy.IPAddress, proxy.Port,
				proxy.Country, proxy.ResponseTime,
				proxy.Https)
		}

		sendMessage(bot, update.Message.Chat.ID, fmt.Sprintf("%s %d %s", "Found", len(proxyList), "working proxies"))

		sendMessage(bot, update.Message.Chat.ID, proxyInfo)
		// sendMessage(bot, update.Message.Chat.ID, fmt.Sprintf("%s %s", proxyList[0].IPAddress, proxyList[0].Port))

	}
}

// Deletes a message from a specific chat
func deleteMessage(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	deleteMsg := tgbotapi.DeleteMessageConfig{
		ChatID:    chatID,
		MessageID: messageID,
	}

	if _, err := bot.DeleteMessage(deleteMsg); err != nil {
		log.Fatal("Failed to delete message: ", err)
	}
}

// Sends a message to a specific chat and returns the sent message
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) tgbotapi.Message {
	msg := tgbotapi.NewMessage(chatID, text)
	sentMessage, err := bot.Send(msg)
	if err != nil {
		return tgbotapi.Message{}
	}
	return sentMessage
}

func fetchProxies() []*goquery.Document {
	urls := []string{
		"https://www.socks-proxy.net/",
		"https://free-proxy-list.net/",
		"https://www.us-proxy.org/",
		"https://www.sslproxies.org",
		"https://free-proxy-list.net/uk-proxy.html",
		"https://free-proxy-list.net/anonymous-proxy.html",
	}

	// Fetch documents
	docs, err := fetchDocuments(urls)
	if err != nil {
		log.Fatal("Failed to fetch documents: ", err)
	}

	return docs
}

// fetchAndParseProxies fetches and parses proxies concurrently
// parseProxiesFromDocuments fetches and parses proxies concurrently
func parseProxiesFromDocuments(docs []*goquery.Document) []Proxy {
	// Parse proxies from documents
	proxies := make(map[string]Proxy)
	for _, doc := range docs {
		ps, err := parseProxies(doc)
		if err != nil {
			log.Fatal("Failed to parse proxies: ", err)
		}
		for _, p := range ps {
			if _, exists := proxies[p.IPAddress]; !exists {
				proxies[p.IPAddress] = p
			}
		}
	}

	// Convert map back to slice
	uniqueProxies := make([]Proxy, 0, len(proxies))
	for _, p := range proxies {
		uniqueProxies = append(uniqueProxies, p)
	}

	// Concurrently check proxies
	return uniqueProxies
}

// fetchDocuments fetches a list of documents
func fetchDocuments(urls []string) ([]*goquery.Document, error) {
	var docs []*goquery.Document
	for _, url := range urls {
		res, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// parseProxies parses proxies from a document
func parseProxies(doc *goquery.Document) ([]Proxy, error) {
	var proxies []Proxy
	doc.Find("table.table.table-striped.table-bordered").Each(func(i int, tableHtml *goquery.Selection) {
		tableHtml.Find("tr").Each(func(i int, rowHtml *goquery.Selection) {
			cells := rowHtml.Find("td")
			if cells.Length() == 8 {
				proxy := Proxy{
					IPAddress:   strings.TrimSpace(cells.Eq(0).Text()),
					Port:        strings.TrimSpace(cells.Eq(1).Text()),
					Code:        strings.TrimSpace(cells.Eq(2).Text()),
					Country:     strings.TrimSpace(cells.Eq(3).Text()),
					Anonymity:   strings.TrimSpace(cells.Eq(4).Text()),
					Google:      strings.TrimSpace(cells.Eq(5).Text()),
					Https:       strings.TrimSpace(cells.Eq(6).Text()),
					LastChecked: strings.TrimSpace(cells.Eq(7).Text()),
				}
				proxies = append(proxies, proxy)
			}
		})
	})
	return proxies, nil
}

// checkProxiesConcurrently checks proxies concurrently and returns a list of working proxies
func checkProxiesConcurrently(proxies []Proxy, concurrentGoroutines int) []Proxy {
	var wg sync.WaitGroup
	wg.Add(concurrentGoroutines)

	var workingProxies []Proxy
	var mu sync.Mutex

	proxyGroups := make([][]Proxy, concurrentGoroutines)
	for i := range proxies {
		proxyGroups[i%concurrentGoroutines] = append(proxyGroups[i%concurrentGoroutines], proxies[i])
	}

	for i := 0; i < concurrentGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			for _, proxy := range proxyGroups[i] {
				proxy.ResponseTime, proxy.Working = checkProxy(proxy)

				if proxy.Working {
					mu.Lock()
					workingProxies = append(workingProxies, proxy)
					log.Printf("Proxy info: %+v", proxy)
					mu.Unlock()
				}
			}
		}(i)
	}

	wg.Wait()
	if len(workingProxies) > 10 {
		workingProxies = workingProxies[:10]
	}
	// Sort proxies by response time
	sort.Slice(workingProxies, func(i, j int) bool {
		return workingProxies[i].ResponseTime < workingProxies[j].ResponseTime
	})

	return workingProxies
}

// checkProxy checks if a proxy is working and returns its response time and status
func checkProxy(p Proxy) (time.Duration, bool) {
	proxyURL, err := url.Parse(fmt.Sprintf("http://%s:%s", p.IPAddress, p.Port))
	if err != nil {
		fmt.Println("failed to parse proxy URL: %w", err)
		return 0, false
	}

	netTransport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	httpClient := &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}

	// List of websites to check
	websites := []string{
		"https://www.youtube.com/",
		"https://medium.com/",


		// "https://www.linkedin.com/",
	
	}

	start := time.Now()
	for _, website := range websites {
		request, err := http.NewRequest("GET", website, nil)
		if err != nil {
			fmt.Println("failed to create request: %w", err)
			return 0, false
		}

		response, err := httpClient.Do(request)
		if err != nil {
			return time.Since(start), false
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusFound {
			return time.Since(start), false
		}
	}

	fmt.Println("200 status ", proxyURL)
	return time.Since(start), true
}
