param(
    [string]$BaseUrl = "http://localhost:3000"
)

$results = @()
$tempFile = [System.IO.Path]::GetTempFileName()

function Add-TestResult {
    param($CaseName, $ExpectedStatus, $ActualStatus)
    $passed = ($ExpectedStatus -eq $ActualStatus)
    $script:results += [PSCustomObject]@{
        "Kasus Uji"         = $CaseName
        "Status Diharapkan" = $ExpectedStatus
        "Status Didapat"    = $ActualStatus
        "Hasil"             = if ($passed) { "LULUS" } else { "GAGAL" }
    }
}

Write-Host "=== MEMULAI SMOKE TEST OTOMATIS SIAKAD MINI ===" -ForegroundColor Cyan

function Invoke-CurlStatus {
    param(
        [string]$Method,
        [string]$Url,
        [string]$Token = "",
        [string]$Body = ""
    )
    $argsList = @("-s", "-o", $tempFile, "-w", "%{http_code}", "-X", $Method, $Url, "-H", "Content-Type: application/json")
    if ($Token -ne "") {
        $argsList += "-H"
        $argsList += "Authorization: Bearer $Token"
    }
    if ($Body -ne "") {
        $argsList += "-d"
        $argsList += $Body
    }
    $status = & curl.exe @argsList
    return [int]$status.Trim()
}

# Login Admin (200 OK)
$loginAdminBody = '{"email":"admin@siakad.test","password":"Admin12345!"}'
$statusAdmin = Invoke-CurlStatus "POST" "$BaseUrl/api/v1/auth/login" "" $loginAdminBody
$adminLoginJson = Get-Content $tempFile -Raw | ConvertFrom-Json
$adminToken = $adminLoginJson.data.access_token
Add-TestResult "1. Login Admin" 200 $statusAdmin

# Login Mahasiswa Seed (200 OK)
$loginMhsBody = '{"email":"187222000001@student.siakad.test","password":"187222000001"}'
$statusMhs = Invoke-CurlStatus "POST" "$BaseUrl/api/v1/auth/login" "" $loginMhsBody
$mhsLoginJson = Get-Content $tempFile -Raw | ConvertFrom-Json
$mhsToken = $mhsLoginJson.data.access_token
Add-TestResult "2. Login Mahasiswa" 200 $statusMhs

# GET /students tanpa token (401 Unauthorized)
$status401 = Invoke-CurlStatus "GET" "$BaseUrl/api/v1/students"
Add-TestResult "3. GET Students Tanpa Token" 401 $status401

# GET /students oleh Mahasiswa (403 Forbidden)
$status403 = Invoke-CurlStatus "GET" "$BaseUrl/api/v1/students" $mhsToken
Add-TestResult "4. GET Students oleh Mahasiswa" 403 $status403

# GET /students/99999 tidak ditemukan (404 Not Found)
$status404 = Invoke-CurlStatus "GET" "$BaseUrl/api/v1/students/99999" $adminToken
Add-TestResult "5. GET Student ID Tidak Ada" 404 $status404

# POST /students validasi gagal (422 Unprocessable Entity)
$status422 = Invoke-CurlStatus "POST" "$BaseUrl/api/v1/students" $adminToken '{}'
Add-TestResult "6. POST Student Validasi Gagal" 422 $status422

# Buat Mahasiswa Baru (Idempotent pakai Get-Random) & Ambil KRS (201 Created)
$randNim = "18" + (Get-Random -Minimum 1000000000 -Maximum 9999999999).ToString()
$createStudentPayload = "{\`"nim\`":\`"$randNim\`",\`"nama\`":\`"Mahasiswa Test\`",\`"email\`":\`"$randNim@student.siakad.test\`",\`"prodi\`":\`"Teknik Informatika\`",\`"angkatan\`":2024}"
$statusCreateMhs = Invoke-CurlStatus "POST" "$BaseUrl/api/v1/students" $adminToken $createStudentPayload

# Login mahasiswa baru untuk tes KRS
$loginNewMhsBody = "{\`"email\`":\`"$randNim@student.siakad.test\`",\`"password\`":\`"$randNim\`"}"
$statusNewLogin = Invoke-CurlStatus "POST" "$BaseUrl/api/v1/auth/login" "" $loginNewMhsBody
$newMhsJson = Get-Content $tempFile -Raw | ConvertFrom-Json
$newMhsToken = $newMhsJson.data.access_token

$enrollPayload = '{"course_id":1,"tahun_akademik":"2026/2027-Ganjil"}'
$statusEnroll201 = Invoke-CurlStatus "POST" "$BaseUrl/api/v1/enrollments" $newMhsToken $enrollPayload
Add-TestResult "7. POST Enrollments (Ambil KRS)" 201 $statusEnroll201

# POST Enrollments Duplikat (409 Conflict)
$status409 = Invoke-CurlStatus "POST" "$BaseUrl/api/v1/enrollments" $newMhsToken $enrollPayload
Add-TestResult "8. POST Enrollments Duplikat" 409 $status409

# DELETE Enrollments (204 No Content) - Membatalkan KRS
$status204 = Invoke-CurlStatus "DELETE" "$BaseUrl/api/v1/enrollments/1" $newMhsToken
Add-TestResult "9. DELETE Enrollments" 204 $status204

# Kasus 429 Rate Limit Login Gagal (>5x per menit) - DIJALANKAN PALING AKHIR
Write-Host "[INFO] Menjalankan uji 429 Rate Limit (6x login salah beruntun). IP akan dikunci sementara..." -ForegroundColor Yellow
$wrongLoginBody = '{"email":"admin@siakad.test","password":"SalahPassword123"}'
for ($i = 1; $i -le 6; $i++) {
    $status429 = Invoke-CurlStatus "POST" "$BaseUrl/api/v1/auth/login" "" $wrongLoginBody
}
Add-TestResult "10. Rate Limit Login Gagal" 429 $status429

# Cleanup file sementara
Remove-Item $tempFile -ErrorAction SilentlyContinue

# Cetak Tabel Ringkasan
Write-Host "`n=== RINGKASAN HASIL SMOKE TEST ===" -ForegroundColor Green
$results | Format-Table -AutoSize

$failedCount = ($results \vert{} Where-Object {$_.Hasil -eq "GAGAL" }).Count
if ($failedCount -gt 0) {
    Write-Host "Terdapat $failedCount pengujian yang GAGAL." -ForegroundColor Red
    exit 1
} else {
    Write-Host "Semua pengujian BERHASIL (LULUS)!" -ForegroundColor Green
    exit 0
}