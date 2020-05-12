package db

import (
	"fmt"
	"net/url"

	"github.com/spf13/viper"
)

// MySQL ...
const MySQL = "MYSQL"

// MSSQL ...
const MSSQL = "MSSQL"

// PostGres ...
const PostGres = "POSTGRES"

const mySQLFormat = "%s@(%s)/%s"
const msSQLFormat = "sqlserver://%s@%s?database=%s"
const postGresFormat = "user=%s password=%s host=%s port=%s dbname=%s"

// ConnStringBuilder ...
type ConnStringBuilder struct {
	Host     string
	Port     string
	Username string
	Password string
	Dbname   string
	Format   string
	Options  map[string]string
}

// SetHost ...
func (c *ConnStringBuilder) SetHost(Host string) *ConnStringBuilder {
	c.Host = Host
	return c
}

// SetPort ...
func (c *ConnStringBuilder) SetPort(Port string) *ConnStringBuilder {
	c.Port = Port
	return c
}

// SetUsername ...
func (c *ConnStringBuilder) SetUsername(Username string) *ConnStringBuilder {
	c.Username = Username
	return c
}

// SetPassword ...
func (c *ConnStringBuilder) SetPassword(Password string) *ConnStringBuilder {
	c.Password = Password
	return c
}

// SetDbname ...
func (c *ConnStringBuilder) SetDbname(Dbname string) *ConnStringBuilder {
	c.Dbname = Dbname
	return c
}

// SetFormat ...
func (c *ConnStringBuilder) SetFormat(Format string) *ConnStringBuilder {
	c.Format = Format
	return c
}

// SetOptions ...
func (c *ConnStringBuilder) SetOptions(Options map[string]string) *ConnStringBuilder {
	c.Options = Options
	return c
}

// SetWithConfig ...
func (c *ConnStringBuilder) SetWithConfig(cfg *viper.Viper) *ConnStringBuilder {
	c.Username = cfg.GetString("username")
	c.Password = cfg.GetString("password")
	c.Dbname = cfg.GetString("dbname")
	c.Port = cfg.GetString("port")
	c.Host = cfg.GetString("host")
	c.Format = cfg.GetString("format")
	return c
}

// Build ...
func (c *ConnStringBuilder) Build() string {
	host := c.Host
	auth := c.Username

	if c.Port != "" {
		host = host + ":" + c.Port
	}

	if c.Password != "" {
		auth = auth + ":" + c.Password
	}

	if c.Options == nil {
		c.Options = generateDefaultOption()
	}

	if c.Format == "" || c.Format == MySQL {
		return fmt.Sprintf(mySQLFormat, auth, host, c.Dbname) + encodeOptions(c.Options)
	} else if c.Format == MSSQL {
		return fmt.Sprintf(msSQLFormat, auth, host, c.Dbname)
	} else if c.Format == PostGres {
		return fmt.Sprintf(postGresFormat, c.Username, c.Password, c.Host, c.Port, c.Dbname)
	}
	return ""
}

func generateDefaultOption() map[string]string {
	options := make(map[string]string)
	options["loc"] = "Local"
	options["parseTime"] = "True"
	options["charset"] = "utf8"
	return options
}

func encodeOptions(Options map[string]string) string {
	if len(Options) != 0 {
		query := url.Values{}
		for k, v := range Options {
			query.Add(k, v)
		}
		return "?" + query.Encode()
	}
	return ""
}
