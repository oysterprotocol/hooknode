<?php

//PHP built in function to get load.  It gets three load scores you see printed beloW
$load = sys_getloadavg();

//this function calls the Ubuntu CLI using shell_exec to get the number of prcessors
function get_processor_cores_number() {
    $command = "cat /proc/cpuinfo | grep processor | wc -l";

    return  (int) shell_exec($command);
}
$get_request = $_GET["param1"];

print($get_request);

// Calculate AND  Build a return packet

$return_packet = array('Average load over one minute' => $load[0]/ get_processor_cores_number(), 'Average load over five minutes' => $load[1]/ get_processor_cores_number(), 'Average load over fifteen minutes' => $load[2]/ get_processor_cores_number());

//and we can encode this into json.  We will generally encode responses in json.
$encoded = json_encode($return_packet);
echo $encoded;