package bmap

import "encoding/xml"

// Bmap represents the structure of a .bmap XML file
type Bmap struct {
	XMLName           xml.Name `xml:"bmap"`
	Version           string   `xml:"version,attr"`
	ImageSize         int64    `xml:"ImageSize"`
	BlockSize         int      `xml:"BlockSize"`
	BlocksCount       int64    `xml:"BlocksCount"`
	MappedBlocksCount int64    `xml:"MappedBlocksCount"`
	ChecksumType      string   `xml:"ChecksumType"`
	BmapFileChecksum  string   `xml:"BmapFileChecksum"`
	BmapFileSHA1      string   `xml:"BmapFileSHA1"`
	BlockMap          []Range  `xml:"BlockMap>Range"`
}

// Range represents a single range entry in the BlockMap
type Range struct {
	Checksum string `xml:"chksum,attr"`
	Text     string `xml:",chardata"`
}
