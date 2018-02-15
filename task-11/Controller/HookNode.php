<?php
namespace Controller;

require_once("Dao/BrokerNodeDAO.php");

use Dao\BrokerNodeDAO;

class HookNode {

	private $brokerNodeDao;

	function __construct() {
		$this->brokerNodeDao = new BrokerNodeDAO();
	}
	
	
	/**
	* Function to verify with a given IpAddress exists
	* @author Arthur Mastropietro
	*/
	
	public function verifyRegisteredBrokerNode($ipAddress) {
		return $this->brokerNodeDao->verifyRegisteredBrokerNode($ipAddress);
	}
	
}

