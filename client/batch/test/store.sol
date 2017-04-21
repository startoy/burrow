pragma solidity ^0.4.4;

contract Store {

	struct Item {
		uint id;
		string name;
	}

	mapping (uint=>Item) items;
	uint counter;

	function saveItem(uint id, string name) {
		var newItem = Item(id, name);
		items[id] = newItem;
		counter++;
	}

	function getCount() returns (uint count) {
		return counter;
	}
}
