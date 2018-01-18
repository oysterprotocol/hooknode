<?php
require_once("database/HookNodeDAO.php");

class HookNode {

	private $dao;

	function __construct() {
		$this->dao = new HookNodeDAO();
	}
	
	private function selectHookNode() {
        $node = $this->dao->getHigherScoreNode();
        return $node;
    }

	/**
	* Function to increase node's score
	* @author Arthur Mastropietro - 16/01
	*/
	
	private function increaseHookNodeScore($nodeId) {
		$this->dao->increaseNodeScore($nodeId);
		return true;
	}
	
	/**
	* Function to decrease node's score
	* @author Arthur Mastropietro - 16/01
	*/
	
	private function decreaseHookNodeScore($nodeId) {
		$this->dao->decreaseNodeScore($nodeId);
		return true;
	}
	
}

