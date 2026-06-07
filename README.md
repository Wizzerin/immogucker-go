# Immogucker API & Dashboard

Immogucker is an asynchronous Go-based microservice and web dashboard designed to scrape real estate listings from multiple platforms (WG-Gesucht and Kleinanzeigen). It utilizes a worker pool architecture to process scraping tasks in the background, stores results in a PostgreSQL database, and sends email notifications with an attached Excel report of the found apartments.

## 🚀 Features
* **Multi-Platform Scraping:** Currently supports WG-Gesucht and Kleinanzeigen, utilizing an extensible Factory Pattern architecture (SOLID principles) to easily integrate new platforms.
* **Interactive Web Dashboard:** Features a lightweight, JavaScript-free web UI built with Go Templates, HTMX, and Pico.css for initiating tasks and viewing real-time status updates.
* **Asynchronous Processing:** Uses Go channels and a Worker Pool to handle multiple scraping tasks concurrently without blocking the API.
* **Smart Filtering & Scraping:** Simulates real user behavior (rate limiting, custom headers) and filters out deactivated or suspicious listings based on dynamic minimum and maximum price thresholds.
* **Excel Reports:** Automatically generates structured `.xlsx` files containing all parsed listings with active hyperlinks, attaching them to email notifications via MIME (`multipart/mixed`).
* **Database Integration:** Stores tasks and apartment data using PostgreSQL with fully automated migrations (`golang-migrate`).
* **Graceful Shutdown:** Ensures all workers finish their current tasks and database connections close safely before the service stops.
* **Dockerized:** Fully containerized using Docker and Docker Compose for a seamless, cross-platform setup.

## 🛠 Prerequisites
* [Docker](https://www.docker.com/) and Docker Compose installed on your machine.

## ⚙️ Quick Start

1. **Clone the repository:**
   ```bash
   git clone [https://github.com/Wizzerin/immogucker-go.git](https://github.com/Wizzerin/immogucker-go.git)
   cd immogucker-go
