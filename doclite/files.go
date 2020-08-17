package doclite

import (
	"encoding/binary"
	"encoding/json"
	"strings"
)

const (
	demarcationbByteString = " "
	demarcationByte        = byte(32)
)

func (c *Cache) write(n *Node) error {
	if c.db.file == nil {
		//for testing db.file shouldn't be nil
		return nil
	}

	data := n.document.data
	if len(n.document.data) > dataSize {
		data = n.document.data[:dataSize]
		c.overflowDoc(n)
	} else {
		sizeFill := dataSize - len(n.document.data)
		
		if sizeFill > 0 {
			data = append(data, []byte(strings.Repeat(demarcationbByteString, sizeFill))...)
		}
	}
	if err := write(c.db.file, n.document.offset, data); err != nil {
		return err
	}
	return nil
}

func (c *Cache) read(n *Node) ([]byte, error) {
	if c.db.file == nil {
		return []byte{}, nil
	}
	buf := make([]byte, dataSize)
	_, err := read(c.db.file, n.document.offset, buf)
	lb := dataSize
	for lb >= 1 {

		if buf[lb-1] != demarcationByte {
			break
		}
		lb--
	}
	if lb == dataSize {
		buf = append(buf, c.getOverflowData(n)...)
	} else {
		buf = buf[:lb]
	}
	return buf, err
}

type overflowNode struct {
	ID   int64
	Data []byte
}

func (c *Cache) overflowDoc(n *Node) error {
	ofn := &overflowNode{ID: n.document.id, Data: n.document.data[dataSize:]}
	buf := make([]byte, 8)
	data, err := json.Marshal(ofn)
	if err != nil {
		return err
	}

	binary.BigEndian.PutUint64(buf, uint64(len(data)))
	buf = append(buf, data...)

	err = c.writeOverflowfile(buf)
	if c.db.isTesting {
		c.db.overflows = append(c.db.overflows, *ofn)
		c.db.metadata.OverflowSize += int64(len(buf))
	}
	return err
}

func (c *Cache) getOverflowData(n *Node) []byte {
	for _, ofn := range c.db.overflows {
		if ofn.ID == n.document.id {
			return ofn.Data
		}
	}
	var (
		sizeBuf [8]byte
	)
	x := int64(0)
	for {
		// read the first 8 byte to decode the size of the overflow data
		w, err := c.readOverflowfile(x, sizeBuf[:])
		if err != nil || w == 0 {
			break
		}

		x += int64(w)
		size := int64(binary.BigEndian.Uint64(sizeBuf[:])) // convert the byte to int
		buf := make([]byte, size)
		w, err = c.readOverflowfile(x, buf)
		if err != nil || w == 0 {
			break
		}
		x += size
		ofn := &overflowNode{}
		json.Unmarshal(buf, ofn)
		c.db.overflows = append(c.db.overflows, *ofn)
		if n.document.id == ofn.ID {
			return ofn.Data
		}

		c.db.overflows = append(c.db.overflows, *ofn)

	}

	return []byte{}
}

func (c *Cache) writeOverflowfile(data []byte) error {
	return write(c.db.overflowfile, c.db.metadata.OverflowSize, data)
}

func (c *Cache) readOverflowfile(offset int64, buf []byte) (int, error) {
	return read(c.db.overflowfile, offset, buf)
}