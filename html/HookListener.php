<?php

require_once("HookNode.php");

if (HookNode::verifyRegisteredBroker($_SERVER['REMOTE_ADDR'])) {
    $req = new stdClass();

    foreach ($_POST as $key => $value ) {
        $req->$key = $value;
    }

    $req->responseAddress = $_SERVER['REMOTE_ADDR'] . ':' .  $_SERVER['REMOTE_PORT'];

    processRequest($req);
} else {
    die("PERMISSION DENIED");
}

function processRequest($request)
{
    var_dump($request);

    if (isset($request->command)) {

        switch ($request->command) {
            case 'attachToTangle':
                HookNode::attachTx($request);
                break;
            case 'getNodeInfo':
                echo "getNodeInfo";
                break;
            case 'broadcastTransactions':
                echo 'broadcastTransactions';
                break;
            case 'getNeighbors':
                echo "getNeighbors";
                break;
            default:
                echo "i equals 2";
                break;
        }
    } else {
        die("NO COMMAND PROVIDED");
    }
}


