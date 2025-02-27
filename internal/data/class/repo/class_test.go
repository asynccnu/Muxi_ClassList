package repo

import (
	"context"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	bizmodel "github.com/asynccnu/Muxi_ClassList/internal/biz/model"
	"github.com/asynccnu/Muxi_ClassList/internal/data/class/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"regexp"
	"testing"
)

func Test_extractClassIDs(t *testing.T) {

	class1 := &model.ClassDO{ID: "1"}
	class2 := &model.ClassDO{ID: "2"}

	type args struct {
		classes []*model.ClassDO
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "normal",
			args: args{
				classes: []*model.ClassDO{class1, class2},
			},
			want: []string{class1.ID, class2.ID},
		},
		{
			name: "none",
			args: args{
				classes: nil,
			},
			want: nil,
		},
		{
			name: "have nil",
			args: args{
				classes: []*model.ClassDO{class1, nil},
			},
			want: []string{class1.ID},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := extractClassIDs(tt.args.classes)
			assert.ElementsMatch(t, res, tt.want)
		})
	}
}

func Test_batchNewBizClasses(t *testing.T) {
	class1 := &model.ClassDO{ID: "1"}
	class2 := &model.ClassDO{ID: "2"}

	type args struct {
		classes []*model.ClassDO
	}
	tests := []struct {
		name string
		args args
		want []*bizmodel.ClassBiz
	}{
		{
			name: "none",
			args: args{
				classes: nil,
			},
			want: nil,
		},
		{
			name: "normal",
			args: args{
				classes: []*model.ClassDO{class1, class2},
			},
			want: []*bizmodel.ClassBiz{{ID: class1.ID}, {ID: class2.ID}},
		},
		{
			name: "have nil",
			args: args{
				classes: []*model.ClassDO{class1, nil},
			},
			want: []*bizmodel.ClassBiz{{ID: class1.ID}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := batchNewBizClasses(tt.args.classes)
			assert.ElementsMatch(t, res, tt.want)
		})
	}
}

func initDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}))
	if err != nil {
		t.Fatal(err)
	}
	return gormDB, mock
}

func TestClassRepo_getClassesFromDBByIDs(t *testing.T) {
	gormDB, mock := initDB(t)

	classRepo := &ClassRepo{db: gormDB}

	class1 := &model.ClassDO{ID: "1"}
	class2 := &model.ClassDO{ID: "2"}

	type args struct {
		classIDs []string
	}
	tests := []struct {
		name    string
		mock    func(t *testing.T, mock sqlmock.Sqlmock, args2 args)
		args    args
		want    []*model.ClassDO
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "normal - single class ID",
			args: args{
				classIDs: []string{class1.ID},
			},
			mock: func(t *testing.T, mock sqlmock.Sqlmock, args2 args) {
				sql := fmt.Sprintf("SELECT * FROM `%s` WHERE id IN (?)", model.ClassDOTableName)
				mock.ExpectQuery(regexp.QuoteMeta(sql)).
					WithArgs(args2.classIDs[0]).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "created_at", "updated_at", "day", "teacher", "where", "class_when", "week_duration", "class_name", "credit", "weeks", "semester", "year"}).
							AddRow(class1.ID, class1.CreatedAt, class1.UpdatedAt, class1.Day, class1.Teacher, class1.Where, class1.ClassWhen, class1.WeekDuration, class1.Classname, class1.Credit, class1.Weeks, class1.Semester, class1.Year))

			},
			want:    []*model.ClassDO{class1},
			wantErr: assert.NoError,
		},
		{
			name: "normal - multiple class IDs",
			args: args{
				classIDs: []string{class1.ID, class2.ID},
			},
			mock: func(t *testing.T, mock sqlmock.Sqlmock, args2 args) {
				sql := fmt.Sprintf("SELECT * FROM `%s` WHERE id IN (?,?)", model.ClassDOTableName)

				mock.ExpectQuery(regexp.QuoteMeta(sql)).
					WithArgs(args2.classIDs[0], args2.classIDs[1]).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "created_at", "updated_at", "day", "teacher", "where", "class_when", "week_duration", "class_name", "credit", "weeks", "semester", "year"}).
							AddRow(class1.ID, class1.CreatedAt, class1.UpdatedAt, class1.Day, class1.Teacher, class1.Where, class1.ClassWhen, class1.WeekDuration, class1.Classname, class1.Credit, class1.Weeks, class1.Semester, class1.Year).
							AddRow(class2.ID, class2.CreatedAt, class2.UpdatedAt, class2.Day, class2.Teacher, class2.Where, class2.ClassWhen, class2.WeekDuration, class2.Classname, class2.Credit, class2.Weeks, class2.Semester, class2.Year))

			},
			want:    []*model.ClassDO{class1, class2},
			wantErr: assert.NoError,
		},
		{
			name: "none - empty class IDs",
			args: args{
				classIDs: []string{},
			},
			mock: func(t *testing.T, mock sqlmock.Sqlmock, args2 args) {
				// 不需要设置 mock.ExpectQuery，因为不会执行 SQL 查询
			},
			want:    nil,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(t, mock, tt.args)
			res, err := classRepo.getClassesFromDBByIDs(context.Background(), tt.args.classIDs...)
			tt.wantErr(t, err)
			assert.ElementsMatch(t, res, tt.want)
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestClassRepo_getClassesFromDB(t *testing.T) {
	gormDB, mock := initDB(t)

	classRepo := &ClassRepo{db: gormDB}

	class1 := &model.ClassDO{ID: "1"}

	type args struct {
		stuID    string
		year     string
		semester string
	}
	tests := []struct {
		name    string
		args    args
		mock    func(t *testing.T, mock sqlmock.Sqlmock, args2 args)
		want    []*model.ClassDO
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "normal",
			args: args{
				stuID:    "123",
				year:     "2023",
				semester: "1",
			},
			mock: func(t *testing.T, mock sqlmock.Sqlmock, args2 args) {
				sql := fmt.Sprintf("SELECT * FROM `%s` LEFT JOIN %s ON %s.id = %s.cla_id WHERE %s.stu_id = ? AND %s.year = ? AND %s.semester = ?",
					model.ClassDOTableName, model.StudentClassRelationDOTableName, model.ClassDOTableName, model.StudentClassRelationDOTableName, model.StudentClassRelationDOTableName, model.StudentClassRelationDOTableName, model.StudentClassRelationDOTableName)
				mock.ExpectQuery(regexp.QuoteMeta(sql)).
					WithArgs(args2.stuID, args2.year, args2.semester).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "created_at", "updated_at", "day", "teacher", "where", "class_when", "week_duration", "class_name", "credit", "weeks", "semester", "year"}).
							AddRow(class1.ID, class1.CreatedAt, class1.UpdatedAt, class1.Day, class1.Teacher, class1.Where, class1.ClassWhen, class1.WeekDuration, class1.Classname, class1.Credit, class1.Weeks, class1.Semester, class1.Year))

			},
			want:    []*model.ClassDO{class1},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(t, mock, tt.args)
			res, err := classRepo.getClassesFromDB(context.Background(), tt.args.stuID, tt.args.year, tt.args.semester)
			tt.wantErr(t, err)
			assert.ElementsMatch(t, res, tt.want)
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestClassRepo_saveClassInDB(t *testing.T) {
	gormDB, mock := initDB(t)

	classRepo := &ClassRepo{db: gormDB}

	type args struct {
		stuID    string
		year     string
		semester string
		classes  []*bizmodel.ClassBiz
	}
	tests := []struct {
		name    string
		args    args
		mock    func(t *testing.T, mock sqlmock.Sqlmock, args2 args)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Transaction execution successful",
			args: args{
				stuID:    "123",
				year:     "2023",
				semester: "1",
				classes:  []*bizmodel.ClassBiz{{ID: "1"}, {ID: "2"}},
			},
			mock: func(t *testing.T, mock sqlmock.Sqlmock, args2 args) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE *").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT *").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT *").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: assert.NoError,
		},
		{
			name: "Some transactions failed to execute",
			args: args{
				stuID:    "123",
				year:     "2023",
				semester: "1",
				classes:  []*bizmodel.ClassBiz{{ID: "1"}, {ID: "2"}},
			},
			mock: func(t *testing.T, mock sqlmock.Sqlmock, args2 args) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE *").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT *").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT *").WillReturnError(errors.New("failed to execute"))
				mock.ExpectRollback()
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(t, mock, tt.args)
			err := classRepo.saveClassInDB(context.Background(), tt.args.stuID, tt.args.year, tt.args.semester, tt.args.classes)
			tt.wantErr(t, err)
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}
