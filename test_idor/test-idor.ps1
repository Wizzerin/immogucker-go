$BaseUrl = "http://localhost:8080/api/v1"

# IMPORTANT: Change these to real users that exist in your database
$UserA_Creds = @{ email="user1@example.com"; password="123" }
$UserB_Creds = @{ email="user2@example.com"; password="123" }

Clear-Host
Write-Host "=== START SECURITY TEST (IDOR) ===" -ForegroundColor Cyan

# 1. Authenticate User A
Write-Host "`n[1] Authenticating User A..."
try {
    # FIX: Removed ConvertTo-Json and ContentType. This forces application/x-www-form-urlencoded
    $loginResponseA = Invoke-WebRequest -Uri "$BaseUrl/auth/login" -Method Post -Body $UserA_Creds -SessionVariable sessionA
    Write-Host "User A session acquired." -ForegroundColor Green
} catch {
    Write-Host "Failed to login User A. Check credentials. $($_.Exception.Message)" -ForegroundColor Red
    exit
}

# 2. Create task for User A
Write-Host "`n[2] Creating task for User A..."
$taskBody = @{ city="Berlin"; min_price=500; max_price=1000 } | ConvertTo-Json
try {
    $createResponse = Invoke-RestMethod -Uri "$BaseUrl/tasks" -Method Post -Body $taskBody -ContentType "application/json" -WebSession $sessionA
    $taskId = $createResponse.task_id
    Write-Host "Task created! UUID: $taskId" -ForegroundColor Green
} catch {
    Write-Host "Failed to create task: $($_.Exception.Message)" -ForegroundColor Red
    exit
}

# 3. User A reads own task
Write-Host "`n[3] User A fetching own task..."
try {
    $statusResponse = Invoke-RestMethod -Uri "$BaseUrl/tasks/$taskId" -Method Get -WebSession $sessionA
    Write-Host "Success! Task found. Status: $($statusResponse.status)" -ForegroundColor Green
} catch {
    Write-Host "Error: User A could not get own task: $($_.Exception.Message)" -ForegroundColor Red
    exit
}

# 4. Authenticate User B
Write-Host "`n[4] Authenticating User B (Attacker)..."
try {
    $loginResponseB = Invoke-WebRequest -Uri "$BaseUrl/auth/login" -Method Post -Body $UserB_Creds -SessionVariable sessionB
    Write-Host "User B session acquired." -ForegroundColor Green
} catch {
    Write-Host "Failed to login User B. Check credentials. $($_.Exception.Message)" -ForegroundColor Red
    exit
}

# 5. IDOR Test: User B tries to read User A's task
Write-Host "`n[5] WARNING: User B attempting to access User A's task ($taskId)..." -ForegroundColor Yellow
try {
    $hackResponse = Invoke-RestMethod -Uri "$BaseUrl/tasks/$taskId" -Method Get -WebSession $sessionB

    # If it reaches here, the server returned 200 OK (VULNERABILITY)
    Write-Host "CRITICAL VULNERABILITY (IDOR)! User B accessed User A's task!" -ForegroundColor Red
    Write-Host ($hackResponse | ConvertTo-Json) -ForegroundColor Red
} catch {
    # Check if the server correctly blocked access
    $response = $_.Exception.Response
    $statusCode = $response.StatusCode.value__

    if ($statusCode -eq 404 -or $statusCode -eq 401) {
        Write-Host "TEST PASSED! Server hid the task (Status $statusCode)." -ForegroundColor Green
    } else {
        Write-Host "Server returned error: $statusCode. Message: $($_.Exception.Message)" -ForegroundColor Yellow
    }
}

Write-Host "`n=== TEST COMPLETE ===" -ForegroundColor Cyan
Read-Host -Prompt "Press Enter to exit..."
