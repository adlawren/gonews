package db

import "github.com/jinzhu/gorm"

type DB interface {
	FirstOrCreate(interface{}, ...interface{}) DB
	Find(interface{}, ...interface{}) DB
	Order(interface{}, ...bool) DB
	Close() error
	Model(interface{}) DB
	Update(...interface{}) DB
	Where(interface{}, ...interface{}) DB
	Related(interface{}, ...string) DB
}

type gormDB struct {
	db *gorm.DB
}

func FromGormDB(db *gorm.DB) DB {
	return &gormDB{db: db}
}

func (gdb *gormDB) FirstOrCreate(out interface{}, where ...interface{}) DB {
	return &gormDB{db: gdb.db.FirstOrCreate(out, where...)}
}

func (gdb *gormDB) Find(out interface{}, where ...interface{}) DB {
	return &gormDB{db: gdb.db.Find(out, where...)}
}

func (gdb *gormDB) Order(value interface{}, reorder ...bool) DB {
	return &gormDB{db: gdb.db.Order(value, reorder...)}
}

func (gdb *gormDB) Close() error {
	return gdb.db.Close()
}

func (gdb *gormDB) Model(value interface{}) DB {
	return &gormDB{db: gdb.db.Model(value)}
}

func (gdb *gormDB) Update(attrs ...interface{}) DB {
	return &gormDB{db: gdb.db.Update(attrs...)}
}

func (gdb *gormDB) Where(query interface{}, args ...interface{}) DB {
	return &gormDB{db: gdb.db.Where(query, args...)}
}

func (gdb *gormDB) Related(value interface{}, foreignKeys ...string) DB {
	return &gormDB{db: gdb.db.Related(value, foreignKeys...)}
}
