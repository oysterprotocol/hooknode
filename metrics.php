<?php
//this function calls the Ubuntu CLI using shell_exec to get the number of prcessors
function get_processor_cores_number() {
   $command = "cat /proc/cpuinfo | grep processor | wc -l";

   return  (int) shell_exec($command);
}

//Function to get load.  It gets three load scores you see printed below
function get_cpu_load_average(){
    
	$sys_cmd = "cat /proc/loadavg";
	$sys_ret = shell_exec($sys_cmd);
	$load = explode(" ", $sys_ret);
	//We get the number of processor cores
	$number_of_cores = get_processor_cores_number();
    // We build the return array
	$return_array = array('Average load over one minute' => $load[0]/ $number_of_cores , 'Average load over five minutes' => $load[1]/ $number_of_cores , 'Average load over fifteen minutes' => $load[2]/ $number_of_cores );
	
	return ($return_array);	
}

$load = get_cpu_load_average();
print("Average load over one minute: " . $load['Average load over one minute']."\n");
print("Average load over five minutes: " . $load['Average load over five minutes']."\n");
print("Average load over fifteen minutes: " . $load['Average load over fifteen minutes']."\n");


$input = $_GET["param1"];

print($input);
//and then we can  build a return packet
$return_packet = $load; 
//and we can encode this into json.  We will generally encode responses in json.
$encoded = json_encode($return_packet);
echo encoded

?>
