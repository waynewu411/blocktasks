package repository

type Repository interface {
	Transaction(fn func(Repository) error) error

	TaskDao() TaskDao
	LogDao() LogDao
}
