package types

import (
	"database/sql"
	"time"
)

func Convert2NullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func Convert2NullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

//type NullString sql.NullString
//
//func (ns *NullString) MarshalJSON() ([]byte, error) {
//	if !ns.Valid {
//		return []byte("null"), nil
//	}
//	return json.Marshal(ns.String)
//}
//
//func (ns *NullString) UnmarshalJSON(data []byte) error {
//	var x *string
//	if err := json.Unmarshal(data, &x); err != nil {
//		return err
//	}
//	ns.Valid = x != nil
//	if ns.Valid {
//		ns.String = *x
//	}
//	return nil
//}
