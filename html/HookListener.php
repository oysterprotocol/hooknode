<?php

require_once("HookNode.php");

/*
 * TODOS: This calls a stubbed-out method that Arthur is working on.  His
 * work should either be added into that stubbed-out method, or remove that method
 * and replace with his method and replace the call.
 */

if (HookNode::verifyRegisteredBroker($_SERVER['REMOTE_ADDR'])) {
    $req = new stdClass();

    $request_body = file_get_contents('php://input');
    $req = json_decode($request_body);
    $req->responseAddress = $_SERVER['REMOTE_ADDR'] . ':' .  $_SERVER['REMOTE_PORT'];

    $my_file = '/home/OUTPUT.txt';
    $handle = fopen($my_file, 'a') or die('Cannot open file:  '.$my_file);
    $data = "Request TEST! \n";
    $data .= var_export($req, true);
//
//    foreach ($_POST as $key => $value ) {
//        $req->$key = $value;
//        $data .= "key: \n";
//        $data .= $key . "\n";
//    }

//    $data .= "\nTrytes is \n:";
//    $data .= implode("\n", $_POST['trytes']);
//    $data .= "\nType of trytes is: \n";
//    $data .= gettype($_POST['trytes']);

    fwrite($handle, $data);


    processRequest($req);
} else {
    die("PERMISSION DENIED");
}

function processRequest($request)
{
    if (isset($request->command)) {

        switch ($request->command) {
            case 'attachToTangle':
                HookNode::attachTx($request);
                sleep(45);
                //HookNode::attachTx($request);  // for good measure
                break;
            default:
                die("UNRECOGNIZED COMMAND");
                break;
        }
    } else {
        die("NO COMMAND PROVIDED");
    }
}


