<?php 


class HookNodeModel {
	
	private $ipAddress;
	private $timestamp;
	private $score;
	
	public setIpAddress($__ipAddress){
		$this->ipAddress = $__ipAddress;
	}
	
	public setTimestamp($$__timestamp){
		$this->timestamp = $__timestamp;
	}
	
	public setScore($__score){
		$this->score = $__score;
	}
	
	public getIpAddress() {
		return $this->ipAddress;
	}
	
	public getTimestamp() {
		return $this->timestamp;
	}
	
	public getScore() {
		return $this->score;
	}
}