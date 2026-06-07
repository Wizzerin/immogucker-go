# Immogucker API & Dashboard

Immogucker is a secure, asynchronous Go-based microservice and web dashboard designed to scrape real estate listings from multiple platforms (WG-Gesucht and Kleinanzeigen). It utilizes a worker pool architecture to process scraping tasks in the background, stores results in a PostgreSQL database, and sends email notifications with an attached Excel report of the found apartments.

## 🚀 Features
* **Authentication & Data Isolation:** Secure user registration, unique usernames, and session-based login using HttpOnly cookies. Features strict data isolation (protection against IDOR), ensuring users can only access and manage their own scraping tasks.
* **Email Verification:** Mandates email confirmation via SMTP before unlocking the scraper functionality, ensuring high-quality user data and preventing spam.
* **Interactive Web Dashboard:** A lightweight, JavaScript-free web UI built with Go Templates, HTMX, and Pico.css for initiating tasks and viewing real-time status updates. Fully protected by authentication middleware.
* **Multi-Platform Scraping:** Currently supports WG-Gesucht and Kleinanzeigen, utilizing an extensible Factory Pattern architecture (SOLID principles) to easily integrate new platforms.
* **Asynchronous Processing:** Uses Go channels and a Worker Pool to handle multiple scraping tasks concurrently without blocking the API.
* **Smart Filtering & Scraping:** Simulates real user behavior (rate limiting, custom headers) and filters out deactivated or suspicious listings. Supports dynamic filtering by budget (min/max price), apartment size (min/max m²), and room count (min/max). Includes strict filtering to exclude "Tauschangebote" (swap offers).
* **Excel Reports:** Automatically generates structured `.xlsx` files containing all parsed listings with active hyperlinks, attaching them to email notifications via MIME (`multipart/mixed`).
* **Database Integration:** Stores users, sessions, tasks, and apartment data using PostgreSQL with fully automated migrations (`golang-migrate`).
* **Graceful Shutdown:** Ensures all workers finish their current tasks and database connections close safely before the service stops.
* **Dockerized:** Fully containerized using Docker and Docker Compose for a seamless, cross-platform setup.

## 🛠 Prerequisites
* [Docker](https://www.docker.com/) and Docker Compose installed on your machine.

## ⚙️ Quick Start

1. **Clone the repository:**
   ```bash
   git clone [https://github.com/Wizzerin/immogucker-go.git](https://github.com/Wizzerin/immogucker-go.git)
   cd immogucker-go
Configure Environment Variables:
Copy the example environment file and fill in your SMTP credentials.

Bash
cp .env.example .env
Note: Open the newly created .env file and insert your actual email and App Password (see the section below).

Start the Service:
Build and start the containers in detached mode:

Bash
docker compose up -d --build
Access the Application:
Once the containers are running, open your browser. You will need to register an account and verify your email to access the dashboard tools.

Web Dashboard (Login): http://localhost:8080/login

Swagger API Docs: http://localhost:8080/swagger/index.html

🛡️ Security Testing (IDOR)
To prove the effectiveness of the implemented session-based authentication and strict data isolation, the repository includes an automated PowerShell test script. This script verifies that the API is not vulnerable to Insecure Direct Object Reference (IDOR).

How to run the test:

Ensure the Docker containers are running.

Register two distinct test accounts via the Web UI (http://localhost:8080/register).

Open the file test_idor/test-idor.ps1 in a text editor.

Replace the credentials in $UserA_Creds and $UserB_Creds with the newly registered accounts.

Run the script using PowerShell:

PowerShell
cd test_idor
.\test-idor.ps1
The script will authenticate both users, create a task for User A, and attempt to access it using User B's session. A successful test will result in a green TEST PASSED! (Status 404) message.

📧 How to Setup Email Notifications (SMTP Password)
To allow the application to send verification links and email notifications with Excel reports, you must use an App Password, not your regular email password.

For Gmail Users:

Go to your Google Account Management.

Navigate to Security -> 2-Step Verification (must be enabled).

Scroll down to App passwords.

Create a new app password (name it "Immogucker").

Copy the generated 16-character password and paste it into SMTP_PASSWORD in your .env file.

For Yahoo Users:

Go to your Yahoo Account Security page.

Click on Generate app password or Manage app passwords.

Select "Other app" and name it "Immogucker".

Copy the generated password into your .env file.

🛑 Stopping the Service
To stop the application, shut down the worker pool, and halt the database, run:

Bash
docker compose down
