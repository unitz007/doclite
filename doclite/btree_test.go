package doclite

import (
	"encoding/json"
	"fmt"
	"testing"
)

func testBtree(bt *Btree, t *testing.T) *Btree {
	var (
		numOfInsert = 100
	)
	for i := 0; i < numOfInsert; i++ {
		bt.Insert([]byte(fmt.Sprintf("%d docklite", i)))
	}

	if bt.NumDocuments != int64(numOfInsert) {
		t.Errorf("%d is not  equal to number of docs inserted", bt.NumDocuments)
	}
	add := 0
	if numOfInsert%MinKeys != 0 {
		add++
	}
	if bt.NumRoots != numOfInsert/MinKeys+add {
		t.Errorf("%d is not  equal to number of root docs", bt.NumRoots)
	}

	x := 0
	pageNo := bt.Pages[x]
	for i := 1; i <= numOfInsert; i++ {
		if i%MinKeys == 0 {
			x++
			pageNo = bt.Pages[x]
		}
		n, err := bt.Find(int64(i))
		if err != nil {
			t.Errorf("%v getting data failed", err)
		}
		if i%MinKeys != 0 && n.document.offset != pageNo*pageSize+int64(((i-1)%MinKeys)*dataSize) {
			t.Errorf("unmatching offset")
		}
	}

	bt.InsertSubCollection("newcollection")

	return bt.Get("newcollection")
}
func TestBtree(t *testing.T) {
	db := &DB{metadata: &Meta{}}
	bt := db.newBtree("")
	bt1 := testBtree(bt, t)

	if bt.Get("newcollection") == nil {
		t.Errorf("collection newcollection not found")
	}
	testBtree(bt1, t)

	data, err := json.Marshal(bt)
	if err != nil {
		t.Errorf("%v failed marshaling btree", err)
	}
	fmt.Println(string(data))
}

// TestBtreeDiskInitExactMultiple verifies that diskInitBtree sets the correct
// numChildren for the last root when NumDocuments is an exact multiple of MinKeys.
// This is a regression test for the bug where NumDocuments % MinKeys == 0 caused
// numChildren to be 0, making all documents in the last root invisible on reload.
func TestBtreeDiskInitExactMultiple(t *testing.T) {
	// Use exactly 2 * MinKeys documents so that NumDocuments % MinKeys == 0
	numDocs := 2 * MinKeys

	db := &DB{metadata: &Meta{}}
	bt := db.newBtree("")

	// Insert exactly 2 * MinKeys documents
	for i := 0; i < numDocs; i++ {
		bt.Insert([]byte(fmt.Sprintf("doc-%d", i)))
	}

	// Verify all documents were inserted
	if bt.NumDocuments != int64(numDocs) {
		t.Fatalf("expected NumDocuments=%d, got %d", numDocs, bt.NumDocuments)
	}

	if bt.NumRoots != 2 {
		t.Fatalf("expected NumRoots=2, got %d", bt.NumRoots)
	}

	// Verify all documents are findable via Find()
	for i := int64(1); i <= int64(numDocs); i++ {
		n, err := bt.Find(i)
		if err != nil {
			t.Errorf("Find(%d) failed: %v", i, err)
			continue
		}
		expected := fmt.Sprintf("doc-%d", i-1)
		actual := string(n.document.data)
		if actual != expected {
			t.Errorf("Find(%d): expected %q, got %q", i, expected, actual)
		}
	}

	// Verify the last root's numChildren is MinKeys, not 0
	lastRoot := bt.roots[len(bt.roots)-1]
	if lastRoot.numChildren != MinKeys {
		t.Errorf("last root numChildren = %d, want %d (MinKeys)", lastRoot.numChildren, MinKeys)
	}

	// Simulate a disk reload by clearing in-memory roots and re-initializing
	originalRoots := bt.roots
	bt.roots = nil
	bt.initBtreeRoot = false
	bt.db = &DB{metadata: &Meta{}, file: nil} // nil file so read returns empty (ok for test)
	bt.findPool = make(map[int64]int64)
	bt.diskInitBtree()

	// After diskInitBtree, verify all documents are still findable
	for i := int64(1); i <= int64(numDocs); i++ {
		_, err := bt.Find(i)
		if err != nil {
			t.Errorf("after diskInitBtree, Find(%d) failed: %v", i, err)
		}
	}

	// Verify last root's numChildren is correct after disk re-init
	lastRootAfterReload := bt.roots[len(bt.roots)-1]
	if lastRootAfterReload.numChildren != MinKeys {
		t.Errorf("after diskInitBtree, last root numChildren = %d, want %d (MinKeys)",
			lastRootAfterReload.numChildren, MinKeys)
	}

	// Verify first root's numChildren is still MinKeys (unchanged by the fix)
	firstRootAfterReload := bt.roots[0]
	if firstRootAfterReload.numChildren != MinKeys {
		t.Errorf("after diskInitBtree, first root numChildren = %d, want %d (MinKeys)",
			firstRootAfterReload.numChildren, MinKeys)
	}

	// Restore for cleanup (avoid mutating shared test state)
	bt.roots = originalRoots
}

func TestBinarySearch(t *testing.T) {
	nodes := []*Node{}
	ids := []int64{}
	for i := 0; i < 100; i++ {
		id := int64(i)
		ids = append(ids, id)
		doc := &Document{id: id, data: []byte{}}
		n := &Node{document: doc}
		nodes = append(nodes, n)
	}
	for i := 0; i < 100; i++ {
		if ids[i] != nodes[indexOfNodes(ids[i], nodes, 100)].document.id {
			t.Errorf(" nodes not sorted %d %d", i, indexOfNodes(ids[i], nodes, 100))
		}
	}

	list := []*Node{{document: &Document{id: 1, data: []byte{}}}, {document: &Document{id: 33, data: []byte{}}}, {document: &Document{id: 65, data: []byte{}}}, {document: &Document{id: 97, data: []byte{}}}}

	if indexOfNodes(int64(3), list, 4) != 0 {
		t.Errorf(" wrong node")
	}

	if indexOfNodes(int64(40), list, 4) != 1 {
		t.Errorf(" wrong node")
	}

	if indexOfNodes(int64(78), list, 4) != 2 {
		t.Errorf(" wrong node")
	}

	if indexOfNodes(int64(98), list, 4) != 3 {
		t.Errorf(" wrong node")
	}
}
