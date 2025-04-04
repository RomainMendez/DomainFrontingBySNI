$proxyHost = "2.2.2.2"
$proxyPort = 8080

$frontingDomain = "somedomain.com"   # The controlled fronting endpoint
$realTarget = "example.com"      # The real target (SNI)

# Create a TCP connection to the Proxy
$client = New-Object System.Net.Sockets.TcpClient($proxyHost, $proxyPort)
$stream = $client.GetStream()
$writer = New-Object System.IO.StreamWriter($stream)
$reader = New-Object System.IO.StreamReader($stream)

# Send an HTTP CONNECT request to establish a tunnel to "bad.com"
$connectRequest = "CONNECT $frontingDomain`:443 HTTP/1.1`r`n" +
                  "Host: $frontingDomain`r`n" +
                  "`r`n"

$writer.Write($connectRequest)
$writer.Flush()

# Read Proxy Response
$proxyResponse = $reader.ReadLine()
if (-not $proxyResponse.Contains("200")) {
    Write-Host "Proxy failed to establish connection: $proxyResponse"
    exit
}

# Upgrade to TLS over the Proxy
$sslStream = New-Object System.Net.Security.SslStream($stream, $false, {$true})
$sslStream.AuthenticateAsClient($realTarget)  # 🔥 This sets the SNI to `good.com`

# Now Send the Real HTTP Request
$writer = New-Object System.IO.StreamWriter($sslStream)
$writer.WriteLine("GET / HTTP/1.1")
$writer.WriteLine("Host: $realTarget")  # 🔥 Host header must match `good.com`
$writer.WriteLine("Connection: close")
$writer.WriteLine()
$writer.Flush()

# Read Response
$response = New-Object System.IO.StreamReader($sslStream)
$responseText = $response.ReadToEnd()

Write-Output $responseText
