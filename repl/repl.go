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
			fmt.Printf("line=%q\n", line)

			// TODO: dont just fail because bad queries should be written out to the repl
			return err
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

	for _, row := range rows {
		var line []string
		for _, col := range columns {
			line = append(line, fmt.Sprintf("%v", row[col]))
		}
		table.Append(line)
	}

	table.Render()

	return nil
}

func (c *CQL) exec(line string) error {
	// TODO: parse and do other things
	return c.executeQuery(line)
}
