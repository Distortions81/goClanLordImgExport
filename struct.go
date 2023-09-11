package main

const TYPE_IDREF = 0x50446635    //PDf5
const TYPE_IMAGE = 0x42697432    //BIT2
const TYPE_COLOR = 0x436c7273    //CLRS
const TYPE_NAME = 0x43496d34     //Name
const TYPE_MYSTERY1 = 0x4c697431 //No idea
const TYPE_MYSTERY2 = 0x56657273 //No idea
const TYPE_MYSTERY3 = 0x4c617933 //No idea

type dataLocation struct {
	offset    uint32 //Location in the file
	size      uint32 //Data size
	entryType uint32 //Data type
	id        uint32 //Data ID

	colorBytes []byte
	name       string
	imageID    uint32
	colorID    uint32
}
