$proxyUri = "http://your.proxy.server:8080"  # Change this to your corporate proxy

$frontingDomain = "somedomain.com"   # The controlled fronting endpoint
$realTarget = "example.com"      # The real target (SNI)
# Create Proxy Handler with Windows Authentication
$handler = New-Object System.Net.Http.HttpClientHandler
$handler.Proxy = New-Object System.Net.WebProxy($proxyUri, $true)
$handler.UseDefaultCredentials = $true  # 🔥 Enables NTLM/Kerberos authentication

# Create HTTP Client
$client = New-Object System.Net.Http.HttpClient($handler)

# Prepare the HTTP request
$request = New-Object System.Net.Http.HttpRequestMessage([System.Net.Http.HttpMethod]::Get, "https://$realTarget")
$request.Headers.Host = $realTarget  # 🔥 Forces SNI to be set correctly

# Send request through the authenticated proxy
$response = $client.SendAsync($request).Result

# Output response
$response.StatusCode
$response.Content.ReadAsStringAsync().Result