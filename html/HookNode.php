<?php

require_once("requests/IriWrapper.php");
require_once("requests/IriData.php");

class HookNode
{

    public static function verifyRegisteredBroker($brokerIp)
    {
        return true;
        /*TODO

        Put Arthur's stuff in here or get rid of this method and replace with
        the new method.

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

        $result = $req->makeRequest($command);

        if (!is_null($result) && property_exists($result, 'duration') &&
            !property_exists($result, 'error')) {
            self::storeTransactions($trytesToBroadcast);
        } else {
            throw new Exception('broadcastTransaction failed!');
        }
    }

    private static function storeTransactions($trytes)
    {
        $req = new IriWrapper();

        $command = new stdClass();
        $command->command = "storeTransactions";
        $command->trytes = $trytes;

        $result = $req->makeRequest($command);

        if (!is_null($result) && property_exists($result, 'duration') &&
            !property_exists($result, 'error')) {
            return $result;
        } else {
            throw new Exception('storeTransactions failed!');
        }
    }

    private static function sendResponse()
    {

    }
}