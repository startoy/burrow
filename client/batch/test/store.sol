pragma solidity ^0.4.4;

contract Store {

	struct Item {
		uint id;
		string name;
		bytes8 junk;
		address blob;
	}

	mapping (uint=>Item) items;
	uint counter;

	function saveItem(uint id, string name, bytes8 junk, address blob) {
		var newItem = Item(id, name, junk, blob);
		items[id] = newItem;
		counter++;
	}

	function getCount() returns (uint count) {
		return counter;
	}
}
