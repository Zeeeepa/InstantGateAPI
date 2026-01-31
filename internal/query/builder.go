package query

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/proyaai/instantgate/internal/database/mysql"
)

type Builder struct {
	sb     sq.StatementBuilderType
	schema *mysql.SchemaCache
}

func NewBuilder(schema *mysql.SchemaCache) *Builder {
	return &Builder{
		sb:     sq.StatementBuilder.PlaceholderFormat(sq.Question),
		schema: schema,
	}
}

func (b *Builder) BuildSelect(table string, params *QueryParams) (string, []interface{}, error) {
	tableSchema, exists := b.schema.Get(table)
	if !exists {
		return "", nil, fmt.Errorf("table '%s' not found", table)
	}

	var columns []string
	if len(params.Fields) > 0 {
		for _, field := range params.Fields {
			if _, ok := tableSchema.Columns[field]; !ok {
				return "", nil, fmt.Errorf("unknown column '%s' in table '%s'", field, table)
			}
			columns = append(columns, field)
		}
	} else {
		for colName := range tableSchema.Columns {
			columns = append(columns, colName)
		}
	}

	query := b.sb.Select(columns...).From(table)

	for _, filter := range params.Filters {
		if _, ok := tableSchema.Columns[filter.Field]; !ok {
			return "", nil, fmt.Errorf("unknown column '%s' in table '%s'", filter.Field, table)
		}

		query = applyFilter(query, filter)
	}

	if params.Sorting != nil {
		if _, ok := tableSchema.Columns[params.Sorting.Field]; !ok {
			return "", nil, fmt.Errorf("unknown column '%s' for sorting", params.Sorting.Field)
		}
		orderClause := params.Sorting.Field
		if params.Sorting.Direction == "desc" {
			orderClause += " DESC"
		} else {
			orderClause += " ASC"
		}
		query = query.OrderBy(orderClause)
	}

	if params.Pagination != nil {
		if params.Pagination.Limit > 0 {
			query = query.Limit(uint64(params.Pagination.Limit))
		}
		if params.Pagination.Offset > 0 {
			query = query.Offset(uint64(params.Pagination.Offset))
		}
	}

	return query.ToSql()
}

func (b *Builder) BuildSelectByID(table string, id interface{}, fields []string) (string, []interface{}, error) {
	tableSchema, exists := b.schema.Get(table)
	if !exists {
		return "", nil, fmt.Errorf("table '%s' not found", table)
	}

	if tableSchema.PrimaryKey == "" {
		return "", nil, fmt.Errorf("table '%s' has no primary key", table)
	}

	var columns []string
	if len(fields) > 0 {
		for _, field := range fields {
			if _, ok := tableSchema.Columns[field]; !ok {
				return "", nil, fmt.Errorf("unknown column '%s' in table '%s'", field, table)
			}
			columns = append(columns, field)
		}
	} else {
		for colName := range tableSchema.Columns {
			columns = append(columns, colName)
		}
	}

	query := b.sb.Select(columns...).
		From(table).
		Where(sq.Eq{tableSchema.PrimaryKey: id}).
		Limit(1)

	return query.ToSql()
}

func (b *Builder) BuildCount(table string, params *QueryParams) (string, []interface{}, error) {
	_, exists := b.schema.Get(table)
	if !exists {
		return "", nil, fmt.Errorf("table '%s' not found", table)
	}

	query := b.sb.Select("COUNT(*) as count").From(table)

	for _, filter := range params.Filters {
		query = applyFilter(query, filter)
	}

	return query.ToSql()
}

func (b *Builder) BuildInsert(table string, data map[string]interface{}) (string, []interface{}, error) {
	tableSchema, exists := b.schema.Get(table)
	if !exists {
		return "", nil, fmt.Errorf("table '%s' not found", table)
	}

	columns := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for col, val := range data {
		colInfo, ok := tableSchema.Columns[col]
		if !ok {
			return "", nil, fmt.Errorf("unknown column '%s' in table '%s'", col, table)
		}

		if colInfo.IsAutoIncrement {
			continue
		}

		columns = append(columns, col)
		values = append(values, val)
	}

	query := b.sb.Insert(table).
		Columns(columns...).
		Values(values...)

	return query.ToSql()
}

func (b *Builder) BuildUpdate(table string, id interface{}, data map[string]interface{}) (string, []interface{}, error) {
	tableSchema, exists := b.schema.Get(table)
	if !exists {
		return "", nil, fmt.Errorf("table '%s' not found", table)
	}

	if tableSchema.PrimaryKey == "" {
		return "", nil, fmt.Errorf("table '%s' has no primary key", table)
	}

	updateData := make(map[string]interface{})
	for col, val := range data {
		colInfo, ok := tableSchema.Columns[col]
		if !ok {
			return "", nil, fmt.Errorf("unknown column '%s' in table '%s'", col, table)
		}

		if colInfo.IsPrimaryKey || colInfo.IsAutoIncrement {
			continue
		}

		updateData[col] = val
	}

	if len(updateData) == 0 {
		return "", nil, fmt.Errorf("no updateable columns provided")
	}

	query := b.sb.Update(table).
		SetMap(updateData).
		Where(sq.Eq{tableSchema.PrimaryKey: id})

	return query.ToSql()
}

func (b *Builder) BuildDelete(table string, id interface{}) (string, []interface{}, error) {
	tableSchema, exists := b.schema.Get(table)
	if !exists {
		return "", nil, fmt.Errorf("table '%s' not found", table)
	}

	if tableSchema.PrimaryKey == "" {
		return "", nil, fmt.Errorf("table '%s' has no primary key", table)
	}

	query := b.sb.Delete(table).
		Where(sq.Eq{tableSchema.PrimaryKey: id})

	return query.ToSql()
}

func applyFilter(query sq.SelectBuilder, filter Filter) sq.SelectBuilder {
	switch filter.Operator {
	case OpEqual:
		return query.Where(sq.Eq{filter.Field: filter.Value})
	case OpNotEqual:
		return query.Where(sq.NotEq{filter.Field: filter.Value})
	case OpGreater:
		return query.Where(sq.Gt{filter.Field: filter.Value})
	case OpGreaterEqual:
		return query.Where(sq.GtOrEq{filter.Field: filter.Value})
	case OpLess:
		return query.Where(sq.Lt{filter.Field: filter.Value})
	case OpLessEqual:
		return query.Where(sq.LtOrEq{filter.Field: filter.Value})
	case OpLike:
		return query.Where(sq.Like{filter.Field: filter.Value})
	case OpNotLike:
		return query.Where(sq.NotLike{filter.Field: filter.Value})
	case OpIn:
		return query.Where(sq.Eq{filter.Field: filter.Values})
	case OpNotIn:
		return query.Where(sq.NotEq{filter.Field: filter.Values})
	default:
		return query.Where(sq.Eq{filter.Field: filter.Value})
	}
}
