// SQL datasource and commands for Cookoo.
// This provides basic SQL support for Cookoo.
package sql

import (
	dbsql "database/sql"
	"sync"
)

// NewDbDatasource creates a new SQL datasource.
//
// Currently, this returns an actual *"database/sql".Db instance. Note that
// you do not *need* to use this function in order to create a new database
// datasource. You can simply place a database handle into the context
// as a datasource.
//
// Example:
//	ds, err := sql.NewDatasource("mysql", "root@/mpbtest")
//	if err != nil {
//		panic("Could not create a database connection.")
//		return
//	}
//
//	cxt.AddDatasource("db", ds)
//
// In the example above, we create a new datasource and then add it to
// the context. This should be done at server init, before web.Serve
// or router.HandleRequest().
func NewDbDatasource(driverName, datasourceName string) (*dbsql.DB, error) {
	return dbsql.Open(driverName, datasourceName)
}

// NewStmtCache creates a new cache for prepared statements.
//
// Initial capacity determines how big the cache will be.
//
// Warning: The implementation of the caching layer will likely
// change from relatively static to an LRU. To avoid memory leaks, the
// statement cache will automatically clear itself each time it hits
// 1000 distinct statements.
func NewStmtCache(dbHandle *dbsql.DB, initialCapacity int) StmtCache {
	c := new(StmtCacheMap)
	c.cache = make(map[string]*dbsql.Stmt, initialCapacity)
	c.capacity = initialCapacity
	c.dbh = dbHandle

	return c
}

// A StmtCache caches SQL prepared statements.
//
// It's intended use is as a datsource for a long-running SQL-backed
// application. Prepared statements can exist across requests and be
// shared by separate goroutines. For frequently executed statements,
// this is both more performant and more secure (at least for some
// drivers).
//
// IMPORTANT: Statments are cached by string key, so it is important that to
// get the most out of the cache, you re-use the same strings. Otherwise,
// 'SELECT surname, name FROM names' will generate a different cache entry
// than 'SELECT name, surname FROM names'.
//
// The cache is driver-agnostic.
type StmtCache interface {
	Prepare(statement string) (*dbsql.Stmt, error)
	Clear() error
}

type StmtCacheMap struct {
	cache    map[string]*dbsql.Stmt
	capacity int
	dbh      *dbsql.DB
	mu sync.Mutex
}

// Deprecated. Use Prepare()
/*
func (c *StmtCacheMap) Get(statement string) (*dbsql.Stmt, error) {
	return c.Prepare(statement)
}
*/

// Prepare gets a prepared statement from a SQL string.
//
// This will return a cached statement if one exists, otherwise
// this will generate one, insert it into the cache, and return
// the new statement.
//
// It is assumed that the underlying database layer can handle
// parallelism with prepared statements, and we make no effort
// to deal with locking or synchronization.
// For compatibility with database/sql.DB.Prepare
func (c *StmtCacheMap) Prepare(statement string) (*dbsql.Stmt, error) {
	// Protect the statment cache.
	c.mu.Lock()
	defer c.mu.Unlock()

	if stmt, ok := c.cache[statement]; ok {
		return stmt, nil
	}
	// Else we prepare the statement and then cache it.
	stmt, err := c.dbh.Prepare(statement)
	if err != nil {
		return nil, err
	}

	// Hack: Until we have a statement cache, we stop memory leaks by clearing
	// the entire cache when it hits 1000.
	if len(c.cache) > 1000 {
		c.cache = make(map[string]*dbsql.Stmt, c.capacity)
	}

	// Cache by string key.
	c.cache[statement] = stmt
	return stmt, nil
}

// Clear clears the cache.
//
// Right now, it is suggested that the 
func (c *StmtCacheMap) Clear() error {
	// While I don't think this is a good idea, it might be necessary. On the
	// flip side, it might cause race conditions if one goroutine is running
	// a query while another is clearing the cache. For now, leaving this
	// to the memory manager.
	//for _, stmt := range c.cache {
	//	stmt.Close()
	//}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*dbsql.Stmt, c.capacity)
	return nil
}
