<?php

$oy_hook_payout = "0x00F72426ccE219B5Fa1b7E4C680F4440669Ba273";

$oy_hook_broker = array("broker1.oyster.ws");//can sign up for multiple broker nodes

$oy_hook_permitted = array();

foreach ($oy_hook_broker as $oy_hook_unique) $oy_hook_permitted[gethostbyname($oy_hook_unique)] = true;

if (!isset($oy_hook_permitted[$_SERVER['REMOTE_ADDR']])) die("PERMISSION DENIED");

echo "PASS";


//validates connection with broker node

//provides access to local iri api

//provides ethereum address for PRL payment