param(
  [string]$BaseUrl = "http://127.0.0.1:8080/api",
  [Parameter(Mandatory = $true)][string]$UserName,
  [Parameter(Mandatory = $true)][string]$Password,
  [switch]$PreviewOnly,
  [switch]$SyncAll,
  [int]$MaxSelect = 10
)

$ErrorActionPreference = "Stop"

function Invoke-JsonRequest {
  param(
    [Parameter(Mandatory = $true)][ValidateSet("GET", "POST")] [string]$Method,
    [Parameter(Mandatory = $true)][string]$Url,
    [string]$Token = "",
    [hashtable]$Body = $null
  )

  $headers = @{}
  if ($Token) {
    $headers["token"] = $Token
  }

  if ($Method -eq "GET") {
    return Invoke-RestMethod -Method Get -Uri $Url -Headers $headers -TimeoutSec 90
  }

  if ($Body -ne $null) {
    return Invoke-RestMethod -Method Post -Uri $Url -Headers $headers -Body ($Body | ConvertTo-Json -Depth 10) -ContentType "application/json" -TimeoutSec 180
  }
  return Invoke-RestMethod -Method Post -Uri $Url -Headers $headers -ContentType "application/json" -TimeoutSec 180
}

Write-Host "==> 登录管理员账号并获取 token..." -ForegroundColor Cyan
$loginResp = Invoke-JsonRequest -Method "POST" -Url "$BaseUrl/email_login" -Body @{
  user_name = $UserName
  password  = $Password
}

if (-not $loginResp -or $loginResp.code -ne 0 -or -not $loginResp.data) {
  throw "登录失败：$($loginResp | ConvertTo-Json -Depth 8)"
}
$token = [string]$loginResp.data
Write-Host "登录成功" -ForegroundColor Green

Write-Host "==> 检索最新图片候选..." -ForegroundColor Cyan
$previewResp = Invoke-JsonRequest -Method "GET" -Url "$BaseUrl/settings/site_info/sync_fengfeng_images_preview" -Token $token
if (-not $previewResp -or $previewResp.code -ne 0) {
  throw "图片预览失败：$($previewResp | ConvertTo-Json -Depth 8)"
}

$preview = $previewResp.data
Write-Host ("检索结果：来源总数={0}，扫描图片={1}，可新增={2}，重复={3}，无效={4}" -f `
  $preview.source_total, $preview.latest_scanned, $preview.new_candidate, $preview.duplicate_count, $preview.invalid_count) -ForegroundColor Yellow

if ($preview.candidates) {
  Write-Host "候选示例（前 5 条）：" -ForegroundColor DarkCyan
  $preview.candidates | Select-Object -First 5 | ForEach-Object {
    Write-Host ("- [{0}] {1} ({2})" -f $_.category, $_.url, $_.article_title)
  }
}

if ($PreviewOnly.IsPresent) {
  Write-Host "PreviewOnly 模式，不执行抓取。" -ForegroundColor Yellow
  exit 0
}

$body = @{}
if ($SyncAll.IsPresent) {
  $body.sync_all = $true
  Write-Host "==> 执行一键抓取全部候选图片..." -ForegroundColor Cyan
} else {
  $max = [Math]::Max(1, $MaxSelect)
  $selected = @()
  if ($preview.candidates) {
    $selected = $preview.candidates | Select-Object -First $max | ForEach-Object { $_.url }
  }
  if (-not $selected -or $selected.Count -eq 0) {
    Write-Host "没有可抓取的图片候选，任务结束。" -ForegroundColor Yellow
    exit 0
  }
  $body.image_urls = $selected
  $body.sync_all = $false
  Write-Host ("==> 抓取前 {0} 条候选图片..." -f $selected.Count) -ForegroundColor Cyan
}

$syncResp = Invoke-JsonRequest -Method "POST" -Url "$BaseUrl/settings/site_info/sync_fengfeng_images" -Token $token -Body $body
if (-not $syncResp -or $syncResp.code -ne 0) {
  throw "图片抓取失败：$($syncResp | ConvertTo-Json -Depth 8)"
}

$sync = $syncResp.data
Write-Host ("抓取完成：来源总数={0}，扫描图片={1}，新增={2}，重复={3}，失败={4}，跳过={5}" -f `
  $sync.source_total, $sync.latest_scanned, $sync.created, $sync.duplicate_count, $sync.failed_count, $sync.skipped) -ForegroundColor Green

$syncResp | ConvertTo-Json -Depth 12
