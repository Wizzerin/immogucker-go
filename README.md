# Immogucker API

Immogucker is an asynchronous Go-based microservice designed to scrape real estate listings from WG-Gesucht. It utilizes a worker pool architecture to process scraping tasks in the background, stores results in a PostgreSQL database, and sends email notifications with the found apartments.

## 🚀 Features
* **Asynchronous Processing:** Uses Go channels and a Worker Pool to handle multiple scraping tasks without blocking the API.
* **Smart Scraping:** Simulates real user behavior with rate limiting and custom headers to avoid bans.
* **Database Integration:** Stores tasks and apartment data using PostgreSQL with fully automated migrations.
* **Graceful Shutdown:** Ensures all workers finish their current tasks and database connections close safely before the service stops.
* **Dockerized:** Fully containerized using Docker and Docker Compose for a seamless setup.

## 🛠 Prerequisites
* [Docker](https://www.docker.com/) and Docker Compose installed on your machine.

## ⚙️ Quick Start

1. **Clone the repository:**
   ```bash
   git clone [https://github.com/Wizzerin/immogucker-go.git](https://github.com/Wizzerin/immogucker-go.git)
   cd immogucker-go
   ```

2. **Configure Environment Variables:**
   Copy the example environment file and fill in your SMTP credentials.
   ```bash
   cp .env.example .env
   ```
   *Note: Open the newly created `.env` file and insert your actual email and App Password (see the section below).*

3. **Start the Service:**
   Build and start the containers in detached mode:
   ```bash
   docker compose up -d --build
   ```

4. **Access the API Documentation:**
   Once the containers are running, open your browser and navigate to the Swagger UI:
   👉 **http://localhost:8080/swagger/index.html**

## 📧 How to Setup Email Notifications (SMTP Password)

To allow the application to send email notifications, you must use an **App Password**, not your regular email password.

**For Gmail Users:**
1. Go to your Google Account Management.
2. Navigate to **Security** -> **2-Step Verification** (must be enabled).
3. Scroll down to **App passwords**.
4. Create a new app password (name it "Immogucker").
5. Copy the generated 16-character password and paste it into `SMTP_PASSWORD` in your `.env` file.

**For Yahoo Users:**
1. Go to your Yahoo Account Security page.
2. Click on **Generate app password** or **Manage app passwords**.
3. Select "Other app" and name it "Immogucker".
4. Copy the generated password into your `.env` file.

## 🛑 Stopping the Service
To stop the application and database, run:
```bash
docker compose down
