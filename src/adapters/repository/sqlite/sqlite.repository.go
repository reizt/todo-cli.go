package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/reizt/todo/src/core"
)

var (
	todoTableName   = "todo"
	todoColumnNames = struct {
		ID          string
		Title       string
		Description string
	}{
		ID:          "id",
		Title:       "title",
		Description: "description",
	}
	ErrDatabaseInit  = errors.New("failed to init database")
	ErrDatabaseClose = errors.New("failed to close database")
)

type repository struct {
	db *sql.DB
}

func migrate(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS ` + todoTableName + ` (
			` + todoColumnNames.ID + ` VARCHAR(255) NOT NULL,
			` + todoColumnNames.Title + ` VARCHAR(255) NOT NULL,
			` + todoColumnNames.Description + ` TEXT,
			PRIMARY KEY (id)
		);
	`
	_, err := db.Prepare(query)
	if err != nil {
		return err
	}
	db.Exec(query)
	return nil
}

func Init() (*repository, error) {
	storagePath := os.Getenv("TODO_CLI_SQLITE_STORAGE_PATH")
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		fmt.Println(err.Error())
		return nil, ErrDatabaseInit
	}

	if err := migrate(db); err != nil {
		fmt.Println(err.Error())
		return nil, ErrDatabaseInit
	}

	return &repository{db}, nil
}

func (repo repository) Close() error {
	err := repo.db.Close()
	if err != nil {
		fmt.Println(err.Error())
		return ErrDatabaseClose
	}
	return nil
}

func (repo repository) FindMany(input core.IRepositoryFindManyInput) (*([]core.Todo), error) {
	rows, err := repo.db.Query("SELECT * FROM todo;")
	if err != nil {
		return nil, core.ErrRepositoryFindMany
	}
	defer rows.Close()

	todos := []core.Todo{}
	for rows.Next() {
		todo := core.Todo{}
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Description); err != nil {
			return nil, core.ErrRepositoryUnexpected
		}
		todos = append(todos, todo)
	}
	return &todos, nil
}

func (repo repository) FindById(id string) (*(core.Todo), error) {
	selectQuery := "SELECT * FROM todo WHERE " + todoColumnNames.ID + " = ?;"
	row := repo.db.QueryRow(selectQuery, id)
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			return nil, core.ErrRepositoryNotFound
		}
		fmt.Println(err.Error())
		return nil, core.ErrRepositoryUnexpected
	}

	todo := core.Todo{}
	if err := row.Scan(&todo.ID, &todo.Title, &todo.Description); err != nil {
		if err == sql.ErrNoRows {
			return nil, core.ErrRepositoryNotFound
		}
		fmt.Println(err.Error())
		return nil, core.ErrRepositoryUnexpected
	}
	return &todo, nil
}

func (repo repository) Insert(input core.IRepositoryInsertInput) (*core.Todo, error) {
	insertQuery := fmt.Sprintf(
		"INSERT INTO %s (%s, %s, %s) VALUES (?, ?, ?);",
		todoTableName,
		todoColumnNames.ID,
		todoColumnNames.Title,
		todoColumnNames.Description,
	)
	_, err := repo.db.Exec(insertQuery, input.ID, input.Title, input.Description)
	if err != nil {
		return nil, core.ErrRepositoryInsertFailed
	}

	todo := core.Todo(input)
	return &todo, nil
}

func (repo repository) Update(id string, input core.IRepositoryUpdateInput) (err error) {
	if input.Title != nil && input.Description != nil {
		updateQuery := fmt.Sprintf(
			"UPDATE %s SET %s = ?, %s = ? WHERE %s = ?;",
			todoTableName,
			todoColumnNames.Title,
			todoColumnNames.Description,
			todoColumnNames.ID,
		)
		_, err = repo.db.Exec(updateQuery, input.Title, input.Description, id)
	} else if input.Title != nil {
		updateQuery := fmt.Sprintf(
			"UPDATE %s SET %s = ? WHERE %s = ?;",
			todoTableName,
			todoColumnNames.Title,
			todoColumnNames.ID,
		)
		_, err = repo.db.Exec(updateQuery, input.Title, id)
	} else {
		updateQuery := fmt.Sprintf(
			"UPDATE %s SET %s = ? WHERE %s = ?;",
			todoTableName,
			todoColumnNames.Description,
			todoColumnNames.ID,
		)
		_, err = repo.db.Exec(updateQuery, input.Description, id)
	}
	if err != nil {
		return core.ErrRepositoryUpdateFailed
	}

	return nil
}

func (repo repository) Delete(id string) error {
	deleteQuery := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = ?;",
		todoTableName,
		todoColumnNames.ID,
	)
	_, err := repo.db.Exec(deleteQuery, id)
	if err != nil {
		return core.ErrRepositoryDeleteFailed
	}

	return nil
}

func (repo repository) DeleteAll() error {
	deleteQuery := fmt.Sprintf(
		"DELETE FROM %s;",
		todoTableName,
	)
	_, err := repo.db.Exec(deleteQuery)
	if err != nil {
		return core.ErrRepositoryDeleteAllFailed
	}

	return nil
}