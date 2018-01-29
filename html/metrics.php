<?php

class Metrics
{
    //this function calls the Ubuntu CLI using shell_exec to get the number of prcessors
    public static function get_processor_cores_number() {
        $command = "cat /proc/cpuinfo | grep processor | wc -l";

        return  (int) shell_exec($command);
    }

    public static function getLoadReport() {
        //PHP built in function to get load.  It gets three load scores you see printed beloW
        $load = sys_getloadavg();

        // Calculate AND  Build a return packet
        $cores = Metrics::get_processor_cores_number();

        //Error handling zero core
        if ($cores < 1) {
            throw new Exception('Number of cores is zero!');
        }

        $return_packet = array('Average load over one minute' => $load[0]/ $cores, 
                               'Average load over five minutes' => $load[1]/  $cores,
                               'Average load over fifteen minutes' => $load[2]/  $cores);

        //and we can encode this into json.  We will generally encode responses in json.
        $encoded = json_encode($return_packet);
        echo $encoded;
    }
}



