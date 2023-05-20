package backend

import (
	"log"
	"strconv"

	"github.com/gocql/gocql"
)

type CassandraBackend struct {
	session *gocql.Session
	table   string
	prefix  string
	size    int64

	verbose bool
}

func NewCassandraBackend(
	session *gocql.Session,
	table string,
	prefix string,
	size int64,
	verbose bool,
) *CassandraBackend {
	return &CassandraBackend{
		session: session,
		table:   table,
		prefix:  prefix,
		size:    size,

		verbose: verbose,
	}
}

func (b *CassandraBackend) ReadAt(p []byte, off int64) (n int, err error) {
	if b.verbose {
		log.Printf("ReadAt(len(p) = %v, off = %v)", len(p), off)
	}

	var val []byte
	if err := b.session.Query(`select data from `+b.table+` where key = ? limit 1`, b.prefix+"-"+strconv.FormatInt(off, 10)).Scan(&val); err != nil {
		if err == gocql.ErrNotFound {
			return len(p), nil
		}

		return 0, err
	}

	return copy(p, val), nil
}

func (b *CassandraBackend) WriteAt(p []byte, off int64) (n int, err error) {
	if b.verbose {
		log.Printf("WriteAt(len(p) = %v, off = %v)", len(p), off)
	}

	if err := b.session.Query(`insert into `+b.table+` (key, data) values (?, ?)`, b.prefix+"-"+strconv.FormatInt(off, 10), p).Exec(); err != nil {
		return 0, err
	}

	return len(p), nil
}

func (b *CassandraBackend) Size() (int64, error) {
	if b.verbose {
		log.Println("Size()")
	}

	return b.size, nil
}

func (b *CassandraBackend) Sync() error {
	if b.verbose {
		log.Println("Sync()")
	}

	return nil
}
