<?php 

require_once ("Database.class.php");
require_once ("HookNodeModel.php");

class HookNodeDAO extends Database {
	
	function __construct() {
		parent::__construct();
	}
	
	public function getHigherScoreNode() {
		$node = $this->query("SELECT id, ip_address, timestamp, score from node ORDER BY score DESC, timestamp DESC LIMIT 1");
		$model = new HookNodeModel();
		$singleNode = $node[0];
		$model->setIpAddress($singleNode["ip_address"]);
		$model->setTimestamp($singleNode["timestamp"]);
		$model->setScore($singleNode["score"]);
		return $model;
	}
	
	public function getAllNodes() {
		$nodes = $this->query("SELECT id, ip_address, timestamp, score from node");
		return $nodes;
	}
	
	public function increaseNodeScore($id) {
		$update = $this->query("UPDATE node SET score = score + 1 WHERE id = :id", array("id" => $id));
		return $update;
	}
	
	public function decreaseNodeScore($id) {
		$update = $this->query("UPDATE node SET score = score - 1 WHERE id = :id", array("id" => $id));
		return $update;
	}
	
	public function insertNode($data) {
		$this->bindMore(array("ip_address" => $data["ipAddress"], "timestamp" => $data["timestamp"], "score" => $data["score"]));
		$insert = $this->query("INSERT INTO node(ip_address, timestamp, score) VALUES(:ip_address, :timestamp, :score)");
		return $insert;
	}
	
}