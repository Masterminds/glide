// active package contains basic Active Record support.
//
// This package contains support for the design pattern called
// "Active Record". This pattern combines a structural data representation
// with the code necessary to persist the data.
//
// In the typical implementation, a record is stored in a relational
// database, and each record corresponds to a row in the database. The
// pattern here does not enforce use of an RDB. You may use whatever
// storage you see fit.
//
// This is not an active record *implementation*. It is only a set of
// interfaces describing what an active record should look like.
//
// Significantly, this does not specify how the record object obtains a
// handle to the underlying database. The assumption is that a constructor
// function will handle this detail.
package active

import (
	"github.com/Masterminds/cookoo"
)

// Record describes the data storage methods on an active record.
type Record interface {

	// Insert adds a new record.
	//
	// Implementations *should* set an attribute to the value of the
	// unique ID of the record.
	Insert() error

	// Update modifies an existing record.
	Update() error
	// Save creates a new record if non exists, and modifies an existing record.
	//
	// If a new record is created, implementations should set an attribute to the
	// unique ID of the record.
	Save() error


	// Load should load the data from persistent storage and set the local
	// attributes accordingly.
	//
	// For this to work correctly, it may require that the local unique ID attribute
	// be manually set.
	//
	// Attribute values on the struct may be overwritten when this is executed.
	//
	// Load MUST NOT have side-effects when a load fails to find a record. It MUST
	// NOT alter the existing record if it fails to find a new record.
	//
	// This allows the following desired behavior:
	//
	// 	r := NewRecord()
	//  r.Id = someId
	// 	r.Load()
	// 	r.Save() // Creates record with someId if none was found.
	Load() error

}

// Records describes the data access methods on a collection of records.
type Records interface {
	// Fetch all of the records of this type.
	All() ([]interface{}, error)
	// Fetch paged records for this type.
	Paged(offset, limit int) ([]interface{}, error)
	// Find all records that map the given filter.
	// Find(filter map[string]interface{}, offset, limit int) ([]interface{}, error)
}

// Load loads an already initialized record.
func Load(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	r := p.Get("record", nil).(Record)
	return r, r.Load()	
}

func Save(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	r := p.Get("record", nil).(Record)
	return r, r.Save()	
}
