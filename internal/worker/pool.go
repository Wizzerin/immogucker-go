package worker

import (
	"database/sql"
	"log"
	"sync"

	"github.com/Wizzerin/immogucker-go/internal/models"
	"github.com/Wizzerin/immogucker-go/internal/notifier"
	"github.com/Wizzerin/immogucker-go/internal/repository"
	"github.com/Wizzerin/immogucker-go/internal/scraper"
)

// StartPool initializes and starts the specified number of worker goroutines
func StartPool(db *sql.DB, taskChan <-chan string, workerCount int, wg *sync.WaitGroup) {
	for i := 1; i <= workerCount; i++ {
		wg.Add(1)
		go worker(i, db, taskChan, wg)
	}
	log.Printf("[Worker Pool] Started: %d workers ready", workerCount)
}

func worker(id int, db *sql.DB, tasks <-chan string, wg *sync.WaitGroup) {
	defer wg.Done() // Ensure the WaitGroup counter is decremented upon exit

	// Worker continuously listens to the tasks channel
	for taskID := range tasks {
		log.Printf("[Worker %d] Picked up task: %s", id, taskID)

		// Update task status to 'processing' in the database
		err := repository.UpdateTaskStatus(db, taskID, "processing")
		if err != nil {
			log.Printf("[Worker %d] Failed to update task status: %v", id, err)
			continue
		}

		taskData, err := repository.GetTaskForWorker(db, taskID)
		if err != nil {
			log.Printf("[Worker %d] Failed to retrieve task data: %v", id, err)
			repository.UpdateTaskStatus(db, taskID, "failed")
			continue
		}

		log.Printf("[Worker %d] Searching for apartments in %s up to %d €", id, taskData.City, taskData.MaxPrice)

		// Execute the scraping process
		providers := []scraper.Provider{scraper.ProviderWGGesucht, scraper.ProviderKleinanzeigen}
		var allResults []models.Apartment

		for _, provider := range providers {
			parser, err := scraper.NewScraper(provider)
			if err != nil {
				log.Printf("[Worker %d] Failed to initialize scraper for %s: %v", id, provider, err)
				continue
			}

			results, err := parser.Parse(taskData, taskID)
			if err != nil {
				log.Printf("[Worker %d] Scraping failed for %s: %v", id, provider, err)
				continue
			}

			allResults = append(allResults, results...)
		}

		if len(allResults) == 0 {
			log.Printf("[Worker %d] No results found across all platforms for %s", id, taskData.City)
			repository.UpdateTaskStatus(db, taskID, "completed")
			continue
		}

		err = repository.SaveApartment(db, allResults)
		if err != nil {
			log.Printf("[Worker %d] Failed to save apartments: %v", id, err)
			repository.UpdateTaskStatus(db, taskID, "failed")
			continue
		}

		// Send email notification with results
		err = notifier.SendResults(taskData.Email, allResults)
		if err != nil {
			log.Printf("[Worker %d] Failed to send email: %v", id, err)
		} else {
			log.Printf("[Worker %d] Notification successfully sent to %s", id, taskData.Email)
		}

		// Mark task as completed
		err = repository.UpdateTaskStatus(db, taskID, "completed")
		if err != nil {
			log.Printf("[Worker %d] Failed to update status to completed: %v", id, err)
			continue
		}

		log.Printf("[Worker %d] Successfully completed task: %s. Apartments found: %d", id, taskID, len(allResults))
	}

	log.Printf("[Worker %d] Shutting down (channel closed)", id)
}
