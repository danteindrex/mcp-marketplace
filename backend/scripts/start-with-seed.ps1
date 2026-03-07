param(
  [string]$SuperAdminEmail = "admin@platform.local",
  [string]$SuperAdminPassword = "",
  [string]$BaseUrl = "http://localhost:8080",
  [string]$MongoUri = "mongodb://localhost:27017",
  [string]$MongoDbName = "mcp_marketplace"
)

if (-not $SuperAdminPassword) {
  $SuperAdminPassword = [Guid]::NewGuid().ToString("N") + "!Aa1"
}

if (-not $env:JWT_SECRET) {
  $env:JWT_SECRET = [Guid]::NewGuid().ToString("N")
}

$env:SUPER_ADMIN_EMAIL = $SuperAdminEmail
$env:SUPER_ADMIN_PASSWORD = $SuperAdminPassword
$env:BASE_URL = $BaseUrl
$env:MONGO_URI = $MongoUri
$env:MONGO_DB_NAME = $MongoDbName
$env:ALLOW_INSECURE_DEFAULTS = "false"

Write-Host "Seeding super admin..."
go run ./cmd/seed-super-admin
if ($LASTEXITCODE -ne 0) {
  throw "Super admin seeding failed"
}

Write-Host ""
Write-Host "Super admin credentials:"
Write-Host "  Email:    $SuperAdminEmail"
Write-Host "  Password: $SuperAdminPassword"
Write-Host ""
Write-Host "Starting backend server..."
go run ./cmd/server
