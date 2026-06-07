package scraper

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Wizzerin/immogucker-go/internal/models"
	"github.com/gocolly/colly/v2"
)

type KleinanzeigenScraper struct{}

type kaCityData struct {
	Slug string
	ID   string
}

var kaCityIDs = map[string]kaCityData{
	"Neuss":       {"neuss", "c203l2108"},
	"Düsseldorf":  {"duesseldorf", "c203l2056"},
	"Duesseldorf": {"duesseldorf", "c203l2056"},
	"Köln":        {"koeln", "c203l945"},
	"Koeln":       {"koeln", "c203l945"},
	"Berlin":      {"berlin", "c203l3331"},
}

func (s *KleinanzeigenScraper) Parse(task models.WorkerTask, taskID string) ([]models.Apartment, error) {
	var apartments []models.Apartment

	var targetCity kaCityData
	if data, exists := kaCityIDs[task.City]; exists {
		targetCity = data
	} else {
		for _, data := range kaCityIDs {
			if task.City == data.ID {
				targetCity = data
				break
			}
		}
	}

	if targetCity.Slug == "" {
		return nil, fmt.Errorf("city %s is not supported by Kleinanzeigen scraper", task.City)
	}

	c := colly.NewCollector(
		colly.AllowedDomains("www.kleinanzeigen.de", "kleinanzeigen.de"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*kleinanzeigen.de*",
		Delay:       4 * time.Second,
		RandomDelay: 3 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		r.Headers.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8")
		log.Printf("[Kleinanzeigen] Requesting: %s", r.URL.String())
	})

	c.OnHTML("article.aditem", func(e *colly.HTMLElement) {
		titleEl := e.DOM.Find(".aditem-main--middle h2 a")
		title := strings.TrimSpace(titleEl.Text())
		link, _ := titleEl.Attr("href")

		priceRaw := e.DOM.Find(".aditem-main--middle--price-shipping--price").Text()

		if title == "" || link == "" || priceRaw == "" {
			return
		}

		priceClean := extractDigits(priceRaw)
		if priceClean == "" {
			return
		}

		price, err := strconv.Atoi(priceClean)
		if err != nil {
			return
		}

		if price < task.MinPrice || price > task.MaxPrice {
			log.Printf("[Kleinanzeigen] Filtered out by price (%d €): %s", price, title)
			return
		}

		if !strings.HasPrefix(link, "http") {
			link = "https://www.kleinanzeigen.de" + link
		}

		apartments = append(apartments, models.Apartment{
			TaskID: taskID,
			Title:  "[KA] " + title,
			Price:  price,
			Link:   link,
		})
	})

	modifiers := "+wohnung_mieten.swap_s:nein" // Исключаем обмен

	if task.MinSize > 0 || task.MaxSize > 0 {
		max := ""
		if task.MaxSize > 0 {
			max = strconv.Itoa(task.MaxSize)
		}
		modifiers += fmt.Sprintf("+wohnung_mieten.qm_d:%d,%s", task.MinSize, max)
	}

	if task.MinRooms > 0 || task.MaxRooms > 0 {
		max := ""
		if task.MaxRooms > 0 {
			max = strconv.Itoa(task.MaxRooms)
		}
		modifiers += fmt.Sprintf("+wohnung_mieten.zimmer_d:%d,%s", task.MinRooms, max)
	}

	searchURL := fmt.Sprintf("https://www.kleinanzeigen.de/s-wohnung-mieten/%s/preis:%d:%d/%s%s", targetCity.Slug, task.MinPrice, task.MaxPrice, targetCity.ID, modifiers)
	log.Printf("[Parser] Starting data collection from URL: %s", searchURL)

	err := c.Visit(searchURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	return apartments, nil
}
