package lib

type RowKey []interface{}
type TableId interface{}

type Parent interface {
	SetFieldCallback(tableId TableId, rowKey RowKey, columnIndex int, value []byte)
}

type DatastoreStructWithParent struct {
	DatastoreStruct
	parent Parent
	rowKey RowKey
}

func NewDatastoreStructWithParent(store DatastoreSlot, sizes []int, parent Parent, rowKey RowKey) *DatastoreStructWithParent {
	return &DatastoreStructWithParent{
		DatastoreStruct: *NewDatastoreStruct(store, sizes),
		parent:          parent,
		rowKey:          rowKey,
	}
}

func (s *DatastoreStructWithParent) SetField(index int, data []byte) {
	s.DatastoreStruct.SetField(index, data)
	if s.parent != nil {
		s.parent.SetFieldCallback(nil, s.rowKey, index, data)
	}
}
