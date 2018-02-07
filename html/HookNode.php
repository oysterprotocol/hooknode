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
                if ($trytesToBroadcast != NULL) {

                    // DELETE THIS
                    $my_file = '/home/OUTPUT.txt';
                    $handle = fopen($my_file, 'a') or die('Cannot open file:  '.$my_file);
                    $data = "\n\n=========== attachTx! =============\n";
                    $data .= var_export($trytes, true);
                    $data .= "\n=========== attachTx! =============\n\n";
                    fwrite($handle, $data);

                    self::broadcastTransactions($trytesToBroadcast);
                    sleep(5);
                    self::broadcastTransactions($trytesToBroadcast);
                    sleep(5);
                    self::broadcastTransactions($trytesToBroadcast);
                    sleep(5);
                    self::broadcastTransactions($trytesToBroadcast);
                    sleep(5);
                    self::broadcastTransactions($trytesToBroadcast);
                }
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

        $my_file = '/home/OUTPUT.txt';
        $handle = fopen($my_file, 'a') or die('Cannot open file:  '.$my_file);
        $data = "\n\n\n=========== attachToTangle! BEFORE =============\n";
        $data .= var_export($command, true);
        $data .= "\n=========== attachToTangle! BEFORE =============\n\n\n";

        $resultOfAttach = $req->makeRequest($command);

        // DELETE THIS
        $my_file = '/home/OUTPUT.txt';
        $handle = fopen($my_file, 'a') or die('Cannot open file:  '.$my_file);
        $data = "\n\n\n=========== attachToTangle! cmd =============\n";
        $data .= var_export($resultOfAttach, true);
        $data .= "\n=========== attachToTangle! cmd =============\n\n\n";
        fwrite($handle, $data);

        if (!is_null($resultOfAttach) && property_exists($resultOfAttach, 'trytes')) {
            $my_file = '/home/OUTPUT.txt';
            $handle = fopen($my_file, 'a') or die('Cannot open file:  '.$my_file);
            $data = "Attached to tangle with this result! \n";
            $data .= implode("\n\n", $resultOfAttach->trytes);
            fwrite($handle, $data);
            return $resultOfAttach->trytes;
        } else {
            $my_file = '/home/OUTPUT.txt';
            $handle = fopen($my_file, 'a') or die('Cannot open file:  '. $my_file);
            $data = "Attached to tangle failed! \n";
            $data .= $resultOfAttach;
            fwrite($handle, $data);
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
            $my_file = '/home/OUTPUT.txt';
            $handle = fopen($my_file, 'a') or die('Cannot open file:  '.$my_file);
            $data = "Broadcasted! \n";
            fwrite($handle, $data);
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
            $my_file = '/home/OUTPUT.txt';
            $handle = fopen($my_file, 'a') or die('Cannot open file:  '.$my_file);
            $data = "Stored! \n";
            fwrite($handle, $data);
            return $result;
        } else {
            throw new Exception('storeTransactions failed!');
        }
    }
}
