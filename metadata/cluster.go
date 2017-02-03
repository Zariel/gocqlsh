package metadata

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/gocql/gocql"
)

func New(db *gocql.Session) *Cassandra {
	// TODO: should we just dial a single conn and not use a session?
	return &Cassandra{
		db: db,
	}
}

type Cassandra struct {
	db *gocql.Session
}

type Cluster struct {
	Version    Version
	CQLVersion string
	Name       string
	Protocol   string
	Address    net.IP
}

func (c *Cassandra) ClusterMeta() (*Cluster, error) {
	meta := &Cluster{}
	// TODO: the driver needs a way to export this on a per cluster basis, cluster metadata?
	err := c.db.Query("SELECT release_version, cql_version, cluster_name, native_protocol_version, listen_address FROM system.local").Scan(
		&meta.Version, &meta.CQLVersion, &meta.Name, &meta.Protocol, &meta.Address)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

type Version struct {
	Major, Minor, Patch int
}

func (c *Version) Set(v string) error {
	if v == "" {
		return nil
	}

	return c.UnmarshalCQL(nil, []byte(v))
}

func (c *Version) UnmarshalCQL(info gocql.TypeInfo, data []byte) error {
	return c.unmarshal(data)
}

func (c *Version) unmarshal(data []byte) error {
	version := strings.TrimSuffix(string(data), "-SNAPSHOT")
	version = strings.TrimPrefix(version, "v")
	v := strings.Split(version, ".")

	if len(v) < 2 {
		return fmt.Errorf("invalid version string: %s", data)
	}

	var err error
	c.Major, err = strconv.Atoi(v[0])
	if err != nil {
		return fmt.Errorf("invalid major version %v: %v", v[0], err)
	}

	c.Minor, err = strconv.Atoi(v[1])
	if err != nil {
		return fmt.Errorf("invalid minor version %v: %v", v[1], err)
	}

	if len(v) > 2 {
		c.Patch, err = strconv.Atoi(v[2])
		if err != nil {
			return fmt.Errorf("invalid patch version %v: %v", v[2], err)
		}
	}

	return nil
}

func (c Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", c.Major, c.Minor, c.Patch)
}
