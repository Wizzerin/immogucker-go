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

type WGGesuchtScraper struct{}

var wgCityIDs = map[string]string{
	"Neuss":       "224",
	"Düsseldorf":  "30",
	"Duesseldorf": "30",
	"Köln":        "73",
	"Koeln":       "73",
	"Berlin":      "8",
}

func (s *WGGesuchtScraper) Parse(city string, minPrice, maxPrice int, taskID string) ([]models.Apartment, error) {
	var apartments []models.Apartment

	var targetCityName string
	var targetCityID string

	if id, exists := wgCityIDs[city]; exists {
		targetCityName = city
		targetCityID = id
	} else {
		for name, id := range wgCityIDs {
			if city == id {
				targetCityName = name
				targetCityID = id
				break
			}
		}
	}

	if targetCityID == "" {
		return nil, fmt.Errorf("city %s is not supported by the parser (unknown ID or Name)", city)
	}

	c := colly.NewCollector(
		colly.AllowedDomains("www.wg-gesucht.de", "wg-gesucht.de"),
	)

	// Configure limits to simulate real user behavior and avoid IP bans
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*wg-gesucht.de*",
		Delay:       3 * time.Second,
		RandomDelay: 2 * time.Second,
	})

	// Set headers to bypass basic bot protection
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		r.Headers.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8")
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
		log.Printf("[Scraper] Requesting: %s", r.URL.String())
	})

	// Search for listing cards (selectors might require periodic updates)
	// Trigger on the broad .wgg_card class
	c.OnHTML(".wgg_card", func(e *colly.HTMLElement) {
		titleEl := e.DOM.Find("h2.truncate_title a")
		title := strings.TrimSpace(titleEl.Text())
		link, _ := titleEl.Attr("href")

		priceStr := e.DOM.Find(".middle .col-xs-3 b").First().Text()

		if title == "" || link == "" || priceStr == "" {
			return // Skip empty blocks or advertisement banners
		}

		// Parse price: find the first bold text in the middle block
		priceRaw := e.DOM.Find(".middle b").First().Text()
		if priceRaw == "" {
			priceRaw = e.DOM.Find(".card_body b").First().Text()
		}

		// Extract strictly numeric digits
		priceClean := extractDigits(priceRaw)
		price, err := strconv.Atoi(priceClean)
		if err != nil {
			return // Skip if price parsing fails
		}

		// Filter results on the fly to save memory
		if price < minPrice || price > maxPrice {
			log.Printf("[Parser] Apartment filtered out by price (%d € > %d €): %s", price, maxPrice, title)
			return
		}

		// Ensure the link is an absolute URL
		if !strings.HasPrefix(link, "http") {
			link = "https://www.wg-gesucht.de" + link
		}

		apartments = append(apartments, models.Apartment{
			TaskID: taskID,
			Title:  title,
			Price:  price,
			Link:   link,
		})
	})

	// Build the target URL using the mapped city ID
	searchURL := fmt.Sprintf("https://www.wg-gesucht.de/wohnungen-in-%s.%s.2.1.0.html?rent_types[0]=1&rent_types[1]=2", targetCityName, targetCityID)
	log.Printf("[Parser] Starting data collection from URL: %s", searchURL)

	err := c.Visit(searchURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	return apartments, nil
}
