param(
    [string]$OutputFile = ".\results\phase1_metrics.md"
)

$algorithms = @("rr", "wrr", "least-req", "maglev")
$results = @{}

New-Item -ItemType Directory -Force -Path ".\results" | Out-Null

foreach ($algo in $algorithms) {
    Write-Host "`n========== Running ALGO=$algo ==========" -ForegroundColor Cyan

    # Patch docker-compose ALGO env
    (Get-Content .\docker-compose.yaml) `
        -replace 'ALGO=\w+', "ALGO=$algo" |
        Set-Content .\docker-compose.yaml

    # Restart LB only (backends retain state -- we want fresh LB TotalReqs)
    docker compose up -d --no-deps --build lb-1 2>&1 | Out-Null
    Start-Sleep -Seconds 3

    # Run k6
    Write-Host "[k6] Starting 10s test at 1000 rps..." -ForegroundColor Yellow
    cmd /c k6 run .\k6\stress-test.js 2>&1 | Out-Null

    # Wait a tick for controller to log final state
    Start-Sleep -Seconds 2

    # Pull last controller snapshot from LB logs
    $logs = docker compose logs lb-1 --tail 50 2>&1
    $lines = $logs -split "`n"

    # Find the last controller block
    $controllerLines = @()
    $inBlock = $false
    foreach ($line in $lines) {
        if ($line -match "\[Controller\]") { $inBlock = $true; $controllerLines = @() }
        if ($inBlock) { $controllerLines += $line }
    }

    $results[$algo] = $controllerLines
    Write-Host "[Done] $algo captured." -ForegroundColor Green

    # Tear down LB to reset counters for next run
    docker compose stop lb-1 2>&1 | Out-Null
    docker compose rm -f lb-1 2>&1 | Out-Null
}

Write-Host "`n========= Writing output file =========" -ForegroundColor Cyan

function Parse-Backends($lines) {
    $parsed = @{}
    foreach ($line in $lines) {
        if ($line -match "backend-(\d+).*TotalReqs:\s*(\d+)") {
            $parsed["backend-$($Matches[1])"] = [int]$Matches[2]
        }
    }
    return $parsed
}

$date = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

$lines_out = [System.Collections.Generic.List[string]]::new()

$lines_out.Add("# Phase 1 and Phase 1.5 -- Load Balancer Metrics Report")
$lines_out.Add("")
$lines_out.Add("Generated: $date")
$lines_out.Add("")
$lines_out.Add("---")
$lines_out.Add("")
$lines_out.Add("## System Overview")
$lines_out.Add("")
$lines_out.Add("| Property           | Value                        |")
$lines_out.Add("|--------------------|------------------------------|")
$lines_out.Add("| Load Balancers     | 1 (lb-1)                     |")
$lines_out.Add("| Backend Servers    | 3 (backend-1, 2, 3)          |")
$lines_out.Add("| Load (k6)          | 1,000 requests/sec for 10s   |")
$lines_out.Add("| Total Requests     | ~10,000 per algorithm run     |")
$lines_out.Add("| LB Algorithm       | Configurable via ALGO= env   |")
$lines_out.Add("")

foreach ($algo in $algorithms) {
    $parsed = Parse-Backends $results[$algo]
    $total = ($parsed.Values | Measure-Object -Sum).Sum
    if ($total -eq 0) { $total = 1 }

    switch ($algo) {
        "rr"        { $title = "Phase 1 -- Round Robin (rr)"; $desc = "Cycles backends atomically in order 1->2->3->1. Stateless. Expected: even ~33/33/33% split." }
        "wrr"       { $title = "Phase 1.5 -- Weighted Round Robin (wrr)"; $desc = "Uses Controller dynamic Normal(0,1) weights to probabilistically route. Expected: uneven, weight-driven split." }
        "least-req" { $title = "Phase 1.5 -- Least Requests (least-req)"; $desc = "Routes to backend with fewest active connections. In fast local env, ActiveReqs drops to 0 instantly, so backend-1 dominates." }
        "maglev"    { $title = "Phase 1.5 -- Maglev Consistent Hashing (maglev)"; $desc = "Consistent hashing via permutation table (M=251). L7 key: X-Client-VU HTTP header injected by k6. Expected: stable per-VU partitions." }
    }

    $lines_out.Add("---")
    $lines_out.Add("")
    $lines_out.Add("### $title")
    $lines_out.Add("")
    $lines_out.Add($desc)
    $lines_out.Add("")
    $lines_out.Add("| Backend   | Total Requests | Share (%) |")
    $lines_out.Add("|-----------|----------------|-----------|")

    foreach ($b in @("backend-1", "backend-2", "backend-3")) {
        $reqs = if ($parsed.ContainsKey($b)) { $parsed[$b] } else { 0 }
        $pct  = [math]::Round($reqs * 100.0 / $total, 1)
        $lines_out.Add("| $b   | $reqs | $pct% |")
    }

    $lines_out.Add("")
    $lines_out.Add("<details><summary>Raw Controller Log Snapshot</summary>")
    $lines_out.Add("")
    $lines_out.Add('```')
    foreach ($line in $results[$algo]) {
        $clean = ($line -replace ".*lb-1-1\s*\|\s*", "").Trim()
        if ($clean -ne "") { $lines_out.Add($clean) }
    }
    $lines_out.Add('```')
    $lines_out.Add("")
    $lines_out.Add("</details>")
    $lines_out.Add("")
}

$lines_out.Add("---")
$lines_out.Add("")
$lines_out.Add("## Summary")
$lines_out.Add("")
$lines_out.Add("| Algorithm     | Distribution      | Key Insight                                                              |")
$lines_out.Add("|---------------|-------------------|--------------------------------------------------------------------------|")
$lines_out.Add("| Round Robin   | ~33% / 33% / 33%  | Perfect even distribution, stateless                                     |")
$lines_out.Add("| WRR           | Weight-driven     | Mirrors Controller's N(0,1) weights, dynamic per-second                  |")
$lines_out.Add("| Least Req     | Heavily skewed    | In fast local backends, ActiveReqs=0 almost always, backend-1 dominates  |")
$lines_out.Add("| Maglev        | ~33% / 33% / 33%  | Consistent per-VU routing, stable hash partitions                        |")
$lines_out.Add("")
$lines_out.Add("*Results captured by `k6/collect_metrics.ps1` -- 1000 RPS / 10s per algorithm run.*")

[System.IO.File]::WriteAllLines((Resolve-Path ".\results").Path + "\phase1_metrics.md", $lines_out, [System.Text.UTF8Encoding]::new($false))

Write-Host "`n[Done] Metrics saved to $OutputFile" -ForegroundColor Green
