<?php
$oy_hook_payout = "0x00F72426ccE219B5Fa1b7E4C680F4440669Ba273";//address to receive PRL payouts

$oy_hook_port = "14500";

$oy_hook_broker = array("broker1.oyster.ws");//can sign up for multiple broker nodes

$oy_hook_permitted = array();

foreach ($oy_hook_broker as $oy_hook_unique) $oy_hook_permitted[gethostbyname($oy_hook_unique)] = true;

if (!isset($oy_hook_permitted[$_SERVER['REMOTE_ADDR']])) die("PERMISSION DENIED");

// Create a stream
$opts = [
    "http" => [
        "method" => "POST",
        "header" => "Content-Type:application/json\r\nX-IOTA-API-Version: 1.4\r\n",
        "content" => array("command" => "getNodeInfo")
    ]
];

function curl($url) {
    global $oy_hook_port;

    $ch = curl_init();

    curl_setopt($ch, CURLOPT_HEADER, 0);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
    curl_setopt($ch, CURLOPT_URL, "http://127.0.0.1:".$oy_hook_port);

    $data = curl_exec($ch);
    curl_close($ch);

    return $data;
}

//$file = file_get_contents('http://localhost:'.$oy_hook_port, false, stream_context_create($opts));

var_dump(curl("http://127.0.0.1:14500"));

var_dump($file);

echo "PASS";


//validates connection with broker node

//provides access to local iri api

//provides ethereum address for PRL payment