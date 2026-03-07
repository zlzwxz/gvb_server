param(
  [string]$BaseUrl = "http://127.0.0.1:8080/api",
  [Parameter(Mandatory = $true)][string]$UserName,
  [Parameter(Mandatory = $true)][string]$Password,
  [switch]$PreviewOnly,
  [switch]$SyncAll,
  [int]$MaxSelect = 10,
  [int]$Limit = 200
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
    return Invoke-RestMethod -Method Get -Uri $Url -Headers $headers -TimeoutSec 60
  }

  if ($Body -ne $null) {
    return Invoke-RestMethod -Method Post -Uri $Url -Headers $headers -Body ($Body | ConvertTo-Json -Depth 8) -ContentType "application/json" -TimeoutSec 120
  }
  return Invoke-RestMethod -Method Post -Uri $Url -Headers $headers -ContentType "application/json" -TimeoutSec 120
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

Write-Host "==> 检索最新可抓取文章数量..." -ForegroundColor Cyan
if ($Limit -lt 0) {
  $safeLimit = 200
} elseif ($Limit -eq 0) {
  $safeLimit = 0
} else {
  $safeLimit = [Math]::Min($Limit, 10000)
}
$previewResp = Invoke-JsonRequest -Method "GET" -Url "$BaseUrl/settings/site_info/sync_fengfeng_preview?limit=$safeLimit" -Token $token
if (-not $previewResp -or $previewResp.code -ne 0) {
  throw "预览失败：$($previewResp | ConvertTo-Json -Depth 8)"
}

$preview = $previewResp.data
Write-Host ("检索结果：来源总数={0}，本次扫描={1}，可新增={2}，重复={3}，无效={4}" -f `
  $preview.source_total, $preview.latest_scanned, $preview.new_candidate, $preview.duplicate_count, $preview.invalid_count) -ForegroundColor Yellow
if ($preview.candidates) {
  Write-Host "候选文章示例（前 5 条）：" -ForegroundColor DarkCyan
  $preview.candidates | Select-Object -First 5 | ForEach-Object {
    Write-Host ("- {0} ({1})" -f $_.title, $_.article_id)
  }
}

if ($PreviewOnly.IsPresent) {
  Write-Host "PreviewOnly 模式，不执行入库同步。" -ForegroundColor Yellow
  exit 0
}

if ($SyncAll.IsPresent) {
  Write-Host "==> 开始一键抓取全部候选文章..." -ForegroundColor Cyan
  $syncBody = @{ sync_all = $true; include_update = $true; limit = $safeLimit }
} else {
  $max = [Math]::Max(1, $MaxSelect)
  $selectedIDs = @()
  if ($preview.candidates) {
    $selectedIDs = $preview.candidates | Select-Object -First $max | ForEach-Object { $_.article_id }
  }
  if (-not $selectedIDs -or $selectedIDs.Count -eq 0) {
    Write-Host "没有可抓取的新文章，任务结束。" -ForegroundColor Yellow
    exit 0
  }
  Write-Host ("==> 开始抓取选中的 {0} 篇文章..." -f $selectedIDs.Count) -ForegroundColor Cyan
  $syncBody = @{ article_ids = $selectedIDs; sync_all = $false; include_update = $true; limit = $safeLimit }
}

$syncResp = Invoke-JsonRequest -Method "POST" -Url "$BaseUrl/settings/site_info/sync_fengfeng" -Token $token -Body $syncBody
if (-not $syncResp -or $syncResp.code -ne 0) {
  throw "同步失败：$($syncResp | ConvertTo-Json -Depth 8)"
}

$sync = $syncResp.data
Write-Host ("同步完成：来源总数={0}，本次扫描={1}，新增={2}，已更新={3}，重复={4}，失败={5}，跳过={6}" -f `
  $sync.source_total, $sync.latest_scanned, $sync.created, $sync.updated_count, $sync.duplicate_count, $sync.failed_count, $sync.skipped) -ForegroundColor Green

# 输出完整 JSON 结果，便于排查线上问题。
$syncResp | ConvertTo-Json -Depth 12
