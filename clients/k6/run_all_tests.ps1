$tests = @(
    "test_flow_consistency",
    "test_http11_hol",
    "test_http2_mux",
    "test_sticky_rules",
    "test_realistic_load",
    "test_health_failover",
    "test_sticky_failover"
)

New-Item -ItemType Directory -Force -Path ".\results" | Out-Null
$report = ".\results\phase1_l7_tests.md"

[System.IO.File]::WriteAllText(
    (Join-Path (Get-Location) "results\phase1_l7_tests.md"),
    "# Phase 1.6 - L7 Test Results`nGenerated: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')`n`n",
    [System.Text.UTF8Encoding]::new($false)
)

foreach ($name in $tests) {
    $file = Join-Path $PSScriptRoot "${name}.js"
    Write-Host ""
    Write-Host "===== $name =====" -ForegroundColor Cyan

    $output = (cmd /c k6 run $file 2>&1) -join "`n"
    Write-Host $output

    $exitCode = $LASTEXITCODE
    $status = if ($exitCode -eq 0) { "PASSED" } else { "FAILED (exit $exitCode)" }
    Write-Host "Result: $status" -ForegroundColor $(if ($exitCode -eq 0) { "Green" } else { "Red" })

    $section = "---`n### $name -- $status`n`n" + '```' + "`n$output`n" + '```' + "`n`n"
    [System.IO.File]::AppendAllText(
        (Join-Path (Get-Location) "results\phase1_l7_tests.md"),
        $section,
        [System.Text.UTF8Encoding]::new($false)
    )
}

Write-Host ""
Write-Host "All tests complete. Report: $report" -ForegroundColor Yellow
