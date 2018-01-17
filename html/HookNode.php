<?php

require_once("requests/IriWrapper.php");
require_once("requests/IriData.php");

class HookNode
{

    public function __construct($nodeUrl, $apiVersion = '1.4', $nodeId = 'OYSTERPEARL')
    {
    }

    public static function verifyRegisteredBroker($brokerIp) {
        return true;
        /*TODO

        hooknode needs to check if it knows the broker
        for now, just pretend we do

        We should report the result of this check to the broker
        so it can either wait or move on to the next hooknode node

        */
    }

    public static function attachTx($transactionObject)
    {
        $trytesToBroadcast = NULL;

        if (property_exists($transactionObject, 'trunkTransaction')) {
            try {
                $trytesToBroadcast = self::attachToTangle($transactionObject);
            } catch (Exception $e) {
                echo "Caught exception: " . $e->getMessage() . $GLOBALS['nl'];
                return;
            }
        }

        if ($trytesToBroadcast != NULL) {
            try {
                self::broadcastTransactions($trytesToBroadcast);
            } catch (Exception $e) {
                echo "Caught exception: " . $e->getMessage() . $GLOBALS['nl'];
                return;
            }
        }
    }

    private static function attachToTangle($transactionObject)
    {
        $req = new IriWrapper();

        $command = new stdClass();
        $command->command = "attachToTangle";

        $command->minWeightMagnitude = IriData::$minWeightMagnitude;

        $command->trunkTransaction = $transactionObject->trunkTransaction;
        $command->branchTransaction = $transactionObject->branchTransaction;
        $command->trytes = $transactionObject->trytes;

        $resultOfAttach = $req->makeRequest($command);

        if (!is_null($resultOfAttach) && property_exists($resultOfAttach, 'trytes')) {
            return $resultOfAttach->trytes;
        } else {
            throw new Exception('attachToTangle failed!');
        }
    }

    private static function broadcastTransactions($trytesToBroadcast)
    {
        $req = new IriWrapper();

        $command = new stdClass();
        $command->command = "broadcastTransactions";
        $command->trytes = $trytesToBroadcast;

        $req->makeRequest($command);
    }

    private static function sendResponse() {

    }
}