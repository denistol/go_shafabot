package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func getProducts(url string, c context.Context) ProductsMap {
	var prefix = "https://shafa.ua"

	var links []*cdp.Node
	err := chromedp.Run(c,
		chromedp.Navigate(url),
		// Scroll down
		chromedp.Sleep(1*time.Second),
		chromedp.KeyEvent(kb.End),
		chromedp.Sleep(3*time.Second),
		// Get links
		chromedp.Nodes("a", &links, chromedp.ByQueryAll),
	)
	if err != nil {
		log.Fatal("Error:", err)
	}

	var products = make(map[int]ProductData)

	for _, node := range links {
		href := node.AttributeValue("href")
		regex := regexp.MustCompile(`/\d{6,}`)
		m := regex.FindAllString(href, 1)
		if !strings.Contains(href, "https://") && len(m) == 1 {
			Id, err := strconv.Atoi(strings.ReplaceAll(m[0], "/", ""))
			if err == nil {
				var Url = prefix + href
				var Timestamp = time.Now().Unix()
				var p = ProductData{Timestamp, Url}
				products[Id] = p
			}
		}
	}

	return products
}

func main() {
	var cfg = ConfigConstructor()
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Panic(err)
	}

	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)
	defer cancel()

	for {
		for _, u := range cfg.Urls {
			var detectedProducts = getProducts(u, ctx)
			var newProducts = cfg.handleProducts(detectedProducts)

			var detectedProductsCount = len(detectedProducts)
			var newProductsCount = len(newProducts)

			fmt.Println("Checked: ", u)
			fmt.Println("Detected: ", detectedProductsCount)
			fmt.Println("New: ", newProductsCount)
			fmt.Println("---------------")

			if newProductsCount != 0 {
				for _, v := range newProducts {
					for _, id := range cfg.ChatIds {
						msg := tgbotapi.NewMessage(id, v.Url)
						bot.Send(msg)
					}
				}
			}
		}
		time.Sleep(time.Duration(cfg.Timeout) * time.Second)
	}

}
