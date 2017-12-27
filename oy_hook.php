<?php
$oy_hook_payout = "0x00F72426ccE219B5Fa1b7E4C680F4440669Ba273";//address to receive PRL payouts

$oy_hook_port = "14500";

$oy_hook_broker = array("broker1.oyster.ws");//can sign up for multiple broker nodes

$oy_hook_permitted = array();

foreach ($oy_hook_broker as $oy_hook_unique) $oy_hook_permitted[gethostbyname($oy_hook_unique)] = true;

if (!isset($oy_hook_permitted[$_SERVER['REMOTE_ADDR']])) die("PERMISSION DENIED");

if (!isset($_GET['oy_command'])||($_GET['oy_command']!="getNodeInfo"&&$_GET['oy_command']!="getNeighbors"&&$_GET['oy_command']!="getPayout"&&$_GET['oy_command']!="getLoad")) exit;

if ($_GET['oy_command']=="getPayout") {
    echo $oy_hook_payout;
    exit;
}

if ($_GET['oy_command']=="getLoad") {
    //TODO needs function to detect node's current workload. Ideally with /proc/cpuinfo and a shell invocation of 'uptime'.
    exit;
}

function oy_iri($oy_url, $oy_command) {
    $oy_headers = array(
        "Content-Type:application/json",
        "X-IOTA-API-Version: 1.4"
    );

    $ch = curl_init();

    curl_setopt($ch, CURLOPT_HEADER, 0);
    curl_setopt($ch, CURLOPT_HTTPHEADER, $oy_headers);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
    curl_setopt($ch, CURLOPT_URL, $oy_url);
    curl_setopt($ch, CURLOPT_POST, 1);
    curl_setopt($ch, CURLOPT_POSTFIELDS, "{\"command\":\"".$oy_command."\"}");

    $data = curl_exec($ch);
    curl_close($ch);

    return $data;
}

echo oy_iri("http://localhost:".$oy_hook_port, $_GET['oy_command']);

//provides ethereum address for PRL payment