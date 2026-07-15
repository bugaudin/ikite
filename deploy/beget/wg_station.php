<?php
// Windguru station proxy for shared hosting (PHP 5.6+).
// Deploy as wg_station.php?id_station=2763
$id = 0;
if (isset($_GET['id_station'])) {
    $id = intval($_GET['id_station']);
} elseif (isset($_GET['sid'])) {
    $id = intval($_GET['sid']);
}

if ($id <= 0) {
    header('HTTP/1.1 400 Bad Request');
    header('Content-Type: text/plain; charset=utf-8');
    echo 'missing id_station';
    exit;
}

$url = 'https://www.windguru.net/int/iapi.php?q=station&id_station=' . $id . '&weather=false';
$headers = array(
    'authority: www.windguru.net',
    'accept: */*',
    'accept-language: en-US,en;q=0.9',
    'origin: https://www.windguru.cz',
    'referer: https://www.windguru.cz/',
    'user-agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36',
);

$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, $url);
curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_ENCODING, '');
curl_setopt($ch, CURLOPT_TIMEOUT, 45);
curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);
curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'GET');
$body = curl_exec($ch);
$code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
$err = curl_error($ch);
curl_close($ch);

if ($body === false || $code !== 200) {
    header('HTTP/1.1 502 Bad Gateway');
    header('Content-Type: text/plain; charset=utf-8');
    echo 'windguru fetch failed: HTTP ' . $code . ' ' . $err;
    exit;
}

header('Content-Type: application/json; charset=utf-8');
echo $body;
