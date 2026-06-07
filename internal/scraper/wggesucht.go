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

type wgCityData struct {
	Slug string
	ID   string
}

var wgCityIDs = map[string]wgCityData{
	"Neuss":       {"Neuss", "224"},
	"Düsseldorf":  {"Duesseldorf", "30"},
	"Duesseldorf": {"Duesseldorf", "30"},
	"Köln":        {"Koeln", "73"},
	"Koeln":       {"Koeln", "73"},
	"Berlin":      {"Berlin", "8"},
}

func (s *WGGesuchtScraper) Parse(task models.WorkerTask, taskID string) ([]models.Apartment, error) {
	var apartments []models.Apartment

	var targetCity wgCityData
	if data, exists := wgCityIDs[task.City]; exists {
		targetCity = data
	} else {
		for _, data := range wgCityIDs {
			if task.City == data.ID {
				targetCity = data
				break
			}
		}
	}

	if targetCity.Slug == "" {
		return nil, fmt.Errorf("city %s is not supported by WG-Gesucht scraper", task.City)
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
		if price < task.MinPrice || price > task.MaxPrice {
			log.Printf("[Parser] Apartment filtered out by price (%d € > %d €): %s", price, task.MaxPrice, title)
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

	queryParams := "?rent_types[0]=1&rent_types[1]=2"

	if task.MaxSize > 0 || task.MaxSize > 0 {
		maxSize := task.MaxSize
		if maxSize == 0 {
			maxSize = 999
		}
		queryParams += fmt.Sprintf("&min_size=%d&max_size=%d", task.MinSize, task.MaxSize)
	}

	if task.MinRooms > 0 || task.MaxRooms > 0 {
		maxRooms := task.MaxRooms
		if maxRooms == 0 {
			maxRooms = 99
		}
		queryParams += fmt.Sprintf("&rmMin=%d&rmMax=%d", task.MinRooms, maxRooms)
	}

	// Build the target URL using the mapped city ID
	searchURL := fmt.Sprintf("https://www.wg-gesucht.de/wohnungen-in-%s.%s.2.1.0.html%s", targetCity.Slug, targetCity.ID, queryParams)
	log.Printf("[Parser] Starting data collection from URL: %s", searchURL)

	err := c.Visit(searchURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	return apartments, nil
}
