<?php
// Generic HTTP proxy for shared hosting (PHP 5.6+).
// Upload as proxy_post.php — all target URLs and headers come from the client (POST JSON).
//
// Request body (application/json):
//   { "url": "https://example.com/path?query=1", "method": "GET", "headers": { "Accept": "*/*" }, "body": "" }

header('Content-Type: application/json; charset=utf-8');

$raw = file_get_contents('php://input');
$data = json_decode($raw, true);
if (!is_array($data) || empty($data['url'])) {
    http_response_code(400);
    echo json_encode(array('error' => 'missing url'));
    exit;
}

$url = $data['url'];
$method = isset($data['method']) ? strtoupper($data['method']) : 'GET';
$body = isset($data['body']) ? $data['body'] : '';
$headers = array();
if (isset($data['headers']) && is_array($data['headers'])) {
    foreach ($data['headers'] as $name => $value) {
        $headers[] = $name . ': ' . $value;
    }
}

$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, $url);
curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $method);
curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_ENCODING, '');
curl_setopt($ch, CURLOPT_TIMEOUT, 90);
curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);
if ($method !== 'GET' && $method !== 'HEAD' && $body !== '') {
    curl_setopt($ch, CURLOPT_POSTFIELDS, $body);
}

$response = curl_exec($ch);
$code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
$ctype = curl_getinfo($ch, CURLINFO_CONTENT_TYPE);
$err = curl_error($ch);
curl_close($ch);

if ($response === false) {
    http_response_code(502);
    echo json_encode(array('error' => 'curl failed', 'detail' => $err));
    exit;
}

http_response_code($code);
if ($ctype) {
    header('Content-Type: ' . $ctype);
} else {
    header('Content-Type: application/octet-stream');
}
echo $response;
