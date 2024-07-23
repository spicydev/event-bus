package spannerDb

import (
	"time"
)

type AccountXref struct {
	XrefVal        string    `spanner:"XREF_VAL"`
	XrefType       string    `spanner:"XREF_TYPE"`
	UserGuid       string    `spanner:"USER_GUID"`
	UserId         int64     `spanner:"USER_ID"`
	CreatedTs      time.Time `spanner:"CREATE_TS"`
	LastModifiedTs time.Time `spanner:"LAST_MOD_TS"`
}

type Account struct {
	UserGuid       string    `spanner:"USER_GUID"`
	UserId         int64     `spanner:"USER_ID"`
	FirstName      string    `spanner:"FIRST_NM"`
	LastName       string    `spanner:"LAST_NM"`
	Phone          string    `spanner:"PHONE"`
	Email          string    `spanner:"EMAIL"`
	CreatedTs      time.Time `spanner:"CREATE_TS"`
	LastModifiedTs time.Time `spanner:"LAST_MOD_TS"`
}
