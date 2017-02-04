package repl

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/gocql/gocqlsh/metadata"

	"github.com/chzyer/readline"
	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
)

type CQL struct {
	db   *gocql.Session
	r    *readline.Instance
	meta *metadata.Cassandra
}

func New(db *gocql.Session, r *readline.Instance) *CQL {
	// TODO: protbably want to pass this in for testing
	r.Config.AutoComplete = &cqlCompleter{db}
	meta := metadata.New(db)
	return &CQL{
		db:   db,
		r:    r,
		meta: meta,
	}
}

func (c *CQL) err(err error) {
	// TODO: improve error display
	if _, err := fmt.Fprintf(c.r, "error: %v\n", aurora.Red(err)); err != nil {
		panic(err)
	}
}

func (c *CQL) Run() error {
	clusterInfo, err := c.meta.ClusterMeta()
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(c.r, "Connected to %s at %v\n", aurora.Magenta(clusterInfo.Name), clusterInfo.Address); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(c.r, "[gocqlsh | Cassandra %s | CQL Spec %s | Native Protocol %s]\n", clusterInfo.Version,
		clusterInfo.CQLVersion, clusterInfo.Protocol); err != nil {
		return err
	}

	for {
		line, err := c.r.Readline()
		if err != nil {
			return err
		}

		if err := c.exec(line); err != nil {
			c.err(err)
		}

	}
}

func (c *CQL) executeQuery(query string) error {
	iter := c.db.Query(query).Iter()

	table := tablewriter.NewWriter(c.r)
	table.SetAutoFormatHeaders(false)
	var header, columns []string
	for _, col := range iter.Columns() {
		header = append(header, aurora.Red(col.Name).String())
		columns = append(columns, col.Name)
	}

	table.SetHeader(header)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)

	rows, err := iter.SliceMap()
	if err != nil {
		// TODO: write errors to repl
		return err
	}

	line := make([]string, len(columns))
	for _, row := range rows {
		for i, col := range columns {
			line[i] = fmt.Sprintf("%v", row[col])
		}
		table.Append(line)
	}

	table.Render()

	// TODO: store page state and query here so that we can let the user page through results with space

	return iter.Close()
}

func (c *CQL) exec(line string) error {
	// TODO: parse and do other things
	return c.executeQuery(line)
}
