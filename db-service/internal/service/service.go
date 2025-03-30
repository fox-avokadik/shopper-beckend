package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	database "db-service/proto"
	"gorm.io/gorm"
)

type Server struct {
	database.UnimplementedDatabaseServiceServer
	db *gorm.DB
}

func NewDatabaseService(db *gorm.DB) *Server {
	return &Server{db: db}
}

func (s *Server) ExecuteQuery(ctx context.Context, req *database.ExecuteQueryRequest) (*database.ExecuteQueryResponse, error) {
	if err := validateQueryRequest(req); err != nil {
		return errorResponse(err), nil
	}

	params, err := parseQueryParams(req.Params)
	if err != nil {
		return errorResponse(err), nil
	}

	rows, err := executeSQLQuery(s.db, req.Query, params)
	if err != nil {
		return errorResponse(err), nil
	}
	defer rows.Close()

	results, err := processQueryResults(rows)
	if err != nil {
		return errorResponse(err), nil
	}

	return formatQueryResponse(results)
}

func validateQueryRequest(req *database.ExecuteQueryRequest) error {
	if req.Query == "" {
		return fmt.Errorf("query cannot be empty")
	}
	return nil
}

func parseQueryParams(grpcParams []*database.QueryParam) ([]interface{}, error) {
	params := make([]interface{}, len(grpcParams))
	for i, param := range grpcParams {
		switch v := param.Value.(type) {
		case *database.QueryParam_StrValue:
			params[i] = v.StrValue
		case *database.QueryParam_IntValue:
			params[i] = v.IntValue
		case *database.QueryParam_UintValue:
			params[i] = v.UintValue
		case *database.QueryParam_FloatValue:
			params[i] = v.FloatValue
		case *database.QueryParam_DoubleValue:
			params[i] = v.DoubleValue
		case *database.QueryParam_BoolValue:
			params[i] = v.BoolValue
		case *database.QueryParam_BytesValue:
			params[i] = v.BytesValue
		case *database.QueryParam_TimeValue:
			params[i] = v.TimeValue.AsTime()
		default:
			return nil, fmt.Errorf("unsupported parameter type")
		}
	}

	return params, nil
}

func executeSQLQuery(db *gorm.DB, query string, params []interface{}) (*sql.Rows, error) {
	return db.Raw(query, params...).Rows()
}

func processQueryResults(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		rowData, err := scanRow(rows, columns)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, rowData)
	}
	return results, nil
}

func scanRow(rows *sql.Rows, columns []string) (map[string]interface{}, error) {
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	rowData := make(map[string]interface{})
	for i, col := range columns {
		rowData[col] = convertValue(values[i])
	}
	return rowData, nil
}

func convertValue(val interface{}) interface{} {
	switch v := val.(type) {
	case time.Time:
		return v.Format(time.RFC3339)
	case []byte:
		return string(v)
	default:
		return v
	}
}

func formatQueryResponse(results []map[string]interface{}) (*database.ExecuteQueryResponse, error) {
	if len(results) == 0 {
		return errorResponse(fmt.Errorf("no rows found")), nil
	}

	var resultJSON []byte
	var err error

	if len(results) == 1 {
		resultJSON, err = json.Marshal(results[0])
	} else {
		resultJSON, err = json.Marshal(results)
	}

	if err != nil {
		return errorResponse(fmt.Errorf("failed to marshal JSON: %w", err)), nil
	}

	return &database.ExecuteQueryResponse{
		Result: string(resultJSON),
	}, nil
}

func errorResponse(err error) *database.ExecuteQueryResponse {
	return &database.ExecuteQueryResponse{
		Error: err.Error(),
	}
}
