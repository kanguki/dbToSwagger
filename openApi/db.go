package openApi

import (
	"database/sql"
	"fmt"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
)

type Db interface {
	Read(chan RawData, DBOptions)
}
type DBOptions struct {
	ClientId, Domain string
}
type RawData struct {
	name, tags, uriPattern, summary, security, parameters, requestBody, responses, version, forward_data string
}

type OpenApiDb struct {
	// Url string //"name:password@tcp(host:port)/db_name"
	Conn *sql.DB
}

func (d *OpenApiDb) Read(data chan RawData, opts DBOptions) {
	if hasSuspiciousCharacter(opts.ClientId) || hasSuspiciousCharacter(opts.Domain) {
		Log("input is suspicious: %+v", opts)
		close(data)
		return
	}
	query := fmt.Sprintf(`
select o.operation_id, o.tags, o.uri_pattern, o.summary, o.security, o.parameters, o.request_body, o.responses, o.version, s.forward_data from 
t_open_api o join t_scope s on o.id = s.id
join t_scope_scope_group_map sgm on s.id = sgm.scope_id
join t_scope_group g on sgm.scope_group_id = g.id
join t_login_method_scope_group_map lgm on g.id = lgm.scope_group_id
join t_login_method l on l.id = lgm.login_method_id
join t_client_login_method_map clm on l.id = clm.login_method_id
join t_client c on c.id = clm.client_id
where c.client_id = "%v" and c.domain = "%v"; `, opts.ClientId, opts.Domain)
	rawData, err := d.Conn.Query(query)
	if err != nil {
		Log("error reading from db: %v", err)
		close(data)
		return
	}
	go func() {
		defer close(data)
		for rawData.Next() {
			var name, tags, uriPattern, summary, security, parameters, requestBody, responses, version, forward_data sql.NullString
			err := rawData.Scan(&name, &tags, &uriPattern, &summary, &security, &parameters, &requestBody, &responses, &version, &forward_data)
			if err != nil {
				Log("error reading from db: %v", err)
				return
			}
			r := RawData{
				name:         name.String,
				tags:         tags.String,
				uriPattern:   uriPattern.String,
				summary:      summary.String,
				security:     security.String,
				parameters:   parameters.String,
				requestBody:  requestBody.String,
				responses:    responses.String,
				version:      version.String,
				forward_data: forward_data.String,
			}
			Debug("%v", r)
			data <- r
		}
		Debug("total records: %v", len(data))
	}()
}

func hasSuspiciousCharacter(s string) bool {
	hasSuspiciousCharacter, err := regexp.MatchString(`[^-A-Za-z0-9_ ]+`, s)
	if err != nil {
		Debug("checkInjectingInput failed: %v", err)
		return false
	}
	return hasSuspiciousCharacter
}
